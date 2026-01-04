package application

import (
	"context"
	"fmt"

	"go1/internal/modules/order/domain"
	"go1/internal/modules/order/domain/entity"
	"go1/internal/modules/order/workflow"
	"go1/internal/pkg/request"

	"go.temporal.io/sdk/client"
)

func (s *orderService) CreateRideOrder(ctx context.Context, input CreateRideOrderInput) (*OrderOutput, error) {
	userCtx, ok := request.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("internal: user context missing")
	}
	customerID, driverID, err := s.resolveParticipants(userCtx.Role, userCtx.UserID, input.CustomerID, input.DriverID)
	if err != nil {
		return nil, err
	}

	order := entity.NewRideOrder(
		userCtx.UserID,
		userCtx.Role,
		entity.CustomerVO{ID: customerID},
		entity.DriverVO{ID: driverID},
	)

	if err := s.enrichData(order, ctx, input); err != nil {
		return nil, err
	}

	// 2. Validate Business Rules/Policies
	vCtx := domain.ValidationContext{
		UserID: customerID,
	}
	if err := s.rideValidator.ValidateCreate(ctx, order, vCtx); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 3. Entity Internal Validation (Data Consistency)
	if err := order.Validate(); err != nil {
		return nil, err
	}

	// 4. Set WorkflowID and Save (Infrastructure)
	workflowID := "order_" + order.ID
	order.WorkflowID = workflowID

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 5. Trigger Workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "ORDER_TASK_QUEUE",
	}

	_, err = s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflow.CreateOrderWorkflow, order.ID)
	if err != nil {
		// Log error but don't fail the request? Or return error?
		// Since Order is created, we probably want to return success but log failure to start workflow.
		// Or failing is better so client retries?
		// User said "success sẽ tạo workflow", implying strict dependency.
		// Use fmt.Errorf to return error.
		return nil, fmt.Errorf("failed to start workflow: %w", err)
	}

	return s.mapper.ToOrderOutput(order), nil
}

func (s *orderService) enrichData(order *entity.RideOrderEntity, ctx context.Context, input CreateRideOrderInput) error {
	service, err := s.getService(ctx, input.ServiceID, input.ServiceType)
	if err != nil {
		return err
	}
	order.SetService(*service)

	payment, err := s.getPayment(ctx, input.CustomerID, input.PaymentMethod)
	if err != nil {
		return err
	}
	order.SetPayment(*payment)

	var points []entity.PointVO
	for i, p := range input.Points {
		points = append(points, entity.PointVO{
			Lat:     p.Lat,
			Lng:     p.Lng,
			Address: p.Address,
			Type:    p.Type,
			Order:   i,
		})
	}

	order.SetPoints(points)
	return nil
}
