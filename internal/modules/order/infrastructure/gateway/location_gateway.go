package gateway

import (
	"context"
	"go1/internal/modules/order/domain/entity"
)

type LocationGateway struct {
	// client LocationClient
}

func NewLocationGateway() *LocationGateway {
	return &LocationGateway{}
}

// GetDriverLocation implementation
func (l *LocationGateway) GetDriverLocation(ctx context.Context, driverID string) (*entity.PointVO, error) {
	// TODO: Call external service
	return &entity.PointVO{
		Lat: 10.762622,
		Lng: 106.660172,
	}, nil
}
