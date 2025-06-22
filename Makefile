.PHONY: help build run test clean docker-build docker-run docker-compose-up docker-compose-down k8s-deploy k8s-delete dev-setup

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev-setup: ## Set up development environment
	@./scripts/dev-setup.sh

build: ## Build the Go application
	@echo "Building signaling server..."
	@go build -o bin/signaling ./cmd/signaling

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f signaling
	@docker system prune -f

# Docker
docker-build: ## Build Docker images
	@./scripts/build.sh

docker-run: docker-build ## Run with Docker (single container)
	@echo "Running signaling server with Docker..."
	@docker run -p 8080:8080 --rm signaling-server:latest

docker-compose-up: ## Start all services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@cd deployments/docker-compose && docker compose up --build

docker-compose-down: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	@cd deployments/docker-compose && docker compose down

docker-compose-logs: ## View Docker Compose logs
	@cd deployments/docker-compose && docker compose logs -f

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	@./scripts/deploy.sh

k8s-delete: ## Delete Kubernetes deployment
	@echo "Deleting Kubernetes deployment..."
	@kubectl delete namespace webrtc-signaling --ignore-not-found=true

k8s-logs: ## View Kubernetes logs
	@kubectl logs -f deployment/signaling-deployment -n webrtc-signaling

k8s-status: ## Check Kubernetes deployment status
	@echo "Checking Kubernetes deployment status..."
	@kubectl get pods -n webrtc-signaling
	@echo ""
	@kubectl get services -n webrtc-signaling
	@echo ""
	@kubectl get ingress -n webrtc-signaling

k8s-port-forward: ## Port forward Kubernetes service to localhost
	@echo "Port forwarding signaling service to localhost:8080..."
	@kubectl port-forward service/signaling-service 8080:8080 -n webrtc-signaling

# Development helpers
fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Minikube helpers
minikube-start: ## Start minikube
	@echo "Starting minikube..."
	@minikube start
	@minikube addons enable ingress

minikube-stop: ## Stop minikube
	@echo "Stopping minikube..."
	@minikube stop

minikube-build: ## Build images in minikube
	@echo "Building images in minikube..."
	@eval $$(minikube docker-env) && ./scripts/build.sh

minikube-deploy: minikube-build ## Deploy to minikube
	@./scripts/deploy.sh

minikube-service: ## Open minikube service
	@minikube service signaling-service -n webrtc-signaling

# Production helpers
prod-build: ## Build for production
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o bin/signaling ./cmd/signaling

# Monitoring
health-check: ## Check application health
	@echo "Checking application health..."
	@curl -f http://localhost:8080/health || echo "Health check failed"

ready-check: ## Check application readiness
	@echo "Checking application readiness..."
	@curl -f http://localhost:8080/ready || echo "Readiness check failed"

# Database
redis-cli: ## Connect to Redis CLI (Docker Compose)
	@cd deployments/docker-compose && docker compose exec redis redis-cli

# All-in-one commands
dev: docker-compose-up ## Start development environment

stop: docker-compose-down ## Stop development environment

restart: docker-compose-down docker-compose-up ## Restart development environment

full-deploy: docker-build k8s-deploy ## Build and deploy to Kubernetes

# Variables
IMAGE_TAG ?= latest
REGISTRY ?= 
KUBECTL_CONTEXT ?= 

.DEFAULT_GOAL := help
