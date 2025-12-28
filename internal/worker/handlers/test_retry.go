package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"github.com/IBM/sarama"
)

type TestRetryHandler struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
}

func NewTestRetryHandler(pg *postgres.Postgres, rds *redis.RedisClient) *TestRetryHandler {
	return &TestRetryHandler{
		postgres: pg,
		redis:    rds,
	}
}

func (h *TestRetryHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Extract attempt count from headers
	attemptCount := 0
	for _, header := range message.Headers {
		if string(header.Key) == "x-attempt" {
			if val, err := strconv.Atoi(string(header.Value)); err == nil {
				attemptCount = val
			}
		}
	}

	var msg TestMessage
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		logger.Log.Error("❌ Failed to unmarshal message",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "raw_value", Value: string(message.Value)})
		return err
	}

	logger.Log.Info("🔄 [RETRY TEST] Processing message",
		logger.Field{Key: "topic", Value: message.Topic},
		logger.Field{Key: "partition", Value: message.Partition},
		logger.Field{Key: "offset", Value: message.Offset},
		logger.Field{Key: "key", Value: string(message.Key)},
		logger.Field{Key: "message_id", Value: msg.ID},
		logger.Field{Key: "message_content", Value: msg.Message},
		logger.Field{Key: "attempt", Value: attemptCount},
		logger.Field{Key: "timestamp", Value: message.Timestamp})

	// Log all headers
	logger.Log.Info("Message headers",
		logger.Field{Key: "count", Value: len(message.Headers)})
	for _, header := range message.Headers {
		logger.Log.Info("  Header",
			logger.Field{Key: "key", Value: string(header.Key)},
			logger.Field{Key: "value", Value: string(header.Value)})
	}

	// Simulate processing that ALWAYS fails to trigger retry → DLQ flow
	logger.Log.Warn("⚠️  Simulating processing error (will trigger retry/DLQ)",
		logger.Field{Key: "message_id", Value: msg.ID},
		logger.Field{Key: "current_attempt", Value: attemptCount})

	// Return error to trigger retry mechanism
	return fmt.Errorf("simulated error for message ID %d (attempt %d)", msg.ID, attemptCount)
}
