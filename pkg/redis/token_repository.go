package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenRepository implements repository.RedisTokenRepository
type TokenRepository struct {
	client     *Client
	keyBuilder *KeyBuilder
}

// NewTokenRepository creates a new TokenRepository
func NewTokenRepository(client *Client) *TokenRepository {
	return &TokenRepository{
		client:     client,
		keyBuilder: NewKeyBuilder(),
	}
}

// StoreRefreshToken stores a refresh token in Redis with TTL
func (r *TokenRepository) StoreRefreshToken(ctx context.Context, userID int64, tokenID, tokenHash string, expiresAt time.Time) error {
	key := r.keyBuilder.RefreshToken(strconv.FormatInt(userID, 10), tokenID)
	ttl := time.Until(expiresAt)

	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}

	// Store token hash with expiration
	err := r.client.Set(ctx, key, tokenHash, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

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

	// Scan for all keys matching the pattern
	// IMPORTANT: SCAN is a blocking operation, so we add context checks
	// and limit the number of iterations to prevent infinite loops
	var cursor uint64
	const maxIterations = 1000 // Safety limit to prevent infinite loops
	iteration := 0

	for {
		// Check context deadline/cancellation on each iteration
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled while deleting tokens: %w", ctx.Err())
		default:
		}

		// Safety check: prevent infinite loops
		iteration++
		if iteration > maxIterations {
			return fmt.Errorf("exceeded maximum iterations (%d) while scanning tokens", maxIterations)
		}

		var keys []string
		var err error

		// SCAN with smaller batch size to reduce blocking time
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, 50).Result()
		if err != nil {
			return fmt.Errorf("failed to scan refresh tokens: %w", err)
		}

		if len(keys) > 0 {
			// Delete in batch
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete refresh tokens (iteration %d): %w", iteration, err)
			}
		}

		// Cursor 0 means we've completed the scan
		if cursor == 0 {
			break
		}
	}

	return nil
}
