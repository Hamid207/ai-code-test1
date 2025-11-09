package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// RateLimitRepository implements repository.RedisRateLimitRepository
type RateLimitRepository struct {
	client     *Client
	keyBuilder *KeyBuilder
}

// NewRateLimitRepository creates a new RateLimitRepository
func NewRateLimitRepository(client *Client) *RateLimitRepository {
	return &RateLimitRepository{
		client:     client,
		keyBuilder: NewKeyBuilder(),
	}
}

// IncrementUserRequest increments the request count for a user
// Returns: current count, reset time, error
func (r *RateLimitRepository) IncrementUserRequest(ctx context.Context, userID int64, window time.Duration) (int64, time.Time, error) {
	key := r.keyBuilder.RateLimitUser(strconv.FormatInt(userID, 10))
	return r.incrementRequest(ctx, key, window)
}

// IncrementIPRequest increments the request count for an IP address
// Returns: current count, reset time, error
func (r *RateLimitRepository) IncrementIPRequest(ctx context.Context, ipAddress string, window time.Duration) (int64, time.Time, error) {
	key := r.keyBuilder.RateLimitIP(ipAddress)
	return r.incrementRequest(ctx, key, window)
}

// GetUserRequestCount gets the current request count for a user
func (r *RateLimitRepository) GetUserRequestCount(ctx context.Context, userID int64) (int64, error) {
	key := r.keyBuilder.RateLimitUser(strconv.FormatInt(userID, 10))
	return r.getRequestCount(ctx, key)
}

// GetIPRequestCount gets the current request count for an IP
func (r *RateLimitRepository) GetIPRequestCount(ctx context.Context, ipAddress string) (int64, error) {
	key := r.keyBuilder.RateLimitIP(ipAddress)
	return r.getRequestCount(ctx, key)
}

// incrementRequest is a helper function that implements the rate limiting logic
// using atomic increment with TTL
func (r *RateLimitRepository) incrementRequest(ctx context.Context, key string, window time.Duration) (int64, time.Time, error) {
	// Use a Lua script to atomically increment and set TTL if key doesn't exist
	// This ensures the TTL is set only on first request in the window
	luaScript := `
		local current = redis.call('INCR', KEYS[1])
		if current == 1 then
			redis.call('EXPIRE', KEYS[1], ARGV[1])
		end
		return current
	`

	result, err := r.client.Eval(ctx, luaScript, []string{key}, int(window.Seconds())).Result()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to increment request count: %w", err)
	}

	count, ok := result.(int64)
	if !ok {
		return 0, time.Time{}, fmt.Errorf("unexpected result type from Lua script")
	}

	// Get the TTL to determine reset time
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return count, time.Time{}, fmt.Errorf("failed to get TTL: %w", err)
	}

	resetTime := time.Now().Add(ttl)

	return count, resetTime, nil
}

// getRequestCount retrieves the current request count for a key
func (r *RateLimitRepository) getRequestCount(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		// Key doesn't exist means no requests yet
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get request count: %w", err)
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse request count: %w", err)
	}

	return count, nil
}
