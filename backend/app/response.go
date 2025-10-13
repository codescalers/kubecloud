package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse struct contains data returned from endpoints
type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedData is a unified container for paginated responses
type PaginatedData[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
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

// InternalServerError returns internal server error
func InternalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
		Error:   "",
	})
}
