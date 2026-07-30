package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Addr          string
	DatabaseURL   string
	JWTSecret     string
	AdminKey      string
	TokenTTLHours int
	LogLevel      slog.Level
}

type fileConfig struct {
	Addr          string `json:"addr"`
	DatabaseURL   string `json:"database_url"`
	JWTSecret     string `json:"jwt_secret"`
	AdminKey      string `json:"admin_key"`
	TokenTTLHours int    `json:"token_ttl_hours"`
	LogLevel      string `json:"log_level"`
}

func Load() (Config, error) {
	values := fileConfig{
		Addr:          "0.0.0.0:8088",
		DatabaseURL:   "postgres://postgres@localhost:5432/automatic_tools?sslmode=disable",
		JWTSecret:     "dev-secret-change-me",
		AdminKey:      "dev-admin-key-change-me",
		TokenTTLHours: 24 * 30,
		LogLevel:      "info",
	}

	configPath := env("AUTOMATIC_TOOLS_CONFIG_FILE", "config.json")
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := json.Unmarshal(data, &values); err != nil {
			return Config{}, fmt.Errorf("parse %s: %w", configPath, err)
		}
	} else if !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("read %s: %w", configPath, err)
	}

	if port := os.Getenv("AUTOMATIC_TOOLS_PORT"); port != "" && os.Getenv("AUTOMATIC_TOOLS_ADDR") == "" {
		values.Addr = "0.0.0.0:" + port
	}
	values.Addr = env("AUTOMATIC_TOOLS_ADDR", values.Addr)
	values.DatabaseURL = env("DATABASE_URL", values.DatabaseURL)
	values.JWTSecret = env("AUTOMATIC_TOOLS_JWT_SECRET", values.JWTSecret)
	values.AdminKey = env("AUTOMATIC_TOOLS_ADMIN_KEY", values.AdminKey)
	values.TokenTTLHours = envInt("AUTOMATIC_TOOLS_TOKEN_TTL_HOURS", values.TokenTTLHours)
	values.LogLevel = env("AUTOMATIC_TOOLS_LOG_LEVEL", values.LogLevel)

	var level slog.Level
	if err := level.UnmarshalText([]byte(values.LogLevel)); err != nil {
		return Config{}, fmt.Errorf("invalid log_level %q: %w", values.LogLevel, err)
	}

	return Config{
		Addr:          values.Addr,
		DatabaseURL:   values.DatabaseURL,
		JWTSecret:     values.JWTSecret,
		AdminKey:      values.AdminKey,
		TokenTTLHours: values.TokenTTLHours,
		LogLevel:      level,
	}, nil
}

func env(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
