package config

import "os"

type Config struct {
	AppName      string
	Address      string
	DatabasePath string
}

func Load() Config {
	return Config{
		AppName:      envOrDefault("APP_NAME", "GistLike"),
		Address:      envOrDefault("APP_ADDR", ":8080"),
		DatabasePath: envOrDefault("APP_DB_PATH", "data/snippets.db"),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
