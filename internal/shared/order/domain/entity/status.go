package entity

// OrderStatus defines the status of an order
type OrderStatus string

const (
	Scheduling             OrderStatus = "SCHEDULING"
	StatusFinding          OrderStatus = "FINDING"
	StatusAssigned         OrderStatus = "ASSIGNED"
	StatusInProcess        OrderStatus = "IN PROCESS"
	StatusCompleted        OrderStatus = "COMPLETED"
	StatusCancelled        OrderStatus = "CANCELLED"
	WaitingForPayment      OrderStatus = "WAITING FOR PAYMENT"
	PendingForConfirmation OrderStatus = "PENDING FOR CONFIRMATION"
)
