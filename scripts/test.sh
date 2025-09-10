#!/bin/sh

set -e

echo "🧪 Running Tests for OTP Auth Service"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo "${GREEN}✅ $1${NC}"
}

print_info() {
    echo "${BLUE}ℹ️  $1${NC}"
}

print_warning() {
    echo "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo "${RED}❌ $1${NC}"
}

# Check if we're in Docker
if [ -f /.dockerenv ]; then
    print_info "Running in Docker container"
    DOCKER_MODE=true
else
    print_info "Running locally"
    DOCKER_MODE=false
fi

# Generate swagger docs first if we have the tools
print_info "Generating Swagger documentation..."
if command -v swag >/dev/null 2>&1; then
    swag init -g cmd/main.go -o docs
    /usr/bin/sed -i '' '/LeftDelim/d; /RightDelim/d' docs/docs.go
    print_status "Swagger docs generated"
elif [ "$DOCKER_MODE" = true ]; then
    # In Docker, swag should be available
    swag init -g cmd/main.go -o docs
    /usr/bin/sed -i '' '/LeftDelim/d; /RightDelim/d' docs/docs.go
    print_status "Swagger docs generated"
else
    print_warning "swag command not found, skipping swagger generation"
fi

# Run different types of tests
print_info "Running unit tests..."
go test -v -race -short ./service/... ./repository/... ./controller/... ./validator/...

print_info "Running integration tests..."
go test -v -race ./test/...

print_info "Running all tests with coverage..."
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
if [ -f coverage.out ]; then
    print_info "Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    
    # Show coverage summary
    print_info "Coverage Summary:"
    go tool cover -func=coverage.out | tail -1
    
    print_status "Coverage report generated: coverage.html"
else
    print_warning "No coverage file generated"
fi

# Run go vet
print_info "Running go vet..."
go vet ./...

# Run go fmt check
print_info "Checking code formatting..."
if [ -n "$(go fmt ./...)" ]; then
    print_error "Code is not properly formatted. Run 'go fmt ./...' to fix."
    exit 1
else
    print_status "Code formatting is correct"
fi

print_status "All tests completed successfully!"

echo ""
echo "${BLUE}📊 Test Results Summary:${NC}"
echo "  • Unit tests: ✅"
echo "  • Integration tests: ✅"
echo "  • Code coverage: See coverage.html"
echo "  • Code formatting: ✅"
echo "  • Go vet: ✅"
