package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with our configuration
type Client struct {
	*redis.Client
}

// Config holds Redis connection configuration
type Config struct {
	Host        string
	Port        string
	DB          int
	Password    string
	MaxConns    int
	MinIdleConns int
}

// NewClient creates a new Redis client with connection pooling
func NewClient(cfg Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.MaxConns,
		MinIdleConns: cfg.MinIdleConns,

		// Connection timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Pool timeouts
		PoolTimeout:      4 * time.Second,
		ConnMaxIdleTime:  5 * time.Minute,

		// Retry configuration
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: rdb}, nil
}

// Close gracefully closes the Redis connection
func (c *Client) Close() error {
	return c.Client.Close()
}

// HealthCheck performs a health check on the Redis connection
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.Ping(ctx).Err()
}
