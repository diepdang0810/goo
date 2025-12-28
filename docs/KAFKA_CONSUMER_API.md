# Kafka Consumer API - Simplified with Auto-Unmarshal

## Overview

The Kafka Consumer handler API has been simplified with **automatic JSON unmarshaling** and **auto-logging**.

## The Problem (Before)

Every handler had boilerplate code:

```go
func (h *Handler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
    // 1. Manual unmarshal (repetitive!)
    var payload MyType
    if err := json.Unmarshal(message.Value, &payload); err != nil {
        logger.Log.Error("Failed to unmarshal", ...)
        return nil // Skip invalid JSON
    }

    // 2. Manual metadata extraction
    attempt := extractAttemptFromHeaders(message.Headers)

    // 3. Manual logging
    logger.Log.Info("Processing message",
        logger.Field{Key: "topic", Value: message.Topic},
        logger.Field{Key: "partition", Value: message.Partition},
        ...)

    // 4. Finally... business logic
    return processPayload(payload)
}
```

**Problems:**
- ❌ Repetitive unmarshal code in every handler
- ❌ Manual error handling for invalid JSON
- ❌ Manual metadata extraction
- ❌ Manual logging
- ❌ Hard to see the actual business logic

## The Solution (Now)

Use `kafka.HandleJSON` - auto-unmarshal + auto-log!

```go
func (h *Handler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, payload MyType, meta *kafka.MessageMetadata) error {
        // Payload is already unmarshaled!
        // Metadata is already extracted!
        // Auto-logged on entry and exit!

        // Just write business logic!
        return processPayload(payload)
    })
}
```

## API Reference

### kafka.HandleJSON[T]

Generic helper that handles JSON unmarshaling and logging automatically.

```go
func HandleJSON[T any](handler TypedMessageHandler[T]) MessageHandler
```

### TypedMessageHandler

Your handler receives:
- **payload** (T): Already unmarshaled message
- **meta** (*MessageMetadata): Message metadata

```go
type TypedMessageHandler[T any] func(
    ctx context.Context,
    payload T,
    meta *MessageMetadata,
) error
```

### MessageMetadata

Contains all Kafka message metadata:

```go
type MessageMetadata struct {
    Topic     string                    // Topic name
    Partition int32                     // Partition number
    Offset    int64                     // Message offset
    Key       string                    // Message key
    Headers   []*sarama.RecordHeader    // Message headers
    Timestamp int64                     // Unix timestamp
}
```

## Usage Examples

### 1. Simple Handler - User Created Event

**Before (Verbose):**
```go
func (h *UserCreatedHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
    var user domain.User
    if err := json.Unmarshal(message.Value, &user); err != nil {
        logger.Log.Error("Failed to unmarshal", logger.Field{Key: "error", Value: err})
        return nil
    }

    logger.Log.Info("Processing user_created",
        logger.Field{Key: "user_id", Value: user.ID},
        logger.Field{Key: "email", Value: user.Email})

    // Business logic...
    return sendWelcomeEmail(user)
}
```

**After (Clean):**
```go
func (h *UserCreatedHandler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, user domain.User, meta *kafka.MessageMetadata) error {
        // User already unmarshaled! Auto-logged!

        logger.Log.Info("📧 Sending welcome email",
            logger.Field{Key: "email", Value: user.Email})

        return sendWelcomeEmail(user)
    })
}
```

**Auto-logged output:**
```
📥 Processing message
  topic: user_created
  partition: 0
  offset: 42
  key: 123
  attempt: 0
  headers_count: 0

📧 Sending welcome email
  email: user@example.com

✅ Message processed successfully
  topic: user_created
  offset: 42
```

### 2. Handler with Dependencies

```go
type OrderCreatedHandler struct {
    postgres *postgres.Postgres
    redis    *redis.RedisClient
}

func NewOrderCreatedHandler(pg *postgres.Postgres, rds *redis.RedisClient) *OrderCreatedHandler {
    return &OrderCreatedHandler{postgres: pg, redis: rds}
}

func (h *OrderCreatedHandler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, order Order, meta *kafka.MessageMetadata) error {
        // Order already unmarshaled!

        // Access dependencies
        orderRepo := repository.NewOrderRepository(h.postgres)

        // Business logic
        if err := orderRepo.Save(ctx, &order); err != nil {
            return err // Will trigger retry/DLQ
        }

        return nil
    })
}
```

### 3. Using Metadata

```go
func (h *Handler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, event Event, meta *kafka.MessageMetadata) error {
        // Access metadata
        logger.Log.Info("Processing event",
            logger.Field{Key: "key", Value: meta.Key},
            logger.Field{Key: "partition", Value: meta.Partition})

        // Check if retry
        for _, header := range meta.Headers {
            if string(header.Key) == "x-attempt" {
                logger.Log.Info("This is a retry",
                    logger.Field{Key: "attempt", Value: string(header.Value)})
            }
        }

        return processEvent(event)
    })
}
```

### 4. Worker Registration

**Before:**
```go
AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
    return handlers.NewUserCreatedHandler(pg, rds).Handle
})
```

**After (note the `()`):**
```go
AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
    return handlers.NewUserCreatedHandler(pg, rds).Handle()  // Call Handle()
})
```

## Auto-Logging Features

### On Message Received

```
📥 Processing message
  topic: user_created
  partition: 0
  offset: 42
  key: user_123
  attempt: 0
  headers_count: 0
```

### On Success

```
✅ Message processed successfully
  topic: user_created
  offset: 42
```

### On Unmarshal Error

```
❌ Failed to unmarshal message
  topic: user_created
  partition: 0
  offset: 42
  error: invalid character 'x' looking for beginning of value
  raw_value: invalid json data
```

**Note:** Invalid JSON is **automatically skipped** (returns nil) to prevent infinite retries.

### On Handler Error

```
❌ Handler failed
  topic: user_created
  error: database connection failed
```

**Note:** Handler errors **trigger retry/DLQ** mechanism.

## Comparison: Before vs After

### Example: User Created Handler

**❌ Before (18 lines of boilerplate):**
```go
func (h *UserCreatedHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
    // Boilerplate 1: Extract metadata
    attempt := 0
    for _, h := range message.Headers {
        if string(h.Key) == "x-attempt" {
            fmt.Sscanf(string(h.Value), "%d", &attempt)
        }
    }

    // Boilerplate 2: Unmarshal
    var user domain.User
    if err := json.Unmarshal(message.Value, &user); err != nil {
        logger.Log.Error("Failed to unmarshal", logger.Field{Key: "error", Value: err})
        return nil
    }

    // Boilerplate 3: Log entry
    logger.Log.Info("Processing user",
        logger.Field{Key: "topic", Value: message.Topic},
        logger.Field{Key: "partition", Value: message.Partition},
        logger.Field{Key: "user_id", Value: user.ID})

    // Finally... business logic (3 lines)
    sendWelcomeEmail(user)

    // Boilerplate 4: Log exit
    logger.Log.Info("Done", ...)

    return nil
}
```

**✅ After (5 lines total!):**
```go
func (h *UserCreatedHandler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, user domain.User, meta *kafka.MessageMetadata) error {
        // Just business logic!
        return sendWelcomeEmail(user)
    })
}
```

## Benefits

### 1. **No Boilerplate**
- ✅ Auto-unmarshal JSON to your type
- ✅ Auto-extract metadata
- ✅ Auto-log entry/exit/errors
- ✅ Auto-skip invalid JSON

### 2. **Type Safety**
```go
// Type-safe! Compiler checks payload type
kafka.HandleJSON(func(ctx context.Context, user User, meta *kafka.MessageMetadata) error {
    user.ID   // ✅ Autocomplete works!
    user.Name // ✅ Compile-time safety!
})
```

### 3. **Clean Business Logic**
```go
// Can immediately see what this handler does
return kafka.HandleJSON(func(ctx context.Context, order Order, meta *kafka.MessageMetadata) error {
    return h.processOrder(order)  // Clear business logic!
})
```

### 4. **Consistent Logging**
All handlers log the same format automatically:
- Topic, partition, offset, key
- Attempt count (for retries)
- Success/failure status

### 5. **Error Handling**
```go
return kafka.HandleJSON(func(ctx context.Context, event Event, meta *kafka.MessageMetadata) error {
    // Return error → triggers retry/DLQ
    if err := process(event); err != nil {
        return err  // Auto-logged + retry
    }
    return nil  // Auto-logged success
})
```

## Advanced: Raw Handler

If you need full control (no unmarshal), use `HandleRaw`:

```go
func (h *Handler) Handle() kafka.MessageHandler {
    return kafka.HandleRaw(func(ctx context.Context, message *sarama.ConsumerMessage) error {
        // Full control - no auto-unmarshal
        // Use when you need raw bytes or custom unmarshaling
        return processRawBytes(message.Value)
    })
}
```

## Migration Guide

### Step 1: Update Handler Method Signature

**Before:**
```go
func (h *Handler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error
```

**After:**
```go
func (h *Handler) Handle() kafka.MessageHandler
```

### Step 2: Wrap with HandleJSON

**Before:**
```go
func (h *Handler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
    var payload MyType
    json.Unmarshal(message.Value, &payload)
    // ...
}
```

**After:**
```go
func (h *Handler) Handle() kafka.MessageHandler {
    return kafka.HandleJSON(func(ctx context.Context, payload MyType, meta *kafka.MessageMetadata) error {
        // payload already unmarshaled!
        // ...
    })
}
```

### Step 3: Remove Boilerplate

Remove:
- ❌ `json.Unmarshal` calls
- ❌ Manual error handling for JSON
- ❌ Manual metadata extraction
- ❌ Entry/exit logging

### Step 4: Update Worker Registration

**Before:**
```go
.AddTopic("topic", func(pg, rds) kafka.MessageHandler {
    return handlers.NewHandler(pg, rds).Handle  // No ()
})
```

**After:**
```go
.AddTopic("topic", func(pg, rds) kafka.MessageHandler {
    return handlers.NewHandler(pg, rds).Handle()  // Call Handle()
})
```

## Complete Example

```go
// Handler struct
type UserCreatedHandler struct {
    postgres *postgres.Postgres
    redis    *redis.RedisClient
}

// Constructor
func NewUserCreatedHandler(pg *postgres.Postgres, rds *redis.RedisClient) *UserCreatedHandler {
    return &UserCreatedHandler{postgres: pg, redis: rds}
}

// Handler method - returns MessageHandler
func (h *UserCreatedHandler) Handle() kafka.MessageHandler {
    // Use HandleJSON for auto-unmarshal + auto-log
    return kafka.HandleJSON(func(ctx context.Context, user domain.User, meta *kafka.MessageMetadata) error {
        // User is already unmarshaled! No boilerplate!

        logger.Log.Info("Processing new user",
            logger.Field{Key: "user_id", Value: user.ID},
            logger.Field{Key: "email", Value: user.Email})

        // Business logic with dependencies
        userRepo := repository.NewUserRepository(h.postgres)
        userCache := cache.NewUserCache(h.redis)

        // Example: Send welcome email
        if err := h.sendWelcomeEmail(user); err != nil {
            return err  // Triggers retry/DLQ
        }

        // Example: Cache user
        userCache.Set(ctx, user)

        return nil  // Success!
    })
}

// Worker registration
func (a *App) initWorker() error {
    w, err := NewWorkerBuilder(cfg).
        WithPostgres(postgres).
        WithRedis(redis).
        AddTopic("user_created", func(pg *postgres.Postgres, rds *redis.RedisClient) kafka.MessageHandler {
            return handlers.NewUserCreatedHandler(pg, rds).Handle()  // Call Handle()
        }).
        Build()

    return err
}
```

## Summary

✅ **Before:** 15-20 lines of boilerplate per handler
✅ **After:** 3-5 lines of pure business logic

✅ **Before:** Manual unmarshal, logging, error handling
✅ **After:** Auto-unmarshal, auto-log, auto-skip invalid JSON

✅ **Before:** Hard to see business logic
✅ **After:** Crystal clear what the handler does

Just use:
```go
kafka.HandleJSON(func(ctx context.Context, payload YourType, meta *kafka.MessageMetadata) error {
    // Your business logic here
    return nil
})
```

That's it! 🎉
