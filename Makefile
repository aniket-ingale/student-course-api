# Load .env.local if present so DATABASE_URL etc. are available to targets.
ifneq (,$(wildcard .env.local))
include .env.local
export
endif

.PHONY: build run test test-integration migrate-up migrate-down migrate-create fmt vet compose-up compose-down tidy

build: ## Build all packages
	go build ./...

run: ## Run the API server locally
	go run ./cmd/server

test: ## Run unit tests
	go test ./...

test-integration: ## Run integration tests (requires Docker for testcontainers)
	go test -tags=integration ./...

migrate-up: ## Apply all migrations
	migrate -database "$(DATABASE_URL)" -path migrations up

migrate-down: ## Roll back the most recent migration
	migrate -database "$(DATABASE_URL)" -path migrations down 1

migrate-create: ## Create a new migration pair: make migrate-create name=add_x
	migrate create -ext sql -dir migrations -seq $(name)

fmt: ## Format code
	gofmt -l -w .

vet: ## Run static analysis
	go vet ./...

compose-up: ## Build and start the full stack
	docker compose up --build

compose-down: ## Stop the stack and remove volumes
	docker compose down -v

tidy: ## Sync dependencies
	go mod tidy
