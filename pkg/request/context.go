package request

import (
	"context"
	"go1/pkg/utils"
)

type contextKey string

const userContextKey contextKey = "user_context"

type UserContext struct {
	UserID    string
	Role      utils.UserRole
	Platform  string
	PartnerID string
}

func WithUser(ctx context.Context, user *UserContext) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(userContextKey).(*UserContext)
	return user, ok
}
