package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           int
	DatabaseURL    string
	JWTSecret      string
	RequestTimeout time.Duration
	LogLevel       slog.Level
}

func Load() Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			port = parsed
		}
	}

	timeout := 5 * time.Second
	if v := os.Getenv("REQUEST_TIMEOUT_SECONDS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			timeout = time.Duration(parsed) * time.Second
		}
	}

	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}

	return Config{
		Port:           port,
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		JWTSecret:      secret,
		RequestTimeout: timeout,
		LogLevel:       level,
	}
}
