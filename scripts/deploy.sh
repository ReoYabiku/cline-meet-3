#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="webrtc-signaling"
KUBECTL_CONTEXT="${KUBECTL_CONTEXT:-}"

echo -e "${GREEN}Deploying WebRTC Signaling Server to Kubernetes...${NC}"

# Set kubectl context if specified
if [ ! -z "$KUBECTL_CONTEXT" ]; then
    echo -e "${YELLOW}Setting kubectl context to ${KUBECTL_CONTEXT}...${NC}"
    kubectl config use-context $KUBECTL_CONTEXT
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if cluster is accessible
echo -e "${YELLOW}Checking cluster connectivity...${NC}"
if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi

# Create namespace
echo -e "${YELLOW}Creating namespace...${NC}"
kubectl apply -f deployments/kubernetes/namespace.yaml

# Apply ConfigMaps and Secrets
echo -e "${YELLOW}Applying ConfigMaps and Secrets...${NC}"
kubectl apply -f deployments/kubernetes/configmap.yaml
kubectl apply -f deployments/kubernetes/secret.yaml

# Deploy Redis
echo -e "${YELLOW}Deploying Redis...${NC}"
kubectl apply -f deployments/kubernetes/redis/

# Wait for Redis to be ready
echo -e "${YELLOW}Waiting for Redis to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/redis-deployment -n $NAMESPACE

# Deploy Coturn
echo -e "${YELLOW}Deploying Coturn...${NC}"
kubectl apply -f deployments/kubernetes/coturn/

# Wait for Coturn to be ready
echo -e "${YELLOW}Waiting for Coturn to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/coturn-deployment -n $NAMESPACE

# Deploy Signaling Server
echo -e "${YELLOW}Deploying Signaling Server...${NC}"
kubectl apply -f deployments/kubernetes/signaling/

# Wait for Signaling Server to be ready
echo -e "${YELLOW}Waiting for Signaling Server to be ready...${NC}"
kubectl wait --for=condition=available --timeout=300s deployment/signaling-deployment -n $NAMESPACE

# Apply Ingress
echo -e "${YELLOW}Applying Ingress...${NC}"
kubectl apply -f deployments/kubernetes/ingress.yaml

# Show deployment status
echo -e "${GREEN}Deployment completed!${NC}"
echo -e "${BLUE}Checking deployment status...${NC}"

kubectl get pods -n $NAMESPACE
echo ""
kubectl get services -n $NAMESPACE
echo ""
kubectl get ingress -n $NAMESPACE

# Show useful commands
echo -e "${GREEN}Useful commands:${NC}"
echo -e "${YELLOW}View logs:${NC}"
echo "  kubectl logs -f deployment/signaling-deployment -n $NAMESPACE"
echo "  kubectl logs -f deployment/redis-deployment -n $NAMESPACE"
echo "  kubectl logs -f deployment/coturn-deployment -n $NAMESPACE"
echo ""
echo -e "${YELLOW}Port forward for local access:${NC}"
echo "  kubectl port-forward service/signaling-service 8080:8080 -n $NAMESPACE"
echo ""
echo -e "${YELLOW}Scale deployment:${NC}"
echo "  kubectl scale deployment signaling-deployment --replicas=5 -n $NAMESPACE"
echo ""
echo -e "${YELLOW}Check HPA status:${NC}"
echo "  kubectl get hpa -n $NAMESPACE"

echo -e "${GREEN}Deployment script completed!${NC}"
