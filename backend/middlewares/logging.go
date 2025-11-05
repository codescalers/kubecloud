package middlewares

import (
	"kubecloud/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
)

// GinLoggerMiddleware creates a middleware that disables gin's default logging
// and uses our custom zerolog-based logging instead
// Skips logging for health and metrics endpoints
func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Skip logging for health and metrics endpoints
		if path == "/health" || path == "/metrics" || path == "/api/v1/health" {
			return
		}

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		// Use the shared logger which is configured with file output
		logEvent := logger.GetLogger().With().
			Str("method", method).
			Str("path", path).
			Int("status", statusCode).
			Str("ip", clientIP).
			Str("user_agent", c.Request.UserAgent()).
			Dur("latency", latency).
			Int("body_size", bodySize).
			Logger()

		// Log based on status code
		if len(c.Errors) > 0 {
			logEvent.Error().Str("errors", c.Errors.String()).Msg("Request completed with errors")
		} else if statusCode >= 500 {
			logEvent.Error().Msg("Request completed with server error")
		} else if statusCode >= 400 {
			logEvent.Warn().Msg("Request completed with client error")
		} else {
			logEvent.Debug().Msg("Request completed successfully")
		}
	}
}
