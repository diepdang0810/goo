package activity

import (
	"context"
	"go1/internal/modules/order/domain"
	"go1/pkg/service/payment"
)

type OrderActivities struct {
	paymentService *payment.PaymentService
	repo           domain.OrderRepository
}

func NewOrderActivities(repo domain.OrderRepository) *OrderActivities {
	return &OrderActivities{
		paymentService: payment.NewPaymentService(),
		repo:           repo,
	}
}

func (a *OrderActivities) GetPaymentMethod(ctx context.Context, userID int64) ([]payment.PaymentMethod, error) {
	return a.paymentService.GetPaymentMethodByUserID(ctx, userID)
}

func (a *OrderActivities) Pay(ctx context.Context, amount float64, paymentMethodID string) (bool, error) {
	return a.paymentService.Pay(ctx, amount, paymentMethodID)
}

func (a *OrderActivities) CreateOrder(ctx context.Context, order *domain.Order) (int64, error) {
	err := a.repo.Create(ctx, order)
	if err != nil {
		return 0, err
	}
	return order.ID, nil
}

func (a *OrderActivities) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	return a.repo.UpdateStatus(ctx, orderID, status)
}
