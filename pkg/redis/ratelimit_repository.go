package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Lua script for atomic rate limiting (package-level constant for performance)
// This script atomically increments the counter and returns both count and TTL
// Reducing this to a constant eliminates per-request string allocation
const rateLimitScript = `
	local current = redis.call('INCR', KEYS[1])
	if current == 1 then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end
	local ttl = redis.call('TTL', KEYS[1])
	return {current, ttl}
`

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
	// Use a pre-compiled Lua script (defined at package level) for better performance
	// This prevents race conditions between INCR and TTL commands
	// Returns: {count, ttl_seconds}
	result, err := r.client.Eval(ctx, rateLimitScript, []string{key}, int(window.Seconds())).Result()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to increment request count: %w", err)
	}

	// Parse result array
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) != 2 {
		return 0, time.Time{}, fmt.Errorf("unexpected result format from Lua script")
	}

	count, ok := resultArray[0].(int64)
	if !ok {
		return 0, time.Time{}, fmt.Errorf("unexpected count type from Lua script")
	}

	ttlSeconds, ok := resultArray[1].(int64)
	if !ok {
		return 0, time.Time{}, fmt.Errorf("unexpected TTL type from Lua script")
	}

	// Calculate reset time based on TTL from Redis
	// Using time.Now() at this point is more accurate than after separate TTL call
	resetTime := time.Now().Add(time.Duration(ttlSeconds) * time.Second)

	return count, resetTime, nil
}

// getRequestCount retrieves the current request count for a key
func (r *RateLimitRepository) getRequestCount(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		// Key doesn't exist means no requests yet
		if err == redis.Nil {
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
