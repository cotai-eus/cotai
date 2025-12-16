#!/bin/bash

# Build script for CotAI Keycloak Extensions
# Builds the Docker image with custom Keycloak SPIs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

IMAGE_NAME="cotai-keycloak-custom"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Building CotAI Keycloak Custom Image${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "Image: ${YELLOW}${FULL_IMAGE_NAME}${NC}"
echo ""

# Build the Docker image
echo -e "${GREEN}[1/3]${NC} Building Docker image..."
docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  -t "${FULL_IMAGE_NAME}" \
  -f Dockerfile \
  .

if [ $? -eq 0 ]; then
  echo -e "${GREEN}✓${NC} Docker image built successfully"
else
  echo -e "${RED}✗${NC} Docker build failed"
  exit 1
fi

# Verify the JAR is in the image
echo ""
echo -e "${GREEN}[2/3]${NC} Verifying extension JAR..."
docker run --rm "${FULL_IMAGE_NAME}" \
  ls -lh /opt/keycloak/providers/

if [ $? -eq 0 ]; then
  echo -e "${GREEN}✓${NC} Extension JAR verified in image"
else
  echo -e "${RED}✗${NC} Extension JAR not found in image"
  exit 1
fi

# Show image details
echo ""
echo -e "${GREEN}[3/3]${NC} Image details:"
docker images "${FULL_IMAGE_NAME}"

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "Next steps:"
echo -e "  1. Update docker-compose.dev.yml to use: ${YELLOW}${FULL_IMAGE_NAME}${NC}"
echo -e "  2. Run: ${YELLOW}docker compose -f docker-compose.dev.yml up -d keycloak${NC}"
echo -e "  3. Access Keycloak: ${YELLOW}http://localhost:8080${NC}"
echo ""
