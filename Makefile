# Makefile for MarketFlow Application

.PHONY: help build run test clean docker-up docker-down lint vet fmt deps mod-tidy install-tools

# Default target
help:
	@echo "MarketFlow - Cryptocurrency Market Data Processing Application"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build the application binary"
	@echo "  run           - Run the application"
	@echo "  test          - Run all tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-up     - Start PostgreSQL and Redis containers"
	@echo "  docker-down   - Stop and remove containers"
	@echo "  lint          - Run golangci-lint"
	@echo "  vet           - Run go vet"
	@echo "  fmt           - Format code with go fmt"
	@echo "  deps          - Download dependencies"
	@echo "  mod-tidy      - Tidy go modules"
	@echo "  install-tools - Install development tools"

# Build configuration
BINARY_NAME=marketflow
BUILD_DIR=build
MAIN_PACKAGE=./cmd

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-v $(LDFLAGS)

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build
	@echo "Starting MarketFlow application..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Run with custom port
run-port: build
	@echo "Starting MarketFlow application on port 9090..."
	./$(BUILD_DIR)/$(BINARY_NAME) --port 9090

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Test coverage:"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	$(GOCMD) tool cover -func=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Start PostgreSQL and Redis containers
docker-up:
	@echo "Starting PostgreSQL and Redis containers..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services are ready!"

# Stop and remove containers
docker-down:
	@echo "Stopping containers..."
	docker-compose down
	@echo "Containers stopped"

# Development database setup
db-reset: docker-down docker-up
	@echo "Database reset completed"

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOGET) -d -v ./...

# Tidy modules
mod-tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check code quality
check: fmt vet lint test
	@echo "All checks passed!"

# Load exchange Docker images (example for amd64)
load-exchanges:
	@echo "Loading exchange Docker images..."
	@if [ -f "exchange1_amd64.tar" ]; then \
		docker load -i exchange1_amd64.tar; \
	fi
	@if [ -f "exchange2_amd64.tar" ]; then \
		docker load -i exchange2_amd64.tar; \
	fi
	@if [ -f "exchange3_amd64.tar" ]; then \
		docker load -i exchange3_amd64.tar; \
	fi

# Start exchange containers
start-exchanges:
	@echo "Starting exchange containers..."
	@docker run -p 40101:40101 --name exchange1 -d exchange1 || echo "exchange1 already running"
	@docker run -p 40102:40102 --name exchange2 -d exchange2 || echo "exchange2 already running"
	@docker run -p 40103:40103 --name exchange3 -d exchange3 || echo "exchange3 already running"
	@echo "Exchange containers started"

# Stop exchange containers
stop-exchanges:
	@echo "Stopping exchange containers..."
	@docker stop exchange1 exchange2 exchange3 2>/dev/null || true
	@docker rm exchange1 exchange2 exchange3 2>/dev/null || true
	@echo "Exchange containers stopped"

# Complete setup for development
dev-setup: install-tools deps docker-up load-exchanges start-exchanges
	@echo "Development environment setup completed!"
	@echo "You can now run 'make run' to start the application"

# Production build
build-prod:
	@echo "Building production binary..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Production build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t marketflow:latest .

# Show application help
app-help: build
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Monitor logs (requires application to be running)
logs:
	@echo "Monitoring application logs..."
	@echo "Note: This assumes the application is logging to a file"
	@echo "Adjust the log file path as needed"

# Quick start (setup + run)
quick-start: dev-setup build run

# Environment variables for local development
.env:
	@echo "Creating .env file for local development..."
	@echo "PORT=8080" > .env
	@echo "DB_HOST=localhost" >> .env
	@echo "DB_PORT=5432" >> .env
	@echo "DB_USER=postgres" >> .env
	@echo "DB_PASSWORD=postgres" >> .env
	@echo "DB_NAME=market" >> .env
	@echo "REDIS_HOST=localhost" >> .env
	@echo "REDIS_PORT=6379" >> .env
	@echo "EXCHANGE1_HOST=localhost" >> .env
	@echo "EXCHANGE1_PORT=40101" >> .env
	@echo "EXCHANGE2_HOST=localhost" >> .env
	@echo "EXCHANGE2_PORT=40102" >> .env
	@echo "EXCHANGE3_HOST=localhost" >> .env
	@echo "EXCHANGE3_PORT=40103" >> .env
	@echo ".env file created"