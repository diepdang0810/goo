package consumers

import (
	"context"

	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"
)

type TestSuccessHandlerV2 struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
}

func NewTestSuccessHandlerV2(pg *postgres.Postgres, rds *redis.RedisClient) *TestSuccessHandlerV2 {
	return &TestSuccessHandlerV2{
		postgres: pg,
		redis:    rds,
	}
}

// Handle uses the new simplified API with auto-unmarshal
func (h *TestSuccessHandlerV2) Handle() kafka.MessageHandler {
	// Use HandleJSON helper - auto-unmarshal + auto-log!
	return kafka.HandleJSON(func(ctx context.Context, msg TestMessage, meta *kafka.MessageMetadata) error {
		// Message is already unmarshaled! No json.Unmarshal needed!
		logger.Log.Info("✅ [SUCCESS V2] Processing message",
			logger.Field{Key: "message_id", Value: msg.ID},
			logger.Field{Key: "message_content", Value: msg.Message})

		// Business logic here
		// Can access: h.postgres, h.redis

		return nil
	})
}
