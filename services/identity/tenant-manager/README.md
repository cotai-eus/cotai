# Tenant Manager Service

Serviço de gerenciamento de tenants para a plataforma CotAI.

## Visão Geral

O Tenant Manager é responsável por:
- CRUD de tenants via REST API
- Provisionamento automático de schemas PostgreSQL
- Configuração de Row-Level Security (RLS)
- Publicação de eventos de lifecycle no Kafka
- API gRPC para validação interna de tenants

## Status de Implementação

- ✅ **Phase A**: Project Setup (In Progress)
- ⏳ **Phase B**: Domain Layer
- ⏳ **Phase C**: Database Layer
- ⏳ **Phase D**: Use Cases
- ⏳ **Phase E**: REST API
- ⏳ **Phase F**: gRPC API
- ⏳ **Phase G**: Kafka Events
- ⏳ **Phase H**: Observability
- ⏳ **Phase I**: Docker Integration
- ⏳ **Phase J**: Testing & Documentation

## Quick Start

```bash
# Install dependencies
make deps

# Build binary
make build

# Run locally
make run

# Run tests
make test
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
# Edit .env with your settings
```

## Architecture

```
├── cmd/server/          # Entry point
├── internal/
│   ├── app/            # Configuration
│   ├── domain/         # Business logic (DDD)
│   ├── usecase/        # Use cases
│   ├── infrastructure/ # Database, Kafka, etc.
│   └── delivery/       # HTTP + gRPC handlers
├── proto/              # Protocol buffer definitions
├── migrations/         # Tenant schema templates
└── test/               # Tests
```

## API Endpoints

### REST API (Port 8082)

- `POST /api/v1/tenants` - Create tenant
- `GET /api/v1/tenants` - List tenants
- `GET /api/v1/tenants/{id}` - Get tenant
- `PATCH /api/v1/tenants/{id}` - Update tenant
- `POST /api/v1/tenants/{id}/suspend` - Suspend tenant
- `POST /api/v1/tenants/{id}/activate` - Activate tenant
- `DELETE /api/v1/tenants/{id}` - Delete tenant (soft)

### gRPC API (Port 9082)

- `GetTenant` - Retrieve tenant by ID
- `ValidateTenant` - Validate tenant status
- `GetTenantBySlug` - Retrieve by slug
- `ListTenants` - List all tenants

## Development

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development guide.

## Validation

See [docs/VALIDATION.md](docs/VALIDATION.md) for validation procedures.

## License

Proprietary - CotAI Platform
