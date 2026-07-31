package logic

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"automatictools/backend/internal/store"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)
var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

const maxPasswordBytes = 72

type RegisterRequest struct {
	Username   string `json:"username,omitempty"`
	Password   string `json:"password"`
	Email      string `json:"email"`
	EmailCode  string `json:"emailCode"`
	Phone      string `json:"phone,omitempty"`
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

type LoginRequest struct {
	Account    string `json:"account,omitempty"`
	Username   string `json:"username,omitempty"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Password   string `json:"password"`
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

type AuthResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type tokenClaims struct {
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func (a *Service) Register(ctx context.Context, req RegisterRequest, meta RequestMeta) (AuthResponse, error) {
	username := normalizeUsername(req.Username)
	email := normalizeEmail(req.Email)
	phone := normalizePhone(req.Phone)
	if username == "" {
		generatedUsername, err := newGeneratedUsername()
		if err != nil {
			return AuthResponse{}, err
		}
		username = generatedUsername
	}
	if err := validateAuth(username, req.Password); err != nil {
		return AuthResponse{}, err
	}
	if email == "" {
		return AuthResponse{}, badRequest("邮箱不能为空。")
	}
	if err := validateContacts(email, phone); err != nil {
		return AuthResponse{}, err
	}
	req.EmailCode = strings.TrimSpace(req.EmailCode)
	if !verificationCodePattern.MatchString(req.EmailCode) {
		return AuthResponse{}, badRequest("邮箱验证码必须是 6 位数字。")
	}
	if err := validateOptionalDevice(req.DeviceID, req.DeviceName, req.Platform); err != nil {
		return AuthResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	now := unixNow()
	user := store.User{
		Username:     username,
		Email:        nullableString(email),
		Phone:        nullableString(phone),
		PasswordHash: string(passwordHash),
		Status:       "active",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	var verificationErr error
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		verification, err := a.verifyRegistrationEmailCode(tx, email, req.EmailCode, now)
		if err != nil {
			var appErr Error
			if errors.As(err, &appErr) {
				verificationErr = appErr
				return nil
			}
			return err
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Model(&store.EmailVerificationCode{}).
			Where("id = ?", verification.ID).
			Update("consumed_at", now).Error
	})
	if verificationErr != nil {
		return AuthResponse{}, verificationErr
	}
	if err != nil {
		if registrationConflict := mapRegistrationConflict(err); registrationConflict != nil {
			return AuthResponse{}, registrationConflict
		}
		return AuthResponse{}, err
	}

	_ = a.audit(ctx, meta, &user.ID, "user.register", "")
	if req.DeviceID != "" {
		_ = a.upsertDevice(ctx, user.ID, req.DeviceID, req.DeviceName, req.Platform)
	}

	token, err := a.issueUserToken(user.ID, username)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: token,
		User:  userDTO(user),
	}, nil
}

func mapRegistrationConflict(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return nil
	}
	switch pgErr.ConstraintName {
	case "uq_users_email":
		return conflict("邮箱已被注册。")
	case "uq_users_phone":
		return conflict("手机号已被注册。")
	default:
		return conflict("用户名已存在。")
	}
}

func (a *Service) Login(ctx context.Context, req LoginRequest, meta RequestMeta) (AuthResponse, error) {
	account := strings.TrimSpace(req.Account)
	if account == "" {
		account = strings.TrimSpace(req.Username)
	}
	if account == "" {
		account = strings.TrimSpace(req.Email)
	}
	if account == "" {
		account = strings.TrimSpace(req.Phone)
	}
	if err := validateLogin(account, req.Password); err != nil {
		return AuthResponse{}, err
	}
	if err := validateOptionalDevice(req.DeviceID, req.DeviceName, req.Platform); err != nil {
		return AuthResponse{}, err
	}

	var user store.User
	err := a.findLoginUser(ctx, account, &user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AuthResponse{}, unauthorized("账号或密码错误。")
		}
		return AuthResponse{}, err
	}

	if user.Status != "active" {
		return AuthResponse{}, forbidden("账号不可用。")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{}, unauthorized("账号或密码错误。")
	}

	now := unixNow()
	err = a.db.WithContext(ctx).
		Model(&store.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		}).Error
	if err != nil {
		return AuthResponse{}, err
	}
	user.LastLoginAt = &now
	user.UpdatedAt = now

	_ = a.audit(ctx, meta, &user.ID, "user.login", "")
	if req.DeviceID != "" {
		_ = a.upsertDevice(ctx, user.ID, req.DeviceID, req.DeviceName, req.Platform)
	}

	token, err := a.issueUserToken(user.ID, user.Username)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		Token: token,
		User:  userDTO(user),
	}, nil
}

func (a *Service) Authenticate(ctx context.Context, header string) (int64, string, error) {
	userID, _, err := a.parseUserToken(header)
	if err != nil {
		return 0, "", err
	}

	var user store.User
	err = a.db.WithContext(ctx).
		Select("id", "username", "status").
		Where("id = ? AND status = ?", userID, "active").
		Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", unauthorized("用户不存在或已停用。")
		}
		return 0, "", err
	}
	return user.ID, user.Username, nil
}

func (a *Service) parseUserToken(header string) (int64, string, error) {
	if !strings.HasPrefix(header, "Bearer ") {
		return 0, "", unauthorized("缺少登录令牌。")
	}
	tokenText := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (any, error) {
		return []byte(a.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return 0, "", unauthorized("登录令牌无效或已过期。")
	}

	if claims.Subject == "" || (claims.TokenType != "" && claims.TokenType != "user") {
		return 0, "", unauthorized("登录令牌无效。")
	}

	var userID int64
	if _, err := fmtSscan(claims.Subject, &userID); err != nil {
		return 0, "", unauthorized("登录令牌无效。")
	}
	return userID, claims.Username, nil
}

func (a *Service) findLoginUser(ctx context.Context, account string, user *store.User) error {
	if strings.Contains(account, "@") {
		return a.db.WithContext(ctx).
			Where("LOWER(email) = ?", normalizeEmail(account)).
			Take(user).Error
	}

	normalizedUsername := normalizeUsername(account)
	err := a.db.WithContext(ctx).
		Where("username = ?", normalizedUsername).
		Take(user).Error
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	phone := normalizePhone(account)
	if phonePattern.MatchString(phone) {
		return a.db.WithContext(ctx).
			Where("phone = ?", phone).
			Take(user).Error
	}
	return gorm.ErrRecordNotFound
}

func (a *Service) issueUserToken(userID int64, username string) (string, error) {
	return a.issueToken(userID, username, "user")
}

func (a *Service) issueToken(subjectID int64, username string, tokenType string) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		Username:  username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmtSprint(subjectID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(a.cfg.TokenTTLHours) * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(a.cfg.JWTSecret))
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizePhone(phone string) string {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	return replacer.Replace(strings.TrimSpace(phone))
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validateAuth(username string, password string) error {
	if !usernamePattern.MatchString(username) {
		return badRequest("用户名必须是 3 到 32 位，只能包含字母、数字、下划线、横线或点。")
	}
	if utf8.RuneCountInString(password) < 6 {
		return badRequest("密码至少需要 6 个字符。")
	}
	if len([]byte(password)) > maxPasswordBytes {
		return badRequest("密码不能超过 72 个字节。")
	}
	return nil
}

func validateLogin(account string, password string) error {
	if account == "" {
		return badRequest("账号不能为空。")
	}
	if utf8.RuneCountInString(account) > 254 {
		return badRequest("账号不能超过 254 个字符。")
	}
	if password == "" {
		return badRequest("密码不能为空。")
	}
	if len([]byte(password)) > maxPasswordBytes {
		return badRequest("密码不能超过 72 个字节。")
	}
	return nil
}

func validateOptionalDevice(deviceID string, deviceName string, platform string) error {
	if utf8.RuneCountInString(strings.TrimSpace(deviceID)) > 255 {
		return badRequest("deviceId 不能超过 255 个字符。")
	}
	if utf8.RuneCountInString(strings.TrimSpace(deviceName)) > 255 {
		return badRequest("deviceName 不能超过 255 个字符。")
	}
	if utf8.RuneCountInString(strings.TrimSpace(platform)) > 32 {
		return badRequest("platform 不能超过 32 个字符。")
	}
	return nil
}

func validateContacts(email string, phone string) error {
	if email != "" {
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email || len(email) > 254 {
			return badRequest("邮箱格式不正确。")
		}
	}
	if phone != "" && !phonePattern.MatchString(phone) {
		return badRequest("手机号必须是 6 到 20 位数字，可以以加号开头。")
	}
	return nil
}
