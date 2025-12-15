# CRM Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Resource Mgmt |
| Responsabilidade | Gestão de fornecedores e contatos |
| Stack | Node.js 20 + NestJS |
| Database | PostgreSQL (tenant schema) |
| Protocolos | REST, Kafka events |

## Responsabilidades
- CRUD de `Fornecedor`, contatos e documentos fiscais.
- Catálogo de produtos/serviços oferecidos pelo fornecedor.
- Integração com Kanban/Quote para sugestões automáticas.

## Endpoints (v1)
- `GET /fornecedores` — filtros: segmento, uf, status.
- `POST /fornecedores` — cria fornecedor.
- `GET /fornecedores/:id`
- `PATCH /fornecedores/:id` — dados cadastrais.
- `GET /fornecedores/:id/produtos`
- `POST /fornecedores/:id/produtos` — catálogo.
- `GET /fornecedores/:id/contatos`
- `POST /fornecedores/:id/contatos`
Headers: `Authorization`, `X-Tenant-ID`.

## Modelo de Dados
```sql
CREATE TABLE fornecedores (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  razao_social VARCHAR(255) NOT NULL,
  nome_fantasia VARCHAR(255),
  cnpj VARCHAR(14) UNIQUE NOT NULL,
  segmento VARCHAR(100),
  uf CHAR(2),
  cidade VARCHAR(100),
  status VARCHAR(20) DEFAULT 'ativo',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE fornecedores_produtos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fornecedor_id UUID REFERENCES fornecedores(id),
  descricao TEXT,
  categoria VARCHAR(100),
  preco_medio DECIMAL(15,2),
  unidade VARCHAR(10)
);

CREATE TABLE fornecedores_contatos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  fornecedor_id UUID REFERENCES fornecedores(id),
  nome VARCHAR(255),
  email VARCHAR(255),
  telefone VARCHAR(30)
);
```

## Eventos (Kafka)
| Evento | Tópico | Trigger |
|--------|--------|---------|
| `fornecedor.created` | `fornecedor.lifecycle` | novo fornecedor |
| `fornecedor.updated` | `fornecedor.lifecycle` | atualização |

## Autorização
- Roles: `admin`, `manager`, `commercial` (CRUD); `supplier` pode editar próprio perfil via portal.
- RLS por `tenant_id`.

## Observabilidade
- Metrics: `crm_requests_total`, `crm_db_latency_seconds`.
- Tracing: Jaeger interceptor NestJS.

## Exemplo
```http
POST /v1/fornecedores
Authorization: Bearer <token>
X-Tenant-ID: tenant_abc
{
  "razao_social": "ACME LTDA",
  "cnpj": "12345678000199",
  "segmento": "tecnologia",
  "uf": "SP"
}
```

*Relaciona-se com Kanban/Quote para matching automático.*