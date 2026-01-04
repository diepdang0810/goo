# Kafka Retry/DLQ Testing Guide

## Overview

This guide shows how to test the Kafka consumer with retry and DLQ (Dead Letter Queue) mechanism using Sarama.

## Test Setup

### Available Topics

1. **test_success**: Messages process successfully (no retry)
2. **test_retry**: Messages always fail to test retry → DLQ flow

### Automatic Topic Creation

Kafka will auto-create topics when you publish/consume. The following topics will be created automatically:

- `test_success` - Base topic
- `test_success.dlq` - DLQ topic (auto-created when needed)
- `test_retry` - Base topic
- `test_retry.retry` - Retry topic (auto-created when subscribed)
- `test_retry.dlq` - DLQ topic (auto-created when max retries reached)

### Retry Configuration

From `config/config.yaml`:

```yaml
kafka:
  retry:
    topics:
      test_retry:
        enableRetry: true
        maxAttempts: 3      # Will retry 3 times
        backoffMs: 2000     # 2 seconds between retries
```

## Testing Steps

### 1. Start Infrastructure

```bash
make up
```

### 2. Start Services

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

### 3. Check Available Topics API

```bash
curl http://localhost:8080/api/v1/kafka-test/topics | jq '.'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "available_topics": [
      {
        "name": "test_success",
        "description": "Topic that processes successfully (no errors)",
        "retry": "disabled"
      },
      {
        "name": "test_retry",
        "description": "Topic that fails and retries (with DLQ after max attempts)",
        "retry": "enabled",
        "maxAttempts": "3",
        "backoff": "2000ms"
      }
    ],
    "note": "Use POST /api/v1/kafka-test/publish to send messages"
  }
}
```

### 4. Test Successful Processing (No Retry)

```bash
curl -X POST http://localhost:8080/api/v1/kafka-test/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test_success",
    "count": 10,
    "message": "Test successful processing"
  }' | jq '.'
```

**Expected Worker Logs:**
```
✅ [SUCCESS] Processing message
  topic: test_success
  partition: 0
  offset: 0
  message_id: 1
  message_content: Test successful processing (message #1)
  attempt: 0

Message headers:
  count: 0

✅ Message processed successfully
  message_id: 1
```

### 5. Test Retry → DLQ Flow

```bash
curl -X POST http://localhost:8080/api/v1/kafka-test/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "test_retry",
    "count": 10,
    "message": "Test retry and DLQ"
  }' | jq '.'
```

**Expected Worker Logs:**

**Attempt 1 (base topic):**
```
🔄 [RETRY TEST] Processing message
  topic: test_retry
  partition: 0
  offset: 0
  message_id: 1
  attempt: 0

⚠️  Simulating processing error (will trigger retry/DLQ)
  message_id: 1
  current_attempt: 0

Handler error
  topic: test_retry
  attempt: 1
  maxAttempts: 3
  error: simulated error for message ID 1 (attempt 0)

Message sent to retry topic
  retryTopic: test_retry.retry
  attempts: 1
```

**Attempt 2 (retry topic):**
```
🔄 [RETRY TEST] Processing message
  topic: test_retry.retry
  partition: 0
  offset: 0
  message_id: 1
  attempt: 1

Message headers:
  count: 1
  Header:
    key: x-attempt
    value: 1

Handler error
  topic: test_retry.retry
  attempt: 2
  maxAttempts: 3

Message sent to retry topic
  retryTopic: test_retry.retry
  attempts: 2
```

**Attempt 3 (retry topic):**
```
🔄 [RETRY TEST] Processing message
  topic: test_retry.retry
  partition: 0
  offset: 1
  message_id: 1
  attempt: 2

Message headers:
  count: 1
  Header:
    key: x-attempt
    value: 2

Handler error
  topic: test_retry.retry
  attempt: 3
  maxAttempts: 3

Message sent to retry topic
  retryTopic: test_retry.retry
  attempts: 3
```

**Final attempt → DLQ:**
```
🔄 [RETRY TEST] Processing message
  topic: test_retry.retry
  partition: 0
  offset: 2
  message_id: 1
  attempt: 3

Handler error
  topic: test_retry.retry
  attempt: 4
  maxAttempts: 3

Message sent to DLQ
  dlqTopic: test_retry.dlq
  attempts: 4
```

### 6. Inspect DLQ Messages

```bash
docker exec -it go1_kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic test_retry.dlq \
  --from-beginning \
  --property print.headers=true \
  --property print.key=true
```

**Expected Output:**
```
x-attempt:4	1	{"id":1,"message":"Test retry and DLQ (message #1)","topic":"test_retry"}
x-attempt:4	2	{"id":2,"message":"Test retry and DLQ (message #2)","topic":"test_retry"}
...
```

### 7. Check All Topics Created

```bash
docker exec -it go1_kafka kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list | grep test_
```

**Expected:**
```
test_retry
test_retry.dlq
test_retry.retry
test_success
test_success.dlq
```

## Key Observations

### 1. **Sequential Processing Per Partition**
Each partition processes messages sequentially to maintain ordering:
```
Starting to consume partition
  topic: test_retry
  partition: 0
  initialOffset: 0
```

### 2. **Headers Preserved**
The `x-attempt` header tracks retry count and is preserved across retry/DLQ:
```
Message headers:
  Header:
    key: x-attempt
    value: 3
```

### 3. **Parallel Partition Processing**
Sarama automatically spawns goroutines for each partition:
```
Starting to consume partition topic: test_retry partition: 0
Starting to consume partition topic: test_retry partition: 1
Starting to consume partition topic: test_retry partition: 2
```

### 4. **Graceful Rebalancing**
When worker restarts or rebalances:
```
Consumer group session setup completed
Session context cancelled, exiting consume loop
Consumer group session cleanup completed
```

## Debugging Commands

### View All Messages in Topic
```bash
docker exec -it go1_kafka kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic test_retry \
  --from-beginning \
  --max-messages 10
```

### View Consumer Groups
```bash
docker exec -it go1_kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --list
```

### View Consumer Group Details
```bash
docker exec -it go1_kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group order-worker-group \
  --describe
```

### Reset Consumer Group Offset
```bash
docker exec -it go1_kafka kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --group order-worker-group \
  --topic test_retry \
  --reset-offsets \
  --to-earliest \
  --execute
```

## API Reference

### POST /api/v1/kafka-test/publish

Publish test messages to Kafka.

**Request Body:**
```json
{
  "topic": "test_retry",
  "count": 10,
  "message": "Your test message"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "success": 10,
    "failed": 0,
    "published_ids": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10],
    "failed_reasons": []
  }
}
```

### GET /api/v1/kafka-test/topics

Get available test topics.

**Response:**
```json
{
  "success": true,
  "data": {
    "available_topics": [...]
  }
}
```

## Expected Flow Summary

**test_success topic:**
```
Message → Handler → ✅ Success → Commit Offset
```

**test_retry topic:**
```
Message → Handler → ❌ Error (attempt 1)
  ↓
Retry Topic → Handler → ❌ Error (attempt 2)
  ↓
Retry Topic → Handler → ❌ Error (attempt 3)
  ↓
Retry Topic → Handler → ❌ Error (attempt 4)
  ↓
DLQ Topic (final destination)
```

All messages preserve:
- Original payload
- Original key (for partitioning)
- Headers (including x-attempt counter)
