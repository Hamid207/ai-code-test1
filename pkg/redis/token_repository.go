package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// scanBatchSize is the number of keys to retrieve per SCAN iteration
	// Smaller batch size reduces Redis blocking time
	scanBatchSize = 50

	// maxScanIterations prevents infinite loops in SCAN operations
	// With batch size 50, this allows scanning up to 50,000 keys
	maxScanIterations = 1000

	// minTokenTTL is the minimum TTL for token storage
	// Provides safety margin for network latency and processing time
	minTokenTTL = 100 * time.Millisecond
)

// TokenRepository implements repository.RedisTokenRepository
type TokenRepository struct {
	client     *Client
	keyBuilder *KeyBuilder
	logger     Logger
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(client *Client) *TokenRepository {
	return &TokenRepository{
		client:     client,
		keyBuilder: NewKeyBuilder(),
		logger:     defaultLogger,
	}
}

// WithLogger sets a custom logger for this repository
func (r *TokenRepository) WithLogger(logger Logger) *TokenRepository {
	r.logger = logger
	return r
}

// StoreRefreshToken stores a refresh token in Redis with TTL
func (r *TokenRepository) StoreRefreshToken(ctx context.Context, userID int64, tokenID, tokenHash string, expiresAt time.Time) error {
	key := r.keyBuilder.RefreshToken(strconv.FormatInt(userID, 10), tokenID)
	ttl := time.Until(expiresAt)

	r.logger.Debug("storing refresh token",
		zap.Int64("user_id", userID),
		zap.String("token_id", tokenID),
		zap.Duration("ttl", ttl),
	)

	// Safety margin to account for network latency and processing time
	// If token expires too soon, reject it to prevent edge cases
	if ttl <= minTokenTTL {
		r.logger.Warn("token TTL too short",
			zap.Int64("user_id", userID),
			zap.Duration("ttl", ttl),
			zap.Duration("min_ttl", minTokenTTL),
		)
		return fmt.Errorf("token expires too soon (TTL: %v, minimum: %v)", ttl, minTokenTTL)
	}

	// Store token hash with expiration
	err := r.client.Set(ctx, key, tokenHash, ttl).Err()
	if err != nil {
		r.logger.Error("failed to store refresh token",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	r.logger.Info("refresh token stored successfully",
		zap.Int64("user_id", userID),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// GetRefreshToken retrieves a refresh token from Redis
func (r *TokenRepository) GetRefreshToken(ctx context.Context, userID int64, tokenID string) (string, error) {
	key := r.keyBuilder.RefreshToken(strconv.FormatInt(userID, 10), tokenID)

	tokenHash, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	return tokenHash, nil
}

// DeleteRefreshToken removes a refresh token from Redis
func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, userID int64, tokenID string) error {
	key := r.keyBuilder.RefreshToken(strconv.FormatInt(userID, 10), tokenID)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// DeleteAllUserTokens removes all refresh tokens for a user
func (r *TokenRepository) DeleteAllUserTokens(ctx context.Context, userID int64) error {
	pattern := r.keyBuilder.RefreshTokenPattern(strconv.FormatInt(userID, 10))

	r.logger.Debug("deleting all user tokens",
		zap.Int64("user_id", userID),
		zap.String("pattern", pattern),
	)

	// Scan for all keys matching the pattern
	// IMPORTANT: SCAN is a blocking operation, so we add context checks
	// and limit the number of iterations to prevent infinite loops
	var cursor uint64
	iteration := 0
	totalDeleted := 0

	for {
		// Check context deadline/cancellation on each iteration
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled while deleting tokens: %w", ctx.Err())
		default:
		}

		// Safety check: prevent infinite loops
		iteration++
		if iteration > maxScanIterations {
			return fmt.Errorf("exceeded maximum iterations (%d) while scanning tokens", maxScanIterations)
		}

		var keys []string
		var err error

		// SCAN with smaller batch size to reduce blocking time
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, scanBatchSize).Result()
		if err != nil {
			return fmt.Errorf("failed to scan refresh tokens: %w", err)
		}

		if len(keys) > 0 {
			// Delete in batch
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				r.logger.Error("failed to delete token batch",
					zap.Int64("user_id", userID),
					zap.Int("iteration", iteration),
					zap.Int("keys_count", len(keys)),
					zap.Error(err),
				)
				return fmt.Errorf("failed to delete refresh tokens (iteration %d): %w", iteration, err)
			}
			totalDeleted += len(keys)
		}

		// Cursor 0 means we've completed the scan
		if cursor == 0 {
			break
		}
	}

	r.logger.Info("deleted all user tokens",
		zap.Int64("user_id", userID),
		zap.Int("total_deleted", totalDeleted),
		zap.Int("iterations", iteration),
	)

	return nil
}
