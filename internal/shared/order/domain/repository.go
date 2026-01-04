package domain

import (
	"context"
	"go1/internal/shared/order/domain/entity"
)

// OrderRepository defines the interface for order storage
type OrderRepository interface {
	Create(ctx context.Context, order *entity.RideOrderEntity) error
	GetByID(ctx context.Context, id string) (*entity.RideOrderEntity, error)
	Update(ctx context.Context, order *entity.RideOrderEntity) error
	UpdateStatus(ctx context.Context, id string, status string) error
	Delete(ctx context.Context, id string) error
}
