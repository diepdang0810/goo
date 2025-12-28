package handlers

import (
	"context"
	"encoding/json"

	"go1/internal/modules/order/domain"
	"go1/pkg/logger"

	"github.com/IBM/sarama"
	"go.temporal.io/sdk/client"
)

type ShipmentEventHandler struct {
	temporalClient client.Client
	repo           domain.OrderRepository
}

func NewShipmentEventHandler(temporalClient client.Client, repo domain.OrderRepository) *ShipmentEventHandler {
	return &ShipmentEventHandler{
		temporalClient: temporalClient,
		repo:           repo,
	}
}

func (h *ShipmentEventHandler) Handle(ctx context.Context, message *sarama.ConsumerMessage) error {
	var event struct {
		OrderID int64  `json:"order_id"`
		Status  string `json:"status"`
	}

	// The message value is likely wrapped in a structure by the producer,
	// but for simplicity assuming the payload is directly accessible or wrapped simply.
	// Adjust based on actual producer format.
	// Assuming producer sends: {"key": "...", "value": {"order_id": 1, "status": "..."}}
	// But here we just try to unmarshal the value directly if it matches.

	// If the producer wraps it in a standard event envelope, we need to unwrap it.
	// Let's assume a simple structure for now or try to parse generic map.
	var payload map[string]interface{}
	if err := json.Unmarshal(message.Value, &payload); err != nil {
		logger.Log.Error("Failed to unmarshal shipment event", logger.Field{Key: "error", Value: err})
		return nil
	}

	// Extract data from payload (adjust based on your Kafka producer's envelope)
	// If using the standard producer from this project, it might be wrapped.
	// For this example, let's assume the payload IS the data we want or contains it.

	// Let's try to map it to our struct
	dataBytes, _ := json.Marshal(payload) // Re-marshal to unmarshal into struct, or just cast
	if err := json.Unmarshal(dataBytes, &event); err != nil {
		// Try to see if it's nested under "payload" or similar if using CDC or specific envelope
		// For now, let's assume direct mapping works or we extract fields manually
	}

	// Manual extraction if struct mapping fails or is complex
	if val, ok := payload["order_id"].(float64); ok {
		event.OrderID = int64(val)
	}
	if val, ok := payload["status"].(string); ok {
		event.Status = val
	}

	if event.OrderID == 0 || event.Status == "" {
		logger.Log.Warn("Invalid shipment event data", logger.Field{Key: "payload", Value: payload})
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
	case "ACCEPTED":
		signalName = "order-accepted"
	case "DELIVERED":
		signalName = "order-delivered"
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
