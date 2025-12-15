# Chat Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Collaboration |
| Responsabilidade | Mensageria em tempo real (chat de equipes) |
| Stack | Node.js 20 + Socket.io |
| Database | PostgreSQL (histórico) + Redis (pub/sub, presence) |
| Protocolos | WebSocket (Socket.io), REST (histórico), Kafka eventos |

## Responsabilidades
- Salas por licitação/tenant; threads por assunto.
- Mensagens em tempo real, presença/typing, anexos leves.
- Histórico persistido com RLS por tenant.
- Eventos de notificação (mencions) para Notification Service.

## Modelos
```sql
CREATE TABLE conversas (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  licitacao_id UUID,
  titulo TEXT,
  created_by UUID,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE mensagens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversa_id UUID REFERENCES conversas(id),
  tenant_id UUID NOT NULL,
  usuario_id UUID NOT NULL,
  conteudo TEXT,
  tipo VARCHAR(20) DEFAULT 'text',
  metadata JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_msg_conversa ON mensagens(conversa_id, created_at);
```

## Eventos (Kafka)
| Evento | Tópico | Payload |
|--------|--------|---------|
| `chat.message.sent` | `chat.events` | {conversaId, mensagemId, usuarioId, tenantId}
| `chat.user.typing` | `chat.events` | {conversaId, usuarioId}

## REST Endpoints
- `GET /api/v1/conversas?licitacaoId=...`
- `POST /api/v1/conversas` {licitacaoId, titulo}
- `GET /api/v1/conversas/:id/mensagens?cursor=...`
- `POST /api/v1/conversas/:id/mensagens` {conteudo, tipo}
Auth: JWT, header `X-Tenant-ID`.

## WebSocket Fluxo
- Conectar: `wss://api.cotai.com/ws/chat?token=...&tenantId=...`
- Eventos: `message:new`, `typing`, `presence`, `message:ack`.
- Backpressure: limitar 20 msgs/seg por socket; ban temporário se abuso.

## Segurança
- RLS por tenant; validação de participação na conversa.
- Sanitização HTML; limitar anexos (<=5MB) via signed URL S3.

## Métricas
- `chat_active_connections`
- `chat_messages_total{tenant}`
- `chat_delivery_latency_seconds` (p95)

## Health
- `/health/live` (process), `/health/ready` (Redis + DB).

*Integrações: Notification Service para push/email; Kanban para criar conversa por licitação.*