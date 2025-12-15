```instructions
---
applyTo: '**'
---
# Copilot / AI Agent Instructions for this repository

[ADDED] Portuguese Copilot instructions integrated from maintainer input. Preserve frontmatter and update when architecture or commands change.

# Copilot Instructions

- Contexto geral: plataforma cloud-native de licitações dividida em microsserviços por bounded context (Acquisition, Core Bidding, Resource Mgmt, Collaboration, Identity). Consulte [README.md](../README.md) para visão macro, diagramas e decisões de arquitetura (schema-per-tenant, RabbitMQ para comandos, Kafka para eventos, OCR/NLP distribuído).
- Tenant: isolamento via PostgreSQL schema-per-tenant com RLS. Resolva tenant por claim `tenant_id` do JWT ou header `X-Tenant-ID`; subdomínio é fallback. Sempre configure `search_path` para `tenant_{uuid},public` e inclua `tenant_id` em eventos e logs. Referências: [docs/architecture/multi-tenancy.md](../docs/architecture/multi-tenancy.md).
- Comunicação: REST/gRPC para consultas síncronas; RabbitMQ (`cotai.commands` direct, filas `crawler.jobs`, `ocr.process`, DLQ `cotai.commands.dlq`) para comandos; Kafka para eventos (`licitacao.status.changed`, `edital.raw`, `tenant.lifecycle`, `audit.events`). Modelos e exemplos em [docs/architecture/communication-patterns.md](../docs/architecture/communication-patterns.md).
- Fluxos críticos: jornada do edital (crawler → normalizer → workflow → kanban/quote) em [README.md](../README.md#6-fluxo-do-core-jornada-do-edital) e [docs/workflows/edital-journey.md](../docs/workflows/edital-journey.md). Workflow Engine (Temporal) controla transições e dispara OCR/NLP; Kanban API emite eventos de status e aplica regras do workflow.
- Serviços principais e docs:
  - Edge: Frontend Next.js em [docs/services/edge/frontend.md](../docs/services/edge/frontend.md) (OIDC PKCE, cookies httpOnly; `npm ci && npm run lint && npm run build && npm run start`), API Gateway Kong em [docs/services/edge/api-gateway.md](../docs/services/edge/api-gateway.md) (plugins oidc/jwt, rate limit por tenant, Prometheus plugin).
  - Identity: Auth Service (Keycloak) em [docs/services/identity/auth-service.md](../docs/services/identity/auth-service.md); Tenant Manager Go com provisioning de schemas e gRPC em [docs/services/identity/tenant-manager.md](../docs/services/identity/tenant-manager.md); Audit Service Go consumindo `audit.events` → ClickHouse/S3 em [docs/services/identity/audit-service.md](../docs/services/identity/audit-service.md).
  - Acquisition: Scheduler Node + BullMQ publica em RabbitMQ (cron jobs, rate limits) em [docs/services/acquisition/scheduler.md](../docs/services/acquisition/scheduler.md); Crawler Workers Python+Scrapy consomem `crawler.jobs`, salvam PDFs em S3 e publicam `edital.raw` em [docs/services/acquisition/crawler-workers.md](../docs/services/acquisition/crawler-workers.md); Normalizer em [docs/services/acquisition/normalizer.md](../docs/services/acquisition/normalizer.md).
  - Core Bidding: Workflow Engine Go + Temporal em [docs/services/core-bidding/workflow-engine.md](../docs/services/core-bidding/workflow-engine.md); Kanban API NestJS em [docs/services/core-bidding/kanban-api.md](../docs/services/core-bidding/kanban-api.md); OCR/NLP Service Python (RabbitMQ in, Kafka out, S3 storage) em [docs/services/core-bidding/ocr-nlp-service.md](../docs/services/core-bidding/ocr-nlp-service.md); Data Extractor em [docs/services/core-bidding/data-extractor.md](../docs/services/core-bidding/data-extractor.md).
- Ambiente local: requisitos e bootstrap de deps (Postgres, Redis, Kafka, RabbitMQ) em [docs/deployment/environment-setup.md](../docs/deployment/environment-setup.md). Use `docker compose up -d postgres redis kafka rabbitmq` antes de rodar serviços individuais. Perfis: `dev` (logs verbosos, seeds, CORS amplo) vs `prod` (TLS obrigatório, secrets em Vault/Secrets Manager).
- Segurança/compliance: siga [docs/architecture/security-compliance.md](../docs/architecture/security-compliance.md) (OAuth2/OIDC RS256, MFA, rate limiting por tenant, CSP, encryption at rest KMS). Eventos de auditoria devem incluir `tenantId`, `userId`, `action`, `resource` e `correlationId`.
- Observabilidade: métricas Prometheus padrão em `/metrics`; tracing Jaeger em serviços Go/Nest; Kong e Keycloak expõem métricas próprias. Logs estruturados (JSON) e sem dados sensíveis.
- Convenções de API: REST v1 com envelopes `{data, meta, links}`; erros `{error: {code, message, details[]}}`; headers `Authorization: Bearer`, `X-Tenant-ID`, `X-Request-ID` e `X-Correlation-ID`. gRPC usa protos em `proto/*` (ver exemplos em docs).
- Mensageria: configure retries + DLQ em RabbitMQ; para Kafka use chave = aggregateId, inclua headers `tenant-id` e `event-type`; consumidores devem ser idempotentes.
- Dados: todas as tabelas incluem `tenant_id`; habilite RLS e defina `app.current_tenant` antes de queries. Para serviços que criam schemas, reutilize provisioning do Tenant Manager ou scripts de exemplo em [docs/architecture/multi-tenancy.md](../docs/architecture/multi-tenancy.md#provisioning-de-novo-tenant).
- Testes/qualidade: Node services usam `npm test` e `npm run lint`; Go `go test ./...`; Python `pytest`. Smoke básico: `curl -f http://localhost:3000/health` (veja environment setup). Priorize testes de fluxo de eventos (Kafka/RabbitMQ) quando alterar integrações.
- Padrões de código: preferir NestJS interceptors para tracing/metrics; usar schemas OpenAPI 3.1 e versionamento por path `/v1/`; em Python, mantenha pipelines Scrapy idempotentes e assíncronos; em Go, injetar contexto com `tenant_id` e `correlation_id`.
- Quando algo faltar: cada serviço tem README em `docs/services/*`. Se não houver comando de build/run na pasta de código, peça confirmação aos maintainers antes de assumir defaults.

Atualize este arquivo quando novas decisões arquiteturais ou comandos de build surgirem. Diga ao maintainer se alguma seção estiver incompleta ou desatualizada.

[REPLACED] The previous English guidance (original body) has been replaced by the Portuguese content above. If you want to preserve the original English text, keep an archived copy elsewhere.

```
---
applyTo: '**'
---
# Copilot / AI Agent Instructions for this repository

Keep guidance concise and executable. Use the repository `README.md` and `docs/` as the primary source of truth for architecture and domain boundaries.

Key pointers
- **Big picture:** This is a cloud-native, microservice platform (see `README.md`). Major subsystems: Acquisition (crawlers, normalizer), Core Bidding (workflow/Kanban, OCR/NLP), Resource Management (CRM, stock), Collaboration (chat, notifications), Identity (multi-tenant + RBAC). Use `docs/` for diagrams and deeper explanations.
- **Tenant model:** Schema-per-tenant is used (Postgres schemas like `tenant_{uuid}`) and tenant resolution is implemented via JWT claim or subdomain. Do not change tenant-storage decisions without consulting maintainers.
- **Messaging:** Commands -> RabbitMQ, Events -> Kafka. Crawlers and heavy processing use RabbitMQ; lifecycle events flow through Kafka.
- **OCR / NLP pipeline:** PDFs are stored in S3; OCR and extractors run as separate services (Python/Tesseract or AWS Textract). Look for worker/service code under `src/*`.

Where to look first
- `README.md` (root): canonical architecture, bounded contexts, and service list — start here for domain intent and constraints.
- `docs/`: diagrams and operational notes.
- `src/`: services and service-specific code (e.g., `src/getway-api`). Inspect per-service folders for manifests and run instructions.

Repository patterns and conventions
- Microservices: mixed stacks (Node.js/NestJS, Go, Python). Expect per-service manifests rather than a single root manifest.
- Infra as code: Terraform + Helm are referenced in docs/README; look for `infra/` or `charts/` folders when implementing infra changes.
- Messaging and async patterns: prefer publishing domain events to Kafka; use RabbitMQ for command/queue processing with retries and DLQs.
- Observability: Prometheus + Grafana + Jaeger are used; add metrics and traces when touching service code paths.

Developer workflows (what the agent can do)
- When modifying or adding services, search the service folder for build files (`package.json`, `go.mod`, `pyproject.toml`, `Dockerfile`). If none are present, ask the repo owner for the service-level commands.
- Create focused small PRs that update a single service or infra chart. Include a short README snippet in the service folder with start/test commands if missing.
- Add metrics and tracing when modifying business logic flows (use existing conventions in `README.md`).

Safety and merge guidance
- Do not alter global tenant storage architecture (schema-per-tenant) or messaging backbone (RabbitMQ/Kafka) without an explicit design doc and maintainer sign-off.
- For database migrations, follow schema-per-tenant strategy — note that migrations may need to run per-schema. Confirm approach with maintainers before automating.

If you need missing operational commands
- This repo currently has no top-level `package.json`, `go.mod`, `pyproject.toml`, `Makefile`, or `Dockerfile` in the repository root. When you cannot discover a service's build or run commands from its folder, open an issue or request clarification from the maintainers and include a suggested minimal command example.

Examples from this codebase
- Architecture and bounded contexts: see `README.md` (root) for domain model and texture of services.
- Starting point for analysis: inspect `src/getway-api` for an API surface to learn routing and auth patterns.

Behavioral rules for the agent
- Preserve established design decisions; when proposing changes, add a short justification and open an issue first.
- Avoid global refactors across services in a single PR — prefer per-service patches.
- When uncertain about runtime commands or CI behavior, ask humans rather than guessing.

Next steps for humans
- Confirm per-service build/test/run commands and add them to each service folder's README.
- If you want stricter automation rules, add an `AGENT.md` with CI/CD expectations and required approvals.

If anything here is unclear or you want more detail on a specific service, tell me which service folder to analyze next.
