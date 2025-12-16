package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
	"go.uber.org/zap"
)

// PostgresDB wraps the database connection
type PostgresDB struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// Config holds database configuration
type Config struct {
	Host            string
	Port            int
	Database        string
	User            string
	Password        string
	SSLMode         string
	MaxConns        int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg Config, logger *zap.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return &PostgresDB{
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// DB returns the underlying database connection
func (p *PostgresDB) DB() *sqlx.DB {
	return p.db
}

// Health checks database health
func (p *PostgresDB) Health(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

// Ping checks database connectivity (alias for Health)
func (p *PostgresDB) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
