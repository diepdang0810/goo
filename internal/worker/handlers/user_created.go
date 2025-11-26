package handlers

import (
	"context"
	"encoding/json"

	"go1/internal/modules/user/domain"
	"go1/pkg/logger"

	"github.com/twmb/franz-go/pkg/kgo"
)

type UserCreatedHandler struct {
	// You can inject use cases or services here
	// userService application.UserService
}

func NewUserCreatedHandler() *UserCreatedHandler {
	return &UserCreatedHandler{}
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
	// Example: Send welcome email, create profile, etc.
	
	return nil
}
