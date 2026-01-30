# GoBank API

A production-grade banking/financial REST API built with Go, demonstrating enterprise-level development practices including clean architecture, robust security, high performance, and modern DevOps practices.

## Features

### Core Functionality
- **User Management**: Registration, authentication with JWT (access + refresh tokens), profile management
- **Account Management**: Create checking/savings accounts, multi-currency support (USD, EUR, GBP)
- **Money Transfers**: Transfer between accounts with full ACID compliance, idempotency support
- **Transaction History**: Complete audit trail with pagination

### Technical Highlights
- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Security**: JWT authentication, bcrypt password hashing, rate limiting, audit logging
- **Performance**: Redis caching, PostgreSQL connection pooling, graceful shutdown
- **DevOps**: Docker, Kubernetes manifests, GitHub Actions CI/CD, Prometheus metrics

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.22+ |
| Framework | Gin |
| Database | PostgreSQL 15+ |
| Cache | Redis 7+ |
| Authentication | JWT |
| Containerization | Docker |
| Orchestration | Kubernetes |
| CI/CD | GitHub Actions |
| Monitoring | Prometheus |

## Project Structure

```
gobank/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── domain/                 # Business entities and interfaces
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   ├── usecase/                # Business logic
│   ├── adapter/                # Interface adapters
│   │   ├── handler/            # HTTP handlers
│   │   ├── repository/         # Repository implementations
│   │   └── middleware/
│   ├── infrastructure/         # External services
│   │   ├── config/
│   │   ├── database/
│   │   ├── logger/
│   │   └── server/
│   └── pkg/                    # Shared utilities
├── migrations/                 # Database migrations
├── deployments/
│   ├── docker/
│   └── kubernetes/
└── .github/workflows/          # CI/CD pipelines
```

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker)
- Redis 7+ (or use Docker)
- Make (optional)

### Quick Start with Docker

```bash
# Clone the repository
git clone https://github.com/yourusername/gobank.git
cd gobank

# Start all services
docker compose up -d

# The API will be available at http://localhost:8080
```

### Local Development

```bash
# Install dependencies
go mod download

# Copy environment file
cp .env.example .env

# Run database migrations
make migrate-up

# Run the application
make run
```

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login and get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Invalidate refresh token |

### Users
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/users/me` | Get current user profile |
| PUT | `/api/v1/users/me` | Update profile |

### Accounts
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/accounts` | Create new account |
| GET | `/api/v1/accounts` | List user's accounts |
| GET | `/api/v1/accounts/:id` | Get account details |
| GET | `/api/v1/accounts/:id/transactions` | Get account transactions |

### Transfers
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/transfers` | Create transfer |
| GET | `/api/v1/transfers` | List transfers |
| GET | `/api/v1/transfers/:id` | Get transfer details |

### Health & Monitoring
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/metrics` | Prometheus metrics |

## API Usage Examples

### Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123",
    "full_name": "John Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### Create Account
```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{
    "account_type": "checking",
    "currency": "USD"
  }'
```

### Create Transfer
```bash
curl -X POST http://localhost:8080/api/v1/transfers \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Idempotency-Key: unique-key-123" \
  -d '{
    "from_account_id": "uuid-from",
    "to_account_id": "uuid-to",
    "amount": "100.00"
  }'
```

## Development

### Available Make Commands

```bash
make build          # Build the application
make run            # Run the application
make test           # Run tests
make lint           # Run linter
make docker-up      # Start Docker services
make docker-down    # Stop Docker services
make migrate-up     # Run migrations
make migrate-down   # Rollback migrations
make help           # Show all commands
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Deployment

### Docker

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
docker compose up -d
```

### Kubernetes

```bash
# Apply Kubernetes manifests
kubectl apply -f deployments/kubernetes/
```

## Security Features

- **JWT Authentication**: Short-lived access tokens (15 min) with refresh token rotation
- **Password Hashing**: bcrypt with cost factor 12
- **Rate Limiting**: Redis-based sliding window rate limiting
- **Input Validation**: Comprehensive request validation
- **SQL Injection Prevention**: Parameterized queries throughout
- **Audit Logging**: All financial operations are logged
- **Security Headers**: CORS, Content-Type enforcement, XSS protection

## Architecture Decisions

1. **Clean Architecture**: Separates business logic from infrastructure concerns
2. **Repository Pattern**: Abstracts data access for easy testing and switching databases
3. **Dependency Injection**: All dependencies are injected, enabling easy mocking
4. **Database Transactions**: Financial operations use proper transaction isolation
5. **Idempotency**: Transfer operations support idempotency keys

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
