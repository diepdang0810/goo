package request

import (
	"context"
)

type contextKey string

const userContextKey contextKey = "user_context"

type UserContext struct {
	UserID   string
	Audience string
	Role     string
	Platform string
}

func WithUser(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(userContextKey).(*UserContext)
	return user, ok
}
