package scheduler

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrInvalidCollectionInterval is returned when the collection interval is invalid
	ErrInvalidCollectionInterval = errors.New("collection interval cannot be zero or negative")

	// ErrInvalidGracefulShutdownTimeout is returned when the graceful shutdown timeout is invalid
	ErrInvalidGracefulShutdownTimeout = errors.New("graceful shutdown timeout cannot be zero or negative")
)

// Config represents the configuration for the scheduler module
type Config struct {
	// CollectionInterval specifies the interval between data collection cycles
	// Default: 5 seconds
	CollectionInterval time.Duration `json:"collection_interval" yaml:"collection_interval" default:"5s"`

	// GracefulShutdownTimeout specifies the timeout for graceful shutdown
	// Default: 30 seconds
	GracefulShutdownTimeout time.Duration `json:"graceful_shutdown_timeout" yaml:"graceful_shutdown_timeout" default:"30s"`
}

// DefaultConfig returns a default configuration for the scheduler module
func DefaultConfig() *Config {
	return &Config{
		CollectionInterval:      5 * time.Second,
		GracefulShutdownTimeout: 30 * time.Second,
	}
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	if c.CollectionInterval <= 0 {
		return fmt.Errorf("%w: %v", ErrInvalidCollectionInterval, c.CollectionInterval)
	}

	if c.GracefulShutdownTimeout <= 0 {
		return fmt.Errorf("%w: %v", ErrInvalidGracefulShutdownTimeout, c.GracefulShutdownTimeout)
	}

	return nil
}

// SchedulerModuleConfig represents the module configuration for registration
// This will be used for config module registration
type SchedulerModuleConfig struct {
	// Config is the scheduler configuration
	Config *Config `json:"config" yaml:"config"`
}

var (
	// Global module config instance for registration
	moduleConfig *SchedulerModuleConfig
)

// init registers the scheduler module configuration
func init() {
	// Initialize the module configuration with defaults
	moduleConfig = &SchedulerModuleConfig{
		Config: DefaultConfig(),
	}
}

// GetModuleConfig returns the global module configuration
func GetModuleConfig() *SchedulerModuleConfig {
	return moduleConfig
}

// SetModuleConfig sets the global module configuration
func SetModuleConfig(config *SchedulerModuleConfig) {
	if config != nil {
		moduleConfig = config
	}
}
