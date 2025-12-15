# Agenda Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Collaboration |
| Responsabilidade | Eventos, lembretes e SLAs |
| Stack | Node.js 20 |
| Database | PostgreSQL |
| Protocolos | REST; publica eventos em Kafka |

## Responsabilidades
- Calendário por tenant/usuário.
- Lembretes de prazos (OCR, cotações, entregas).
- Integração com Notification Service para alertas.

## Modelo de Dados
```sql
CREATE TABLE eventos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  titulo TEXT NOT NULL,
  descricao TEXT,
  inicio TIMESTAMPTZ NOT NULL,
  fim TIMESTAMPTZ,
  licitacao_id UUID,
  tipo VARCHAR(30),
  created_by UUID,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE lembretes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  evento_id UUID REFERENCES eventos(id),
  offset_minutes INT NOT NULL,
  channel VARCHAR(20) DEFAULT 'push',
  status VARCHAR(20) DEFAULT 'pendente'
);
```

## Endpoints
- `GET /api/v1/eventos?from=&to=`
- `POST /api/v1/eventos` — cria evento/lembretes.
- `PATCH /api/v1/eventos/:id`
- `DELETE /api/v1/eventos/:id`
Auth: JWT + `X-Tenant-ID`.

## Lembretes → Kafka
Evento `agenda.reminder.due` em tópico `agenda.events`:
```json
{ "eventoId": "EVT123", "tenantId": "tenant_abc", "channel": "push" }
```
Notification Service consome e envia.

## SLA e Schedules
- Worker agenda roda a cada minuto (cron) para disparar lembretes vencidos.

## Métricas
- `agenda_events_total{status}`
- `agenda_reminders_sent_total{channel}`

*Relaciona-se com Kanban (prazos) e Notification (envio).*