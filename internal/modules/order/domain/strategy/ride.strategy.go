package strategy

import (
	"errors"
	"go1/internal/modules/order/domain"
)

type RideStrategy struct{}

type CreateRideOrderCommand struct {
	UserID       int64
	RestaurantID string
	Pickup       string
	Dropoff      string
}

func (c CreateRideOrderCommand) ServiceType() domain.ServiceType {
	return domain.TypeRide
}

func (s *RideStrategy) ValidateCreate(cmd CreateRideOrderCommand) error {
	if cmd.RestaurantID == "" {
		return errors.New("restaurant_id is required")
	}
	if cmd.Pickup == "" {
		return errors.New("pickup is required")
	}
	if cmd.Dropoff == "" {
		return errors.New("dropoff is required")
	}
	return nil
}

func (s *RideStrategy) InitialStatus() domain.Status {
	return domain.StatusCreated
}

type RidePayload struct {
	RestaurantID string `json:"restaurant_id"`
	Pickup       string `json:"pickup"`
	Dropoff      string `json:"dropoff"`
}

func (s *RideStrategy) BuildPayload(cmd CreateRideOrderCommand) any {
	return RidePayload{
		RestaurantID: cmd.RestaurantID,
		Pickup:       cmd.Pickup,
		Dropoff:      cmd.Dropoff,
	}
}
