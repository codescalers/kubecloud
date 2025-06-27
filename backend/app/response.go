package app

import "github.com/gin-gonic/gin"

// APIResponse struct contains data returned from endpoints
type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success returns data for successful requests
func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	})
}

// Error returns data from failed endpoints
func Error(c *gin.Context, status int, message string, err string) {
	c.JSON(status, APIResponse{
		Status:  status,
		Message: message,
		Error:   err,
	})
}
