# API Gateway (Kong)

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Responsabilidade | Roteamento, auth, rate limiting |
| Stack | Kong + Redis (counters) |
| Deploy | Helm chart no K8s |

## Rotas Principais
- `/auth/*` → Auth Service
- `/kanban/*` → Kanban API
- `/crm/*` → CRM Service
- `/quote/*` → Quote Service
- `/chat/*` → Chat Service (WS upgrade)
- `/notifications/*` → Notification Service

## Plugins
- `oidc` (Keycloak) ou `jwt` — valida tokens.
- `acl` — perfis por rota.
- `rate-limiting` (Redis) — por tenant e IP.
- `cors` — domínios autorizados.
- `request-transformer` — injeta `X-Tenant-ID` do JWT.
- `prometheus` — métricas.

## Exemplo de Declaração (Kong YAML)
```yaml
services:
  - name: kanban-api
    url: http://kanban-api:3000
    routes:
      - name: kanban
        paths: ["/kanban"]
plugins:
  - name: jwt
  - name: rate-limiting
    config:
      minute: 200
      policy: redis
      redis_host: redis
      redis_port: 6379
```

## Segurança
- TLS mútuo opcional para backends internos.
- WAF externo (CloudFront/WAF) na borda.
- Headers de tracing: `x-request-id`, `x-correlation-id` propagados.

## Observabilidade
- `/metrics` Prometheus plugin.
- Logs estruturados para ELK/ClickHouse.

*Gateway centraliza autenticação e rate limiting; integra com tenant resolution.*