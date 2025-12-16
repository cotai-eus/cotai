package mapper

import (
	"github.com/cotai/tenant-manager/internal/domain"
	tenantv1 "github.com/cotai/tenant-manager/proto/tenant/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DomainToProto converts domain.Tenant to proto Tenant
func DomainToProto(tenant *domain.Tenant) *tenantv1.Tenant {
	if tenant == nil {
		return nil
	}

	return &tenantv1.Tenant{
		Id:            tenant.ID.String(),
		TenantId:      tenant.TenantID.String(),
		Name:          tenant.TenantName,
		Slug:          tenant.TenantSlug,
		SchemaName:    tenant.DatabaseSchema,
		Status:        StatusDomainToProto(tenant.Status),
		Plan:          string(tenant.PlanTier),
		ContactEmail:  tenant.PrimaryContactEmail,
		ContactName:   tenant.PrimaryContactName,
		BillingEmail:  tenant.BillingEmail,
		CreatedAt:     timestamppb.New(tenant.CreatedAt),
		UpdatedAt:     timestamppb.New(tenant.UpdatedAt),
	}
}

// StatusDomainToProto converts domain.TenantStatus to proto TenantStatus
func StatusDomainToProto(status domain.TenantStatus) tenantv1.TenantStatus {
	switch status {
	case domain.StatusProvisioning:
		return tenantv1.TenantStatus_TENANT_STATUS_PROVISIONING
	case domain.StatusActive:
		return tenantv1.TenantStatus_TENANT_STATUS_ACTIVE
	case domain.StatusSuspended:
		return tenantv1.TenantStatus_TENANT_STATUS_SUSPENDED
	case domain.StatusArchived:
		return tenantv1.TenantStatus_TENANT_STATUS_ARCHIVED
	case domain.StatusDeleted:
		return tenantv1.TenantStatus_TENANT_STATUS_DELETED
	default:
		return tenantv1.TenantStatus_TENANT_STATUS_UNSPECIFIED
	}
}

// StatusProtoToDomain converts proto TenantStatus to domain.TenantStatus
func StatusProtoToDomain(status tenantv1.TenantStatus) domain.TenantStatus {
	switch status {
	case tenantv1.TenantStatus_TENANT_STATUS_PROVISIONING:
		return domain.StatusProvisioning
	case tenantv1.TenantStatus_TENANT_STATUS_ACTIVE:
		return domain.StatusActive
	case tenantv1.TenantStatus_TENANT_STATUS_SUSPENDED:
		return domain.StatusSuspended
	case tenantv1.TenantStatus_TENANT_STATUS_ARCHIVED:
		return domain.StatusArchived
	case tenantv1.TenantStatus_TENANT_STATUS_DELETED:
		return domain.StatusDeleted
	default:
		return ""
	}
}
