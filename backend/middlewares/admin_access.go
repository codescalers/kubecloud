package middlewares

import (
	"kubecloud/internal"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware validates requests to admin endpoints
func AdminMiddleware(tokenManager internal.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenManager.VerifyToken(tokenStr)
		if err != nil || !claims.Admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		c.Set("user_id", strconv.Itoa(claims.UserID))
		c.Set("username", claims.Username)
		c.Set("admin", claims.Admin)
		c.Next()
	}
}
