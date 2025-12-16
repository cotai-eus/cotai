package usecase

import (
	"context"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SuspendTenantCommand represents the input for suspending a tenant
type SuspendTenantCommand struct {
	TenantID uuid.UUID
	Reason   string
}

// SuspendTenantUseCase handles tenant suspension
type SuspendTenantUseCase struct {
	repo      domain.TenantRepository
	publisher EventPublisher
	logger    *zap.Logger
}

// NewSuspendTenantUseCase creates a new SuspendTenantUseCase
func NewSuspendTenantUseCase(
	repo domain.TenantRepository,
	publisher EventPublisher,
	logger *zap.Logger,
) *SuspendTenantUseCase {
	return &SuspendTenantUseCase{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// Execute executes the suspend tenant use case
func (uc *SuspendTenantUseCase) Execute(ctx context.Context, cmd SuspendTenantCommand) (*domain.Tenant, error) {
	uc.logger.Warn("Suspending tenant",
		zap.String("tenant_id", cmd.TenantID.String()),
		zap.String("reason", cmd.Reason),
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

	// Suspend tenant
	if err := tenant.Suspend(cmd.Reason); err != nil {
		uc.logger.Error("Failed to suspend tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to suspend tenant: %w", err)
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
		if err := uc.publisher.PublishTenantSuspended(publishCtx, tenant); err != nil {
			uc.logger.Error("Failed to publish tenant.suspended event",
				zap.String("tenant_id", tenant.TenantID.String()),
				zap.Error(err),
			)
		}
	}()

	uc.logger.Info("Tenant suspended",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	return tenant, nil
}
