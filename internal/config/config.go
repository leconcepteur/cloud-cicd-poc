package config

import (
	"log"
	"os"
	"strconv"
)

const defaultSessionSecret = "dev-secret-change-in-production"

type Config struct {
	Port          string
	Env           string
	DatabaseURL   string
	RedisURL      string
	SessionSecret string
	SessionMaxAge int
}

func Load() *Config {
	sessionMaxAge := 30 * 24 * 60 * 60 // 30 days in seconds
	if val := os.Getenv("SESSION_MAX_AGE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			sessionMaxAge = parsed
		}
	}

	env := getEnv("ENV", "development")
	sessionSecret := getEnv("SESSION_SECRET", defaultSessionSecret)

	// Warn if using default session secret in non-development environments
	if env != "development" && sessionSecret == defaultSessionSecret {
		log.Printf("WARNING: Using default SESSION_SECRET in %s environment. Please set a secure SESSION_SECRET environment variable!", env)
	}

	return &Config{
		Port:          getEnv("PORT", "8080"),
		Env:           env,
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tictactoe?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		SessionSecret: sessionSecret,
		SessionMaxAge: sessionMaxAge,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
