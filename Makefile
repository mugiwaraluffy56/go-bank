.PHONY: all build run test lint clean docker-build docker-up docker-down migrate-up migrate-down help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=gobank
BINARY_PATH=./cmd/api

# Docker
DOCKER_COMPOSE=docker compose

# Database
DB_URL=postgres://postgres:postgres@localhost:5432/gobank?sslmode=disable

all: lint test build

## build: Build the application binary
build:
	@echo "Building..."
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-w -s" -o $(BINARY_NAME) $(BINARY_PATH)

## run: Run the application
run:
	@echo "Running..."
	$(GOCMD) run $(BINARY_PATH)/main.go

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run --timeout=5m

## clean: Clean build files
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest -f deployments/docker/Dockerfile .

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

## docker-down: Stop all services
docker-down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

## docker-logs: View Docker logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## docker-ps: List running containers
docker-ps:
	$(DOCKER_COMPOSE) ps

## migrate-up: Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	migrate -path migrations -database "$(DB_URL)" up

## migrate-down: Run database migrations down
migrate-down:
	@echo "Running migrations down..."
	migrate -path migrations -database "$(DB_URL)" down

## migrate-create: Create a new migration (usage: make migrate-create name=migration_name)
migrate-create:
	@echo "Creating migration..."
	migrate create -ext sql -dir migrations -seq $(name)

## install-tools: Install development tools
install-tools:
	@echo "Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## swagger: Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	swag init -g cmd/api/main.go -o api/docs

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
