# Data Extractor

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Core Bidding |
| Responsabilidade | NLP para itens, prazos, valores e requisitos |
| Stack | Python 3.11 + spaCy + FastAPI |
| Storage | Elasticsearch 8 (index de busca) |
| Protocolos | Kafka consumer (ocr.completed), REST (search), gRPC opcional |

## Responsabilidades
- Consumir `ocr.completed` (Kafka) e processar texto.
- Extrair itens (descrição, qtde, unidade), prazos, valores estimados, requisitos.
- Calcular `score` de aderência (catálogo interno).
- Indexar no Elasticsearch para busca e similaridade.
- Publicar `nlp.completed` → Kafka para Workflow Engine / Kanban.

## Pipeline NLP
1) Limpeza de texto: lower, remover stopwords, normalizar números.
2) Sentence splitting; regex para datas/valores; modelos spaCy custom (PT-BR).
3) Extração de itens: regex + matcher de unidades (kg, un, m²).
4) Classificação de segmento: zero-shot / fastText pré-treinado.
5) Score de aderência: cosine similarity com catálogo de produtos.

## Evento de Entrada (Kafka `ocr.completed`)
```json
{
  "tenantId": "tenant_abc",
  "licitacaoId": "LIC123",
  "textS3Key": "ocr/tenant_abc/lic123.txt",
  "engine": "textract",
  "pages": 18
}
```

## Evento de Saída (Kafka `nlp.completed`)
```json
{
  "eventId": "evt_nlp_123",
  "tenantId": "tenant_abc",
  "licitacaoId": "LIC123",
  "items": [
    {"descricao": "Notebook 16GB", "quantidade": 10, "unidade": "UN"}
  ],
  "deadlines": {
    "abertura": "2025-12-20T14:00:00Z",
    "entrega": "2026-01-10T23:59:59Z"
  },
  "valorEstimado": 150000,
  "score": 0.82,
  "segmento": "tecnologia",
  "metadata": {"engine": "spacy"},
  "timestamp": "2025-12-15T10:40:00Z"
}
```

## REST Endpoints
- `GET /health`
- `GET /search?q=notebook+16gb&tenant_id=...` — busca no Elasticsearch.
- `GET /licitacoes/:id/extract` — retorna última extração.
Auth: JWT + `X-Tenant-ID`.

## Index Elasticsearch
```json
PUT /licitacoes
{
  "mappings": {
    "properties": {
      "licitacaoId": {"type": "keyword"},
      "tenantId": {"type": "keyword"},
      "texto": {"type": "text", "analyzer": "portuguese"},
      "items": {
        "type": "nested",
        "properties": {
          "descricao": {"type": "text", "analyzer": "portuguese"},
          "quantidade": {"type": "float"},
          "unidade": {"type": "keyword"}
        }
      },
      "deadlines": {"properties": {"abertura": {"type": "date"}}},
      "score": {"type": "float"},
      "segmento": {"type": "keyword"}
    }
  }
}
```

## Cálculo de Score
- Similaridade TF-IDF + embeddings (sentence-transformers) com catálogo interno.
- Score final: $score = 0.6\*sim + 0.4\*peso_segmento$.

## Variáveis de Ambiente
```bash
KAFKA_BROKERS=kafka:9092
ES_HOST=http://elasticsearch:9200
S3_BUCKET=cotai-ocr
MODEL_PATH=/models/spacy-pt
CATALOGO_INDEX=produtos
```

## Métricas
- `nlp_jobs_total{status}`
- `nlp_duration_seconds{quantile}`
- `nlp_score_histogram` (buckets)
- `nlp_items_extracted_total`

## Segurança
- Leitura S3 com credenciais limitadas.
- Escrita/Leitura ES com user dedicado.
- Sanitização de payloads antes de indexar.

*Referência: Workflow Engine e Kanban API consomem `nlp.completed`.*