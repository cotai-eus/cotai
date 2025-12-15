# Tenant Manager

## Visão Geral

| Atributo | Valor |
|----------|-------|
| **Bounded Context** | Identity & Compliance |
| **Responsabilidade** | CRUD tenants, provisioning de schemas, lifecycle management |
| **Stack** | Go 1.21+ |
| **Database** | PostgreSQL (schema: `public`) |
| **Porta** | 8082 |
| **Protocolo** | REST + gRPC |

---

## Responsabilidades

1. **Cadastro de tenants** (onboarding)
2. **Provisioning de schemas** PostgreSQL
3. **Gestão de lifecycle** (active, suspended, deleted)
4. **Configurações por tenant** (features, limites)
5. **Migração para database dedicado** (enterprise)
6. **Billing metadata** (planos, quotas)

---

## Arquitetura

```
┌──────────────┐    gRPC     ┌────────────────┐
│ Auth Service │ ◄──────────►│ Tenant Manager │
└──────────────┘             └───────┬────────┘
                                     │
┌──────────────┐    REST             │
│ Admin Portal │ ◄───────────────────┤
└──────────────┘                     │
                              ┌──────▼──────┐
                              │  PostgreSQL │
                              │   (public)  │
                              └─────────────┘
```

---

## API Endpoints

### REST API

#### Criar Tenant

```http
POST /api/v1/tenants
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "Empresa ACME",
  "slug": "acme",
  "plan": "professional",
  "adminEmail": "admin@acme.com.br",
  "settings": {
    "timezone": "America/Sao_Paulo",
    "language": "pt-BR",
    "features": ["ocr", "nlp", "crawler"]
  }
}
```

**Response (201):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Empresa ACME",
  "slug": "acme",
  "schemaName": "tenant_550e8400e29b41d4a716446655440000",
  "status": "provisioning",
  "plan": "professional",
  "createdAt": "2025-12-15T10:30:00Z"
}
```

#### Listar Tenants

```http
GET /api/v1/tenants?page=1&per_page=20&status=active
Authorization: Bearer <admin_token>
```

**Response:**
```json
{
  "data": [
    {
      "id": "550e8400-...",
      "name": "Empresa ACME",
      "slug": "acme",
      "status": "active",
      "plan": "professional",
      "usersCount": 15,
      "createdAt": "2025-12-15T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 45
  }
}
```

#### Obter Tenant

```http
GET /api/v1/tenants/{tenant_id}
Authorization: Bearer <admin_token>
```

#### Atualizar Tenant

```http
PATCH /api/v1/tenants/{tenant_id}
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "name": "ACME Corporation",
  "settings": {
    "features": ["ocr", "nlp", "crawler", "ai-scoring"]
  }
}
```

#### Suspender Tenant

```http
POST /api/v1/tenants/{tenant_id}/suspend
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "reason": "payment_overdue",
  "suspendUntil": "2025-12-31T23:59:59Z"
}
```

#### Reativar Tenant

```http
POST /api/v1/tenants/{tenant_id}/activate
Authorization: Bearer <admin_token>
```

#### Deletar Tenant (soft)

```http
DELETE /api/v1/tenants/{tenant_id}
Authorization: Bearer <admin_token>
```

---

### gRPC API

```protobuf
// proto/tenant.proto
syntax = "proto3";
package identity.v1;

service TenantService {
  rpc GetTenant(GetTenantRequest) returns (Tenant);
  rpc ValidateTenant(ValidateTenantRequest) returns (ValidateResponse);
  rpc GetTenantBySlug(GetBySlugRequest) returns (Tenant);
  rpc ListTenants(ListTenantsRequest) returns (ListTenantsResponse);
}

message Tenant {
  string id = 1;
  string name = 2;
  string slug = 3;
  string schema_name = 4;
  TenantStatus status = 5;
  string plan = 6;
  TenantSettings settings = 7;
  int64 created_at = 8;
}

enum TenantStatus {
  PROVISIONING = 0;
  ACTIVE = 1;
  SUSPENDED = 2;
  DELETED = 3;
}

message TenantSettings {
  string timezone = 1;
  string language = 2;
  repeated string features = 3;
  map<string, int32> quotas = 4;
}
```

---

## Modelo de Dados

```sql
-- Schema: public

CREATE TYPE tenant_status AS ENUM (
  'provisioning', 'active', 'suspended', 'deleted'
);

CREATE TYPE tenant_plan AS ENUM (
  'starter', 'professional', 'enterprise'
);

CREATE TABLE tenants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(100) UNIQUE NOT NULL,
  schema_name VARCHAR(100) UNIQUE NOT NULL,
  status tenant_status DEFAULT 'provisioning',
  plan tenant_plan DEFAULT 'starter',
  settings JSONB DEFAULT '{}',
  quotas JSONB DEFAULT '{
    "users": 10,
    "licitacoes_per_month": 100,
    "storage_gb": 5
  }',
  database_host VARCHAR(255), -- Para enterprise (database dedicado)
  suspended_at TIMESTAMPTZ,
  suspended_reason VARCHAR(255),
  deleted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tenant_features (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id),
  feature_code VARCHAR(50) NOT NULL,
  enabled BOOLEAN DEFAULT true,
  config JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(tenant_id, feature_code)
);

CREATE TABLE tenant_usage (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID REFERENCES tenants(id),
  period DATE NOT NULL,
  users_count INT DEFAULT 0,
  licitacoes_count INT DEFAULT 0,
  storage_bytes BIGINT DEFAULT 0,
  api_calls INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(tenant_id, period)
);

-- Índices
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenant_usage_period ON tenant_usage(tenant_id, period);
```

---

## Schema Provisioning

```go
// internal/provisioning/schema.go
package provisioning

import (
    "context"
    "database/sql"
    "fmt"
)

type SchemaProvisioner struct {
    db         *sql.DB
    migrations string
}

func (p *SchemaProvisioner) ProvisionTenant(ctx context.Context, tenantID string) error {
    schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenantID, "-", ""))
    
    // 1. Criar schema
    _, err := p.db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
    if err != nil {
        return fmt.Errorf("failed to create schema: %w", err)
    }
    
    // 2. Aplicar migrations
    if err := p.applyMigrations(ctx, schemaName); err != nil {
        return fmt.Errorf("failed to apply migrations: %w", err)
    }
    
    // 3. Configurar RLS
    if err := p.enableRLS(ctx, schemaName); err != nil {
        return fmt.Errorf("failed to enable RLS: %w", err)
    }
    
    // 4. Seed data (opcional)
    if err := p.seedInitialData(ctx, schemaName, tenantID); err != nil {
        return fmt.Errorf("failed to seed data: %w", err)
    }
    
    return nil
}

func (p *SchemaProvisioner) applyMigrations(ctx context.Context, schema string) error {
    // Usar golang-migrate ou similar
    m, err := migrate.New(p.migrations, fmt.Sprintf("postgres://...?search_path=%s", schema))
    if err != nil {
        return err
    }
    return m.Up()
}
```

---

## Eventos Publicados

| Evento | Tópico Kafka | Trigger |
|--------|--------------|---------|
| `tenant.created` | `tenant.lifecycle` | Após provisioning completo |
| `tenant.activated` | `tenant.lifecycle` | Reativação |
| `tenant.suspended` | `tenant.lifecycle` | Suspensão |
| `tenant.deleted` | `tenant.lifecycle` | Soft delete |
| `tenant.plan.changed` | `tenant.lifecycle` | Upgrade/downgrade |

```json
{
  "eventId": "evt_abc123",
  "eventType": "tenant.created",
  "tenantId": "550e8400-...",
  "timestamp": "2025-12-15T10:30:00Z",
  "payload": {
    "name": "Empresa ACME",
    "slug": "acme",
    "plan": "professional",
    "schemaName": "tenant_550e8400..."
  }
}
```

---

## Variáveis de Ambiente

```bash
# Server
PORT=8082
GRPC_PORT=9082
ENV=production

# Database
DATABASE_URL=postgres://user:pass@postgres:5432/cotai?sslmode=require
DATABASE_MAX_CONNS=25

# Migrations
MIGRATIONS_PATH=file://migrations

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_TENANT=tenant.lifecycle

# Auth
AUTH_GRPC_ADDR=auth-service:9080
JWT_PUBLIC_KEY_URL=https://auth.cotai.com.br/realms/cotai/protocol/openid-connect/certs
```

---

## Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tenant-manager ./cmd/server

FROM alpine:3.18
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /tenant-manager .
COPY migrations ./migrations
EXPOSE 8082 9082
CMD ["./tenant-manager"]
```

---

## Exemplo de Uso

```bash
# Criar tenant via CLI
curl -X POST https://api.cotai.com.br/v1/tenants \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Nova Empresa",
    "slug": "nova-empresa",
    "plan": "professional",
    "adminEmail": "admin@nova-empresa.com.br"
  }'

# Verificar status de provisioning
curl https://api.cotai.com.br/v1/tenants/550e8400-... \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

---

## Health Check

```http
GET /health
```

```json
{
  "status": "healthy",
  "version": "1.2.0",
  "checks": {
    "database": "ok",
    "kafka": "ok",
    "grpc": "ok"
  }
}
```

---

*Referência: [Multi-Tenancy](../architecture/multi-tenancy.md) | [Auth Service](./auth-service.md)*
