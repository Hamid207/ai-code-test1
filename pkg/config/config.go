package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	ServerPort     string
	AppleTeamID    string
	AppleClientID  string
	AllowedOrigins []string
	DatabaseURL    string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		AppleTeamID:    getEnv("APPLE_TEAM_ID", ""),
		AppleClientID:  getEnv("APPLE_CLIENT_ID", ""),
		AllowedOrigins: parseAllowedOrigins(getEnv("ALLOWED_ORIGINS", "")),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
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
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
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

// parseAllowedOrigins parses comma-separated origins
func parseAllowedOrigins(origins string) []string {
	if origins == "" {
		return []string{} // Empty list - no CORS allowed by default (secure)
	}

	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))
	for _, origin := range parts {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
