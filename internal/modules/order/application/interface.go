package application

import (
	"context"
)

type OrderService interface {
	CreateRideOrder(ctx context.Context, input CreateRideOrderInput) (*OrderOutput, error)
	GetByID(ctx context.Context, id string) (*OrderOutput, error)
}
