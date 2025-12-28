package handlers

import (
	"context"

	"go1/internal/modules/user/domain"
	"go1/pkg/kafka"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"
)

type UserCreatedHandler struct {
	postgres *postgres.Postgres
	redis    *redis.RedisClient
	// You can inject use cases or services here
	// userService application.UserService
}

func NewUserCreatedHandler(pg *postgres.Postgres, rds *redis.RedisClient) *UserCreatedHandler {
	return &UserCreatedHandler{
		postgres: pg,
		redis:    rds,
	}
}

// Handle uses the new simplified API with auto-unmarshal
func (h *UserCreatedHandler) Handle() kafka.MessageHandler {
	// Use HandleJSON helper - auto-unmarshal + auto-log!
	return kafka.HandleJSON(func(ctx context.Context, user domain.User, meta *kafka.MessageMetadata) error {
		// User is already unmarshaled! No json.Unmarshal needed!

		logger.Log.Info("📧 Processing user_created event",
			logger.Field{Key: "user_id", Value: user.ID},
			logger.Field{Key: "email", Value: user.Email},
			logger.Field{Key: "name", Value: user.Name})

		// TODO: Add business logic here
		// Example with database access:
		// userRepo := repository.NewPostgresUserRepository(h.postgres)
		// userCache := caching.NewRedisUserCache(h.redis)
		// userUsecase := usecase.NewUserUsecase(userRepo, userCache, nil)
		//
		// Example: Send welcome email, create profile, sync to external service, etc.

		return nil
	})
}
