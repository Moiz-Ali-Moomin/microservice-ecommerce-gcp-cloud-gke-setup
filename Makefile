.PHONY: build test clean docker-build

SERVICES := api-gateway auth-service feature-flag-service user-service attribution-service offer-service landing-service cart-service redirect-service conversion-webhook notification-service audit-service analytics-ingest-service reporting-service admin-backoffice-service

build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		cd services/$$service && go build ./... || exit 1; \
		cd ../..; \
	done

test:
	@echo "Running tests..."
	@go test ./services/...

docker-build:
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building docker image for $$service..."; \
		docker build -t $$service:latest -f services/$$service/Dockerfile services/$$service || exit 1; \
	done

tidy:
	@echo "Tidying modules..."
	@for service in $(SERVICES); do \
		echo "Tidying $$service..."; \
		cd services/$$service && go mod tidy; \
		cd ../..; \
	done
	@cd services/common && go mod tidy
