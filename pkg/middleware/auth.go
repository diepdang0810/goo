package middleware

import (
	"strings"

	"go1/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware checks for authentication (bypassed by default)
func AuthMiddleware(bypass bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bypass {
			// Bypass authentication
			c.Next()
			return
		}

		// Check Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		
		// TODO: Validate JWT token here
		// For now, just log and continue
		logger.Log.Info("Auth token received", logger.Field{Key: "token", Value: token[:10] + "..."})

		c.Next()
	}
}
