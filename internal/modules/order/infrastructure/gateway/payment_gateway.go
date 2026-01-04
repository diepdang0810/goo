package gateway

import (
	"context"
	"go1/internal/modules/order/domain/entity"
)

type PaymentGateway struct {
}

func NewPaymentGateway() *PaymentGateway {
	return &PaymentGateway{}
}

func (p *PaymentGateway) GetPaymentInfo(ctx context.Context, userID string, method string) (*entity.PaymentVO, error) {
	// TODO: Call external service
	return &entity.PaymentVO{
		Method:      method,
		PaymentType: "prepaid",
		Config:      map[string]interface{}{"card_id": "123"},
	}, nil
}
