package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// TenantRepository implements domain.TenantRepository
type TenantRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sqlx.DB, logger *zap.Logger) *TenantRepository {
	return &TenantRepository{
		db:     db,
		logger: logger,
	}
}

// tenantRow represents a database row from tenant_registry table
type tenantRow struct {
	ID                  uuid.UUID      `db:"id"`
	TenantID            uuid.UUID      `db:"tenant_id"`
	TenantName          string         `db:"tenant_name"`
	TenantSlug          string         `db:"tenant_slug"`
	DatabaseSchema      string         `db:"database_schema"`
	SchemaVersion       string         `db:"schema_version"`
	Status              string         `db:"status"`
	PlanTier            string         `db:"plan_tier"`
	MaxUsers            int            `db:"max_users"`
	MaxStorageGB        int            `db:"max_storage_gb"`
	PrimaryContactEmail sql.NullString `db:"primary_contact_email"`
	PrimaryContactName  sql.NullString `db:"primary_contact_name"`
	BillingEmail        sql.NullString `db:"billing_email"`
	Settings            []byte         `db:"settings"` // JSONB
	Features            []byte         `db:"features"` // JSONB
	CreatedAt           sql.NullTime   `db:"created_at"`
	UpdatedAt           sql.NullTime   `db:"updated_at"`
	ActivatedAt         sql.NullTime   `db:"activated_at"`
	SuspendedAt         sql.NullTime   `db:"suspended_at"`
	DeletedAt           sql.NullTime   `db:"deleted_at"`
	CreatedBy           uuid.NullUUID  `db:"created_by"`
	UpdatedBy           uuid.NullUUID  `db:"updated_by"`
}

// Create creates a new tenant in the database
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		INSERT INTO public.tenant_registry (
			id, tenant_id, tenant_name, tenant_slug, database_schema, schema_version,
			status, plan_tier, max_users, max_storage_gb,
			primary_contact_email, primary_contact_name, billing_email,
			settings, features,
			created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	settings, _ := json.Marshal(tenant.Settings)
	features, _ := json.Marshal(tenant.Features)

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.TenantID,
		tenant.TenantName,
		tenant.TenantSlug,
		tenant.DatabaseSchema,
		tenant.SchemaVersion,
		string(tenant.Status),
		string(tenant.PlanTier),
		tenant.MaxUsers,
		tenant.MaxStorageGB,
		tenant.PrimaryContactEmail,
		tenant.PrimaryContactName,
		tenant.BillingEmail,
		settings,
		features,
		tenant.CreatedAt,
		tenant.UpdatedAt,
		tenant.CreatedBy,
	)

	if err != nil {
		r.logger.Error("Failed to create tenant",
			zap.Error(err),
			zap.String("tenant_id", tenant.TenantID.String()),
		)
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	r.logger.Info("Tenant created",
		zap.String("tenant_id", tenant.TenantID.String()),
		zap.String("slug", tenant.TenantSlug),
	)

	return nil
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT * FROM public.tenant_registry WHERE id = $1
	`

	var row tenantRow
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return r.rowToTenant(&row)
}

// GetByTenantID retrieves a tenant by tenant_id
func (r *TenantRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	query := `
		SELECT * FROM public.tenant_registry WHERE tenant_id = $1
	`

	var row tenantRow
	err := r.db.GetContext(ctx, &row, query, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by tenant_id: %w", err)
	}

	return r.rowToTenant(&row)
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	query := `
		SELECT * FROM public.tenant_registry WHERE tenant_slug = $1
	`

	var row tenantRow
	err := r.db.GetContext(ctx, &row, query, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to get tenant by slug: %w", err)
	}

	return r.rowToTenant(&row)
}

// List retrieves all tenants with pagination
func (r *TenantRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Tenant, int, error) {
	// Build query with filters
	query := `SELECT * FROM public.tenant_registry WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM public.tenant_registry WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		countQuery += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, string(filter.Status))
		argPos++
	}

	if filter.PlanTier != "" {
		query += fmt.Sprintf(" AND plan_tier = $%d", argPos)
		countQuery += fmt.Sprintf(" AND plan_tier = $%d", argPos)
		args = append(args, string(filter.PlanTier))
		argPos++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (tenant_name ILIKE $%d OR tenant_slug ILIKE $%d)", argPos, argPos)
		countQuery += fmt.Sprintf(" AND (tenant_name ILIKE $%d OR tenant_slug ILIKE $%d)", argPos, argPos)
		args = append(args, searchPattern)
		argPos++
	}

	// Get total count
	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Add pagination
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, filter.Limit(), filter.Offset())

	// Execute query
	var rows []tenantRow
	err = r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Convert rows to domain tenants
	tenants := make([]*domain.Tenant, 0, len(rows))
	for _, row := range rows {
		tenant, err := r.rowToTenant(&row)
		if err != nil {
			r.logger.Warn("Failed to convert tenant row", zap.Error(err))
			continue
		}
		tenants = append(tenants, tenant)
	}

	return tenants, total, nil
}

// Update updates an existing tenant
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
		UPDATE public.tenant_registry SET
			tenant_name = $1,
			status = $2,
			plan_tier = $3,
			max_users = $4,
			max_storage_gb = $5,
			primary_contact_email = $6,
			primary_contact_name = $7,
			billing_email = $8,
			settings = $9,
			features = $10,
			updated_at = $11,
			activated_at = $12,
			suspended_at = $13,
			deleted_at = $14,
			updated_by = $15
		WHERE tenant_id = $16
	`

	settings, _ := json.Marshal(tenant.Settings)
	features, _ := json.Marshal(tenant.Features)

	result, err := r.db.ExecContext(ctx, query,
		tenant.TenantName,
		string(tenant.Status),
		string(tenant.PlanTier),
		tenant.MaxUsers,
		tenant.MaxStorageGB,
		tenant.PrimaryContactEmail,
		tenant.PrimaryContactName,
		tenant.BillingEmail,
		settings,
		features,
		tenant.UpdatedAt,
		tenant.ActivatedAt,
		tenant.SuspendedAt,
		tenant.DeletedAt,
		tenant.UpdatedBy,
		tenant.TenantID,
	)

	if err != nil {
		r.logger.Error("Failed to update tenant",
			zap.Error(err),
			zap.String("tenant_id", tenant.TenantID.String()),
		)
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrTenantNotFound
	}

	r.logger.Info("Tenant updated",
		zap.String("tenant_id", tenant.TenantID.String()),
	)

	return nil
}

// Delete soft-deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE public.tenant_registry SET
			status = $1,
			deleted_at = $2,
			updated_at = $3
		WHERE tenant_id = $4 AND status != $1
	`

	result, err := r.db.ExecContext(ctx, query,
		string(domain.StatusDeleted),
		sql.NullTime{Time: sql.NullTime{}.Time, Valid: true},
		sql.NullTime{}.Time,
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrTenantNotFound
	}

	return nil
}

// ExistsBySlug checks if a tenant with the given slug exists
func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM public.tenant_registry WHERE tenant_slug = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, slug)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant slug existence: %w", err)
	}

	return exists, nil
}

// CountByStatus counts tenants by status
func (r *TenantRepository) CountByStatus(ctx context.Context, status domain.TenantStatus) (int, error) {
	query := `SELECT COUNT(*) FROM public.tenant_registry WHERE status = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, string(status))
	if err != nil {
		return 0, fmt.Errorf("failed to count tenants by status: %w", err)
	}

	return count, nil
}

// rowToTenant converts a database row to a domain Tenant
func (r *TenantRepository) rowToTenant(row *tenantRow) (*domain.Tenant, error) {
	tenant := &domain.Tenant{
		ID:             row.ID,
		TenantID:       row.TenantID,
		TenantName:     row.TenantName,
		TenantSlug:     row.TenantSlug,
		DatabaseSchema: row.DatabaseSchema,
		SchemaVersion:  row.SchemaVersion,
		Status:         domain.TenantStatus(row.Status),
		PlanTier:       domain.PlanTier(row.PlanTier),
		MaxUsers:       row.MaxUsers,
		MaxStorageGB:   row.MaxStorageGB,
	}

	// Handle nullable fields
	if row.PrimaryContactEmail.Valid {
		tenant.PrimaryContactEmail = row.PrimaryContactEmail.String
	}
	if row.PrimaryContactName.Valid {
		tenant.PrimaryContactName = row.PrimaryContactName.String
	}
	if row.BillingEmail.Valid {
		tenant.BillingEmail = row.BillingEmail.String
	}

	// Parse JSONB fields
	if len(row.Settings) > 0 {
		if err := json.Unmarshal(row.Settings, &tenant.Settings); err != nil {
			r.logger.Warn("Failed to unmarshal settings", zap.Error(err))
			tenant.Settings = make(map[string]interface{})
		}
	} else {
		tenant.Settings = make(map[string]interface{})
	}

	if len(row.Features) > 0 {
		if err := json.Unmarshal(row.Features, &tenant.Features); err != nil {
			r.logger.Warn("Failed to unmarshal features", zap.Error(err))
			tenant.Features = make(map[string]interface{})
		}
	} else {
		tenant.Features = make(map[string]interface{})
	}

	// Handle timestamps
	if row.CreatedAt.Valid {
		tenant.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		tenant.UpdatedAt = row.UpdatedAt.Time
	}
	if row.ActivatedAt.Valid {
		tenant.ActivatedAt = &row.ActivatedAt.Time
	}
	if row.SuspendedAt.Valid {
		tenant.SuspendedAt = &row.SuspendedAt.Time
	}
	if row.DeletedAt.Valid {
		tenant.DeletedAt = &row.DeletedAt.Time
	}

	// Handle UUIDs
	if row.CreatedBy.Valid {
		tenant.CreatedBy = &row.CreatedBy.UUID
	}
	if row.UpdatedBy.Valid {
		tenant.UpdatedBy = &row.UpdatedBy.UUID
	}

	return tenant, nil
}
