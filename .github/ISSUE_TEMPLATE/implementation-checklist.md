---
name: Implementation Task - Checklist
about: Use this template to track implementation tasks against the central implementation checklist
title: '[IMPLEMENT] <serviço/feature> - curto resumo'
labels: ['implementation', 'needs-triage']
assignees: []
---

## Context
- Serviço / Componente: 
- Epic / Ticket relacionado: 
- Referência de design / doc: [Implementation Checklist](../../docs/implementation-checklist.md)

## Objetivo
Descrever brevemente o que será entregue.

## Checklist de Implementação (marcar conforme concluído)
- [ ] Arquitetura e decisões documentadas (ADRs atualizadas)
- [ ] Diagramas e dependências mapeados
- [ ] Variáveis de configuração validadas (env schema)
- [ ] Health/readiness endpoints expostos
- [ ] Métricas, logs estruturados e tracing instrumentados
- [ ] Retries/timeouts/circuit-breakers configurados
- [ ] Migrations aplicáveis (idempotentes) e testadas em staging
- [ ] Mensageria: contratos (events/commands) validados e DLQ configurada
- [ ] Security: authn/authz implementado, secrets via ExternalSecrets/Vault
- [ ] Docker image: multi-stage, non-root, scanned (SBOM)
- [ ] Helm values preparados e revisados (ver docs/helm-values-checklist.md)
- [ ] CI jobs (lint/tests/scan) configurados e verdes
- [ ] Runbook mínimo criado (rollbacks, observability links)

## Ambiente de Teste / Observações
- Branch: 
- Ambiente de deploy: 
- Observações adicionais:
