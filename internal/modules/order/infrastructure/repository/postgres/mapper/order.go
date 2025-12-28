package mapper

import (
	"go1/internal/modules/order/domain"
	"go1/internal/modules/order/infrastructure/repository/postgres/model"
)

func ToOrderDomain(m *model.OrderModel) *domain.Order {
	return &domain.Order{
		ID:         m.ID,
		UserID:     m.UserID,
		Amount:     m.Amount,
		Status:     domain.Status(m.Status),
		WorkflowID: m.WorkflowID,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
