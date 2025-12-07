# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Clean Architecture Go microservices project with event-driven capabilities. It consists of two separate services:
- **API Service** (`cmd/app`): REST API server using Gin
- **Worker Service** (`cmd/worker`): Kafka consumer for async event processing

Both services share business logic through the `internal/modules/` directory, which contains domain entities, use cases, and infrastructure implementations organized by module (currently `user`).

## Development Commands

### Running Services

```bash
# Run API server with hot reload (development)
make dev

# Run API server directly
make run
# OR
go run cmd/app/main.go

# Run Worker service
go run cmd/worker/main.go

# Build both binaries
go build -o bin/app cmd/app/main.go
go build -o bin/worker cmd/worker/main.go
```

### Infrastructure

```bash
# Start all infrastructure (Postgres, Redis, Kafka, Prometheus, Grafana, Jaeger)
make up

# Stop all infrastructure
make down

# Run database migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create
# Then enter migration name when prompted
```

### Testing

```bash
# Run all tests
make test

# Run tests for a specific package
go test -v ./internal/modules/user/...

# Run a single test
go test -v ./path/to/package -run TestName
```

### Kafka Setup

After running `make up`, create the required Kafka topics:

```bash
docker exec go1_kafka kafka-topics --bootstrap-server localhost:9092 --create --topic user_created --partitions 1 --replication-factor 1
```

## Architecture

### Clean Architecture Layers

The project follows Clean Architecture with strict dependency rules. Each module in `internal/modules/` is organized into:

1. **Domain Layer** (`domain/`): Core business entities and interfaces
   - `entity.go`: Domain models (e.g., `User`)
   - `repository.go`: Repository interfaces
   - `cache.go`: Cache interfaces
   - `event.go`: Event publisher interfaces
   - All other layers depend on domain interfaces

2. **Application Layer** (`usecase/`): Business logic orchestration
   - `usecase.go`: Use case implementations
   - `dto.go`: Input/Output data transfer objects
   - Depends only on domain interfaces

3. **Infrastructure Layer** (`infrastructure/`): External dependencies
   - `repository/postgres/`: PostgreSQL implementation with models and mappers
   - `caching/redis/`: Redis implementation with models and mappers
   - `message_queue/kafka/`: Kafka event publisher implementation
   - Implements domain interfaces

4. **Presentation Layer** (`internal/api/handlers/`): HTTP handlers
   - `handler.go`: Gin HTTP handlers
   - `dto.go`: HTTP request/response DTOs
   - `router.go`: Route registration
   - Depends on use cases

### Module Initialization

Each module has a `module.go` file that wires up dependencies:

```go
// internal/modules/user/module.go
func Init(router *gin.Engine, db *pgxpool.Pool, redisClient *redis.RedisClient, kafkaProducer *kafka.KafkaProducer) {
    repo := repository.NewPostgresUserRepository(db)
    cache := caching.NewRedisUserCache(redisClient)
    event := message_queue.NewKafkaUserEvent(kafkaProducer)

    uc := usecase.NewUserUsecase(repo, cache, event)
    handler := userHandler.NewUserHandler(uc)

    userHandler.RegisterRoutes(router, handler)
}
```

This is called in `internal/api/server/server.go` during server initialization.

### Service Architecture

**API Service Flow:**
1. HTTP request → Handler (Presentation)
2. Handler → UseCase (Application)
3. UseCase → Repository/Cache/Event (Infrastructure)
4. UseCase publishes events to Kafka asynchronously (errors logged but don't fail the request)

**Worker Service Flow:**
1. Worker initializes with PostgreSQL and Redis connections
2. Kafka Consumer → Worker router (`internal/worker/worker.go`)
3. Worker routes by topic to registered handlers (`internal/worker/handlers/`)
4. Handlers receive dependencies (postgres, redis) via constructor injection
5. Handlers process events and can access database/cache as needed

**Worker Dependency Injection:**
Handlers receive postgres and redis connections via constructor injection:
```go
// cmd/worker/main.go
w, err := worker.NewWorkerBuilder(cfg).
    WithPostgres(pg).
    WithRedis(redisClient).
    AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
        return handlers.NewUserCreatedHandler(pg, rds).Handle
    }).
    Build()
```

Handlers can then use these dependencies to access repositories, caches, or use cases:
```go
// internal/worker/handlers/user_created.go
type UserCreatedHandler struct {
    postgres *postgres.Postgres
    redis    *redis.RedisClient
}

func (h *UserCreatedHandler) Handle(ctx context.Context, record *kgo.Record) error {
    // Can instantiate repositories/use cases with h.postgres and h.redis
    userRepo := repository.NewPostgresUserRepository(h.postgres)
    userCache := caching.NewRedisUserCache(h.redis)
    // ... business logic
}
```

### Error Handling

The project uses custom error codes in `pkg/apperrors/`:
- `1001`: Email already exists
- `1002`: User not found

Errors are returned as structured JSON responses via `pkg/response/response.go`.

### Observability

The application is fully instrumented:
- **Metrics**: Prometheus metrics exposed at `/metrics` (request count, duration, CCU)
- **Tracing**: OpenTelemetry → Jaeger for distributed tracing
- **Logging**: Structured logging with Zap

All middleware is registered in `internal/api/server/server.go`:
- `middleware.MetricsMiddleware()`: Prometheus instrumentation
- `middleware.TracingMiddleware()`: OpenTelemetry tracing
- `middleware.AuthMiddleware(true)`: Auth bypass mode (set to `false` for enforcement)
- `middleware.CORSMiddleware()`: CORS handling

### Configuration

Configuration is loaded via Viper from `config/config.yaml` and can be overridden with environment variables (e.g., `APP_PORT=9090`). See `config/config.go` for the loader and individual config structs in `config/*.go`.

## Adding New Features

### Adding a New Module

1. Create module directory: `internal/modules/{module_name}/`
2. Define domain entities and interfaces in `domain/`
3. Implement use cases in `usecase/`
4. Implement infrastructure adapters in `infrastructure/`
5. Create HTTP handlers in `internal/api/handlers/{module_name}/`
6. Create `module.go` for dependency injection
7. Call `{module}.Init()` in `internal/api/server/server.go`

### Adding Event Handlers to Worker

1. Define domain event interface in the module's `domain/event.go`
2. Implement Kafka publisher in module's `infrastructure/message_queue/kafka.go`
3. Create handler in `internal/worker/handlers/{event_name}.go`
4. Register handler in `internal/worker/worker.go:setupHandlers()`
5. Add topic to consumer subscription in `worker.go:Run()`

### Database Migrations

Use `make migrate-create` to generate migration files in `migrations/`. The project uses golang-migrate with sequential naming (e.g., `000001_create_users_table.up.sql`).

## Tech Stack

- **Go**: 1.24.0
- **Web Framework**: Gin
- **Database**: PostgreSQL 15 (pgx/v5 driver)
- **Caching**: Redis 7 (go-redis/v9)
- **Message Queue**: Kafka via franz-go
- **Observability**: Prometheus, Grafana, Jaeger, OpenTelemetry
- **Config**: Viper
- **Logging**: Zap
- **Hot Reload**: Air

## MCP Server Configuration

This project uses MCP (Model Context Protocol) servers for enhanced Claude Code integration. Configuration is stored in `.mcp.json` at the project root.

### GitHub MCP Server

The GitHub MCP server enables Claude to interact with GitHub repositories, issues, pull requests, and workflows.

**Setup:**

1. Generate a GitHub Personal Access Token:
   - Visit: https://github.com/settings/tokens
   - Click "Generate new token (classic)"
   - Select required scopes: `repo`, `read:user`, `workflow`
   - Copy the generated token

2. Add the token to your shell configuration:
   ```bash
   echo 'export GITHUB_TOKEN=your_github_token_here' >> ~/.zshrc
   source ~/.zshrc
   ```

3. The `.mcp.json` file is already configured to use the `GITHUB_TOKEN` environment variable

4. Verify the setup:
   ```bash
   claude mcp list
   ```

**Security Note:** Never commit your actual token to version control. The `.mcp.json` file uses environment variable expansion (`${GITHUB_TOKEN}`) to keep tokens secure.

### Available MCP Commands

Within Claude Code, use:
```
/mcp    # Check MCP server status and manage authentication
```

### Adding More MCP Servers

To add additional MCP servers, update `.mcp.json`:

```json
{
  "mcpServers": {
    "github": {
      "type": "http",
      "url": "https://api.github.com/mcp",
      "headers": {
        "Authorization": "Bearer ${GITHUB_TOKEN}"
      }
    },
    "your-server": {
      "type": "http",
      "url": "https://your-api.com/mcp",
      "headers": {
        "Authorization": "Bearer ${YOUR_TOKEN}"
      }
    }
  }
}
```

## Key Dependencies

- Infrastructure clients are in `pkg/`: `postgres`, `redis`, `kafka`
- Shared utilities: `pkg/logger`, `pkg/metrics`, `pkg/response`, `pkg/apperrors`
- OpenTelemetry setup: `pkg/telemetry`

## Common Patterns

- **Repository Pattern**: Database access abstracted behind interfaces
- **Cache-Aside**: Try cache first, fallback to DB, then populate cache
- **Event Publishing**: Publish events after domain operations (errors logged, don't fail request)
- **Mappers**: Convert between domain entities, DB models, and cache models
- **Graceful Shutdown**: Server handles SIGINT/SIGTERM with 5s timeout
