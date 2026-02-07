# Ecommerce Platform Root Makefile

.PHONY: build-all test lint tidy fmt docker-build docker-build-one help

SERVICES_DIR := services
SERVICES := $(wildcard $(SERVICES_DIR)/*)

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

fmt: ## Format all code (Go, YAML, JSON, MD)
	@echo "Formatting Go files..."
	@for /d %%i in ($(SERVICES_DIR)\*) do ( \
		if exist "%%i\go.mod" ( \
			echo Formatting %%i... & \
			pushd "%%i" && go fmt ./... && popd \
		) \
	)
	@echo "Formatting YAML, JSON, Markdown with Prettier..."
	@npx -y prettier --write "**/*.{yaml,yml,json,md}"

build-all: ## Build all Go services
	@echo "Building all services..."
	@cd $(SERVICES_DIR) && go build -v ./...

test: ## Run unit tests across all services
	@echo "Running tests..."
	@cd $(SERVICES_DIR) && go test -v -cover ./...

lint: ## Run linting (requires revive or golangci-lint)
	@echo "Linting..."
	@cd $(SERVICES_DIR) && go vet ./...

tidy: ## Run go mod tidy for all services
	@echo "Tidying modules..."
	@cd $(SERVICES_DIR) && go mod tidy
	@for service in $(SERVICES); do \
		if [ -f $$service/go.mod ]; then \
			echo "Tidying $$service..."; \
			cd $$service && go mod tidy && cd ../..; \
		fi \
	done

docker-build: ## Build all Docker images (Context: Root)
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		if [ -f $$service/Dockerfile ]; then \
			NAME=$$(basename $$service); \
			echo "Building $$NAME..."; \
			docker build -t ecommerce/$$NAME:local -f $$service/Dockerfile .; \
		fi \
	done

docker-build-one: ## Build a single service image (Usage: make docker-build-one SVC=cart-service)
	@if [ -z "$(SVC)" ]; then \
		echo "SVC argument required. Example: make docker-build-one SVC=cart-service"; \
		exit 1; \
	fi
	@echo "Building $(SVC)..."
	@docker build -t ecommerce/$(SVC):local -f $(SERVICES_DIR)/$(SVC)/Dockerfile .
