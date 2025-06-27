package middlewares

import (
	"kubecloud/internal"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserMiddleware function validates users token
func UserMiddleware(tokenManager internal.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
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
