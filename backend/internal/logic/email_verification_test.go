package logic

import (
	"regexp"
	"testing"

	"automatictools/backend/internal/config"
)

func TestVerificationCodeGeneration(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9]{6}$`)
	for range 20 {
		code, err := newVerificationCode()
		if err != nil {
			t.Fatal(err)
		}
		if !pattern.MatchString(code) {
			t.Fatalf("unexpected verification code: %q", code)
		}
	}
}

func TestVerificationCodeHashUsesEmailAndCode(t *testing.T) {
	service := New(Dependencies{Config: config.Config{JWTSecret: "test-secret"}})
	first := service.hashVerificationCode("user@example.com", "123456")
	if first != service.hashVerificationCode("user@example.com", "123456") {
		t.Fatal("same email and code should produce the same hash")
	}
	if first == service.hashVerificationCode("other@example.com", "123456") {
		t.Fatal("different emails should produce different hashes")
	}
	if first == service.hashVerificationCode("user@example.com", "654321") {
		t.Fatal("different codes should produce different hashes")
	}
}

func TestMaskEmail(t *testing.T) {
	if got := maskEmail("user@example.com"); got != "u***@example.com" {
		t.Fatalf("unexpected masked email: %q", got)
	}
}
