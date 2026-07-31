package logic

import (
	"context"
	"errors"
	"strings"

	"automatictools/backend/internal/store"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminAuthResponse struct {
	Token string   `json:"token"`
	Admin AdminDTO `json:"admin"`
}

func (a *Service) AdminLogin(ctx context.Context, req AdminLoginRequest, meta RequestMeta) (AdminAuthResponse, error) {
	username := normalizeUsername(req.Username)

	var admin store.Admin
	err := a.db.WithContext(ctx).
		Where("username = ?", username).
		Take(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AdminAuthResponse{}, unauthorized("管理员用户名或密码错误。")
		}
		return AdminAuthResponse{}, err
	}

	if admin.Status != "active" {
		return AdminAuthResponse{}, forbidden("管理员账号不可用。")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return AdminAuthResponse{}, unauthorized("管理员用户名或密码错误。")
	}

	now := unixNow()
	err = a.db.WithContext(ctx).
		Model(&store.Admin{}).
		Where("id = ?", admin.ID).
		Updates(map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		}).Error
	if err != nil {
		return AdminAuthResponse{}, err
	}
	admin.LastLoginAt = &now
	admin.UpdatedAt = now

	_ = a.audit(ctx, meta, nil, "admin.login", `{"adminId":`+fmtSprint(admin.ID)+`}`)

	token, err := a.issueToken(admin.ID, admin.Username, "admin")
	if err != nil {
		return AdminAuthResponse{}, err
	}

	return AdminAuthResponse{
		Token: token,
		Admin: AdminDTO{
			ID:          admin.ID,
			Username:    admin.Username,
			Status:      admin.Status,
			CreatedAt:   admin.CreatedAt,
			LastLoginAt: admin.LastLoginAt,
		},
	}, nil
}

func (a *Service) AuthenticateAdmin(ctx context.Context, header string) (int64, string, error) {
	if !strings.HasPrefix(header, "Bearer ") {
		return 0, "", unauthorized("缺少管理员登录令牌。")
	}

	tokenText := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (any, error) {
		return []byte(a.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid || claims.Subject == "" || claims.TokenType != "admin" {
		return 0, "", unauthorized("管理员登录令牌无效或已过期。")
	}

	var adminID int64
	if _, err := fmtSscan(claims.Subject, &adminID); err != nil {
		return 0, "", unauthorized("管理员登录令牌无效。")
	}

	var admin store.Admin
	err = a.db.WithContext(ctx).
		Select("id", "username", "status").
		Where("id = ? AND status = ?", adminID, "active").
		Take(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, "", unauthorized("管理员账号不存在或已停用。")
		}
		return 0, "", err
	}
	return admin.ID, admin.Username, nil
}
