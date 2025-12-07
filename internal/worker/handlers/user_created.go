package handlers

import (
	"context"
	"encoding/json"

	"go1/internal/modules/user/domain"
	"go1/pkg/logger"
	"go1/pkg/postgres"
	"go1/pkg/redis"

	"github.com/twmb/franz-go/pkg/kgo"
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

func (h *UserCreatedHandler) Handle(ctx context.Context, record *kgo.Record) error {
	var user domain.User
	if err := json.Unmarshal(record.Value, &user); err != nil {
		logger.Log.Error("Failed to unmarshal user", logger.Field{Key: "error", Value: err})
		return err
	}

	logger.Log.Info("Processing user_created event",
		logger.Field{Key: "user_id", Value: user.ID},
		logger.Field{Key: "email", Value: user.Email},
	)

	// TODO: Add business logic here
	// Example with database access:
	// userRepo := repository.NewPostgresUserRepository(h.postgres)
	// userCache := caching.NewRedisUserCache(h.redis)
	// userUsecase := usecase.NewUserUsecase(userRepo, userCache, nil)
	//
	// Example: Send welcome email, create profile, sync to external service, etc.

	return nil
}
