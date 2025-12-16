package domain

import (
	"context"

	"github.com/google/uuid"
)

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *Tenant) error

	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)

	// GetByTenantID retrieves a tenant by tenant_id
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error)

	// GetBySlug retrieves a tenant by slug
	GetBySlug(ctx context.Context, slug string) (*Tenant, error)

	// List retrieves all tenants with pagination
	List(ctx context.Context, filter ListFilter) ([]*Tenant, int, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *Tenant) error

	// Delete soft-deletes a tenant
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsBySlug checks if a tenant with the given slug exists
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// CountByStatus counts tenants by status
	CountByStatus(ctx context.Context, status TenantStatus) (int, error)
}

// ListFilter defines filters for listing tenants
type ListFilter struct {
	Page     int
	PerPage  int
	Status   TenantStatus
	PlanTier PlanTier
	Search   string
}

// Offset calculates the offset for pagination
func (f ListFilter) Offset() int {
	if f.Page <= 0 {
		return 0
	}
	return (f.Page - 1) * f.PerPage
}

// Limit returns the page size
func (f ListFilter) Limit() int {
	if f.PerPage <= 0 {
		return 20 // default page size
	}
	if f.PerPage > 100 {
		return 100 // max page size
	}
	return f.PerPage
}
