# Signed Blob Storage Service - Makefile
# 
# This Makefile provides convenient targets for building, testing, and running
# the Signed Blob Storage Service. All targets use British English conventions
# and follow standard GNU Make practices.

.PHONY: help check generate-proto build run build-and-run generate-keys unit-test e2e-test tests test build-client clean

.DEFAULT_GOAL := help

## Display this help message with available targets
help: ## Show this help message
	@echo "Signed Blob Storage Service - Available Make Targets"
	@echo ""
	@echo "Build & Development:"
	@awk 'BEGIN {FS = ":.*##"; category=""} /^## / {category=substr($$0,4); printf "\n\033[1m%s\033[0m\n", category; next} /^[a-zA-Z_-]+:.*?##/ {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Usage examples:"
	@echo "   make check            # Verify all required tools are installed"
	@echo "   make build-and-run    # Build and start the complete service"
	@echo "   make test             # Run all tests (unit + integration)"
	@echo "   make build-client     # Build CLI client only"

## Prerequisites and environment checking
check: ## Verify all required tools and dependencies are installed
	@echo "Checking required tools and dependencies..."
	@echo ""
	@echo "Essential tools:"
	@command -v go >/dev/null 2>&1 && echo "  [OK] Go $(shell go version | cut -d' ' -f3)" || { echo "  [ERROR] Go not found - install Go 1.24+"; exit 1; }
	@command -v docker >/dev/null 2>&1 && echo "  [OK] Docker $(shell docker --version | cut -d' ' -f3 | tr -d ',')" || { echo "  [ERROR] Docker not found - install Docker"; exit 1; }
	@echo ""
	@echo "Development tools:"
	@command -v buf >/dev/null 2>&1 && echo "  [OK] Buf $(shell buf --version 2>/dev/null || echo 'version unknown')" || echo "  [WARN] Buf not found - run ./install-buf-tools-linux.sh"
	@command -v grpcurl >/dev/null 2>&1 && echo "  [OK] grpcurl $(shell grpcurl --version 2>/dev/null | head -n1 || echo 'version unknown')" || echo "  [WARN] grpcurl not found - run ./install-grpcurl-linux.sh (optional for testing)"
	@echo ""
	@echo "Project files:"
	@test -f buf.yaml && echo "  [OK] buf.yaml configuration found" || { echo "  [ERROR] buf.yaml not found"; exit 1; }
	@test -f buf.gen.yaml && echo "  [OK] buf.gen.yaml configuration found" || { echo "  [ERROR] buf.gen.yaml not found"; exit 1; }
	@test -f docker-compose.yaml && echo "  [OK] docker-compose.yaml found" || { echo "  [ERROR] docker-compose.yaml not found"; exit 1; }
	@test -f scripts/generate-rsa-keys.sh && echo "  [OK] RSA key generation script found" || { echo "  [ERROR] scripts/generate-rsa-keys.sh not found"; exit 1; }
	@echo ""
	@echo "Environment check completed successfully!"
	@echo "Run 'make build-and-run' to start the service"

## Protocol Buffer generation and project building
generate-proto: ## Generate Go code from Protocol Buffer definitions using Buf
	buf generate --template buf.gen.yaml

build: generate-proto generate-keys ## Build Docker images with generated code and RSA keys
	go mod vendor && docker compose build

run: build ## Start the service stack (requires prior build)
	docker compose up

build-and-run: build ## Build and start the complete service in one command
	go mod vendor && docker compose up --build

## Cryptographic key generation
generate-keys: ## Generate RSA-PSS private/public key pair for digital signatures
	./scripts/generate-rsa-keys.sh

## Testing and quality assurance
unit-test: ## Run unit tests for all Go packages
	go test -v ./... -count=1

e2e-test: ## Run end-to-end integration tests using testcontainers
	go test -v -tags=e2e ./e2e -count=1

tests: unit-test e2e-test ## Run complete test suite (unit + integration tests)
test: unit-test e2e-test ## Alias for 'tests' target

## Client application building
build-client: ## Build the command-line client application
	go build -o client ./cmd/client/

## Maintenance and cleanup
clean: ## Remove generated files and Docker resources
	docker compose down --volumes --remove-orphans
	docker system prune -f
	rm -f client
	rm -f private_key.pem public_key.pem