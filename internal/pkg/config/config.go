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
	Env            string
	Host           string
	Port           int
	LogLevel       slog.Level
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	JWTAccessTTL   int
	S3Endpoint     string
	S3AccessKey    string
	S3SecretKey    string
	S3AccessKeySource string
	S3Bucket       string
	S3UseSSL       bool
	PublicBaseURL  string
	MigrationDir        string
	MigrateOnStart      bool
	MarketDataStatusURL string
	CORSOrigins         []string
}

func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Env = getenvDefault("GOEN_ENV", "development")
	cfg.Host = getenvDefault("HOST", "0.0.0.0")
	cfg.Port = getenvIntDefault("PORT", 8080)
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	
	// Redis: support unified URL or individual parts
	cfg.RedisURL = os.Getenv("REDIS_URL")
	redisHost := os.Getenv("REDIS_HOST")
	// FORCE REBUILD if URL is missing OR looks mangled (typical production interpolation error)
	if redisHost != "" || cfg.RedisURL == "" || strings.Contains(cfg.RedisURL, "default:s") {
		host := getenvDefault("REDIS_HOST", "sunflower-redis")
		port := getenvIntDefault("REDIS_PORT", 6379)
		pass := os.Getenv("REDIS_PASSWORD")
		if pass != "" {
			cfg.RedisURL = fmt.Sprintf("redis://default:%s@%s:%d/0", pass, host, port)
		} else {
			cfg.RedisURL = fmt.Sprintf("redis://%s:%d/0", host, port)
		}
	}

	cfg.JWTSecret = getenvDefault("JWT_SECRET", "dev-secret")
	cfg.JWTAccessTTL = getenvIntDefault("JWT_ACCESS_TTL_MINUTES", 60)

	// SeaweedFS: support unified endpoint or individual parts
	cfg.S3Endpoint = os.Getenv("SEAWEEDFS_ENDPOINT")
	if os.Getenv("SEAWEEDFS_HOST") != "" || cfg.S3Endpoint == "" {
		host := getenvDefault("SEAWEEDFS_HOST", "sunflower-seaweedfs")
		port := getenvIntDefault("SEAWEEDFS_PORT", 8333)
		cfg.S3Endpoint = fmt.Sprintf("%s:%d", host, port)
	}

	// Try multiple environment variable names for keys to be robust
	cfg.S3AccessKey, cfg.S3AccessKeySource = getenvFallbackWithSource("SEAWEEDFS_ACCESS_KEY_ID", "S3_ACCESS_KEY_ID", "SEAWEEDFS_ACCESS_KEY")
	cfg.S3SecretKey, _ = getenvFallbackWithSource("SEAWEEDFS_SECRET_ACCESS_KEY", "S3_SECRET_ACCESS_KEY", "SEAWEEDFS_SECRET_KEY")
	
	cfg.S3Bucket = getenvDefault("SEAWEEDFS_BUCKET", "goen")
	cfg.S3UseSSL = getenvBoolDefault("SEAWEEDFS_USE_SSL", false)
	cfg.PublicBaseURL = os.Getenv("PUBLIC_BASE_URL")
	cfg.MigrationDir = getenvDefault("MIGRATION_DIR", "scripts")
	cfg.MigrateOnStart = getenvBoolDefault("MIGRATE_ON_START", true)
	cfg.MarketDataStatusURL = os.Getenv("MARKET_DATA_STATUS_URL")

	corsStr := getenvDefault("CORS_ORIGINS", "*")
	if corsStr == "*" {
		cfg.CORSOrigins = []string{"*"}
	} else {
		cfg.CORSOrigins = strings.Split(corsStr, ",")
	}

	levelStr := strings.ToLower(getenvDefault("LOG_LEVEL", "info"))
	switch levelStr {
	case "debug":
		cfg.LogLevel = slog.LevelDebug
	case "info":
		cfg.LogLevel = slog.LevelInfo
	case "warn":
		cfg.LogLevel = slog.LevelWarn
	case "error":
		cfg.LogLevel = slog.LevelError
	default:
		cfg.LogLevel = slog.LevelInfo
	}

	if cfg.DatabaseURL == "" && cfg.Env == "production" {
		return nil, errors.New("DATABASE_URL is required in production")
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
	v := strings.ToLower(os.Getenv(key))
	if v == "" {
		return def
	}
	return v == "true" || v == "1"
}

func getenvFallback(keys ...string) string {
	v, _ := getenvFallbackWithSource(keys...)
	return v
}

func getenvFallbackWithSource(keys ...string) (string, string) {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v, k
		}
	}
	return "", "NONE"
}
