# Kanban API

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Core Bidding |
| Responsabilidade | CRUD de licitações e transições de status |
| Stack | Node.js 20 + NestJS |
| Database | PostgreSQL (schema tenant) |
| Protocolos | REST (público), gRPC interno, Kafka eventos |

## Responsabilidades
- CRUD de `Licitacao`, `ItemEdital`, `Cotacao` básica.
- Aplicar transições de status (manual/automática) conforme workflow.
- Expor endpoints para Frontend / Portal fornecedor.
- Publicar eventos de status para Kafka.

## Rotas REST (v1)
- `GET /licitacoes` — lista com filtros (status, data, orgão, valor).
- `POST /licitacoes` — cria licitação (uso interno/workflow).
- `GET /licitacoes/:id` — detalha.
- `PATCH /licitacoes/:id/status` — transição (body: `{to, reason}`).
- `GET /licitacoes/:id/itens` — itens extraídos/validados.
- `POST /licitacoes/:id/cotacoes` — registrar cotação de fornecedor.
- `GET /licitacoes/:id/historico` — histórico de transições.
Headers: `Authorization: Bearer`, `X-Tenant-ID`.

## gRPC (interno)
```protobuf
service LicitacaoService {
  rpc Get(GetRequest) returns (Licitacao);
  rpc UpdateStatus(UpdateStatusRequest) returns (Licitacao);
  rpc List(ListRequest) returns (stream Licitacao);
}
```

## Modelo de Dados (resumo)
```sql
CREATE TABLE licitacoes (
  id UUID PRIMARY KEY,
  numero VARCHAR(100),
  objeto TEXT,
  status VARCHAR(20) NOT NULL,
  score NUMERIC(5,2),
  data_abertura TIMESTAMPTZ,
  data_encerramento TIMESTAMPTZ,
  orgao VARCHAR(255),
  orgao_cnpj VARCHAR(14),
  valor_estimado DECIMAL(15,2),
  documentos JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE itens_edital (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  licitacao_id UUID REFERENCES licitacoes(id),
  descricao TEXT,
  quantidade NUMERIC,
  unidade VARCHAR(10),
  preco_estimado DECIMAL(15,2)
);

CREATE TABLE historico_status (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  licitacao_id UUID,
  from_status VARCHAR(20),
  to_status VARCHAR(20),
  reason TEXT,
  changed_by UUID,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Eventos (Kafka)
| Evento | Tópico | Payload |
|--------|--------|---------|
| `licitacao.created` | `licitacao.lifecycle` | {id, status, tenantId}
| `licitacao.status.changed` | `licitacao.status.changed` | {id, from, to, reason}
| `licitacao.cotacao.received` | `cotacao.received` | {licitacaoId, fornecedorId, valor}

## Regras de Negócio
- Apenas transições válidas (consultar Workflow Engine).
- `status` protegido: só Workflow Engine altera estados automáticos.
- Row-Level Security com `tenant_id` em todas as tabelas.

## Autenticação/Autorização
- JWT via Auth Service; roles: `admin`, `manager`, `commercial`, `supplier`.
- Fornecedor (supplier) só pode criar cotação na sua licitação.

## Observabilidade
- Metrics: `kanban_requests_total{route,status}`, `kanban_db_latency_seconds`.
- Tracing com Jaeger (NestJS interceptor).

## Exemplo de Transição
```http
PATCH /v1/licitacoes/LIC123/status
Authorization: Bearer <token>
X-Tenant-ID: tenant_abc
Content-Type: application/json
{
  "to": "COTAR",
  "reason": "score >= 0.7"
}
```

## Health
- `/health/live` — processo
- `/health/ready` — DB e Kafka alcançáveis

*Referência: usa eventos descritos em docs/architecture/communication-patterns.md*