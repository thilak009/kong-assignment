package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/pkg/log"
)



// AbortWithError sends an error response with automatic trace ID population
func AbortWithError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, models.ErrorResponse{
		Message: message,
		TraceId: log.GetRequestID(c.Request.Context()),
	})
}

// AbortWithErrorDetails sends an error response with details and automatic trace ID population
func AbortWithErrorDetails(c *gin.Context, statusCode int, errorType, message string, details interface{}) {
	c.AbortWithStatusJSON(statusCode, models.ErrorResponse{
		Type:    errorType,
		Message: message,
		TraceId: log.GetRequestID(c.Request.Context()),
		Details: details,
	})
}

// SendError sends an error response without aborting (for non-abort scenarios)
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Message: message,
		TraceId: log.GetRequestID(c.Request.Context()),
	})
}
