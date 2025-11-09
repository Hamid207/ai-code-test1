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
	if c.AppleClientID == "" {
		return fmt.Errorf("APPLE_CLIENT_ID is required")
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
