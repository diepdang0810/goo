# Kafka Producer API - Simplified Usage

## Overview

The Kafka Producer API has been simplified with **automatic logging** and a clean interface.

## API Signature

```go
func (k *KafkaProducer) Publish(topic string, message any, key ...string) error
```

### Parameters

- **topic** (string, required): Kafka topic name
- **message** (any, required): Message payload
  - `[]byte` → sent as-is
  - `string` → converted to bytes
  - **struct/map** → **auto-marshaled to JSON!** 🎉
- **key** (string, optional): Partition key
  - Pass as variadic parameter
  - Omit for default partitioning
  - Use for consistent routing (same key → same partition)

### Returns

- `error`: nil on success, error on failure
- **Auto-logs** every publish attempt (success or failure)

## Usage Examples

### 1. Basic - Publish without key (default partitioning)

```go
producer, _ := kafka.NewProducer("localhost:9092")

// Simple string message
err := producer.Publish("test_topic", "Hello Kafka")

// Byte array message
data := []byte("Hello Kafka")
err := producer.Publish("test_topic", data)
```

**Auto-logged output:**
```
📤 Message published
  topic: test_topic
  partition: 2
  offset: 42
  key:
  value_size: 11
```

### 2. With partition key (consistent routing)

```go
// Use user ID as key → same user always goes to same partition
userID := "user_123"
message := []byte(`{"user_id": "user_123", "action": "login"}`)

err := producer.Publish("user_events", message, userID)
```

**Auto-logged output:**
```
📤 Message published
  topic: user_events
  partition: 1
  offset: 100
  key: user_123
  value_size: 45
```

### 3. Real-world example - Order Created Event

```go
// In your service/usecase
func (u *OrderUsecase) CreateOrder(ctx context.Context, dto *CreateOrderDTO) error {
    // Create order in database
    order, err := u.repo.Create(ctx, dto)
    if err != nil {
        return err
    }

    // Publish struct directly - auto-marshaled to JSON!
    if err := u.kafka.Publish("order_created", order, order.ID); err != nil {
        // Error is already logged by producer
        // Don't fail the request, just log
        logger.Log.Warn("Failed to publish order_created event",
            logger.Field{Key: "order_id", Value: order.ID})
    }

    return nil
}
```

**Auto-logged output:**
```
📤 Message published
  topic: order_created
  partition: 0
  offset: 523
  key: ord_123
  value_size: 156
```

### 4. API Handler - Test Publishing

```go
func (h *Handler) PublishTestMessages(c *gin.Context) {
    for i := 1; i <= 10; i++ {
        // Pass struct directly - auto-marshaled to JSON!
        msg := TestMessage{
            ID:      i,
            Message: fmt.Sprintf("Test message %d", i),
        }

        key := strconv.Itoa(i)
        if err := h.producer.Publish("test_topic", msg, key); err != nil {
            // Already logged by producer
            continue
        }
    }
}
```

**Auto-logged output (10 messages):**
```
📤 Message published topic: test_topic partition: 0 offset: 1 key: 1 value_size: 45
📤 Message published topic: test_topic partition: 1 offset: 1 key: 2 value_size: 45
📤 Message published topic: test_topic partition: 2 offset: 1 key: 3 value_size: 45
📤 Message published topic: test_topic partition: 0 offset: 2 key: 4 value_size: 45
📤 Message published topic: test_topic partition: 1 offset: 2 key: 5 value_size: 45
📤 Message published topic: test_topic partition: 2 offset: 2 key: 6 value_size: 45
📤 Message published topic: test_topic partition: 0 offset: 3 key: 7 value_size: 45
📤 Message published topic: test_topic partition: 1 offset: 3 key: 8 value_size: 45
📤 Message published topic: test_topic partition: 2 offset: 3 key: 9 value_size: 45
📤 Message published topic: test_topic partition: 0 offset: 4 key: 10 value_size: 46
```

### 5. Error handling with auto-logging

```go
err := producer.Publish("invalid_topic", message, key)
if err != nil {
    // Error is already logged by producer:
    // ❌ Failed to publish message
    //   topic: invalid_topic
    //   key: user_123
    //   value_size: 45
    //   error: kafka: Failed to produce message...

    // Just handle the error in your business logic
    return fmt.Errorf("failed to publish event: %w", err)
}
```

## Comparison: Old vs New API

### ❌ Old API (Complex)

```go
ctx := context.Background()
eventData, err := json.Marshal(order)
if err != nil {
    return err
}

err = producer.Publish(ctx, "order_created", []byte(order.ID), eventData)

// Manual logging
if err != nil {
    logger.Log.Error("Failed to publish",
        logger.Field{Key: "error", Value: err})
} else {
    logger.Log.Info("Published",
        logger.Field{Key: "topic", Value: "order_created"})
}
```

### ✅ New API (Super Simple)

```go
err := producer.Publish("order_created", order, order.ID)
// Auto-marshaled to JSON + Auto-logged! 🎉
```

## Benefits

### 1. **Simpler API**
- No `context.Context` needed (fire-and-forget)
- No byte conversion (handles `string`, `[]byte`, etc.)
- Optional key (variadic parameter)

### 2. **Auto-logging**
- ✅ Success: partition, offset, key, size
- ❌ Failure: topic, key, size, error
- 📋 Headers: automatic for retry/DLQ

### 3. **Better observability**
Every message is logged with:
- Topic name
- Partition assigned
- Offset in partition
- Key used (for debugging partitioning)
- Message size

### 4. **Less boilerplate**
```go
// Before: 10+ lines
ctx := context.Background()
key := []byte(userID)
data, err := json.Marshal(event)
if err != nil {
    return err
}
err = producer.Publish(ctx, topic, key, data)
logger.Log.Info("Published", ...)

// After: 1 line!
producer.Publish(topic, event, userID)
// Auto-marshal + Auto-log! 🚀
```

## Advanced: Internal Retry/DLQ

For retry/DLQ mechanism, use the internal method:

```go
// Used internally by consumer for retry/DLQ
func (k *KafkaProducer) PublishWithHeaders(
    ctx context.Context,
    topic string,
    key, value []byte,
    headers []sarama.RecordHeader,
) error
```

This is automatically called by the consumer when:
- Message fails and needs retry
- Max retries reached and goes to DLQ

**Auto-logged output:**
```
📤 Message published (with headers)
  topic: test_retry.retry
  partition: 0
  offset: 10
  key: 5
  headers_count: 1
  📋 Header:
    key: x-attempt
    value: 2
```

## Configuration

Producer is configured with safe defaults:

```go
config.Producer.RequiredAcks = sarama.WaitForAll  // Wait for all replicas
config.Producer.Retry.Max = 5                     // Retry up to 5 times
config.Producer.Return.Successes = true           // Track successes
```

## Testing

Use the test API endpoints:

```bash
# Publish 10 messages with auto-logging
curl -X POST http://localhost:8080/api/v1/kafka-test/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test_success",
    "count": 10,
    "message": "Test message"
  }'
```

Check logs to see:
```
📤 Message published topic: test_success partition: 0 offset: 1 key: 1
📤 Message published topic: test_success partition: 1 offset: 1 key: 2
📤 Message published topic: test_success partition: 2 offset: 1 key: 3
...
```

## Summary

✅ **Simple**: 1 line to publish with optional key
✅ **Auto-logged**: Every publish is logged automatically
✅ **Type-safe**: Accepts `string`, `[]byte`, or any type
✅ **Observability**: Partition, offset, key, size tracked
✅ **Error-friendly**: Errors are logged and returned

Just call:
```go
producer.Publish(topic, message, optionalKey)
```

That's it! 🎉
