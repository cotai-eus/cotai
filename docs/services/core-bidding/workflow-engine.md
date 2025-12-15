# Workflow Engine

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Core Bidding |
| Responsabilidade | Orquestrar estados Kanban e automações |
| Stack | Go 1.21 + Temporal.io |
| Database | PostgreSQL (tenant schema) + Redis (caches) |
| Protocolos | gRPC (workers), REST (admin), Kafka events |

## Responsabilidades
- Orquestrar transições de `Licitacao` no board (Recebido → Analisando → Cotar → Cotado → Sem Resposta)
- Disparar atividades assíncronas (OCR, NLP, notificações)
- Regras automáticas: score, SLA, retries
- Garantir idempotência e compensações (sagas)

## Arquitetura
```
Frontend → API Gateway → Kanban API → Workflow Engine (Temporal)
                                     ↘ Kafka (status events)
```
Workers: `ocrActivity`, `nlpExtractActivity`, `notifyActivity`, `timeoutActivity`.

## Fluxo Principal (State Machine)
1) Evento `edital.normalized` (Kafka) → inicia workflow `ProcessLicitacao`.
2) State RECEBIDO: persiste licitação, agenda OCR.
3) State ANALISANDO: aguarda OCR+NLP; calcula score; decide próximo estado.
4) State COTAR: notifica fornecedores; aguarda cotações; SLA 48h.
5) State COTADO ou SEM_RESPOSTA conforme respostas/SLA.
6) Emite `licitacao.status.changed` em Kafka a cada transição.

## API (Admin/Observability)
- `GET /api/v1/workflows/:licitacaoId` — status e histórico.
- `POST /api/v1/workflows/:licitacaoId/retry` — reprocessa atividades falhas.
- `POST /api/v1/workflows/:licitacaoId/cancel` — cancela workflow.
- `GET /health` — liveness/readiness.
Auth: JWT + `X-Tenant-ID`.

## gRPC (Workers → Engine)
```protobuf
service WorkflowControl {
  rpc StartProcess(StartRequest) returns (StartResponse);
  rpc Resume(ResumeRequest) returns (Empty);
  rpc GetStatus(GetStatusRequest) returns (StatusResponse);
}
```

## Eventos (Kafka)
| Evento | Tópico | Payload |
|--------|--------|---------|
| `licitacao.status.changed` | `licitacao.status.changed` | {licitacaoId, from, to, tenantId, reason}
| `licitacao.ocr.requested`  | `ocr.requests`             | {licitacaoId, s3Key, tenantId}
| `licitacao.ocr.completed`  | `ocr.completed`            | {licitacaoId, textS3Key}
| `licitacao.nlp.completed`  | `nlp.completed`            | {licitacaoId, items, deadlines, score}

## Modelo de Dados (PostgreSQL schema tenant)
```sql
CREATE TABLE licitacoes_workflow (
  id UUID PRIMARY KEY,
  status VARCHAR(20) NOT NULL,
  score NUMERIC(5,2),
  sla_due_at TIMESTAMPTZ,
  retries INT DEFAULT 0,
  last_error TEXT,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE workflow_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  licitacao_id UUID,
  from_status VARCHAR(20),
  to_status VARCHAR(20),
  reason TEXT,
  metadata JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Regras Automáticas
- RECEBIDO → ANALISANDO: OCR iniciado (PDF baixado).
- ANALISANDO → COTAR: `score >= 70%` e itens extraídos.
- ANALISANDO → SEM_RESPOSTA: `score < 30%` ou fora do segmento.
- COTAR → COTADO: ≥1 proposta válida.
- COTAR → SEM_RESPOSTA: SLA 48h expirado.

## Temporal Setup
- Task Queue: `workflow-engine`.
- Retry policy: maxAttempts=3, backoff=30s→5m.
- Timeouts: activity 2m OCR, 1m NLP; workflow run 7d.

## Segurança
- Auth via JWT RS256; autorização por role (`licitacao:write`).
- Tenant enforcement: search_path + RLS.
- Idempotência: workflowId = licitacaoId; activityId = step + attempt.

## Observabilidade
- Metrics: `workflow_transitions_total{from,to}`, `workflow_duration_seconds{state}`.
- Tracing: Jaeger instrumentation (Temporal interceptors).

## Exemplo de Chamada REST
```http
POST /api/v1/workflows/LIC123/retry
Authorization: Bearer <token>
X-Tenant-ID: tenant_abc
```

## Health / Readiness
- `/health/live` — process up
- `/health/ready` — DB + Temporal connected

*Referência: ver tópicos Kafka em docs/architecture/communication-patterns.md*