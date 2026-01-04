package order

import (
	"go1/internal/modules/order/application"
	"go1/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service application.OrderService
}

func NewOrderHandler(service application.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindingError(c, err)
		return
	}

	var points []application.OrderPointInput
	for _, p := range req.Points {
		points = append(points, application.OrderPointInput{
			Lat:     p.Lat,
			Lng:     p.Lng,
			Type:    p.Type,
			Address: p.Address,
			Phone:   p.Phone,
		})
	}

	input := application.CreateRideOrderInput{
		ServiceID:     req.ServiceID,
		ServiceType:   req.ServiceType,
		PaymentMethod: req.PaymentMethod,
		Points:        points,
	}

	order, err := h.service.CreateRideOrder(c.Request.Context(), input)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")

	order, err := h.service.GetByID(c.Request.Context(), idStr)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, order)
}
