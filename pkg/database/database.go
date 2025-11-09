package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

// PoolConfig holds configurable pool settings
type PoolConfig struct {
	MaxConns int32
	MinConns int32
}

// pgxLogger implements tracelog.Logger interface for query logging
type pgxLogger struct{}

func (l *pgxLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	// Format: [DB] level: message {data}
	logMsg := fmt.Sprintf("[DB] %s: %s", level, msg)
	if len(data) > 0 {
		logMsg = fmt.Sprintf("%s %v", logMsg, data)
	}
	log.Println(logMsg)
}

// NewPool creates a new PostgreSQL connection pool
func NewPool(ctx context.Context, databaseURL string, poolCfg PoolConfig) (*pgxpool.Pool, error) {
	// Parse the connection string and configure the pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure connection pool settings (now configurable!)
	config.MaxConns = poolCfg.MaxConns
	config.MinConns = poolCfg.MinConns
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute
	config.ConnConfig.ConnectTimeout = 10 * time.Second

	// Enable query logging for observability
	// Log level: Error only in production (can be made configurable)
	config.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger:   &pgxLogger{},
		LogLevel: tracelog.LogLevelError, // Only log errors
	}

	// Create the connection pool with timeout
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection with timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
