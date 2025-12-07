package domain

import "context"

type UserCache interface {
	Get(ctx context.Context, id int64) (*User, error)
	Set(ctx context.Context, user *User) error

	Delete(ctx context.Context, id int64) error
}
