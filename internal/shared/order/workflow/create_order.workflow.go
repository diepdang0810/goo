package workflow

import (
	"time"

	"go1/internal/shared/order/activity"

	"go.temporal.io/sdk/workflow"
)

// Define Activity Stub for type safety
var a *activity.OrderActivities

// --- Types ---

type WorkflowEvent int

const (
	EventUnknown WorkflowEvent = iota
	EventDispatched
	EventDelivered
	EventCancelled
	EventTimeout
)

// --- Workflow Entry Point ---

func CreateOrderWorkflow(ctx workflow.Context, orderID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow Started", "OrderID", orderID)

	ctx = withActivityOptions(ctx)

	// Phase 1: Finding Driver (Wait for Dispatch)
	// Business Rule: Order can be cancelled while finding a driver.
	// Business Rule: If no driver found within 1 minute, timeout and cancel.
	logger.Info("Waiting for dispatch signal", "OrderID", orderID)
	event, err := waitForDispatchOrCancel(ctx)
	if err != nil {
		logger.Error("Error waiting for dispatch", "Error", err)
		return err
	}

	if event == EventCancelled {
		logger.Info("Order cancelled during dispatch phase", "OrderID", orderID)
		return processCancellation(ctx, orderID)
	}

	if event == EventTimeout {
		logger.Info("Order timed out finding driver (1 min)", "OrderID", orderID)
		return processCancellation(ctx, orderID)
	}

	// Phase 2: Driver Found (Dispatched)
	logger.Info("Processing dispatch", "OrderID", orderID)
	if err := processDispatch(ctx, orderID); err != nil {
		logger.Error("Error processing dispatch", "Error", err)
		return err
	}

	// Phase 3: In Transit (Wait for Delivery)
	// Business Rule: Order can be cancelled during transit.
	logger.Info("Waiting for delivery signal", "OrderID", orderID)
	event, err = waitForDeliveryOrCancel(ctx)
	if err != nil {
		logger.Error("Error waiting for delivery", "Error", err)
		return err
	}

	if event == EventCancelled {
		logger.Info("Order cancelled during delivery phase", "OrderID", orderID)
		return processCancellation(ctx, orderID)
	}

	// Phase 4: Order Completed
	logger.Info("Processing completion", "OrderID", orderID)
	err = processCompletion(ctx, orderID)
	if err != nil {
		logger.Error("Error processing completion", "Error", err)
		return err
	}

	logger.Info("Workflow Completed Successfully", "OrderID", orderID)
	return nil
}

// --- Helper Functions ---

func waitForDispatchOrCancel(ctx workflow.Context) (WorkflowEvent, error) {
	var dispatchSignal struct {
		OrderID        string
		DispatchStatus string
	}
	var cancelSignal struct {
		OrderID string
		Status  string
	}

	var event WorkflowEvent = EventUnknown
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, "order-dispatched"), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &dispatchSignal)
		event = EventDispatched
	})

	selector.AddReceive(workflow.GetSignalChannel(ctx, "order-canceled"), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &cancelSignal)
		event = EventCancelled
	})

	// Add 1 minute timeout
	selector.AddFuture(workflow.NewTimer(ctx, 1*time.Minute), func(f workflow.Future) {
		event = EventTimeout
	})

	selector.Select(ctx)
	return event, nil
}

func waitForDeliveryOrCancel(ctx workflow.Context) (WorkflowEvent, error) {
	var deliverySignal struct {
		OrderID string
		Status  string
	}
	var cancelSignal struct {
		OrderID string
		Status  string
	}

	var event WorkflowEvent = EventUnknown
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(workflow.GetSignalChannel(ctx, "order-delivered"), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &deliverySignal)
		event = EventDelivered
	})

	selector.AddReceive(workflow.GetSignalChannel(ctx, "order-canceled"), func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &cancelSignal)
		event = EventCancelled
	})

	selector.Select(ctx)
	return event, nil
}

func processCancellation(ctx workflow.Context, orderID string) error {
	return workflow.ExecuteActivity(ctx, a.SetOrderCancelled, orderID).Get(ctx, nil)
}

func processDispatch(ctx workflow.Context, orderID string) error {
	return workflow.ExecuteActivity(ctx, a.SetOrderDispatched, orderID).Get(ctx, nil)
}

func processCompletion(ctx workflow.Context, orderID string) error {
	return workflow.ExecuteActivity(ctx, a.SetOrderCompleted, orderID).Get(ctx, nil)
}

func withActivityOptions(ctx workflow.Context) workflow.Context {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	return workflow.WithActivityOptions(ctx, ao)
}
