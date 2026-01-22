package config

import (
	"os"
	"strconv"
)

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

	return &Config{
		Port:          getEnv("PORT", "8080"),
		Env:           getEnv("ENV", "development"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/tictactoe?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		SessionSecret: getEnv("SESSION_SECRET", "dev-secret-change-in-production"),
		SessionMaxAge: sessionMaxAge,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
