package scheduler

import (
	"errors"
	"fmt"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
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
	CollectionInterval time.Duration `json:"collection_interval" yaml:"collection_interval" env:"SCHEDULER_COLLECTION_INTERVAL" default:"5s"`

	// GracefulShutdownTimeout specifies the timeout for graceful shutdown
	// Default: 30 seconds
	GracefulShutdownTimeout time.Duration `json:"graceful_shutdown_timeout" yaml:"graceful_shutdown_timeout" env:"SCHEDULER_GRACEFUL_SHUTDOWN_TIMEOUT" default:"30s"`
}

// DefaultConfig returns a default configuration for the scheduler module
func DefaultConfig() *Config {
	return &Config{
		CollectionInterval:      5 * time.Second,
		GracefulShutdownTimeout: 30 * time.Second,
	}
}

// NewConfig creates a new scheduler configuration using the provided loader.
func NewConfig(loader *config.Loader) (*Config, error) {
	cfg := &Config{}
	cfg.SetDefaults()

	if err := loader.LoadModule("scheduler", cfg); err != nil {
		return nil, fmt.Errorf("failed to load scheduler config: %w", err)
	}

	return cfg, cfg.Validate()
}

// SetDefaults sets the default values for the scheduler configuration.
func (c *Config) SetDefaults() {
	if c.CollectionInterval == 0 {
		c.CollectionInterval = 5 * time.Second
	}
	if c.GracefulShutdownTimeout == 0 {
		c.GracefulShutdownTimeout = 30 * time.Second
	}
}

// Validate validates the configuration and returns an error if invalid
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("scheduler config cannot be nil")
	}

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

// String returns a string representation of the configuration.
// Sensitive data is masked for security.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf(
		"SchedulerConfig{CollectionInterval: %s, GracefulShutdownTimeout: %s}",
		c.CollectionInterval.String(),
		c.GracefulShutdownTimeout.String(),
	)
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() config.Config {
	if c == nil {
		return nil
	}

	return &Config{
		CollectionInterval:      c.CollectionInterval,
		GracefulShutdownTimeout: c.GracefulShutdownTimeout,
	}
}
