package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	DatabaseURL string
	Port        string
}

// Load reads and validates configuration from the environment.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.Port == "" {
		return nil, fmt.Errorf("PORT is required")

	}
	return cfg, nil
}
