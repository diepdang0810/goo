package domain

import (
	"errors"
	"time"
)

var ErrInvalidTransition = errors.New("invalid transition")



type ServiceType string

const (
	TypeFood  ServiceType = "FOOD"
	TypeRide  ServiceType = "RIDE"
	TypeHotel ServiceType = "HOTEL"
)

type Status string

const (
	StatusCreated     Status = "CREATED"
	StatusPaid        Status = "PAID"
	StatusCooking     Status = "COOKING"
	StatusDelivering  Status = "DELIVERING"
	StatusCompleted   Status = "COMPLETED"

	// common cancel / fail
	StatusCancelled   Status = "CANCELLED"
)

type Order struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Amount     float64   `json:"amount"`
	Status     Status    `json:"status"`
	ServiceType ServiceType `json:"service_type"`
	WorkflowID string    `json:"workflow_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
func (o *Order) Transition(
	to Status,
	sm StateMachine,
) error {
	if !sm.CanTransition(o.Status, to) {
		return ErrInvalidTransition
	}
	o.Status = to
	return nil
}