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

// OK returns a successful response with data (200)
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// Created returns a created response with data (201)
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// Accepted returns an accepted response for async operations (202)
func Accepted(c *gin.Context, message string, workflowID string, status string) {
	c.JSON(http.StatusAccepted, Response{
		WorkflowID: workflowID,
		Status:     status,
		Message:    message,
	})
}

// AcceptedWithData returns an accepted response with custom data (202)
func AcceptedWithData(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusAccepted, message, data)
}

// BadRequest returns a bad request error response (400)
func BadRequest(c *gin.Context, error string) {
	Error(c, http.StatusBadRequest, "Bad Request", error)
}

// Unauthorized returns an unauthorized error response (401)
func Unauthorized(c *gin.Context, error string) {
	Error(c, http.StatusUnauthorized, "Unauthorized", error)
}

// Forbidden returns a forbidden error response (403)
func Forbidden(c *gin.Context, error string) {
	Error(c, http.StatusForbidden, "Forbidden", error)
}

// NotFound returns a not found error response (404)
func NotFound(c *gin.Context, error string) {
	Error(c, http.StatusNotFound, "Resource not found", error)
}

// Conflict returns a conflict error response (409)
func Conflict(c *gin.Context, error string) {
	Error(c, http.StatusConflict, "Conflict", error)
}

// InternalServerError returns internal server error (500)
func InternalServerError(c *gin.Context, error string) {
	Error(c, http.StatusInternalServerError, "Internal server error", error)
}

// JSONResponse returns a raw JSON response (for special cases like health checks)
// Use this when you need to return a custom structure that doesn't fit APIResponse format
func JSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}
