# Multi-stage Dockerfile for CotAI Keycloak Extensions
# Stage 1: Build the extension JAR with Maven
# Stage 2: Copy extension into Keycloak image

# ============================================================================
# Stage 1: Build
# ============================================================================
FROM maven:3.9.6-eclipse-temurin-17-alpine AS builder

WORKDIR /build

# Copy POM first for dependency caching
COPY pom.xml .

# Download dependencies (cached layer)
RUN mvn dependency:go-offline -B

# Copy source code
COPY src/ ./src/

# Build the extension JAR
# Skip tests during build (run separately in CI)
RUN mvn clean package -DskipTests -B

# Verify the JAR was created
RUN ls -lh /build/target/*.jar

# ============================================================================
# Stage 2: Keycloak with Custom Extensions
# ============================================================================
FROM quay.io/keycloak/keycloak:23.0.3

USER root

# Copy the extension JAR to Keycloak providers directory
COPY --from=builder /build/target/cotai-keycloak-extensions-*.jar /opt/keycloak/providers/

# Set proper permissions
RUN chown keycloak:keycloak /opt/keycloak/providers/*.jar

USER keycloak

# Keycloak will automatically load providers from /opt/keycloak/providers/
# No need to manually trigger rebuild in start-dev mode

# Environment variables for configuration (can be overridden)
ENV KAFKA_BOOTSTRAP_SERVERS=kafka:9092
ENV KAFKA_PRODUCER_CLIENT_ID=keycloak-event-publisher
ENV KAFKA_ACKS=1
ENV KAFKA_RETRIES=3
ENV KAFKA_COMPRESSION_TYPE=snappy

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=5 \
  CMD curl -f http://localhost:8080/health/ready || exit 1

# Default command (can be overridden in docker-compose)
ENTRYPOINT ["/opt/keycloak/bin/kc.sh"]
CMD ["start-dev"]
