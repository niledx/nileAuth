.PHONY: help migrate-up migrate-down migrate-version migrate-force run build test docker-build

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

migrate-up: ## Apply all pending migrations
	@cd .. && go run ./go/cmd/migrate -command up

migrate-down: ## Rollback last migration
	@cd .. && go run ./go/cmd/migrate -command down

migrate-version: ## Show current migration version
	@cd .. && go run ./go/cmd/migrate -command version

migrate-force: ## Force migration version (use: make migrate-force VERSION=1)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	@cd .. && go run ./go/cmd/migrate -command force -version $(VERSION)

run: ## Run the application
	@cd .. && go run ./go

build: ## Build the application
	@cd .. && go build -o bin/nile-auth ./go

build-migrate: ## Build the migration tool
	@cd .. && go build -o bin/migrate ./go/cmd/migrate

test: ## Run tests
	@cd .. && go test ./go/...

docker-build: ## Build Docker image
	@docker build -t nile-auth -f go/Dockerfile .

docker-up: ## Start services with docker-compose
	@docker-compose up -d

docker-down: ## Stop services with docker-compose
	@docker-compose down

docker-logs: ## View docker-compose logs
	@docker-compose logs -f

