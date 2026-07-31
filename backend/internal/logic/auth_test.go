package logic

import (
	"context"
	"strings"
	"testing"

	"automatictools/backend/internal/config"
)

func TestNormalizeContacts(t *testing.T) {
	if got := normalizeEmail("  User@Example.COM "); got != "user@example.com" {
		t.Fatalf("unexpected normalized email: %q", got)
	}
	if got := normalizePhone(" +86 138-0013-8000 "); got != "+8613800138000" {
		t.Fatalf("unexpected normalized phone: %q", got)
	}
}

func TestValidateContacts(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		phone   string
		wantErr bool
	}{
		{name: "empty contacts are allowed"},
		{name: "valid contacts", email: "user@example.com", phone: "+8613800138000"},
		{name: "invalid email", email: "not-an-email", wantErr: true},
		{name: "phone too short", phone: "12345", wantErr: true},
		{name: "phone contains letters", phone: "13800abc000", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateContacts(tt.email, tt.phone)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateContacts() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAuthPasswordLength(t *testing.T) {
	if err := validateAuth("test_user", "123456"); err != nil {
		t.Fatalf("valid auth request rejected: %v", err)
	}
	if err := validateAuth("test_user", "密码太短"); err == nil {
		t.Fatal("password shorter than six characters should be rejected")
	}
	if err := validateAuth("test_user", strings.Repeat("a", maxPasswordBytes+1)); err == nil {
		t.Fatal("password longer than bcrypt limit should be rejected")
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name     string
		account  string
		password string
		wantErr  bool
	}{
		{name: "valid username", account: "test_user", password: "123456"},
		{name: "valid email", account: "user@example.com", password: "123456"},
		{name: "valid phone", account: "+8613800138000", password: "123456"},
		{name: "missing account", password: "123456", wantErr: true},
		{name: "missing password", account: "test_user", wantErr: true},
		{name: "account too long", account: strings.Repeat("a", 255), password: "123456", wantErr: true},
		{
			name:     "password too long",
			account:  "test_user",
			password: strings.Repeat("a", maxPasswordBytes+1),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogin(tt.account, tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOptionalDevice(t *testing.T) {
	tests := []struct {
		name       string
		deviceID   string
		deviceName string
		platform   string
		wantErr    bool
	}{
		{name: "empty device is allowed"},
		{name: "valid device", deviceID: "device-1", deviceName: "Windows PC", platform: "windows"},
		{name: "device id too long", deviceID: strings.Repeat("a", 256), wantErr: true},
		{name: "device name too long", deviceName: strings.Repeat("a", 256), wantErr: true},
		{name: "platform too long", platform: strings.Repeat("a", 33), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptionalDevice(tt.deviceID, tt.deviceName, tt.platform)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateOptionalDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNullableString(t *testing.T) {
	if nullableString("") != nil {
		t.Fatal("empty string should be stored as NULL")
	}
	if got := nullableString("value"); got == nil || *got != "value" {
		t.Fatalf("unexpected non-empty value: %v", got)
	}
}

func TestUserAndAdminTokensAreIsolated(t *testing.T) {
	service := New(Dependencies{
		Config: config.Config{
			JWTSecret:     "test-secret",
			TokenTTLHours: 1,
		},
	})

	userToken, err := service.issueUserToken(1, "user")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.parseUserToken("Bearer " + userToken); err != nil {
		t.Fatalf("user token should parse as user: %v", err)
	}
	if _, _, err := service.AuthenticateAdmin(context.Background(), "Bearer "+userToken); err == nil {
		t.Fatal("user token must not authenticate as administrator")
	}

	adminToken, err := service.issueToken(1, "admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := service.parseUserToken("Bearer " + adminToken); err == nil {
		t.Fatal("administrator token must not authenticate as user")
	}
}
