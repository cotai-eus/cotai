package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewTenant(t *testing.T) {
	tests := []struct {
		name      string
		tenantName string
		slug      string
		plan      PlanTier
		email     string
		wantErr   error
	}{
		{
			name:       "valid tenant",
			tenantName: "Test Company",
			slug:       "test-company",
			plan:       PlanProfessional,
			email:      "admin@test.com",
			wantErr:    nil,
		},
		{
			name:       "empty name",
			tenantName: "",
			slug:       "test",
			plan:       PlanProfessional,
			email:      "admin@test.com",
			wantErr:    ErrEmptyTenantName,
		},
		{
			name:       "empty slug",
			tenantName: "Test Company",
			slug:       "",
			plan:       PlanProfessional,
			email:      "admin@test.com",
			wantErr:    ErrEmptyTenantSlug,
		},
		{
			name:       "invalid slug with uppercase",
			tenantName: "Test Company",
			slug:       "Test-Company",
			plan:       PlanProfessional,
			email:      "admin@test.com",
			wantErr:    ErrInvalidTenantSlug,
		},
		{
			name:       "invalid email",
			tenantName: "Test Company",
			slug:       "test-company",
			plan:       PlanProfessional,
			email:      "invalid-email",
			wantErr:    ErrInvalidEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := NewTenant(tt.tenantName, tt.slug, tt.plan, tt.email)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, tenant)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tenant)
				assert.Equal(t, tt.tenantName, tenant.TenantName)
				assert.Equal(t, tt.slug, tenant.TenantSlug)
				assert.Equal(t, tt.plan, tenant.PlanTier)
				assert.Equal(t, tt.email, tenant.PrimaryContactEmail)
				assert.Equal(t, StatusProvisioning, tenant.Status)
				assert.NotEqual(t, uuid.Nil, tenant.TenantID)
				assert.NotEqual(t, uuid.Nil, tenant.ID)
			}
		})
	}
}

func TestTenant_Activate(t *testing.T) {
	tenant, _ := NewTenant("Test Company", "test-company", PlanProfessional, "admin@test.com")

	// First activation should succeed
	err := tenant.Activate()
	assert.NoError(t, err)
	assert.Equal(t, StatusActive, tenant.Status)
	assert.NotNil(t, tenant.ActivatedAt)

	// Second activation should fail
	err = tenant.Activate()
	assert.ErrorIs(t, err, ErrTenantAlreadyActive)
}

func TestTenant_Suspend(t *testing.T) {
	tenant, _ := NewTenant("Test Company", "test-company", PlanProfessional, "admin@test.com")
	tenant.Activate()

	// Suspend tenant
	reason := "Payment overdue"
	err := tenant.Suspend(reason)
	assert.NoError(t, err)
	assert.Equal(t, StatusSuspended, tenant.Status)
	assert.NotNil(t, tenant.SuspendedAt)
	assert.Equal(t, reason, tenant.Settings["suspension_reason"])

	// Second suspension should fail
	err = tenant.Suspend("Another reason")
	assert.ErrorIs(t, err, ErrTenantAlreadySuspended)
}

func TestTenant_Delete(t *testing.T) {
	tenant, _ := NewTenant("Test Company", "test-company", PlanProfessional, "admin@test.com")
	tenant.Activate()

	// Delete tenant
	err := tenant.Delete()
	assert.NoError(t, err)
	assert.Equal(t, StatusDeleted, tenant.Status)
	assert.NotNil(t, tenant.DeletedAt)

	// Second deletion should fail
	err = tenant.Delete()
	assert.ErrorIs(t, err, ErrTenantAlreadyDeleted)
}

func TestTenant_ChangePlan(t *testing.T) {
	tenant, _ := NewTenant("Test Company", "test-company", PlanBasic, "admin@test.com")

	// Change to professional plan
	err := tenant.ChangePlan(PlanProfessional)
	assert.NoError(t, err)
	assert.Equal(t, PlanProfessional, tenant.PlanTier)
	assert.Equal(t, 100, tenant.MaxUsers)
	assert.Equal(t, 500, tenant.MaxStorageGB)

	// Change to same plan should fail
	err = tenant.ChangePlan(PlanProfessional)
	assert.ErrorIs(t, err, ErrPlanAlreadySet)
}

func TestFormatSchemaName(t *testing.T) {
	tenantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	schemaName := FormatSchemaName(tenantID)

	expected := "tenant_550e8400e29b41d4a716446655440000"
	assert.Equal(t, expected, schemaName)
	assert.NotContains(t, schemaName, "-")
}

func TestPlanTier_IsValid(t *testing.T) {
	tests := []struct {
		plan  PlanTier
		valid bool
	}{
		{PlanFree, true},
		{PlanBasic, true},
		{PlanProfessional, true},
		{PlanEnterprise, true},
		{PlanTier("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.plan), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.plan.IsValid())
		})
	}
}

func TestTenantStatus_IsValid(t *testing.T) {
	tests := []struct {
		status TenantStatus
		valid  bool
	}{
		{StatusProvisioning, true},
		{StatusActive, true},
		{StatusSuspended, true},
		{StatusArchived, true},
		{StatusDeleted, true},
		{TenantStatus("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}
