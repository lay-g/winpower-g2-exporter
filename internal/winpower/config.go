package winpower

import (
	"fmt"
	"net/url"
	"time"
)

// Config holds the configuration for WinPower client.
type Config struct {
	// BaseURL is the base URL of the WinPower system (e.g., "https://winpower.example.com")
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`

	// Username for authentication
	Username string `yaml:"username" mapstructure:"username"`

	// Password for authentication
	Password string `yaml:"password" mapstructure:"password"`

	// Timeout for HTTP requests
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`

	// SkipSSLVerify skips SSL certificate verification (for self-signed certificates)
	SkipSSLVerify bool `yaml:"skip_ssl_verify" mapstructure:"skip_ssl_verify"`

	// RefreshThreshold is the time before expiration to refresh the token
	RefreshThreshold time.Duration `yaml:"refresh_threshold" mapstructure:"refresh_threshold"`

	// UserAgent is the User-Agent header for HTTP requests
	UserAgent string `yaml:"user_agent" mapstructure:"user_agent"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Timeout:          15 * time.Second,
		SkipSSLVerify:    false,
		RefreshThreshold: 5 * time.Minute,
		UserAgent:        "Mozilla/5.0 (compatible; WinPower-Exporter/1.0)",
	}
}

// Validate validates the configuration and returns an error if invalid.
// This implements the ConfigValidator interface for integration with the config module.
func (c *Config) Validate() error {
	// Validate BaseURL
	if c.BaseURL == "" {
		return &ConfigError{
			Field:   "base_url",
			Message: "cannot be empty",
		}
	}

	// Validate BaseURL format
	parsedURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return &ConfigError{
			Field:   "base_url",
			Message: "invalid URL format",
			Err:     err,
		}
	}

	// Ensure BaseURL has a valid scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &ConfigError{
			Field:   "base_url",
			Message: fmt.Sprintf("invalid URL scheme %q, must be http or https", parsedURL.Scheme),
		}
	}

	// Validate username
	if c.Username == "" {
		return &ConfigError{
			Field:   "username",
			Message: "cannot be empty",
		}
	}

	// Validate password
	if c.Password == "" {
		return &ConfigError{
			Field:   "password",
			Message: "cannot be empty",
		}
	}

	// Validate timeout
	if c.Timeout <= 0 {
		return &ConfigError{
			Field:   "timeout",
			Message: fmt.Sprintf("must be positive, got %v", c.Timeout),
		}
	}

	// Validate refresh threshold
	if c.RefreshThreshold < time.Minute {
		return &ConfigError{
			Field:   "refresh_threshold",
			Message: fmt.Sprintf("must be at least 1 minute, got %v", c.RefreshThreshold),
		}
	}

	// Ensure refresh threshold is less than token expiry (1 hour)
	if c.RefreshThreshold >= time.Hour {
		return &ConfigError{
			Field:   "refresh_threshold",
			Message: fmt.Sprintf("must be less than 1 hour, got %v", c.RefreshThreshold),
		}
	}

	return nil
}

// WithDefaults fills in missing optional fields with default values.
func (c *Config) WithDefaults() *Config {
	defaults := DefaultConfig()

	if c.Timeout == 0 {
		c.Timeout = defaults.Timeout
	}

	if c.RefreshThreshold == 0 {
		c.RefreshThreshold = defaults.RefreshThreshold
	}

	if c.UserAgent == "" {
		c.UserAgent = defaults.UserAgent
	}

	return c
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	return &Config{
		BaseURL:          c.BaseURL,
		Username:         c.Username,
		Password:         c.Password,
		Timeout:          c.Timeout,
		SkipSSLVerify:    c.SkipSSLVerify,
		RefreshThreshold: c.RefreshThreshold,
		UserAgent:        c.UserAgent,
	}
}

// Sanitize returns a copy of the config with sensitive fields masked for logging.
func (c *Config) Sanitize() map[string]interface{} {
	return map[string]interface{}{
		"base_url":          c.BaseURL,
		"username":          c.Username,
		"password":          "***REDACTED***",
		"timeout":           c.Timeout.String(),
		"skip_ssl_verify":   c.SkipSSLVerify,
		"refresh_threshold": c.RefreshThreshold.String(),
		"user_agent":        c.UserAgent,
	}
}
