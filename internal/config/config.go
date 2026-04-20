package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL       string
	JWTSecret         string
	JWTExpiry         time.Duration
	Port              string
	CORSOrigins       string
	ClientOrigin      string
	IPaymuVA          string
	IPaymuAPIKey      string
	IPaymuBaseURL     string
	IPaymuCallbackURL string
}

func Load() *Config {
	expiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		expiry = 24 * time.Hour
	}

	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://pemilo:pemilo_secret@localhost:5432/pemilo?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-to-a-strong-secret-key"),
		JWTExpiry:         expiry,
		Port:              getEnv("PORT", "8080"),
		CORSOrigins:       getEnv("CORS_ORIGINS", "http://localhost:3000"),
		ClientOrigin:      getEnv("CLIENT_ORIGINS", "http://localhost:3000"),
		IPaymuVA:          getEnv("IPAYMU_VA", ""),
		IPaymuAPIKey:      getEnv("IPAYMU_API_KEY", ""),
		IPaymuBaseURL:     getEnv("IPAYMU_BASE_URL", "https://sandbox.ipaymu.com/api/v2"),
		IPaymuCallbackURL: getEnv("IPAYMU_CALLBACK_URL", "http://localhost:8080/api/payments/ipaymu/webhook"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
