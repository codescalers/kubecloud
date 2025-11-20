package middlewares

import (
	"net/http"
	"strings"

	"kubecloud/internal/auth"

	"github.com/gin-gonic/gin"
)

// UserMiddleware function validates users token from Authorization header
// Used for regular API endpoints
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

// SSEMiddleware validates user token for Server-Sent Events connections
// Accepts token from query parameter since EventSource cannot set custom headers
func SSEMiddleware(tokenManager auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from Authorization header first (for clients that support it)
		authHeader := c.GetHeader("Authorization")
		var tokenStr string

		if authHeader != "" {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				// TrimPrefix didn't work, try query parameter instead
				tokenStr = c.Query("token")
			}
		} else {
			// Fall back to query parameter for EventSource connections
			tokenStr = c.Query("token")
		}

		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization token missing"})
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
