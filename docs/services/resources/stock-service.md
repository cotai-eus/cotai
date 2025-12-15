# Stock Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Resource Mgmt |
| Responsabilidade | Controle de estoque e insumos |
| Stack | Node.js 20 |
| Database | PostgreSQL (tenant schema) |
| Protocolos | REST |

## Responsabilidades
- Cadastro de produtos internos e níveis de estoque.
- Reservas de itens para propostas/cotações.
- Movimentações de entrada/saída (NF, devolução).

## Endpoints
- `GET /produtos` — filtros: categoria, sku.
- `POST /produtos`
- `GET /produtos/:id`
- `PATCH /produtos/:id`
- `POST /produtos/:id/reservas` — reserva para licitação/cotação.
- `POST /produtos/:id/movimentacoes` — entrada/saída.
Headers: `Authorization`, `X-Tenant-ID`.

## Modelo de Dados
```sql
CREATE TABLE produtos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  sku VARCHAR(50) UNIQUE,
  nome VARCHAR(255) NOT NULL,
  categoria VARCHAR(100),
  unidade VARCHAR(10),
  estoque_atual NUMERIC DEFAULT 0,
  estoque_minimo NUMERIC DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE reservas (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  produto_id UUID REFERENCES produtos(id),
  licitacao_id UUID,
  quantidade NUMERIC,
  status VARCHAR(20) DEFAULT 'ativa',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE movimentacoes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  produto_id UUID REFERENCES produtos(id),
  tipo VARCHAR(20) CHECK (tipo IN ('entrada','saida')),
  quantidade NUMERIC,
  documento VARCHAR(100),
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Regras
- Reserva reduz `estoque_disponivel` virtual; confirmação libera/baixa.
- Não permitir estoque negativo.
- Alertas quando `estoque_atual < estoque_minimo`.

## Observabilidade
- Metrics: `stock_reservas_total{status}`, `stock_movimentacoes_total{tipo}`.

## Exemplo
```http
POST /v1/produtos
{
  "nome": "Cabo UTP Cat6",
  "sku": "CABO-CAT6-01",
  "categoria": "rede",
  "unidade": "UN",
  "estoque_minimo": 100
}
```

*Integra com Quote/CRM para disponibilidade de insumos.*