#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME="signaling-server"
IMAGE_TAG="${IMAGE_TAG:-latest}"
COTURN_IMAGE_NAME="signaling-coturn"

echo -e "${GREEN}Building WebRTC Signaling Server...${NC}"

# Build Go application
echo -e "${YELLOW}Building Go binary...${NC}"
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o signaling ./cmd/signaling

# Build Docker images
echo -e "${YELLOW}Building Docker images...${NC}"

# Build signaling server image
echo -e "${YELLOW}Building signaling server image...${NC}"
docker build -f deployments/docker/signaling/Dockerfile -t ${IMAGE_NAME}:${IMAGE_TAG} .

# Build coturn image
echo -e "${YELLOW}Building coturn image...${NC}"
docker build -f deployments/docker/coturn/Dockerfile -t ${COTURN_IMAGE_NAME}:${IMAGE_TAG} .

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}Images built:${NC}"
echo -e "  - ${IMAGE_NAME}:${IMAGE_TAG}"
echo -e "  - ${COTURN_IMAGE_NAME}:${IMAGE_TAG}"

# Optional: Tag for registry
if [ ! -z "$REGISTRY" ]; then
    echo -e "${YELLOW}Tagging images for registry ${REGISTRY}...${NC}"
    docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}
    docker tag ${COTURN_IMAGE_NAME}:${IMAGE_TAG} ${REGISTRY}/${COTURN_IMAGE_NAME}:${IMAGE_TAG}
    
    echo -e "${GREEN}Registry tags created:${NC}"
    echo -e "  - ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
    echo -e "  - ${REGISTRY}/${COTURN_IMAGE_NAME}:${IMAGE_TAG}"
fi

# Clean up binary
rm -f signaling

echo -e "${GREEN}Build script completed!${NC}"
