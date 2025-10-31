package server

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
	Path  string `json:"path"`
	Time  string `json:"ts"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error, path string) *ErrorResponse {
	return &ErrorResponse{
		Error: err.Error(),
		Path:  path,
		Time:  time.Now().Format(time.RFC3339),
	}
}

// loggerMiddleware creates a Gin middleware for structured logging
func (s *HTTPServer) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build path with query
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request
		s.log.Info("HTTP request",
			"method", c.Request.Method,
			"path", path,
			"status", statusCode,
			"latency", latency.String(),
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// recoveryMiddleware creates a Gin middleware for panic recovery
func (s *HTTPServer) recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				s.log.Error("Panic recovered",
					"error", err,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
				)

				// Return error response
				c.JSON(500, NewErrorResponse(
					fmt.Errorf("internal server error"),
					c.Request.URL.Path,
				))

				// Abort the request
				c.Abort()
			}
		}()

		c.Next()
	}
}
