package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort          string
	DatabaseURL         string
	JWTSecret           string
	EncryptionKey       string
	EncryptionSaltFront string
	EncryptionSaltBack  string
	Environment         string
	AllowedOrigins      string
	OwnerUsername       string
	OwnerPassword       string
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		JWTSecret:           getEnv("JWT_SECRET", "change-this-secret-in-production"),
		EncryptionKey:       getEnv("ENCRYPTION_KEY", ""),
		EncryptionSaltFront: getEnv("ENCRYPTION_SALT_FRONT", ""),
		EncryptionSaltBack:  getEnv("ENCRYPTION_SALT_BACK", ""),
		Environment:         getEnv("ENVIRONMENT", "development"),
		AllowedOrigins:      getEnv("ALLOWED_ORIGINS", "*"),
		OwnerUsername:       getEnv("OWNER_USERNAME", "owner"),
		OwnerPassword:       getEnv("OWNER_PASSWORD", "change-this-password"),
	}

	// Validate required fields
	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if config.EncryptionKey == "" || len(config.EncryptionKey) != 32 {
		return nil, fmt.Errorf("ENCRYPTION_KEY must be exactly 32 characters for AES-256")
	}
	if config.EncryptionSaltFront == "" {
		return nil, fmt.Errorf("ENCRYPTION_SALT_FRONT is required")
	}
	if config.EncryptionSaltBack == "" {
		return nil, fmt.Errorf("ENCRYPTION_SALT_BACK is required")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
