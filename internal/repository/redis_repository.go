package repository

import (
	"context"
	"time"

	"github.com/Hamid207/ai-code-test1/internal/model"
)

// RedisTokenRepository defines operations for managing tokens in Redis
type RedisTokenRepository interface {
	// StoreRefreshToken stores a refresh token in Redis with TTL
	StoreRefreshToken(ctx context.Context, userID int64, tokenID, tokenHash string, expiresAt time.Time) error

	// GetRefreshToken retrieves a refresh token from Redis
	GetRefreshToken(ctx context.Context, userID int64, tokenID string) (string, error)

	// DeleteRefreshToken removes a refresh token from Redis
	DeleteRefreshToken(ctx context.Context, userID int64, tokenID string) error

	// DeleteAllUserTokens removes all refresh tokens for a user
	DeleteAllUserTokens(ctx context.Context, userID int64) error
}

// RedisBlacklistRepository defines operations for managing token blacklist
type RedisBlacklistRepository interface {
	// AddToBlacklist adds a token to the blacklist with TTL
	AddToBlacklist(ctx context.Context, tokenID string, expiresAt time.Time) error

	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

// RedisRateLimitRepository defines operations for rate limiting
type RedisRateLimitRepository interface {
	// IncrementUserRequest increments the request count for a user
	// Returns: current count, reset time, error
	IncrementUserRequest(ctx context.Context, userID int64, window time.Duration) (int64, time.Time, error)

	// IncrementIPRequest increments the request count for an IP address
	// Returns: current count, reset time, error
	IncrementIPRequest(ctx context.Context, ipAddress string, window time.Duration) (int64, time.Time, error)

	// GetUserRequestCount gets the current request count for a user
	GetUserRequestCount(ctx context.Context, userID int64) (int64, error)

	// GetIPRequestCount gets the current request count for an IP
	GetIPRequestCount(ctx context.Context, ipAddress string) (int64, error)
}

// RedisCacheRepository defines operations for caching
type RedisCacheRepository interface {
	// SetUserCache stores user data in cache with TTL
	SetUserCache(ctx context.Context, userID int64, user *model.User, ttl time.Duration) error

	// GetUserCache retrieves user data from cache
	GetUserCache(ctx context.Context, userID int64) (*model.User, error)

	// InvalidateUserCache removes user data from cache
	InvalidateUserCache(ctx context.Context, userID int64) error

	// SetGeneric stores any JSON-serializable data in cache
	SetGeneric(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// GetGeneric retrieves and deserializes data from cache
	GetGeneric(ctx context.Context, key string, dest interface{}) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error
}
