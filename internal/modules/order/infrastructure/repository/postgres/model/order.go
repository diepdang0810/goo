package model

import "time"

type OrderModel struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	Amount     float64   `db:"amount"`
	Status     string    `db:"status"`
	WorkflowID string    `db:"workflow_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
