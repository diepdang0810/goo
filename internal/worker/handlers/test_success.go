package handlers

import (
	"context"
	"encoding/json"

	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"github.com/IBM/sarama"
)

type TestSuccessHandler struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
}

type TestMessage struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Topic   string `json:"topic"`
}

func NewTestSuccessHandler(pg *postgres.Postgres, rds *redis.RedisClient) *TestSuccessHandler {
	return &TestSuccessHandler{
		postgres: pg,
		redis:    rds,
	}
}

func (h *TestSuccessHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	// Extract attempt count from headers
	attemptCount := 0
	for _, header := range message.Headers {
		if string(header.Key) == "x-attempt" {
			logger.Log.Info("Found x-attempt header", logger.Field{Key: "value", Value: string(header.Value)})
		}
	}

	var msg TestMessage
	if err := json.Unmarshal(message.Value, &msg); err != nil {
		logger.Log.Error("Failed to unmarshal message",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "raw_value", Value: string(message.Value)})
		return err
	}

	logger.Log.Info("✅ [SUCCESS] Processing message",
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

	// Simulate successful processing
	logger.Log.Info("✅ Message processed successfully",
		logger.Field{Key: "message_id", Value: msg.ID})

	return nil
}
