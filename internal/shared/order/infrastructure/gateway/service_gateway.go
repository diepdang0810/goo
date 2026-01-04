package gateway

import (
	"context"
	"go1/internal/shared/order/domain/entity"
)

type ServiceGateway struct {
}

func NewServiceGateway() *ServiceGateway {
	return &ServiceGateway{}
}

func (s *ServiceGateway) GetService(ctx context.Context, serviceID int32, serviceType string) (*entity.ServiceVO, error) {
	// TODO: Call actual external service & compare serviceType
	return &entity.ServiceVO{
		ID:                 serviceID,
		Type:               serviceType,
		Name:               "Bike",
		PricingMode:        "GPS",
		AutoCancelInterval: 300,
		DriverLockTime:     15,
		Addons:             []string{"insurance"},
		Enable:             true,
	}, nil
}
