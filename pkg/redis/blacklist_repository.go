package redis

import (
	"context"
	"fmt"
	"time"
)

const (
	// minBlacklistTTL is the minimum TTL for blacklisted tokens
	// Provides safety margin for network latency to ensure token is blacklisted
	minBlacklistTTL = 100 * time.Millisecond
)

// BlacklistRepository implements repository.RedisBlacklistRepository
type BlacklistRepository struct {
	client     *Client
	keyBuilder *KeyBuilder
}

// NewBlacklistRepository creates a new BlacklistRepository
func NewBlacklistRepository(client *Client) *BlacklistRepository {
	return &BlacklistRepository{
		client:     client,
		keyBuilder: NewKeyBuilder(),
	}
}

// AddToBlacklist adds a token to the blacklist with TTL
// The token will automatically expire from the blacklist when it naturally expires
func (r *BlacklistRepository) AddToBlacklist(ctx context.Context, tokenID string, expiresAt time.Time) error {
	key := r.keyBuilder.BlacklistToken(tokenID)
	ttl := time.Until(expiresAt)

	// Safety margin to account for network latency
	// If token expires very soon, still blacklist it with minimum TTL
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}
	if ttl < minBlacklistTTL {
		// Token expires very soon, use minimum TTL to ensure blacklisting
		ttl = minBlacklistTTL
	}

	// Store a simple marker value (we only care about key existence)
	// The value "1" indicates the token is blacklisted
	err := r.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *BlacklistRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := r.keyBuilder.BlacklistToken(tokenID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists > 0, nil
}
