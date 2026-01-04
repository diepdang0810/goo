package activity

import (
	"context"
	"go1/internal/modules/order/domain"
)

type OrderActivities struct {
	repo domain.OrderRepository
}

func NewOrderActivities(repo domain.OrderRepository) *OrderActivities {
	return &OrderActivities{repo: repo}
}

func (a *OrderActivities) Pay(ctx context.Context, amount float64, methodID string) (bool, error) {
	// TODO: Use real payment service
	return true, nil
}

func (a *OrderActivities) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	return a.repo.UpdateStatus(ctx, orderID, status)
}

func (a *OrderActivities) SetOrderDispatched(ctx context.Context, orderID string) error {
	return a.repo.UpdateStatus(ctx, orderID, "DISPATCHED")
}

func (a *OrderActivities) SetOrderCompleted(ctx context.Context, orderID string) error {
	return a.repo.UpdateStatus(ctx, orderID, "COMPLETED")
}

func (a *OrderActivities) SetOrderCancelled(ctx context.Context, orderID string) error {
	return a.repo.UpdateStatus(ctx, orderID, "CANCELLED")
}
