package app

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]{3,32}$`)
var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

type authRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone,omitempty"`
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	Platform   string `json:"platform,omitempty"`
}

type authResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

type tokenClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		a.handleErr(w, r, badRequest("请求体必须是合法 JSON。"))
		return
	}

	username := normalizeUsername(req.Username)
	email := normalizeEmail(req.Email)
	phone := normalizePhone(req.Phone)
	if err := validateAuth(username, req.Password); err != nil {
		a.handleErr(w, r, err)
		return
	}
	if err := validateContacts(email, phone); err != nil {
		a.handleErr(w, r, err)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	now := unixNow()
	var userID int64
	err = a.db.QueryRowContext(
		r.Context(),
		`INSERT INTO users(username, email, phone, password_hash, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, 'active', $5, $6)
		 RETURNING id`,
		username,
		nullableString(email),
		nullableString(phone),
		string(passwordHash),
		now,
		now,
	).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "uq_users_email":
				a.handleErr(w, r, conflict("邮箱已被注册。"))
			case "uq_users_phone":
				a.handleErr(w, r, conflict("手机号已被注册。"))
			default:
				a.handleErr(w, r, conflict("用户名已存在。"))
			}
			return
		}
		a.handleErr(w, r, err)
		return
	}

	_ = a.audit(r, &userID, "user.register", "")
	if req.DeviceID != "" {
		_ = a.upsertDevice(r, userID, req.DeviceID, req.DeviceName, req.Platform)
	}

	token, err := a.issueToken(userID, username)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User: UserDTO{
			ID:       userID,
			Username: username,
			Email:    email,
			Phone:    phone,
		},
	})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		a.handleErr(w, r, badRequest("请求体必须是合法 JSON。"))
		return
	}

	username := normalizeUsername(req.Username)
	row := a.db.QueryRowContext(
		r.Context(),
		`SELECT id, username, email, phone, password_hash, status FROM users WHERE username = $1`,
		username,
	)

	var userID int64
	var storedUsername string
	var email sql.NullString
	var phone sql.NullString
	var passwordHash string
	var status string
	if err := row.Scan(&userID, &storedUsername, &email, &phone, &passwordHash, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.handleErr(w, r, unauthorized("用户名或密码错误。"))
			return
		}
		a.handleErr(w, r, err)
		return
	}

	if status != "active" {
		a.handleErr(w, r, forbidden("账号不可用。"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		a.handleErr(w, r, unauthorized("用户名或密码错误。"))
		return
	}

	now := unixNow()
	if _, err := a.db.ExecContext(r.Context(), `UPDATE users SET last_login_at = $1, updated_at = $2 WHERE id = $3`, now, now, userID); err != nil {
		a.handleErr(w, r, err)
		return
	}

	_ = a.audit(r, &userID, "user.login", "")
	if req.DeviceID != "" {
		_ = a.upsertDevice(r, userID, req.DeviceID, req.DeviceName, req.Platform)
	}

	token, err := a.issueToken(userID, storedUsername)
	if err != nil {
		a.handleErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token: token,
		User: UserDTO{
			ID:       userID,
			Username: storedUsername,
			Email:    email.String,
			Phone:    phone.String,
			Status:   status,
		},
	})
}

func (a *App) currentUser(r *http.Request) (int64, string, error) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return 0, "", unauthorized("缺少登录令牌。")
	}
	tokenText := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(tokenText, claims, func(token *jwt.Token) (any, error) {
		return []byte(a.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, "", unauthorized("登录令牌无效或已过期。")
	}

	if claims.Subject == "" {
		return 0, "", unauthorized("登录令牌无效。")
	}

	var userID int64
	if _, err := fmtSscan(claims.Subject, &userID); err != nil {
		return 0, "", unauthorized("登录令牌无效。")
	}
	return userID, claims.Username, nil
}

func (a *App) issueToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmtSprint(userID),
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

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func validateAuth(username string, password string) error {
	if !usernamePattern.MatchString(username) {
		return badRequest("用户名必须是 3 到 32 位，只能包含字母、数字、下划线、横线或点。")
	}
	if len(password) < 6 {
		return badRequest("密码至少需要 6 个字符。")
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
