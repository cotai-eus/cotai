package domain

import "errors"

// Domain errors
var (
	// Validation errors
	ErrEmptyTenantName       = errors.New("tenant name cannot be empty")
	ErrInvalidTenantName     = errors.New("invalid tenant name")
	ErrTenantNameTooLong     = errors.New("tenant name cannot exceed 255 characters")
	ErrEmptyTenantSlug       = errors.New("tenant slug cannot be empty")
	ErrInvalidSlug           = errors.New("invalid tenant slug")
	ErrTenantSlugTooLong     = errors.New("tenant slug cannot exceed 100 characters")
	ErrInvalidTenantSlug     = errors.New("tenant slug must contain only lowercase letters, numbers, and hyphens")
	ErrEmptyEmail            = errors.New("email cannot be empty")
	ErrEmailTooLong          = errors.New("email cannot exceed 255 characters")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidPlanTier       = errors.New("invalid plan tier")

	// Business logic errors
	ErrTenantNotFound              = errors.New("tenant not found")
	ErrTenantAlreadyExists         = errors.New("tenant already exists")
	ErrSlugAlreadyExists           = errors.New("tenant slug already exists")
	ErrTenantDeleted               = errors.New("tenant is deleted")
	ErrTenantAlreadyActive         = errors.New("tenant is already active")
	ErrTenantAlreadySuspended      = errors.New("tenant is already suspended")
	ErrTenantAlreadyDeleted        = errors.New("tenant is already deleted")
	ErrCannotSuspendDeletedTenant  = errors.New("cannot suspend deleted tenant")
	ErrPlanAlreadySet              = errors.New("tenant already has this plan")

	// Repository errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrTransactionFailed  = errors.New("transaction failed")
	ErrQueryFailed        = errors.New("query execution failed")

	// Provisioning errors
	ErrSchemaCreationFailed   = errors.New("schema creation failed")
	ErrMigrationFailed        = errors.New("migration failed")
	ErrRLSEnablementFailed    = errors.New("RLS enablement failed")
	ErrProvisioningFailed     = errors.New("provisioning failed")
)

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrTenantNotFound)
}

// IsAlreadyExistsError checks if error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrTenantAlreadyExists) || errors.Is(err, ErrSlugAlreadyExists)
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrEmptyTenantName) ||
		errors.Is(err, ErrTenantNameTooLong) ||
		errors.Is(err, ErrEmptyTenantSlug) ||
		errors.Is(err, ErrTenantSlugTooLong) ||
		errors.Is(err, ErrInvalidTenantSlug) ||
		errors.Is(err, ErrEmptyEmail) ||
		errors.Is(err, ErrEmailTooLong) ||
		errors.Is(err, ErrInvalidEmail) ||
		errors.Is(err, ErrInvalidPlanTier)
}
