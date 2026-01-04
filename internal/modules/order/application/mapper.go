package application

import "go1/internal/modules/order/domain/entity"

type OrderMapper struct {
}

func NewOrderMapper() *OrderMapper {
	return &OrderMapper{}
}

func (m *OrderMapper) ToOrderOutput(order *entity.RideOrderEntity) *OrderOutput {
	if order == nil {
		return nil
	}

	return &OrderOutput{
		ID:            order.ID,
		CreatedBy:     order.CreatedBy,
		Status:        string(order.Status),
		Payment:       order.Payment,
		Metadata:      order.Metadata,
		WorkflowID:    order.WorkflowID,
		Service:       order.Service,
		Customer:      order.Customer,
		Driver:        order.Driver,
		Points:        order.Points,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
}

