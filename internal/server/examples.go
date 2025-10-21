package server

import (
	"time"
)

// This file contains configuration examples for different deployment scenarios.
// These examples demonstrate best practices for various use cases and environments.

// ExampleConfigDevelopment returns a configuration optimized for development environments.
//
// This configuration enables debugging features, uses localhost binding,
// and implements shorter timeouts for rapid development iteration.
//
// Features:
//   - Debug mode with detailed logging
//   - Pprof profiling enabled for performance analysis
//   - CORS enabled for frontend development
//   - Localhost-only binding for security
//   - Shorter timeouts for faster iteration
//
// Usage:
//
//	config := ExampleConfigDevelopment()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigDevelopment() *Config {
	return &Config{
		Port:            8080,             // Development port
		Host:            "127.0.0.1",      // Localhost only
		Mode:            "debug",          // Debug mode with detailed logging
		ReadTimeout:     10 * time.Second, // Shorter timeouts for development
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     30 * time.Second,
		EnablePprof:     true,  // Enable debugging
		EnableCORS:      true,  // Enable for frontend development
		EnableRateLimit: false, // Disable for development convenience
	}
}

// ExampleConfigProduction returns a configuration optimized for production environments.
//
// This configuration prioritizes security, performance, and stability
// while following production deployment best practices.
//
// Features:
//   - Release mode for optimal performance
//   - Standard Prometheus port (9090)
//   - All interfaces binding for load balancer access
//   - Production-ready timeouts
//   - Security features (debugging disabled)
//   - Rate limiting enabled for protection
//
// Usage:
//
//	config := ExampleConfigProduction()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigProduction() *Config {
	return &Config{
		Port:            9090,             // Standard Prometheus port
		Host:            "0.0.0.0",        // All interfaces for load balancer
		Mode:            "release",        // Optimized performance
		ReadTimeout:     30 * time.Second, // Production-ready timeouts
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnablePprof:     false, // Security: disable debugging
		EnableCORS:      false, // Security: disable unless needed
		EnableRateLimit: true,  // Enable protection
	}
}

// ExampleConfigHighSecurity returns a configuration optimized for high-security environments.
//
// This configuration implements the most restrictive settings suitable
// for environments requiring maximum security controls.
//
// Features:
//   - Localhost-only binding
//   - Very short timeouts to prevent attacks
//   - All optional features disabled
//   - Rate limiting enabled
//   - Release mode for security
//
// Usage:
//
//	config := ExampleConfigHighSecurity()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigHighSecurity() *Config {
	return &Config{
		Port:            9090,
		Host:            "127.0.0.1",      // Restrict to localhost
		Mode:            "release",        // Production mode
		ReadTimeout:     15 * time.Second, // Shorter timeouts for security
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     30 * time.Second,
		EnablePprof:     false, // No debugging in production
		EnableCORS:      false, // No cross-origin access
		EnableRateLimit: true,  // Protection against abuse
	}
}

// ExampleConfigContainer returns a configuration optimized for container/Docker environments.
//
// This configuration accounts for container networking considerations
// and typical deployment patterns in orchestrated environments.
//
// Features:
//   - All interfaces binding for container networking
//   - Standard Prometheus port
//   - Moderate timeouts for container environments
//   - Release mode for performance
//   - Rate limiting disabled (handled at orchestration level)
//
// Usage:
//
//	config := ExampleConfigContainer()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigContainer() *Config {
	return &Config{
		Port:            9090,
		Host:            "0.0.0.0", // All interfaces for container networking
		Mode:            "release",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     45 * time.Second, // Shorter for container environments
		EnablePprof:     false,
		EnableCORS:      false,
		EnableRateLimit: false, // Handle at reverse proxy level
	}
}

// ExampleConfigHighThroughput returns a configuration optimized for high-throughput scenarios.
//
// This configuration is tuned for environments handling large volumes
// of requests with emphasis on performance and resource efficiency.
//
// Features:
//   - Shorter timeouts for faster request turnover
//   - Release mode for maximum performance
//   - Rate limiting enabled to prevent abuse
//   - Optimized keep-alive timeout
//
// Usage:
//
//	config := ExampleConfigHighThroughput()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigHighThroughput() *Config {
	return &Config{
		Port:            9090,
		Host:            "0.0.0.0",
		Mode:            "release",
		ReadTimeout:     10 * time.Second, // Faster request processing
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     30 * time.Second, // Shorter keep-alive for connection reuse
		EnablePprof:     false,
		EnableCORS:      false,
		EnableRateLimit: true, // Prevent abuse under high load
	}
}

// ExampleConfigLargePayloads returns a configuration optimized for handling large request/response payloads.
//
// This configuration extends timeouts to accommodate file uploads,
// large metrics outputs, or slow client connections.
//
// Features:
//   - Extended read/write timeouts for large payloads
//   - Longer idle timeout for slow clients
//   - Release mode for stable performance
//   - All optional features disabled for simplicity
//
// Usage:
//
//	config := ExampleConfigLargePayloads()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigLargePayloads() *Config {
	return &Config{
		Port:            9090,
		Host:            "0.0.0.0",
		Mode:            "release",
		ReadTimeout:     60 * time.Second,  // Support large uploads
		WriteTimeout:    120 * time.Second, // Support large downloads
		IdleTimeout:     300 * time.Second, // Longer keep-alive for slow clients
		EnablePprof:     false,
		EnableCORS:      false,
		EnableRateLimit: false,
	}
}

// ExampleConfigTesting returns a configuration optimized for automated testing environments.
//
// This configuration uses test mode and minimal timeouts to ensure
// fast test execution while maintaining functionality.
//
// Features:
//   - Test mode for mocked dependencies
//   - Minimal timeouts for fast test execution
//   - Localhost binding for test isolation
//   - All optional features disabled
//
// Usage:
//
//	config := ExampleConfigTesting()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigTesting() *Config {
	return &Config{
		Port:            0,               // Random port for test isolation
		Host:            "127.0.0.1",     // Localhost for testing
		Mode:            "test",          // Testing mode
		ReadTimeout:     5 * time.Second, // Minimal timeouts for fast tests
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     10 * time.Second,
		EnablePprof:     false,
		EnableCORS:      false,
		EnableRateLimit: false,
	}
}

// ExampleConfigBehindProxy returns a configuration optimized for deployment behind a reverse proxy.
//
// This configuration assumes TLS termination and request routing
// are handled by a reverse proxy (nginx, Traefik, etc.).
//
// Features:
//   - HTTP-only (TLS handled by proxy)
//   - Standard configuration for proxy deployment
//   - Rate limiting disabled (handled by proxy)
//   - CORS disabled (handled by proxy)
//
// Usage:
//
//	config := ExampleConfigBehindProxy()
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigBehindProxy() *Config {
	return &Config{
		Port:            9090,
		Host:            "0.0.0.0", // Accessible to proxy
		Mode:            "release",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		EnablePprof:     false, // Security: disable behind proxy
		EnableCORS:      false, // Handled by proxy
		EnableRateLimit: false, // Handled by proxy
	}
}

// ExampleConfigMinimal returns a minimal configuration with only essential settings.
//
// This configuration provides the most basic setup suitable for
// simple deployments or as a starting point for customization.
//
// Features:
//   - Default configuration with minimal modifications
//   - Security-focused (all optional features disabled)
//   - Standard production settings
//
// Usage:
//
//	config := ExampleConfigMinimal()
//	// Customize as needed
//	config.Port = 8080
//	srv := NewHTTPServer(config, logger, metrics, health)
func ExampleConfigMinimal() *Config {
	config := DefaultConfig()
	// Ensure all optional features are disabled for minimal setup
	config.EnablePprof = false
	config.EnableCORS = false
	config.EnableRateLimit = false
	return config
}

// ConfigurationBuilder provides a fluent interface for building configurations.
//
// This builder pattern allows for more readable and flexible configuration
// construction, especially when combining multiple options.
//
// Example:
//
//	config := NewConfigurationBuilder().
//	    WithPort(8080).
//	    WithHost("127.0.0.1").
//	    WithDebugMode().
//	    WithPprof().
//	    Build()
type ConfigurationBuilder struct {
	config *Config
}

// NewConfigurationBuilder creates a new configuration builder with default values.
func NewConfigurationBuilder() *ConfigurationBuilder {
	return &ConfigurationBuilder{
		config: DefaultConfig(),
	}
}

// WithPort sets the server port.
func (b *ConfigurationBuilder) WithPort(port int) *ConfigurationBuilder {
	b.config.Port = port
	return b
}

// WithHost sets the server host.
func (b *ConfigurationBuilder) WithHost(host string) *ConfigurationBuilder {
	b.config.Host = host
	return b
}

// WithMode sets the server mode.
func (b *ConfigurationBuilder) WithMode(mode string) *ConfigurationBuilder {
	b.config.Mode = mode
	return b
}

// WithTimeouts sets all timeout values.
func (b *ConfigurationBuilder) WithTimeouts(read, write, idle time.Duration) *ConfigurationBuilder {
	b.config.ReadTimeout = read
	b.config.WriteTimeout = write
	b.config.IdleTimeout = idle
	return b
}

// WithDebugMode enables debug mode.
func (b *ConfigurationBuilder) WithDebugMode() *ConfigurationBuilder {
	b.config.Mode = "debug"
	return b
}

// WithReleaseMode enables release mode.
func (b *ConfigurationBuilder) WithReleaseMode() *ConfigurationBuilder {
	b.config.Mode = "release"
	return b
}

// WithPprof enables or disables pprof.
func (b *ConfigurationBuilder) WithPprof(enabled bool) *ConfigurationBuilder {
	b.config.EnablePprof = enabled
	return b
}

// WithCORS enables or disables CORS.
func (b *ConfigurationBuilder) WithCORS(enabled bool) *ConfigurationBuilder {
	b.config.EnableCORS = enabled
	return b
}

// WithRateLimit enables or disables rate limiting.
func (b *ConfigurationBuilder) WithRateLimit(enabled bool) *ConfigurationBuilder {
	b.config.EnableRateLimit = enabled
	return b
}

// WithSecurity enables security-focused settings.
func (b *ConfigurationBuilder) WithSecurity() *ConfigurationBuilder {
	b.config.EnablePprof = false
	b.config.EnableCORS = false
	b.config.EnableRateLimit = true
	b.config.Mode = "release"
	return b
}

// WithPerformance enables performance-focused settings.
func (b *ConfigurationBuilder) WithPerformance() *ConfigurationBuilder {
	b.config.Mode = "release"
	b.config.ReadTimeout = 10 * time.Second
	b.config.WriteTimeout = 10 * time.Second
	b.config.IdleTimeout = 30 * time.Second
	return b
}

// Build creates the final configuration and validates it.
func (b *ConfigurationBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, err
	}
	return b.config, nil
}

// MustBuild creates the final configuration and panics on validation errors.
//
// This method is useful for configuration that should be correct by design
// and where validation errors indicate programming errors.
func (b *ConfigurationBuilder) MustBuild() *Config {
	config, err := b.Build()
	if err != nil {
		panic(err)
	}
	return config
}

// ExampleUsage demonstrates how to use the configuration builder.
func ExampleUsage() *Config {
	// Using builder pattern for complex configuration
	config, err := NewConfigurationBuilder().
		WithPort(8080).
		WithHost("127.0.0.1").
		WithDebugMode().
		WithTimeouts(15*time.Second, 15*time.Second, 45*time.Second).
		WithPprof(true).
		WithCORS(true).
		Build()

	if err != nil {
		panic(err)
	}

	return config
}
