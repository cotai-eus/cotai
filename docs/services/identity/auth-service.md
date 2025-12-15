# Auth Service

## Visão Geral

| Atributo | Valor |
|----------|-------|
| **Bounded Context** | Identity & Compliance |
| **Responsabilidade** | Autenticação, emissão de tokens, gestão de sessões |
| **Stack** | Keycloak 22+ / Node.js (fallback) |
| **Database** | PostgreSQL (schema: `auth`) |
| **Porta** | 8080 (Keycloak) / 3001 (Node.js) |
| **Protocolo** | REST + OIDC |

---

## Responsabilidades

1. **Autenticação de usuários** (login/logout)
2. **Emissão de tokens** (access, refresh, ID tokens)
3. **Gestão de sessões** (SSO, session timeout)
4. **Multi-factor Authentication** (TOTP, SMS)
5. **Federação de identidade** (Google, Microsoft, SAML)
6. **Password policies** (complexity, expiration)
7. **Brute-force protection**

---

## Arquitetura

```
┌─────────────┐     ┌──────────────┐     ┌────────────┐
│   Frontend  │────►│ API Gateway  │────►│   Keycloak │
│   (React)   │     │    (Kong)    │     │  (Auth)    │
└─────────────┘     └──────────────┘     └──────┬─────┘
                                                │
                                         ┌──────▼─────┐
                                         │ PostgreSQL │
                                         │  (auth)    │
                                         └────────────┘
```

---

## Configuração Keycloak

### Realm Configuration

```json
{
  "realm": "cotai",
  "enabled": true,
  "sslRequired": "external",
  "registrationAllowed": false,
  "loginWithEmailAllowed": true,
  "duplicateEmailsAllowed": false,
  "bruteForceProtected": true,
  "permanentLockout": false,
  "maxFailureWaitSeconds": 900,
  "minimumQuickLoginWaitSeconds": 60,
  "waitIncrementSeconds": 60,
  "quickLoginCheckMilliSeconds": 1000,
  "maxDeltaTimeSeconds": 43200,
  "failureFactor": 5,
  "passwordPolicy": "length(12) and upperCase(1) and lowerCase(1) and digit(1) and specialChars(1) and notUsername()"
}
```

### Client Configuration

```json
{
  "clientId": "cotai-web",
  "name": "CotAI Web Application",
  "protocol": "openid-connect",
  "publicClient": true,
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": false,
  "serviceAccountsEnabled": false,
  "authorizationServicesEnabled": false,
  "redirectUris": [
    "https://app.cotai.com.br/*",
    "http://localhost:3000/*"
  ],
  "webOrigins": [
    "https://app.cotai.com.br",
    "http://localhost:3000"
  ],
  "defaultClientScopes": [
    "openid",
    "profile",
    "email",
    "roles",
    "tenant"
  ]
}
```

---

## API Endpoints

### OpenID Connect Discovery

```http
GET /.well-known/openid-configuration
```

### Authorization

```http
GET /realms/cotai/protocol/openid-connect/auth
    ?response_type=code
    &client_id=cotai-web
    &redirect_uri=https://app.cotai.com.br/callback
    &scope=openid profile email tenant
    &state=xyz123
    &code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM
    &code_challenge_method=S256
```

### Token Exchange

```http
POST /realms/cotai/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=abc123
&redirect_uri=https://app.cotai.com.br/callback
&client_id=cotai-web
&code_verifier=dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "id_token": "eyJhbGciOiJSUzI1NiIs...",
  "scope": "openid profile email tenant"
}
```

### Refresh Token

```http
POST /realms/cotai/protocol/openid-connect/token
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token
&refresh_token=eyJhbGciOiJIUzI1NiIs...
&client_id=cotai-web
```

### Logout

```http
POST /realms/cotai/protocol/openid-connect/logout
Content-Type: application/x-www-form-urlencoded

refresh_token=eyJhbGciOiJIUzI1NiIs...
&client_id=cotai-web
```

### User Info

```http
GET /realms/cotai/protocol/openid-connect/userinfo
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

**Response:**
```json
{
  "sub": "user-uuid-123",
  "email": "joao@empresa.com.br",
  "email_verified": true,
  "name": "João Silva",
  "preferred_username": "joao.silva",
  "tenant_id": "tenant-uuid-456",
  "roles": ["admin", "commercial"]
}
```

---

## Custom Mapper: Tenant ID

```java
// TenantIdMapper.java
public class TenantIdMapper extends AbstractOIDCProtocolMapper {
    @Override
    protected void setClaim(IDToken token, ProtocolMapperModel model,
                           UserSessionModel session, KeycloakSession keycloak,
                           ClientSessionContext ctx) {
        UserModel user = session.getUser();
        String tenantId = user.getFirstAttribute("tenant_id");
        token.getOtherClaims().put("tenant_id", tenantId);
    }
}
```

---

## Modelo de Dados

```sql
-- Schema: auth (gerenciado pelo Keycloak)
-- Principais tabelas customizadas:

CREATE TABLE user_tenant_mapping (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL,
    tenant_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, tenant_id)
);

CREATE TABLE login_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(36) NOT NULL,
    tenant_id UUID,
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL,
    failure_reason VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_login_history_user ON login_history(user_id, created_at DESC);
```

---

## Eventos Publicados

| Evento | Tópico Kafka | Payload |
|--------|--------------|---------|
| `user.login.success` | `auth.events` | `{ userId, tenantId, ip, timestamp }` |
| `user.login.failed` | `auth.events` | `{ email, ip, reason, timestamp }` |
| `user.logout` | `auth.events` | `{ userId, tenantId, timestamp }` |
| `user.password.changed` | `auth.events` | `{ userId, changedBy, timestamp }` |
| `user.mfa.enabled` | `auth.events` | `{ userId, method, timestamp }` |

---

## Variáveis de Ambiente

```bash
# Keycloak
KC_DB=postgres
KC_DB_URL=jdbc:postgresql://postgres:5432/keycloak
KC_DB_USERNAME=keycloak
KC_DB_PASSWORD=${KC_DB_PASSWORD}
KC_HOSTNAME=auth.cotai.com.br
KC_PROXY=edge
KC_HTTP_ENABLED=true
KC_HEALTH_ENABLED=true
KC_METRICS_ENABLED=true

# SMTP
KC_SPI_EMAIL_SENDER=ses
KC_SPI_EMAIL_SES_REGION=sa-east-1
```

---

## Docker Compose

```yaml
services:
  keycloak:
    image: quay.io/keycloak/keycloak:22.0
    command: start
    environment:
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak
      KC_DB_USERNAME: keycloak
      KC_DB_PASSWORD: ${KC_DB_PASSWORD}
      KC_HOSTNAME: auth.cotai.com.br
      KC_PROXY: edge
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

---

## Exemplo de Uso (Frontend)

```typescript
// src/auth/keycloak.ts
import Keycloak from 'keycloak-js';

const keycloak = new Keycloak({
  url: 'https://auth.cotai.com.br',
  realm: 'cotai',
  clientId: 'cotai-web',
});

export const initAuth = async (): Promise<boolean> => {
  const authenticated = await keycloak.init({
    onLoad: 'check-sso',
    silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
    pkceMethod: 'S256',
  });
  
  if (authenticated) {
    // Refresh token antes de expirar
    setInterval(() => {
      keycloak.updateToken(60).catch(() => keycloak.login());
    }, 60000);
  }
  
  return authenticated;
};

export const getToken = (): string | undefined => keycloak.token;
export const getTenantId = (): string => keycloak.tokenParsed?.tenant_id;
export const logout = (): void => keycloak.logout();
```

---

## Métricas Prometheus

```yaml
# Keycloak metrics endpoint: /metrics
keycloak_logins_total{realm="cotai", provider="keycloak", client_id="cotai-web"} 1523
keycloak_failed_login_attempts_total{realm="cotai", provider="keycloak"} 45
keycloak_registrations_total{realm="cotai"} 89
keycloak_active_sessions{realm="cotai"} 234
```

---

## Health Check

```http
GET /health/ready
```

```json
{
  "status": "UP",
  "checks": [
    { "name": "Database", "status": "UP" },
    { "name": "Infinispan", "status": "UP" }
  ]
}
```

---

*Referência: [Arquitetura](../../README.md) | [Segurança](../architecture/security-compliance.md)*
