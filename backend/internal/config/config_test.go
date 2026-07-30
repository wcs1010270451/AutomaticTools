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
  "admin_key": "file-admin",
  "token_ttl_hours": 12,
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
	if cfg.JWTSecret != "file-secret" || cfg.AdminKey != "file-admin" {
		t.Fatal("file secrets were not loaded")
	}
	if cfg.TokenTTLHours != 12 {
		t.Fatalf("unexpected token TTL: %d", cfg.TokenTTLHours)
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
