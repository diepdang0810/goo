package shipment

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *ShipmentHandler) {
	group := r.Group("/shipment")
	{
		group.POST("/accept/:orderId", h.Accept)
		group.POST("/delivery/:orderId", h.Delivery)
	}
}
