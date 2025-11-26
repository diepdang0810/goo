package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *UserHandler) {
	group := r.Group("/users")
	{
		group.POST("", h.Create)
		group.GET("", h.Fetch)
		group.GET("/:id", h.GetByID)
	}
}
