package order

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *OrderHandler) {
	group := r.Group("/orders")
	{
		group.POST("", h.Create)
		group.GET("/:id", h.GetByID)
	}
}
