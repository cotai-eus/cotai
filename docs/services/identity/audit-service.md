# Audit Service

## Visão Geral

| Atributo | Valor |
|----------|-------|
| **Bounded Context** | Identity & Compliance |
| **Responsabilidade** | Logs de auditoria, compliance LGPD, retenção de dados |
| **Stack** | Go 1.21+ |
| **Storage** | Kafka → S3 (cold) / ClickHouse (query) |
| **Porta** | 8083 |
| **Protocolo** | gRPC + Kafka Consumer |

---

## Responsabilidades

1. **Consumir eventos de auditoria** de todos os serviços
2. **Armazenar logs imutáveis** (append-only)
3. **Indexar para busca** rápida (ClickHouse)
4. **Arquivar em storage frio** (S3) para compliance
5. **Gerenciar retenção** (90 dias hot, 7 anos cold)
6. **Gerar relatórios** de compliance LGPD

---

## Arquitetura

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Serviços   │────►│    Kafka     │────►│Audit Service │
│  (eventos)   │     │ audit.events │     │  (consumer)  │
└──────────────┘     └──────────────┘     └──────┬───────┘
                                                 │
                           ┌─────────────────────┼─────────────────────┐
                           │                     │                     │
                    ┌──────▼──────┐       ┌──────▼──────┐       ┌──────▼──────┐
                    │ ClickHouse  │       │     S3      │       │  Prometheus │
                    │   (query)   │       │   (cold)    │       │  (metrics)  │
                    └─────────────┘       └─────────────┘       └─────────────┘
```

---

## API Endpoints

### Query API (REST)

#### Buscar Logs de Auditoria

```http
GET /api/v1/audit/logs?tenant_id={id}&from=2025-12-01&to=2025-12-15
Authorization: Bearer <admin_token>
```

**Query Parameters:**
| Param | Tipo | Descrição |
|-------|------|-----------|
| `tenant_id` | UUID | Filtrar por tenant |
| `user_id` | UUID | Filtrar por usuário |
| `action` | string | Tipo de ação (e.g., `licitacao.create`) |
| `resource_type` | string | Tipo de recurso |
| `resource_id` | UUID | ID do recurso |
| `from` | date | Data início |
| `to` | date | Data fim |
| `page` | int | Página |
| `per_page` | int | Items por página (max 100) |

**Response:**
```json
{
  "data": [
    {
      "id": "log_abc123",
      "timestamp": "2025-12-15T10:30:00Z",
      "tenantId": "tenant_123",
      "userId": "user_456",
      "action": "licitacao.status.update",
      "resourceType": "Licitacao",
      "resourceId": "lic_789",
      "changes": {
        "before": { "status": "RECEBIDO" },
        "after": { "status": "ANALISANDO" }
      },
      "context": {
        "ip": "192.168.1.100",
        "userAgent": "Mozilla/5.0...",
        "requestId": "req_def"
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 1523
  }
}
```

#### Exportar Logs (LGPD)

```http
POST /api/v1/audit/export
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "tenantId": "tenant_123",
  "userId": "user_456",
  "from": "2025-01-01",
  "to": "2025-12-31",
  "format": "json",
  "includeDeleted": true
}
```

**Response (202):**
```json
{
  "exportId": "exp_xyz789",
  "status": "processing",
  "estimatedSize": "15MB",
  "downloadUrl": null,
  "expiresAt": null
}
```

#### Verificar Status da Exportação

```http
GET /api/v1/audit/export/{export_id}
Authorization: Bearer <admin_token>
```

**Response:**
```json
{
  "exportId": "exp_xyz789",
  "status": "completed",
  "downloadUrl": "https://s3.../exports/exp_xyz789.json.gz",
  "expiresAt": "2025-12-16T10:30:00Z",
  "fileSize": "14.2MB"
}
```

#### Relatório de Acesso (LGPD Art. 18)

```http
GET /api/v1/audit/reports/data-access?user_id={id}&period=30d
Authorization: Bearer <admin_token>
```

---

### gRPC API

```protobuf
// proto/audit.proto
syntax = "proto3";
package audit.v1;

service AuditService {
  rpc LogEvent(AuditEvent) returns (LogResponse);
  rpc QueryLogs(QueryRequest) returns (stream AuditLog);
  rpc GetUserDataReport(UserDataRequest) returns (DataReport);
}

message AuditEvent {
  string event_id = 1;
  string tenant_id = 2;
  string user_id = 3;
  string action = 4;
  string resource_type = 5;
  string resource_id = 6;
  bytes changes = 7; // JSON
  bytes context = 8; // JSON
  int64 timestamp = 9;
}

message AuditLog {
  string id = 1;
  string tenant_id = 2;
  string user_id = 3;
  string action = 4;
  string resource_type = 5;
  string resource_id = 6;
  string changes_json = 7;
  string context_json = 8;
  int64 timestamp = 9;
}
```

---

## Estrutura de Evento

### Formato Padrão (Kafka)

```json
{
  "eventId": "evt_abc123def456",
  "eventType": "audit.log",
  "version": "1.0",
  "timestamp": "2025-12-15T10:30:00.000Z",
  "source": "kanban-api",
  "tenantId": "tenant_123",
  "correlationId": "req_xyz789",
  "payload": {
    "userId": "user_456",
    "action": "licitacao.status.update",
    "resourceType": "Licitacao",
    "resourceId": "lic_789",
    "changes": {
      "before": { "status": "RECEBIDO" },
      "after": { "status": "ANALISANDO" }
    },
    "context": {
      "ip": "192.168.1.100",
      "userAgent": "Mozilla/5.0...",
      "sessionId": "sess_abc"
    }
  }
}
```

---

## Modelo de Dados (ClickHouse)

```sql
CREATE TABLE audit_logs (
    id UUID,
    tenant_id UUID,
    user_id UUID,
    action LowCardinality(String),
    resource_type LowCardinality(String),
    resource_id UUID,
    changes String, -- JSON
    context String, -- JSON
    ip IPv4,
    user_agent String,
    request_id String,
    correlation_id String,
    source LowCardinality(String),
    timestamp DateTime64(3),
    
    INDEX idx_action action TYPE bloom_filter GRANULARITY 4,
    INDEX idx_resource resource_type TYPE bloom_filter GRANULARITY 4
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, timestamp, id)
TTL timestamp + INTERVAL 90 DAY TO VOLUME 's3_cold';

-- Materialized view para métricas
CREATE MATERIALIZED VIEW audit_stats
ENGINE = SummingMergeTree()
ORDER BY (tenant_id, action, date)
AS SELECT
    tenant_id,
    action,
    toDate(timestamp) as date,
    count() as count
FROM audit_logs
GROUP BY tenant_id, action, date;
```

---

## Consumer Kafka

```go
// internal/consumer/audit_consumer.go
package consumer

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
)

type AuditConsumer struct {
    reader     *kafka.Reader
    clickhouse *clickhouse.Conn
    s3Client   *s3.Client
}

func (c *AuditConsumer) Start(ctx context.Context) error {
    batch := make([]AuditEvent, 0, 1000)
    ticker := time.NewTicker(5 * time.Second)
    
    for {
        select {
        case <-ctx.Done():
            return c.flush(batch)
        case <-ticker.C:
            if len(batch) > 0 {
                c.flush(batch)
                batch = batch[:0]
            }
        default:
            msg, err := c.reader.ReadMessage(ctx)
            if err != nil {
                continue
            }
            
            var event AuditEvent
            json.Unmarshal(msg.Value, &event)
            batch = append(batch, event)
            
            if len(batch) >= 1000 {
                c.flush(batch)
                batch = batch[:0]
            }
        }
    }
}

func (c *AuditConsumer) flush(events []AuditEvent) error {
    // Batch insert no ClickHouse
    batch, _ := c.clickhouse.PrepareBatch(ctx, "INSERT INTO audit_logs")
    for _, e := range events {
        batch.Append(e.ID, e.TenantID, e.UserID, ...)
    }
    return batch.Send()
}
```

---

## S3 Archiving

```go
// internal/archiver/s3_archiver.go
package archiver

func (a *S3Archiver) ArchiveOldLogs(ctx context.Context, olderThan time.Time) error {
    // 1. Query logs antigos do ClickHouse
    rows, _ := a.ch.Query(ctx, `
        SELECT * FROM audit_logs 
        WHERE timestamp < ? 
        ORDER BY timestamp
        LIMIT 100000
    `, olderThan)
    
    // 2. Comprimir e enviar para S3
    var buf bytes.Buffer
    gzw := gzip.NewWriter(&buf)
    enc := json.NewEncoder(gzw)
    
    for rows.Next() {
        var log AuditLog
        rows.ScanStruct(&log)
        enc.Encode(log)
    }
    gzw.Close()
    
    key := fmt.Sprintf("audit/%s/%s.json.gz", 
        olderThan.Format("2006/01"), 
        uuid.New().String())
    
    _, err := a.s3.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String("cotai-audit-archive"),
        Key:    aws.String(key),
        Body:   bytes.NewReader(buf.Bytes()),
    })
    
    // 3. Deletar do ClickHouse após confirmação S3
    if err == nil {
        a.ch.Exec(ctx, "ALTER TABLE audit_logs DELETE WHERE timestamp < ?", olderThan)
    }
    
    return err
}
```

---

## Variáveis de Ambiente

```bash
# Server
PORT=8083
GRPC_PORT=9083

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=audit.events
KAFKA_GROUP_ID=audit-service

# ClickHouse
CLICKHOUSE_HOST=clickhouse:9000
CLICKHOUSE_DATABASE=audit
CLICKHOUSE_USER=audit
CLICKHOUSE_PASSWORD=${CH_PASSWORD}

# S3
AWS_REGION=sa-east-1
S3_BUCKET=cotai-audit-archive
S3_PREFIX=audit/

# Retention
RETENTION_HOT_DAYS=90
RETENTION_COLD_YEARS=7
```

---

## Docker Compose

```yaml
services:
  audit-service:
    build: ./audit-service
    ports:
      - "8083:8083"
      - "9083:9083"
    environment:
      KAFKA_BROKERS: kafka:9092
      CLICKHOUSE_HOST: clickhouse:9000
    depends_on:
      - kafka
      - clickhouse
      
  clickhouse:
    image: clickhouse/clickhouse-server:23.8
    volumes:
      - clickhouse_data:/var/lib/clickhouse
    ports:
      - "8123:8123"
      - "9000:9000"
```

---

## Métricas

```yaml
# Prometheus
audit_events_processed_total{source="kanban-api", action="licitacao.create"} 15234
audit_events_failed_total{reason="parse_error"} 12
audit_storage_bytes{tier="hot"} 5368709120
audit_storage_bytes{tier="cold"} 107374182400
audit_query_duration_seconds{quantile="0.99"} 0.45
```

---

## LGPD Compliance

| Requisito | Implementação |
|-----------|---------------|
| Art. 18, I (confirmação) | Query logs por user_id |
| Art. 18, II (acesso) | Endpoint de exportação |
| Art. 18, VI (eliminação) | Anonimização, não deleção física |
| Art. 37 (registro) | Todos eventos armazenados |
| Art. 46 (segurança) | Encryption, access control |

---

*Referência: [Segurança](../architecture/security-compliance.md) | [Kafka Patterns](../architecture/communication-patterns.md)*
