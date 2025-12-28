package strategy

import (
	"go1/internal/modules/order/domain"
)

type CreateOrderCommand interface {
	ServiceType() domain.ServiceType
}

type OrderStrategy interface {
	ValidateCreate(cmd CreateOrderCommand) error
	InitialStatus() domain.Status
	BuildPayload(cmd CreateOrderCommand) any
}
