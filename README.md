# Go1 Project

A modular Clean Architecture Golang project with integrated observability and event-driven architecture.

## Features

- **Clean Architecture**: Separation of concerns with Domain, Application, Infrastructure, and Presentation layers
- **Event-Driven**: Kafka integration for async messaging
- **Workflow Engine**: Temporal for durable execution and distributed transactions (Saga Pattern)
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
- **Workflow Engine**: Temporal
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
│   └── shared                  # Shared Business Logic
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
- Temporal Server (port 7233)
- Temporal UI (port 8088)
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
make dev
```

**Terminal 2 - Start Worker:**
```bash
make dev-worker
```


### Step 4: Monitor

- **API Metrics**: http://localhost:8080/metrics
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger Tracing**: http://localhost:16686
- **Kafka Console UI**: http://localhost:8084
- **Temporal UI**: http://localhost:8088
- **Debezium API**: http://localhost:8083 (Kafka Connect REST API)

## API Endpoints

### Order Management
- `POST /orders`: Create an order
- `GET /orders/:id`: Get order by ID

### Observability
- `GET /metrics`: Prometheus metrics endpoint

### Make Commands
```bash
# Infrastructure
make up           # Start all Docker services (Postgres, Redis, Kafka, etc.)
make down         # Stop all Docker services

# API Server
make dev          # Run API server with hot reload

# Worker
make dev-worker   # Run worker

# Database
make migrate-up   # Run migrations
make migrate-down # Rollback migrations
```

# Validating
## Workflow Testing
We have a script to test the full order workflow:
```bash
go run scripts/test_flow/main.go
```

