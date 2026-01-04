package application

import (
	"time"

	"go1/internal/modules/order/domain/entity"
)

type CreateRideOrderInput struct {
	ServiceID     int32             `json:"service_id"`
	ServiceType   string            `json:"service_type"`
	PaymentMethod string            `json:"payment_method"`
	Points        []OrderPointInput `json:"points"`
	CustomerID    string            `json:"customer_id"` // Optional: For Admin/Driver to specify customer
	DriverID      string            `json:"driver_id"`   // Optional: For Admin to specify driver
}

type OrderPointInput struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Type    string  `json:"type"` // 'pickup' | 'dropoff'
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
}

type OrderOutput struct {
	ID         string                 `json:"id"`
	CreatedBy  string                 `json:"created_by"`
	Status     string                 `json:"status"`
	Payment    entity.PaymentVO       `json:"payment"` // Exposed full payment info or keep flat? Let's use VO for richness
	Metadata   map[string]interface{} `json:"metadata"`
	WorkflowID string                 `json:"workflow_id"`
	Service    entity.ServiceVO       `json:"service"`
	Customer   entity.CustomerVO      `json:"customer"`
	Driver     entity.DriverVO        `json:"driver,omitempty"`
	Points     []entity.PointVO       `json:"points"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}
