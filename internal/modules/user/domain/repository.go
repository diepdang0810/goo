package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	Fetch(ctx context.Context) ([]User, error)

	DeleteByID(ctx context.Context, id int64) error
}
