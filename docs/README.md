# CotAI - Documentação Técnica de Microserviços

> Documentação completa da arquitetura de microserviços da plataforma CotAI de Gestão de Licitações.

## 📋 Índice de Navegação

### Arquitetura Fundamental
| Documento | Descrição |
|-----------|-----------|
| [Multi-Tenancy](./architecture/multi-tenancy.md) | Isolamento schema-per-tenant, RLS, provisioning |
| [Comunicação](./architecture/communication-patterns.md) | RabbitMQ, Kafka, gRPC, REST patterns |
| [Segurança](./architecture/security-compliance.md) | OAuth 2.0, RBAC, LGPD, encryption |

---

## 🔐 Identity & Compliance

| Serviço | Stack | Documentação |
|---------|-------|--------------|
| Auth Service | Keycloak / Node.js + PostgreSQL | [auth-service.md](./services/identity/auth-service.md) |
| Tenant Manager | Go + PostgreSQL | [tenant-manager.md](./services/identity/tenant-manager.md) |
| Audit Service | Go + Kafka → S3/ClickHouse | [audit-service.md](./services/identity/audit-service.md) |

---

## 📥 Acquisition & Ingestion

| Serviço | Stack | Documentação |
|---------|-------|--------------|
| Scheduler | Node.js + BullMQ + Redis | [scheduler.md](./services/acquisition/scheduler.md) |
| Crawler Workers | Python + Scrapy | [crawler-workers.md](./services/acquisition/crawler-workers.md) |
| Normalizer | Python + PostgreSQL | [normalizer.md](./services/acquisition/normalizer.md) |

---

## ⚙️ Core Bidding Engine

| Serviço | Stack | Documentação |
|---------|-------|--------------|
| Workflow Engine | Go + Temporal.io + PostgreSQL | [workflow-engine.md](./services/core-bidding/workflow-engine.md) |
| OCR/NLP Service | Python + Tesseract/Textract + S3 | [ocr-nlp-service.md](./services/core-bidding/ocr-nlp-service.md) |
| Data Extractor | Python + spaCy + Elasticsearch | [data-extractor.md](./services/core-bidding/data-extractor.md) |
| Kanban API | Node.js + NestJS + PostgreSQL | [kanban-api.md](./services/core-bidding/kanban-api.md) |

---

## 📦 Resource Management

| Serviço | Stack | Documentação |
|---------|-------|--------------|
| CRM Service | Node.js + NestJS + PostgreSQL | [crm-service.md](./services/resources/crm-service.md) |
| Stock Service | Node.js + PostgreSQL | [stock-service.md](./services/resources/stock-service.md) |
| Quote Service | Node.js + PostgreSQL | [quote-service.md](./services/resources/quote-service.md) |

---

## 💬 Collaboration

| Serviço | Stack | Documentação |
|---------|-------|--------------|
| Chat Service | Node.js + Socket.io + Redis | [chat-service.md](./services/collaboration/chat-service.md) |
| Notification Service | Node.js + Firebase + PostgreSQL | [notification-service.md](./services/collaboration/notification-service.md) |
| Agenda Service | Node.js + PostgreSQL | [agenda-service.md](./services/collaboration/agenda-service.md) |

---

## 🌐 Edge Layer

| Serviço | Stack | Documentação |
|---------|-------|--------------|
| Frontend | React 18 + Next.js + TypeScript | [frontend.md](./services/edge/frontend.md) |
| API Gateway | Kong + Redis | [api-gateway.md](./services/edge/api-gateway.md) |

---

## 🚀 Deployment & Infrastructure

| Documento | Descrição |
|-----------|-----------|
| [Environment Setup](./deployment/environment-setup.md) | Configuração ambiente desenvolvimento |
| [Docker Guidelines](./deployment/docker-guidelines.md) | Padrões Dockerfile, compose |
| [Kubernetes Deploy](./deployment/kubernetes-deploy.md) | Helm charts, HPA, secrets |

---

## ✅ Checklist de Implementação

- [Implementation Checklist](./implementation-checklist.md) — checklist acionável cobrindo arquitetura, implantação, configuração por serviço, workflows e controles de qualidade.

---

## 📊 Fluxos de Negócio

| Documento | Descrição |
|-----------|-----------|
| [Jornada do Edital](./workflows/edital-journey.md) | Fluxo completo discovery → cotação |
| [Pipeline de Crawlers](./workflows/crawler-pipeline.md) | Agendamento, coleta, normalização |

---

## Convenções

- **APIs REST**: OpenAPI 3.1, versionamento via path (`/v1/`)
- **Autenticação**: JWT via header `Authorization: Bearer <token>`
- **Tenant Resolution**: Header `X-Tenant-ID` ou claim JWT `tenant_id`
- **Mensageria**: RabbitMQ (comandos), Kafka (eventos de domínio)
- **Observabilidade**: Prometheus metrics em `/metrics`, tracing Jaeger

---

*Última atualização: Dezembro 2025*
