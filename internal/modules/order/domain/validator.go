package domain

import (
	"context"
	"go1/internal/modules/order/domain/entity"
)

// RideOrderValidator defines the contract for validating ride orders
type RideOrderValidator interface {
	ValidateCreate(ctx context.Context, order *entity.RideOrderEntity, vCtx ValidationContext) error
}

// ValidationContext contains external data needed for validation
type ValidationContext struct {
	UserID         string
	DriverLocation *entity.PointVO
	// Add other context data as needed
}
