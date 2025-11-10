package redis

import (
	"go.uber.org/zap"
)

// Logger interface for Redis operations
// This allows for dependency injection and testing
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
}

// noopLogger is a logger that does nothing
// Used as default when no logger is provided
type noopLogger struct{}

func (n *noopLogger) Debug(msg string, fields ...zap.Field) {}
func (n *noopLogger) Info(msg string, fields ...zap.Field)  {}
func (n *noopLogger) Warn(msg string, fields ...zap.Field)  {}
func (n *noopLogger) Error(msg string, fields ...zap.Field) {}

// defaultLogger is used when no logger is explicitly provided
var defaultLogger Logger = &noopLogger{}

// SetDefaultLogger sets the default logger for all Redis operations
// This should be called during application initialization
func SetDefaultLogger(logger Logger) {
	if logger != nil {
		defaultLogger = logger
	}
}

// zapLogger wraps zap.Logger to implement our Logger interface
type zapLogger struct {
	*zap.Logger
}

// NewZapLogger creates a new Logger from a zap.Logger
func NewZapLogger(logger *zap.Logger) Logger {
	return &zapLogger{Logger: logger}
}
