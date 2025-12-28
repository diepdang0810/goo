package order

import (
	"net/http"

	"go1/internal/modules/order/domain"
	usecase "go1/internal/modules/order/usecase"
	"go1/pkg/response"
	"go1/pkg/utils"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	usecase usecase.OrderUsecaseInterface
}

func NewOrderHandler(uc usecase.OrderUsecaseInterface) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	// fmt.Println("CreateOrderRequest struct:", req) // Remove debug print or use logger
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleBindingError(c, err)
		return
	}

	input := usecase.CreateOrderInput{
		UserID:       req.UserID,
		Amount:       req.Amount,
		Type:         domain.ServiceType(req.Type),
		RestaurantID: req.RestaurantID,
		Items:        req.Items,
		Pickup:       req.Pickup,
		Dropoff:      req.Dropoff,
	}

	order, err := h.usecase.Create(c.Request.Context(), input)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Created(c, order)
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := utils.ParseInt(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	order, err := h.usecase.GetByID(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, order)
}
