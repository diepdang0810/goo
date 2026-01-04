package domain

import (
	"context"
	"go1/internal/shared/order/domain/entity"
)

type RideOrderValidator interface {
	ValidateCreate(ctx context.Context, order *entity.RideOrderEntity) error
}
