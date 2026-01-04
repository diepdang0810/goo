package consumers

import (
	"context"

	"go1/internal/modules/order/domain/entity"
	"go1/internal/modules/order/workflow"
	"go1/pkg/kafka"
	"go1/pkg/logger"

	"go.temporal.io/sdk/client"
)

type OrderConsumer struct {
	temporalClient client.Client
}

func NewOrderConsumer(temporalClient client.Client) *OrderConsumer {
	return &OrderConsumer{
		temporalClient: temporalClient,
	}
}

type OrderEvent struct {
	kafka.CDCEvent
	entity.RideOrderEntity
	Metadata string `json:"metadata"` // Shadow metadata map from entity to handle JSON string
}

func (h *OrderConsumer) Handle() kafka.MessageHandler {
	return kafka.HandleJSON(func(ctx context.Context, event OrderEvent, meta *kafka.MessageMetadata) error {
		// Simulate transient error for retry testing
		// attempt := 0
		// for _, header := range meta.Headers {
		// 	if string(header.Key) == "x-attempt" {
		// 		fmt.Sscanf(string(header.Value), "%d", &attempt)
		// 		break
		// 	}
		// }
		// if attempt < 1 {
		// 	err := fmt.Errorf("simulated error for retry testing (attempt %d)", attempt)
		// 	logger.Log.Error("Simulating transient error", logger.Field{Key: "error", Value: err})
		// 	return err
		// }

		if event.IsCreated() {
			workflowID := event.WorkflowID
			workflowOptions := client.StartWorkflowOptions{
				ID:        workflowID,
				TaskQueue: "ORDER_TASK_QUEUE",
			}

			run, err := h.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflow.CreateOrderWorkflow, event.ID)
			if err != nil {
				logger.Log.Error("Failed to start workflow from CDC", logger.Field{Key: "error", Value: err})
				return err
			}

			logger.Log.Info("Workflow started from CDC",
				logger.Field{Key: "workflowID", Value: workflowID},
				logger.Field{Key: "runID", Value: run.GetID()})
		}

		return nil
	})
}
