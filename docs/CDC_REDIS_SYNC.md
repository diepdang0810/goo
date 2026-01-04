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

When getting an order by ID (`GET /orders/:id`):

```go
// internal/modules/order/usecase/usecase.go
func (u *OrderUsecase) GetByID(ctx context.Context, id string) (*OrderOutput, error) {
    // 1. Try cache first
    if order, err := u.cache.Get(ctx, id); err == nil {
        return u.toOutput(order), nil  // ✅ Fast path - cache hit
    }

    // 2. Fallback to DB on cache miss
    order, err := u.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Populate cache for next time
    u.cache.Set(ctx, order)
    return u.toOutput(order), nil
}
```

### 2. **Automatic Cache Sync via CDC**

When order data changes in PostgreSQL:

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

This creates a connector named `orders-connector` that:
- Monitors the `public.orders` table
- Publishes changes to `dbserver1.public.orders` Kafka topic
- Uses PostgreSQL's native `pgoutput` plugin
- Transforms events with `ExtractNewRecordState` for simplified structure

**Verify connector status:**

```bash
./scripts/debezium/check-connector.sh
```

Expected output:
```json
{
  "name": "orders-connector",
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
Subscribed to topics: [order_created dbserver1.public.orders test_success test_retry ...]
```

## Testing the CDC Flow

### Test 1: Create Order (INSERT)

```bash
# Create a new order
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": 1,
    "service_type": "delivery",
    "customer_id": "cust_123",
    "points": [{"lat": 10.0, "lng": 106.0, "type": "pickup"}]
  }'

# Response: {"id": "ord_1", "status": "new", ...}
```

**What happens:**
1. Order inserted into PostgreSQL
2. Debezium captures INSERT event
3. CDC handler caches order in Redis
4. Worker logs: `🔄 CDC: Order created/snapshot order_id=ord_1`
5. Worker logs: `✅ Order cached successfully order_id=ord_1`

**Verify cache:**
```bash
# Check Redis directly
docker exec -it go1_redis redis-cli
> GET order:ord_1
# Should return the cached order JSON
```

**Get order (should hit cache):**
```bash
curl http://localhost:8080/orders/ord_1
```

Check API logs for: `Order found in cache id=ord_1` ✅

### Test 2: Update Order (UPDATE)

```bash
# First, create an order to get an ID
curl -X POST http://localhost:8080/orders ...

# Update via database (simulating external change)
docker exec -it go1_postgres psql -U postgres -d go1_db -c \
  "UPDATE orders SET status='completed' WHERE id='ord_1';"
```

**What happens:**
1. Database updated directly
2. Debezium captures UPDATE event
3. CDC handler updates Redis cache
4. Worker logs: `🔄 CDC: Order updated order_id=ord_1`
5. Worker logs: `✅ Order cache updated successfully order_id=ord_1`

**Verify cache updated:**
```bash
curl http://localhost:8080/orders/ord_1
# Should show "completed" from cache
```

### Test 3: Delete Order (DELETE)

```bash
# Delete an order
curl -X DELETE http://localhost:8080/orders/ord_1
```

**What happens:**
1. Order deleted from PostgreSQL
2. Debezium captures DELETE event
3. CDC handler removes from Redis
4. Worker logs: `🔄 CDC: Order deleted order_id=ord_1`
5. Worker logs: `✅ Order removed from cache successfully order_id=ord_1`

**Verify cache deleted:**
```bash
docker exec -it go1_redis redis-cli
> GET order:ord_1
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
# - dbserver1.public.orders (CDC events)
# - order_created (application events)
```

### Monitor CDC Events

```bash
# Consume CDC events in real-time
docker exec go1_kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic dbserver1.public.orders \
  --from-beginning
```

**Example CDC Event (after ExtractNewRecordState):**

```json
{
  "id": "ord_1",
  "status": "new",
  "service_id": 1,
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
curl http://localhost:8083/connectors/orders-connector/status

# Get connector config
curl http://localhost:8083/connectors/orders-connector
```

### Check Redis Cache

```bash
# Connect to Redis
docker exec -it go1_redis redis-cli

# List all order keys
KEYS order:*

# Get specific order
GET order:ord_1

# Check TTL (should be ~600 seconds = 10 minutes)
TTL order:ord_1

# Flush all cache (for testing)
FLUSHALL
```

### Worker Logs

The CDC handler logs all operations:

```
🔄 CDC: Order created/snapshot order_id=ord_1
✅ Order cached successfully order_id=ord_1

🔄 CDC: Order updated order_id=ord_1
✅ Order cache updated successfully order_id=ord_1

🔄 CDC: Order deleted order_id=ord_1
✅ Order removed from cache successfully order_id=ord_1
```

## Configuration

### Debezium Connector Config

The connector is configured in `scripts/debezium/register-connector.sh`:

```json
{
  "name": "orders-connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "postgres",
    "database.password": "postgres",
    "database.dbname": "go1_db",
    "database.server.name": "go1",
    "table.include.list": "public.orders",
    "topic.prefix": "dbserver1",
    "plugin.name": "pgoutput",
    "slot.name": "debezium_slot"
  }
}
```

**Key settings:**
- `table.include.list`: Only monitor `public.orders` table
- `topic.prefix`: CDC events go to `dbserver1.public.orders`
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
     --topic dbserver1.public.orders
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

- [ ] Add CDC for other tables (users, products, etc.)
- [ ] Implement cache warming on application startup
- [ ] Add Prometheus metrics for CDC lag monitoring
- [ ] Implement distributed cache with Redis Cluster
- [ ] Add cache versioning for breaking schema changes
