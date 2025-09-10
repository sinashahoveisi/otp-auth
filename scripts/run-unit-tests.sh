#!/bin/sh

set -e

echo "🧪 Running Unit Tests (No External Dependencies)"
echo "================================================"

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

print_info "🔧 Generating Swagger documentation..."
if command -v swag >/dev/null 2>&1; then
    swag init -g cmd/main.go -o docs
    /usr/bin/sed -i '' '/LeftDelim/d; /RightDelim/d' docs/docs.go
    print_status "Swagger docs generated"
else
    print_warning "swag command not found, skipping swagger generation"
fi

print_info "🧪 Running unit tests..."
go test -v -race -short ./validator/...

if [ $? -eq 0 ]; then
    print_status "Unit tests passed!"
else
    print_error "Unit tests failed!"
    exit 1
fi

print_info "📊 Generating coverage report..."
go test -cover ./validator/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
coverage_percentage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
print_status "Coverage report generated: coverage.html"
print_info "Total coverage: ${coverage_percentage}"

print_info "🔍 Running code quality checks..."

print_info "Running go vet analysis..."
go vet ./...
if [ $? -eq 0 ]; then
    print_status "Go vet analysis passed!"
else
    print_error "Go vet analysis failed!"
    exit 1
fi

print_info "Checking code formatting..."
if [ "$(go fmt ./... | wc -l)" -eq 0 ]; then
    print_status "Code formatting is correct!"
else
    print_warning "Code formatting needs improvement. Run 'go fmt ./...' to fix."
    go fmt ./...
    print_status "Code formatting fixed!"
fi

echo ""
print_status "🎉 All Unit Tests and Quality Checks Completed Successfully!"
echo ""
echo "📋 Unit Test Summary:"
echo "  ✅ Validator Tests: Comprehensive phone number validation"
echo "  ✅ Code Quality: Formatting, linting, and best practices"
echo "  ✅ Coverage Report: Available in coverage.html"
echo "  ✅ Total Coverage: ${coverage_percentage}"
echo ""
echo "📁 Generated Files:"
echo "  - coverage.out: Coverage data"
echo "  - coverage.html: HTML coverage report"
echo "  - docs/swagger.json: API documentation"
echo "  - docs/swagger.yaml: API specification"
echo ""
echo "🎯 Unit tests provide fast feedback without external dependencies!"
