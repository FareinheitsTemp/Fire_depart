package config

import "os"

// Config — налаштування застосунку.
type Config struct {
	HTTPAddr    string
	DatabaseDSN string
}

// Load зчитує конфігурацію зі змінних оточення з дефолтами для локальної розробки.
func Load() Config {
	return Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseDSN: getenv("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/fire_depart?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
