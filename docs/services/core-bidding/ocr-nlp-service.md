# OCR/NLP Service

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Bounded Context | Core Bidding |
| Responsabilidade | Extrair texto de PDFs e gerar texto bruto |
| Stack | Python 3.11, Tesseract 5 / AWS Textract, FastAPI |
| Storage | S3 (PDFs e texto extraído) |
| Protocolos | RabbitMQ (comandos), REST (status), Kafka (eventos) |

## Responsabilidades
- Consumir comandos `ocr.extract` (RabbitMQ) enviados pelo Workflow Engine.
- Realizar OCR (Tesseract local) ou delegar a Textract.
- Armazenar texto extraído em S3 (`s3://cotai-ocr/{tenant}/{licitacao}.txt`).
- Publicar `ocr.completed` em Kafka para o Data Extractor.

## Fluxo
1) Mensagem `ocr.extract` recebida com `s3PdfKey`.
2) Baixa PDF do S3, escolhe engine (Textract para >20p ou imagem complexa).
3) Extrai texto, normaliza encoding, remove duplicações de cabeçalho/rodapé.
4) Salva texto em S3 e emite evento `ocr.completed`.

## Mensagem de Comando (RabbitMQ)
```json
{
  "eventId": "evt_ocr_123",
  "licitacaoId": "LIC123",
  "tenantId": "tenant_abc",
  "s3PdfKey": "editais/tenant_abc/lic123.pdf",
  "priority": "high"
}
```

Exchange: `cotai.commands` (direct), routingKey: `ocr.extract`, queue: `ocr.process`.

## Evento de Saída (Kafka)
Tópico `ocr.completed`:
```json
{
  "eventId": "evt_ocr_done_123",
  "tenantId": "tenant_abc",
  "licitacaoId": "LIC123",
  "textS3Key": "ocr/tenant_abc/lic123.txt",
  "pages": 18,
  "engine": "textract",
  "durationMs": 9200,
  "timestamp": "2025-12-15T10:30:00Z"
}
```

## REST Endpoints (Monitoramento)
- `GET /health` — liveness/readiness (S3 + RabbitMQ + Kafka).
- `GET /metrics` — Prometheus.
- `GET /jobs/:id` — status do job OCR (cache Redis opcional).

Auth: JWT + `X-Tenant-ID`.

## Configuração de Engines
```python
if pages > 20 or scanned_pdf:
    use_textract()
else:
    use_tesseract(lang="por+eng", psm=6)
```

## Extração (Tesseract)
```python
import pytesseract
from pdf2image import convert_from_bytes

def ocr_local(pdf_bytes):
    images = convert_from_bytes(pdf_bytes, dpi=300)
    texts = [pytesseract.image_to_string(img, lang="por+eng", config="--oem 1 --psm 6") for img in images]
    return "\n\n".join(texts)
```

## Normalização de Texto
- Remover múltiplos espaços, quebras extras.
- Preservar tabelas simples com separador `;`.
- Detectar encoding e normalizar para UTF-8.

## Variáveis de Ambiente
```bash
RABBITMQ_URL=amqp://user:pass@rabbitmq:5672
KAFKA_BROKERS=kafka:9092
S3_BUCKET=cotai-ocr
AWS_REGION=sa-east-1
USE_TEXTRACT=true
TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata
```

## Dockerfile (resumo)
```dockerfile
FROM python:3.11-slim
RUN apt-get update && apt-get install -y tesseract-ocr libtesseract-dev poppler-utils
WORKDIR /app
COPY pyproject.toml poetry.lock ./
RUN pip install poetry && poetry install --no-dev
COPY src ./src
CMD ["poetry", "run", "python", "-m", "src.main"]
```

## Métricas
- `ocr_jobs_total{engine,status}`
- `ocr_duration_seconds{engine}`
- `ocr_pages_total`
- `ocr_textract_invocations_total`

## SLA e Retries
- Timeout por job: 2m (local), 5m (Textract async).
- Retries: 3 (exponential backoff).
- DLQ: `cotai.commands.dlq`.

## Segurança
- Tokens de serviço para Textract (IAM role).
- S3 com SSE-KMS; URLs com tempo limitado.
- Logs sem persistir PDFs; somente metadados.

*Referência: eventos em docs/architecture/communication-patterns.md*