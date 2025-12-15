# Quote Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Resource Mgmt |
| Responsabilidade | Gerenciar cotações de fornecedores |
| Stack | Node.js 20 |
| Database | PostgreSQL (tenant schema) |
| Protocolos | REST, Kafka events |

## Responsabilidades
- Enviar solicitações de cotação para fornecedores (via Notification Service).
- Receber propostas e registrar valores/prazos.
- Consolidar melhor cenário de preços para a licitação.

## Endpoints
- `POST /licitacoes/:id/solicitar-cotacoes` — dispara convites.
- `GET /licitacoes/:id/cotacoes` — lista propostas recebidas.
- `POST /licitacoes/:id/cotacoes` — fornecedor submete proposta.
- `POST /cotacoes/:cotacaoId/finalizar` — fecha cotação vencedora.
Headers: `Authorization`, `X-Tenant-ID` (supplier usa token portal).

## Modelo de Dados
```sql
CREATE TABLE cotacoes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  licitacao_id UUID NOT NULL,
  fornecedor_id UUID NOT NULL,
  valor_total DECIMAL(15,2) NOT NULL,
  prazo_entrega INTERVAL,
  validade_proposta DATE,
  status VARCHAR(20) DEFAULT 'recebida',
  itens JSONB,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Eventos (Kafka)
| Evento | Tópico | Payload |
|--------|--------|---------|
| `cotacao.requested` | `cotacao.lifecycle` | {licitacaoId, fornecedores[]}
| `cotacao.received` | `cotacao.received` | {licitacaoId, cotacaoId, valor}
| `cotacao.closed` | `cotacao.lifecycle` | {licitacaoId, cotacaoId, vencedor: fornecedorId}

## Regras
- Evitar propostas duplicadas do mesmo fornecedor para a mesma licitação.
- Validar prazo de validade >= data atual.
- Seleção vencedora pode considerar menor preço ou score ponderado.

## Observabilidade
- Metrics: `quote_requests_total`, `quote_received_total`, `quote_response_time_seconds`.

## Exemplo (fornecedor)
```http
POST /v1/licitacoes/LIC123/cotacoes
Authorization: Bearer <supplier_token>
{
  "fornecedor_id": "FORN456",
  "valor_total": 120000.00,
  "prazo_entrega": "30 days",
  "validade_proposta": "2026-01-15"
}
```

*Integra com Notification (convites) e Kanban (status COTADO).*