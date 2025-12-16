package provisioning

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// RLSManager handles Row-Level Security policy management
type RLSManager struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewRLSManager creates a new RLS manager
func NewRLSManager(db *sqlx.DB, logger *zap.Logger) *RLSManager {
	return &RLSManager{
		db:     db,
		logger: logger,
	}
}

// EnableRLSForSchema enables RLS on all tables in a schema
func (m *RLSManager) EnableRLSForSchema(ctx context.Context, schemaName string, tenantID uuid.UUID) error {
	// Get all tables in the schema
	tables, err := m.getTablesInSchema(ctx, schemaName)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	m.logger.Info("Enabling RLS for schema",
		zap.String("schema", schemaName),
		zap.Int("tables", len(tables)),
	)

	for _, table := range tables {
		if err := m.enableRLSForTable(ctx, schemaName, table, tenantID); err != nil {
			m.logger.Error("Failed to enable RLS for table",
				zap.String("table", table),
				zap.Error(err),
			)
			return fmt.Errorf("failed to enable RLS for table %s: %w", table, err)
		}
	}

	m.logger.Info("RLS enabled for all tables",
		zap.String("schema", schemaName),
	)

	return nil
}

// enableRLSForTable enables RLS on a specific table
func (m *RLSManager) enableRLSForTable(ctx context.Context, schemaName, tableName string, tenantID uuid.UUID) error {
	fullyQualifiedTable := fmt.Sprintf("%s.%s", schemaName, tableName)

	// Enable RLS on the table
	enableQuery := fmt.Sprintf("ALTER TABLE %s ENABLE ROW LEVEL SECURITY", fullyQualifiedTable)
	if _, err := m.db.ExecContext(ctx, enableQuery); err != nil {
		return fmt.Errorf("failed to enable RLS: %w", err)
	}

	// Create SELECT policy
	selectPolicy := fmt.Sprintf(`
		CREATE POLICY tenant_isolation_select ON %s
		FOR SELECT
		USING (tenant_id = current_setting('app.current_tenant', true)::uuid)
	`, fullyQualifiedTable)
	if _, err := m.db.ExecContext(ctx, selectPolicy); err != nil {
		m.logger.Warn("Failed to create SELECT policy (may already exist)", zap.Error(err))
	}

	// Create INSERT policy
	insertPolicy := fmt.Sprintf(`
		CREATE POLICY tenant_isolation_insert ON %s
		FOR INSERT
		WITH CHECK (tenant_id = current_setting('app.current_tenant', true)::uuid)
	`, fullyQualifiedTable)
	if _, err := m.db.ExecContext(ctx, insertPolicy); err != nil {
		m.logger.Warn("Failed to create INSERT policy (may already exist)", zap.Error(err))
	}

	// Create UPDATE policy
	updatePolicy := fmt.Sprintf(`
		CREATE POLICY tenant_isolation_update ON %s
		FOR UPDATE
		USING (tenant_id = current_setting('app.current_tenant', true)::uuid)
	`, fullyQualifiedTable)
	if _, err := m.db.ExecContext(ctx, updatePolicy); err != nil {
		m.logger.Warn("Failed to create UPDATE policy (may already exist)", zap.Error(err))
	}

	// Create DELETE policy
	deletePolicy := fmt.Sprintf(`
		CREATE POLICY tenant_isolation_delete ON %s
		FOR DELETE
		USING (tenant_id = current_setting('app.current_tenant', true)::uuid)
	`, fullyQualifiedTable)
	if _, err := m.db.ExecContext(ctx, deletePolicy); err != nil {
		m.logger.Warn("Failed to create DELETE policy (may already exist)", zap.Error(err))
	}

	m.logger.Debug("RLS enabled for table",
		zap.String("table", fullyQualifiedTable),
	)

	return nil
}

// getTablesInSchema returns all tables in a schema
func (m *RLSManager) getTablesInSchema(ctx context.Context, schemaName string) ([]string, error) {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	var tables []string
	err := m.db.SelectContext(ctx, &tables, query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}

	return tables, nil
}

// DisableRLSForTable disables RLS on a specific table
func (m *RLSManager) DisableRLSForTable(ctx context.Context, schemaName, tableName string) error {
	fullyQualifiedTable := fmt.Sprintf("%s.%s", schemaName, tableName)

	query := fmt.Sprintf("ALTER TABLE %s DISABLE ROW LEVEL SECURITY", fullyQualifiedTable)
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to disable RLS: %w", err)
	}

	m.logger.Debug("RLS disabled for table",
		zap.String("table", fullyQualifiedTable),
	)

	return nil
}

// CheckRLSEnabled checks if RLS is enabled on a table
func (m *RLSManager) CheckRLSEnabled(ctx context.Context, schemaName, tableName string) (bool, error) {
	query := `
		SELECT relrowsecurity
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
	`

	var enabled bool
	err := m.db.GetContext(ctx, &enabled, query, schemaName, tableName)
	if err != nil {
		return false, fmt.Errorf("failed to check RLS status: %w", err)
	}

	return enabled, nil
}
