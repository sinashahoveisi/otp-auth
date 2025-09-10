# OTP Authentication Service

A robust, production-ready backend service implemented in Go that provides OTP-based authentication and user management features. This service follows clean architecture principles with Redis-powered rate limiting, JWT session management, comprehensive testing, and CI/CD integration.

## 🚀 Features

- **Session-Token OTP Authentication**: Enhanced security with session tokens for OTP verification
- **Redis-Powered Rate Limiting**: High-performance rate limiting with configurable TTL and automatic cleanup
- **JWT Session Management**: Redis-backed JWT token storage with logout functionality and token revocation
- **User Management**: Complete CRUD operations with pagination and search capabilities
- **Phone Number Validation**: Robust validation to prevent invalid numbers (e.g., "salamsalam")
- **Database Migrations**: Automated PostgreSQL schema management with optimized migrations
- **Docker Support**: Full containerization with Redis, PostgreSQL, and multi-environment support
- **API Documentation**: Comprehensive Swagger/OpenAPI with proper authentication configuration
- **Comprehensive Testing**: Unit tests, integration tests, and end-to-end scenario testing
- **CI/CD Integration**: GitHub Actions and GitLab CI pipelines with security scanning
- **Monitoring**: Built-in health checks, structured logging, and performance metrics

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
├── cmd/                    # Application entry point
├── config/                 # Configuration management
├── entity/                 # Domain entities and DTOs
├── repository/            # Data access layer
├── service/               # Business logic layer
├── controller/            # HTTP request handlers
├── handler/               # Route definitions and middleware
├── validator/             # Request validation
├── migrations/           # Database migrations
├── pkg/logger/           # Logging utilities
├── test/                 # Integration tests
└── docs/                 # Swagger documentation
```

## 📋 Requirements

- Go 1.23+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (recommended)

## 🏢 CI/CD & Quality Assurance

This project includes professional CI/CD pipelines:

- **GitHub Actions**: Automated testing, building, security scanning, and publishing
- **GitLab CI**: Complete pipeline with PostgreSQL services and multi-stage testing
- **Security Scanning**: Trivy container scanning and Go vulnerability checks
- **Quality Gates**: Code coverage reporting and lint checks
- **Multi-Architecture Builds**: Support for AMD64 and ARM64 platforms

## 🚀 Quick Start

### Using Docker (Recommended)

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd otp-auth
   ```

2. **Start the service:**
   ```bash
   docker-compose up -d
   ```

3. **The service will be available at:**
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger/index.html
   - Health Check: http://localhost:8080/health

### Local Development

1. **Install dependencies:**
   ```bash
   make deps
   make install-tools
   ```

2. **Set up environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start PostgreSQL and Redis:**
   ```bash
   docker-compose up -d db redis
   ```

4. **Run the service:**
   ```bash
   make dev
   ```

## 🔧 Configuration

The service is configured via environment variables. All variables are properly standardized and cleaned:

### Application Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_SERVER_PORT` | 8080 | Application server port |
| `SWAGGER_ENABLED` | true | Enable/disable Swagger documentation |
| `LOGGER_LEVEL` | info | Logging level (debug, info, warn, error) |
| `LOGGER_MODE` | production | Logging mode (development, production) |

### Database Configuration (PostgreSQL)
| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_HOST` | db | PostgreSQL host |
| `DATABASE_PORT` | 5432 | PostgreSQL port |
| `DATABASE_USER` | otp_auth | Database username |
| `DATABASE_PASSWORD` | otp_auth | Database password |
| `DATABASE_NAME` | otp_auth | Database name |
| `DATABASE_SSL_MODE` | disable | SSL mode for database connection |

### Redis Configuration (Rate Limiting & Token Storage)
| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | redis | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `REDIS_PASSWORD` | "" | Redis password (optional) |
| `REDIS_DB` | 0 | Redis database number |

### JWT Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | (required) | JWT signing secret |
| `JWT_EXPIRATION_TIME` | 24h | JWT token expiration |

### OTP Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `OTP_LENGTH` | 6 | OTP code length |
| `OTP_EXPIRATION_TIME` | 2m | OTP expiration time |

### Rate Limiting Configuration
| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_MAX_REQUESTS` | 3 | Max OTP requests per window |
| `RATE_LIMIT_WINDOW_DURATION` | 10m | Rate limit window duration |

## 🔌 API Endpoints

### Public Endpoints

#### Send OTP
```http
POST /api/v1/otp/send
Content-Type: application/json

{
  "phone_number": "+1234567890"
}
```

**Response:**
```json
{
  "message": "OTP sent successfully",
  "phone_number": "+1234567890",
  "token": "session_token_for_verification", 
  "expires_at": "2024-01-15T12:02:00Z"
}
```

#### Verify OTP (Enhanced with Session Token)
```http
POST /api/v1/otp/verify
Content-Type: application/json

{
  "token": "session_token_from_send_response",
  "code": "123456"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone_number": "+1234567890",
    "registered_at": "2024-01-15T12:00:00Z",
    "last_login_at": "2024-01-15T12:00:00Z",
    "is_active": true
  },
  "expires_at": "2024-01-16T12:00:00Z",
  "message": "Authentication successful"
}
```

### Protected Endpoints (Require JWT)

Add the JWT token to the Authorization header:
```
Authorization: Bearer your_jwt_token_here
```

#### Get User by ID
```http
GET /api/v1/users/{id}
Authorization: Bearer your_jwt_token_here
```

#### List Users with Pagination
```http
GET /api/v1/users?page=1&page_size=20&search=123
Authorization: Bearer your_jwt_token_here
```

#### Logout (Token Revocation)
```http
POST /api/v1/auth/logout
Authorization: Bearer your_jwt_token_here
```

**Response:**
```json
{
  "message": "Successfully logged out"
}
```

**Response:**
```json
{
  "users": [
    {
      "id": 1,
      "phone_number": "+1234567890",
      "registered_at": "2024-01-15T12:00:00Z",
      "last_login_at": "2024-01-15T12:00:00Z",
      "is_active": true
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

## 🧪 Testing

This project includes a comprehensive test suite with multiple testing levels:

### Run all tests:
```bash
make test
```

### Run unit tests only:
```bash
make unit-test
```

### Run scenario tests (End-to-End):
```bash
make scenario-test
```

### Run with coverage:
```bash
make test-coverage
```

### Test Infrastructure Features:
- **Unit Tests**: Comprehensive controller, service, and repository testing
- **Integration Tests**: Database and Redis integration verification
- **Scenario Tests**: End-to-end testing with full environment setup
- **Mock Testing**: Proper mocking for isolated unit testing
- **Coverage Reporting**: HTML and terminal coverage reports
- **Docker Test Environment**: Isolated testing with test databases
- **Validation Testing**: Phone number and request validation testing

## 🛠️ Development

### Available Make commands:
```bash
make help           # Show all available commands
make dev            # Start development server with hot reload
make dev-stop       # Stop development environment
make dev-logs       # View development logs
make build          # Build the application
make test           # Run all tests
make unit-test      # Run unit tests only
make scenario-test  # Run end-to-end scenario tests
make test-coverage  # Run tests with coverage report
make clean          # Clean build artifacts
make install-tools  # Install development tools (air, swag, etc.)
make swagger        # Generate Swagger documentation
```

### Development workflow:
1. Start development environment: `make docker-dev`
2. Make changes to the code
3. Tests run automatically on file changes
4. API documentation is auto-generated

## 💾 Database

### Database Architecture

#### PostgreSQL (Primary Database)
We chose PostgreSQL for persistent data storage:

1. **ACID Compliance**: Ensures data consistency for critical authentication operations
2. **Concurrent Performance**: Handles multiple OTP requests efficiently with proper locking
3. **Indexing**: Excellent support for phone number and timestamp-based queries
4. **JSON Support**: Future flexibility for storing additional user metadata
5. **Reliability**: Battle-tested for production workloads
6. **Open Source**: No licensing costs and strong community support

#### Redis (Caching & Session Layer)
Redis provides high-performance caching and session management:

1. **Rate Limiting**: Fast, memory-based rate limiting with configurable TTL
2. **JWT Session Storage**: Secure token storage with automatic expiration
3. **High Performance**: Sub-millisecond response times for authentication checks
4. **Scalability**: Horizontal scaling support for high-traffic scenarios
5. **TTL Support**: Automatic cleanup of expired tokens and rate limit entries

### Schema Overview

**PostgreSQL Tables:**
- **users**: Stores user information and registration data
- **otps**: Manages OTP codes with session tokens and expiration tracking
- **schema_migrations**: Tracks applied database migrations

**Redis Data Structures:**
- **Rate Limits**: `rate_limit:{phone_number}` with TTL-based expiration
- **JWT Tokens**: `token:{user_id}:{token_hash}` for session management

### Migrations

Migrations run automatically on startup. To run manually:
```bash
make db-migrate
```

## 📈 Monitoring & Logging

### Health Check
```http
GET /health
```

### Structured Logging
The service uses structured JSON logging with configurable levels:
- `debug`: Detailed information for debugging
- `info`: General application flow
- `warn`: Warning conditions
- `error`: Error conditions

### Metrics
- Request/response logging
- Database query performance
- Rate limiting statistics
- OTP generation and verification rates

## 🔐 Security Considerations

1. **Enhanced Session Security**: OTP verification uses session tokens instead of phone numbers
2. **Redis-Powered Rate Limiting**: High-performance rate limiting with automatic TTL cleanup
3. **JWT Token Management**: Redis-backed token storage with logout and revocation capabilities
4. **Phone Number Validation**: Robust validation prevents invalid inputs (e.g., "salamsalam")
5. **JWT Expiration**: Tokens expire after 24 hours with Redis-enforced cleanup
6. **OTP Expiration**: OTP codes expire after 2 minutes with database cleanup
7. **Input Validation**: Comprehensive validation using structured validation rules
8. **SQL Injection Protection**: Uses parameterized queries exclusively
9. **CORS Configuration**: Configurable CORS settings for web clients
10. **Token Hashing**: JWT tokens are hashed (SHA256) before storage in Redis
11. **Session Isolation**: Each session token is unique and tied to specific OTP requests

## 🚀 Deployment

### Production Deployment

1. **Build and push Docker image:**
   ```bash
   make docker-build
   docker tag otp-auth:latest your-registry/otp-auth:latest
   docker push your-registry/otp-auth:latest
   ```

2. **Deploy with Docker Compose:**
   ```bash
   docker-compose -f docker-compose.yml up -d
   ```

### Environment Variables for Production

Ensure these are properly configured:
- `JWT_SECRET`: Use a strong, randomly generated secret (32+ characters)
- `DATABASE_PASSWORD`: Use a strong database password
- `REDIS_PASSWORD`: Use a strong Redis password for production
- `LOGGER_MODE`: Set to "production" 
- `LOGGER_LEVEL`: Set to "info" or "warn" for production
- `SWAGGER_ENABLED`: Set to "false" in production

### Production Deployment with Redis

The service requires both PostgreSQL and Redis in production:

```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  app:
    image: otp-auth:latest
    environment:
      - DATABASE_HOST=postgres
      - REDIS_HOST=redis
      - LOGGER_MODE=production
      - SWAGGER_ENABLED=false
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ${DATABASE_NAME}
      POSTGRES_USER: ${DATABASE_USER}
      POSTGRES_PASSWORD: ${DATABASE_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
```

## 🤝 API Examples

### Complete Authentication Flow

1. **Send OTP:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/otp/send \
     -H "Content-Type: application/json" \
     -d '{"phone_number": "+1234567890"}'
   ```

2. **Check console for OTP code** (printed to stdout)

3. **Verify OTP (using session token):**
   ```bash
   curl -X POST http://localhost:8080/api/v1/otp/verify \
     -H "Content-Type: application/json" \
     -d '{"token": "SESSION_TOKEN_FROM_STEP_1", "code": "123456"}'
   ```

4. **Use JWT token for protected endpoints:**
   ```bash
   curl -X GET http://localhost:8080/api/v1/users/1 \
     -H "Authorization: Bearer YOUR_JWT_TOKEN"
   ```

5. **Logout (revoke JWT token):**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/logout \
     -H "Authorization: Bearer YOUR_JWT_TOKEN"
   ```

## 📚 Additional Documentation

- [API Documentation (Swagger)](http://localhost:8080/swagger/index.html)
- [Database Schema](./migrations/)
- [Testing Guide](./test/)
- [Contributing Guidelines](./CONTRIBUTING.md)

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed:**
   - Ensure PostgreSQL is running: `docker-compose up -d db`
   - Check database credentials in `.env` (use DATABASE_* variables)
   - Verify network connectivity

2. **Redis Connection Failed:**
   - Ensure Redis is running: `docker-compose up -d redis`
   - Check Redis configuration in `.env` (REDIS_* variables)
   - Verify Redis authentication if password is set

3. **OTP Not Received:**
   - Check console output (OTPs are printed there)
   - Verify phone number format (+1234567890 format required)
   - Ensure phone number passes validation (no "salamsalam" type inputs)

4. **Rate Limiting:**
   - Wait for the rate limit window to reset (10 minutes by default)
   - Check Redis for rate limit entries: `KEYS rate_limit:*`
   - Rate limits automatically expire with Redis TTL

5. **JWT Token Invalid:**
   - Ensure token is not expired (24h default)
   - Check JWT secret configuration (JWT_SECRET)
   - Verify token format in Authorization header: `Bearer <token>`
   - Check if token was revoked (logout invalidates tokens in Redis)

6. **Session Token Issues:**
   - Use session token from `/otp/send` response in `/otp/verify`
   - Session tokens expire with OTP (2 minutes default)
   - Each OTP send generates a new session token

### Getting Help

For issues and questions:
1. Check the troubleshooting section above
2. Review the logs for error messages
3. Open an issue on GitHub
4. Contact the development team

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Note**: This service prints OTP codes to the console instead of sending SMS as per requirements. In a production environment, integrate with an SMS service provider.
