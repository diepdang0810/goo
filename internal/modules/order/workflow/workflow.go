package workflow

import (
	"errors"
	"time"

	"go1/internal/modules/order/activity"
	"go1/internal/modules/order/domain"
	"go1/pkg/service/payment"

	"go.temporal.io/sdk/workflow"
)

// Define Activity Stub for type safety
var a *activity.OrderActivities

// --- State Machine Infrastructure ---

type State interface {
	Execute(ctx workflow.Context, data *OrderWorkflowData) (State, error)
}

type OrderWorkflowData struct {
	Input   domain.Order
	OrderID int64
}

// --- Workflow Entry Point ---

func CreateOrderWorkflow(ctx workflow.Context, input domain.Order) (int64, error) {
	ctx = withActivityOptions(ctx)

	data := &OrderWorkflowData{
		Input: input,
	}

	// Initial State
	var currentState State = &PaymentState{}

	for currentState != nil {
		nextState, err := currentState.Execute(ctx, data)
		if err != nil {
			return 0, err
		}
		currentState = nextState
	}

	return data.OrderID, nil
}

func withActivityOptions(ctx workflow.Context) workflow.Context {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	return workflow.WithActivityOptions(ctx, ao)
}

// --- States Implementation ---

// 1. Payment State
type PaymentState struct{}

func (s *PaymentState) Execute(ctx workflow.Context, data *OrderWorkflowData) (State, error) {
	var paymentMethods []payment.PaymentMethod
	if err := workflow.ExecuteActivity(ctx, a.GetPaymentMethod, data.Input.UserID).Get(ctx, &paymentMethods); err != nil {
		return nil, err
	}

	if len(paymentMethods) == 0 {
		return nil, errors.New("no payment methods found")
	}

	var paid bool
	// Use the first payment method for simplicity
	if err := workflow.ExecuteActivity(ctx, a.Pay, data.Input.Amount, paymentMethods[0].ID).Get(ctx, &paid); err != nil {
		return nil, err
	}

	if !paid {
		return nil, errors.New("payment failed")
	}

	return &CreateOrderState{}, nil
}

// 2. Create Order State
type CreateOrderState struct{}

func (s *CreateOrderState) Execute(ctx workflow.Context, data *OrderWorkflowData) (State, error) {
	data.Input.Status = "FINDING"
	if err := workflow.ExecuteActivity(ctx, a.CreateOrder, &data.Input).Get(ctx, &data.OrderID); err != nil {
		return nil, err
	}
	return &FindingState{}, nil
}

// 3. Finding State (Waiting for Accept)
type FindingState struct{}

func (s *FindingState) Execute(ctx workflow.Context, data *OrderWorkflowData) (State, error) {
	// Wait for Accept (and ensure Delivery hasn't come yet)
	if err := waitForSignal(ctx, "order-accepted", "order-delivered"); err != nil {
		return nil, err
	}

	if err := workflow.ExecuteActivity(ctx, a.UpdateOrderStatus, data.OrderID, "PROCESSING").Get(ctx, nil); err != nil {
		return nil, err
	}

	return &ProcessingState{}, nil
}

// 4. Processing State (Waiting for Delivery)
type ProcessingState struct{}

func (s *ProcessingState) Execute(ctx workflow.Context, data *OrderWorkflowData) (State, error) {
	// Wait for Delivery
	if err := waitForSignal(ctx, "order-delivered", ""); err != nil {
		return nil, err
	}

	if err := workflow.ExecuteActivity(ctx, a.UpdateOrderStatus, data.OrderID, "COMPLETED").Get(ctx, nil); err != nil {
		return nil, err
	}

	return nil, nil // Workflow Completed
}

// waitForSignal waits for expectedSignal. If unexpectedSignal is received first, it returns an error.
func waitForSignal(ctx workflow.Context, expectedSignal string, unexpectedSignal string) error {
	var signalData struct{ Status string }
	var unexpectedData struct{ Status string }

	selector := workflow.NewSelector(ctx)
	receivedUnexpected := false
	receivedExpected := false

	selector.AddReceive(workflow.GetSignalChannel(ctx, expectedSignal), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signalData)
		receivedExpected = true
	})

	if unexpectedSignal != "" {
		selector.AddReceive(workflow.GetSignalChannel(ctx, unexpectedSignal), func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &unexpectedData)
			receivedUnexpected = true
		})
	}

	// Wait until one of the signals is received
	selector.Select(ctx)

	if receivedUnexpected {
		return errors.New("invalid sequence: received " + unexpectedSignal + " before " + expectedSignal)
	}

	if !receivedExpected {
		// Should not happen if Select returns, but good for safety
		return errors.New("signal channel closed unexpectedly")
	}

	return nil
}
