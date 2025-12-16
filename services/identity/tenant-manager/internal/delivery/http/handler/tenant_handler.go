package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/cotai/tenant-manager/internal/delivery/http/dto"
	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/cotai/tenant-manager/internal/usecase"
)

// TenantHandler handles tenant HTTP requests
type TenantHandler struct {
	createTenantUC  *usecase.CreateTenantUseCase
	getTenantUC     *usecase.GetTenantUseCase
	listTenantsUC   *usecase.ListTenantsUseCase
	updateTenantUC  *usecase.UpdateTenantUseCase
	suspendTenantUC *usecase.SuspendTenantUseCase
	activateTenantUC *usecase.ActivateTenantUseCase
	deleteTenantUC  *usecase.DeleteTenantUseCase
	validator       *validator.Validate
	logger          *zap.Logger
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(
	createTenantUC *usecase.CreateTenantUseCase,
	getTenantUC *usecase.GetTenantUseCase,
	listTenantsUC *usecase.ListTenantsUseCase,
	updateTenantUC *usecase.UpdateTenantUseCase,
	suspendTenantUC *usecase.SuspendTenantUseCase,
	activateTenantUC *usecase.ActivateTenantUseCase,
	deleteTenantUC *usecase.DeleteTenantUseCase,
	logger *zap.Logger,
) *TenantHandler {
	return &TenantHandler{
		createTenantUC:   createTenantUC,
		getTenantUC:      getTenantUC,
		listTenantsUC:    listTenantsUC,
		updateTenantUC:   updateTenantUC,
		suspendTenantUC:  suspendTenantUC,
		activateTenantUC: activateTenantUC,
		deleteTenantUC:   deleteTenantUC,
		validator:        validator.New(),
		logger:           logger,
	}
}

// CreateTenant creates a new tenant
// POST /api/v1/tenants
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req dto.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload", nil)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := h.parseValidationErrors(err.(validator.ValidationErrors))
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", validationErrors)
		return
	}

	// Convert DTO to use case command
	cmd := usecase.CreateTenantCommand{
		Name:       req.Name,
		Slug:       req.Slug,
		Plan:       req.ToTenantPlan(),
		AdminEmail: req.AdminEmail,
		AdminName:  req.AdminName,
		Settings:   req.Settings,
	}

	// Execute use case
	result, err := h.createTenantUC.Execute(ctx, cmd)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert domain to response DTO
	response := dto.FromDomain(result.Tenant)

	h.logger.Info("Tenant created successfully",
		zap.String("tenant_id", result.Tenant.TenantID.String()),
		zap.String("slug", result.Tenant.TenantSlug),
		zap.Duration("provisioning_duration", result.ProvisioningDuration),
	)

	h.respondSuccess(w, http.StatusCreated, response)
}

// ListTenants lists tenants with pagination
// GET /api/v1/tenants?page=1&pageSize=20&status=active&plan=professional&search=acme
func (h *TenantHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := dto.ParseListTenantsQuery(r)

	// Convert to use case query
	ucQuery := usecase.ListTenantsQuery{
		Page:     query.Page,
		PerPage:  query.PageSize,
		Status:   query.ToTenantStatus(),
		PlanTier: query.ToTenantPlan(),
		Search:   query.Search,
	}

	// Execute use case
	result, err := h.listTenantsUC.Execute(ctx, ucQuery)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to response DTO
	response := dto.NewListTenantsResponse(result.Tenants, result.Total, query.Page, query.PageSize)

	h.respondSuccess(w, http.StatusOK, response)
}

// GetTenant retrieves a single tenant by ID
// GET /api/v1/tenants/{id}
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse tenant ID from URL
	idParam := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tenant ID format", nil)
		return
	}

	// Execute use case
	tenant, err := h.getTenantUC.ExecuteByID(ctx, tenantID)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to response DTO
	response := dto.FromDomain(tenant)

	h.respondSuccess(w, http.StatusOK, response)
}

// UpdateTenant updates tenant information
// PATCH /api/v1/tenants/{id}
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse tenant ID
	idParam := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tenant ID format", nil)
		return
	}

	// Parse request
	var req dto.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload", nil)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := h.parseValidationErrors(err.(validator.ValidationErrors))
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", validationErrors)
		return
	}

	// Convert to use case command
	cmd := usecase.UpdateTenantCommand{
		TenantID:     tenantID,
		Name:         req.Name,
		ContactEmail: req.ContactEmail,
		ContactName:  req.ContactName,
		BillingEmail: req.BillingEmail,
		Settings:     req.Settings,
	}

	// Execute use case
	tenant, err := h.updateTenantUC.Execute(ctx, cmd)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to response DTO
	response := dto.FromDomain(tenant)

	h.logger.Info("Tenant updated successfully",
		zap.String("tenant_id", tenant.TenantID.String()),
	)

	h.respondSuccess(w, http.StatusOK, response)
}

// SuspendTenant suspends a tenant
// POST /api/v1/tenants/{id}/suspend
func (h *TenantHandler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse tenant ID
	idParam := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tenant ID format", nil)
		return
	}

	// Parse request
	var req dto.SuspendTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON payload", nil)
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		validationErrors := h.parseValidationErrors(err.(validator.ValidationErrors))
		h.respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", validationErrors)
		return
	}

	// Convert to use case command
	cmd := usecase.SuspendTenantCommand{
		TenantID: tenantID,
		Reason:   req.Reason,
	}

	// Execute use case
	tenant, err := h.suspendTenantUC.Execute(ctx, cmd)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to response DTO
	response := dto.FromDomain(tenant)

	h.logger.Warn("Tenant suspended",
		zap.String("tenant_id", tenant.TenantID.String()),
		zap.String("reason", req.Reason),
	)

	h.respondSuccess(w, http.StatusOK, response)
}

// ActivateTenant activates or reactivates a tenant
// POST /api/v1/tenants/{id}/activate
func (h *TenantHandler) ActivateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse tenant ID
	idParam := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tenant ID format", nil)
		return
	}

	// Convert to use case command
	cmd := usecase.ActivateTenantCommand{
		TenantID: tenantID,
	}

	// Execute use case
	tenant, err := h.activateTenantUC.Execute(ctx, cmd)
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	// Convert to response DTO
	response := dto.FromDomain(tenant)

	h.logger.Info("Tenant activated",
		zap.String("tenant_id", tenant.TenantID.String()),
	)

	h.respondSuccess(w, http.StatusOK, response)
}

// DeleteTenant soft deletes a tenant
// DELETE /api/v1/tenants/{id}
func (h *TenantHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse tenant ID
	idParam := chi.URLParam(r, "id")
	tenantID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid tenant ID format", nil)
		return
	}

	// Convert to use case command
	cmd := usecase.DeleteTenantCommand{
		TenantID: tenantID,
	}

	// Execute use case
	if err := h.deleteTenantUC.Execute(ctx, cmd); err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	h.logger.Warn("Tenant deleted",
		zap.String("tenant_id", tenantID.String()),
	)

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// handleUseCaseError maps domain errors to HTTP responses
func (h *TenantHandler) handleUseCaseError(w http.ResponseWriter, err error) {
	h.logger.Error("Use case error", zap.Error(err))

	switch err {
	case domain.ErrTenantNotFound:
		h.respondError(w, http.StatusNotFound, "TENANT_NOT_FOUND", "Tenant not found", nil)
	case domain.ErrSlugAlreadyExists:
		h.respondError(w, http.StatusConflict, "SLUG_EXISTS", "Tenant slug already exists", nil)
	case domain.ErrTenantDeleted:
		h.respondError(w, http.StatusGone, "TENANT_DELETED", "Tenant has been deleted", nil)
	case domain.ErrInvalidPlanTier:
		h.respondError(w, http.StatusBadRequest, "INVALID_PLAN", "Invalid plan tier", nil)
	case domain.ErrInvalidTenantName:
		h.respondError(w, http.StatusBadRequest, "INVALID_NAME", "Invalid tenant name", nil)
	case domain.ErrInvalidSlug:
		h.respondError(w, http.StatusBadRequest, "INVALID_SLUG", "Invalid tenant slug", nil)
	case domain.ErrInvalidEmail:
		h.respondError(w, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email address", nil)
	case domain.ErrTenantAlreadyActive:
		h.respondError(w, http.StatusConflict, "ALREADY_ACTIVE", "Tenant is already active", nil)
	case domain.ErrTenantAlreadySuspended:
		h.respondError(w, http.StatusConflict, "ALREADY_SUSPENDED", "Tenant is already suspended", nil)
	case domain.ErrCannotSuspendDeletedTenant:
		h.respondError(w, http.StatusConflict, "CANNOT_SUSPEND_DELETED", "Cannot suspend deleted tenant", nil)
	case context.Canceled:
		h.respondError(w, http.StatusRequestTimeout, "REQUEST_CANCELED", "Request was canceled", nil)
	case context.DeadlineExceeded:
		h.respondError(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request timeout", nil)
	default:
		h.respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
	}
}

// parseValidationErrors converts validator errors to field errors
func (h *TenantHandler) parseValidationErrors(errs validator.ValidationErrors) []dto.FieldError {
	fieldErrors := make([]dto.FieldError, 0, len(errs))

	for _, err := range errs {
		fieldErrors = append(fieldErrors, dto.FieldError{
			Field:   err.Field(),
			Message: h.getValidationErrorMessage(err),
		})
	}

	return fieldErrors
}

// getValidationErrorMessage generates user-friendly validation messages
func (h *TenantHandler) getValidationErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + err.Param() + " characters"
	case "max":
		return "Must be at most " + err.Param() + " characters"
	case "lowercase":
		return "Must be lowercase"
	case "alphanum_hyphen":
		return "Must contain only alphanumeric characters and hyphens"
	case "oneof":
		return "Must be one of: " + err.Param()
	default:
		return "Invalid value"
	}
}

// respondSuccess sends a success response
func (h *TenantHandler) respondSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.SuccessResponse{
		Data: data,
	}

	json.NewEncoder(w).Encode(response)
}

// respondError sends an error response
func (h *TenantHandler) respondError(w http.ResponseWriter, status int, code, message string, details []dto.FieldError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}

	json.NewEncoder(w).Encode(response)
}
