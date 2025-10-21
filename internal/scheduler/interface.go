package scheduler

import (
	"context"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// WinPowerClient defines the interface for WinPower client operations
// This interface will be used by the scheduler to trigger data collection
type WinPowerClient interface {
	// CollectDeviceData retrieves data for all devices
	CollectDeviceData(ctx context.Context) error

	// IsConnected checks if the client is connected and authenticated
	IsConnected() bool
}

// Scheduler defines the interface for the scheduler implementation
type Scheduler interface {
	// Start starts the scheduler with the given context
	// Returns an error if the scheduler is already running or cannot be started
	Start(ctx context.Context) error

	// Stop stops the scheduler gracefully
	// Waits for ongoing operations to complete within the configured timeout
	Stop() error

	// IsRunning returns true if the scheduler is currently running
	IsRunning() bool
}

// SchedulerFactory defines the interface for creating scheduler instances
type SchedulerFactory interface {
	// CreateScheduler creates a new scheduler instance with the given dependencies
	CreateScheduler(config *Config, client WinPowerClient, logger log.Logger) (Scheduler, error)
}
