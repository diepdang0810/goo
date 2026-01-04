# Go1 Project

A modular Clean Architecture Golang project with integrated observability and event-driven architecture.

## Features

- **Clean Architecture**: Separation of concerns with Domain, Application, Infrastructure, and Presentation layers
- **Event-Driven**: Kafka integration for async messaging
- **CDC-Based Caching**:
  - Debezium Change Data Capture for automatic cache synchronization
  - Redis cache-first read pattern with auto-invalidation
  - Zero manual cache management
- **Observability**:
  - Prometheus for metrics
  - Grafana for visualization
  - Jaeger for distributed tracing
  - OpenTelemetry instrumentation
- **Middleware**: CORS, Authentication (bypass mode), Metrics, Tracing
- **Error Handling**: Structured error responses with custom error codes
- **Database**: PostgreSQL with migrations and logical replication
- **Development**: Hot reload with Air

## Tech Stack

- **Go**: 1.25.4
- **Web Framework**: Gin
- **Database**: PostgreSQL 15 (with CDC via logical replication)
- **Caching**: Redis 7
- **Message Queue**: Kafka (Confluent 7.3.1)
- **CDC**: Debezium 2.5 (Kafka Connect + PostgreSQL connector)
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
│   │       └── order
│   │           ├── handler.go  # Order HTTP handlers
│   │           ├── dto.go      # Request/Response DTOs
│   │           └── router.go   # Route definitions
│   ├── worker                   # Worker Service  
│   │   ├── worker.go           # Worker orchestration
│   │   └── handlers            # Message handlers
│   │       └── order_created.go # Order created event handler
│   └── modules                  # Shared Business Logic
│       └── order               # Order Module (Clean Architecture)
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

## Complete Setup & Running Guide

### Step 1: Start Infrastructure

```bash
make up
```

This starts all required services:
- PostgreSQL (port 5432) - with WAL logical replication
- Redis (port 6379)
- Kafka (port 9099)
- Zookeeper (port 2181)
- Debezium (port 8083) - CDC connector runtime (Kafka Connect API)
- Kafka Console (port 8084) - Redpanda Console UI
- Prometheus (port 9090)
- Grafana (port 3000)
- Jaeger (port 16686)

**Verify services are running:**
```bash
docker ps
```

### Step 2: Run Database Migrations

```bash
make migrate-up
```

This creates the required database tables.

### Step 2.5: Setup Debezium CDC (Optional but Recommended)

**Register the Debezium connector for automatic cache synchronization:**

```bash
./scripts/debezium/register-connector.sh
```

This enables Change Data Capture (CDC) so any PostgreSQL changes automatically sync to Redis cache.

**Verify connector status:**
```bash
./scripts/debezium/check-connector.sh
```

You should see `"state": "RUNNING"` in the output.

📚 **For detailed CDC documentation**, see: [docs/CDC_REDIS_SYNC.md](docs/CDC_REDIS_SYNC.md)

### Step 3: Start the Services

You need **TWO terminals**:

**Terminal 1 - Start API Server:**
```bash
make run
# or with hot reload (recommended for development):
make dev
```

Wait for the log:
```
✅ Connected to PostgreSQL
✅ Connected to Redis
✅ Connected to Kafka producer
Server is running on :8080
```

**Terminal 2 - Start Worker:**
```bash
make run-worker
```

Wait for the log:
```
✅ Connected to PostgreSQL
✅ Connected to Redis
✅ Connected to Kafka producer
Worker started
```

### Step 4: Test the Setup

**Test API:**
```bash
curl http://localhost:8080/orders/ORDER_ID
```

**Create an order (will publish event to Kafka):**
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": 1,
    "service_type": "delivery",
    "customer_id": "cust_123",
    "points": [{"lat": 10.0, "lng": 106.0, "type": "pickup"}]
  }'
```

**Check Worker logs** - you should see:
```
📥 Processing message
  topic: order_created
  ...
✅ Message processed successfully
```

### Step 5: Monitor

- **API Metrics**: http://localhost:8080/metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger Tracing**: http://localhost:16686
- **Kafka Console UI**: http://localhost:8084
- **Debezium API**: http://localhost:8083 (Kafka Connect REST API)

### Stopping

**Stop services gracefully:**
- Press `Ctrl+C` in Terminal 1 (API)
- Press `Ctrl+C` in Terminal 2 (Worker)

**Stop infrastructure:**
```bash
make down
```

## Running the Application

The project has **two services** that need to be run separately:
1. **API Server** (`cmd/app`) - HTTP REST API on port 8080
2. **Worker** (`cmd/worker`) - Kafka message consumer

### Quick Start (Recommended)

**Terminal 1 - API Server:**
```bash
make run
# or with hot reload:
make dev
```

**Terminal 2 - Worker:**
```bash
make run-worker
```

The API will be available at: http://localhost:8080

### Alternative Methods

#### Method 1: Using Go directly

**Terminal 1 - API Server:**
```bash
go run cmd/app/main.go
```

**Terminal 2 - Worker:**
```bash
go run cmd/worker/main.go
```

#### Method 2: Build and run binaries

**Build both services:**
```bash
make build-all
# or separately:
make build        # builds bin/app
make build-worker # builds bin/worker
```

**Run the binaries:**

**Terminal 1:**
```bash
./bin/app
```

**Terminal 2:**
```bash
./bin/worker
```

### Important Notes

- ⚠️ **Both services must run simultaneously** for full functionality
- ⚠️ **Start infrastructure first** with `make up` before running services
- ⚠️ **Run migrations** with `make migrate-up` before first use
- ✅ API server logs will show on Terminal 1
- ✅ Worker logs will show on Terminal 2

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

### Order Management
- `POST /orders`: Create an order
- `GET /orders/:id`: Get order by ID

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
- `1001`: Resource not found
- `1002`: Validation error

## Development

### Hot Reload
The project uses [Air](https://github.com/cosmtrek/air) for hot reloading in development mode.

### Make Commands
```bash
# Infrastructure
make up           # Start all Docker services (Postgres, Redis, Kafka, etc.)
make down         # Stop all Docker services

# API Server
make run          # Run API server
make dev          # Run API server with hot reload
make build        # Build API binary (bin/app)

# Worker
make run-worker   # Run worker
make build-worker # Build worker binary (bin/worker)

# Build both
make build-all    # Build both API and Worker binaries

# Database
make migrate-up   # Run migrations
make migrate-down # Rollback migrations

# Testing
make test         # Run all tests
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
