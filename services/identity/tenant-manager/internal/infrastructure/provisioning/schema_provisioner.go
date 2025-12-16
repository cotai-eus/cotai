package provisioning

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// SchemaProvisioner handles tenant schema provisioning
type SchemaProvisioner struct {
	db             *sqlx.DB
	migrationsPath string
	logger         *zap.Logger
}

// NewSchemaProvisioner creates a new schema provisioner
func NewSchemaProvisioner(db *sqlx.DB, migrationsPath string, logger *zap.Logger) *SchemaProvisioner {
	return &SchemaProvisioner{
		db:             db,
		migrationsPath: migrationsPath,
		logger:         logger,
	}
}

// ProvisionTenant provisions a complete tenant schema
func (p *SchemaProvisioner) ProvisionTenant(ctx context.Context, tenantID uuid.UUID) error {
	schemaName := FormatSchemaName(tenantID)

	p.logger.Info("Starting tenant provisioning",
		zap.String("tenant_id", tenantID.String()),
		zap.String("schema", schemaName),
	)

	startTime := time.Now()

	// Step 1: Create schema
	if err := p.createSchema(ctx, schemaName); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Step 2: Run migrations
	if err := p.runMigrations(ctx, schemaName); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Step 3: Seed initial data (optional)
	if err := p.seedInitialData(ctx, schemaName, tenantID); err != nil {
		p.logger.Warn("Failed to seed initial data", zap.Error(err))
		// Don't fail provisioning if seeding fails
	}

	duration := time.Since(startTime)
	p.logger.Info("Tenant provisioning completed",
		zap.String("tenant_id", tenantID.String()),
		zap.Duration("duration", duration),
	)

	return nil
}

// DeProvisionTenant removes a tenant schema
func (p *SchemaProvisioner) DeProvisionTenant(ctx context.Context, tenantID uuid.UUID) error {
	schemaName := FormatSchemaName(tenantID)

	p.logger.Warn("Deprovisioning tenant schema",
		zap.String("tenant_id", tenantID.String()),
		zap.String("schema", schemaName),
	)

	// Drop schema cascade (removes all tables, functions, etc.)
	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop schema: %w", err)
	}

	p.logger.Info("Tenant schema deprovisioned",
		zap.String("tenant_id", tenantID.String()),
	)

	return nil
}

// SchemaExists checks if a tenant schema exists
func (p *SchemaProvisioner) SchemaExists(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	schemaName := FormatSchemaName(tenantID)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM information_schema.schemata
			WHERE schema_name = $1
		)
	`

	var exists bool
	err := p.db.GetContext(ctx, &exists, query, schemaName)
	if err != nil {
		return false, fmt.Errorf("failed to check schema existence: %w", err)
	}

	return exists, nil
}

// createSchema creates a new PostgreSQL schema
func (p *SchemaProvisioner) createSchema(ctx context.Context, schemaName string) error {
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute CREATE SCHEMA: %w", err)
	}

	p.logger.Debug("Schema created", zap.String("schema", schemaName))
	return nil
}

// runMigrations runs all migrations for a tenant schema
func (p *SchemaProvisioner) runMigrations(ctx context.Context, schemaName string) error {
	// Create a temporary connection with search_path set to the tenant schema
	connStr := p.db.DriverName()

	// Get underlying SQL DB
	sqlDB := p.db.DB

	// Create driver instance for migrations
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		MigrationsTable: "schema_migrations",
		SchemaName:      schemaName,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		p.migrationsPath,
		connStr,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run all migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	p.logger.Debug("Migrations completed", zap.String("schema", schemaName))
	return nil
}

// seedInitialData seeds initial data for a tenant
func (p *SchemaProvisioner) seedInitialData(ctx context.Context, schemaName string, tenantID uuid.UUID) error {
	// Set search_path to tenant schema
	_, err := p.db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Example: Insert default configuration or lookup tables
	// This is tenant-specific initial data

	p.logger.Debug("Initial data seeded", zap.String("schema", schemaName))
	return nil
}

// FormatSchemaName formats tenant ID into PostgreSQL schema name
// Example: "550e8400-e29b-41d4-a716-446655440000" -> "tenant_550e8400e29b41d4a716446655440000"
func FormatSchemaName(tenantID uuid.UUID) string {
	cleanID := ""
	for _, ch := range tenantID.String() {
		if ch != '-' {
			cleanID += string(ch)
		}
	}
	return fmt.Sprintf("tenant_%s", cleanID)
}

// GetSchemaInfo returns information about a tenant schema
func (p *SchemaProvisioner) GetSchemaInfo(ctx context.Context, schemaName string) (*SchemaInfo, error) {
	query := `
		SELECT
			schemaname,
			pg_size_pretty(pg_total_relation_size(schemaname||'.*')) as size
		FROM pg_tables
		WHERE schemaname = $1
		GROUP BY schemaname
		LIMIT 1
	`

	var info SchemaInfo
	err := p.db.GetContext(ctx, &info, query, schemaName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("schema not found: %s", schemaName)
		}
		return nil, fmt.Errorf("failed to get schema info: %w", err)
	}

	return &info, nil
}

// SchemaInfo holds information about a tenant schema
type SchemaInfo struct {
	SchemaName string `db:"schemaname"`
	Size       string `db:"size"`
}
