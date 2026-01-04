package middleware

import (
	"go1/internal/pkg/request"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for x-user-id header
func AuthMiddleware(bypass bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bypass {
			c.Next()
			return
		}

		// Check x-user-id header
		userID := c.GetHeader("x-user-id")
		if userID == "" {
			c.JSON(401, gin.H{"error": "Unauthorized: x-user-id required"})
			c.Abort()
			return
		}

		user := &request.UserContext{
			UserID:   userID,
			Audience: c.GetHeader("x-user-audience"),
			Platform: c.GetHeader("x-platform"),
		}

		// Set context
		ctx := request.WithUser(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
