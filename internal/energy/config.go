package energy

import (
	"fmt"
	"time"
)

// DefaultConfig returns a default configuration for the energy module.
func DefaultConfig() *Config {
	return &Config{
		Precision:            0.01,  // 0.01 Wh precision
		EnableStats:          true,  // Enable statistics by default
		MaxCalculationTime:   1 * time.Second.Nanoseconds(),
		NegativePowerAllowed: true,  // Allow negative power for energy feedback
	}
}

// Validate checks if the configuration parameters are valid.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate precision
	if c.Precision <= 0 {
		return fmt.Errorf("precision must be positive, got: %f", c.Precision)
	}

	// Validate max calculation time (in nanoseconds)
	if c.MaxCalculationTime <= 0 {
		return fmt.Errorf("max calculation time must be positive, got: %d", c.MaxCalculationTime)
	}

	// Max calculation time should be reasonable (not too long)
	if c.MaxCalculationTime > 10*time.Second.Nanoseconds() {
		return fmt.Errorf("max calculation time should not exceed 10 seconds, got: %d", c.MaxCalculationTime)
	}

	return nil
}

// String returns a string representation of the configuration.
// Sensitive data is masked for security.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}

	return fmt.Sprintf(
		"Config{Precision: %.4f, EnableStats: %t, MaxCalculationTime: %s, NegativePowerAllowed: %t}",
		c.Precision,
		c.EnableStats,
		time.Duration(c.MaxCalculationTime).String(),
		c.NegativePowerAllowed,
	)
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	return &Config{
		Precision:            c.Precision,
		EnableStats:          c.EnableStats,
		MaxCalculationTime:   c.MaxCalculationTime,
		NegativePowerAllowed: c.NegativePowerAllowed,
	}
}

// WithPrecision sets the precision and returns the configuration for chaining.
func (c *Config) WithPrecision(precision float64) *Config {
	c.Precision = precision
	return c
}

// WithStatsEnabled enables or disables statistics collection.
func (c *Config) WithStatsEnabled(enabled bool) *Config {
	c.EnableStats = enabled
	return c
}

// WithMaxCalculationTime sets the maximum calculation time.
func (c *Config) WithMaxCalculationTime(duration time.Duration) *Config {
	c.MaxCalculationTime = duration.Nanoseconds()
	return c
}

// WithNegativePowerAllowed sets whether negative power is allowed.
func (c *Config) WithNegativePowerAllowed(allowed bool) *Config {
	c.NegativePowerAllowed = allowed
	return c
}