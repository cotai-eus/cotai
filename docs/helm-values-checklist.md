# Helm Values Quick-Checklist

Use this checklist when preparing `values.yaml` for Helm deployments. It is a pragmatic set of keys and validations to review before creating PRs and running deployments.

- **Image & Release**
  - `image.repository` set and approved
  - `image.tag` locked or follows semver pattern
  - `image.pullPolicy` set appropriately (`IfNotPresent`/`Always`)

- **Replicas & Autoscaling**
  - `replicaCount` sensible default for staging
  - `autoscaling.enabled` true/false, `minReplicas`/`maxReplicas` defined
  - HPA metrics and thresholds configured (CPU/queue depth)

- **Resources**
  - `resources.requests` and `limits` set for CPU/memory
  - `priorityClassName` set if required

- **Probes**
  - `livenessProbe` and `readinessProbe` configured and tested
  - `startupProbe` configured for slow-start services

- **Configuration & Secrets**
  - Env vars via `envFrom` referencing `Secret` or `ConfigMap`
  - Sensitive values injected via ExternalSecrets or Vault integration
  - `secrets.rotation` policy documented

- **Networking & Security**
  - `service.type` correct (ClusterIP/LoadBalancer) and annotations set
  - NetworkPolicy configured for pod-to-pod restrictions
  - TLS certs referenced (Ingress/Service) and `tls.enabled`

- **Affinity & Tolerations**
  - Node affinity / anti-affinity for HA
  - Tolerations for critical workloads if needed

- **Storage & Persistence**
  - PVCs defined with storageClass and size
  - Retention/backup policy for persistent volumes

- **Observability**
  - `metrics.enabled` and serviceMonitor present for Prometheus
  - Sidecar for logs/agent if required
  - Tracing headers propagated (correlation-id)

- **Lifecycle & Migrations**
  - Job hooks for DB migrations present and idempotent
  - Pre-install/post-upgrade hooks validated

- **Misc**
  - `imagePullSecrets` configured for private registries
  - `rbac.create` and `serviceAccount.name` correct
  - `podDisruptionBudget` set for critical services

Refer to [docs/deployment/kubernetes-deploy.md](./deployment/kubernetes-deploy.md) and [docs/implementation-checklist.md](./implementation-checklist.md) for policies and examples.
