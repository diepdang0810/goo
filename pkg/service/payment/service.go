package payment

import (
	"context"
	"time"
)

type PaymentMethod struct {
	ID     string `json:"id"`
	UserID int64  `json:"user_id"`
	Type   string `json:"type"`
	Last4  string `json:"last4"`
}

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

func (s *PaymentService) GetPaymentMethodByUserID(ctx context.Context, userID int64) ([]PaymentMethod, error) {
	// Mock data
	return []PaymentMethod{
		{
			ID:     "pm_123456789",
			UserID: userID,
			Type:   "credit_card",
			Last4:  "4242",
		},
		{
			ID:     "pm_987654321",
			UserID: userID,
			Type:   "paypal",
			Last4:  "",
		},
	}, nil
}

func (s *PaymentService) Pay(ctx context.Context, amount float64, paymentMethodID string) (bool, error) {
	// Simulate processing time
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(5 * time.Second):
		// Simulate successful payment
		return true, nil
	}
}
