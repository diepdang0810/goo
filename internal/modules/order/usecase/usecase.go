package usecase

import (
	"context"

	"go1/internal/modules/order/domain"
	"go1/internal/modules/order/domain/service"
	"go1/internal/modules/order/domain/strategy"

	"go.temporal.io/sdk/client"
)

type OrderUsecase struct {
	repo            domain.OrderRepository
	temporalClient  client.Client
	strategyFactory *service.StrategyFactory
}

func NewOrderUsecase(repo domain.OrderRepository, temporalClient client.Client) *OrderUsecase {
	return &OrderUsecase{
		repo:            repo,
		temporalClient:  temporalClient,
		strategyFactory: &service.StrategyFactory{},
	}
}

func (u *OrderUsecase) Create(ctx context.Context, input strategy.CreateOrderCommand) (*OrderOutput, error) {
	strategy := u.strategyFactory.Get(input.ServiceType())

	if err := strategy.ValidateCreate(input); err != nil {
		return nil, err
	}

	order := &domain.Order{
		ID:         1,
		UserID:     1,
		Amount:     1,
		ServiceType:input.ServiceType(),
		Status:     strategy.InitialStatus(),
	}

	return u.toOutput(order), u.repo.Create(ctx, order)
}

func (u *OrderUsecase) GetByID(ctx context.Context, id int64) (*OrderOutput, error) {
	order, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return u.toOutput(order), nil
}

func (u *OrderUsecase) toOutput(order *domain.Order) *OrderOutput {
	return &OrderOutput{
		ID:        order.ID,
		UserID:    order.UserID,
		Amount:    order.Amount,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}
}
