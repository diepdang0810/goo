package consumers

import (
	"context"

	"go1/internal/modules/order/domain"
	"go1/pkg/kafka"
	"go1/pkg/logger"

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

type DispatchEvent struct {
	OrderID        string `json:"order_id"`
	DispatchStatus string `json:"dispatch_status"`
}

func (h *DispatchConsumer) Handle() kafka.MessageHandler {
	return kafka.HandleJSON(func(ctx context.Context, event DispatchEvent, meta *kafka.MessageMetadata) error {
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
	})
}
