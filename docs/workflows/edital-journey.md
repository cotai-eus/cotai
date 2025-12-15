# Jornada do Edital (Core)

## Etapas
1) **Descoberta**: Scheduler → RabbitMQ → Crawler coleta PNCP/portais.
2) **Normalização**: Normalizer dedup/enriquece; grava em PG; publica `edital.normalized`.
3) **Recebido**: Workflow Engine cria licitação e agenda OCR.
4) **Analisando**: OCR extrai texto; Data Extractor gera itens/prazos/score; Workflow decide.
5) **Cotar**: Kanban abre cotações; Notification envia convites; Quote registra respostas; CRM sugere fornecedores.
6) **Cotado**: Consolidar propostas; Workflow marca; agenda follow-up.
7) **Sem Resposta**: SLA 48h expirado ou score baixo.

## Eventos Principais (Kafka)
- `edital.normalized`
- `licitacao.status.changed`
- `ocr.completed`
- `nlp.completed`
- `cotacao.received`
- `notification.sent`

## SLAs
- OCR: até 5 minutos.
- NLP: até 1 minuto.
- Cotações: 48h padrão.

## Métricas de Negócio
- Lead time Recebido→Cotado.
- Taxa de resposta de fornecedores.
- Score médio de aderência.

*Fluxo detalha transições automáticas descritas no README.*