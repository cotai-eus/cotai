# Implementation Checklist

## Arquitetura do Sistema
- [ ] Mapear domínios, dependências síncronas/assíncronas e contratos conforme [docs/architecture/communication-patterns.md](docs/architecture/communication-patterns.md); garantir correlação de mensagens, idempotência e versionamento de APIs/eventos.
- [ ] Confirmar estratégia multi-tenant (schema-per-tenant, RLS, pools por tenant, provisioning e isolamento) seguindo [docs/architecture/multi-tenancy.md](docs/architecture/multi-tenancy.md); planejar migrações coordenadas por tenant e limites de recursos.
- [ ] Validar requisitos de segurança e compliance (OAuth2/OIDC, RBAC, criptografia em trânsito/repouso, mascaramento/PII, auditoria) conforme [docs/architecture/security-compliance.md](docs/architecture/security-compliance.md).
- [ ] Definir SLO/SLA/RPO/RTO, métricas e limites de erro/latência; registrar ADRs para decisões críticas.
- [ ] Garantir observabilidade: logs estruturados, métricas em /metrics, tracing distribuído, correlação por tenant e correlation-id.

## Implantação (Build, Infra, Kubernetes)
- [ ] Imagens: multi-stage, usuário não-root, varrer secrets, gerar SBOM e scan de vulnerabilidades conforme [docs/deployment/docker-guidelines.md](docs/deployment/docker-guidelines.md).
- [ ] Ambientes: provisionar dependências (Postgres, Redis, Kafka, RabbitMQ, S3/MinIO, Elastic) conforme [docs/deployment/environment-setup.md](docs/deployment/environment-setup.md); automatizar seeds e migrations.
- [ ] Kubernetes: readiness/liveness, requests/limits, HPA, PodDisruptionBudget, anti-affinity, rollout/canary e rollback de acordo com [docs/deployment/kubernetes-deploy.md](docs/deployment/kubernetes-deploy.md).
- [ ] Configuração e secrets: ExternalSecrets/Vault, rotação, TLS interno, policies de least privilege (service accounts, network policies, securityContext).
- [ ] CI/CD gates: build → test → scan → SBOM → image signing → deploy canary → promoção; rollback automático, artefatos versionados e promoção entre ambientes.

## Configuração Individual por Serviço
- [ ] Validar variáveis obrigatórias e tipos na inicialização; expor /health e /ready; métricas e tracing propagando tenant e correlation-id.
- [ ] Resiliência: timeouts, retries com backoff, circuit breaker e bulkhead para integrações externas.
- [ ] Mensageria/eventos: DLQ e retry policy por fila/tópico, deduplicação e idempotência, contratos versionados.
- [ ] Dados: migrations idempotentes por tenant, índices e retenção, arquivamento e políticas de backup/restore.
- [ ] Segurança por serviço: authn/authz (JWT/OIDC), escopos por tenant, sanitização de PII, auditoria quando aplicável.
- [ ] Edge: gateway/ingress, JWT/OIDC, rate limiting e roteamento seguindo [docs/services/edge/api-gateway.md](docs/services/edge/api-gateway.md) e [docs/services/edge/frontend.md](docs/services/edge/frontend.md).
- [ ] Identity: Keycloak/OIDC, auditoria em ClickHouse/S3, provisionamento de tenants conforme [docs/services/identity/auth-service.md](docs/services/identity/auth-service.md), [docs/services/identity/tenant-manager.md](docs/services/identity/tenant-manager.md) e [docs/services/identity/audit-service.md](docs/services/identity/audit-service.md).
- [ ] Acquisition: crawlers, normalizer e scheduler (Rabbit→Kafka, rate limits, dedupe) conforme [docs/services/acquisition/crawler-workers.md](docs/services/acquisition/crawler-workers.md), [docs/services/acquisition/normalizer.md](docs/services/acquisition/normalizer.md) e [docs/services/acquisition/scheduler.md](docs/services/acquisition/scheduler.md).
- [ ] Core Bidding: workflow kanban, OCR/NLP, data extractor e engine (regras de estado, SLAs, Tesseract/Textract, Temporal) conforme [docs/services/core-bidding/kanban-api.md](docs/services/core-bidding/kanban-api.md), [docs/services/core-bidding/ocr-nlp-service.md](docs/services/core-bidding/ocr-nlp-service.md), [docs/services/core-bidding/data-extractor.md](docs/services/core-bidding/data-extractor.md) e [docs/services/core-bidding/workflow-engine.md](docs/services/core-bidding/workflow-engine.md).
- [ ] Resources: CRM/Quote/Stock (reservas, cotação, estoque) seguindo [docs/services/resources/crm-service.md](docs/services/resources/crm-service.md), [docs/services/resources/quote-service.md](docs/services/resources/quote-service.md) e [docs/services/resources/stock-service.md](docs/services/resources/stock-service.md).
- [ ] Collaboration: agenda/chat/notification (WebSocket/Redis, templates, retries) conforme [docs/services/collaboration/agenda-service.md](docs/services/collaboration/agenda-service.md), [docs/services/collaboration/chat-service.md](docs/services/collaboration/chat-service.md) e [docs/services/collaboration/notification-service.md](docs/services/collaboration/notification-service.md).

## Workflows
- [ ] Mapear passos, SLAs e eventos do crawler pipeline conforme [workflows/crawler-pipeline.md](workflows/crawler-pipeline.md); definir KPIs e alertas por etapa.
- [ ] Mapear jornada do edital (descoberta → recebido → analisando → cotar → cotado/sem resposta) conforme [workflows/edital-journey.md](workflows/edital-journey.md); definir guardrails de transição e compensações.
- [ ] Idempotência e reprocessamento: IDs determinísticos, checkpoints, política de replay/compensação, tratamento de atrasos e reordenação.
- [ ] Testes end-to-end: caminhos felizes, falhas externas, backlog de fila, carga multi-tenant, caos básico em etapas críticas.
- [ ] Operação: runbooks por workflow, critérios de pausa/retomada, dashboards por etapa, procedimentos de retomada pós-incidente.

## Qualidade, Observabilidade e Operação
- [ ] Contratos: testes de contrato (OpenAPI/Protobuf), lint/format, testes unitários/integrados com fixtures de Kafka/Rabbit/Temporal.
- [ ] Performance e capacidade: testes de carga e tuning (DB, cache, filas), limites por tenant, orçamentos de erro e latência.
- [ ] Alertas mínimos: latência p95/p99, error rate, backlog de fila, throughput Kafka/RabbitMQ, uso de CPU/memória, falhas de deploy.
- [ ] Segurança contínua: scans SAST/DAST, varredura de dependências, rotação de chaves/creds, revisões de acesso e auditoria periódica.
- [ ] Runbooks e DR: planos de rollback, procedimentos de migração reversível, backups verificados, RPO/RTO monitorados.
