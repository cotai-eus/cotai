#!/bin/bash
# CotAI Development Environment Validation Script
# Tests connectivity to all infrastructure services

set -e

echo "=================================="
echo "CotAI Dev Environment Validation"
echo "=================================="
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
check_service() {
    local service_name=$1
    local check_command=$2

    echo -n "Checking $service_name... "
    if eval "$check_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Check Docker is running
echo "=== Docker Status ==="
check_service "Docker daemon" "docker info"
echo ""

# Check containers are running
echo "=== Container Status ==="
RUNNING_CONTAINERS=$(docker compose -f docker-compose.dev.yml ps --format json 2>/dev/null | grep -c '"State":"running"' || echo "0")
echo "Running containers: $RUNNING_CONTAINERS/15"
echo ""

# Check PostgreSQL databases
echo "=== PostgreSQL Databases ==="
check_service "Postgres Identity (5436)" "docker exec cotai-postgres-identity pg_isready -U cotai_dev"
check_service "Postgres Core (5433)" "docker exec cotai-postgres-core pg_isready -U cotai_dev"
check_service "Postgres Resources (5434)" "docker exec cotai-postgres-resources pg_isready -U cotai_dev"
check_service "Postgres Collab (5435)" "docker exec cotai-postgres-collab pg_isready -U cotai_dev"
echo ""

# Check Redis
echo "=== Cache Services ==="
check_service "Redis (6380)" "docker exec cotai-redis redis-cli -a dev_password ping"
echo ""

# Check Message Brokers
echo "=== Message Brokers ==="
check_service "RabbitMQ (5672)" "docker exec cotai-rabbitmq rabbitmq-diagnostics -q ping"
check_service "Kafka (9092)" "docker exec cotai-kafka kafka-broker-api-versions --bootstrap-server localhost:9092"
echo ""

# Check Search & Storage
echo "=== Search & Storage ==="
check_service "Elasticsearch (9200)" "curl -s http://localhost:9200/_cluster/health"
check_service "MinIO (9000)" "curl -s http://localhost:9000/minio/health/live"
echo ""

# Check Observability
echo "=== Observability Stack ==="
check_service "Prometheus (9090)" "curl -s http://localhost:9090/-/healthy"
check_service "Grafana (3000)" "curl -s http://localhost:3000/api/health"
check_service "Jaeger (16686)" "curl -s http://localhost:16686/"
echo ""

# Check Authentication
echo "=== Authentication ==="
check_service "Keycloak (8080)" "curl -s http://localhost:8080/health/ready"
echo ""

# Summary
echo "=================================="
echo "Validation Complete!"
echo "=================================="
echo ""
echo "Access URLs:"
echo "  - Grafana:         http://localhost:3000 (admin/admin)"
echo "  - Prometheus:      http://localhost:9090"
echo "  - RabbitMQ UI:     http://localhost:15672 (cotai_dev/dev_password)"
echo "  - Jaeger UI:       http://localhost:16686"
echo "  - Kafka UI:        http://localhost:8081"
echo "  - Keycloak:        http://localhost:8080 (admin/admin)"
echo "  - MinIO Console:   http://localhost:9001 (cotai_dev/dev_password)"
echo ""
echo "For detailed status, see: DEVELOPMENT_ENVIRONMENT_STATUS.md"
