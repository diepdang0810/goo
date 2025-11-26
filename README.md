# Go1 Project

## Project Structure

```
.
├── cmd
│   └── app
│       └── main.go           # Application entry point
├── config                    # Configuration (Viper)
├── internal
│   ├── server                # Server setup
│   └── user                  # User Module
│       ├── application       # Application Layer (Usecase, DTOs)
│       ├── domain            # Domain Layer (Entities, Interfaces)
│       ├── infrastructure    # Infrastructure Layer
│       │   ├── caching       # Redis implementation (with model/mapper)
│       │   ├── message_queue # Kafka implementation
│       │   └── repository    # Postgres implementation (with model/mapper)
│       ├── presentation      # Presentation Layer (HTTP Handlers, DTOs)
│       └── module.go         # Module initialization
├── migrations                # SQL Migrations
├── pkg                       # Shared packages (Logger, Postgres, Redis, Kafka, Response, Utils)
└── docker-compose.yml
```

Modular Clean Architecture Golang Project.

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, but recommended)

## Setup

1.  **Clone the repository**
2.  **Start Infrastructure**
    ```bash
    make up
    ```
    This starts Postgres, Redis, Kafka, and Zookeeper.

3.  **Run Migrations**
    ```bash
    make migrate-up
    ```

## Running the Application

### Development Mode (Hot Reload)
```bash
make dev
```

### Production Build
```bash
make build
./bin/app
```

## Configuration

Configuration is managed via `config/config.yaml`. You can also override settings using environment variables (e.g., `APP_PORT=9090`).

## API Endpoints

- `POST /users`: Create a user
- `GET /users`: List users
- `GET /users/:id`: Get user by ID (Cached)
