package server

import (
	"fmt"
	"time"
)

// Config defines comprehensive configuration for the HTTP server.
//
// This struct encapsulates all configurable aspects of the HTTP server
// including network settings, timeouts, middleware options, and debugging
// features. Configuration values can be provided via YAML/JSON files,
// environment variables, or programmatic construction.
//
// Default values are designed for production use with security and performance
// in mind. Debugging and development features are explicitly disabled by default.
//
// YAML example:
//
//	port: 9090
//	host: "0.0.0.0"
//	mode: "release"
//	read_timeout: "30s"
//	write_timeout: "30s"
//	idle_timeout: "60s"
//	enable_pprof: false
//	enable_cors: false
//	enable_rate_limit: false
//
// Environment variables (with WINPOWER_EXPORTER_ prefix):
//
//	WINPOWER_EXPORTER_PORT=9090
//	WINPOWER_EXPORTER_HOST=0.0.0.0
//	WINPOWER_EXPORTER_MODE=release
type Config struct {
	// Port defines the TCP port for the HTTP server to listen on.
	//
	// Valid range: 1-65535
	// Default: 9090
	// Common values:
	//   - 9090: Default Prometheus metrics port
	//   - 8080: Standard HTTP port for development
	//   - 80: Standard HTTP port (requires root privileges)
	//
	// Security considerations:
	//   - Avoid using privileged ports (< 1024) unless necessary
	//   - Use different ports for different environments
	//   - Ensure firewall rules allow the chosen port
	Port int `yaml:"port" json:"port"`

	// Host defines the network interface for the HTTP server to bind to.
	//
	// Default: "0.0.0.0" (all interfaces)
	// Common values:
	//   - "0.0.0.0": Listen on all available interfaces
	//   - "127.0.0.1": Listen only on localhost (development)
	//   - "192.168.1.100": Listen on specific IP address
	//   - "::": Listen on all IPv6 interfaces
	//
	// Security considerations:
	//   - Use "127.0.0.1" in development to limit access
	//   - Use "0.0.0.0" in production behind a load balancer
	//   - Avoid binding to specific IPs in cloud environments
	Host string `yaml:"host" json:"host"`

	// Mode defines the Gin framework runtime mode.
	//
	// Valid values: "debug", "release", "test"
	// Default: "release"
	//
	// Mode behaviors:
	//   - "debug": Detailed logging, panic stack traces, development features
	//   - "release": Optimized performance, minimal logging, production ready
	//   - "test": Testing mode with mocked dependencies
	//
	// Recommendations:
	//   - Use "debug" during development and troubleshooting
	//   - Use "release" in production environments
	//   - Use "test" for automated testing
	Mode string `yaml:"mode" json:"mode"`

	// ReadTimeout defines the maximum duration for reading the entire request.
	//
	// This timeout includes the time to read request headers and body.
	// It helps prevent slowloris attacks and resource exhaustion.
	//
	// Default: 30 seconds
	// Recommended range: 10-60 seconds
	//
	// Considerations:
	//   - Larger values support file uploads and slow clients
	//   - Smaller values improve protection against attacks
	//   - Should be longer than expected request processing time
	//   - Set to 0 for no timeout (not recommended)
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout"`

	// WriteTimeout defines the maximum duration for writing the response.
	//
	// This timeout controls how long the server waits for the client
	// to acknowledge the response before closing the connection.
	//
	// Default: 30 seconds
	// Recommended range: 10-60 seconds
	//
	// Considerations:
	//   - Should accommodate slow client connections
	//   - Larger values support large response payloads
	//   - Smaller values prevent connection hijacking
	//   - Set to 0 for no timeout (not recommended)
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`

	// IdleTimeout defines the maximum time to wait for the next request.
	//
	// This timeout applies to keep-alive connections. If no request
	// is received within this period, the connection is closed.
	//
	// Default: 60 seconds
	// Recommended range: 30-300 seconds
	//
	// Benefits:
	//   - Prevents connection exhaustion
	//   - Frees up server resources
	//   - Encourages efficient connection reuse
	//   - Balances performance and resource usage
	IdleTimeout time.Duration `yaml:"idle_timeout" json:"idle_timeout"`

	// EnablePprof enables Go's built-in pprof profiling endpoints.
	//
	// When enabled, provides access to /debug/pprof/ endpoints for
	// performance analysis and debugging. This should be disabled
	// in production environments due to security concerns.
	//
	// Default: false
	// Endpoints when enabled:
	//   - /debug/pprof/: Profiling index page
	//   - /debug/pprof/cmdline: Command line arguments
	//   - /debug/pprof/profile: CPU profiling
	//   - /debug/pprof/symbol: Symbol lookup
	//   - /debug/pprof/trace: Execution tracing
	//
	// Security considerations:
	//   - Exposes internal application details
	//   - Can be used for performance analysis attacks
	//   - Should be behind authentication in production
	//   - Consider enabling only temporarily for debugging
	EnablePprof bool `yaml:"enable_pprof" json:"enable_pprof"`

	// EnableCORS enables Cross-Origin Resource Sharing middleware.
	//
	// When enabled, the server will handle CORS preflight requests
	// and add appropriate CORS headers to responses. This is useful
	// when the server is accessed from web browsers on different domains.
	//
	// Default: false
	//
	// CORS features when enabled:
	//   - Handles OPTIONS preflight requests
	//   - Adds Access-Control-Allow-Origin headers
	//   - Supports configurable allowed origins
	//   - Can be customized for specific use cases
	//
	// Use cases:
	//   - Web dashboards accessing the metrics endpoint
	//   - Development with frontend on different port
	//   - API access from browser-based tools
	//
	// Security considerations:
	//   - Can expose the API to any website
	//   - Should be configured with specific origins in production
	//   - Consider authentication alternatives for sensitive data
	EnableCORS bool `yaml:"enable_cors" json:"enable_cors"`

	// EnableRateLimit enables request rate limiting middleware.
	//
	// When enabled, the server will limit the number of requests per
	// client IP address to prevent abuse and ensure fair resource usage.
	// This helps protect against DoS attacks and resource exhaustion.
	//
	// Default: false
	//
	// Rate limiting features when enabled:
	//   - Limits requests per IP address
	//   - Returns 429 status code when limits exceeded
	//   - Includes Retry-After headers
	//   - Uses token bucket algorithm
	//
	// Benefits:
	//   - Prevents request flooding attacks
	//   - Ensures fair resource allocation
	//   - Reduces server load during traffic spikes
	//   - Provides graceful degradation under load
	//
	// Configuration:
	//   - Rate limits are currently hardcoded
	//   - Future versions may support configurable limits
	//   - Consider using reverse proxy rate limiting for production
	EnableRateLimit bool `yaml:"enable_rate_limit" json:"enable_rate_limit"`
}

// DefaultConfig returns a server configuration with production-ready defaults.
//
// This function provides a pre-configured Config instance with sensible
// default values optimized for production use. The defaults prioritize
// security, performance, and stability while following common conventions.
//
// Default values and rationale:
//   - Port: 9090 - Standard Prometheus metrics port
//   - Host: "0.0.0.0" - Listen on all interfaces for production deployment
//   - Mode: "release" - Optimized for production performance
//   - ReadTimeout: 30s - Sufficient for most HTTP requests
//   - WriteTimeout: 30s - Balanced for response delivery
//   - IdleTimeout: 60s - Reasonable keep-alive timeout
//   - EnablePprof: false - Security: disable debugging in production
//   - EnableCORS: false - Security: disable unless explicitly needed
//   - EnableRateLimit: false - Performance: disable unless needed
//
// Usage:
//
//	config := server.DefaultConfig()
//	config.Port = 8080  // Override specific values as needed
//	config.EnablePprof = true  // Enable debugging features
//
// Returns:
//   - *Config: A new configuration instance with default values
//
// Note: The returned Config can be safely modified after creation.
//
//	Changes should be made before creating the HTTP server.
func DefaultConfig() *Config {
	return &Config{
		Port:            9090,
		Host:            "0.0.0.0",
		Mode:            "release",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnablePprof:     false,
		EnableCORS:      false,
		EnableRateLimit: false,
	}
}

// Validate checks the configuration for validity and returns detailed errors.
//
// This method performs comprehensive validation of all configuration fields
// to ensure they meet security, performance, and functionality requirements.
// Validation failures include specific error messages to help with debugging
// and configuration corrections.
//
// Validation rules:
//   - Port: Must be between 1 and 65535 (valid TCP port range)
//   - Host: Must not be empty string
//   - Mode: Must be one of "debug", "release", "test"
//   - ReadTimeout: Must be positive duration (> 0)
//   - WriteTimeout: Must be positive duration (> 0)
//   - IdleTimeout: Must be positive duration (> 0)
//
// Error handling:
//   - Each validation failure returns a specific, descriptive error
//   - Multiple validation errors are reported one at a time
//   - Error messages include both the problematic value and expected range
//
// Usage:
//
//	config := server.DefaultConfig()
//	config.Port = 70000  // Invalid port
//
//	if err := config.Validate(); err != nil {
//	    log.Fatalf("Invalid configuration: %v", err)
//	    // Output: "Invalid configuration: invalid port: 70000 (must be between 1 and 65535)"
//	}
//
// Returns:
//   - error: nil if configuration is valid, or descriptive error if invalid
//
// Best practices:
//   - Call Validate() before creating HTTP servers
//   - Handle validation errors gracefully in application startup
//   - Log validation errors for debugging configuration issues
//   - Consider using environment variables for configuration overrides
func (c *Config) Validate() error {
	// Validate port range (1-65535)
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1 and 65535)", c.Port)
	}

	// Validate host is not empty
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Validate mode is one of allowed values
	validModes := map[string]bool{
		"debug":   true,
		"release": true,
		"test":    true,
	}
	if !validModes[c.Mode] {
		return fmt.Errorf("invalid mode: %s (must be one of: debug, release, test)", c.Mode)
	}

	// Validate timeouts are positive
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive (got: %v)", c.ReadTimeout)
	}
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive (got: %v)", c.WriteTimeout)
	}
	if c.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive (got: %v)", c.IdleTimeout)
	}

	return nil
}
