package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	ServerPort    string
	AppleTeamID   string
	AppleClientID string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:    getEnv("SERVER_PORT", "8080"),
		AppleTeamID:   getEnv("APPLE_TEAM_ID", ""),
		AppleClientID: getEnv("APPLE_CLIENT_ID", ""),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate ensures required configuration is present
func (c *Config) validate() error {
	if c.AppleClientID == "" {
		return fmt.Errorf("APPLE_CLIENT_ID is required")
	}
	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
