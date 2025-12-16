package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateTenantCommand represents the input for creating a tenant
type CreateTenantCommand struct {
	Name       string
	Slug       string
	Plan       domain.PlanTier
	AdminEmail string
	AdminName  string
	Settings   map[string]interface{}
}

// CreateTenantResult represents the output of creating a tenant
type CreateTenantResult struct {
	Tenant       *domain.Tenant
	SchemaName   string
	ProvisioningDuration time.Duration
}

// CreateTenantUseCase handles tenant creation with full orchestration
type CreateTenantUseCase struct {
	repo        domain.TenantRepository
	provisioner SchemaProvisioner
	publisher   EventPublisher
	logger      *zap.Logger
}

// SchemaProvisioner interface for schema provisioning
type SchemaProvisioner interface {
	ProvisionTenant(ctx context.Context, tenantID uuid.UUID) error
	SchemaExists(ctx context.Context, tenantID uuid.UUID) (bool, error)
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	PublishTenantCreated(ctx context.Context, tenant *domain.Tenant) error
	PublishTenantUpdated(ctx context.Context, tenant *domain.Tenant) error
	PublishTenantSuspended(ctx context.Context, tenant *domain.Tenant) error
	PublishTenantActivated(ctx context.Context, tenant *domain.Tenant) error
	PublishTenantDeleted(ctx context.Context, tenant *domain.Tenant) error
}

// NewCreateTenantUseCase creates a new CreateTenantUseCase
func NewCreateTenantUseCase(
	repo domain.TenantRepository,
	provisioner SchemaProvisioner,
	publisher EventPublisher,
	logger *zap.Logger,
) *CreateTenantUseCase {
	return &CreateTenantUseCase{
		repo:        repo,
		provisioner: provisioner,
		publisher:   publisher,
		logger:      logger,
	}
}

// Execute executes the create tenant use case
func (uc *CreateTenantUseCase) Execute(ctx context.Context, cmd CreateTenantCommand) (*CreateTenantResult, error) {
	startTime := time.Now()

	uc.logger.Info("Creating tenant",
		zap.String("name", cmd.Name),
		zap.String("slug", cmd.Slug),
		zap.String("plan", string(cmd.Plan)),
	)

	// Step 1: Validate command
	if err := uc.validateCommand(cmd); err != nil {
		uc.logger.Error("Invalid create tenant command", zap.Error(err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Step 2: Check slug uniqueness
	exists, err := uc.repo.ExistsBySlug(ctx, cmd.Slug)
	if err != nil {
		uc.logger.Error("Failed to check slug existence", zap.Error(err))
		return nil, fmt.Errorf("failed to check slug: %w", err)
	}
	if exists {
		return nil, domain.ErrSlugAlreadyExists
	}

	// Step 3: Create tenant entity (status: provisioning)
	tenant, err := domain.NewTenant(cmd.Name, cmd.Slug, cmd.Plan, cmd.AdminEmail)
	if err != nil {
		uc.logger.Error("Failed to create tenant entity", zap.Error(err))
		return nil, fmt.Errorf("failed to create tenant entity: %w", err)
	}

	// Apply additional settings
	if cmd.AdminName != "" {
		tenant.PrimaryContactName = cmd.AdminName
	}
	if cmd.Settings != nil {
		tenant.Settings = cmd.Settings
	}

	// Step 4: Insert tenant record into database
	if err := uc.repo.Create(ctx, tenant); err != nil {
		uc.logger.Error("Failed to insert tenant into database",
			zap.String("tenant_id", tenant.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create tenant record: %w", err)
	}

	uc.logger.Info("Tenant record created in database",
		zap.String("tenant_id", tenant.TenantID.String()),
	)

	// Step 5: Provision schema (critical operation)
	provisionStart := time.Now()
	if err := uc.provisioner.ProvisionTenant(ctx, tenant.TenantID); err != nil {
		uc.logger.Error("Schema provisioning failed",
			zap.String("tenant_id", tenant.TenantID.String()),
			zap.Error(err),
		)

		// Try to rollback: mark tenant as failed
		// In production, you might want to implement retry logic or manual intervention
		return nil, fmt.Errorf("schema provisioning failed: %w", err)
	}
	provisioningDuration := time.Since(provisionStart)

	uc.logger.Info("Schema provisioning completed",
		zap.String("tenant_id", tenant.TenantID.String()),
		zap.Duration("duration", provisioningDuration),
	)

	// Step 6: Activate tenant
	if err := tenant.Activate(); err != nil {
		uc.logger.Error("Failed to activate tenant", zap.Error(err))
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}

	// Step 7: Update tenant status to active
	if err := uc.repo.Update(ctx, tenant); err != nil {
		uc.logger.Error("Failed to update tenant status",
			zap.String("tenant_id", tenant.TenantID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	uc.logger.Info("Tenant activated",
		zap.String("tenant_id", tenant.TenantID.String()),
	)

	// Step 8: Publish event (async, non-blocking)
	// We don't fail the operation if event publishing fails
	go func() {
		publishCtx := context.Background()
		if err := uc.publisher.PublishTenantCreated(publishCtx, tenant); err != nil {
			uc.logger.Error("Failed to publish tenant.created event",
				zap.String("tenant_id", tenant.TenantID.String()),
				zap.Error(err),
			)
		}
	}()

	totalDuration := time.Since(startTime)
	uc.logger.Info("Tenant creation completed",
		zap.String("tenant_id", tenant.TenantID.String()),
		zap.Duration("total_duration", totalDuration),
	)

	return &CreateTenantResult{
		Tenant:               tenant,
		SchemaName:           tenant.DatabaseSchema,
		ProvisioningDuration: provisioningDuration,
	}, nil
}

// validateCommand validates the create tenant command
func (uc *CreateTenantUseCase) validateCommand(cmd CreateTenantCommand) error {
	if cmd.Name == "" {
		return domain.ErrEmptyTenantName
	}
	if cmd.Slug == "" {
		return domain.ErrEmptyTenantSlug
	}
	if !cmd.Plan.IsValid() {
		return domain.ErrInvalidPlanTier
	}
	if cmd.AdminEmail == "" {
		return domain.ErrEmptyEmail
	}
	return nil
}
