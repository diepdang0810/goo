package domain

import (
	"context"
	"go1/internal/shared/order/domain/entity"
)

// PricingGateway defines the contract for pricing services
type PricingGateway interface {
	EstimatePrice(ctx context.Context, input EstimatePriceInput) (string, error)
}

// ServiceGateway defines the contract for service management
type ServiceGateway interface {
	GetService(ctx context.Context, id int32, serviceType string) (*entity.ServiceVO, error)
}

// PaymentGateway defines the contract for payment services
type PaymentGateway interface {
	GetPaymentInfo(ctx context.Context, userID string, method string) (*entity.PaymentVO, error)
}

// LocationGateway defines the contract for location services
type LocationGateway interface {
	GetDriverLocation(ctx context.Context, driverID string) (*entity.PointVO, error)
}

// Input structs for gateways

type EstimatePriceInput struct {
	ServiceID     string
	IsSchedule    bool
	OrderTime     int64 // Unix timestamp
	OrderID       string
	Points        []entity.PointVO
	ServiceAddons []string
}
