package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Address   string
	JWTSecret string
	TokenTTL  time.Duration
	Database  DatabaseConfig
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

func Load() Config {
	return Config{
		Address:   env("APP_ADDRESS", ""),
		JWTSecret: env("JWT_SECRET", ""),
		TokenTTL:  durationEnv("JWT_TTL", 15*time.Minute),
		Database: DatabaseConfig{
			URL:      env("DB_URL", ""),
			Host:     env("DB_HOST", "localhost"),
			Port:     intEnv("DB_PORT", 5432),
			User:     env("DB_USER", ""),
			Password: env("DB_PASSWORD", ""),
			Name:     env("DB_NAME", ""),
		},
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func intEnv(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	port, err := strconv.Atoi(value)
	if err != nil || port <= 0 {
		return fallback
	}
	return port
}
