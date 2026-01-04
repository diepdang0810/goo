package order

import (
	"fmt"
	"go1/pkg/utils"
	"time"
)

type CreateOrderRequest struct {
	ServiceID     int32               `json:"service_id" binding:"required"`
	ServiceType   string              `json:"service_type" binding:"required"`
	PaymentMethod string              `json:"payment_method"`
	Points        []OrderPointRequest `json:"points" binding:"required,dive"`
	CustomerID    string              `json:"customer_id"`
	DriverID      string              `json:"driver_id"`
}

func (r *CreateOrderRequest) Validate() error {
	switch utils.ServiceType(r.ServiceType) {
	case utils.ServiceTypeRideTaxi, utils.ServiceTypeRideHour, utils.ServiceTypeRideRoute,
		utils.ServiceTypeRideTrip, utils.ServiceTypeRideShare, utils.ServiceTypeRideShuttle:
		return nil
	default:
		return fmt.Errorf("invalid service_type: %s", r.ServiceType)
	}
}

type OrderPointRequest struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	Type    string  `json:"type" binding:"required"` // 'pickup', 'dropoff'
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
}

type OrderResponse struct {
	ID        string    `json:"id"`
	CreatedBy string    `json:"created_by"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
