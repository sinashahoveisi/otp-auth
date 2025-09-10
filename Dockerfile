# Build stage
FROM docker.arvancloud.ir/golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk --no-cache add git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Install swag CLI tool and generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/main.go -o docs
RUN sed -i '/LeftDelim/d; /RightDelim/d' docs/docs.go

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o otp-auth ./cmd/main.go

# Final stage
FROM docker.arvancloud.ir/alpine:3.18

# Install ca-certificates and PostgreSQL client tools for testing
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/otp-auth .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy scripts
COPY --from=builder /app/scripts ./scripts

# Make scripts executable
RUN chmod +x ./scripts/*.sh 2>/dev/null || true

EXPOSE 8080

CMD ["./otp-auth"]
