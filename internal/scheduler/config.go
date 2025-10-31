package scheduler

import (
	"fmt"
	"time"
)

// Config defines the configuration for the scheduler module.
type Config struct {
	// CollectionInterval is the interval between data collection cycles.
	// Default: 5 seconds
	CollectionInterval time.Duration `yaml:"collection_interval" json:"collection_interval"`

	// GracefulShutdownTimeout is the maximum time to wait for graceful shutdown.
	// Default: 5 seconds
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		CollectionInterval:      5 * time.Second,
		GracefulShutdownTimeout: 5 * time.Second,
	}
}

// Validate validates the configuration values.
func (c *Config) Validate() error {
	if c.CollectionInterval <= 0 {
		return fmt.Errorf("collection_interval must be positive, got: %v", c.CollectionInterval)
	}

	if c.GracefulShutdownTimeout <= 0 {
		return fmt.Errorf("graceful_shutdown_timeout must be positive, got: %v", c.GracefulShutdownTimeout)
	}

	// Minimum interval constraint (prevent too frequent collections)
	minInterval := 1 * time.Second
	if c.CollectionInterval < minInterval {
		return fmt.Errorf("collection_interval must be at least %v, got: %v", minInterval, c.CollectionInterval)
	}

	// Maximum interval constraint (ensure reasonable collection frequency)
	maxInterval := 1 * time.Hour
	if c.CollectionInterval > maxInterval {
		return fmt.Errorf("collection_interval must not exceed %v, got: %v", maxInterval, c.CollectionInterval)
	}

	return nil
}
