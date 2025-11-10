package redis

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
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
	logger     Logger
}

// NewBlacklistRepository creates a new BlacklistRepository
func NewBlacklistRepository(client *Client) *BlacklistRepository {
	return &BlacklistRepository{
		client:     client,
		keyBuilder: NewKeyBuilder(),
		logger:     defaultLogger,
	}
}

// WithLogger sets a custom logger for this repository
func (r *BlacklistRepository) WithLogger(logger Logger) *BlacklistRepository {
	r.logger = logger
	return r
}

// AddToBlacklist adds a token to the blacklist with TTL
// The token will automatically expire from the blacklist when it naturally expires
func (r *BlacklistRepository) AddToBlacklist(ctx context.Context, tokenID string, expiresAt time.Time) error {
	key := r.keyBuilder.BlacklistToken(tokenID)
	ttl := time.Until(expiresAt)

	r.logger.Debug("adding token to blacklist",
		zap.String("token_id", tokenID),
		zap.Duration("ttl", ttl),
	)

	// Safety margin to account for network latency
	// If token expires very soon, still blacklist it with minimum TTL
	if ttl <= 0 {
		r.logger.Debug("token already expired, skipping blacklist",
			zap.String("token_id", tokenID),
		)
		// Token already expired, no need to blacklist
		return nil
	}
	if ttl < minBlacklistTTL {
		r.logger.Debug("using minimum TTL for near-expiry token",
			zap.String("token_id", tokenID),
			zap.Duration("original_ttl", ttl),
			zap.Duration("min_ttl", minBlacklistTTL),
		)
		// Token expires very soon, use minimum TTL to ensure blacklisting
		ttl = minBlacklistTTL
	}

	// Store a simple marker value (we only care about key existence)
	// The value "1" indicates the token is blacklisted
	err := r.client.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		r.logger.Error("failed to blacklist token",
			zap.String("token_id", tokenID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	r.logger.Info("token blacklisted successfully",
		zap.String("token_id", tokenID),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (r *BlacklistRepository) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	key := r.keyBuilder.BlacklistToken(tokenID)

	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.logger.Error("failed to check blacklist",
			zap.String("token_id", tokenID),
			zap.Error(err),
		)
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	isBlacklisted := exists > 0

	r.logger.Debug("blacklist check result",
		zap.String("token_id", tokenID),
		zap.Bool("is_blacklisted", isBlacklisted),
	)

	return isBlacklisted, nil
}
