# Environment Setup

## Pré-requisitos
- Node.js 20, Go 1.21, Python 3.11
- Docker + Docker Compose
- kubectl + helm
- PostgreSQL 15, Redis, Kafka, RabbitMQ

## Passos Locais (dev)
```bash
# 1) Clonar
git clone <repo>
cd dev

# 2) Variáveis (exemplo)
cp .env.example .env

# 3) Subir deps básicas
docker compose up -d postgres redis kafka rabbitmq

# 4) Rodar services individuais (ex.)
cd docs/services/edge && npm ci && npm run dev
```

## Banco
- Multi-tenant: schemas `tenant_{uuid}`; usar migrações por serviço.
- Criar DBs: `auth`, `core`, `resources`, `collab` conforme serviços.

## Mensageria
- RabbitMQ: exchange `cotai.commands`, filas `crawler.jobs`, `ocr.process`.
- Kafka: tópicos `edital.raw`, `edital.normalized`, `licitacao.status.changed`, `audit.events`.

## Observabilidade
- Prometheus, Grafana, Jaeger: subir via helm charts opcionais.

## Perfis
- `dev`: logs verbosos, seeds dummy, CORS amplo.
- `prod`: CORS restrito, TLS obrigatório, secrets em Vault/Secrets Manager.

## Smoke Tests
```bash
curl -f http://localhost:3000/health || exit 1
```

*Use README e docs de cada serviço para comandos específicos.*