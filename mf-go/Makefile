.PHONY: build run test lint migrate migrate-down docker-up docker-down clean help

APP_NAME := masterfabric
BUILD_DIR := bin
MAIN_PATH := ./cmd/server
MIGRATION_DIR := internal/infrastructure/postgres/migrations
DB_DSN ?= postgres://masterfabric:masterfabric@localhost:5432/masterfabric?sslmode=disable

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## build: Build the application binary
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

## run: Run the application
run:
	@echo "Running $(APP_NAME)..."
	go run $(MAIN_PATH)

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v -race -count=1 ./...

## test-cover: Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## migrate: Run database migrations up
migrate:
	@echo "Running migrations..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" up

## migrate-down: Rollback last migration
migrate-down:
	@echo "Rolling back last migration..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" down

## migrate-status: Show migration status
migrate-status:
	@echo "Migration status..."
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" status

## docker-up: Start Docker Compose services
docker-up:
	@echo "Starting Docker services..."
	docker compose -f deployments/docker-compose.yml up -d

## docker-down: Stop Docker Compose services
docker-down:
	@echo "Stopping Docker services..."
	docker compose -f deployments/docker-compose.yml down

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) coverage.out coverage.html tmp

## seed: Seed the database with sample data
seed:
	@echo "Seeding database..."
	go run ./scripts/seed.go
