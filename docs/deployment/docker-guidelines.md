# Docker Guidelines

## Princípios
- Imagens pequenas (alpine/slim) e sem ferramentas extras.
- Multi-stage builds (builder → runtime).
- Usuário não-root no runtime.
- HEALTHCHECK sempre que possível.

## Node.js (exemplo NestJS)
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
USER node
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

## Go
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/server

FROM gcr.io/distroless/base
COPY --from=builder /bin/service /service
USER 65532
CMD ["/service"]
```

## Python (FastAPI)
```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY pyproject.toml poetry.lock ./
RUN pip install poetry && poetry install --no-dev
COPY . .
EXPOSE 8000
CMD ["poetry", "run", "uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```

## Boas Práticas
- Fixar versões de dependências.
- Usar `.dockerignore` para reduzir contexto.
- Variáveis sensíveis via env/secret, não `ARG`.
- Definir `CMD` simples; evitar `ENTRYPOINT` rígido.
- Compressão de camadas (buildx) e SBOM opcional.

*Aplicar padrão por serviço; ajustar portas conforme docs.*