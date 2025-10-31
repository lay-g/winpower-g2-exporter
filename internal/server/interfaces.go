package server

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Server defines the HTTP server interface
type Server interface {
	// Start starts the HTTP server
	Start() error

	// Stop gracefully shuts down the HTTP server
	Stop(ctx context.Context) error
}

// MetricsService defines the interface for serving Prometheus metrics
type MetricsService interface {
	// HandleMetrics is the Gin handler for the /metrics endpoint
	HandleMetrics(c *gin.Context)
}

// HealthService defines the interface for health check
type HealthService interface {
	// Check performs health check and returns status and details
	Check(ctx context.Context) (status string, details map[string]any)
}

// Logger defines the minimal logging interface required by the server
type Logger interface {
	// Info logs an informational message
	Info(msg string, keysAndValues ...interface{})

	// Error logs an error message
	Error(msg string, keysAndValues ...interface{})

	// Warn logs a warning message
	Warn(msg string, keysAndValues ...interface{})

	// Debug logs a debug message
	Debug(msg string, keysAndValues ...interface{})
}
