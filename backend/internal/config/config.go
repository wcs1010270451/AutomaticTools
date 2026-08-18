package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type Config struct {
	Addr                 string
	DatabaseURL          string
	JWTSecret            string
	AdminUsername        string
	AdminPassword        string
	TokenTTLHours        int
	SMTPHost             string
	SMTPPort             int
	SMTPUsername         string
	SMTPPassword         string
	SMTPFrom             string
	SMTPFromName         string
	SMTPEncryption       string
	AlipayEnabled        bool
	AlipayAppID          string
	AlipayPrivateKeyFile string
	AlipayPublicKeyFile  string
	AlipayNotifyURL      string
	AlipaySellerID       string
	AlipayProduction     bool
	AlipayTimeoutSeconds int
	LogLevel             slog.Level
}

type fileConfig struct {
	Addr                 string `json:"addr"`
	DatabaseURL          string `json:"database_url"`
	JWTSecret            string `json:"jwt_secret"`
	AdminUsername        string `json:"admin_username"`
	AdminPassword        string `json:"admin_password"`
	TokenTTLHours        int    `json:"token_ttl_hours"`
	SMTPHost             string `json:"smtp_host"`
	SMTPPort             int    `json:"smtp_port"`
	SMTPUsername         string `json:"smtp_username"`
	SMTPPassword         string `json:"smtp_password"`
	SMTPFrom             string `json:"smtp_from"`
	SMTPFromName         string `json:"smtp_from_name"`
	SMTPEncryption       string `json:"smtp_encryption"`
	AlipayEnabled        bool   `json:"alipay_enabled"`
	AlipayAppID          string `json:"alipay_app_id"`
	AlipayPrivateKeyFile string `json:"alipay_private_key_file"`
	AlipayPublicKeyFile  string `json:"alipay_public_key_file"`
	AlipayNotifyURL      string `json:"alipay_notify_url"`
	AlipaySellerID       string `json:"alipay_seller_id"`
	AlipayProduction     bool   `json:"alipay_production"`
	AlipayTimeoutSeconds int    `json:"alipay_timeout_seconds"`
	LogLevel             string `json:"log_level"`
}

func Load() (Config, error) {
	values := fileConfig{
		Addr:                 "0.0.0.0:8088",
		DatabaseURL:          "postgres://postgres@localhost:5432/automatic_tools?sslmode=disable",
		JWTSecret:            "dev-secret-change-me",
		AdminUsername:        "admin",
		AdminPassword:        "123456",
		TokenTTLHours:        24 * 30,
		SMTPPort:             465,
		SMTPFromName:         "AutomaticTools",
		SMTPEncryption:       "starttls",
		AlipayTimeoutSeconds: 15,
		LogLevel:             "info",
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
	values.AdminUsername = env("AUTOMATIC_TOOLS_ADMIN_USERNAME", values.AdminUsername)
	values.AdminPassword = env("AUTOMATIC_TOOLS_ADMIN_PASSWORD", values.AdminPassword)
	values.TokenTTLHours = envInt("AUTOMATIC_TOOLS_TOKEN_TTL_HOURS", values.TokenTTLHours)
	values.SMTPHost = env("AUTOMATIC_TOOLS_SMTP_HOST", values.SMTPHost)
	values.SMTPPort = envInt("AUTOMATIC_TOOLS_SMTP_PORT", values.SMTPPort)
	values.SMTPUsername = env("AUTOMATIC_TOOLS_SMTP_USERNAME", values.SMTPUsername)
	values.SMTPPassword = env("AUTOMATIC_TOOLS_SMTP_PASSWORD", values.SMTPPassword)
	values.SMTPFrom = env("AUTOMATIC_TOOLS_SMTP_FROM", values.SMTPFrom)
	values.SMTPFromName = env("AUTOMATIC_TOOLS_SMTP_FROM_NAME", values.SMTPFromName)
	values.SMTPEncryption = env("AUTOMATIC_TOOLS_SMTP_ENCRYPTION", values.SMTPEncryption)
	if value := os.Getenv("AUTOMATIC_TOOLS_ALIPAY_ENABLED"); value != "" {
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return Config{}, fmt.Errorf("invalid AUTOMATIC_TOOLS_ALIPAY_ENABLED: %w", parseErr)
		}
		values.AlipayEnabled = parsed
	}
	values.AlipayAppID = env("AUTOMATIC_TOOLS_ALIPAY_APP_ID", values.AlipayAppID)
	values.AlipayPrivateKeyFile = env("AUTOMATIC_TOOLS_ALIPAY_PRIVATE_KEY_FILE", values.AlipayPrivateKeyFile)
	values.AlipayPublicKeyFile = env("AUTOMATIC_TOOLS_ALIPAY_PUBLIC_KEY_FILE", values.AlipayPublicKeyFile)
	values.AlipayNotifyURL = env("AUTOMATIC_TOOLS_ALIPAY_NOTIFY_URL", values.AlipayNotifyURL)
	values.AlipaySellerID = env("AUTOMATIC_TOOLS_ALIPAY_SELLER_ID", values.AlipaySellerID)
	if value := os.Getenv("AUTOMATIC_TOOLS_ALIPAY_PRODUCTION"); value != "" {
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return Config{}, fmt.Errorf("invalid AUTOMATIC_TOOLS_ALIPAY_PRODUCTION: %w", parseErr)
		}
		values.AlipayProduction = parsed
	}
	if value := os.Getenv("AUTOMATIC_TOOLS_ALIPAY_TIMEOUT_SECONDS"); value != "" {
		parsed, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return Config{}, fmt.Errorf("invalid AUTOMATIC_TOOLS_ALIPAY_TIMEOUT_SECONDS: %w", parseErr)
		}
		values.AlipayTimeoutSeconds = parsed
	}
	values.LogLevel = env("AUTOMATIC_TOOLS_LOG_LEVEL", values.LogLevel)
	if values.AlipayTimeoutSeconds <= 0 {
		return Config{}, fmt.Errorf("alipay_timeout_seconds must be greater than zero")
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(values.LogLevel)); err != nil {
		return Config{}, fmt.Errorf("invalid log_level %q: %w", values.LogLevel, err)
	}

	return Config{
		Addr:                 values.Addr,
		DatabaseURL:          values.DatabaseURL,
		JWTSecret:            values.JWTSecret,
		AdminUsername:        values.AdminUsername,
		AdminPassword:        values.AdminPassword,
		TokenTTLHours:        values.TokenTTLHours,
		SMTPHost:             values.SMTPHost,
		SMTPPort:             values.SMTPPort,
		SMTPUsername:         values.SMTPUsername,
		SMTPPassword:         values.SMTPPassword,
		SMTPFrom:             values.SMTPFrom,
		SMTPFromName:         values.SMTPFromName,
		SMTPEncryption:       values.SMTPEncryption,
		AlipayEnabled:        values.AlipayEnabled,
		AlipayAppID:          values.AlipayAppID,
		AlipayPrivateKeyFile: values.AlipayPrivateKeyFile,
		AlipayPublicKeyFile:  values.AlipayPublicKeyFile,
		AlipayNotifyURL:      values.AlipayNotifyURL,
		AlipaySellerID:       values.AlipaySellerID,
		AlipayProduction:     values.AlipayProduction,
		AlipayTimeoutSeconds: values.AlipayTimeoutSeconds,
		LogLevel:             level,
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
