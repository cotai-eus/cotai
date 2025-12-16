package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TenantStatus represents the lifecycle state of a tenant
type TenantStatus string

const (
	StatusProvisioning TenantStatus = "provisioning"
	StatusActive       TenantStatus = "active"
	StatusSuspended    TenantStatus = "suspended"
	StatusArchived     TenantStatus = "archived"
	StatusDeleted      TenantStatus = "deleted"
)

// IsValid checks if the tenant status is valid
func (s TenantStatus) IsValid() bool {
	switch s {
	case StatusProvisioning, StatusActive, StatusSuspended, StatusArchived, StatusDeleted:
		return true
	default:
		return false
	}
}

// PlanTier represents the subscription plan
type PlanTier string

const (
	PlanFree         PlanTier = "free"
	PlanBasic        PlanTier = "basic"
	PlanProfessional PlanTier = "professional"
	PlanEnterprise   PlanTier = "enterprise"
)

// IsValid checks if the plan tier is valid
func (p PlanTier) IsValid() bool {
	switch p {
	case PlanFree, PlanBasic, PlanProfessional, PlanEnterprise:
		return true
	default:
		return false
	}
}

// Tenant represents the tenant aggregate root (DDD)
type Tenant struct {
	// Primary key
	ID uuid.UUID `db:"id"`

	// Business identifiers
	TenantID   uuid.UUID `db:"tenant_id"`
	TenantName string    `db:"tenant_name"`
	TenantSlug string    `db:"tenant_slug"`

	// Schema information
	DatabaseSchema string `db:"database_schema"`
	SchemaVersion  string `db:"schema_version"`

	// Status and plan
	Status   TenantStatus `db:"status"`
	PlanTier PlanTier     `db:"plan_tier"`

	// Quotas
	MaxUsers     int `db:"max_users"`
	MaxStorageGB int `db:"max_storage_gb"`

	// Contact information
	PrimaryContactEmail string `db:"primary_contact_email"`
	PrimaryContactName  string `db:"primary_contact_name"`
	BillingEmail        string `db:"billing_email"`

	// JSONB fields
	Settings map[string]interface{} `db:"settings"`
	Features map[string]interface{} `db:"features"`

	// Audit fields
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	ActivatedAt *time.Time `db:"activated_at"`
	SuspendedAt *time.Time `db:"suspended_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	CreatedBy   *uuid.UUID `db:"created_by"`
	UpdatedBy   *uuid.UUID `db:"updated_by"`
}

// NewTenant creates a new tenant with default values
func NewTenant(name, slug string, plan PlanTier, email string) (*Tenant, error) {
	if err := validateTenantName(name); err != nil {
		return nil, err
	}

	if err := validateTenantSlug(slug); err != nil {
		return nil, err
	}

	if !plan.IsValid() {
		return nil, ErrInvalidPlanTier
	}

	if err := validateEmail(email); err != nil {
		return nil, err
	}

	tenantID := uuid.New()
	now := time.Now()

	return &Tenant{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		TenantName:          name,
		TenantSlug:          slug,
		DatabaseSchema:      FormatSchemaName(tenantID),
		SchemaVersion:       "1.0.0",
		Status:              StatusProvisioning,
		PlanTier:            plan,
		MaxUsers:            getDefaultMaxUsers(plan),
		MaxStorageGB:        getDefaultMaxStorage(plan),
		PrimaryContactEmail: email,
		BillingEmail:        email,
		Settings:            make(map[string]interface{}),
		Features:            make(map[string]interface{}),
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// FormatSchemaName formats tenant ID into PostgreSQL schema name
// Example: "550e8400-e29b-41d4-a716-446655440000" -> "tenant_550e8400e29b41d4a716446655440000"
func FormatSchemaName(tenantID uuid.UUID) string {
	cleanID := strings.ReplaceAll(tenantID.String(), "-", "")
	return fmt.Sprintf("tenant_%s", cleanID)
}

// Business Methods

// Activate activates a tenant
func (t *Tenant) Activate() error {
	if t.Status == StatusDeleted {
		return ErrTenantDeleted
	}

	if t.Status == StatusActive {
		return ErrTenantAlreadyActive
	}

	t.Status = StatusActive
	now := time.Now()
	t.ActivatedAt = &now
	t.UpdatedAt = now

	return nil
}

// Suspend suspends a tenant
func (t *Tenant) Suspend(reason string) error {
	if t.Status == StatusDeleted {
		return ErrTenantDeleted
	}

	if t.Status == StatusSuspended {
		return ErrTenantAlreadySuspended
	}

	t.Status = StatusSuspended
	now := time.Now()
	t.SuspendedAt = &now
	t.UpdatedAt = now

	// Store suspension reason in settings
	if t.Settings == nil {
		t.Settings = make(map[string]interface{})
	}
	t.Settings["suspension_reason"] = reason

	return nil
}

// Delete soft-deletes a tenant
func (t *Tenant) Delete() error {
	if t.Status == StatusDeleted {
		return ErrTenantAlreadyDeleted
	}

	t.Status = StatusDeleted
	now := time.Now()
	t.DeletedAt = &now
	t.UpdatedAt = now

	return nil
}

// UpdateName updates the tenant name
func (t *Tenant) UpdateName(name string) error {
	if err := validateTenantName(name); err != nil {
		return err
	}

	t.TenantName = name
	t.UpdatedAt = time.Now()

	return nil
}

// UpdateContactEmail updates the primary contact email
func (t *Tenant) UpdateContactEmail(email string) error {
	if err := validateEmail(email); err != nil {
		return err
	}

	t.PrimaryContactEmail = email
	t.UpdatedAt = time.Now()

	return nil
}

// ChangePlan changes the subscription plan
func (t *Tenant) ChangePlan(newPlan PlanTier) error {
	if !newPlan.IsValid() {
		return ErrInvalidPlanTier
	}

	if t.PlanTier == newPlan {
		return ErrPlanAlreadySet
	}

	oldPlan := t.PlanTier
	t.PlanTier = newPlan
	t.MaxUsers = getDefaultMaxUsers(newPlan)
	t.MaxStorageGB = getDefaultMaxStorage(newPlan)
	t.UpdatedAt = time.Now()

	// Store plan change history in settings
	if t.Settings == nil {
		t.Settings = make(map[string]interface{})
	}
	t.Settings["plan_changed_from"] = oldPlan
	t.Settings["plan_changed_at"] = time.Now().Format(time.RFC3339)

	return nil
}

// Query Methods

// CanProvision checks if tenant can be provisioned
func (t *Tenant) CanProvision() bool {
	return t.Status == StatusProvisioning
}

// IsActive checks if tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive
}

// IsSuspended checks if tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == StatusSuspended
}

// IsDeleted checks if tenant is deleted
func (t *Tenant) IsDeleted() bool {
	return t.Status == StatusDeleted
}

// Helper functions

func getDefaultMaxUsers(plan PlanTier) int {
	switch plan {
	case PlanFree:
		return 5
	case PlanBasic:
		return 20
	case PlanProfessional:
		return 100
	case PlanEnterprise:
		return 1000
	default:
		return 5
	}
}

func getDefaultMaxStorage(plan PlanTier) int {
	switch plan {
	case PlanFree:
		return 5 // GB
	case PlanBasic:
		return 50
	case PlanProfessional:
		return 500
	case PlanEnterprise:
		return 5000
	default:
		return 5
	}
}

// Validation functions

func validateTenantName(name string) error {
	if len(name) == 0 {
		return ErrEmptyTenantName
	}

	if len(name) > 255 {
		return ErrTenantNameTooLong
	}

	return nil
}

func validateTenantSlug(slug string) error {
	if len(slug) == 0 {
		return ErrEmptyTenantSlug
	}

	if len(slug) > 100 {
		return ErrTenantSlugTooLong
	}

	// Slug must be lowercase alphanumeric with hyphens
	for _, ch := range slug {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return ErrInvalidTenantSlug
		}
	}

	return nil
}

func validateEmail(email string) error {
	if len(email) == 0 {
		return ErrEmptyEmail
	}

	if len(email) > 255 {
		return ErrEmailTooLong
	}

	// Basic email validation (contains @)
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}

	return nil
}
