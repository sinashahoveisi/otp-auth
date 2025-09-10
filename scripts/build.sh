#!/bin/bash
set -e

echo "Building OTP Auth Service..."

# Generate swagger docs
echo "Generating Swagger documentation..."
swag init -g cmd/main.go -o docs

# Build the application
echo "Building Go binary..."
go build -o otp-auth ./cmd/main.go

echo "Build completed successfully!"
echo "Binary: ./otp-auth"
