package server

import (
	"time"
)

// Config holds the configuration for the HTTP server
type Config struct {
	// Port is the port number the server listens on (1-65535)
	Port int `yaml:"port" validate:"min=1,max=65535"`

	// Host is the hostname or IP address to bind to
	Host string `yaml:"host" validate:"required"`

	// Mode is the Gin mode: debug, release, or test
	Mode string `yaml:"mode" validate:"oneof=debug release test"`

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration `yaml:"read_timeout" validate:"min=1s"`

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration `yaml:"write_timeout" validate:"min=1s"`

	// IdleTimeout is the maximum duration to wait for the next request when keep-alives are enabled
	IdleTimeout time.Duration `yaml:"idle_timeout" validate:"min=1s"`

	// EnablePprof enables the /debug/pprof endpoints for profiling
	EnablePprof bool `yaml:"enable_pprof"`

	// ShutdownTimeout is the maximum duration to wait for graceful shutdown
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" validate:"min=1s"`
}

// DefaultConfig returns the default server configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		Host:            "0.0.0.0",
		Mode:            "release",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnablePprof:     false,
		ShutdownTimeout: 30 * time.Second,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidConfig
	}
	if c.Host == "" {
		return ErrInvalidConfig
	}
	if c.Mode != "debug" && c.Mode != "release" && c.Mode != "test" {
		return ErrInvalidConfig
	}
	if c.ReadTimeout < time.Second {
		return ErrInvalidConfig
	}
	if c.WriteTimeout < time.Second {
		return ErrInvalidConfig
	}
	if c.IdleTimeout < time.Second {
		return ErrInvalidConfig
	}
	if c.ShutdownTimeout < time.Second {
		return ErrInvalidConfig
	}
	return nil
}
