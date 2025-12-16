package dto

import (
	"net/http"
	"strconv"

	"github.com/cotai/tenant-manager/internal/domain"
)

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	Name       string                 `json:"name" validate:"required,min=3,max=255"`
	Slug       string                 `json:"slug" validate:"required,min=2,max=100,lowercase,alphanum_hyphen"`
	Plan       string                 `json:"plan" validate:"required,oneof=free basic professional enterprise"`
	AdminEmail string                 `json:"adminEmail" validate:"required,email"`
	AdminName  string                 `json:"adminName,omitempty" validate:"omitempty,max=255"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
}

// ToTenantPlan converts string to domain.PlanTier
func (r *CreateTenantRequest) ToTenantPlan() domain.PlanTier {
	return domain.PlanTier(r.Plan)
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name         *string                `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	ContactEmail *string                `json:"contactEmail,omitempty" validate:"omitempty,email"`
	ContactName  *string                `json:"contactName,omitempty" validate:"omitempty,max=255"`
	BillingEmail *string                `json:"billingEmail,omitempty" validate:"omitempty,email"`
	Settings     map[string]interface{} `json:"settings,omitempty"`
}

// SuspendTenantRequest represents the request to suspend a tenant
type SuspendTenantRequest struct {
	Reason string `json:"reason" validate:"required,min=10,max=500"`
}

// ListTenantsQuery represents query parameters for listing tenants
type ListTenantsQuery struct {
	Page     int    `json:"page" validate:"omitempty,min=1"`
	PageSize int    `json:"pageSize" validate:"omitempty,min=1,max=100"`
	Status   string `json:"status" validate:"omitempty,oneof=provisioning active suspended archived deleted"`
	Plan     string `json:"plan" validate:"omitempty,oneof=free basic professional enterprise"`
	Search   string `json:"search" validate:"omitempty,max=255"`
}

// ParseListTenantsQuery parses query parameters from HTTP request
func ParseListTenantsQuery(r *http.Request) ListTenantsQuery {
	query := ListTenantsQuery{
		Page:     1,
		PageSize: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			query.PageSize = ps
		}
	}

	query.Status = r.URL.Query().Get("status")
	query.Plan = r.URL.Query().Get("plan")
	query.Search = r.URL.Query().Get("search")

	return query
}

// ToTenantStatus converts string to domain.TenantStatus
func (q *ListTenantsQuery) ToTenantStatus() domain.TenantStatus {
	return domain.TenantStatus(q.Status)
}

// ToTenantPlan converts string to domain.PlanTier
func (q *ListTenantsQuery) ToTenantPlan() domain.PlanTier {
	return domain.PlanTier(q.Plan)
}
