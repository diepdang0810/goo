package service

import (
	"go1/internal/modules/order/domain"
	"go1/internal/modules/order/domain/statemachine"
	"go1/internal/modules/order/domain/strategy"
)

type StrategyFactory struct{}

func (f *StrategyFactory) Get(t domain.ServiceType) strategy.OrderStrategy {
	switch t {
	case domain.TypeFood:
		return &strategy.FoodStrategy{}
	case domain.TypeRide:
		return &strategy.RideStrategy{}
	default:
		// In a real app, might want to return error or handle better
		panic("unsupported order type")
	}
}

type StateMachineFactory struct{}

func (f *StateMachineFactory) Get(t domain.ServiceType) domain.StateMachine {
	switch t {
	case domain.TypeFood:
		return &statemachine.FoodStateMachine{}
	case domain.TypeRide:
		return &statemachine.RideStateMachine{}
	default:
		panic("unsupported order type")
	}
}
