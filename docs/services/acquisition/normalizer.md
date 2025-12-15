# Normalizer Service

## Visão Geral

| Atributo | Valor |
|----------|-------|
| **Bounded Context** | Acquisition & Ingestion |
| **Responsabilidade** | Padronização de dados, deduplicação, enriquecimento |
| **Stack** | Python 3.11+ + FastAPI |
| **Database** | PostgreSQL (schema por tenant) |
| **Porta** | 8020 |
| **Protocolo** | Kafka Consumer + REST |

---

## Responsabilidades

1. **Consumir eventos** `edital.raw` do Kafka
2. **Padronizar dados** (formatos, campos, códigos)
3. **Deduplicar editais** (fingerprint + hash)
4. **Enriquecer metadados** (geolocalização, classificação)
5. **Persistir no banco** do tenant
6. **Publicar evento** `edital.normalized` para Workflow Engine

---

## Arquitetura

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Kafka     │────►│  Normalizer  │────►│    Kafka     │
│ edital.raw   │     │   Service    │     │edital.normal │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────▼──────┐
                     │  PostgreSQL │
                     │  (tenant_x) │
                     └─────────────┘
```

---

## Estrutura do Projeto

```
normalizer/
├── pyproject.toml
├── Dockerfile
├── src/
│   ├── __init__.py
│   ├── main.py
│   ├── api/
│   │   ├── __init__.py
│   │   └── routes.py
│   ├── consumer/
│   │   ├── __init__.py
│   │   └── kafka_consumer.py
│   ├── services/
│   │   ├── __init__.py
│   │   ├── normalizer.py
│   │   ├── deduplicator.py
│   │   └── enricher.py
│   ├── models/
│   │   ├── __init__.py
│   │   └── edital.py
│   └── db/
│       ├── __init__.py
│       └── repository.py
└── tests/
```

---

## API Endpoints

### Status e Métricas

```http
GET /health
```

```json
{
  "status": "healthy",
  "kafka": "connected",
  "database": "connected",
  "processed_last_hour": 234
}
```

### Normalizar Manualmente

```http
POST /api/v1/normalize
Authorization: Bearer <token>
X-Tenant-ID: <tenant_id>
Content-Type: application/json

{
  "source": "pncp",
  "sourceId": "12345",
  "numero": "PE-001/2025",
  "objeto": "Aquisição de equipamentos de TI",
  "modalidade": "pregao_eletronico",
  "orgao": "Prefeitura Municipal de São Paulo",
  "valor_estimado": 150000.00,
  "data_abertura": "2025-12-20T14:00:00Z"
}
```

**Response:**
```json
{
  "id": "ed_abc123",
  "status": "normalized",
  "fingerprint": "fp_xyz789",
  "isDuplicate": false,
  "enrichments": {
    "ibge_code": "3550308",
    "regiao": "Sudeste",
    "porte_orgao": "grande"
  }
}
```

### Verificar Duplicata

```http
POST /api/v1/check-duplicate
Authorization: Bearer <token>
X-Tenant-ID: <tenant_id>
Content-Type: application/json

{
  "numero": "PE-001/2025",
  "orgao_cnpj": "46395000000139",
  "objeto": "Aquisição de equipamentos de TI"
}
```

**Response:**
```json
{
  "isDuplicate": true,
  "existingId": "ed_def456",
  "similarity": 0.95,
  "matchedFields": ["numero", "orgao_cnpj"]
}
```

---

## Consumer Kafka

```python
# src/consumer/kafka_consumer.py
from aiokafka import AIOKafkaConsumer
import asyncio
import json

class EditalConsumer:
    def __init__(self, normalizer_service, publisher):
        self.normalizer = normalizer_service
        self.publisher = publisher
        
    async def start(self):
        consumer = AIOKafkaConsumer(
            'edital.raw',
            bootstrap_servers=os.environ['KAFKA_BROKERS'],
            group_id='normalizer-service',
            auto_offset_reset='earliest',
            enable_auto_commit=False,
        )
        
        await consumer.start()
        try:
            async for msg in consumer:
                await self.process_message(msg)
                await consumer.commit()
        finally:
            await consumer.stop()
    
    async def process_message(self, msg):
        event = json.loads(msg.value.decode('utf-8'))
        tenant_id = event.get('tenantId')
        payload = event.get('payload', {})
        
        try:
            # 1. Normalizar
            normalized = await self.normalizer.normalize(payload)
            
            # 2. Verificar duplicata
            duplicate = await self.normalizer.check_duplicate(
                tenant_id, normalized
            )
            
            if duplicate:
                await self.handle_duplicate(tenant_id, normalized, duplicate)
                return
            
            # 3. Enriquecer
            enriched = await self.normalizer.enrich(normalized)
            
            # 4. Persistir
            edital_id = await self.normalizer.save(tenant_id, enriched)
            
            # 5. Publicar evento
            await self.publisher.publish_normalized(tenant_id, edital_id, enriched)
            
        except Exception as e:
            await self.handle_error(event, e)
```

---

## Serviço de Normalização

```python
# src/services/normalizer.py
from datetime import datetime
import re

class NormalizerService:
    def __init__(self, db_repository, enricher):
        self.db = db_repository
        self.enricher = enricher
    
    async def normalize(self, raw_data: dict) -> dict:
        """Padroniza campos e formatos."""
        return {
            'source': raw_data.get('source'),
            'source_id': raw_data.get('source_id'),
            'numero': self._normalize_numero(raw_data.get('numero')),
            'objeto': self._normalize_texto(raw_data.get('objeto')),
            'objeto_resumido': self._extract_resumo(raw_data.get('objeto')),
            'modalidade': self._normalize_modalidade(raw_data.get('modalidade')),
            'orgao': self._normalize_texto(raw_data.get('orgao')),
            'orgao_cnpj': self._normalize_cnpj(raw_data.get('orgao_cnpj')),
            'uf': raw_data.get('uf', '').upper(),
            'municipio': self._normalize_texto(raw_data.get('municipio')),
            'valor_estimado': self._normalize_valor(raw_data.get('valor_estimado')),
            'data_publicacao': self._parse_date(raw_data.get('data_publicacao')),
            'data_abertura': self._parse_date(raw_data.get('data_abertura')),
            'data_encerramento': self._parse_date(raw_data.get('data_encerramento')),
            'url_original': raw_data.get('url_original'),
            'documentos': raw_data.get('documentos', []),
            'raw_data': raw_data.get('raw_data'),
            'normalized_at': datetime.utcnow(),
        }
    
    def _normalize_numero(self, numero: str) -> str:
        if not numero:
            return None
        # Remove espaços extras, padroniza formato
        return re.sub(r'\s+', ' ', numero.strip().upper())
    
    def _normalize_texto(self, texto: str) -> str:
        if not texto:
            return None
        # Remove espaços extras, normaliza unicode
        import unicodedata
        texto = unicodedata.normalize('NFKC', texto)
        return re.sub(r'\s+', ' ', texto.strip())
    
    def _normalize_modalidade(self, modalidade: str) -> str:
        MODALIDADE_MAP = {
            'pregão eletrônico': 'PREGAO_ELETRONICO',
            'pregao eletronico': 'PREGAO_ELETRONICO',
            'concorrência': 'CONCORRENCIA',
            'tomada de preços': 'TOMADA_PRECOS',
            'convite': 'CONVITE',
            'dispensa': 'DISPENSA',
            'inexigibilidade': 'INEXIGIBILIDADE',
        }
        if not modalidade:
            return 'OUTROS'
        key = modalidade.lower().strip()
        return MODALIDADE_MAP.get(key, 'OUTROS')
    
    def _normalize_cnpj(self, cnpj: str) -> str:
        if not cnpj:
            return None
        return re.sub(r'\D', '', cnpj)[:14]
    
    def _normalize_valor(self, valor) -> float:
        if valor is None:
            return None
        if isinstance(valor, (int, float)):
            return float(valor)
        # Parse string com vírgula/ponto
        valor_str = str(valor).replace('.', '').replace(',', '.')
        try:
            return float(valor_str)
        except ValueError:
            return None
    
    def _parse_date(self, date_str: str) -> datetime:
        if not date_str:
            return None
        formats = [
            '%Y-%m-%dT%H:%M:%S.%fZ',
            '%Y-%m-%dT%H:%M:%SZ',
            '%Y-%m-%dT%H:%M:%S',
            '%Y-%m-%d',
            '%d/%m/%Y',
        ]
        for fmt in formats:
            try:
                return datetime.strptime(date_str, fmt)
            except ValueError:
                continue
        return None
    
    def _extract_resumo(self, objeto: str, max_len: int = 200) -> str:
        if not objeto:
            return None
        texto = self._normalize_texto(objeto)
        if len(texto) <= max_len:
            return texto
        return texto[:max_len-3] + '...'
```

---

## Deduplicador

```python
# src/services/deduplicator.py
import hashlib
from typing import Optional

class Deduplicator:
    def __init__(self, db_repository):
        self.db = db_repository
    
    async def check_duplicate(
        self, 
        tenant_id: str, 
        edital: dict
    ) -> Optional[dict]:
        """Verifica se edital já existe."""
        
        # 1. Match exato por source_id
        existing = await self.db.find_by_source_id(
            tenant_id, 
            edital['source'], 
            edital['source_id']
        )
        if existing:
            return {'type': 'exact', 'existing': existing}
        
        # 2. Match por fingerprint
        fingerprint = self.generate_fingerprint(edital)
        existing = await self.db.find_by_fingerprint(tenant_id, fingerprint)
        if existing:
            return {'type': 'fingerprint', 'existing': existing}
        
        # 3. Match por similaridade (número + órgão)
        if edital.get('numero') and edital.get('orgao_cnpj'):
            existing = await self.db.find_similar(
                tenant_id,
                numero=edital['numero'],
                orgao_cnpj=edital['orgao_cnpj']
            )
            if existing:
                similarity = self.calculate_similarity(edital, existing)
                if similarity > 0.9:
                    return {'type': 'similar', 'existing': existing, 'similarity': similarity}
        
        return None
    
    def generate_fingerprint(self, edital: dict) -> str:
        """Gera fingerprint único baseado em campos chave."""
        components = [
            edital.get('numero', ''),
            edital.get('orgao_cnpj', ''),
            edital.get('modalidade', ''),
            str(edital.get('data_abertura', '')),
        ]
        content = '|'.join(components).lower()
        return hashlib.sha256(content.encode()).hexdigest()[:32]
    
    def calculate_similarity(self, a: dict, b: dict) -> float:
        """Calcula similaridade entre dois editais."""
        from difflib import SequenceMatcher
        
        obj_a = a.get('objeto', '') or ''
        obj_b = b.get('objeto', '') or ''
        
        return SequenceMatcher(None, obj_a.lower(), obj_b.lower()).ratio()
```

---

## Modelo de Dados

```sql
-- Schema: tenant_{uuid}

CREATE TABLE editais_raw (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source VARCHAR(50) NOT NULL,
    source_id VARCHAR(100) NOT NULL,
    fingerprint VARCHAR(64) NOT NULL,
    
    numero VARCHAR(100),
    objeto TEXT,
    objeto_resumido VARCHAR(250),
    modalidade VARCHAR(50),
    
    orgao VARCHAR(255),
    orgao_cnpj VARCHAR(14),
    uf CHAR(2),
    municipio VARCHAR(100),
    
    valor_estimado DECIMAL(15,2),
    data_publicacao TIMESTAMPTZ,
    data_abertura TIMESTAMPTZ,
    data_encerramento TIMESTAMPTZ,
    
    url_original TEXT,
    documentos JSONB DEFAULT '[]',
    enrichments JSONB DEFAULT '{}',
    raw_data JSONB,
    
    status VARCHAR(20) DEFAULT 'normalized',
    normalized_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(source, source_id)
);

CREATE INDEX idx_editais_fingerprint ON editais_raw(fingerprint);
CREATE INDEX idx_editais_numero_cnpj ON editais_raw(numero, orgao_cnpj);
CREATE INDEX idx_editais_abertura ON editais_raw(data_abertura);
CREATE INDEX idx_editais_modalidade ON editais_raw(modalidade);
CREATE INDEX idx_editais_status ON editais_raw(status);
```

---

## Publicador Kafka

```python
# src/services/publisher.py
from aiokafka import AIOKafkaProducer
import json
from datetime import datetime

class EventPublisher:
    def __init__(self):
        self.producer = None
        
    async def start(self):
        self.producer = AIOKafkaProducer(
            bootstrap_servers=os.environ['KAFKA_BROKERS'],
            value_serializer=lambda v: json.dumps(v, default=str).encode()
        )
        await self.producer.start()
    
    async def publish_normalized(
        self, 
        tenant_id: str, 
        edital_id: str, 
        edital: dict
    ):
        event = {
            'eventId': f"evt_norm_{edital_id}",
            'eventType': 'edital.normalized',
            'timestamp': datetime.utcnow().isoformat(),
            'tenantId': tenant_id,
            'aggregateId': edital_id,
            'payload': {
                'editalId': edital_id,
                'numero': edital['numero'],
                'modalidade': edital['modalidade'],
                'dataAbertura': edital['data_abertura'],
                'documentos': len(edital.get('documentos', [])),
            }
        }
        
        await self.producer.send(
            'edital.normalized',
            key=edital_id.encode(),
            value=event
        )
```

---

## Variáveis de Ambiente

```bash
# Server
PORT=8020

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_GROUP_ID=normalizer-service

# Database
DATABASE_URL=postgres://user:pass@postgres:5432/cotai

# Enrichment APIs
IBGE_API_URL=https://servicodados.ibge.gov.br/api/v1
```

---

## Dockerfile

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY pyproject.toml poetry.lock ./
RUN pip install poetry && poetry install --no-dev

COPY src ./src

EXPOSE 8020
CMD ["poetry", "run", "uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8020"]
```

---

## Métricas

```yaml
normalizer_editais_processed_total{status="success"} 15234
normalizer_editais_processed_total{status="duplicate"} 1823
normalizer_editais_processed_total{status="error"} 45
normalizer_processing_duration_seconds{quantile="0.99"} 0.85
```

---

*Referência: [Crawler Workers](./crawler-workers.md) | [Workflow Engine](../core-bidding/workflow-engine.md)*
