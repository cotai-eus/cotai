# Multi-stage build for Tenant Manager Service

# ====================
# Build Stage
# ====================
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary (optimized for production)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o bin/tenant-manager \
    ./cmd/server

# ====================
# Runtime Stage
# ====================
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 cotai && \
    adduser -D -u 1000 -G cotai cotai

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/tenant-manager .

# Copy migration files (needed for schema provisioning)
COPY --from=builder /app/migrations ./migrations

# Change ownership to non-root user
RUN chown -R cotai:cotai /app

# Switch to non-root user
USER cotai

# Expose HTTP and gRPC ports
EXPOSE 8082 9082

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Run the service
CMD ["./tenant-manager"]
