package usecase

import (
	"time"

	"go1/internal/modules/order/domain"
)

type CreateOrderInput struct {
	UserID int64              `json:"user_id"`
	Amount float64            `json:"amount"`
	Type   domain.ServiceType `json:"type"`

	// Food specific
	RestaurantID string   `json:"restaurant_id"`
	Items        []string `json:"items"`

	// Ride specific
	Pickup  string `json:"pickup"`
	Dropoff string `json:"dropoff"`
}

type OrderOutput struct {
	ID        int64         `json:"id"`
	UserID    int64         `json:"user_id"`
	Amount    float64       `json:"amount"`
	Status    domain.Status `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
