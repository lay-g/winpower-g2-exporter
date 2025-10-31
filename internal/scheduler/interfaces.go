package scheduler

import (
	"context"
)

// Scheduler defines the interface for scheduling periodic data collection.
type Scheduler interface {
	// Start starts the scheduler and begins triggering data collection at configured intervals.
	// It returns an error if the scheduler is already running or if initialization fails.
	// The provided context can be used to control the scheduler's lifecycle.
	Start(ctx context.Context) error

	// Stop stops the scheduler gracefully, waiting for the current collection cycle to complete.
	// It returns an error if the scheduler is not running or if graceful shutdown times out.
	// The provided context can be used to set a deadline for the stop operation.
	Stop(ctx context.Context) error
}

// CollectorInterface defines the interface for data collection operations.
// This interface is based on the collector module's CollectorInterface.
type CollectorInterface interface {
	// CollectDeviceData collects device data and triggers energy calculations.
	CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}

// CollectionResult represents the result of a data collection operation.
// This is a simplified version for scheduler's needs.
type CollectionResult struct {
	Success      bool   `json:"success"`
	DeviceCount  int    `json:"device_count"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Logger defines the interface for structured logging.
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}
