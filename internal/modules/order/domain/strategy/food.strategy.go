package strategy

import (
	"errors"
	"go1/internal/modules/order/domain"
)

type FoodStrategy struct{}

type CreateFoodOrderCommand struct {
	UserID       int64
	RestaurantID string
	Items        []string
}

func (c CreateFoodOrderCommand) ServiceType() domain.ServiceType {
	return domain.TypeFood
}

func (s *FoodStrategy) ValidateCreate(cmd CreateFoodOrderCommand) error {
	if cmd.RestaurantID == "" {
		return errors.New("restaurant_id is required")
	}
	if len(cmd.Items) == 0 {
		return errors.New("items is required")
	}
	return nil
}

func (s *FoodStrategy) InitialStatus() domain.Status {
	return domain.StatusCreated
}

type FoodPayload struct {
	RestaurantID string   `json:"restaurant_id"`
	Items        []string `json:"items"`
}

func (s *FoodStrategy) BuildPayload(cmd CreateFoodOrderCommand) any {
	return FoodPayload{
		RestaurantID: cmd.RestaurantID,
		Items:        cmd.Items,
	}
}
