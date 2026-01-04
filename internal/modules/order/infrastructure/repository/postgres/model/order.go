package model

import "time"

type OrderModel struct {
	ID            string     `db:"id"`
	CreatedBy     string     `db:"created_by"`
	Status        string     `db:"status"`
	SubStatus     string     `db:"sub_status"`
	PromotionCode string     `db:"promotion_code"`
	FeeID         string     `db:"fee_id"`
	HasInsurance  bool       `db:"has_insurance"`
	OrderTime     time.Time  `db:"order_time"`
	CompletedTime *time.Time `db:"completed_time"`
	CancelTime    *time.Time `db:"cancel_time"`
	Platform      string     `db:"platform"`
	IsSchedule    bool       `db:"is_schedule"`
	NowOrder      bool       `db:"now_order"`
	NowOrderCode  string     `db:"now_order_code"`
	PaymentMethod string     `db:"payment_method"`
	Metadata      []byte     `db:"metadata"`
	WorkflowID    string    `db:"workflow_id"`
	ServiceID     int32      `db:"service_id"`
	ServiceType   string     `db:"service_type"`
	ServiceName   string     `db:"service_name"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}
