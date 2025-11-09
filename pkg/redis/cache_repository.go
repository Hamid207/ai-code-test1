package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/redis/go-redis/v9"
)

// CacheRepository implements repository.RedisCacheRepository
type CacheRepository struct {
	client     *Client
	keyBuilder *KeyBuilder
}

// NewCacheRepository creates a new CacheRepository
func NewCacheRepository(client *Client) *CacheRepository {
	return &CacheRepository{
		client:     client,
		keyBuilder: NewKeyBuilder(),
	}
}

// SetUserCache stores user data in cache with TTL
func (r *CacheRepository) SetUserCache(ctx context.Context, userID int64, user *model.User, ttl time.Duration) error {
	key := r.keyBuilder.UserCache(strconv.FormatInt(userID, 10))

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache user data: %w", err)
	}

	return nil
}

// GetUserCache retrieves user data from cache
func (r *CacheRepository) GetUserCache(ctx context.Context, userID int64) (*model.User, error) {
	key := r.keyBuilder.UserCache(strconv.FormatInt(userID, 10))

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user cache not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user cache: %w", err)
	}

	var user model.User
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

// InvalidateUserCache removes user data from cache
func (r *CacheRepository) InvalidateUserCache(ctx context.Context, userID int64) error {
	key := r.keyBuilder.UserCache(strconv.FormatInt(userID, 10))

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate user cache: %w", err)
	}

	return nil
}

// SetGeneric stores any JSON-serializable data in cache
func (r *CacheRepository) SetGeneric(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache data: %w", err)
	}

	return nil
}

// GetGeneric retrieves and deserializes data from cache
func (r *CacheRepository) GetGeneric(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get cache: %w", err)
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key: %w", err)
	}

	return nil
}
