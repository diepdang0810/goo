package consumers

import (
	"context"

	"go1/internal/modules/order/domain"
	"go1/pkg/kafka"
	"go1/pkg/logger"

	"go.temporal.io/sdk/client"
)

type ShipmentConsumer struct {
	temporalClient client.Client
	repo           domain.OrderRepository
}

func NewShipmentConsumer(temporalClient client.Client, repo domain.OrderRepository) *ShipmentConsumer {
	return &ShipmentConsumer{
		temporalClient: temporalClient,
		repo:           repo,
	}
}

type ShipmentEvent struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func (h *ShipmentConsumer) Handle() kafka.MessageHandler {
	return kafka.HandleJSON(func(ctx context.Context, event ShipmentEvent, meta *kafka.MessageMetadata) error {
		// Get Order to find WorkflowID
		order, err := h.repo.GetByID(ctx, event.OrderID)
		if err != nil {
			logger.Log.Error("Failed to get order", logger.Field{Key: "error", Value: err})
			return err
		}

		// Signal Temporal Workflow
		workflowID := order.WorkflowID
		runID := "" // Use empty runID to signal the latest run

		var signalName string
		switch event.Status {
		// REMOVED ACCEPTED
		case "DELIVERED":
			signalName = "order-delivered"
		case "CANCELED": // ADDED CANCELED
			signalName = "order-canceled"
		default:
			logger.Log.Warn("Unknown shipment status for signal", logger.Field{Key: "status", Value: event.Status})
			return nil
		}

		err = h.temporalClient.SignalWorkflow(ctx, workflowID, runID, signalName, event)
		if err != nil {
			logger.Log.Error("Failed to signal workflow", logger.Field{Key: "error", Value: err})
			return err
		}

		return nil
	})
}
