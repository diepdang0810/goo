# Worker Retry & DLQ Architecture

## 📋 Tổng quan

Hệ thống Worker được thiết kế theo **Builder Pattern** với hỗ trợ Retry và Dead Letter Queue (DLQ) tự động, cấu hình linh hoạt theo từng topic.

## 🏗️ Kiến trúc

```
┌─────────────────┐
│ Worker Builder  │  ← Fluent API để config topics
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Kafka Factory   │  ← Build consumer + handlers map
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Kafka Consumer  │  ← Xử lý retry/DLQ logic
└─────────────────┘
```

### Components:

1. **WorkerBuilder** (`builder.go`): Fluent API để thêm topics
2. **KafkaConsumerFactory** (`kafka_factory.go`): Build consumer từ topic configs
3. **Worker** (`worker.go`): Simple router, không quan tâm Kafka details
4. **Consumer** (`pkg/kafka/consumer.go`): Xử lý retry/DLQ tự động

## 🚀 Cách sử dụng

### Thêm topic mới - CỰC KỲ ĐƠN GIẢN!

**File: `cmd/worker/main.go`**

```go
w, err := worker.NewWorkerBuilder(cfg).
    AddTopic("order_created", handlers.NewOrderCreatedHandler().Handle).
    AddTopic("payment_processed", handlers.NewPaymentHandler().Handle).
    Build()
```

Chỉ cần **1 dòng** để thêm topic mới! 🎉

### Cấu hình Retry trong YAML

**File: `config/config.yaml`**

```yaml
kafka:
  brokers: localhost:9099
  groupId: order-worker-group
  retry:
    retrySuffix: ".retry"
    dlqSuffix: ".dlq"
    topics:
      order_created:
        enableRetry: true
        maxAttempts: 5
        backoffMs: 3000
      

      
      # Topic không cần retry
      notification_sent:
        enableRetry: false
```

### Tạo Handler mới

**File: `internal/worker/handlers/order_created.go`**

```go
package handlers

import (
    "context"
    "encoding/json"
    
    "go1/pkg/logger"
    "github.com/twmb/franz-go/pkg/kgo"
)

type OrderCreatedHandler struct {}

func NewOrderCreatedHandler() *OrderCreatedHandler {
    return &OrderCreatedHandler{}
}

func (h *OrderCreatedHandler) Handle(ctx context.Context, record *kgo.Record) error {
    var order Order
    if err := json.Unmarshal(record.Value, &order); err != nil {
        return err // Sẽ tự động retry
    }
    
    // Process order...
    logger.Log.Info("Order processed", logger.Field{Key: "order_id", Value: order.ID})
    return nil
}
```

## 🔄 Flow hoạt động

### Success Flow
```
Message → Handler → ✅ Success → Commit
```

### Retry Flow
```
Message → Handler → ❌ Error (attempt 1)
    ↓
Sleep 2s (backoff)
    ↓
→ topic.retry → Handler → ❌ Error (attempt 2)
    ↓
Sleep 2s
    ↓
→ topic.retry → Handler → ❌ Error (attempt 3)
    ↓
→ topic.dlq (Dead Letter Queue)
```

### Headers tracking
```
x-attempt: 1  → topic.retry → x-attempt: 2 → topic.retry → x-attempt: 3 → topic.dlq
```

## ⚙️ Cấu hình Chi tiết

### Per-Topic Configuration

| Field | Type | Default | Mô tả |
|-------|------|---------|-------|
| `enableRetry` | bool | false | Bật/tắt retry |
| `maxAttempts` | int | 3 | Số lần thử tối đa (bao gồm lần đầu) |
| `backoffMs` | int | 1000 | Độ trễ trước khi retry (ms) |

### Global Configuration

| Field | Type | Default | Mô tả |
|-------|------|---------|-------|
| `retrySuffix` | string | ".retry" | Suffix cho retry topic |
| `dlqSuffix` | string | ".dlq" | Suffix cho DLQ topic |
| `groupId` | string | - | Consumer group ID |

## 📊 Kafka Topics Structure

Với config như trên, hệ thống tự động tạo structure:

```
order_created          ← Base topic
order_created.retry    ← Retry topic (auto-created by consumer)
order_created.dlq      ← Dead letter queue

order_created
order_created.retry
order_created.dlq

notification_sent     ← No retry (theo config)
```

## 🛠️ Development

### Run Worker
```bash
# Development with hot reload
make dev-worker

# Production
go run cmd/worker/main.go
```

### Create Kafka Topics
```bash
# Base topics (tạo thủ công)
docker exec go1_kafka kafka-topics --bootstrap-server localhost:9092 \
  --create --topic order_created --partitions 3 --replication-factor 1

# Retry & DLQ topics tự động được tạo khi có message
```

### Monitor DLQ
```bash
# Xem messages trong DLQ
docker exec go1_kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic order_created.dlq \
  --from-beginning
```

## 🎯 Best Practices

### 1. Idempotent Handlers
Handler phải idempotent vì có thể được gọi nhiều lần:
```go
func (h *Handler) Handle(ctx context.Context, record *kgo.Record) error {
    // ✅ Check if already processed
    if h.repo.IsProcessed(record.Key) {
        return nil
    }
    
    // Process...
    return h.repo.MarkAsProcessed(record.Key)
}
```

### 2. Error Types
Phân biệt lỗi nên retry vs không nên retry:
```go
func (h *Handler) Handle(ctx context.Context, record *kgo.Record) error {
    // ❌ Lỗi validation → KHÔNG NÊN RETRY
    if !isValid(record.Value) {
        logger.Log.Warn("Invalid message, skipping")
        return nil // Return nil để không retry
    }
    
    // ✅ Lỗi network/DB → NÊN RETRY
    if err := h.db.Save(data); err != nil {
        return err // Return error để retry
    }
    
    return nil
}
```

### 3. Backoff Strategy
- Short backoff (1-2s): Lỗi tạm thời (network glitch)
- Long backoff (5-10s): Lỗi service dependency
- Exponential backoff: Cân nhắc implement nếu cần

## 🔍 Troubleshooting

### Message stuck in retry loop?
- Check handler logic
- Check maxAttempts config
- Monitor logs với filter `x-attempt`

### DLQ đầy messages?
- Review handler error handling
- Check external dependencies
- Consider reprocessing from DLQ

### Consumer lag cao?
- Scale workers (tăng instances)
- Tăng partitions
- Optimize handler performance

## 📈 Monitoring

### Metrics to track:
- Retry rate per topic
- DLQ size
- Processing time
- Error rate

### Log patterns:
```json
{
  "level": "error",
  "topic": "order_created",
  "attempt": 2,
  "maxAttempts": 3,
  "error": "..."
}
```

## 🎓 Design Principles

### 1. Separation of Concerns
- **Builder**: API để config topics
- **Factory**: Logic build consumer
- **Worker**: Routing messages
- **Consumer**: Retry/DLQ mechanics

### 2. Open/Closed Principle
- Open for extension: Dễ thêm topic mới
- Closed for modification: Core logic không cần sửa

### 3. Single Responsibility
Mỗi component có 1 nhiệm vụ rõ ràng

### 4. Dependency Inversion
Worker phụ thuộc vào abstraction (MessageHandler interface)

## 🚦 Migration Guide

### Từ old pattern sang new pattern:

**Before:**
```go
// Phải setup nhiều thứ trong worker.go
w := worker.NewWorker(cfg)
w.setupHandlers()
w.buildTopics()
w.configureRetry()
```

**After:**
```go
// Chỉ cần 1 builder chain
w, _ := worker.NewWorkerBuilder(cfg).
    AddTopic("order_created", handler).
    Build()
```

## 📚 References

- [Kafka Retry Pattern](https://www.confluent.io/blog/error-handling-patterns-in-kafka/)
- [Builder Pattern](https://refactoring.guru/design-patterns/builder)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
