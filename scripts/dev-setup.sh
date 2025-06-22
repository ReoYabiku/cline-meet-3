#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up WebRTC Signaling Server Development Environment...${NC}"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go first.${NC}"
    exit 1
fi

echo -e "${GREEN}All required tools are installed!${NC}"

# Create necessary directories
echo -e "${YELLOW}Creating necessary directories...${NC}"
mkdir -p logs
mkdir -p data/redis
mkdir -p data/coturn

# Make scripts executable
echo -e "${YELLOW}Making scripts executable...${NC}"
chmod +x scripts/*.sh

# Download Go dependencies
echo -e "${YELLOW}Downloading Go dependencies...${NC}"
go mod download
go mod tidy

# Build the application for development
echo -e "${YELLOW}Building application...${NC}"
go build -o bin/signaling ./cmd/signaling

echo -e "${GREEN}Development environment setup completed!${NC}"

# Show available commands
echo -e "${BLUE}Available development commands:${NC}"
echo ""
echo -e "${YELLOW}Local development (without Docker):${NC}"
echo "  go run ./cmd/signaling"
echo ""
echo -e "${YELLOW}Docker Compose development:${NC}"
echo "  cd deployments/docker-compose"
echo "  docker compose up --build"
echo ""
echo -e "${YELLOW}Build Docker images:${NC}"
echo "  ./scripts/build.sh"
echo ""
echo -e "${YELLOW}Deploy to Kubernetes:${NC}"
echo "  ./scripts/deploy.sh"
echo ""
echo -e "${YELLOW}Test the application:${NC}"
echo "  Open http://localhost:8080 in your browser"
echo ""
echo -e "${YELLOW}View logs:${NC}"
echo "  docker compose logs -f signaling"
echo "  docker compose logs -f redis"
echo "  docker compose logs -f coturn"
echo ""

# Check if minikube is available for Kubernetes development
if command -v minikube &> /dev/null; then
    echo -e "${BLUE}Minikube detected! Kubernetes development commands:${NC}"
    echo -e "${YELLOW}Start minikube:${NC}"
    echo "  minikube start"
    echo ""
    echo -e "${YELLOW}Enable ingress addon:${NC}"
    echo "  minikube addons enable ingress"
    echo ""
    echo -e "${YELLOW}Build images in minikube:${NC}"
    echo "  eval \$(minikube docker-env)"
    echo "  ./scripts/build.sh"
    echo ""
    echo -e "${YELLOW}Deploy to minikube:${NC}"
    echo "  ./scripts/deploy.sh"
    echo ""
    echo -e "${YELLOW}Access application:${NC}"
    echo "  minikube service signaling-service -n webrtc-signaling"
    echo ""
fi

# Environment variables template
echo -e "${BLUE}Environment Variables (optional):${NC}"
cat << EOF

You can create a .env file with the following variables:

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
READ_TIMEOUT=60
WRITE_TIMEOUT=60

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# STUN/TURN Configuration
STUN_URL=stun:localhost:3478
TURN_URL=turn:localhost:3478

# Docker Registry (for build script)
REGISTRY=your-registry.com

# Kubernetes Context (for deploy script)
KUBECTL_CONTEXT=your-context

EOF

echo -e "${GREEN}Development environment is ready!${NC}"
echo -e "${YELLOW}Start developing with: go run ./cmd/signaling${NC}"
