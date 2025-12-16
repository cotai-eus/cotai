package usecase

import (
	"context"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetTenantUseCase handles retrieving a tenant
type GetTenantUseCase struct {
	repo   domain.TenantRepository
	logger *zap.Logger
}

// NewGetTenantUseCase creates a new GetTenantUseCase
func NewGetTenantUseCase(repo domain.TenantRepository, logger *zap.Logger) *GetTenantUseCase {
	return &GetTenantUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute retrieves a tenant by tenant_id (default method for gRPC)
func (uc *GetTenantUseCase) Execute(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	return uc.ExecuteByTenantID(ctx, tenantID)
}

// ExecuteByID retrieves a tenant by ID
func (uc *GetTenantUseCase) ExecuteByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	tenant, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get tenant by ID",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// ExecuteByTenantID retrieves a tenant by tenant_id
func (uc *GetTenantUseCase) ExecuteByTenantID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := uc.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		uc.logger.Error("Failed to get tenant by tenant_id",
			zap.String("tenant_id", tenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// ExecuteBySlug retrieves a tenant by slug
func (uc *GetTenantUseCase) ExecuteBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	tenant, err := uc.repo.GetBySlug(ctx, slug)
	if err != nil {
		uc.logger.Error("Failed to get tenant by slug",
			zap.String("slug", slug),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// GetBySlug is an alias for ExecuteBySlug (for gRPC compatibility)
func (uc *GetTenantUseCase) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	return uc.ExecuteBySlug(ctx, slug)
}
