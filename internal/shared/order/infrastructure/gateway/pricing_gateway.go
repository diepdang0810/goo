package gateway

import (
	"context"
	"go1/internal/shared/order/domain"
)

type PricingGateway struct {
}

func NewPricingGateway() *PricingGateway {
	return &PricingGateway{}
}

func (p *PricingGateway) EstimatePrice(ctx context.Context, input domain.EstimatePriceInput) (string, error) {
	// TODO: Call external service
	return "fee_123456", nil
}
