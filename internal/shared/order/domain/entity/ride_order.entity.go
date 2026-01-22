package entity

import (
	"time"

	"github.com/oklog/ulid/v2"
)

type RideOrderEntity struct {
	ID            string                 `json:"id"`
	CreatedBy     string                 `json:"created_by"`
	CreatorRole   string                 `json:"creator_role"` // "admin", "driver", "customer"
	Status        OrderStatus            `json:"status"`
	SubStatus     string                 `json:"sub_status"`
	PromotionCode string                 `json:"promotion_code"`
	FeeID         string                 `json:"fee_id"`
	HasInsurance  bool                   `json:"has_insurance"`
	OrderTime     time.Time              `json:"order_time"`
	CompletedTime *time.Time             `json:"completed_time"`
	CancelTime    *time.Time             `json:"cancel_time"`
	Platform      string                 `json:"platform"`
	IsSchedule    bool                   `json:"is_schedule"`
	NowOrder      bool                   `json:"now_order"`
	NowOrderCode  string                 `json:"now_order_code"`
	Payment       PaymentVO              `json:"payment"`
	Metadata      map[string]interface{} `json:"metadata"`
	WorkflowID    string                 `json:"workflow_id"`
	Service       ServiceVO              `json:"service"`
	Customer      CustomerVO             `json:"customer"`
	Driver        DriverVO               `json:"driver,omitempty"`
	Points        []PointVO              `json:"points"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func (o *RideOrderEntity) IsCreatedByAdmin() bool {
	return o.CreatorRole == "admin"
}

func (o *RideOrderEntity) IsCreatedByDriver() bool {
	return o.CreatorRole == "driver"
}

func (o *RideOrderEntity) IsCreatedByCustomer() bool {
	return o.CreatorRole == "customer"
}

func NewRideOrder(
	createdBy string,
	creatorRole string,
	customer CustomerVO,
	driver DriverVO,
) *RideOrderEntity {
	orderID := ulid.Make().String()
	order := &RideOrderEntity{
		ID:          orderID,
		WorkflowID:  "order_" + orderID,
		CreatedBy:   createdBy,
		CreatorRole: creatorRole,
		Customer:    customer,
		Driver:      driver,
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	order.SetInitialStatus()
	return order
}

func (o *RideOrderEntity) SetInitialStatus() {
	if o.IsCreatedByDriver() {
		o.Status = StatusAssigned
	} else {
		o.Status = StatusFinding
	}
}

func (o *RideOrderEntity) Validate() error {
	if len(o.Points) == 0 {
		return &DomainError{Code: "INVALID_POINTS", Message: "at least one point is required"}
	}
	if o.Payment.Method == "" {
		return &DomainError{Code: "INVALID_PAYMENT", Message: "payment method is required"}
	}
	if o.Service.ID == 0 {
		return &DomainError{Code: "INVALID_SERVICE", Message: "service is required"}
	}
	if o.Driver.ID == "" {
		return &DomainError{Code: "INVALID_DRIVER", Message: "driver is required"}
	}
	if o.Customer.ID == "" {
		return &DomainError{Code: "INVALID_CUSTOMER", Message: "customer is required"}
	}
	return nil
}

func (o *RideOrderEntity) SetPayment(payment PaymentVO) {
	o.Payment = payment
}

func (o *RideOrderEntity) SetService(service ServiceVO) {
	o.Service = service
}

func (o *RideOrderEntity) SetPoints(points []PointVO) {
	o.Points = points
}

func (o *RideOrderEntity) SetCustomer(customer CustomerVO) {
	o.Customer = customer
}

func (o *RideOrderEntity) SetDriver(driver DriverVO) {
	o.Driver = driver
}

func (o *RideOrderEntity) GetPoints() []PointVO {
	return o.Points
}

func (o *RideOrderEntity) GetPayment() PaymentVO {
	return o.Payment
}

func (o *RideOrderEntity) GetService() ServiceVO {
	return o.Service
}
