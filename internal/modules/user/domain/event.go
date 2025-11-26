package domain

import "context"

type UserEvent interface {
	PublishUserCreated(ctx context.Context, user *User) error
}
