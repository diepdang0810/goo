package handlers

import (
	"context"
	"fmt"

	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"
)

type TestRetryHandlerV2 struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
}

func NewTestRetryHandlerV2(pg *postgres.Postgres, rds *redis.RedisClient) *TestRetryHandlerV2 {
	return &TestRetryHandlerV2{
		postgres: pg,
		redis:    rds,
	}
}

// Handle uses the new simplified API with auto-unmarshal
func (h *TestRetryHandlerV2) Handle() kafka.MessageHandler {
	// Use HandleJSON helper - auto-unmarshal + auto-log!
	return kafka.HandleJSON(func(ctx context.Context, msg TestMessage, meta *kafka.MessageMetadata) error {
		// Message is already unmarshaled! No json.Unmarshal needed!

		// Get attempt from metadata (already extracted)
		attempt := 0
		for _, header := range meta.Headers {
			if string(header.Key) == "x-attempt" {
				fmt.Sscanf(string(header.Value), "%d", &attempt)
				break
			}
		}

		logger.Log.Info("🔄 [RETRY V2] Processing with auto-unmarshal",
			logger.Field{Key: "message_id", Value: msg.ID},
			logger.Field{Key: "message_content", Value: msg.Message},
			logger.Field{Key: "current_attempt", Value: attempt})

		// Simulate processing error to trigger retry/DLQ
		return fmt.Errorf("simulated error for message ID %d (attempt %d)", msg.ID, attempt)
	})
}
