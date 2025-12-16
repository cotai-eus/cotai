package usecase

import (
	"context"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ActivateTenantCommand represents the input for activating a tenant
type ActivateTenantCommand struct {
	TenantID uuid.UUID
}

// ActivateTenantUseCase handles tenant activation/reactivation
type ActivateTenantUseCase struct {
	repo      domain.TenantRepository
	publisher EventPublisher
	logger    *zap.Logger
}

// NewActivateTenantUseCase creates a new ActivateTenantUseCase
func NewActivateTenantUseCase(
	repo domain.TenantRepository,
	publisher EventPublisher,
	logger *zap.Logger,
) *ActivateTenantUseCase {
	return &ActivateTenantUseCase{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// Execute executes the activate tenant use case
func (uc *ActivateTenantUseCase) Execute(ctx context.Context, cmd ActivateTenantCommand) (*domain.Tenant, error) {
	uc.logger.Info("Activating tenant",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	// Get tenant
	tenant, err := uc.repo.GetByTenantID(ctx, cmd.TenantID)
	if err != nil {
		uc.logger.Error("Failed to get tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Activate tenant
	if err := tenant.Activate(); err != nil {
		uc.logger.Error("Failed to activate tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}

	// Update tenant
	if err := uc.repo.Update(ctx, tenant); err != nil {
		uc.logger.Error("Failed to update tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Publish event (async)
	go func() {
		publishCtx := context.Background()
		if err := uc.publisher.PublishTenantActivated(publishCtx, tenant); err != nil {
			uc.logger.Error("Failed to publish tenant.activated event",
				zap.String("tenant_id", tenant.TenantID.String()),
				zap.Error(err),
			)
		}
	}()

	uc.logger.Info("Tenant activated",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	return tenant, nil
}
