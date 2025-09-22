// Package middleware contains all HTTP middlewares for the application
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thilak009/kong-assignment/models"
	"github.com/thilak009/kong-assignment/pkg/log"
	"github.com/thilak009/kong-assignment/utils"
)

// LoggingMiddleware provides request/response logging
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := log.GetRequestID(c.Request.Context())

		// Process request
		c.Next()

		// Create context with request ID and log with start timestamp
		ctx := context.WithValue(context.Background(), log.RequestIDKey, requestID)
		duration := time.Since(start)

		loggerWithFields := log.With(ctx,
			"duration_ms", duration.Milliseconds(),
			"status_code", c.Writer.Status(),
			"response_size", c.Writer.Size(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		if c.Request.URL.RawQuery != "" {
			loggerWithFields = loggerWithFields.With(ctx, "query", c.Request.URL.RawQuery)
		}

		if len(c.Errors) > 0 {
			loggerWithFields = loggerWithFields.With(ctx, "errors", c.Errors.String())
		}

		// Log at appropriate level with method and path in message
		if c.Writer.Status() >= 500 {
			loggerWithFields.Errorf("%s %s %s %d %d", c.Request.Method, c.Request.URL.Path, c.Request.Proto, c.Writer.Status(), c.Writer.Size())
		} else if c.Writer.Status() >= 400 {
			loggerWithFields.Infof("%s %s %s %d %d", c.Request.Method, c.Request.URL.Path, c.Request.Proto, c.Writer.Status(), c.Writer.Size())
		} else {
			loggerWithFields.Infof("%s %s %s %d %d", c.Request.Method, c.Request.URL.Path, c.Request.Proto, c.Writer.Status(), c.Writer.Size())
		}
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Authorization, Accept, Client-Security-Token, Accept-Encoding, x-access-token")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			fmt.Println("OPTIONS")
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

// RequestIDMiddleware generates a unique ID and attaches it to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()

		// Set in request context (used by logger and can be retrieved by utils)
		ctx := context.WithValue(c.Request.Context(), log.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Authorization header required",
			})
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Invalid authorization header format",
			})
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate the token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Message: "Invalid token",
			})
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}
