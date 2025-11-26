# Go1 Project

A modular Clean Architecture Golang project with integrated observability and event-driven architecture.

## Features

- **Clean Architecture**: Separation of concerns with Domain, Application, Infrastructure, and Presentation layers
- **Event-Driven**: Kafka integration for async messaging
- **Caching**: Redis for performance optimization
- **Observability**: 
  - Prometheus for metrics
  - Grafana for visualization
  - Jaeger for distributed tracing
  - OpenTelemetry instrumentation
- **Middleware**: CORS, Authentication (bypass mode), Metrics, Tracing
- **Error Handling**: Structured error responses with custom error codes
- **Database**: PostgreSQL with migrations
- **Development**: Hot reload with Air

## Tech Stack

- **Go**: 1.25.4
- **Web Framework**: Gin
- **Database**: PostgreSQL 15
- **Caching**: Redis 7
- **Message Queue**: Kafka (Confluent 7.3.1)
- **Observability**: Prometheus, Grafana, Jaeger, OpenTelemetry
- **Config**: Viper
- **Logging**: Zap

## Project Structure

```
.
├── cmd
│   ├── app                      # API server entry point
│   │   └── main.go
│   └── worker                   # Worker entry point
│       └── main.go
├── config                       # Configuration files
│   ├── config.yaml             # Main config file
│   ├── config.go               # Config loader
│   ├── app.go                  # App config struct
│   ├── postgres.go             # Postgres config
│   ├── redis.go                # Redis config
│   ├── kafka.go                # Kafka config
│   └── jaeger.go               # Jaeger config
├── internal
│   ├── api                      # API Service
│   │   ├── server              # HTTP server setup & middleware
│   │   └── handlers            # HTTP handlers (Presentation Layer)
│   │       └── user
│   │           ├── handler.go  # User HTTP handlers
│   │           ├── dto.go      # Request/Response DTOs
│   │           └── router.go   # Route definitions
│   ├── worker                   # Worker Service  
│   │   ├── worker.go           # Worker orchestration
│   │   └── handlers            # Message handlers
│   │       └── user_created.go # User created event handler
│   └── modules                  # Shared Business Logic
│       └── user                # User Module (Clean Architecture)
│           ├── application      # Application Layer
│           │   ├── dto.go      # Input/Output DTOs
│           │   └── usecase.go  # Business logic
│           ├── domain           # Domain Layer
│           │   ├── entity.go   # Domain entities
│           │   ├── repository.go # Repository interfaces
│           │   ├── cache.go    # Cache interfaces
│           │   └── event.go    # Event interfaces
│           ├── infrastructure   # Infrastructure Layer
│           │   ├── caching
│           │   │   ├── model   # Cache models
│           │   │   ├── mapper  # Cache mappers
│           │   │   └── redis.go # Redis implementation
│           │   ├── message_queue
│           │   │   └── kafka.go # Kafka implementation
│           │   └── repository
│           │       ├── model   # DB models
│           │      ├── mapper  # DB mappers
│           │       └── postgres.go # Postgres implementation
│           └── module.go       # Dependency injection
├── migrations                   # SQL migrations
├── pkg                         # Shared packages
│   ├── apperrors               # Custom error types
│   ├── kafka                   # Kafka client (producer & consumer)
│   ├── logger                  # Logging utilities
│   ├── metrics                 # Prometheus metrics
│   ├── middleware              # HTTP middleware
│   ├── postgres                # PostgreSQL client
│   ├── redis                   # Redis client
│   ├── response                # Standard API responses
│   ├── telemetry               # OpenTelemetry setup
│   └── utils                   # Utility functions
├── prometheus.yml              # Prometheus configuration
├── Dockerfile                  # Multi-service Dockerfile
└── docker-compose.yml          # Docker services
```

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make (optional, but recommended)

## Setup

1. **Clone the repository**

2. **Start Infrastructure**
   ```bash
   make up
   ```
   This starts all services: Postgres, Redis, Kafka, Zookeeper, Prometheus, Grafana, and Jaeger.

3. **Run Migrations**
   ```bash
   make migrate-up
   ```

4. **Create Kafka Topic** (for event publishing)
   ```bash
   docker exec go1_kafka kafka-topics --bootstrap-server localhost:9092 --create --topic user_created --partitions 1 --replication-factor 1
   ```

## Running the Application

The project now has two services:
1. **API Server** (`cmd/app`) - HTTP REST API
2. **Worker** (`cmd/worker`) - Kafka message consumer

### Development Mode (Hot Reload)

**Run API server:**
```bash
make dev
```

**Run Worker** (in separate terminal):
```bash
go run cmd/worker/main.go
```

### Production Build

**Build both services:**
```bash
go build -o bin/api ./cmd/app
go build -o bin/worker ./cmd/worker
```

**Run:**
```bash
./bin/api     # Start API server
./bin/worker  # Start worker (separate terminal)
```

### Docker Compose

**Run all services (API + Worker + Infrastructure):**
```bash
make up
```

This starts:
- API server (port 8080)
- Worker (Kafka consumer)
- Postgres, Redis, Kafka
- Prometheus, Grafana, Jaeger

## Configuration

Configuration is managed via `config/config.yaml`. You can also override settings using environment variables (e.g., `APP_PORT=9090`).

```yaml
app:
  name: go1
  port: 8080
  env: development

postgres:
  host: localhost
  port: 5432
  # ...

redis:
  addr: localhost:6379

kafka:
  brokers: localhost:9099

jaeger:
  endpoint: localhost:4318
```

## API Endpoints

### User Management
- `POST /users`: Create a user
- `GET /users`: List all users
- `GET /users/:id`: Get user by ID (cached)

### Observability
- `GET /metrics`: Prometheus metrics endpoint

## Observability & Monitoring

### Prometheus
- **URL**: http://localhost:9090
- **Metrics**: Request count, duration, active requests (CCU)

### Grafana
- **URL**: http://localhost:3000
- **Credentials**: admin/admin
- **Data Source**: Prometheus (http://prometheus:9090)

### Jaeger
- **URL**: http://localhost:16686
- **Features**: Distributed tracing for all HTTP requests

### Redpanda Console (Kafka UI)
- **URL**: http://localhost:8083
- **Features**: View topics, messages, consumer groups

## Error Handling

The application uses structured error responses:

```json
{
  "success": false,
  "error": {
    "code": 1001,
    "message": "Email already exists"
  }
}
```

**Error Codes:**
- `1001`: Email already exists
- `1002`: User not found

## Development

### Hot Reload
The project uses [Air](https://github.com/cosmtrek/air) for hot reloading in development mode.

### Make Commands
```bash
make up          # Start all Docker services
make down        # Stop all Docker services
make dev         # Run with hot reload
make build       # Build binary
make migrate-up  # Run migrations
make migrate-down # Rollback migrations
```

## Architecture Principles

1. **Dependency Inversion**: All layers depend on abstractions (interfaces) defined in the domain layer
2. **Separation of Concerns**: Each layer has a single, well-defined responsibility
3. **Decoupling**: Infrastructure details (DB, cache, messaging) are isolated from business logic
4. **Testability**: Clear boundaries make unit testing straightforward
5. **Service Separation**: API and Worker services share business logic through modules but deploy independently

## Service Architecture

**API Service** (`internal/api/`)
- HTTP server with Gin
- REST endpoints
- Prometheus metrics
- OpenTelemetry tracing

**Worker Service** (`internal/worker/`)
- Kafka consumer
- Event-driven processing
- Shares domain logic with API

**Shared Modules** (`internal/modules/`)
- Business logic (application layer)
- Domain entities and interfaces
- Infrastructure implementations
- Used by both API and Worker
