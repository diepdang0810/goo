package validator

import (
	"context"
	"go1/internal/shared/order/domain"
	"go1/internal/shared/order/domain/entity"
)

type rideOrderValidatorImpl struct{}

func NewRideOrderValidator() domain.RideOrderValidator {
	return &rideOrderValidatorImpl{}
}

func (v *rideOrderValidatorImpl) ValidateCreate(ctx context.Context, order *entity.RideOrderEntity) error {

	return nil
}
