# Crawler Workers

## Visão Geral

| Atributo | Valor |
|----------|-------|
| **Bounded Context** | Acquisition & Ingestion |
| **Responsabilidade** | Coleta de editais de portais públicos (PNCP, etc.) |
| **Stack** | Python 3.11+ + Scrapy |
| **Database** | - (stateless) |
| **Storage** | S3 (PDFs, anexos) |
| **Deploy** | Kubernetes Jobs / AWS Lambda |

---

## Responsabilidades

1. **Consumir jobs** da fila RabbitMQ
2. **Executar crawlers** por fonte (PNCP, portais estaduais)
3. **Extrair metadados** de editais
4. **Baixar documentos** (PDFs, anexos) para S3
5. **Publicar resultados** para Normalizer via Kafka
6. **Respeitar rate limits** e robots.txt

---

## Arquitetura

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   RabbitMQ   │────►│   Crawler    │────►│    Kafka     │
│ crawler.jobs │     │   Workers    │     │edital.raw    │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────▼──────┐
                     │     S3      │
                     │  (docs)     │
                     └─────────────┘
```

---

## Estrutura do Projeto

```
crawler-workers/
├── pyproject.toml
├── Dockerfile
├── src/
│   ├── __init__.py
│   ├── main.py
│   ├── consumer.py
│   ├── spiders/
│   │   ├── __init__.py
│   │   ├── base.py
│   │   ├── pncp.py
│   │   ├── comprasnet.py
│   │   └── bec_sp.py
│   ├── pipelines/
│   │   ├── __init__.py
│   │   ├── s3_upload.py
│   │   └── kafka_publisher.py
│   ├── items.py
│   └── settings.py
└── tests/
```

---

## Spider PNCP

```python
# src/spiders/pncp.py
import scrapy
from datetime import datetime, timedelta
from ..items import EditalItem

class PNCPSpider(scrapy.Spider):
    name = 'pncp'
    allowed_domains = ['pncp.gov.br']
    
    custom_settings = {
        'CONCURRENT_REQUESTS': 4,
        'DOWNLOAD_DELAY': 1,
        'AUTOTHROTTLE_ENABLED': True,
    }
    
    def __init__(self, job_data: dict, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.job_id = job_data['jobId']
        self.tenant_id = job_data['tenantId']
        self.filters = job_data.get('filters', {})
        self.config = job_data.get('config', {})
        
    def start_requests(self):
        base_url = 'https://pncp.gov.br/api/consulta/v1/contratacoes/publicacao'
        
        params = {
            'dataInicial': self.filters.get('dataInicio', 
                (datetime.now() - timedelta(days=1)).strftime('%Y%m%d')),
            'dataFinal': self.filters.get('dataFim',
                datetime.now().strftime('%Y%m%d')),
            'pagina': 1,
            'tamanhoPagina': 50,
        }
        
        if self.filters.get('ufs'):
            params['uf'] = ','.join(self.filters['ufs'])
            
        if self.filters.get('modalidades'):
            params['modalidade'] = ','.join(self.filters['modalidades'])
        
        yield scrapy.Request(
            url=f"{base_url}?{self._build_query(params)}",
            callback=self.parse_list,
            meta={'page': 1}
        )
    
    def parse_list(self, response):
        data = response.json()
        
        for item in data.get('data', []):
            yield scrapy.Request(
                url=f"https://pncp.gov.br/api/consulta/v1/contratacoes/{item['id']}",
                callback=self.parse_detail,
                meta={'basic_info': item}
            )
        
        # Paginação
        page = response.meta['page']
        max_pages = self.config.get('maxPages', 10)
        
        if data.get('hasNext') and page < max_pages:
            next_url = response.url.replace(f'pagina={page}', f'pagina={page+1}')
            yield scrapy.Request(
                url=next_url,
                callback=self.parse_list,
                meta={'page': page + 1}
            )
    
    def parse_detail(self, response):
        data = response.json()
        basic = response.meta['basic_info']
        
        item = EditalItem(
            source='pncp',
            source_id=data['id'],
            tenant_id=self.tenant_id,
            job_id=self.job_id,
            numero=data.get('numeroControlePNCP'),
            objeto=data.get('objetoCompra'),
            modalidade=data.get('modalidadeNome'),
            orgao=data.get('orgaoEntidade', {}).get('razaoSocial'),
            orgao_cnpj=data.get('orgaoEntidade', {}).get('cnpj'),
            uf=data.get('unidadeOrgao', {}).get('ufSigla'),
            municipio=data.get('unidadeOrgao', {}).get('municipioNome'),
            valor_estimado=data.get('valorTotalEstimado'),
            data_publicacao=data.get('dataPublicacaoPncp'),
            data_abertura=data.get('dataAberturaProposta'),
            data_encerramento=data.get('dataEncerramentoProposta'),
            url_original=f"https://pncp.gov.br/app/editais/{data['id']}",
            documentos=[],
            raw_data=data,
        )
        
        # Baixar documentos
        for doc in data.get('arquivos', []):
            item['documentos'].append({
                'tipo': doc.get('tipoDocumentoNome'),
                'nome': doc.get('nomeArquivo'),
                'url': doc.get('url'),
            })
        
        yield item
    
    def _build_query(self, params: dict) -> str:
        return '&'.join(f"{k}={v}" for k, v in params.items() if v)
```

---

## Consumer RabbitMQ

```python
# src/consumer.py
import pika
import json
from scrapy.crawler import CrawlerProcess
from scrapy.utils.project import get_project_settings

class CrawlerConsumer:
    def __init__(self):
        self.connection = pika.BlockingConnection(
            pika.ConnectionParameters(
                host=os.environ['RABBITMQ_HOST'],
                credentials=pika.PlainCredentials(
                    os.environ['RABBITMQ_USER'],
                    os.environ['RABBITMQ_PASSWORD']
                )
            )
        )
        self.channel = self.connection.channel()
        self.channel.queue_declare(queue='crawler.jobs', durable=True)
        self.channel.basic_qos(prefetch_count=1)
        
    def start(self):
        self.channel.basic_consume(
            queue='crawler.jobs',
            on_message_callback=self.process_job
        )
        print('Waiting for crawler jobs...')
        self.channel.start_consuming()
    
    def process_job(self, ch, method, properties, body):
        job_data = json.loads(body)
        source = job_data.get('source', 'pncp')
        
        try:
            self.run_spider(source, job_data)
            ch.basic_ack(delivery_tag=method.delivery_tag)
        except Exception as e:
            # Retry ou DLQ
            if method.redelivered:
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)
            else:
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)
    
    def run_spider(self, source: str, job_data: dict):
        settings = get_project_settings()
        process = CrawlerProcess(settings)
        
        spider_class = self.get_spider_class(source)
        process.crawl(spider_class, job_data=job_data)
        process.start()
    
    def get_spider_class(self, source: str):
        from .spiders import pncp, comprasnet, bec_sp
        return {
            'pncp': pncp.PNCPSpider,
            'comprasnet': comprasnet.ComprasNetSpider,
            'bec': bec_sp.BECSPSpider,
        }.get(source)
```

---

## Item Definition

```python
# src/items.py
import scrapy

class EditalItem(scrapy.Item):
    source = scrapy.Field()
    source_id = scrapy.Field()
    tenant_id = scrapy.Field()
    job_id = scrapy.Field()
    
    numero = scrapy.Field()
    objeto = scrapy.Field()
    modalidade = scrapy.Field()
    
    orgao = scrapy.Field()
    orgao_cnpj = scrapy.Field()
    uf = scrapy.Field()
    municipio = scrapy.Field()
    
    valor_estimado = scrapy.Field()
    data_publicacao = scrapy.Field()
    data_abertura = scrapy.Field()
    data_encerramento = scrapy.Field()
    
    url_original = scrapy.Field()
    documentos = scrapy.Field()  # List[{tipo, nome, url, s3_key}]
    
    raw_data = scrapy.Field()
    crawled_at = scrapy.Field()
```

---

## Pipeline S3

```python
# src/pipelines/s3_upload.py
import boto3
import hashlib
from scrapy.http import Request

class S3UploadPipeline:
    def __init__(self):
        self.s3 = boto3.client('s3')
        self.bucket = os.environ['S3_BUCKET']
    
    def process_item(self, item, spider):
        for doc in item.get('documentos', []):
            if doc.get('url'):
                s3_key = self.upload_document(
                    url=doc['url'],
                    tenant_id=item['tenant_id'],
                    source_id=item['source_id'],
                    filename=doc['nome']
                )
                doc['s3_key'] = s3_key
        
        return item
    
    def upload_document(self, url: str, tenant_id: str, source_id: str, filename: str) -> str:
        import requests
        
        response = requests.get(url, timeout=30)
        content = response.content
        
        # Hash para evitar duplicatas
        content_hash = hashlib.sha256(content).hexdigest()[:12]
        s3_key = f"editais/{tenant_id}/{source_id}/{content_hash}_{filename}"
        
        self.s3.put_object(
            Bucket=self.bucket,
            Key=s3_key,
            Body=content,
            ContentType=response.headers.get('Content-Type', 'application/pdf'),
            Metadata={
                'tenant_id': tenant_id,
                'source_id': source_id,
                'original_url': url,
            }
        )
        
        return s3_key
```

---

## Pipeline Kafka

```python
# src/pipelines/kafka_publisher.py
from kafka import KafkaProducer
import json
from datetime import datetime

class KafkaPublisherPipeline:
    def __init__(self):
        self.producer = KafkaProducer(
            bootstrap_servers=os.environ['KAFKA_BROKERS'].split(','),
            value_serializer=lambda v: json.dumps(v, default=str).encode('utf-8'),
            key_serializer=lambda k: k.encode('utf-8') if k else None,
        )
        self.topic = 'edital.raw'
    
    def process_item(self, item, spider):
        event = {
            'eventId': f"evt_{item['source']}_{item['source_id']}",
            'eventType': 'edital.crawled',
            'timestamp': datetime.utcnow().isoformat(),
            'tenantId': item['tenant_id'],
            'payload': dict(item),
        }
        
        self.producer.send(
            self.topic,
            key=item['source_id'],
            value=event,
        )
        
        return item
    
    def close_spider(self, spider):
        self.producer.flush()
        self.producer.close()
```

---

## Configurações Scrapy

```python
# src/settings.py
BOT_NAME = 'cotai-crawler'
SPIDER_MODULES = ['src.spiders']

# Concurrent requests
CONCURRENT_REQUESTS = 8
CONCURRENT_REQUESTS_PER_DOMAIN = 4
DOWNLOAD_DELAY = 0.5

# AutoThrottle
AUTOTHROTTLE_ENABLED = True
AUTOTHROTTLE_START_DELAY = 1
AUTOTHROTTLE_MAX_DELAY = 10
AUTOTHROTTLE_TARGET_CONCURRENCY = 4

# Retry
RETRY_ENABLED = True
RETRY_TIMES = 3
RETRY_HTTP_CODES = [500, 502, 503, 504, 408, 429]

# Pipelines
ITEM_PIPELINES = {
    'src.pipelines.s3_upload.S3UploadPipeline': 100,
    'src.pipelines.kafka_publisher.KafkaPublisherPipeline': 200,
}

# Logging
LOG_LEVEL = 'INFO'
LOG_FORMAT = '%(asctime)s [%(name)s] %(levelname)s: %(message)s'

# User-Agent rotation
USER_AGENT = 'CotAI-Crawler/1.0 (+https://cotai.com.br/crawler)'
```

---

## Variáveis de Ambiente

```bash
# RabbitMQ
RABBITMQ_HOST=rabbitmq
RABBITMQ_USER=crawler
RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD}

# Kafka
KAFKA_BROKERS=kafka:9092

# S3
AWS_REGION=sa-east-1
S3_BUCKET=cotai-editais-raw

# Crawler
CONCURRENT_REQUESTS=8
DOWNLOAD_DELAY=0.5
```

---

## Dockerfile

```dockerfile
FROM python:3.11-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    libxml2-dev libxslt-dev \
    && rm -rf /var/lib/apt/lists/*

COPY pyproject.toml poetry.lock ./
RUN pip install poetry && poetry install --no-dev

COPY src ./src

CMD ["poetry", "run", "python", "-m", "src.main"]
```

---

## Kubernetes Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: crawler-pncp-{{ .Release.Time }}
spec:
  template:
    spec:
      containers:
      - name: crawler
        image: cotai/crawler-workers:latest
        env:
        - name: RABBITMQ_HOST
          value: rabbitmq
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      restartPolicy: OnFailure
  backoffLimit: 3
```

---

## Métricas

```python
# Exportar métricas Prometheus
from prometheus_client import Counter, Histogram

editais_crawled = Counter(
    'crawler_editais_total',
    'Total editais crawled',
    ['source', 'status']
)

crawl_duration = Histogram(
    'crawler_duration_seconds',
    'Time spent crawling',
    ['source']
)
```

---

*Referência: [Scheduler](./scheduler.md) | [Normalizer](./normalizer.md)*
