# CDC-Based Redis Cache Synchronization

This document describes how Change Data Capture (CDC) with Debezium is used to automatically synchronize PostgreSQL changes to Redis cache.

## Architecture Overview

```
┌─────────────┐      CDC Events      ┌──────────────┐
│  PostgreSQL │ ──────────────────>  │   Debezium   │
│   (WAL)     │                      │   Connector  │
└─────────────┘                      └──────────────┘
                                            │
                                            │ Kafka
                                            ▼
                                     ┌──────────────┐
                                     │ CDC Worker   │
                                     │   Handler    │
                                     └──────────────┘
                                            │
                                            ▼
                                     ┌──────────────┐
                                     │    Redis     │
                                     │    Cache     │
                                     └──────────────┘

┌─────────────┐      Read (cache-first)
│  API Server │ ──────────────────────────> Redis
│             │ <────────────────────────── (fast)
└─────────────┘      Fallback to DB if miss
```

## How It Works

### 1. **Cache-First Read Pattern**

When getting a user by ID (`GET /users/:id`):

```go
// internal/modules/user/usecase/usecase.go
func (u *UserUsecase) GetByID(ctx context.Context, id int64) (*UserOutput, error) {
    // 1. Try cache first
    if user, err := u.cache.Get(ctx, id); err == nil {
        return u.toOutput(user), nil  // ✅ Fast path - cache hit
    }

    // 2. Fallback to DB on cache miss
    user, err := u.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Populate cache for next time
    u.cache.Set(ctx, user)
    return u.toOutput(user), nil
}
```

### 2. **Automatic Cache Sync via CDC**

When user data changes in PostgreSQL:

**INSERT/UPDATE:**
```
1. Application updates PostgreSQL
2. PostgreSQL WAL captures the change
3. Debezium reads WAL and publishes to Kafka topic
4. Worker CDC handler consumes event
5. Handler updates Redis with new data
6. Next read gets fresh data from cache
```

**DELETE:**
```
1. Application deletes from PostgreSQL
2. Debezium captures delete event
3. Worker CDC handler consumes event
4. Handler removes key from Redis
5. Next read will cache miss (expected behavior)
```

## Setup Instructions

### 1. Start Infrastructure

```bash
# Start all services including Debezium
make up

# Wait for services to be ready (~30 seconds)
```

Services started:
- **Debezium (Kafka Connect)**: http://localhost:8083
- PostgreSQL (with `wal_level=logical`)
- Redis
- Kafka

### 2. Register Debezium Connector

Register the PostgreSQL connector to capture changes from the `users` table:

```bash
./scripts/debezium/register-connector.sh
```

This creates a connector named `users-connector` that:
- Monitors the `public.users` table
- Publishes changes to `dbserver1.public.users` Kafka topic
- Uses PostgreSQL's native `pgoutput` plugin
- Transforms events with `ExtractNewRecordState` for simplified structure

**Verify connector status:**

```bash
./scripts/debezium/check-connector.sh
```

Expected output:
```json
{
  "name": "users-connector",
  "connector": {
    "state": "RUNNING",
    "worker_id": "kafka-connect:8083"
  },
  "tasks": [
    {
      "id": 0,
      "state": "RUNNING",
      "worker_id": "kafka-connect:8083"
    }
  ]
}
```

### 3. Start Worker

The worker automatically consumes CDC events from `dbserver1.public.users`:

```bash
go run cmd/worker/main.go
```

You'll see logs like:
```
Worker started groupId=go1_worker topicCount=4
Subscribed to topics: [user_created dbserver1.public.users test_success test_retry ...]
```

## Testing the CDC Flow

### Test 1: Create User (INSERT)

```bash
# Create a new user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Response: {"id": 1, "name": "John Doe", "email": "john@example.com", ...}
```

**What happens:**
1. User inserted into PostgreSQL
2. Debezium captures INSERT event
3. CDC handler caches user in Redis
4. Worker logs: `🔄 CDC: User created/snapshot user_id=1`
5. Worker logs: `✅ User cached successfully user_id=1`

**Verify cache:**
```bash
# Check Redis directly
docker exec -it go1_redis redis-cli
> GET user:1
# Should return the cached user JSON
```

**Get user (should hit cache):**
```bash
curl http://localhost:8080/users/1
```

Check API logs for: `User found in cache id=1` ✅

### Test 2: Update User (UPDATE)

```bash
# First, create a user to get an ID
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "email": "jane@example.com"}'

# Update via database (simulating external change)
docker exec -it go1_postgres psql -U postgres -d go1_db -c \
  "UPDATE users SET name='Jane Smith' WHERE email='jane@example.com';"
```

**What happens:**
1. Database updated directly
2. Debezium captures UPDATE event
3. CDC handler updates Redis cache
4. Worker logs: `🔄 CDC: User updated user_id=2`
5. Worker logs: `✅ User cache updated successfully user_id=2`

**Verify cache updated:**
```bash
curl http://localhost:8080/users/2
# Should show "Jane Smith" from cache
```

### Test 3: Delete User (DELETE)

```bash
# Delete a user
curl -X DELETE http://localhost:8080/users/1
```

**What happens:**
1. User deleted from PostgreSQL
2. Debezium captures DELETE event
3. CDC handler removes from Redis
4. Worker logs: `🔄 CDC: User deleted user_id=1`
5. Worker logs: `✅ User removed from cache successfully user_id=1`

**Verify cache deleted:**
```bash
docker exec -it go1_redis redis-cli
> GET user:1
# Should return (nil)
```

## Monitoring & Debugging

### Check Debezium Status

```bash
# Using the helper script
./scripts/debezium/check-connector.sh

# Or manually
curl http://localhost:8083/connectors/users-connector/status | jq .
```

### Check Kafka Topics

```bash
# List all topics
docker exec go1_kafka kafka-topics --bootstrap-server localhost:9092 --list

# Should see:
# - dbserver1.public.users (CDC events)
# - user_created (application events)
```

### Monitor CDC Events

```bash
# Consume CDC events in real-time
docker exec go1_kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic dbserver1.public.users \
  --from-beginning
```

**Example CDC Event (after ExtractNewRecordState):**

```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2025-12-27T10:00:00Z",
  "updated_at": "2025-12-27T10:00:00Z",
  "__op": "c",
  "__deleted": "false"
}
```

Operation codes:
- `"c"`: Create (INSERT)
- `"u"`: Update (UPDATE)
- `"d"`: Delete (DELETE)
- `"r"`: Read (initial snapshot)

### Debezium REST API

Use the Kafka Connect REST API at http://localhost:8083:

```bash
# List connectors
curl http://localhost:8083/connectors

# Check connector status
curl http://localhost:8083/connectors/users-connector/status

# Get connector config
curl http://localhost:8083/connectors/users-connector
```

### Check Redis Cache

```bash
# Connect to Redis
docker exec -it go1_redis redis-cli

# List all user keys
KEYS user:*

# Get specific user
GET user:1

# Check TTL (should be ~600 seconds = 10 minutes)
TTL user:1

# Flush all cache (for testing)
FLUSHALL
```

### Worker Logs

The CDC handler logs all operations:

```
🔄 CDC: User created/snapshot user_id=1 email=john@example.com
✅ User cached successfully user_id=1

🔄 CDC: User updated user_id=1 email=john@example.com
✅ User cache updated successfully user_id=1

🔄 CDC: User deleted user_id=1
✅ User removed from cache successfully user_id=1
```

## Configuration

### Debezium Connector Config

The connector is configured in `scripts/debezium/register-connector.sh`:

```json
{
  "name": "users-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "postgres",
    "database.password": "postgres",
    "database.dbname": "go1_db",
    "database.server.name": "go1",
    "table.include.list": "public.users",
    "topic.prefix": "dbserver1",
    "plugin.name": "pgoutput",
    "slot.name": "debezium_slot"
  }
}
```

**Key settings:**
- `table.include.list`: Only monitor `public.users` table
- `topic.prefix`: CDC events go to `dbserver1.public.users`
- `plugin.name`: Use PostgreSQL's native `pgoutput` (no extensions needed)
- `slot.name`: Replication slot name (prevents WAL accumulation)

### PostgreSQL WAL Configuration

PostgreSQL is configured with logical replication in `docker-compose.yml`:

```yaml
postgres:
  command:
    - "postgres"
    - "-c"
    - "wal_level=logical"
```

This enables CDC without additional plugins.

### Redis Cache TTL

Cache entries expire after 10 minutes (configured in `redis.go:47`):

```go
c.client.Client.Set(ctx, key, data, 10*time.Minute)
```

## Benefits of CDC-Based Caching

✅ **Automatic Invalidation**: No manual cache invalidation logic
✅ **External Changes**: Captures changes from direct DB updates, admin tools, migrations
✅ **Consistency**: Redis always reflects PostgreSQL state (with minimal lag)
✅ **Decoupled**: API service doesn't need to know about cache invalidation
✅ **Audit Trail**: Kafka stores all change events for debugging

## Limitations

⚠️ **Eventual Consistency**: Small delay between DB update and cache update (~100-500ms)
⚠️ **CDC Lag**: Under heavy load, CDC processing may lag behind database writes
⚠️ **Initial Snapshot**: First connector run creates snapshot of all existing data
⚠️ **Schema Changes**: DDL changes (ALTER TABLE) may require connector restart

## Troubleshooting

### Connector Not Running

```bash
# Check connector status
./scripts/debezium/check-connector.sh

# If FAILED, check logs
docker logs go1_kafka_connect

# Restart connector
./scripts/debezium/delete-connector.sh
./scripts/debezium/register-connector.sh
```

### CDC Events Not Consumed

```bash
# Check if worker is subscribed
docker logs go1_worker | grep "Subscribed to topics"

# Should include: dbserver1.public.users

# Check Kafka consumer group
docker exec go1_kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group go1_worker \
  --describe
```

### Cache Not Updating

1. Verify CDC events are published:
   ```bash
   docker exec go1_kafka kafka-console-consumer \
     --bootstrap-server localhost:9092 \
     --topic dbserver1.public.users
   ```

2. Check worker logs for errors:
   ```bash
   docker logs go1_worker
   ```

3. Verify Redis connection:
   ```bash
   docker exec -it go1_redis redis-cli PING
   # Should return: PONG
   ```

### WAL Disk Usage

If PostgreSQL WAL files accumulate:

```bash
# Check replication slot status
docker exec go1_postgres psql -U postgres -d go1_db -c \
  "SELECT * FROM pg_replication_slots;"

# If slot is inactive, restart connector or delete slot:
docker exec go1_postgres psql -U postgres -d go1_db -c \
  "SELECT pg_drop_replication_slot('debezium_slot');"
```

## Cleanup

```bash
# Delete connector
./scripts/debezium/delete-connector.sh

# Stop all services
make down

# Remove volumes (WARNING: deletes all data)
docker-compose down -v
```

## Next Steps

- [ ] Add CDC for other tables (orders, products, etc.)
- [ ] Implement cache warming on application startup
- [ ] Add Prometheus metrics for CDC lag monitoring
- [ ] Implement distributed cache with Redis Cluster
- [ ] Add cache versioning for breaking schema changes
