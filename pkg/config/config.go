package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	ServerPort     string
	AppleTeamID    string
	AppleClientID  string
	GoogleClientID string
	AllowedOrigins []string
	DatabaseURL    string
	DBMaxConns     int32
	DBMinConns     int32
	JWTSecret      string
	// Redis configuration
	RedisHost        string
	RedisPort        string
	RedisDB          int
	RedisPassword    string
	RedisMaxConns    int
	RedisMinIdleConns int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		AppleTeamID:    getEnv("APPLE_TEAM_ID", ""),
		AppleClientID:  getEnv("APPLE_CLIENT_ID", ""),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		AllowedOrigins: parseAllowedOrigins(getEnv("ALLOWED_ORIGINS", "")),
		DatabaseURL:    getEnv("DATABASE_URL", ""),
		DBMaxConns:     int32(getEnvAsInt("DB_MAX_CONNS", 25)),
		DBMinConns:     int32(getEnvAsInt("DB_MIN_CONNS", 5)),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		// Redis configuration
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
		RedisDB:          getEnvAsInt("REDIS_DB", 0),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisMaxConns:    getEnvAsInt("REDIS_MAX_CONNS", 10),
		RedisMinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 2),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate ensures required configuration is present
func (c *Config) validate() error {
	// At least one OAuth provider must be configured
	if c.AppleClientID == "" && c.GoogleClientID == "" {
		return fmt.Errorf("at least one OAuth provider (APPLE_CLIENT_ID or GOOGLE_CLIENT_ID) is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}

	// Redis configuration validation
	if c.RedisHost == "" {
		return fmt.Errorf("REDIS_HOST cannot be empty")
	}
	if c.RedisPort == "" {
		return fmt.Errorf("REDIS_PORT cannot be empty")
	}

	// Validate Redis port is a valid number in range
	port, err := strconv.Atoi(c.RedisPort)
	if err != nil {
		return fmt.Errorf("REDIS_PORT must be a valid number: %w", err)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("REDIS_PORT must be between 1 and 65535, got %d", port)
	}

	// Validate Redis DB number (Redis supports 0-15 by default)
	if c.RedisDB < 0 || c.RedisDB > 15 {
		return fmt.Errorf("REDIS_DB must be between 0 and 15, got %d", c.RedisDB)
	}

	// Validate connection pool settings
	if c.RedisMaxConns <= 0 {
		return fmt.Errorf("REDIS_MAX_CONNS must be positive, got %d", c.RedisMaxConns)
	}
	if c.RedisMinIdleConns < 0 {
		return fmt.Errorf("REDIS_MIN_IDLE_CONNS cannot be negative, got %d", c.RedisMinIdleConns)
	}
	if c.RedisMinIdleConns > c.RedisMaxConns {
		return fmt.Errorf("REDIS_MIN_IDLE_CONNS (%d) cannot exceed REDIS_MAX_CONNS (%d)", c.RedisMinIdleConns, c.RedisMaxConns)
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

// getEnvAsInt retrieves an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
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
