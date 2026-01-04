package mapper

import (
	"encoding/json"
	"go1/internal/shared/order/domain/entity"
	"go1/internal/shared/order/infrastructure/repository/postgres/model"
)

func ToOrderDomain(m *model.OrderModel) *entity.RideOrderEntity {
	var metadata map[string]interface{}
	if len(m.Metadata) > 0 {
		_ = json.Unmarshal(m.Metadata, &metadata)
	}
	return &entity.RideOrderEntity{
		ID:            m.ID,
		CreatedBy:     m.CreatedBy,
		Status:        entity.OrderStatus(m.Status),
		SubStatus:     m.SubStatus,
		PromotionCode: m.PromotionCode,
		FeeID:         m.FeeID,
		HasInsurance:  m.HasInsurance,
		OrderTime:     m.OrderTime,
		CompletedTime: m.CompletedTime,
		CancelTime:    m.CancelTime,
		Platform:      m.Platform,
		IsSchedule:    m.IsSchedule,
		NowOrder:      m.NowOrder,
		NowOrderCode:  m.NowOrderCode,
		Payment:       entity.PaymentVO{Method: m.PaymentMethod}, // Assume simple mapping for now
		Metadata:      metadata,
		WorkflowID:    m.WorkflowID,
		Service: entity.ServiceVO{
			ID:   m.ServiceID,
			Type: m.ServiceType,
			Name: m.ServiceName,
		},
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
