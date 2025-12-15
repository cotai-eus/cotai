# Notification Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Collaboration |
| Responsabilidade | Envio de push, email, SMS |
| Stack | Node.js 20 |
| Database | PostgreSQL (templates, logs) |
| Provedores | Firebase (push), SMTP/SES (email), SMS (Twilio) |

## Responsabilidades
- Orquestrar notificações multi-canal.
- Templates com variáveis (Mustache/Handlebars).
- Preferências por usuário/tenant; opt-in/out LGPD.
- Retries com DLQ (RabbitMQ opcional) e backoff.

## Modelo de Dados
```sql
CREATE TABLE notification_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID,
  code VARCHAR(50) UNIQUE,
  channel VARCHAR(20),
  subject TEXT,
  body TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE notification_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID,
  user_id UUID,
  channel VARCHAR(20),
  template_code VARCHAR(50),
  status VARCHAR(20),
  error TEXT,
  sent_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## API
- `POST /api/v1/notifications/send` {templateCode, channel, to, data}
- `GET /api/v1/notifications/logs?userId=...`
- `POST /api/v1/preferences` — salvar preferências por canal.
Auth: JWT + `X-Tenant-ID`.

## Eventos Consumidos
- `licitacao.status.changed` (Kafka) → enviar alertas.
- `chat.message.sent` → push para mencionados.
- `cotacao.received` → avisar gestor/comercial.

## Entregas por Canal
- Push: Firebase `fcm_token` (mobile/web push).
- Email: SES/SMTP; DKIM/SPF pré-config.
- SMS: Twilio (templates curtos); fallback push/email.

## Retries
- Tolerância a falha: 3 tentativas, backoff 30s/2m/10m.
- DLQ: `notifications.dlq` (RabbitMQ).

## Métricas
- `notifications_sent_total{channel,status}`
- `notification_latency_seconds{channel}`
- `notifications_bounce_total`

## Segurança
- Templates sanitizados; bloqueio de HTML perigoso.
- Opt-out respeitado; rate limit por user/tenant.

*Integra com Chat, Kanban, Quote; usa eventos de Kafka.*