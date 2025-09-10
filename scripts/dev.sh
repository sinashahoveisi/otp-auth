#!/bin/sh

set -e

echo "🚀 Starting OTP Auth Service Development Environment"
echo "================================================="

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

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

print_status "Docker is running"

# Check if .env file exists
if [ ! -f .env ]; then
    print_warning ".env file not found. Creating default .env file..."
    cat > .env << EOF
POSTGRES_HOST=db
POSTGRES_USER=otp_auth
POSTGRES_PASSWORD=otp_auth
POSTGRES_DB=otp_auth
APP_PORT=8080

# JWT Configuration
JWT_SECRET=your-super-secret-key-change-in-production-make-it-long-and-random
JWT_EXPIRATION_TIME=24h

# OTP Configuration
OTP_LENGTH=6
OTP_EXPIRATION_TIME=2m

# Rate Limiting Configuration
RATE_LIMIT_MAX_REQUESTS=3
RATE_LIMIT_WINDOW_DURATION=10m

# Swagger Configuration
SWAGGER_ENABLED=true

# Logger Configuration
LOGGER_LEVEL=debug
LOGGER_MODE=development
EOF
    print_status "Created .env file with default values"
fi

# Function to cleanup on exit
cleanup() {
    print_info "Shutting down development environment..."
    docker-compose stop dev db
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup INT TERM

print_info "Building development environment..."
docker-compose build dev

print_info "Starting database..."
docker-compose up -d db

# Wait for database to be ready
print_info "Waiting for database to be ready..."
sleep 5

print_status "Starting development server with hot reload..."
print_info "The application will be available at: http://localhost:${APP_PORT:-8080}"
print_info "Swagger UI will be available at: http://localhost:${APP_PORT:-8080}/swagger/index.html"
print_info "Health check available at: http://localhost:${APP_PORT:-8080}/health"
print_info "Press Ctrl+C to stop the development server"

echo ""
echo "${BLUE}📝 Development Features:${NC}"
echo "  • Hot reload on code changes"
echo "  • Debug logging enabled"
echo "  • Swagger documentation enabled"
echo "  • PostgreSQL database included"
echo "  • Source code mounted for live editing"
echo "  • OTP codes printed to console"
echo ""

echo "${BLUE}🧪 API Testing:${NC}"
echo "  • Send OTP: POST /api/v1/otp/send"
echo "  • Verify OTP: POST /api/v1/otp/verify"
echo "  • List Users: GET /api/v1/users (requires JWT)"
echo "  • Get User: GET /api/v1/users/{id} (requires JWT)"
echo ""

# Start the development server
docker-compose up dev
