package consumers

import (
	"context"
	"encoding/json"

	"go1/internal/modules/order/domain"
	"go1/pkg/logger"

	"github.com/IBM/sarama"
	"go.temporal.io/sdk/client"
)

type DispatchConsumer struct {
	temporalClient client.Client
	repo           domain.OrderRepository
}

func NewDispatchConsumer(temporalClient client.Client, repo domain.OrderRepository) *DispatchConsumer {
	return &DispatchConsumer{
		temporalClient: temporalClient,
		repo:           repo,
	}
}

func (h *DispatchConsumer) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	var event struct {
		OrderID        string `json:"order_id"`
		DispatchStatus string `json:"dispatch_status"`
	}

	if err := json.Unmarshal(message.Value, &event); err != nil {
		logger.Log.Error("Failed to unmarshal dispatch event", logger.Field{Key: "error", Value: err})
		return nil
	}

	logger.Log.Info("Processing dispatch event",
		logger.Field{Key: "orderID", Value: event.OrderID},
		logger.Field{Key: "dispatchStatus", Value: event.DispatchStatus})

	// Get Order to find WorkflowID
	order, err := h.repo.GetByID(ctx, event.OrderID)
	if err != nil {
		logger.Log.Error("Failed to get order", logger.Field{Key: "error", Value: err})
		return err
	}

	// Signal Temporal Workflow
	workflowID := order.WorkflowID
	runID := "" // Use empty runID to signal the latest run
	signalName := "order-dispatched"

	err = h.temporalClient.SignalWorkflow(ctx, workflowID, runID, signalName, event)
	if err != nil {
		logger.Log.Error("Failed to signal workflow", logger.Field{Key: "error", Value: err})
		return err
	}

	return nil
}
