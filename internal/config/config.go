package config

import (
	"os"
	"time"
)

type HTTP struct {
	Address      string
	BaseURL      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DB struct {
	DSN string
}

type Log struct {
	Level string
}

type Config struct {
	HTTP HTTP
	DB   DB
	Log  Log
}

func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTP{
			Address:      getEnv("HTTP_ADDRESS", ":8080"),
			BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
			ReadTimeout:  getDuration("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		DB: DB{
			DSN: getEnv("DATABASE_DSN", "file:tinyurl.db?_journal_mode=WAL&_busy_timeout=5000"),
		},
		Log: Log{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
