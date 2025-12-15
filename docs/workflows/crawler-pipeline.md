# Pipeline de Crawlers

## Fluxo
1) Scheduler cria jobs (BullMQ/Redis) e publica comando `crawler.start` (RabbitMQ).
2) Crawler Workers consomem fila, executam spiders (PNCP, ComprasNet...).
3) PDFs e anexos → S3; metadados brutos → evento `edital.raw` (Kafka).
4) Normalizer consome `edital.raw`, dedup/enriquece, persiste e publica `edital.normalized`.
5) Workflow Engine inicia processo da licitação.

## Controles
- Rate limit por fonte; retries com backoff; DLQ `cotai.commands.dlq`.
- Idempotência por `jobId` + `source_id`.

## Métricas
- `crawler_jobs_total{source,status}`
- `crawler_pages_processed` por job
- `crawler_duplicates_total`

## Operação
- HPA baseado em profundidade de fila.
- Circuit breaker em fontes instáveis.

*Integra Acquisition → Core Bidding via Kafka.*