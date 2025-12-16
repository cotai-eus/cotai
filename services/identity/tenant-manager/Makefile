.PHONY: build test lint run docker-build docker-run clean proto-gen

BINARY_NAME=tenant-manager
DOCKER_IMAGE=cotai-tenant-manager
GO_VERSION=1.21

# Build binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags="-w -s" -o bin/$(BINARY_NAME) ./cmd/server

# Run tests
test:
	@echo "Running unit tests..."
	@go test -v -race -coverprofile=coverage.out ./...

# Integration tests
test-integration:
	@echo "Running integration tests..."
	@go test -v -tags=integration ./test/integration/...

# E2E tests
test-e2e:
	@echo "Running E2E tests..."
	@go test -v -tags=e2e ./test/e2e/...

# Coverage report
coverage:
	@go tool cover -html=coverage.out

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run --timeout 5m

# Run service locally
run:
	@echo "Starting $(BINARY_NAME)..."
	@go run ./cmd/server

# Generate protobuf code
proto-gen:
	@echo "Generating protobuf code..."
	@./scripts/gen-proto.sh

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):latest -f deployments/docker/Dockerfile .

# Run Docker container
docker-run:
	@docker run -p 8082:8082 -p 9082:9082 --env-file .env $(DOCKER_IMAGE):latest

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build binary"
	@echo "  test            - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e        - Run E2E tests"
	@echo "  coverage        - Generate coverage report"
	@echo "  lint            - Lint code"
	@echo "  run             - Run service locally"
	@echo "  proto-gen       - Generate protobuf code"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download dependencies"
