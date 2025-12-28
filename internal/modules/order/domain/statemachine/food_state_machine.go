package statemachine

import (
	"go1/internal/modules/order/domain"
)

type FoodStateMachine struct{}

func (sm *FoodStateMachine) CanTransition(from domain.Status, to domain.Status) bool {
	switch from {
	case domain.StatusCreated:
		return to == domain.StatusPaid || to == domain.StatusCancelled
	case domain.StatusPaid:
		return to == domain.StatusCooking || to == domain.StatusCancelled
	case domain.StatusCooking:
		return to == domain.StatusDelivering || to == domain.StatusCancelled
	case domain.StatusDelivering:
		return to == domain.StatusCompleted || to == domain.StatusCancelled
	case domain.StatusCompleted, domain.StatusCancelled:
		return false
	}
	return false
}
