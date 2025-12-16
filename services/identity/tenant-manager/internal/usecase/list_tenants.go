package usecase

import (
	"context"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"go.uber.org/zap"
)

// ListTenantsQuery represents the query for listing tenants
type ListTenantsQuery struct {
	Page     int
	PerPage  int
	Status   domain.TenantStatus
	PlanTier domain.PlanTier
	Search   string
}

// ListTenantsResult represents the result of listing tenants
type ListTenantsResult struct {
	Tenants    []*domain.Tenant
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

// ListTenantsUseCase handles listing tenants with pagination
type ListTenantsUseCase struct {
	repo   domain.TenantRepository
	logger *zap.Logger
}

// NewListTenantsUseCase creates a new ListTenantsUseCase
func NewListTenantsUseCase(repo domain.TenantRepository, logger *zap.Logger) *ListTenantsUseCase {
	return &ListTenantsUseCase{
		repo:   repo,
		logger: logger,
	}
}

// Execute executes the list tenants use case
func (uc *ListTenantsUseCase) Execute(ctx context.Context, query ListTenantsQuery) (*ListTenantsResult, error) {
	// Build filter from query
	filter := domain.ListFilter{
		Page:     query.Page,
		PerPage:  query.PerPage,
		Status:   query.Status,
		PlanTier: query.PlanTier,
		Search:   query.Search,
	}

	// Retrieve tenants
	tenants, total, err := uc.repo.List(ctx, filter)
	if err != nil {
		uc.logger.Error("Failed to list tenants", zap.Error(err))
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	// Calculate total pages
	totalPages := (total + filter.Limit() - 1) / filter.Limit()

	uc.logger.Info("Tenants listed",
		zap.Int("count", len(tenants)),
		zap.Int("total", total),
		zap.Int("page", query.Page),
	)

	return &ListTenantsResult{
		Tenants:    tenants,
		Total:      total,
		Page:       query.Page,
		PerPage:    filter.Limit(),
		TotalPages: totalPages,
	}, nil
}
