package usecase

import (
	"context"
	"fmt"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UpdateTenantCommand represents the input for updating a tenant
type UpdateTenantCommand struct {
	TenantID     uuid.UUID
	Name         *string
	ContactEmail *string
	ContactName  *string
	BillingEmail *string
	Settings     map[string]interface{}
}

// UpdateTenantUseCase handles tenant updates
type UpdateTenantUseCase struct {
	repo      domain.TenantRepository
	publisher EventPublisher
	logger    *zap.Logger
}

// NewUpdateTenantUseCase creates a new UpdateTenantUseCase
func NewUpdateTenantUseCase(
	repo domain.TenantRepository,
	publisher EventPublisher,
	logger *zap.Logger,
) *UpdateTenantUseCase {
	return &UpdateTenantUseCase{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// Execute executes the update tenant use case
func (uc *UpdateTenantUseCase) Execute(ctx context.Context, cmd UpdateTenantCommand) (*domain.Tenant, error) {
	uc.logger.Info("Updating tenant",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	// Get existing tenant
	tenant, err := uc.repo.GetByTenantID(ctx, cmd.TenantID)
	if err != nil {
		uc.logger.Error("Failed to get tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// Apply updates
	if cmd.Name != nil {
		if err := tenant.UpdateName(*cmd.Name); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	if cmd.ContactEmail != nil {
		if err := tenant.UpdateContactEmail(*cmd.ContactEmail); err != nil {
			return nil, fmt.Errorf("failed to update email: %w", err)
		}
	}

	if cmd.ContactName != nil {
		tenant.PrimaryContactName = *cmd.ContactName
	}

	if cmd.BillingEmail != nil {
		tenant.BillingEmail = *cmd.BillingEmail
	}

	if cmd.Settings != nil {
		// Merge settings (don't replace entirely)
		if tenant.Settings == nil {
			tenant.Settings = make(map[string]interface{})
		}
		for k, v := range cmd.Settings {
			tenant.Settings[k] = v
		}
	}

	// Update tenant
	if err := uc.repo.Update(ctx, tenant); err != nil {
		uc.logger.Error("Failed to update tenant",
			zap.String("tenant_id", cmd.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	uc.logger.Info("Tenant updated",
		zap.String("tenant_id", cmd.TenantID.String()),
	)

	return tenant, nil
}
