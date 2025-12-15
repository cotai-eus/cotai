# Kubernetes Deploy

## Padrões
- Namespace por ambiente (`dev`, `stg`, `prod`).
- ConfigMaps para configs não sensíveis; Secrets para credenciais.
- HPA: CPU 70%, Memory 80% (mín 2 pods, máx 10).
- Readiness/Liveness probes obrigatórias.

## Exemplo Helm values (serviço Node)
```yaml
deployment:
  image: cotai/kanban-api:latest
  replicas: 2
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
  env:
    - name: NODE_ENV
      value: production
  livenessProbe:
    httpGet:
      path: /health/live
      port: 3000
    initialDelaySeconds: 20
  readinessProbe:
    httpGet:
      path: /health/ready
      port: 3000
    initialDelaySeconds: 20
service:
  port: 80
  targetPort: 3000
ingress:
  enabled: true
  hosts:
    - host: api.cotai.com.br
      paths:
        - path: /kanban
          pathType: Prefix
```

## Observabilidade
- Prometheus annotations: `prometheus.io/scrape: "true"`, `prometheus.io/port: "3000"`, `prometheus.io/path: "/metrics"`.
- Jaeger tracing env: `JAEGER_AGENT_HOST`, `JAEGER_SAMPLER_TYPE=const`, `JAEGER_SAMPLER_PARAM=1` (ajustar em prod).

## Banco/Mensageria
- Usar StatefulSets externos (RDS, MSK, ElastiCache) quando possível.
- Configurar network policies para isolar tráfego interno.

## Secrets
- Armazenar em Secret ou External Secrets (Vault/ASM/SSM).
- Montar via envFrom ou volume, não commitar.

## Rollouts
- Strategy: RollingUpdate (maxUnavailable=0, maxSurge=1).
- Blue/Green ou Canary (Argo Rollouts) para serviços críticos (Auth, Kanban).

*Aplicar por serviço conforme stacks descritos em docs/services.*