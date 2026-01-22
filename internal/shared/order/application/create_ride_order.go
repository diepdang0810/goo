package application

import (
	"context"
	"fmt"

	"go1/internal/shared/order/domain/entity"
	"go1/pkg/request"
)

func (s *orderService) CreateRideOrder(ctx context.Context, input CreateRideOrderInput) (*OrderOutput, error) {
	userCtx, ok := request.UserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("internal: user context missing")
	}
	fmt.Println("-->userCtx: ", userCtx.Role)
	fmt.Println("-->input: ", input)
	customerID, driverID, err := s.resolveParticipants(userCtx.Role, userCtx.UserID, input.CustomerID, input.DriverID)

	if err != nil {
		return nil, err
	}
	order := entity.NewRideOrder(
		userCtx.UserID,
		string(userCtx.Role),
		entity.CustomerVO{ID: customerID},
		entity.DriverVO{ID: driverID},
	)

	if err := s.enrichData(order, ctx, input); err != nil {
		return nil, err
	}

	// 2. Validate Business Rules/Policies
	if err := s.rideValidator.ValidateCreate(ctx, order); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 3. Entity Internal Validation (Data Consistency)
	if err := order.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
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
	//get pricing
	return nil
}
