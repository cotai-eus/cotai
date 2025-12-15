# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

This is a **technical documentation repository** for CotAI, a cloud-native bidding management platform (plataforma de gestão de licitações). It contains architectural specifications, service designs, and implementation guidelines—not application code. When working with actual service implementations, refer to their respective code repositories.

## Documentation Structure

- [README.md](README.md) - Canonical architecture overview, bounded contexts, tech stack decisions
- [docs/architecture/](docs/architecture/) - Core architectural patterns (multi-tenancy, communication, security)
- [docs/services/](docs/services/) - Individual microservice specifications organized by domain
- [docs/deployment/](docs/deployment/) - Infrastructure and deployment guidelines
- [docs/workflows/](docs/workflows/) - Business process flows and integration patterns
- [docs/implementation-checklist.md](docs/implementation-checklist.md) - Implementation validation checklist
- [.github/instructions/copilot.instructions.md](.github/instructions/copilot.instructions.md) - AI agent guidance (Portuguese)

## Critical Architectural Decisions

These decisions are foundational and must not be changed without maintainer approval and design documentation:

### Multi-Tenancy: Schema-per-Tenant
- PostgreSQL schemas: `tenant_{uuid}` with Row-Level Security (RLS)
- Tenant resolution: JWT claim `tenant_id` (primary) or header `X-Tenant-ID` (fallback)
- Set `search_path` to `tenant_{uuid},public` before queries
- Enterprise tenants can migrate to dedicated databases
- Reference: [docs/architecture/multi-tenancy.md](docs/architecture/multi-tenancy.md)

### Communication Patterns
- **Synchronous**: REST (external APIs) / gRPC (internal services)
- **Asynchronous Commands**: RabbitMQ with queues `crawler.jobs`, `ocr.process`, DLQ `cotai.commands.dlq`
- **Domain Events**: Kafka topics `licitacao.status.changed`, `edital.raw`, `tenant.lifecycle`, `audit.events`
- All events/messages must include `tenant_id` and `correlation_id`
- Reference: [docs/architecture/communication-patterns.md](docs/architecture/communication-patterns.md)

### Bounded Contexts (DDD)
1. **Acquisition** - Crawlers, PNCP ingestion, normalization
2. **Core Bidding** - Kanban workflow, OCR/NLP, automation engine
3. **Resource Management** - CRM, inventory, supplier quotes
4. **Collaboration** - Real-time chat, notifications, agenda
5. **Identity** - Multi-tenant, RBAC, audit, compliance

## Key Business Flow

**Edital Journey** (procurement notice lifecycle):
```
DESCOBERTA → RECEBIDO → ANALISANDO → COTAR → COTADO/SEM RESPOSTA
```
- Automated transitions based on OCR extraction score and supplier responses
- Workflow orchestration via Temporal.io
- State changes emit Kafka events for cross-service coordination
- Full specification: [README.md#6-fluxo-do-core-jornada-do-edital](README.md#6-fluxo-do-core-jornada-do-edital) and [docs/workflows/edital-journey.md](docs/workflows/edital-journey.md)

## Working with Documentation

### Creating or Updating Service Docs
Service specifications follow this structure (see existing examples in [docs/services/](docs/services/)):
- **Overview** - Purpose, bounded context, core responsibilities
- **API Contracts** - REST/gRPC endpoints with request/response schemas
- **Data Model** - Entities, aggregates, relationships (include tenant isolation)
- **Integration Points** - Dependencies, messaging (RabbitMQ/Kafka), external APIs
- **Configuration** - Environment variables, secrets, feature flags
- **Observability** - Metrics exposed, tracing, logging format
- **Deployment** - Resource requirements, scaling policies, health checks

### Architecture Documentation Standards
When documenting architectural decisions:
- Update [README.md](README.md) for system-wide impacts
- Add cross-references between related docs
- Include mermaid diagrams for complex flows
- Specify tenant isolation requirements
- Document message schemas (commands/events)
- Note observability requirements (metrics, traces, logs)

### Security and Compliance Requirements
All service designs must address (see [docs/architecture/security-compliance.md](docs/architecture/security-compliance.md)):
- Authentication/Authorization (OAuth 2.0/OIDC, RBAC scopes)
- Tenant data isolation (schema + RLS enforcement)
- Encryption (at rest: AES-256/KMS, in transit: TLS 1.3)
- Audit logging (include `tenantId`, `userId`, `action`, `resource`, `correlationId`)
- LGPD compliance (consent, retention, erasure)

## Common Patterns

### API Conventions
- **REST**: OpenAPI 3.1, versioned via path `/v1/`, `/v2/`
- **Response envelope**: `{data, meta, links}`
- **Error format**: `{error: {code, message, details[]}}`
- **Required headers**: `Authorization: Bearer <token>`, `X-Tenant-ID`, `X-Request-ID`, `X-Correlation-ID`

### Messaging Patterns
- **RabbitMQ**: Configure retries + DLQ per queue, idempotent consumers
- **Kafka**: Key = `aggregateId`, headers include `tenant-id` and `event-type`
- **Event Schema**: Version events, include timestamp and correlation ID

### Observability
- Metrics endpoint: `/metrics` (Prometheus format)
- Health checks: `/health` (liveness), `/ready` (readiness)
- Distributed tracing: Jaeger with tenant and correlation ID propagation
- Structured JSON logging (no PII/secrets)

## Technology Stack

| Layer | Technology |
|-------|------------|
| Frontend | React 18, Next.js, TypeScript, TailwindCSS |
| API Gateway | Kong, Redis |
| Backend Services | Node.js (NestJS), Go, Python |
| Orchestration | Temporal.io (workflow engine) |
| Databases | PostgreSQL 15 (multi-tenant schemas), Elasticsearch 8 (search) |
| Cache | Redis 7 |
| Messaging | RabbitMQ 3.12 (commands), Kafka 3.x (events) |
| OCR/NLP | AWS Textract / Tesseract 5, spaCy |
| Infrastructure | Kubernetes (EKS), Terraform, Helm |
| Observability | Prometheus, Grafana, Jaeger |

## Cross-References

This CLAUDE.md complements the existing AI agent instructions:
- [.github/instructions/copilot.instructions.md](.github/instructions/copilot.instructions.md) - Comprehensive Portuguese guidance including service-specific commands and development workflows

When implementing services based on this documentation, create per-service CLAUDE.md files in code repositories that reference back to this central architecture documentation.
