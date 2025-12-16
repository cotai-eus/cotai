package observability

import (
	"go.uber.org/zap"
)

// NewLogger creates a new structured logger based on environment
func NewLogger(env string, logLevel string) (*zap.Logger, error) {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// Set log level
	level, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	cfg.Level = level

	// JSON encoding for production
	if env == "production" {
		cfg.Encoding = "json"
	}

	return cfg.Build()
}

// NewNopLogger creates a no-op logger for testing
func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}
