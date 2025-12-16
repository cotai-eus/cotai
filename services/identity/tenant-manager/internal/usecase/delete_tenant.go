package usecase

import (
	"context"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteTenantCommand represents the input for deleting a tenant
type DeleteTenantCommand struct {
	TenantID uuid.UUID
}

// DeleteTenantUseCase handles tenant soft deletion
type DeleteTenantUseCase struct {
	repo      domain.TenantRepository
	publisher EventPublisher
	logger    *zap.Logger
}

// NewDeleteTenantUseCase creates a new DeleteTenantUseCase
func NewDeleteTenantUseCase(
	repo domain.TenantRepository,
	publisher EventPublisher,
	logger *zap.Logger,
) *DeleteTenantUseCase {
	return &DeleteTenantUseCase{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// Execute executes the delete tenant use case (soft delete)
func (uc *DeleteTenantUseCase) Execute(ctx context.Context, cmd DeleteTenantCommand) error {
	uc.logger.Warn("Deleting tenant (soft delete)",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	// Get tenant
	tenant, err := uc.repo.GetByTenantID(ctx, cmd.TenantID)
	if err != nil {
		uc.logger.Error("Failed to get tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Soft delete tenant
	if err := tenant.Delete(); err != nil {
		uc.logger.Error("Failed to delete tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Update tenant
	if err := uc.repo.Update(ctx, tenant); err != nil {
		uc.logger.Error("Failed to update tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	// Publish event (async)
	go func() {
		publishCtx := context.Background()
		if err := uc.publisher.PublishTenantDeleted(publishCtx, tenant.TenantID); err != nil {
			uc.logger.Error("Failed to publish tenant.deleted event",
				zap.String("tenant_id", tenant.TenantID.String()),
				zap.Error(err),
			)
		}
	}()

	uc.logger.Info("Tenant deleted",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	return nil
}
