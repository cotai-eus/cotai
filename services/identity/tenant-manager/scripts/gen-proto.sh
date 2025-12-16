#!/bin/bash

# Script to generate Go code from protobuf definitions

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating protobuf code...${NC}"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}Error: protoc is not installed${NC}"
    echo "Install it with: sudo apt-get install -y protobuf-compiler"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${RED}Error: protoc-gen-go is not installed${NC}"
    echo "Install it with: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${RED}Error: protoc-gen-go-grpc is not installed${NC}"
    echo "Install it with: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# Project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Generate Go code
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/tenant/v1/tenant.proto

echo -e "${GREEN}✓ Protobuf code generated successfully${NC}"
echo -e "${GREEN}  Generated files:${NC}"
echo -e "    - proto/tenant/v1/tenant.pb.go"
echo -e "    - proto/tenant/v1/tenant_grpc.pb.go"
