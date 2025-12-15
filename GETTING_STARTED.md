# CotAI Platform - Getting Started

Guia rápido para desenvolvedores começarem a trabalhar na plataforma CotAI.

## Visão Geral do Projeto

A plataforma CotAI é um sistema cloud-native de gestão de licitações baseado em microsserviços, organizado em 5 bounded contexts (DDD):

1. **Identity**: Autenticação, multi-tenancy, auditoria
2. **Acquisition**: Crawlers, ingestão de editais, normalização
3. **Core Bidding**: Motor de workflow Kanban, OCR/NLP, automação
4. **Resource Management**: CRM fornecedores, cotações, estoque
5. **Collaboration**: Chat real-time, notificações, agenda

## Pré-requisitos

### Ferramentas Obrigatórias

- **Docker**: >= 24.0
- **Docker Compose**: >= 2.20
- **kubectl**: >= 1.28.0
- **Helm**: >= 3.12.0
- **Node.js**: 20.x
- **Go**: 1.21.x
- **Python**: 3.11.x
- **Git**: >= 2.40

### Ferramentas Recomendadas

- **VS Code** com extensões:
  - Go
  - Python
  - ESLint
  - Prettier
  - Docker
  - Kubernetes
- **Postman** ou **Insomnia** para testes de API
- **k9s** para gerenciamento interativo de Kubernetes

## Setup do Ambiente Local

### 1. Clone do Repositório

```bash
git clone https://github.com/cotai/platform.git
cd platform
```

### 2. Setup de Dependências Locais (Docker Compose)

Para desenvolvimento local, use Docker Compose para rodar dependências:

```bash
# Iniciar PostgreSQL, Redis, Kafka, RabbitMQ localmente
docker compose -f docker-compose.dev.yml up -d

# Verificar que todos containers estão rodando
docker compose ps
```

**Serviços disponíveis**:
- PostgreSQL: `localhost:5432` (user: `cotai_dev`, password: `dev_password`)
- Redis: `localhost:6379`
- RabbitMQ: `localhost:5672` (management: `localhost:15672`)
- Kafka: `localhost:9092`

### 3. Setup de Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```bash
cp .env.example .env
```

Edite `.env` e configure:

```env
# Database
DATABASE_URL=postgresql://cotai_dev:dev_password@localhost:5432/cotai_dev

# Redis
REDIS_URL=redis://localhost:6379

# RabbitMQ
RABBITMQ_URL=amqp://cotai_dev:dev_password@localhost:5672/cotai

# Kafka
KAFKA_BROKERS=localhost:9092

# Auth (Keycloak - se rodando localmente)
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=cotai
KEYCLOAK_CLIENT_ID=cotai-dev

# Environment
NODE_ENV=development
GO_ENV=development
PYTHON_ENV=development
```

## Estrutura do Repositório

```
platform/
├── docs/                      # Documentação técnica
│   ├── architecture/          # Decisões arquiteturais
│   ├── services/              # Specs de cada serviço
│   ├── deployment/            # Guias de deployment
│   └── workflows/             # Fluxos de negócio
├── infra/                     # Infraestrutura como código
│   ├── terraform/             # Módulos Terraform
│   ├── helm/                  # Helm charts
│   └── kubernetes/            # Manifestos K8s
├── services/                  # Código dos microserviços
│   ├── identity/              # Bounded context Identity
│   ├── acquisition/           # Bounded context Acquisition
│   ├── core-bidding/          # Bounded context Core Bidding
│   ├── resources/             # Bounded context Resources
│   ├── collaboration/         # Bounded context Collaboration
│   └── edge/                  # API Gateway e Frontend
├── proto/                     # Definições gRPC (Protocol Buffers)
├── shared/                    # Código compartilhado
│   ├── types/                 # Types TypeScript/Go compartilhados
│   ├── utils/                 # Utilitários
│   └── middleware/            # Middleware comum
├── .github/                   # CI/CD workflows
│   └── workflows/
├── README.md                  # Overview do projeto
├── CLAUDE.md                  # Guia para Claude Code
├── IMPLEMENTATION_STATUS.md   # Status da implementação
└── GETTING_STARTED.md         # Este arquivo
```

## Desenvolvimento por Bounded Context

### Identity Context

```bash
cd services/identity/tenant-manager

# Install dependencies
go mod download

# Run service
go run cmd/server/main.go

# Run tests
go test ./...
```

### Acquisition Context

```bash
cd services/acquisition/crawler-workers

# Setup Python venv
python -m venv venv
source venv/bin/activate  # Linux/Mac
# ou: venv\Scripts\activate  # Windows

# Install dependencies
pip install -r requirements.txt

# Run crawler
python -m crawler.cli --source pncp --max-pages 5
```

### Core Bidding Context

```bash
cd services/core-bidding/kanban-api

# Install dependencies
npm ci

# Run migrations
npm run migrate:dev

# Start dev server
npm run dev

# Run tests
npm test
```

## Comandos Úteis

### Database

```bash
# Conectar ao PostgreSQL local
psql -h localhost -U cotai_dev -d cotai_dev

# Executar migrations (exemplo Node.js service)
cd services/core-bidding/kanban-api
npm run migrate:dev

# Seed database com dados de teste
npm run seed:dev
```

### Messaging

```bash
# RabbitMQ Management UI
open http://localhost:15672
# Login: cotai_dev / dev_password

# Publicar mensagem de teste (usando rabbitmqadmin)
rabbitmqadmin publish exchange=cotai.commands \
  routing_key=crawler.start \
  payload='{"source": "pncp", "filters": {}}'

# Consumir mensagens de uma fila
rabbitmqadmin get queue=crawler.jobs requeue=false

# Listar tópicos Kafka
docker exec -it <kafka-container> kafka-topics.sh --bootstrap-server localhost:9092 --list

# Consumir tópico Kafka
docker exec -it <kafka-container> kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic edital.raw \
  --from-beginning
```

### Testes

```bash
# Unit tests (Node.js)
npm test

# Unit tests (Go)
go test ./...

# Unit tests (Python)
pytest

# Integration tests
npm run test:integration

# E2E tests
npm run test:e2e
```

### Linting e Formatting

```bash
# Node.js
npm run lint
npm run format

# Go
go fmt ./...
golangci-lint run

# Python
black .
flake8
mypy .
```

## Workflow de Desenvolvimento

### 1. Criar Feature Branch

```bash
git checkout -b feature/COTAI-123-add-supplier-matching
```

Convenção de nomes:
- `feature/COTAI-XXX-description`: Nova funcionalidade
- `fix/COTAI-XXX-description`: Bug fix
- `refactor/COTAI-XXX-description`: Refatoração
- `docs/COTAI-XXX-description`: Documentação

### 2. Desenvolver e Testar Localmente

```bash
# Fazer alterações no código

# Rodar testes
npm test

# Verificar lint
npm run lint

# Testar localmente
npm run dev
```

### 3. Commit com Conventional Commits

```bash
git add .
git commit -m "feat(crm): add supplier matching algorithm

- Implement cosine similarity for product matching
- Add distance calculation for supplier geo-filtering
- COTAI-123"
```

Prefixos de commit:
- `feat`: Nova feature
- `fix`: Bug fix
- `refactor`: Refatoração
- `docs`: Documentação
- `test`: Testes
- `chore`: Tarefas de manutenção

### 4. Push e Pull Request

```bash
git push origin feature/COTAI-123-add-supplier-matching
```

Criar PR no GitHub com:
- Título descritivo
- Descrição detalhada do que foi alterado
- Referência ao ticket (COTAI-XXX)
- Screenshots se aplicável
- Checklist do template de PR preenchido

## Debugging

### Node.js Services

```bash
# Start com inspector
node --inspect-brk dist/main.js

# VS Code launch.json
{
  "type": "node",
  "request": "attach",
  "name": "Attach to Node",
  "port": 9229
}
```

### Go Services

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start com Delve
dlv debug cmd/server/main.go

# VS Code launch.json
{
  "type": "go",
  "request": "launch",
  "name": "Launch Package",
  "program": "${workspaceFolder}/cmd/server"
}
```

### Python Services

```bash
# pdb
python -m pdb crawler/main.py

# VS Code launch.json
{
  "type": "python",
  "request": "launch",
  "name": "Python: Current File",
  "program": "${file}",
  "console": "integratedTerminal"
}
```

## Troubleshooting Comum

### Docker Compose não inicia

```bash
# Limpar volumes e reconstruir
docker compose down -v
docker compose up -d --build
```

### PostgreSQL connection refused

```bash
# Verificar se PostgreSQL está rodando
docker compose ps postgres

# Ver logs
docker compose logs postgres

# Reiniciar serviço
docker compose restart postgres
```

### Kafka topic não existe

```bash
# Criar tópico manualmente
docker exec -it <kafka-container> kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --topic edital.raw \
  --partitions 12 \
  --replication-factor 1
```

### Port already in use

```bash
# Encontrar processo usando porta 3000
lsof -i :3000  # Mac/Linux
netstat -ano | findstr :3000  # Windows

# Matar processo
kill -9 <PID>
```

## Recursos Adicionais

### Documentação

- [Arquitetura Geral](README.md)
- [Multi-Tenancy](docs/architecture/multi-tenancy.md)
- [Comunicação](docs/architecture/communication-patterns.md)
- [Segurança](docs/architecture/security-compliance.md)
- [Jornada do Edital](docs/workflows/edital-journey.md)

### Guias de Serviços

Cada serviço tem documentação específica em `docs/services/{context}/{service}.md`:

- [Auth Service](docs/services/identity/auth-service.md)
- [Tenant Manager](docs/services/identity/tenant-manager.md)
- [Kanban API](docs/services/core-bidding/kanban-api.md)
- [OCR/NLP Service](docs/services/core-bidding/ocr-nlp-service.md)
- E mais...

### Suporte

- **Issues**: https://github.com/cotai/platform/issues
- **Discussions**: https://github.com/cotai/platform/discussions
- **Slack**: #cotai-dev (interno)

## Próximos Passos

1. Escolha um bounded context para trabalhar
2. Leia a documentação do serviço específico
3. Configure seu ambiente local
4. Pegue uma task no backlog
5. Desenvolva, teste e crie um PR!

Bem-vindo ao time CotAI! 🚀
