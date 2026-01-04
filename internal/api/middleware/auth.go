package middleware

import (
	"go1/pkg/request"
	"go1/pkg/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(bypass bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bypass {
			c.Next()
			return
		}

		userID := c.GetHeader(string(utils.XUserIDHeader))
		if userID == "" {
			c.JSON(401, gin.H{"error": "Unauthorized: X-User-Id required"})
			c.Abort()
			return
		}

		user := &request.UserContext{
			UserID:    userID,
			Role:      utils.UserRole(c.GetHeader(string(utils.XUserAudienceHeader))),
			Platform:  c.GetHeader(string(utils.XUserPlatformHeader)),
			PartnerID: c.GetHeader(string(utils.XPartnerIDHeader)),
		}

		ctx := request.WithUser(c.Request.Context(), user)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
