# Scheduler Service

## Visão Geral

| Atributo | Valor |
|----------|-------|
| **Bounded Context** | Acquisition & Ingestion |
| **Responsabilidade** | Agendamento e disparo de jobs de crawlers |
| **Stack** | Node.js 20+ + BullMQ |
| **Database** | Redis (queues, scheduling) |
| **Porta** | 3010 |
| **Protocolo** | REST + RabbitMQ Publisher |

---

## Responsabilidades

1. **Gerenciar schedules** de crawlers (CRON expressions)
2. **Disparar jobs** para RabbitMQ
3. **Controlar rate limiting** por fonte
4. **Monitorar status** de jobs em execução
5. **Retry automático** com backoff exponencial
6. **Dashboard de monitoramento** de filas

---

## Arquitetura

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ Admin Portal │────►│  Scheduler   │────►│   RabbitMQ   │
│   (config)   │     │   Service    │     │ crawler.jobs │
└──────────────┘     └──────┬───────┘     └──────┬───────┘
                            │                    │
                     ┌──────▼──────┐      ┌──────▼──────┐
                     │    Redis    │      │   Crawler   │
                     │  (BullMQ)   │      │   Workers   │
                     └─────────────┘      └─────────────┘
```

---

## API Endpoints

### Gerenciar Schedules

#### Listar Schedules

```http
GET /api/v1/schedules
Authorization: Bearer <token>
X-Tenant-ID: <tenant_id>
```

**Response:**
```json
{
  "data": [
    {
      "id": "sch_abc123",
      "name": "PNCP Nacional",
      "source": "pncp",
      "cron": "*/30 * * * *",
      "enabled": true,
      "filters": {
        "modalidades": ["pregao_eletronico", "concorrencia"],
        "ufs": ["SP", "RJ", "MG"],
        "segmentos": ["tecnologia", "servicos"]
      },
      "lastRun": "2025-12-15T10:00:00Z",
      "nextRun": "2025-12-15T10:30:00Z",
      "status": "idle"
    }
  ]
}
```

#### Criar Schedule

```http
POST /api/v1/schedules
Authorization: Bearer <token>
X-Tenant-ID: <tenant_id>
Content-Type: application/json

{
  "name": "PNCP - Pregão SP",
  "source": "pncp",
  "cron": "0 */2 * * *",
  "enabled": true,
  "filters": {
    "modalidades": ["pregao_eletronico"],
    "ufs": ["SP"],
    "segmentos": ["tecnologia"]
  },
  "config": {
    "maxPages": 10,
    "timeout": 60000,
    "retries": 3
  }
}
```

**Response (201):**
```json
{
  "id": "sch_xyz789",
  "name": "PNCP - Pregão SP",
  "source": "pncp",
  "cron": "0 */2 * * *",
  "enabled": true,
  "nextRun": "2025-12-15T12:00:00Z",
  "createdAt": "2025-12-15T10:35:00Z"
}
```

#### Atualizar Schedule

```http
PATCH /api/v1/schedules/{schedule_id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "cron": "0 */4 * * *",
  "enabled": false
}
```

#### Deletar Schedule

```http
DELETE /api/v1/schedules/{schedule_id}
Authorization: Bearer <token>
```

### Executar Jobs Manualmente

#### Trigger Imediato

```http
POST /api/v1/schedules/{schedule_id}/run
Authorization: Bearer <token>

{
  "priority": "high",
  "filters": {
    "dataInicio": "2025-12-01",
    "dataFim": "2025-12-15"
  }
}
```

**Response (202):**
```json
{
  "jobId": "job_def456",
  "scheduleId": "sch_abc123",
  "status": "queued",
  "queuedAt": "2025-12-15T10:35:00Z"
}
```

### Monitorar Jobs

#### Status do Job

```http
GET /api/v1/jobs/{job_id}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": "job_def456",
  "scheduleId": "sch_abc123",
  "status": "running",
  "progress": 45,
  "startedAt": "2025-12-15T10:35:05Z",
  "stats": {
    "pagesProcessed": 5,
    "editaisFound": 23,
    "editaisNew": 18,
    "errors": 0
  }
}
```

#### Listar Jobs Recentes

```http
GET /api/v1/jobs?status=completed&limit=20
Authorization: Bearer <token>
```

#### Cancelar Job

```http
POST /api/v1/jobs/{job_id}/cancel
Authorization: Bearer <token>
```

---

## Modelo de Dados (Redis)

```typescript
// src/types/schedule.ts
interface Schedule {
  id: string;
  tenantId: string;
  name: string;
  source: 'pncp' | 'comprasnet' | 'bec' | 'custom';
  cron: string;
  enabled: boolean;
  filters: {
    modalidades?: string[];
    ufs?: string[];
    municipios?: string[];
    segmentos?: string[];
    palavrasChave?: string[];
    valorMinimo?: number;
    valorMaximo?: number;
  };
  config: {
    maxPages: number;
    timeout: number;
    retries: number;
    rateLimit: number; // requests per minute
  };
  lastRun?: Date;
  nextRun?: Date;
  createdAt: Date;
  updatedAt: Date;
}

interface Job {
  id: string;
  scheduleId: string;
  tenantId: string;
  status: 'queued' | 'running' | 'completed' | 'failed' | 'cancelled';
  priority: number;
  progress: number;
  attempts: number;
  filters: Record<string, any>;
  stats: {
    pagesProcessed: number;
    editaisFound: number;
    editaisNew: number;
    editaisDuplicate: number;
    errors: number;
  };
  error?: string;
  startedAt?: Date;
  completedAt?: Date;
  createdAt: Date;
}
```

---

## Implementação BullMQ

```typescript
// src/queues/crawler.queue.ts
import { Queue, Worker, QueueScheduler } from 'bullmq';
import { connection } from './redis';

// Queue para jobs de crawler
export const crawlerQueue = new Queue('crawler-jobs', { connection });

// Scheduler para cron jobs
export const crawlerScheduler = new QueueScheduler('crawler-jobs', { connection });

// Adicionar job recorrente
export async function addScheduledJob(schedule: Schedule): Promise<void> {
  await crawlerQueue.add(
    `crawler:${schedule.source}`,
    {
      scheduleId: schedule.id,
      tenantId: schedule.tenantId,
      source: schedule.source,
      filters: schedule.filters,
      config: schedule.config,
    },
    {
      repeat: {
        pattern: schedule.cron,
        tz: 'America/Sao_Paulo',
      },
      jobId: schedule.id,
      removeOnComplete: 100,
      removeOnFail: 50,
    }
  );
}

// Publicar para RabbitMQ (worker)
import { AmqpConnection } from '@golevelup/nestjs-rabbitmq';

@Injectable()
export class CrawlerJobProcessor {
  constructor(private readonly amqp: AmqpConnection) {}

  @Processor('crawler-jobs')
  async process(job: Job<CrawlerJobData>): Promise<void> {
    const { scheduleId, tenantId, source, filters, config } = job.data;
    
    // Publicar para RabbitMQ (workers Python)
    await this.amqp.publish('cotai.commands', 'crawler.start', {
      jobId: job.id,
      scheduleId,
      tenantId,
      source,
      filters,
      config,
      timestamp: new Date().toISOString(),
    });
    
    // Aguardar resultado via callback ou polling
  }
}
```

---

## Integração RabbitMQ

```typescript
// src/messaging/rabbit.config.ts
import { RabbitMQModule } from '@golevelup/nestjs-rabbitmq';

@Module({
  imports: [
    RabbitMQModule.forRoot({
      exchanges: [
        { name: 'cotai.commands', type: 'direct' },
        { name: 'cotai.events', type: 'topic' },
      ],
      uri: process.env.RABBITMQ_URL,
      connectionInitOptions: { wait: true },
    }),
  ],
})
export class MessagingModule {}

// Consumir resultados
@RabbitSubscribe({
  exchange: 'cotai.events',
  routingKey: 'crawler.completed',
  queue: 'scheduler.crawler-results',
})
async handleCrawlerCompleted(payload: CrawlerResult): Promise<void> {
  await this.jobService.updateJobStatus(payload.jobId, 'completed', payload.stats);
  await this.scheduleService.updateLastRun(payload.scheduleId);
}
```

---

## Fontes Suportadas

| Fonte | URL Base | Rate Limit | Auth |
|-------|----------|------------|------|
| PNCP | api.pncp.gov.br | 60 req/min | API Key |
| ComprasNet | comprasnet.gov.br | 30 req/min | - |
| BEC-SP | bec.sp.gov.br | 20 req/min | - |
| Licitações-e | licitacoes-e.com.br | 30 req/min | - |

---

## Variáveis de Ambiente

```bash
# Server
PORT=3010
NODE_ENV=production

# Redis
REDIS_URL=redis://redis:6379

# RabbitMQ
RABBITMQ_URL=amqp://user:pass@rabbitmq:5672

# Rate Limiting
RATE_LIMIT_PNCP=60
RATE_LIMIT_COMPRASNET=30

# Monitoring
BULL_BOARD_ENABLED=true
BULL_BOARD_PATH=/admin/queues
BULL_BOARD_USERNAME=admin
BULL_BOARD_PASSWORD=${BULL_BOARD_PASSWORD}
```

---

## Dockerfile

```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./
EXPOSE 3010
CMD ["node", "dist/main.js"]
```

---

## Métricas

```yaml
# Prometheus
scheduler_jobs_total{source="pncp", status="completed"} 1523
scheduler_jobs_total{source="pncp", status="failed"} 12
scheduler_job_duration_seconds{source="pncp", quantile="0.99"} 45.2
scheduler_queue_size{queue="crawler-jobs"} 5
scheduler_editais_discovered_total{source="pncp"} 8934
```

---

## Dashboard BullMQ

Acessível em `/admin/queues` com autenticação básica:

- Visualização de filas
- Jobs em execução
- Histórico de falhas
- Retry manual
- Pause/Resume queues

---

## Exemplo de Uso

```bash
# Criar schedule via API
curl -X POST https://api.cotai.com.br/v1/schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "PNCP Diário",
    "source": "pncp",
    "cron": "0 6 * * *",
    "filters": {
      "ufs": ["SP"],
      "modalidades": ["pregao_eletronico"]
    }
  }'

# Executar manualmente
curl -X POST https://api.cotai.com.br/v1/schedules/sch_abc123/run \
  -H "Authorization: Bearer $TOKEN"
```

---

*Referência: [Crawler Workers](./crawler-workers.md) | [Normalizer](./normalizer.md)*
