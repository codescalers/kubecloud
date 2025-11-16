package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse struct contains data returned from endpoints
type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
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
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{
		Status:  status,
		Message: message,
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

// Accepted returns an accepted response with custom data (202)
func Accepted(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusAccepted, message, data)
}

// BadRequest returns a bad request error response (400)
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized returns an unauthorized error response (401)
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden returns a forbidden error response (403)
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound returns a not found error response (404)
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Conflict returns a conflict error response (409)
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// InternalServerError returns internal server error (500)
func InternalServerError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "Internal server error")
}

// ServiceUnavailable returns a service unavailable error response (503)
func ServiceUnavailable(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusServiceUnavailable, APIResponse{
		Status:  http.StatusServiceUnavailable,
		Message: message,
		Data:    data,
	})
}
