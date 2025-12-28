package usecase

import (
	"context"
)

type OrderUsecaseInterface interface {
	Create(ctx context.Context, input CreateOrderInput) (*OrderOutput, error)
	GetByID(ctx context.Context, id int64) (*OrderOutput, error)
}
