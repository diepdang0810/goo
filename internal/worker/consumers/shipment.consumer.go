package consumers

import (
	"context"
	"encoding/json"

	"go1/internal/modules/order/domain"
	"go1/pkg/logger"

	"github.com/IBM/sarama"
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

func (h *ShipmentConsumer) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	var event struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}

	if err := json.Unmarshal(message.Value, &event); err != nil {
		logger.Log.Error("Failed to unmarshal shipment event", logger.Field{Key: "error", Value: err})
		return nil
	}

	if event.OrderID == "" || event.Status == "" {
		logger.Log.Warn("Invalid shipment event data", logger.Field{Key: "payload", Value: message.Value})
		return nil
	}

	logger.Log.Info("Processing shipment event",
		logger.Field{Key: "orderID", Value: event.OrderID},
		logger.Field{Key: "status", Value: event.Status})

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
}
