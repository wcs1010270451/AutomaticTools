package logic

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"automatictools/backend/internal/store"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	registrationCodePurpose     = "register"
	registrationCodeTTL         = 10 * time.Minute
	registrationCodeCooldown    = 60 * time.Second
	registrationCodeMaxAttempts = 5
)

var verificationCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

type SendRegistrationEmailCodeRequest struct {
	Email string `json:"email"`
}

type EmailCodeResponse struct {
	OK          bool `json:"ok"`
	ExpiresIn   int  `json:"expiresIn"`
	ResendAfter int  `json:"resendAfter"`
}

func (a *Service) SendRegistrationEmailCode(
	ctx context.Context,
	req SendRegistrationEmailCodeRequest,
	meta RequestMeta,
) (EmailCodeResponse, error) {
	email := normalizeEmail(req.Email)
	if email == "" {
		return EmailCodeResponse{}, badRequest("邮箱不能为空。")
	}
	if err := validateContacts(email, ""); err != nil {
		return EmailCodeResponse{}, err
	}
	if a.emailSender == nil {
		return EmailCodeResponse{}, serviceUnavailable(
			"邮件服务暂不可用，请稍后重试。",
			errors.New("email sender dependency is nil"),
		)
	}

	code, err := newVerificationCode()
	if err != nil {
		return EmailCodeResponse{}, err
	}
	now := unixNow()
	record := store.EmailVerificationCode{
		Email:     email,
		Purpose:   registrationCodePurpose,
		CodeHash:  a.hashVerificationCode(email, code),
		ExpiresAt: now + int64(registrationCodeTTL/time.Second),
		CreatedAt: now,
	}

	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockKey := "email-code:" + registrationCodePurpose + ":" + email
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
			return err
		}

		var existingUsers int64
		if err := tx.Model(&store.User{}).
			Where("LOWER(email) = ?", email).
			Count(&existingUsers).Error; err != nil {
			return err
		}
		if existingUsers > 0 {
			return conflict("邮箱已被注册。")
		}

		var latest store.EmailVerificationCode
		err := tx.
			Where("email = ? AND purpose = ?", email, registrationCodePurpose).
			Order("created_at DESC").
			Take(&latest).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			remaining := latest.CreatedAt + int64(registrationCodeCooldown/time.Second) - now
			if remaining > 0 {
				return conflict(fmt.Sprintf("请在 %d 秒后重新获取验证码。", remaining))
			}
		}

		return tx.Create(&record).Error
	})
	if err != nil {
		return EmailCodeResponse{}, err
	}

	if err := a.emailSender.SendRegistrationCode(
		email,
		code,
		int(registrationCodeTTL/time.Minute),
	); err != nil {
		_ = a.db.WithContext(ctx).Delete(&store.EmailVerificationCode{}, record.ID).Error
		return EmailCodeResponse{}, serviceUnavailable("验证码邮件发送失败，请稍后重试。", err)
	}

	consumedAt := unixNow()
	_ = a.db.WithContext(ctx).
		Model(&store.EmailVerificationCode{}).
		Where(
			"email = ? AND purpose = ? AND id <> ? AND consumed_at IS NULL",
			email,
			registrationCodePurpose,
			record.ID,
		).
		Update("consumed_at", consumedAt).Error
	_ = a.db.WithContext(ctx).
		Where("expires_at < ?", now-int64(24*time.Hour/time.Second)).
		Delete(&store.EmailVerificationCode{}).Error

	detail, _ := json.Marshal(map[string]string{"email": maskEmail(email)})
	_ = a.audit(ctx, meta, nil, "email.registration_code.send", string(detail))

	return EmailCodeResponse{
		OK:          true,
		ExpiresIn:   int(registrationCodeTTL / time.Second),
		ResendAfter: int(registrationCodeCooldown / time.Second),
	}, nil
}

func (a *Service) verifyRegistrationEmailCode(
	tx *gorm.DB,
	email string,
	code string,
	now int64,
) (store.EmailVerificationCode, error) {
	var record store.EmailVerificationCode
	err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"email = ? AND purpose = ? AND consumed_at IS NULL",
			email,
			registrationCodePurpose,
		).
		Order("created_at DESC").
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return store.EmailVerificationCode{}, badRequest("验证码无效或已过期，请重新获取。")
	}
	if err != nil {
		return store.EmailVerificationCode{}, err
	}

	if record.ExpiresAt <= now {
		if err := tx.Model(&store.EmailVerificationCode{}).
			Where("id = ?", record.ID).
			Update("consumed_at", now).Error; err != nil {
			return store.EmailVerificationCode{}, err
		}
		return store.EmailVerificationCode{}, badRequest("验证码无效或已过期，请重新获取。")
	}
	if record.AttemptCount >= registrationCodeMaxAttempts {
		return store.EmailVerificationCode{}, badRequest("验证码错误次数过多，请重新获取。")
	}
	if !hmac.Equal(
		[]byte(record.CodeHash),
		[]byte(a.hashVerificationCode(email, code)),
	) {
		attemptCount := record.AttemptCount + 1
		updates := map[string]any{"attempt_count": attemptCount}
		if attemptCount >= registrationCodeMaxAttempts {
			updates["consumed_at"] = now
		}
		if err := tx.Model(&store.EmailVerificationCode{}).
			Where("id = ?", record.ID).
			Updates(updates).Error; err != nil {
			return store.EmailVerificationCode{}, err
		}
		if attemptCount >= registrationCodeMaxAttempts {
			return store.EmailVerificationCode{}, badRequest("验证码错误次数过多，请重新获取。")
		}
		return store.EmailVerificationCode{}, badRequest("邮箱验证码错误。")
	}
	return record, nil
}

func (a *Service) hashVerificationCode(email string, code string) string {
	mac := hmac.New(sha256.New, []byte(a.cfg.JWTSecret))
	_, _ = mac.Write([]byte(email))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(registrationCodePurpose))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func newVerificationCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func newGeneratedUsername() (string, error) {
	var random [8]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate username: %w", err)
	}
	return "user_" + hex.EncodeToString(random[:]), nil
}

func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "***"
	}
	return string([]rune(parts[0])[0]) + "***@" + parts[1]
}
