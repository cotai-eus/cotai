# Arquitetura Multi-Tenant

## Visão Geral

A plataforma CotAI utiliza **Schema-per-Tenant** como estratégia de isolamento, combinando densidade de recursos com segurança forte via Row-Level Security (RLS).

---

## Decisão Arquitetural

| Critério | Database-per-tenant | Schema-per-tenant ✓ |
|----------|---------------------|---------------------|
| Isolamento | Máximo | Alto (+ RLS) |
| Custo operacional | Alto | Otimizado |
| Backup/Restore | Simples | Por schema |
| Escalabilidade | Limitada | Melhor densidade |
| Migração enterprise | N/A | Database dedicado |

---

## Estrutura de Schemas

```
PostgreSQL Instance
├── public (shared: migrations metadata, global configs)
├── tenant_abc123-def4-5678-ghij-klmnopqrstuv
│   ├── licitacoes
│   ├── fornecedores
│   ├── cotacoes
│   └── ...
├── tenant_xyz789-...
│   └── ...
└── _audit (logs cross-tenant, compliance)
```

---

## Tenant Resolution

### Via JWT Claim (Preferencial)

```json
{
  "sub": "user-uuid",
  "tenant_id": "abc123-def4-5678-ghij-klmnopqrstuv",
  "roles": ["admin", "commercial"],
  "iat": 1702656000,
  "exp": 1702742400
}
```

### Via Subdomain (Fallback)

```
https://acme.cotai.com.br → tenant: acme
https://globex.cotai.com.br → tenant: globex
```

### Middleware de Resolução (Node.js)

```typescript
// src/middleware/tenant.middleware.ts
import { Injectable, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';

@Injectable()
export class TenantMiddleware implements NestMiddleware {
  use(req: Request, res: Response, next: NextFunction) {
    // 1. Tenta JWT claim
    const tenantFromJwt = req.user?.tenant_id;
    
    // 2. Fallback: header explícito
    const tenantFromHeader = req.headers['x-tenant-id'] as string;
    
    // 3. Fallback: subdomain
    const subdomain = req.hostname.split('.')[0];
    const tenantFromSubdomain = subdomain !== 'api' ? subdomain : null;
    
    req.tenantId = tenantFromJwt || tenantFromHeader || tenantFromSubdomain;
    
    if (!req.tenantId) {
      return res.status(400).json({ error: 'Tenant não identificado' });
    }
    
    next();
  }
}
```

---

## Pool de Conexões

### Configuração por Tenant

```typescript
// src/database/tenant-connection.service.ts
import { Injectable } from '@nestjs/common';
import { Pool } from 'pg';

@Injectable()
export class TenantConnectionService {
  private pools: Map<string, Pool> = new Map();

  async getConnection(tenantId: string): Promise<Pool> {
    if (!this.pools.has(tenantId)) {
      const pool = new Pool({
        host: process.env.DB_HOST,
        database: process.env.DB_NAME,
        user: process.env.DB_USER,
        password: process.env.DB_PASSWORD,
        max: 10, // Conexões por tenant
        idleTimeoutMillis: 30000,
      });
      
      // Set search_path para schema do tenant
      pool.on('connect', async (client) => {
        await client.query(`SET search_path TO tenant_${tenantId}, public`);
      });
      
      this.pools.set(tenantId, pool);
    }
    
    return this.pools.get(tenantId)!;
  }
}
```

---

## Row-Level Security (RLS)

### Habilitação por Tabela

```sql
-- Habilitar RLS
ALTER TABLE licitacoes ENABLE ROW LEVEL SECURITY;

-- Policy de leitura
CREATE POLICY tenant_isolation_select ON licitacoes
  FOR SELECT
  USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Policy de inserção
CREATE POLICY tenant_isolation_insert ON licitacoes
  FOR INSERT
  WITH CHECK (tenant_id = current_setting('app.current_tenant')::uuid);

-- Policy de update
CREATE POLICY tenant_isolation_update ON licitacoes
  FOR UPDATE
  USING (tenant_id = current_setting('app.current_tenant')::uuid);

-- Policy de delete
CREATE POLICY tenant_isolation_delete ON licitacoes
  FOR DELETE
  USING (tenant_id = current_setting('app.current_tenant')::uuid);
```

### Configuração em Runtime

```sql
-- Antes de cada query, setar contexto
SET app.current_tenant = 'abc123-def4-5678-ghij-klmnopqrstuv';

-- Query agora filtrada automaticamente
SELECT * FROM licitacoes; -- Retorna apenas do tenant
```

---

## Provisioning de Novo Tenant

### Fluxo

```
POST /tenants → Tenant Manager → Create Schema → Apply Migrations → Seed Data
```

### Script de Criação

```sql
-- 1. Criar schema
CREATE SCHEMA IF NOT EXISTS tenant_${tenant_id};

-- 2. Aplicar migrations
SET search_path TO tenant_${tenant_id};

CREATE TABLE licitacoes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  numero VARCHAR(50) NOT NULL,
  objeto TEXT,
  status VARCHAR(20) DEFAULT 'RECEBIDO',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE fornecedores (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  razao_social VARCHAR(255) NOT NULL,
  cnpj VARCHAR(14) UNIQUE,
  email VARCHAR(255),
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Índices
CREATE INDEX idx_licitacoes_status ON licitacoes(status);
CREATE INDEX idx_licitacoes_tenant ON licitacoes(tenant_id);

-- 4. Habilitar RLS
ALTER TABLE licitacoes ENABLE ROW LEVEL SECURITY;
ALTER TABLE fornecedores ENABLE ROW LEVEL SECURITY;
```

---

## Migração para Database Dedicado (Enterprise)

### Critérios

- Volume > 100k licitações/ano
- Requisitos de compliance específicos
- SLA diferenciado (RPO/RTO < 5min)

### Processo

1. Criar database dedicado
2. `pg_dump` schema do tenant
3. `pg_restore` para novo database
4. Atualizar registro no Tenant Manager
5. Redirecionar conexões (blue-green)
6. Validar integridade
7. Remover schema antigo

---

## Métricas e Monitoramento

```yaml
# Prometheus metrics
tenant_active_connections{tenant_id="abc123"} 8
tenant_query_duration_seconds{tenant_id="abc123", operation="select"} 0.045
tenant_storage_bytes{tenant_id="abc123"} 1073741824
```

---

## Considerações de Segurança

| Aspecto | Implementação |
|---------|---------------|
| Isolamento de dados | Schema + RLS |
| Auditoria | Logs com tenant_id em todos eventos |
| Backup | Por schema, frequência configurável |
| Encryption | TDE + TLS 1.3 em trânsito |
| Access Control | JWT com tenant_id imutável |

---

*Referência: [README.md](../../README.md) - Seção 3.1*
