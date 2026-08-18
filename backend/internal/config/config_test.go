package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFileAndEnvironmentOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "addr": "127.0.0.1:9000",
  "database_url": "postgres://file",
  "jwt_secret": "file-secret",
  "admin_username": "root",
  "admin_password": "file-admin-password",
  "token_ttl_hours": 12,
  "smtp_host": "smtp.file.example",
  "smtp_port": 465,
  "smtp_username": "mailer@example.com",
  "smtp_password": "smtp-secret",
  "smtp_from": "noreply@example.com",
  "smtp_from_name": "Tools Mailer",
  "smtp_encryption": "tls",
  "log_level": "debug"
}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AUTOMATIC_TOOLS_CONFIG_FILE", path)
	t.Setenv("DATABASE_URL", "postgres://environment")
	t.Setenv("AUTOMATIC_TOOLS_PORT", "9100")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "0.0.0.0:9100" {
		t.Fatalf("unexpected addr: %s", cfg.Addr)
	}
	if cfg.DatabaseURL != "postgres://environment" {
		t.Fatalf("unexpected database URL: %s", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "file-secret" {
		t.Fatal("JWT secret was not loaded")
	}
	if cfg.AdminUsername != "root" || cfg.AdminPassword != "file-admin-password" {
		t.Fatal("admin bootstrap credentials were not loaded")
	}
	if cfg.TokenTTLHours != 12 {
		t.Fatalf("unexpected token TTL: %d", cfg.TokenTTLHours)
	}
	if cfg.SMTPHost != "smtp.file.example" || cfg.SMTPPort != 465 {
		t.Fatalf("unexpected SMTP endpoint: %s:%d", cfg.SMTPHost, cfg.SMTPPort)
	}
	if cfg.SMTPUsername != "mailer@example.com" || cfg.SMTPPassword != "smtp-secret" {
		t.Fatal("SMTP credentials were not loaded")
	}
	if cfg.SMTPFrom != "noreply@example.com" || cfg.SMTPFromName != "Tools Mailer" {
		t.Fatal("SMTP sender identity was not loaded")
	}
	if cfg.SMTPEncryption != "tls" {
		t.Fatalf("unexpected SMTP encryption: %s", cfg.SMTPEncryption)
	}
}

func TestLoadRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"addr":`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AUTOMATIC_TOOLS_CONFIG_FILE", path)

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid JSON error")
	}
}

func TestLoadRejectsInvalidAlipayEnvironment(t *testing.T) {
	t.Setenv("AUTOMATIC_TOOLS_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.json"))
	t.Setenv("AUTOMATIC_TOOLS_ALIPAY_ENABLED", "tru")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid Alipay boolean error")
	}
}
