package scheduler

import (
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// DefaultSchedulerFactory implements the SchedulerFactory interface
type DefaultSchedulerFactory struct{}

// NewDefaultSchedulerFactory creates a new DefaultSchedulerFactory instance
func NewDefaultSchedulerFactory() *DefaultSchedulerFactory {
	return &DefaultSchedulerFactory{}
}

// CreateScheduler creates a new scheduler instance with the given dependencies
func (f *DefaultSchedulerFactory) CreateScheduler(config *Config, client WinPowerClient, logger log.Logger) (Scheduler, error) {
	// Validate config
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create scheduler instance
	scheduler, err := NewDefaultScheduler(config, client, logger)
	if err != nil {
		return nil, err
	}

	return scheduler, nil
}

// DefaultFactory is the default factory instance
var DefaultFactory = NewDefaultSchedulerFactory()

// CreateScheduler is a convenience function that uses the default factory
func CreateScheduler(config *Config, client WinPowerClient, logger log.Logger) (Scheduler, error) {
	return DefaultFactory.CreateScheduler(config, client, logger)
}
