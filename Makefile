.PHONY: help build run test clean migrate-up migrate-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the server binary
	go build -o pemilo-server.exe ./cmd/server

run: ## Run the server
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

test-cover: ## Run tests with coverage
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f pemilo-server.exe coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run linters

dev: ## Run in development mode with hot reload (requires air)
	air

docker-build: ## Build Docker image
	docker build -t pemilo-golang:latest .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env pemilo-golang:latest

migrate-up: ## Run database migrations (manual for now)
	@echo "Running migrations..."
	@for file in migrations/*.sql; do \
		echo "Applying $$file..."; \
		psql $(DATABASE_URL) -f $$file; \
	done

setup: deps ## Initial project setup
	@echo "Setting up project..."
	@cp .env.example .env
	@echo "Please edit .env with your configuration"
	@echo "Then run: make migrate-up"
