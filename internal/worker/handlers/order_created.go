package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"github.com/IBM/sarama"
)

type Order struct {
	ID     string  `json:"id"`
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}

type OrderCreatedHandler struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
	// Inject dependencies here: repository, service, etc.
}

func NewOrderCreatedHandler(pg *postgres.Postgres, rds *redis.RedisClient) *OrderCreatedHandler {
	return &OrderCreatedHandler{
		postgres: pg,
		redis:    rds,
	}
}

func (h *OrderCreatedHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	var order Order
	if err := json.Unmarshal(message.Value, &order); err != nil {
		// Validation error - không nên retry
		logger.Log.Error("Invalid order JSON", logger.Field{Key: "error", Value: err})
		return nil // Return nil để skip retry
	}

	logger.Log.Info("Processing order_created event",
		logger.Field{Key: "order_id", Value: order.ID},
		logger.Field{Key: "user_id", Value: order.UserID},
		logger.Field{Key: "amount", Value: order.Amount},
	)

	// Business logic here
	// Example: Send confirmation email, update inventory, etc.

	// Simulate processing
	if order.Amount <= 0 {
		// Business validation error - không retry
		logger.Log.Warn("Invalid order amount", logger.Field{Key: "amount", Value: order.Amount})
		return nil
	}

	// Simulate network error - nên retry
	// return fmt.Errorf("failed to process order: network timeout")

	logger.Log.Info("Order processed successfully", logger.Field{Key: "order_id", Value: order.ID})
	return nil
}

// Example để test retry:
func (h *OrderCreatedHandler) HandleWithRetry(ctx context.Context, message *sarama.ConsumerMessage) error {
	var order Order
	if err := json.Unmarshal(message.Value, &order); err != nil {
		return nil // Skip invalid JSON
	}

	// Simulate transient error để test retry mechanism
	attempts := getAttemptFromHeaders(message.Headers)
	if attempts < 2 {
		logger.Log.Warn("Simulating transient error",
			logger.Field{Key: "attempt", Value: attempts})
		return fmt.Errorf("simulated network error on attempt %d", attempts)
	}

	// Success on 3rd attempt
	logger.Log.Info("Order processed after retry",
		logger.Field{Key: "order_id", Value: order.ID},
		logger.Field{Key: "attempts", Value: attempts})
	return nil
}

func getAttemptFromHeaders(headers []*sarama.RecordHeader) int {
	for _, h := range headers {
		if string(h.Key) == "x-attempt" {
			// Parse attempt number
			return len(string(h.Value)) // Simple hack for demo
		}
	}
	return 0
}
