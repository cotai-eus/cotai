package messaging

import (
	"time"

	"github.com/cotai/tenant-manager/internal/domain"
)

// EventType represents the type of tenant lifecycle event
type EventType string

const (
	EventTenantCreated     EventType = "tenant.created"
	EventTenantActivated   EventType = "tenant.activated"
	EventTenantSuspended   EventType = "tenant.suspended"
	EventTenantDeleted     EventType = "tenant.deleted"
	EventTenantPlanChanged EventType = "tenant.plan.changed"
	EventTenantUpdated     EventType = "tenant.updated"
)

// TenantLifecycleEvent represents a tenant lifecycle event
type TenantLifecycleEvent struct {
	EventID       string                 `json:"eventId"`
	EventType     EventType              `json:"eventType"`
	TenantID      string                 `json:"tenantId"`
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlationId"`
	Payload       map[string]interface{} `json:"payload"`
}

// TenantToEventPayload converts a domain tenant to event payload
func TenantToEventPayload(tenant *domain.Tenant) map[string]interface{} {
	return map[string]interface{}{
		"id":           tenant.ID.String(),
		"tenantId":     tenant.TenantID.String(),
		"name":         tenant.TenantName,
		"slug":         tenant.TenantSlug,
		"schemaName":   tenant.DatabaseSchema,
		"status":       string(tenant.Status),
		"plan":         string(tenant.PlanTier),
		"contactEmail": tenant.PrimaryContactEmail,
		"contactName":  tenant.PrimaryContactName,
		"billingEmail": tenant.BillingEmail,
		"createdAt":    tenant.CreatedAt.Format(time.RFC3339),
		"updatedAt":    tenant.UpdatedAt.Format(time.RFC3339),
	}
}
