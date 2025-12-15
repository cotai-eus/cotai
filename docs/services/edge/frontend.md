# Frontend (Web/PWA)

## Visão Geral
| Atributo | Valor |
|----------|-------|
| Responsabilidade | Portal cliente/fornecedor, backoffice |
| Stack | Next.js 14, React 18, TypeScript, TailwindCSS |
| Build | `next build` / `next start` |
| Deploy | Vercel/K8s (NGINX ingress) |

## Áreas
- Dashboard licitações, pipeline Kanban
- Portal fornecedor (responder cotações)
- Chat em tempo real, notificações
- Admin (tenants, usuários, features)

## Autenticação
- OIDC via Auth Service (Keycloak) — PKCE.
- Tokens armazenados em cookies httpOnly; refresh token rotacionado.
- Multi-tenant: `tenant_id` no JWT; subdomínio opcional.

## APIs Consumidas
- Kanban API (`/v1/licitacoes`), CRM, Quote, Chat (WS), Notification.
- Gateway expõe `/api` reverso para backend; frontend chama via HTTPS.

## Estrutura de Pastas (sugerida)
```
src/
  app/
    layout.tsx
    page.tsx
    (dashboard)/
    (licitacoes)/[id]/page.tsx
    (fornecedor)/cotacoes/[id]/page.tsx
  components/
    ui/
    charts/
    tables/
  lib/
    api.ts (REST client com fetch + interceptors)
    auth.ts (Keycloak adapter)
    tenant.ts
  hooks/
  styles/
```

## Clientes
- REST: `fetch` com retry/backoff; header `X-Tenant-ID`.
- WebSocket Chat: Socket.io client; reconexão exponencial.

## Observabilidade
- Web vitals (Next analytics) + Sentry para erros.
- Feature flags: LaunchDarkly ou simple env-based.

## Build/Run
```bash
npm ci
npm run lint
npm run build
npm run start
```

## Segurança
- CSP, XSS protection; validação de inputs client-side.
- Desabilitar eval; usar `next/headers` para tokens httpOnly via API routes.

*Interfaces seguem UX do Kanban e portal fornecedor; dark/light via Tailwind tokens.*