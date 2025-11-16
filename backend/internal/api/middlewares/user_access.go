package middlewares

import (
	"net/http"
	"strings"

	"kubecloud/internal/auth"

	"github.com/gin-gonic/gin"
)

// UserMiddleware function validates users token
func UserMiddleware(tokenManager auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			// TrimPrefix didn't work, meaning it doesn't start with "Bearer "
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
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
