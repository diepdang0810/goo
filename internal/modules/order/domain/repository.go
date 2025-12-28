package domain

import "context"

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id int64) (*Order, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
}
