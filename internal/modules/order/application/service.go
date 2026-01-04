package application

import (
	"context"
	"fmt"
	"go1/internal/modules/order/domain"
	"go1/internal/modules/order/domain/entity"
	"go1/pkg/apperrors"

	"go.temporal.io/sdk/client"
)

type orderService struct {
	repo            domain.OrderRepository
	temporalClient  client.Client
	mapper          *OrderMapper
	pricingGateway  domain.PricingGateway
	serviceGateway  domain.ServiceGateway
	paymentGateway  domain.PaymentGateway
	locationGateway domain.LocationGateway
	rideValidator   domain.RideOrderValidator
}

func NewOrderService(
	repo domain.OrderRepository,
	temporalClient client.Client,
	mapper *OrderMapper,
	pricingGateway domain.PricingGateway,
	serviceGateway domain.ServiceGateway,
	paymentGateway domain.PaymentGateway,
	locationGateway domain.LocationGateway,
	rideValidator domain.RideOrderValidator,
) OrderService {
	return &orderService{
		repo:            repo,
		temporalClient:  temporalClient,
		mapper:          mapper,
		pricingGateway:  pricingGateway,
		serviceGateway:  serviceGateway,
		paymentGateway:  paymentGateway,
		locationGateway: locationGateway,
		rideValidator:   rideValidator,
	}
}

// resolveParticipants helper logic to determine customer/driver IDs based on user role
func (s *orderService) resolveParticipants(role, authUserID, inputCustomerID, inputDriverID string) (customerID string, driverID string, err error) {
	switch role {
	case "admin":
		if inputCustomerID == "" {
			return "", "", fmt.Errorf("customer_id is required for admin")
		}
		// Admin explicitly matches specific driver if provided
		return inputCustomerID, inputDriverID, nil
	case "driver":
		if inputCustomerID == "" {
			return "", "", fmt.Errorf("customer_id is required when driver creates order")
		}
		// Driver creates for themselves
		return inputCustomerID, authUserID, nil
	case "customer":
		// Customer creates for themselves
		return authUserID, "", nil
	default:
		// Default safe behavior: treat as customer
		return authUserID, "", nil
	}
}

func (s *orderService) getService(ctx context.Context, serviceID int32, serviceType string) (*entity.ServiceVO, error) {
	serviceVO, err := s.serviceGateway.GetService(ctx, serviceID, serviceType)
	if err != nil {
		return nil, err
	}

	if !serviceVO.Enable {
		return nil, apperrors.NewServiceDisabledError(fmt.Sprintf("service %d is not enabled", serviceID))
	}
	return serviceVO, nil
}

func (s *orderService) getPayment(ctx context.Context, customerID string, paymentMethod string) (*entity.PaymentVO, error) {
	paymentVO, err := s.paymentGateway.GetPaymentInfo(ctx, customerID, paymentMethod)
	if err != nil {
		return nil, err
	}
	return paymentVO, nil
}
