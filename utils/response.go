package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/models"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// RequestIDKey is the key for storing request ID in context
	RequestIDKey ContextKey = "request_id"
)

// GetRequestID extracts the request ID from gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(string(RequestIDKey)); exists {
		return requestID.(string)
	}
	return ""
}

// AbortWithError sends an error response with automatic trace ID population
func AbortWithError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, models.ErrorResponse{
		Message: message,
		TraceId: GetRequestID(c),
	})
}

// AbortWithErrorDetails sends an error response with details and automatic trace ID population
func AbortWithErrorDetails(c *gin.Context, statusCode int, errorType, message string, details interface{}) {
	c.AbortWithStatusJSON(statusCode, models.ErrorResponse{
		Type:    errorType,
		Message: message,
		TraceId: GetRequestID(c),
		Details: details,
	})
}

// SendError sends an error response without aborting (for non-abort scenarios)
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Message: message,
		TraceId: GetRequestID(c),
	})
}
