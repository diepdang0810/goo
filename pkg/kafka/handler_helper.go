package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"go1/pkg/logger"

	"github.com/IBM/sarama"
)

// TypedMessageHandler is a handler that receives an already-unmarshaled message
type TypedMessageHandler[T any] func(ctx context.Context, message T, metadata *MessageMetadata) error

// MessageMetadata contains Kafka message metadata
type MessageMetadata struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
	Headers   []*sarama.RecordHeader
	Timestamp int64
}

// HandleJSON creates a MessageHandler that automatically unmarshals JSON to your type T
// Example:
//   handler := kafka.HandleJSON(func(ctx context.Context, user User, meta *kafka.MessageMetadata) error {
//       logger.Log.Info("Processing user", logger.Field{Key: "user_id", Value: user.ID})
//       return nil
//   })
func HandleJSON[T any](typedHandler TypedMessageHandler[T]) MessageHandler {
	return func(ctx context.Context, message *sarama.ConsumerMessage) error {
		// Auto-unmarshal JSON to type T
		var payload T
		if err := json.Unmarshal(message.Value, &payload); err != nil {
			logger.Log.Error("❌ Failed to unmarshal message",
				logger.Field{Key: "topic", Value: message.Topic},
				logger.Field{Key: "partition", Value: message.Partition},
				logger.Field{Key: "offset", Value: message.Offset},
				logger.Field{Key: "error", Value: err},
				logger.Field{Key: "raw_value", Value: string(message.Value)})
			// Return nil to skip invalid JSON (don't retry bad data)
			return nil
		}

		// Extract metadata
		metadata := &MessageMetadata{
			Topic:     message.Topic,
			Partition: message.Partition,
			Offset:    message.Offset,
			Key:       string(message.Key),
			Headers:   message.Headers,
			Timestamp: message.Timestamp.Unix(),
		}

		// Get attempt count from headers for logging
		attemptCount := 0
		for _, h := range message.Headers {
			if string(h.Key) == "x-attempt" {
				fmt.Sscanf(string(h.Value), "%d", &attemptCount)
				break
			}
		}

		// Auto-log incoming message
		logger.Log.Info("📥 Processing message",
			logger.Field{Key: "topic", Value: metadata.Topic},
			logger.Field{Key: "partition", Value: metadata.Partition},
			logger.Field{Key: "offset", Value: metadata.Offset},
			logger.Field{Key: "key", Value: metadata.Key},
			logger.Field{Key: "attempt", Value: attemptCount},
			logger.Field{Key: "headers_count", Value: len(metadata.Headers)})

		// Call the typed handler with unmarshaled payload
		if err := typedHandler(ctx, payload, metadata); err != nil {
			logger.Log.Error("❌ Handler failed",
				logger.Field{Key: "topic", Value: metadata.Topic},
				logger.Field{Key: "error", Value: err})
			return err
		}

		// Auto-log success
		logger.Log.Info("✅ Message processed successfully",
			logger.Field{Key: "topic", Value: metadata.Topic},
			logger.Field{Key: "offset", Value: metadata.Offset})

		return nil
	}
}

// HandleRaw creates a MessageHandler that passes the raw message without unmarshaling
// Use this when you need full control over the message
func HandleRaw(handler MessageHandler) MessageHandler {
	return handler
}
