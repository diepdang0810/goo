package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"go1/internal/modules/user/domain"
	"go1/internal/modules/user/infrastructure/caching"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"github.com/IBM/sarama"
)

type UserCDCHandler struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
}

func NewUserCDCHandler(pg *postgres.Postgres, rds *redis.RedisClient) *UserCDCHandler {
	return &UserCDCHandler{
		postgres: pg,
		redis:    rds,
	}
}

// CDCEvent represents the structure of Debezium CDC event after ExtractNewRecordState transform
type CDCEvent struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	// Special fields added by ExtractNewRecordState transform
	Operation string `json:"__op"`     // c=create, u=update, d=delete
	Deleted   string `json:"__deleted"` // "true" for delete operations
}

// Handle processes CDC events from Debezium
func (h *UserCDCHandler) Handle() kafka.MessageHandler {
	return func(ctx context.Context, message *sarama.ConsumerMessage) error {
		var event CDCEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			return fmt.Errorf("failed to unmarshal CDC event: %w", err)
		}

		// Initialize cache
		userCache := caching.NewRedisUserCache(h.redis)

		// Determine operation
		operation := event.Operation
		if operation == "" {
			// Fallback: check for __deleted field
			if event.Deleted == "true" {
				operation = "d"
			}
		}

		switch operation {
		case "c", "r": // create or read (initial snapshot)
			return h.handleCreate(ctx, &event, userCache)
		case "u": // update
			return h.handleUpdate(ctx, &event, userCache)
		case "d": // delete
			return h.handleDelete(ctx, &event, userCache)
		default:
			logger.Log.Warn("Unknown CDC operation",
				logger.Field{Key: "operation", Value: operation},
				logger.Field{Key: "user_id", Value: event.ID})
			return nil
		}
	}
}

func (h *UserCDCHandler) handleCreate(ctx context.Context, event *CDCEvent, cache domain.UserCache) error {
	logger.Log.Info("🔄 CDC: User created/snapshot",
		logger.Field{Key: "user_id", Value: event.ID},
		logger.Field{Key: "email", Value: event.Email})

	user := h.eventToDomain(event)
	if err := cache.Set(ctx, user); err != nil {
		logger.Log.Error("Failed to cache user from CDC create event",
			logger.Field{Key: "user_id", Value: event.ID},
			logger.Field{Key: "error", Value: err})
		return err
	}

	logger.Log.Info("✅ User cached successfully",
		logger.Field{Key: "user_id", Value: event.ID})
	return nil
}

func (h *UserCDCHandler) handleUpdate(ctx context.Context, event *CDCEvent, cache domain.UserCache) error {
	logger.Log.Info("🔄 CDC: User updated",
		logger.Field{Key: "user_id", Value: event.ID},
		logger.Field{Key: "email", Value: event.Email})

	user := h.eventToDomain(event)
	if err := cache.Set(ctx, user); err != nil {
		logger.Log.Error("Failed to update user cache from CDC event",
			logger.Field{Key: "user_id", Value: event.ID},
			logger.Field{Key: "error", Value: err})
		return err
	}

	logger.Log.Info("✅ User cache updated successfully",
		logger.Field{Key: "user_id", Value: event.ID})
	return nil
}

func (h *UserCDCHandler) handleDelete(ctx context.Context, event *CDCEvent, cache domain.UserCache) error {
	logger.Log.Info("🔄 CDC: User deleted",
		logger.Field{Key: "user_id", Value: event.ID})

	if err := cache.Delete(ctx, event.ID); err != nil {
		logger.Log.Error("Failed to delete user from cache via CDC event",
			logger.Field{Key: "user_id", Value: event.ID},
			logger.Field{Key: "error", Value: err})
		return err
	}

	logger.Log.Info("✅ User removed from cache successfully",
		logger.Field{Key: "user_id", Value: event.ID})
	return nil
}

func (h *UserCDCHandler) eventToDomain(event *CDCEvent) *domain.User {
	return &domain.User{
		ID:    event.ID,
		Name:  event.Name,
		Email: event.Email,
		// CreatedAt and UpdatedAt will be parsed from string if needed
		// For now, we keep them as zero values since cache doesn't need exact timestamps
	}
}
