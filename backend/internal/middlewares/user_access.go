package middlewares

import (
	"net/http"
	"strings"

	"kubecloud/internal"

	"github.com/gin-gonic/gin"
)

// UserMiddleware function validates users token
func UserMiddleware(tokenManager internal.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		var tokenStr string

		if authHeader != "" {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Check token in query parameter for SSE or other cases
			tokenStr = c.Query("token")
			if tokenStr == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header or token query parameter missing"})
				return
			}
		}

		claims, err := tokenManager.VerifyToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("admin", claims.Admin)
		c.Next()
	}
}
