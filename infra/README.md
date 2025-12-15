# CotAI Infrastructure as Code

Este diretório contém toda a infraestrutura como código (IaC) para a plataforma CotAI, incluindo provisionamento Terraform, Helm charts e manifestos Kubernetes.

## Estrutura

```
infra/
├── terraform/          # Provisionamento de infraestrutura cloud
│   ├── modules/        # Módulos Terraform reutilizáveis
│   │   ├── eks-cluster/      # Cluster Kubernetes (EKS)
│   │   └── rds-postgres/     # Bancos de dados PostgreSQL
│   └── environments/   # Configurações por ambiente
│       ├── dev/
│       ├── staging/
│       └── prod/
├── helm/               # Helm charts para serviços
│   └── charts/
│       ├── rabbitmq/         # Message broker
│       ├── kafka/            # Event streaming
│       ├── redis/            # Cache e real-time
│       ├── prometheus-stack/ # Observabilidade
│       └── jaeger/           # Distributed tracing
└── kubernetes/         # Manifestos Kubernetes base
    └── base/
        ├── namespaces/       # Definições de namespaces
        ├── network-policies/ # Políticas de rede
        └── secrets/          # Templates de secrets
```

## Pré-requisitos

### Ferramentas Necessárias

- **Terraform**: >= 1.5.0
- **kubectl**: >= 1.28.0
- **Helm**: >= 3.12.0
- **AWS CLI**: >= 2.13.0 (para AWS) ou **gcloud** (para GCP)

### Credenciais

```bash
# AWS
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# GCP (alternativa)
gcloud auth application-default login
export GOOGLE_PROJECT="cotai-project"
export GOOGLE_REGION="us-east1"
```

## Guia de Deployment - Fase 0

### Passo 1: Provisionamento do Cluster Kubernetes (Semanas 1-2)

#### 1.1. Configurar Backend Terraform

```bash
cd infra/terraform/environments/dev

# Criar bucket S3 para Terraform state (apenas primeira vez)
aws s3 mb s3://cotai-terraform-state-dev --region us-east-1

# Criar DynamoDB table para state locking
aws dynamodb create-table \
  --table-name cotai-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

#### 1.2. Criar VPC e Subnets (se não existir)

```bash
# Aplicar configuração de rede
terraform init
terraform plan -target=module.vpc
terraform apply -target=module.vpc
```

#### 1.3. Provisionar Cluster EKS

```bash
# Provisionar cluster Kubernetes
terraform plan -target=module.eks_cluster
terraform apply -target=module.eks_cluster

# Configurar kubectl
aws eks update-kubeconfig --name cotai-dev-cluster --region us-east-1
```

#### 1.4. Aplicar Namespaces e Network Policies

```bash
cd ../../kubernetes/base

# Criar namespaces
kubectl apply -f namespaces/namespaces.yaml

# Aplicar network policies
kubectl apply -f network-policies/
```

**Critérios de Aceitação:**
- [ ] Cluster Kubernetes com 3+ nodes em múltiplas AZs
- [ ] kubectl acesso a todos namespaces
- [ ] Network policies aplicadas com default-deny

### Passo 2: Provisionamento de Databases (Semana 2)

#### 2.1. Provisionar RDS PostgreSQL (4 instâncias)

```bash
cd ../../terraform/environments/dev

# Identity database
terraform plan -target=module.rds_identity
terraform apply -target=module.rds_identity

# Core database
terraform plan -target=module.rds_core
terraform apply -target=module.rds_core

# Resources database
terraform plan -target=module.rds_resources
terraform apply -target=module.rds_resources

# Collaboration database
terraform plan -target=module.rds_collab
terraform apply -target=module.rds_collab
```

#### 2.2. Configurar Connection Pooling (PgBouncer)

```bash
# Deploy PgBouncer via Helm (opcional, recomendado para produção)
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install pgbouncer-identity bitnami/pgbouncer \
  --namespace cotai-dev \
  --set postgresql.host=<RDS_IDENTITY_ENDPOINT> \
  --set postgresql.port=5432
```

**Critérios de Aceitação:**
- [ ] 4 instâncias RDS Multi-AZ criadas
- [ ] PostgreSQL acessível de pods K8s
- [ ] Backups automáticos configurados (30 dias)

### Passo 3: Messaging Infrastructure (Semana 2)

#### 3.1. Deploy RabbitMQ Cluster

```bash
cd ../../../helm/charts/rabbitmq

# Criar secret com credenciais
kubectl create secret generic rabbitmq-credentials \
  --from-literal=password=$(openssl rand -base64 32) \
  --namespace cotai-dev

# Criar secret com load definitions
kubectl create secret generic rabbitmq-load-definition \
  --from-file=load_definition.json=definitions.json \
  --namespace cotai-dev

# Deploy RabbitMQ
helm repo add bitnami https://charts.bitnami.com/bitnami

helm install rabbitmq bitnami/rabbitmq \
  --namespace cotai-dev \
  --values values.yaml \
  --wait
```

#### 3.2. Deploy Kafka Cluster

```bash
cd ../kafka

# Deploy Kafka
helm install kafka bitnami/kafka \
  --namespace cotai-dev \
  --values values.yaml \
  --wait

# Verificar tópicos criados
kubectl run kafka-client --rm -ti --restart='Never' --image docker.io/bitnami/kafka:3.6.1 --namespace cotai-dev --command -- \
  kafka-topics.sh --bootstrap-server kafka:9092 --list
```

**Critérios de Aceitação:**
- [ ] RabbitMQ cluster com 3 nodes operacional
- [ ] Kafka cluster com 3 brokers operacional
- [ ] Exchanges e queues criados no RabbitMQ
- [ ] Tópicos Kafka criados com partições corretas

### Passo 4: Cache Layer (Semana 2)

#### 4.1. Deploy Redis Cluster

```bash
cd ../redis

# Criar secret com senha
kubectl create secret generic redis-credentials \
  --from-literal=password=$(openssl rand -base64 32) \
  --namespace cotai-dev

# Deploy Redis Cluster
helm install redis bitnami/redis-cluster \
  --namespace cotai-dev \
  --values values.yaml \
  --wait

# Verificar cluster status
kubectl run redis-client --rm -ti --restart='Never' --image docker.io/bitnami/redis-cluster:7.2.4 --namespace cotai-dev --command -- \
  redis-cli -c -h redis-cluster -a $(kubectl get secret redis-credentials -o jsonpath='{.data.password}' | base64 -d) cluster info
```

**Critérios de Aceitação:**
- [ ] Redis cluster com 6 nodes (3 masters + 3 replicas)
- [ ] Cluster info mostra todos nodes conectados
- [ ] Teste pub/sub bem-sucedido

### Passo 5: Observability Stack (Semana 3)

#### 5.1. Deploy Prometheus + Grafana

```bash
cd ../prometheus-stack

# Add Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

# Deploy kube-prometheus-stack
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace cotai-observability \
  --create-namespace \
  --values values.yaml \
  --wait

# Aplicar custom alerts
kubectl apply -f custom-alerts.yaml
```

#### 5.2. Deploy Jaeger

```bash
cd ../jaeger

# Install Jaeger Operator
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts

kubectl create namespace observability
helm install jaeger-operator jaegertracing/jaeger-operator \
  --namespace observability

# Deploy Jaeger instance
kubectl apply -f values.yaml
```

#### 5.3. Acessar Dashboards

```bash
# Port-forward Grafana
kubectl port-forward -n cotai-observability svc/prometheus-grafana 3000:80

# Port-forward Jaeger
kubectl port-forward -n observability svc/jaeger-query 16686:16686

# Port-forward Prometheus
kubectl port-forward -n cotai-observability svc/prometheus-kube-prometheus-prometheus 9090:9090
```

**URLs:**
- Grafana: http://localhost:3000 (admin / senha definida em values.yaml)
- Jaeger: http://localhost:16686
- Prometheus: http://localhost:9090

**Critérios de Aceitação:**
- [ ] Prometheus coletando métricas de pods de teste
- [ ] Grafana exibindo dashboards de Kubernetes, Kafka, RabbitMQ, Redis
- [ ] Jaeger recebendo traces (testar com pod de exemplo)
- [ ] Alertas customizados carregados no Prometheus

### Passo 6: Security Foundations (Semana 4)

#### 6.1. Deploy External Secrets Operator

```bash
helm repo add external-secrets https://charts.external-secrets.io

helm install external-secrets \
  external-secrets/external-secrets \
  --namespace external-secrets-system \
  --create-namespace

# Configurar SecretStore (AWS Secrets Manager)
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: cotai-dev
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
EOF
```

#### 6.2. Deploy cert-manager

```bash
helm repo add jetstack https://charts.jetstack.io

helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true

# Criar ClusterIssuer (Let's Encrypt)
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: devops@cotai.com.br
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

**Critérios de Aceitação:**
- [ ] External Secrets Operator sincronizando secrets do AWS Secrets Manager
- [ ] cert-manager emitindo certificados automaticamente
- [ ] Certificado wildcard *.cotai.com.br válido

## Verificação Final - Fase 0

Execute os seguintes comandos para validar a infraestrutura:

```bash
# Verificar nodes do cluster
kubectl get nodes -o wide

# Verificar todos pods em todos namespaces
kubectl get pods --all-namespaces

# Verificar PVCs
kubectl get pvc --all-namespaces

# Verificar services
kubectl get svc --all-namespaces

# Health check RabbitMQ
kubectl exec -it rabbitmq-0 -n cotai-dev -- rabbitmqctl cluster_status

# Health check Kafka
kubectl exec -it kafka-0 -n cotai-dev -- kafka-broker-api-versions.sh --bootstrap-server localhost:9092

# Health check Redis
kubectl exec -it redis-cluster-0 -n cotai-dev -- redis-cli -a $(kubectl get secret redis-credentials -n cotai-dev -o jsonpath='{.data.password}' | base64 -d) cluster nodes

# Verificar métricas no Prometheus
curl http://localhost:9090/api/v1/query?query=up
```

## Troubleshooting

### Pods não iniciam

```bash
# Descrever pod para ver eventos
kubectl describe pod <pod-name> -n <namespace>

# Ver logs
kubectl logs <pod-name> -n <namespace>

# Ver logs de container específico
kubectl logs <pod-name> -c <container-name> -n <namespace>
```

### RabbitMQ cluster não forma

```bash
# Verificar logs
kubectl logs rabbitmq-0 -n cotai-dev

# Forçar cluster join
kubectl exec -it rabbitmq-1 -n cotai-dev -- rabbitmqctl stop_app
kubectl exec -it rabbitmq-1 -n cotai-dev -- rabbitmqctl reset
kubectl exec -it rabbitmq-1 -n cotai-dev -- rabbitmqctl join_cluster rabbit@rabbitmq-0.rabbitmq-headless.cotai-dev.svc.cluster.local
kubectl exec -it rabbitmq-1 -n cotai-dev -- rabbitmqctl start_app
```

### Kafka topics não são criados

```bash
# Criar tópico manualmente
kubectl exec -it kafka-0 -n cotai-dev -- kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --topic edital.raw \
  --partitions 12 \
  --replication-factor 3 \
  --config retention.ms=604800000
```

## Próximos Passos

Após completar a Fase 0, prosseguir para:

1. **Fase 1**: Implementação de serviços de Identity
   - Keycloak deployment
   - Tenant Manager service
   - Audit Service

2. **CI/CD Setup**: Configurar GitHub Actions para build e deploy automático

3. **Documentação**: Criar runbooks para operações comuns

## Recursos Adicionais

- [Documentação Terraform EKS](https://registry.terraform.io/modules/terraform-aws-modules/eks/aws/latest)
- [Bitnami Helm Charts](https://github.com/bitnami/charts)
- [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)
- [Jaeger Operator](https://www.jaegertracing.io/docs/latest/operator/)

## Suporte

Para questões sobre infraestrutura, abra uma issue no repositório ou contate a equipe Platform/Infrastructure.
