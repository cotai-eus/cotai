# Tenant Manager Service

Multi-tenant management service for the CotAI platform. Handles tenant lifecycle, schema provisioning, and tenant metadata management.

## Overview

The Tenant Manager Service is responsible for:

- **Tenant CRUD Operations**: Create, read, update, delete tenant records
- **Schema Provisioning**: Automatic PostgreSQL schema creation with RLS policies
- **Lifecycle Management**: Activate, suspend, and archive tenants
- **Event Publishing**: Kafka events for tenant lifecycle changes
- **Internal API**: gRPC service for inter-service communication
- **External API**: REST API for administrative operations

## Implementation Status

- ✅ **Phase A**: Project Setup - Go module, directory structure
- ✅ **Phase B**: Domain Layer - Entities, aggregates, interfaces
- ✅ **Phase C**: Database Layer - PostgreSQL, repositories, schema provisioning
- ✅ **Phase D**: Use Cases - Business logic orchestration
- ✅ **Phase E**: REST API - Chi router, JWT middleware, handlers
- ✅ **Phase F**: gRPC API - Protocol buffers, service implementation
- ✅ **Phase G**: Kafka Events - Event publisher for tenant.lifecycle
- ✅ **Phase H**: Observability - Prometheus metrics, Jaeger tracing
- ✅ **Phase I**: Docker Integration - Dockerfile, docker-compose
- ✅ **Phase J**: Testing & Documentation - Tests and docs

**Status**: ✅ All phases complete and ready for deployment

## Tech Stack

- **Language**: Go 1.21+
- **Web Framework**: Chi Router v5
- **gRPC**: grpc-go + Protocol Buffers
- **Database**: PostgreSQL 15 with sqlx
- **Messaging**: Kafka (IBM/sarama)
- **Observability**: Prometheus + Jaeger
- **Logging**: Zap (structured JSON)

## Quick Start

### Prerequisites

- Go 1.21 or later
- PostgreSQL 15+
- Kafka (optional, will use no-op publisher if not configured)
- Docker & Docker Compose (for containerized setup)

### Running with Docker Compose (Recommended)

```bash
# From project root
docker-compose up tenant-manager

# Or build and run
docker-compose up --build tenant-manager
```

### Running Locally

```bash
# 1. Install dependencies
make deps

# 2. Generate protobuf code
./scripts/gen-proto.sh

# 3. Build binary
make build

# 4. Run service
PORT=8082 GRPC_PORT=9082 ENV=development LOG_LEVEL=debug \
DATABASE_HOST=localhost DATABASE_PORT=5436 DATABASE_NAME=cotai_identity \
DATABASE_USER=cotai_dev DATABASE_PASSWORD=dev_password \
./bin/tenant-manager
```

## API Documentation

### REST API

**Base URL**: `http://localhost:8082`

#### Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/api/v1/tenants` | Create tenant | Admin |
| `GET` | `/api/v1/tenants` | List tenants (paginated) | Admin |
| `GET` | `/api/v1/tenants/{id}` | Get tenant details | Admin |
| `PATCH` | `/api/v1/tenants/{id}` | Update tenant | Admin |
| `DELETE` | `/api/v1/tenants/{id}` | Delete tenant (soft) | Admin |
| `POST` | `/api/v1/tenants/{id}/suspend` | Suspend tenant | Admin |
| `POST` | `/api/v1/tenants/{id}/activate` | Activate/reactivate tenant | Admin |
| `GET` | `/health` | Health check | Public |
| `GET` | `/ready` | Readiness check | Public |
| `GET` | `/metrics` | Prometheus metrics | Public |

#### Example: Create Tenant

**Request**:
```bash
curl -X POST http://localhost:8082/api/v1/tenants \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Empresa ACME",
    "slug": "acme",
    "plan": "professional",
    "adminEmail": "admin@acme.com.br",
    "adminName": "Admin User"
  }'
```

**Response (201)**:
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "tenantId": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Empresa ACME",
    "slug": "acme",
    "schemaName": "tenant_550e8400e29b41d4a716446655440000",
    "status": "active",
    "plan": "professional",
    "contactEmail": "admin@acme.com.br",
    "createdAt": "2025-12-16T10:30:00Z"
  }
}
```

### gRPC API

**Address**: `localhost:9082`

**Service**: `identity.tenant.v1.TenantService`

#### Methods

- `GetTenant(GetTenantRequest) returns (TenantResponse)`
- `GetTenantBySlug(GetBySlugRequest) returns (TenantResponse)`
- `ValidateTenant(ValidateTenantRequest) returns (ValidationResponse)`
- `ListTenants(ListTenantsRequest) returns (ListTenantsResponse)`

#### Example: Get Tenant (grpcurl)

```bash
grpcurl -plaintext \
  -d '{"tenant_id": "550e8400-e29b-41d4-a716-446655440000"}' \
  localhost:9082 identity.tenant.v1.TenantService/GetTenant
```

## Schema Provisioning

When a tenant is created, the service automatically:

1. Creates a PostgreSQL schema: `tenant_{uuid_without_hyphens}`
2. Applies migrations from `migrations/tenant_schema/`
3. Enables Row-Level Security (RLS) on all tables
4. Sets RLS policies for tenant isolation

**Example Schema Name**:
- UUID: `550e8400-e29b-41d4-a716-446655440000`
- Schema: `tenant_550e8400e29b41d4a716446655440000`

## Kafka Events

### Topic: `tenant.lifecycle`

#### Event Types

- `tenant.created` - New tenant provisioned
- `tenant.activated` - Tenant activated or reactivated
- `tenant.suspended` - Tenant suspended
- `tenant.deleted` - Tenant soft-deleted
- `tenant.updated` - Tenant metadata updated

#### Event Schema

```json
{
  "eventId": "evt_abc123",
  "eventType": "tenant.created",
  "tenantId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-12-16T10:30:00Z",
  "correlationId": "req_xyz789",
  "payload": {
    "name": "Empresa ACME",
    "slug": "acme",
    "plan": "professional",
    "schemaName": "tenant_550e8400e29b41d4a716446655440000",
    "status": "active"
  }
}
```

## Observability

### Metrics

**Endpoint**: `http://localhost:8082/metrics`

**Key Metrics**:
- `tenant_manager_tenant_created_total{plan}` - Total tenants created
- `tenant_manager_provisioning_duration_seconds{status}` - Schema provisioning time
- `tenant_manager_active_tenants{plan}` - Active tenants gauge
- `tenant_manager_http_requests_total{method,path,status}` - HTTP requests
- `tenant_manager_grpc_requests_total{method,status}` - gRPC requests

### Tracing

**Jaeger UI**: `http://localhost:16686`

The service emits distributed traces for HTTP requests, gRPC calls, database operations, and Kafka event publishing.

### Logging

Structured JSON logging with correlation IDs, tenant context, and error details.

## Development

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# E2E tests
make test-e2e

# Coverage report
make coverage
```

### Code Generation

```bash
# Generate protobuf code
make proto-gen
```

### Linting

```bash
make lint
```

## Project Structure

```
tenant-manager/
├── cmd/server/           # Application entrypoint
├── internal/
│   ├── app/              # Configuration
│   ├── domain/           # Business logic (entities, interfaces)
│   ├── usecase/          # Application logic
│   ├── delivery/         # HTTP & gRPC handlers
│   ├── infrastructure/   # Database, messaging, observability
│   └── pkg/              # Shared utilities
├── proto/                # Protobuf definitions
├── migrations/           # SQL migration templates
├── scripts/              # Build and deployment scripts
├── deployments/docker/   # Dockerfile
├── test/                 # Integration and E2E tests
└── docs/                 # Documentation
```

## Configuration

See `.env.example` for all available environment variables.

## Troubleshooting

### Database Connection Issues

```bash
# Verify PostgreSQL is running
docker ps | grep postgres-identity

# Check connection
psql -h localhost -p 5436 -U cotai_dev -d cotai_identity
```

### Kafka Connection Issues

```bash
# Check if Kafka is running
docker ps | grep kafka

# Consume tenant events
docker exec cotai-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic tenant.lifecycle \
  --from-beginning
```

### Health Check Failures

```bash
# Check service health
curl http://localhost:8082/health

# View logs
docker logs cotai-tenant-manager
```

## License

Proprietary - CotAI Platform

## Support

For issues or questions, contact the CotAI development team.
