package order

import "time"

type CreateOrderRequest struct {
	UserID int64   `json:"user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Type   string  `json:"type" binding:"required"`

	// Food specific
	RestaurantID string   `json:"restaurant_id"`
	Items        []string `json:"items"`

	// Ride specific
	Pickup  string `json:"pickup"`
	Dropoff string `json:"dropoff"`
}

type OrderResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
