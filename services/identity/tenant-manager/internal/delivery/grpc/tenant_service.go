package grpc

import (
	"context"

	"github.com/cotai/tenant-manager/internal/delivery/grpc/mapper"
	"github.com/cotai/tenant-manager/internal/domain"
	"github.com/cotai/tenant-manager/internal/usecase"
	tenantv1 "github.com/cotai/tenant-manager/proto/tenant/v1"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TenantServiceServer implements the gRPC TenantService
type TenantServiceServer struct {
	tenantv1.UnimplementedTenantServiceServer
	getTenantUC   *usecase.GetTenantUseCase
	listTenantsUC *usecase.ListTenantsUseCase
	logger        *zap.Logger
}

// NewTenantServiceServer creates a new gRPC tenant service server
func NewTenantServiceServer(
	getTenantUC *usecase.GetTenantUseCase,
	listTenantsUC *usecase.ListTenantsUseCase,
	logger *zap.Logger,
) *TenantServiceServer {
	return &TenantServiceServer{
		getTenantUC:   getTenantUC,
		listTenantsUC: listTenantsUC,
		logger:        logger,
	}
}

// GetTenant retrieves a tenant by ID
func (s *TenantServiceServer) GetTenant(ctx context.Context, req *tenantv1.GetTenantRequest) (*tenantv1.TenantResponse, error) {
	// Validate request
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id format")
	}

	// Execute use case
	tenant, err := s.getTenantUC.Execute(ctx, tenantID)
	if err != nil {
		return nil, s.handleError(err)
	}

	// Convert to proto
	return &tenantv1.TenantResponse{
		Tenant: mapper.DomainToProto(tenant),
	}, nil
}

// GetTenantBySlug retrieves a tenant by slug
func (s *TenantServiceServer) GetTenantBySlug(ctx context.Context, req *tenantv1.GetBySlugRequest) (*tenantv1.TenantResponse, error) {
	// Validate request
	if req.Slug == "" {
		return nil, status.Error(codes.InvalidArgument, "slug is required")
	}

	// Execute use case with slug query
	tenant, err := s.getTenantUC.GetBySlug(ctx, req.Slug)
	if err != nil {
		return nil, s.handleError(err)
	}

	// Convert to proto
	return &tenantv1.TenantResponse{
		Tenant: mapper.DomainToProto(tenant),
	}, nil
}

// ValidateTenant checks if a tenant exists and is active
func (s *TenantServiceServer) ValidateTenant(ctx context.Context, req *tenantv1.ValidateTenantRequest) (*tenantv1.ValidationResponse, error) {
	// Validate request
	if req.TenantId == "" {
		return &tenantv1.ValidationResponse{
			Valid:   false,
			Message: "tenant_id is required",
		}, nil
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return &tenantv1.ValidationResponse{
			Valid:   false,
			Message: "invalid tenant_id format",
		}, nil
	}

	// Execute use case
	tenant, err := s.getTenantUC.Execute(ctx, tenantID)
	if err != nil {
		if domain.IsNotFoundError(err) {
			return &tenantv1.ValidationResponse{
				Valid:   false,
				Message: "tenant not found",
			}, nil
		}
		return nil, s.handleError(err)
	}

	// Check if tenant is active
	isValid := tenant.Status == domain.StatusActive
	message := "tenant is valid and active"
	if !isValid {
		message = "tenant is not active"
	}

	return &tenantv1.ValidationResponse{
		Valid:      isValid,
		TenantId:   tenant.TenantID.String(),
		SchemaName: tenant.DatabaseSchema,
		Status:     mapper.StatusDomainToProto(tenant.Status),
		Message:    message,
	}, nil
}

// ListTenants retrieves a paginated list of tenants
func (s *TenantServiceServer) ListTenants(ctx context.Context, req *tenantv1.ListTenantsRequest) (*tenantv1.ListTenantsResponse, error) {
	// Build use case query
	query := usecase.ListTenantsQuery{
		Page:    int(req.Page),
		PerPage: int(req.PageSize),
		Search:  req.Search,
	}

	// Set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 20
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}

	// Parse filters
	if req.Status != "" {
		query.Status = domain.TenantStatus(req.Status)
	}
	if req.Plan != "" {
		query.PlanTier = domain.PlanTier(req.Plan)
	}

	// Execute use case
	result, err := s.listTenantsUC.Execute(ctx, query)
	if err != nil {
		return nil, s.handleError(err)
	}

	// Convert tenants to proto
	protoTenants := make([]*tenantv1.Tenant, len(result.Tenants))
	for i, tenant := range result.Tenants {
		protoTenants[i] = mapper.DomainToProto(tenant)
	}

	return &tenantv1.ListTenantsResponse{
		Tenants:    protoTenants,
		TotalCount: int32(result.Total),
		Page:       int32(result.Page),
		PageSize:   int32(result.PerPage),
		TotalPages: int32(result.TotalPages),
	}, nil
}

// handleError converts domain errors to gRPC errors
func (s *TenantServiceServer) handleError(err error) error {
	s.logger.Error("gRPC service error", zap.Error(err))

	if domain.IsNotFoundError(err) {
		return status.Error(codes.NotFound, err.Error())
	}

	if domain.IsAlreadyExistsError(err) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	if domain.IsValidationError(err) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// Default to internal error
	return status.Error(codes.Internal, "internal server error")
}
