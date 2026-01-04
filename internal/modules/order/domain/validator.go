package domain

import (
	"context"
	"go1/internal/modules/order/domain/entity"
)

type RideOrderValidator interface {
	ValidateCreate(ctx context.Context, order *entity.RideOrderEntity) error
}
