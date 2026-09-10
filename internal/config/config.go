package config

import (
	"os"
	"strings"
)

// Config — налаштування застосунку (змінні оточення з дефолтами для локальної розробки).
type Config struct {
	HTTPAddr       string
	DatabaseDSN    string
	AllowedOrigins []string
}

// Load зчитує конфігурацію.
func Load() Config {
	return Config{
		HTTPAddr:       getenv("HTTP_ADDR", ":8080"),
		DatabaseDSN:    getenv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/fire_depart?sslmode=disable"),
		AllowedOrigins: strings.Split(getenv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
