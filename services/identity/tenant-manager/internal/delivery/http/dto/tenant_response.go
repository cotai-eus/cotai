package dto

import (
	"time"

	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/google/uuid"
)

// TenantResponse represents a tenant in API responses
type TenantResponse struct {
	ID                  uuid.UUID              `json:"id"`
	TenantID            uuid.UUID              `json:"tenantId"`
	Name                string                 `json:"name"`
	Slug                string                 `json:"slug"`
	SchemaName          string                 `json:"schemaName"`
	Status              string                 `json:"status"`
	Plan                string                 `json:"plan"`
	MaxUsers            int                    `json:"maxUsers"`
	MaxStorageGB        int                    `json:"maxStorageGb"`
	PrimaryContactEmail string                 `json:"primaryContactEmail,omitempty"`
	PrimaryContactName  string                 `json:"primaryContactName,omitempty"`
	Settings            map[string]interface{} `json:"settings,omitempty"`
	Features            map[string]interface{} `json:"features,omitempty"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
	ActivatedAt         *time.Time             `json:"activatedAt,omitempty"`
	SuspendedAt         *time.Time             `json:"suspendedAt,omitempty"`
}

// FromDomain converts domain.Tenant to TenantResponse
func FromDomain(tenant *domain.Tenant) *TenantResponse {
	return &TenantResponse{
		ID:                  tenant.ID,
		TenantID:            tenant.TenantID,
		Name:                tenant.TenantName,
		Slug:                tenant.TenantSlug,
		SchemaName:          tenant.DatabaseSchema,
		Status:              string(tenant.Status),
		Plan:                string(tenant.PlanTier),
		MaxUsers:            tenant.MaxUsers,
		MaxStorageGB:        tenant.MaxStorageGB,
		PrimaryContactEmail: tenant.PrimaryContactEmail,
		PrimaryContactName:  tenant.PrimaryContactName,
		Settings:            tenant.Settings,
		Features:            tenant.Features,
		CreatedAt:           tenant.CreatedAt,
		UpdatedAt:           tenant.UpdatedAt,
		ActivatedAt:         tenant.ActivatedAt,
		SuspendedAt:         tenant.SuspendedAt,
	}
}

// ListTenantsResponse represents paginated tenant list
type ListTenantsResponse struct {
	Data  []*TenantResponse `json:"data"`
	Meta  PaginationMeta    `json:"meta"`
	Links PaginationLinks   `json:"links,omitempty"`
}

// NewListTenantsResponse creates a new paginated response
func NewListTenantsResponse(tenants []*domain.Tenant, total, page, pageSize int) *ListTenantsResponse {
	data := make([]*TenantResponse, 0, len(tenants))
	for _, tenant := range tenants {
		data = append(data, FromDomain(tenant))
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return &ListTenantsResponse{
		Data: data,
		Meta: PaginationMeta{
			Page:       page,
			PerPage:    pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// PaginationLinks represents pagination links
type PaginationLinks struct {
	Self  string `json:"self,omitempty"`
	First string `json:"first,omitempty"`
	Last  string `json:"last,omitempty"`
	Next  string `json:"next,omitempty"`
	Prev  string `json:"prev,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldError  `json:"details,omitempty"`
}

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Service   string            `json:"service"`
	Version   string            `json:"version"`
	Checks    map[string]string `json:"checks,omitempty"`
}
