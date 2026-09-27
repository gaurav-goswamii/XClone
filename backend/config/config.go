package config

import (
	"os"
)

// Config holds all backend application configuration
type Config struct {
	Port        string
	Environment string
	DatabaseURL string
	JWTSecret   string
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/xclone?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "xclone-super-secret-key"
	}

	return &Config{
		Port:        port,
		Environment: env,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
	}
}
