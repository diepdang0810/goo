package usecase

import (
	"context"

	"go1/internal/user/domain"
	"go1/pkg/logger"
)

type UserUsecase struct {
	repo  domain.UserRepository
	cache domain.UserCache
	event domain.UserEvent
}

func NewUserUsecase(repo domain.UserRepository, cache domain.UserCache, event domain.UserEvent) *UserUsecase {
	return &UserUsecase{
		repo:  repo,
		cache: cache,
		event: event,
	}
}

func (u *UserUsecase) Create(ctx context.Context, input CreateUserInput) error {
	user := &domain.User{
		Name:  input.Name,
		Email: input.Email,
	}

	if err := u.repo.Create(ctx, user); err != nil {
		return err
	}

	if err := u.event.PublishUserCreated(ctx, user); err != nil {
		logger.Log.Error("Failed to publish user_created event", logger.Field{Key: "error", Value: err})
	}

	return nil
}

func (u *UserUsecase) GetByID(ctx context.Context, id int64) (*UserOutput, error) {
	// Try cache
	if user, err := u.cache.Get(ctx, id); err == nil {
		logger.Log.Info("User found in cache", logger.Field{Key: "id", Value: id})
		return u.toOutput(user), nil
	}

	// Fallback to DB
	user, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Set cache
	if err := u.cache.Set(ctx, user); err != nil {
		logger.Log.Warn("Failed to set cache", logger.Field{Key: "error", Value: err})
	}

	return u.toOutput(user), nil
}

func (u *UserUsecase) Fetch(ctx context.Context) ([]UserOutput, error) {
	users, err := u.repo.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	
	outputs := make([]UserOutput, len(users))
	for i, user := range users {
		outputs[i] = *u.toOutput(&user)
	}
	return outputs, nil
}

func (u *UserUsecase) toOutput(user *domain.User) *UserOutput {
	return &UserOutput{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
