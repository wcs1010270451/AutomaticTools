package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"automatictools/backend/internal/config"
	"automatictools/backend/internal/store"

	"gorm.io/gorm"
)

type capturedEmailSender struct {
	to           string
	code         string
	validMinutes int
}

func (s *capturedEmailSender) SendRegistrationCode(to string, code string, validMinutes int) error {
	s.to = to
	s.code = code
	s.validMinutes = validMinutes
	return nil
}

func TestEmailCodeRegistrationPostgres(t *testing.T) {
	databaseURL := os.Getenv("AUTOMATIC_TOOLS_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set AUTOMATIC_TOOLS_TEST_DATABASE_URL to run the PostgreSQL integration test")
	}

	db, err := store.Open(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatal(err)
	}

	stamp := time.Now().UnixNano()
	email := fmt.Sprintf("integration@auth-%d.example.com", stamp)
	sender := &capturedEmailSender{}
	service := New(Dependencies{
		Config: config.Config{
			JWTSecret:     "integration-test-secret",
			TokenTTLHours: 1,
		},
		DB:          db,
		EmailSender: sender,
	})
	t.Cleanup(func() { cleanupRegistrationIntegrationData(db, email) })

	response, err := service.SendRegistrationEmailCode(
		context.Background(),
		SendRegistrationEmailCodeRequest{Email: email},
		RequestMeta{IP: "127.0.0.1", UserAgent: "integration-test"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !response.OK || response.ExpiresIn != 600 || response.ResendAfter != 60 {
		t.Fatalf("unexpected send-code response: %+v", response)
	}
	if sender.to != email || !regexp.MustCompile(`^[0-9]{6}$`).MatchString(sender.code) || sender.validMinutes != 10 {
		t.Fatalf("unexpected captured email: %+v", sender)
	}

	_, err = service.SendRegistrationEmailCode(
		context.Background(),
		SendRegistrationEmailCodeRequest{Email: email},
		RequestMeta{},
	)
	requireLogicErrorCode(t, err, ErrorCodeConflict)

	wrongCode := "000000"
	if wrongCode == sender.code {
		wrongCode = "999999"
	}
	_, err = service.Register(context.Background(), RegisterRequest{
		Email:     email,
		EmailCode: wrongCode,
		Password:  "password123",
	}, RequestMeta{})
	requireLogicErrorCode(t, err, ErrorCodeBadRequest)

	registered, err := service.Register(context.Background(), RegisterRequest{
		Email:     email,
		EmailCode: sender.code,
		Password:  "password123",
	}, RequestMeta{IP: "127.0.0.1", UserAgent: "integration-test"})
	if err != nil {
		t.Fatal(err)
	}
	if registered.User.Email != email || !regexp.MustCompile(`^user_[0-9a-f]{16}$`).MatchString(registered.User.Username) {
		t.Fatalf("unexpected registered user: %+v", registered.User)
	}
	if registered.Token == "" {
		t.Fatal("registration response did not contain a token")
	}

	userID, username, err := service.Authenticate(context.Background(), "Bearer "+registered.Token)
	if err != nil {
		t.Fatal(err)
	}
	if userID != registered.User.ID || username != registered.User.Username {
		t.Fatalf("unexpected authenticated identity: %d %q", userID, username)
	}

	loggedIn, err := service.Login(context.Background(), LoginRequest{
		Account:  email,
		Password: "password123",
	}, RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if loggedIn.User.ID != registered.User.ID || loggedIn.Token == "" {
		t.Fatalf("unexpected login response: %+v", loggedIn)
	}

	_, err = service.Register(context.Background(), RegisterRequest{
		Email:     email,
		EmailCode: sender.code,
		Password:  "password123",
	}, RequestMeta{})
	requireLogicErrorCode(t, err, ErrorCodeBadRequest)

	var verification store.EmailVerificationCode
	if err := db.Where("email = ?", email).Order("id DESC").Take(&verification).Error; err != nil {
		t.Fatal(err)
	}
	if verification.ConsumedAt == nil || verification.AttemptCount != 1 {
		t.Fatalf("verification code state was not persisted correctly: %+v", verification)
	}
}

func requireLogicErrorCode(t *testing.T, err error, expected ErrorCode) {
	t.Helper()
	var appErr Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected logic error %q, got %v", expected, err)
	}
	if appErr.Code != expected {
		t.Fatalf("expected logic error %q, got %q", expected, appErr.Code)
	}
}

func cleanupRegistrationIntegrationData(db *gorm.DB, email string) {
	var userIDs []int64
	_ = db.Model(&store.User{}).Where("LOWER(email) = ?", email).Pluck("id", &userIDs).Error
	if len(userIDs) > 0 {
		_ = db.Where("user_id IN ?", userIDs).Delete(&store.AuditLog{}).Error
	}
	_ = db.Where("LOWER(email) = ?", email).Delete(&store.User{}).Error
	_ = db.Where("email = ?", email).Delete(&store.EmailVerificationCode{}).Error
	detail, _ := json.Marshal(map[string]string{"email": maskEmail(email)})
	_ = db.Where("action = ? AND detail = ?", "email.registration_code.send", string(detail)).Delete(&store.AuditLog{}).Error
}
