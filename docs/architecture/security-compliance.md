# Segurança e Compliance

## Visão Geral

A plataforma CotAI implementa segurança em múltiplas camadas, desde a borda (WAF) até o dado em repouso, com foco em compliance LGPD.

---

## Camadas de Segurança

| Camada | Controles |
|--------|-----------|
| **Rede** | VPC isolada, Security Groups, WAF, DDoS protection |
| **Edge** | TLS 1.3, Rate Limiting, CORS, CSP headers |
| **Autenticação** | OAuth 2.0 + OIDC, JWT, MFA |
| **Autorização** | RBAC, Attribute-based Access Control |
| **Dados** | Encryption at rest (AES-256), TLS em trânsito |
| **Tenant** | Schema isolation, RLS, Audit trail |
| **Aplicação** | Input validation, SQL injection prevention, XSS protection |

---

## OAuth 2.0 / OpenID Connect

### Fluxo Authorization Code + PKCE

```
┌──────────┐                              ┌──────────────┐
│  Client  │                              │ Auth Service │
│ (React)  │                              │  (Keycloak)  │
└────┬─────┘                              └──────┬───────┘
     │                                           │
     │ 1. GET /authorize?                        │
     │    response_type=code&                    │
     │    client_id=xxx&                         │
     │    redirect_uri=xxx&                      │
     │    code_challenge=xxx&                    │
     │    code_challenge_method=S256             │
     │ ─────────────────────────────────────────►│
     │                                           │
     │ 2. Login page                             │
     │ ◄─────────────────────────────────────────│
     │                                           │
     │ 3. User authenticates                     │
     │ ─────────────────────────────────────────►│
     │                                           │
     │ 4. Redirect with code                     │
     │ ◄─────────────────────────────────────────│
     │                                           │
     │ 5. POST /token                            │
     │    grant_type=authorization_code&         │
     │    code=xxx&                              │
     │    code_verifier=xxx                      │
     │ ─────────────────────────────────────────►│
     │                                           │
     │ 6. { access_token, refresh_token, id_token }
     │ ◄─────────────────────────────────────────│
```

### Estrutura do JWT

```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-id-123"
  },
  "payload": {
    "iss": "https://auth.cotai.com.br/realms/cotai",
    "sub": "user-uuid-123",
    "aud": "cotai-web",
    "exp": 1702742400,
    "iat": 1702656000,
    "tenant_id": "tenant-uuid-456",
    "roles": ["admin", "commercial"],
    "permissions": [
      "licitacao:read",
      "licitacao:write",
      "fornecedor:read"
    ],
    "email": "user@empresa.com.br",
    "name": "João Silva"
  }
}
```

---

## RBAC (Role-Based Access Control)

### Hierarquia de Roles

```yaml
roles:
  super_admin:
    description: "Administrador da plataforma (CotAI)"
    permissions: ["*"]
    
  tenant_admin:
    description: "Administrador do tenant"
    permissions:
      - "users:*"
      - "settings:*"
      - "licitacao:*"
      - "fornecedor:*"
      - "cotacao:*"
      
  manager:
    description: "Gestor comercial"
    permissions:
      - "licitacao:read"
      - "licitacao:write"
      - "cotacao:*"
      - "fornecedor:read"
      - "relatorio:read"
      
  commercial:
    description: "Equipe comercial"
    permissions:
      - "licitacao:read"
      - "cotacao:read"
      - "cotacao:write"
      - "fornecedor:read"
      
  supplier:
    description: "Portal do fornecedor"
    permissions:
      - "cotacao:respond"
      - "catalog:read"
      - "profile:write"
```

### Middleware de Autorização

```typescript
// src/guards/permission.guard.ts
import { CanActivate, ExecutionContext, Injectable } from '@nestjs/common';
import { Reflector } from '@nestjs/core';

@Injectable()
export class PermissionGuard implements CanActivate {
  constructor(private reflector: Reflector) {}

  canActivate(context: ExecutionContext): boolean {
    const requiredPermissions = this.reflector.get<string[]>(
      'permissions',
      context.getHandler()
    );
    
    if (!requiredPermissions) return true;
    
    const request = context.switchToHttp().getRequest();
    const userPermissions = request.user?.permissions || [];
    
    return requiredPermissions.every(perm => 
      this.hasPermission(userPermissions, perm)
    );
  }
  
  private hasPermission(userPerms: string[], required: string): boolean {
    return userPerms.some(p => 
      p === '*' || 
      p === required ||
      (p.endsWith(':*') && required.startsWith(p.replace(':*', ':')))
    );
  }
}

// Uso em controller
@Get('licitacoes')
@Permissions('licitacao:read')
async listLicitacoes() { ... }
```

---

## Rate Limiting

### Configuração Kong

```yaml
# kong.yml
plugins:
  - name: rate-limiting
    config:
      minute: 100
      hour: 1000
      policy: redis
      redis_host: redis
      redis_port: 6379
      
  - name: rate-limiting-advanced
    route: api-heavy-endpoints
    config:
      limits:
        - name: tenant-limit
          config:
            minute: 50
          header_name: X-Tenant-ID
```

### Headers de Response

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1702656060
Retry-After: 30
```

---

## Encryption

### Em Repouso

```yaml
# AWS RDS
storage_encrypted: true
kms_key_id: "arn:aws:kms:sa-east-1:123456789:key/abc123"

# S3
server_side_encryption_configuration:
  rule:
    apply_server_side_encryption_by_default:
      sse_algorithm: "aws:kms"
      kms_master_key_id: "arn:aws:kms:..."
```

### Em Trânsito

```yaml
# TLS 1.3 obrigatório
ssl_protocols: TLSv1.3
ssl_ciphers: TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256
ssl_prefer_server_ciphers: on
```

### Campos Sensíveis (Application-level)

```typescript
// src/crypto/field-encryption.service.ts
import { createCipheriv, createDecipheriv, randomBytes } from 'crypto';

@Injectable()
export class FieldEncryptionService {
  private algorithm = 'aes-256-gcm';
  private key = Buffer.from(process.env.FIELD_ENCRYPTION_KEY, 'hex');

  encrypt(plaintext: string): string {
    const iv = randomBytes(16);
    const cipher = createCipheriv(this.algorithm, this.key, iv);
    
    let encrypted = cipher.update(plaintext, 'utf8', 'hex');
    encrypted += cipher.final('hex');
    
    const authTag = cipher.getAuthTag();
    
    return `${iv.toString('hex')}:${authTag.toString('hex')}:${encrypted}`;
  }

  decrypt(ciphertext: string): string {
    const [ivHex, authTagHex, encrypted] = ciphertext.split(':');
    const iv = Buffer.from(ivHex, 'hex');
    const authTag = Buffer.from(authTagHex, 'hex');
    
    const decipher = createDecipheriv(this.algorithm, this.key, iv);
    decipher.setAuthTag(authTag);
    
    let decrypted = decipher.update(encrypted, 'hex', 'utf8');
    decrypted += decipher.final('utf8');
    
    return decrypted;
  }
}
```

---

## LGPD Compliance

### Mapeamento de Dados Pessoais

| Dado | Classificação | Retenção | Anonimização |
|------|---------------|----------|--------------|
| Email | Pessoal | Conta ativa + 5 anos | Hash |
| CPF/CNPJ | Sensível | Conta ativa + 5 anos | Mascaramento |
| Endereço | Pessoal | Conta ativa | Remoção |
| Logs de acesso | Metadado | 90 dias | IP truncado |

### Consentimento

```typescript
// src/models/consent.entity.ts
@Entity('user_consents')
export class UserConsent {
  @PrimaryGeneratedColumn('uuid')
  id: string;

  @Column()
  userId: string;

  @Column()
  consentType: 'terms' | 'privacy' | 'marketing' | 'data_processing';

  @Column()
  version: string;

  @Column()
  grantedAt: Date;

  @Column({ nullable: true })
  revokedAt: Date;

  @Column()
  ipAddress: string;

  @Column('jsonb')
  metadata: Record<string, any>;
}
```

### Direito ao Esquecimento

```typescript
// src/services/data-erasure.service.ts
@Injectable()
export class DataErasureService {
  async processErasureRequest(userId: string): Promise<ErasureReport> {
    const report: ErasureReport = { userId, erasedAt: new Date(), items: [] };

    // 1. Anonimizar dados pessoais
    await this.userRepository.update(userId, {
      email: `deleted_${userId}@erased.local`,
      name: 'Usuário Removido',
      phone: null,
      address: null,
    });
    report.items.push({ table: 'users', action: 'anonymized' });

    // 2. Remover de sistemas externos
    await this.notificationService.unsubscribeAll(userId);
    report.items.push({ table: 'notification_subscriptions', action: 'deleted' });

    // 3. Manter dados para compliance (anonimizados)
    await this.auditService.logErasure(report);

    return report;
  }
}
```

---

## Auditoria

### Estrutura de Log

```json
{
  "timestamp": "2025-12-15T10:30:00.000Z",
  "level": "INFO",
  "service": "kanban-api",
  "tenantId": "tenant-123",
  "userId": "user-456",
  "action": "licitacao.status.update",
  "resource": {
    "type": "Licitacao",
    "id": "lic-789"
  },
  "changes": {
    "before": { "status": "RECEBIDO" },
    "after": { "status": "ANALISANDO" }
  },
  "context": {
    "ip": "192.168.1.100",
    "userAgent": "Mozilla/5.0...",
    "requestId": "req-abc",
    "correlationId": "corr-def"
  }
}
```

---

## Checklist de Segurança

- [ ] TLS 1.3 em todos endpoints
- [ ] JWT com RS256, rotação de keys
- [ ] Rate limiting por tenant
- [ ] Input validation (Joi/Zod)
- [ ] SQL parameterizado (no raw queries)
- [ ] CORS restritivo
- [ ] CSP headers configurados
- [ ] Secrets em Vault/AWS Secrets Manager
- [ ] Logs sem dados sensíveis
- [ ] Penetration testing trimestral

---

*Referência: [README.md](../../README.md) - Seção 7*
