.PHONY: build test clean docker-build tidy list-services

# Auto-discover services from go.work
SERVICES := $(shell awk '/^\t\.\// {print $$1}' go.work | sed 's|./services/||')

list-services:
	@echo "Detected services:"
	@for s in $(SERVICES); do echo " - $$s"; done

build:
	@echo "Building all services via go.work..."
	@go build ./...

test:
	@echo "Running tests via go.work..."
	@go test ./...

tidy:
	@echo "Tidying all service modules..."
	@for service in $(SERVICES); do \
		echo "Tidying $$service..."; \
		cd services/$$service && go mod tidy || exit 1; \
		cd ../..; \
	done
	@go work sync

docker-build:
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		if [ -f services/$$service/Dockerfile ]; then \
			echo "Building docker image for $$service..."; \
			docker build -t $$service:latest services/$$service || exit 1; \
		fi; \
	done
