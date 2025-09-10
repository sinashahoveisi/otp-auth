APP_NAME=otp-auth

.PHONY: build run test scenario-test clean docker-build docker-run dev dev-build dev-stop dev-bg dev-logs dev-test swagger deps lint fmt install-air air-local help

# Build the application
build: swagger
	@echo "Building application..."
	@go build -o $(APP_NAME) ./cmd/main.go

# Run the application locally
run: build
	@echo "Starting OTP Auth Service..."
	@./$(APP_NAME)

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@swag init -g cmd/main.go -o docs
	@/usr/bin/sed -i '' '/LeftDelim/d; /RightDelim/d' docs/docs.go

# Run tests
test: swagger
	@echo "Running tests..."
	@./scripts/test.sh

unit-test:
	@echo "Running unit tests only..."
	@./scripts/run-unit-tests.sh

# Run scenario tests (comprehensive integration tests)
scenario-test:
	@echo "Running scenario tests..."
	@./scripts/run-scenario-tests.sh

# Run unit tests only
test-unit: swagger
	@echo "Running unit tests..."
	@go test -v -race ./service/... ./repository/... ./controller/... ./validator/...

# Run tests with coverage
test-coverage: swagger
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Build Docker image for production
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME) .

# Run with Docker Compose (production)
docker-run:
	@echo "Running with Docker Compose..."
	@docker-compose up --build

# Start development environment with hot reload
dev:
	@echo "Starting development server..."
	@./scripts/dev.sh

# Build development Docker image
dev-build:
	@echo "Building development Docker image..."
	@docker-compose build dev

# Stop development environment
dev-stop:
	@echo "Stopping development environment..."
	@docker-compose stop dev db redis

# Start development environment in background
dev-bg:
	@echo "Starting development environment in background..."
	@docker-compose up -d dev

# View development logs
dev-logs:
	@echo "Viewing development logs..."
	@docker-compose logs -f dev

# Run tests in development environment
dev-test:
	@echo "Running tests in development environment..."
	@docker-compose run --rm dev go test -v ./...

# Database
db-migrate:
	@echo "Running database migrations..."
	@docker-compose exec app ./otp-auth migrate

db-shell:
	@echo "Connecting to database..."
	@docker-compose exec db psql -U otp_auth -d otp_auth

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(APP_NAME) coverage.out coverage.html build-errors.log
	@rm -rf tmp/ bin/ docs/

# Install Air for local development (if not using Docker)
install-air:
	@echo "Installing Air for hot reload..."
	@go install github.com/cosmtrek/air@v1.49.0

# Run with Air locally (requires PostgreSQL and Redis running separately)
air-local:
	@echo "Running with Air locally..."
	@air

# Install required development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/cosmtrek/air@v1.49.0
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application binary"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run comprehensive tests"
	@echo "  scenario-test - Run scenario integration tests"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build - Build production Docker image"
	@echo "  docker-run   - Run with Docker Compose (production)"
	@echo ""
	@echo "Development commands:"
	@echo "  dev          - Start development environment with hot reload"
	@echo "  dev-build    - Build development Docker image"
	@echo "  dev-stop     - Stop development environment"
	@echo "  dev-bg       - Start development environment in background"
	@echo "  dev-logs     - View development logs"
	@echo "  dev-test     - Run tests in development environment"
	@echo ""
	@echo "Local development:"
	@echo "  install-air  - Install Air for hot reload"
	@echo "  air-local    - Run with Air locally (requires separate DB/Redis)"
	@echo ""
	@echo "Code quality:"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  swagger      - Generate Swagger documentation"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps         - Download dependencies"
	@echo "  deps-update  - Update dependencies"
	@echo "  install-tools - Install development tools"
