# Kafka Quick Start Guide

## 1. Start Infrastructure

```bash
make up
```

Starts: Kafka, PostgreSQL, Redis, Prometheus, Grafana, Jaeger

## 2. Start Services

**Terminal 1 - API Server:**
```bash
make run
# or
go run cmd/app/main.go
```

**Terminal 2 - Worker:**
```bash
go run cmd/worker/main.go
```

## 3. Configuration

All settings in `config/config.yaml`:

```yaml
kafka:
  brokers: localhost:9099
  groupId: user-worker-group
  producer:
    requiredAcks: all         # all, local, none
    retryMax: 5
    compression: none         # none, gzip, snappy, lz4, zstd
    maxMessageBytes: 1000000  # 1MB
  consumer:
    sessionTimeoutMs: 10000      # 10 seconds
    heartbeatIntervalMs: 3000    # 3 seconds
    maxProcessingTimeMs: 300000  # 5 minutes
  retry:
    topics:
      user_created:
        enableRetry: true
        maxAttempts: 3
        backoffMs: 2000
```

See `docs/KAFKA_CONFIGURATION.md` for detailed tuning guide.

## 4. Producer API (Simplified!)

### Basic Usage

```go
// Publish a struct - auto-marshals to JSON
user := domain.User{ID: 1, Email: "test@example.com"}
err := producer.Publish("user_created", user, "key-123")

// Publish a string
err := producer.Publish("logs", "Log message", "optional-key")

// Publish bytes
err := producer.Publish("raw_data", []byte("raw bytes"))
```

**Features:**
- ✅ Auto-marshal structs to JSON
- ✅ Auto-log all publish attempts
- ✅ Optional key parameter (variadic)
- ✅ No manual `json.Marshal` needed
- ✅ No context parameter needed

### Example Handler

```go
func (k *kafkaUserEvent) PublishUserCreated(ctx context.Context, user *domain.User) error {
    // Simple! Just pass the struct
    key := strconv.FormatInt(user.ID, 10)
    return k.producer.Publish("user_created", user, key)
}
```

## 5. Consumer API (Simplified!)

### Basic Usage

```go
func (h *UserCreatedHandler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, user domain.User, meta *kafka.MessageMetadata) error {
        // User is already unmarshaled! No boilerplate!
        logger.Log.Info("Processing user", logger.Field{Key: "email", Value: user.Email})

        // Your business logic here
        return sendWelcomeEmail(user)
    })
}
```

**Features:**
- ✅ Auto-unmarshal JSON to typed struct
- ✅ Auto-log entry/exit/errors
- ✅ Type-safe with generics
- ✅ No manual `json.Unmarshal` needed
- ✅ Clean business logic

### Metadata Available

```go
meta.Topic       // Topic name
meta.Partition   // Partition number
meta.Offset      // Message offset
meta.Key         // Message key
meta.Headers     // Message headers
meta.Timestamp   // Unix timestamp
```

### Handler with Dependencies

```go
type OrderCreatedHandler struct {
    postgres *postgres.Postgres
    redis    *redis.RedisClient
}

func (h *OrderCreatedHandler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, order Order, meta *kafka.MessageMetadata) error {
        // Access dependencies
        orderRepo := repository.NewOrderRepository(h.postgres)

        // Business logic
        return orderRepo.Save(ctx, &order)
    })
}
```

### Register Handler in Worker

```go
w, err := NewWorkerBuilder(cfg).
    WithPostgres(pg).
    WithRedis(rds).
    AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
        return handlers.NewUserCreatedHandler(pg, rds).Handle()  // Call Handle()
    }).
    Build()
```

## 6. Testing

### Test Successful Processing

```bash
curl -X POST http://localhost:8080/api/v1/kafka-test/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test_success",
    "count": 10,
    "message": "Test successful processing"
  }'
```

### Test Retry/DLQ Flow

```bash
curl -X POST http://localhost:8080/api/v1/kafka-test/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test_retry",
    "count": 10,
    "message": "Test retry and DLQ"
  }'
```

### Check Available Topics

```bash
curl http://localhost:8080/api/v1/kafka-test/topics | jq '.'
```

### Inspect DLQ Messages

```bash
docker exec -it go1_kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic test_retry.dlq \
  --from-beginning \
  --property print.headers=true \
  --property print.key=true
```

## 7. Expected Logs

### Producer Success

```
📤 Message published
  topic: user_created
  partition: 0
  offset: 42
  key: user_123
  value_size: 245
```

### Consumer Processing

```
📥 Processing message
  topic: user_created
  partition: 0
  offset: 42
  key: user_123
  attempt: 0

✅ Message processed successfully
  topic: user_created
  offset: 42
```

### Retry Flow

```
❌ Handler failed
  topic: user_created
  error: database connection failed

Message sent to retry topic
  retryTopic: user_created.retry
  attempts: 1
```

### DLQ

```
Message sent to DLQ
  dlqTopic: user_created.dlq
  attempts: 4
```

## 8. Monitoring

### Prometheus Metrics

Visit: http://localhost:9090

Metrics exposed at: http://localhost:8080/metrics

### Grafana Dashboards

Visit: http://localhost:3000 (admin/admin)

### Jaeger Tracing

Visit: http://localhost:16686

## 9. Common Commands

### List Kafka Topics

```bash
docker exec -it go1_kafka kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list
```

### Create Topic

```bash
docker exec -it go1_kafka kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --topic my_topic \
  --partitions 3 \
  --replication-factor 1
```

### View Consumer Groups

```bash
docker exec -it go1_kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --list
```

### Check Consumer Lag

```bash
docker exec -it go1_kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group user-worker-group \
  --describe
```

## 10. Troubleshooting

### Consumer Not Processing

**Check if worker is running:**
```bash
ps aux | grep worker
```

**Check consumer group:**
```bash
docker exec -it go1_kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group user-worker-group \
  --describe
```

### Messages Not Published

**Check API logs:**
```
# Should see:
📤 Message published
  topic: ...
```

**Check producer errors:**
```
❌ Failed to publish message
  error: ...
```

### Retry Not Working

**Check config:**
```yaml
retry:
  topics:
    your_topic:
      enableRetry: true  # Must be enabled!
      maxAttempts: 3
      backoffMs: 2000
```

**Check if retry topic exists:**
```bash
docker exec -it go1_kafka kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list | grep retry
```

## 11. Best Practices

### Producer

✅ **DO:**
- Pass structs directly - auto-marshals
- Use meaningful keys for partitioning
- Let auto-logging handle observability

❌ **DON'T:**
- Manually marshal JSON (unnecessary)
- Pass context (not needed)
- Add your own logging (auto-logged)

### Consumer

✅ **DO:**
- Use `kafka.HandleJSON` for clean handlers
- Return errors to trigger retry/DLQ
- Keep handlers fast (<5min by default)

❌ **DON'T:**
- Manually unmarshal JSON (auto-handled)
- Add entry/exit logging (auto-logged)
- Swallow errors (return them!)

### Configuration

✅ **DO:**
- Start with defaults
- Monitor metrics after changes
- Enable compression in production (`snappy`)

❌ **DON'T:**
- Tune without data
- Set huge timeouts "just in case"
- Disable retries without reason

## 12. Production Checklist

Before deploying to production:

- [ ] Enable compression: `compression: snappy`
- [ ] Set durability: `requiredAcks: all` (critical data)
- [ ] Configure consumer timeouts appropriately
- [ ] Enable retry for important topics
- [ ] Set up monitoring/alerting
- [ ] Test DLQ processing
- [ ] Document custom configs
- [ ] Load test with production volume

## Documentation References

- **Configuration Details:** `docs/KAFKA_CONFIGURATION.md`
- **Testing Guide:** `docs/KAFKA_TESTING.md`
- **Producer API:** `docs/KAFKA_PRODUCER_API.md`
- **Consumer API:** `docs/KAFKA_CONSUMER_API.md`

## Need Help?

1. Check logs (both API and worker)
2. Check Prometheus metrics
3. Inspect Kafka topics directly
4. Review configuration in `config/config.yaml`
5. See detailed docs in `docs/` folder
