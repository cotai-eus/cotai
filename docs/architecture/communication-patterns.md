# Padrões de Comunicação Inter-Serviços

## Visão Geral

A plataforma CotAI utiliza comunicação híbrida: **síncrona** para queries e validações críticas, **assíncrona** para eventos de domínio e processamento pesado.

---

## Matriz de Comunicação

| Padrão | Uso | Tecnologia | Latência |
|--------|-----|------------|----------|
| REST | APIs públicas, CRUD | HTTP/JSON | < 100ms |
| gRPC | Comunicação interna | Protocol Buffers | < 20ms |
| RabbitMQ | Comandos, filas de trabalho | AMQP | eventual |
| Kafka | Eventos de domínio | Event streaming | eventual |

---

## REST APIs (Público)

### Convenções

```yaml
Base URL: https://api.cotai.com.br/v1
Content-Type: application/json
Authorization: Bearer <jwt_token>
X-Tenant-ID: <tenant_uuid> (opcional se JWT contém tenant)
X-Request-ID: <uuid> (tracing)
```

### Padrão de Response

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  },
  "links": {
    "self": "/v1/licitacoes?page=1",
    "next": "/v1/licitacoes?page=2"
  }
}
```

### Códigos de Erro

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dados inválidos",
    "details": [
      { "field": "cnpj", "message": "CNPJ inválido" }
    ]
  }
}
```

| Código HTTP | Uso |
|-------------|-----|
| 200 | Sucesso |
| 201 | Criado |
| 400 | Validação |
| 401 | Não autenticado |
| 403 | Não autorizado |
| 404 | Não encontrado |
| 409 | Conflito |
| 422 | Entidade não processável |
| 429 | Rate limit |
| 500 | Erro interno |

---

## gRPC (Interno)

### Definição de Serviço

```protobuf
// proto/tenant.proto
syntax = "proto3";

package identity;

service TenantService {
  rpc GetTenant(GetTenantRequest) returns (TenantResponse);
  rpc ValidateTenant(ValidateTenantRequest) returns (ValidationResponse);
  rpc ListTenants(ListTenantsRequest) returns (stream TenantResponse);
}

message GetTenantRequest {
  string tenant_id = 1;
}

message TenantResponse {
  string id = 1;
  string name = 2;
  string schema_name = 3;
  string status = 4;
  int64 created_at = 5;
}

message ValidationResponse {
  bool valid = 1;
  string message = 2;
}
```

### Cliente Go

```go
// internal/grpc/tenant_client.go
package grpc

import (
    "context"
    pb "cotai/proto"
    "google.golang.org/grpc"
)

type TenantClient struct {
    conn   *grpc.ClientConn
    client pb.TenantServiceClient
}

func NewTenantClient(addr string) (*TenantClient, error) {
    conn, err := grpc.Dial(addr, grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return &TenantClient{
        conn:   conn,
        client: pb.NewTenantServiceClient(conn),
    }, nil
}

func (c *TenantClient) GetTenant(ctx context.Context, id string) (*pb.TenantResponse, error) {
    return c.client.GetTenant(ctx, &pb.GetTenantRequest{TenantId: id})
}
```

---

## RabbitMQ (Comandos)

### Topologia

```
Exchange: cotai.commands (direct)
├── Queue: crawler.jobs
│   └── Routing Key: crawler.start
├── Queue: ocr.process
│   └── Routing Key: ocr.extract
├── Queue: notification.send
│   └── Routing Key: notification.*
└── DLQ: cotai.commands.dlq
    └── Routing Key: #.failed
```

### Publicação de Comando

```typescript
// src/messaging/rabbit.publisher.ts
import { Injectable } from '@nestjs/common';
import { AmqpConnection } from '@golevelup/nestjs-rabbitmq';

@Injectable()
export class CommandPublisher {
  constructor(private readonly amqp: AmqpConnection) {}

  async publishCrawlerJob(job: CrawlerJobPayload): Promise<void> {
    await this.amqp.publish('cotai.commands', 'crawler.start', {
      jobId: job.id,
      source: job.source,
      filters: job.filters,
      tenantId: job.tenantId,
      timestamp: new Date().toISOString(),
    }, {
      persistent: true,
      headers: {
        'x-retry-count': 0,
        'x-tenant-id': job.tenantId,
      },
    });
  }
}
```

### Consumidor com Retry

```typescript
// src/messaging/crawler.consumer.ts
import { RabbitSubscribe, Nack } from '@golevelup/nestjs-rabbitmq';

@Injectable()
export class CrawlerConsumer {
  @RabbitSubscribe({
    exchange: 'cotai.commands',
    routingKey: 'crawler.start',
    queue: 'crawler.jobs',
    queueOptions: {
      durable: true,
      arguments: {
        'x-dead-letter-exchange': 'cotai.commands.dlq',
        'x-message-ttl': 300000, // 5 min
      },
    },
  })
  async handleCrawlerJob(payload: CrawlerJobPayload, msg: ConsumeMessage) {
    const retryCount = msg.properties.headers['x-retry-count'] || 0;
    
    try {
      await this.crawlerService.execute(payload);
    } catch (error) {
      if (retryCount < 3) {
        // Retry com backoff exponencial
        return new Nack(true); // requeue
      }
      // Enviar para DLQ
      return new Nack(false);
    }
  }
}
```

---

## Kafka (Eventos de Domínio)

### Tópicos

| Tópico | Produtor | Consumidores | Retention |
|--------|----------|--------------|-----------|
| `edital.discovered` | Normalizer | Workflow Engine | 7 dias |
| `licitacao.status.changed` | Kanban API | Audit, Notification | 30 dias |
| `cotacao.received` | Quote Service | Workflow, CRM | 7 dias |
| `audit.log` | Todos | Audit Service | 90 dias |

### Estrutura de Evento

```json
{
  "eventId": "evt_abc123",
  "eventType": "licitacao.status.changed",
  "aggregateId": "lic_xyz789",
  "aggregateType": "Licitacao",
  "tenantId": "tenant_123",
  "timestamp": "2025-12-15T10:30:00Z",
  "version": 1,
  "payload": {
    "previousStatus": "RECEBIDO",
    "newStatus": "ANALISANDO",
    "changedBy": "user_456",
    "reason": "OCR iniciado"
  },
  "metadata": {
    "correlationId": "req_def456",
    "causationId": "evt_ghi789"
  }
}
```

### Produtor (Node.js)

```typescript
// src/messaging/kafka.producer.ts
import { Injectable } from '@nestjs/common';
import { Kafka, Producer } from 'kafkajs';

@Injectable()
export class EventPublisher {
  private producer: Producer;

  constructor() {
    const kafka = new Kafka({
      clientId: 'kanban-api',
      brokers: process.env.KAFKA_BROKERS.split(','),
    });
    this.producer = kafka.producer();
  }

  async publishStatusChanged(event: StatusChangedEvent): Promise<void> {
    await this.producer.send({
      topic: 'licitacao.status.changed',
      messages: [{
        key: event.aggregateId,
        value: JSON.stringify(event),
        headers: {
          'tenant-id': event.tenantId,
          'event-type': event.eventType,
        },
      }],
    });
  }
}
```

### Consumidor (Go)

```go
// internal/kafka/consumer.go
package kafka

import (
    "context"
    "github.com/segmentio/kafka-go"
)

type EventConsumer struct {
    reader *kafka.Reader
}

func NewEventConsumer(brokers []string, topic, groupID string) *EventConsumer {
    return &EventConsumer{
        reader: kafka.NewReader(kafka.ReaderConfig{
            Brokers:  brokers,
            Topic:    topic,
            GroupID:  groupID,
            MinBytes: 10e3,
            MaxBytes: 10e6,
        }),
    }
}

func (c *EventConsumer) Consume(ctx context.Context, handler func(Event) error) error {
    for {
        msg, err := c.reader.ReadMessage(ctx)
        if err != nil {
            return err
        }
        
        var event Event
        json.Unmarshal(msg.Value, &event)
        
        if err := handler(event); err != nil {
            // Log e continua (at-least-once)
            log.Printf("Error processing event: %v", err)
        }
    }
}
```

---

## Circuit Breaker

```typescript
// src/resilience/circuit-breaker.ts
import CircuitBreaker from 'opossum';

const options = {
  timeout: 3000,
  errorThresholdPercentage: 50,
  resetTimeout: 30000,
};

const breaker = new CircuitBreaker(asyncFunction, options);

breaker.on('open', () => console.log('Circuit opened'));
breaker.on('halfOpen', () => console.log('Circuit half-open'));
breaker.on('close', () => console.log('Circuit closed'));

// Uso
await breaker.fire(params);
```

---

## Diagrama de Fluxo

```
┌──────────┐    REST     ┌─────────────┐    gRPC    ┌──────────────┐
│ Frontend │ ──────────► │ API Gateway │ ─────────► │ Auth Service │
└──────────┘             └─────────────┘            └──────────────┘
                               │
                               │ REST
                               ▼
                        ┌─────────────┐
                        │ Kanban API  │
                        └─────────────┘
                               │
              ┌────────────────┼────────────────┐
              │ Kafka          │                │ RabbitMQ
              ▼                ▼                ▼
       ┌────────────┐   ┌────────────┐   ┌────────────┐
       │   Audit    │   │ Workflow   │   │    OCR     │
       │  Service   │   │  Engine    │   │  Service   │
       └────────────┘   └────────────┘   └────────────┘
```

---

*Referência: [README.md](../../README.md) - Seção 3.2*
