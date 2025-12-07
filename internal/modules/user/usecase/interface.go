package usecase

import "context"

// UserUsecaseInterface defines the interface for user use cases
type UserUsecaseInterface interface {
	Create(ctx context.Context, input CreateUserInput) error
	GetByID(ctx context.Context, id int64) (*UserOutput, error)
	Fetch(ctx context.Context) ([]UserOutput, error)
	DeleteByID(ctx context.Context, id int64) error
}
