<!-- Use this PR template to track CI gates and promotion steps -->
## Pull Request Checklist

### Code Quality
- [ ] Lint (ESLint / Flake8 / golangci-lint) ✅
- [ ] Formatting (prettier/black/gofmt) ✅
- [ ] Unit tests (coverage acceptable) ✅

### Tests & Contracts
- [ ] Integration tests (Kafka/Rabbit mocks or test containers) ✅
- [ ] Contract tests (OpenAPI/Protobuf verification) ✅

### Security & Supply Chain
- [ ] SAST/DAST scans ✅
- [ ] Dependency scan (OWASP, Snyk, etc.) ✅
- [ ] SBOM generated and verified ✅

### Build & Image
- [ ] Build artifact created ✅
- [ ] Container image built and scanned ✅
- [ ] Image signed (if applicable) ✅

### Deployment Gates
- [ ] Deploy to `staging` (canary) and run smoke tests ✅
- [ ] Monitor canary metrics (p95 latency, error rate, queue backlog) ✅
- [ ] Promote to `production` after verification ✅

### Operational
- [ ] Helm values reviewed (see `docs/helm-values-checklist.md`) ✅
- [ ] Runbook / rollback instructions attached ✅
- [ ] Monitoring dashboards / alerts updated if needed ✅

---
Referências: [Implementation Checklist](../../docs/implementation-checklist.md)
