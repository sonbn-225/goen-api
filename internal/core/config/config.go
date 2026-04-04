package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env                 string
	Host                string
	Port                int
	LogLevel            slog.Level
	MigrateOnStart      bool
	MigrationDir        string
	DatabaseURL         string
	RedisURL            string
	JWTSecret           string
	JWTAccessTTLMinutes int
	S3Endpoint          string
	S3AccessKey         string
	S3SecretKey         string
	S3Bucket            string
	S3UseSSL            bool
	S3PublicBaseURL     string
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Env = getenvDefault("GOEN_ENV", "development")
	cfg.Host = getenvDefault("HOST", "0.0.0.0")
	cfg.Port = getenvIntDefault("PORT", 8080)
	cfg.MigrateOnStart = getenvBoolDefault("MIGRATE_ON_START", true)
	cfg.MigrationDir = getenvDefault("MIGRATION_DIR", "migrations")
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.RedisURL = os.Getenv("REDIS_URL")
	cfg.JWTAccessTTLMinutes = getenvIntDefault("JWT_ACCESS_TTL_MINUTES", 60)
	cfg.S3Endpoint = os.Getenv("SEAWEEDFS_ENDPOINT")
	cfg.S3AccessKey = os.Getenv("SEAWEEDFS_ACCESS_KEY_ID")
	cfg.S3SecretKey = os.Getenv("SEAWEEDFS_SECRET_ACCESS_KEY")
	cfg.S3Bucket = getenvDefault("SEAWEEDFS_BUCKET", "goen")
	cfg.S3UseSSL = os.Getenv("SEAWEEDFS_USE_SSL") == "true"
	cfg.S3PublicBaseURL = os.Getenv("SEAWEEDFS_PUBLIC_BASE_URL")

	levelStr := strings.ToLower(getenvDefault("LOG_LEVEL", "info"))
	switch levelStr {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn", "warning":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		return nil, fmt.Errorf("invalid LOG_LEVEL: %q", levelStr)
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if cfg.JWTSecret == "" {
		if cfg.Env == "development" {
			cfg.JWTSecret = "dev-insecure-secret"
		} else {
			return nil, errors.New("JWT_SECRET is required in non-development environments")
		}
	}

	if cfg.JWTAccessTTLMinutes < 1 {
		return nil, errors.New("JWT_ACCESS_TTL_MINUTES must be >= 1")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return nil, errors.New("PORT must be between 1 and 65535")
	}

	return cfg, nil
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvIntDefault(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getenvBoolDefault(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
