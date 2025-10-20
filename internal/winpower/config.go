package winpower

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Config represents the configuration for the WinPower module.
// It contains all necessary settings for connecting to and communicating
// with a WinPower G2 system instance.
type Config struct {
	// Connection settings
	URL           string        `yaml:"url" json:"url"`                         // Base URL of the WinPower API (e.g., "https://winpower.example.com:8080")
	SkipTLSVerify bool          `yaml:"skip_tls_verify" json:"skip_tls_verify"` // Skip TLS certificate verification for self-signed certificates
	Timeout       time.Duration `yaml:"timeout" json:"timeout"`                 // HTTP request timeout (recommended: 30s)
	MaxRetries    int           `yaml:"max_retries" json:"max_retries"`         // Maximum retry attempts for failed requests (recommended: 3)

	// Authentication settings
	Username string `yaml:"username" json:"username"` // Username for API authentication
	Password string `yaml:"password" json:"password"` // Password for API authentication
}

// DefaultConfig returns a default configuration for the WinPower module.
// These defaults are suitable for local development and testing.
// For production, you should provide your own configuration using WithXXX methods.
func DefaultConfig() *Config {
	return &Config{
		// Connection settings with development-friendly defaults
		URL:           "https://localhost:8080", // Default local development URL
		Timeout:       30 * time.Second,         // Reasonable timeout for most operations
		MaxRetries:    3,                        // Balanced retry strategy
		SkipTLSVerify: false,                    // Secure by default

		// Authentication settings - use defaults only for development
		Username: "admin", // Default admin user
		Password: "admin", // Default admin password - CHANGE IN PRODUCTION
	}
}

// String returns a string representation of the configuration.
// Sensitive data is masked for security.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}

	maskedPassword := maskSensitive(c.Password)

	return fmt.Sprintf(
		"Config{URL: %s, Username: %s, Password: %s, Timeout: %v, MaxRetries: %d, SkipTLSVerify: %t}",
		c.URL,
		c.Username,
		maskedPassword,
		c.Timeout,
		c.MaxRetries,
		c.SkipTLSVerify,
	)
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	return &Config{
		URL:           c.URL,
		Username:      c.Username,
		Password:      c.Password,
		Timeout:       c.Timeout,
		MaxRetries:    c.MaxRetries,
		SkipTLSVerify: c.SkipTLSVerify,
	}
}

// WithURL sets the URL and returns the configuration for chaining.
func (c *Config) WithURL(url string) *Config {
	c.URL = url
	return c
}

// WithCredentials sets the username and password and returns the configuration for chaining.
func (c *Config) WithCredentials(username, password string) *Config {
	c.Username = username
	c.Password = password
	return c
}

// WithTimeout sets the timeout and returns the configuration for chaining.
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithMaxRetries sets the max retries and returns the configuration for chaining.
func (c *Config) WithMaxRetries(maxRetries int) *Config {
	c.MaxRetries = maxRetries
	return c
}

// WithSkipTLSVerify sets the skip TLS verify flag and returns the configuration for chaining.
func (c *Config) WithSkipTLSVerify(skip bool) *Config {
	c.SkipTLSVerify = skip
	return c
}

// IsSecure returns true if the configuration uses secure settings (HTTPS and no skipped verification).
func (c *Config) IsSecure() bool {
	if c == nil {
		return false
	}
	return strings.HasPrefix(c.URL, "https://") && !c.SkipTLSVerify
}

// IsProductionReady performs basic checks to determine if the configuration is suitable for production.
func (c *Config) IsProductionReady() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	// Check for default/weak credentials
	if c.Username == "admin" && c.Password == "admin" {
		return fmt.Errorf("using default admin credentials - not suitable for production")
	}

	// Check for localhost usage
	if strings.Contains(c.URL, "localhost") || strings.Contains(c.URL, "127.0.0.1") {
		return fmt.Errorf("using localhost URL - not suitable for production")
	}

	// Check for insecure TLS settings
	if !c.IsSecure() {
		return fmt.Errorf("using insecure connection settings - not suitable for production")
	}

	return nil
}

// GetConnectionSummary returns a human-readable summary of connection settings.
func (c *Config) GetConnectionSummary() string {
	if c == nil {
		return "<nil config>"
	}

	protocol := "unknown"
	if strings.HasPrefix(c.URL, "https://") {
		protocol = "HTTPS"
	} else if strings.HasPrefix(c.URL, "http://") {
		protocol = "HTTP"
	}

	security := "secure"
	if c.SkipTLSVerify {
		security = "insecure (TLS verify skipped)"
	}

	return fmt.Sprintf("%s connection to %s (%s, timeout: %v, retries: %d)",
		protocol, c.URL, security, c.Timeout, c.MaxRetries)
}

// ValidationError represents a configuration validation error with structured information
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s (value: %s)", e.Field, e.Message, e.Value)
}

// validationCache caches validation results for performance
type validationCache struct {
	mu    sync.RWMutex
	cache map[string]bool
}

var (
	// Global validation cache for URL patterns
	globalValidationCache = &validationCache{
		cache: make(map[string]bool),
	}
)

// Validation codes for consistent error handling
const (
	ValidationCodeNilConfig       = "NIL_CONFIG"
	ValidationCodeEmptyURL        = "EMPTY_URL"
	ValidationCodeInvalidURL      = "INVALID_URL"
	ValidationCodeInvalidScheme   = "INVALID_SCHEME"
	ValidationCodeInvalidHostname = "INVALID_HOSTNAME"
	ValidationCodeEmptyUsername   = "EMPTY_USERNAME"
	ValidationCodeInvalidUsername = "INVALID_USERNAME"
	ValidationCodeEmptyPassword   = "EMPTY_PASSWORD"
	ValidationCodeInvalidTimeout  = "INVALID_TIMEOUT"
	ValidationCodeTimeoutTooShort = "TIMEOUT_TOO_SHORT"
	ValidationCodeTimeoutTooLong  = "TIMEOUT_TOO_LONG"
	ValidationCodeInvalidRetries  = "INVALID_RETRIES"
	ValidationCodeRetriesTooHigh  = "RETRIES_TOO_HIGH"
)

// validateURLScheme checks if the URL has a valid scheme with caching
func validateURLScheme(rawURL string) error {
	if rawURL == "" {
		return &ValidationError{
			Field:   "url",
			Value:   rawURL,
			Message: "URL cannot be empty",
			Code:    ValidationCodeEmptyURL,
		}
	}

	// Use cache for scheme validation
	cacheKey := "scheme:" + rawURL
	globalValidationCache.mu.RLock()
	if cached, exists := globalValidationCache.cache[cacheKey]; exists {
		globalValidationCache.mu.RUnlock()
		if cached {
			return nil
		}
		return &ValidationError{
			Field:   "url",
			Value:   rawURL,
			Message: "URL must start with http:// or https://",
			Code:    ValidationCodeInvalidScheme,
		}
	}
	globalValidationCache.mu.RUnlock()

	isValid := strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")

	// Cache the result
	globalValidationCache.mu.Lock()
	globalValidationCache.cache[cacheKey] = isValid
	globalValidationCache.mu.Unlock()

	if !isValid {
		return &ValidationError{
			Field:   "url",
			Value:   rawURL,
			Message: "URL must start with http:// or https://",
			Code:    ValidationCodeInvalidScheme,
		}
	}
	return nil
}

// validateHostname performs basic hostname validation with caching and structured errors
func validateHostname(hostname string) error {
	if hostname == "" {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname cannot be empty",
			Code:    ValidationCodeInvalidHostname,
		}
	}

	if len(hostname) > 253 {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname too long (max 253 characters)",
			Code:    ValidationCodeInvalidHostname,
		}
	}

	// Use cache for hostname validation
	cacheKey := "hostname:" + hostname
	globalValidationCache.mu.RLock()
	if cached, exists := globalValidationCache.cache[cacheKey]; exists {
		globalValidationCache.mu.RUnlock()
		if cached {
			return nil
		}
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "invalid hostname format",
			Code:    ValidationCodeInvalidHostname,
		}
	}
	globalValidationCache.mu.RUnlock()

	// Check for localhost and common loopback addresses
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		globalValidationCache.mu.Lock()
		globalValidationCache.cache[cacheKey] = true
		globalValidationCache.mu.Unlock()
		return nil
	}

	// Basic validation - no spaces or control characters
	isValid := !strings.ContainsAny(hostname, " \t\n\r")

	// Cache the result
	globalValidationCache.mu.Lock()
	globalValidationCache.cache[cacheKey] = isValid
	globalValidationCache.mu.Unlock()

	if !isValid {
		return &ValidationError{
			Field:   "hostname",
			Value:   hostname,
			Message: "hostname contains invalid characters",
			Code:    ValidationCodeInvalidHostname,
		}
	}
	return nil
}

// Validate performs comprehensive validation with caching and structured errors for better performance and debugging.
func (c *Config) Validate() error {
	if c == nil {
		return &ValidationError{
			Field:   "config",
			Value:   "nil",
			Message: "config cannot be nil",
			Code:    ValidationCodeNilConfig,
		}
	}

	// Validate URL scheme and format first (cached)
	if err := validateURLScheme(c.URL); err != nil {
		return err
	}

	// Parse URL to validate format and extract hostname
	parsedURL, err := url.Parse(c.URL)
	if err != nil {
		return &ValidationError{
			Field:   "url",
			Value:   c.URL,
			Message: fmt.Sprintf("invalid URL format: %v", err),
			Code:    ValidationCodeInvalidURL,
		}
	}

	// Validate hostname (cached)
	if err := validateHostname(parsedURL.Hostname()); err != nil {
		return err
	}

	// Validate username
	if c.Username == "" {
		return &ValidationError{
			Field:   "username",
			Value:   c.Username,
			Message: "username cannot be empty",
			Code:    ValidationCodeEmptyUsername,
		}
	}

	// Basic username validation - no special characters that could cause issues
	if strings.ContainsAny(c.Username, " \t\n\r:") {
		return &ValidationError{
			Field:   "username",
			Value:   c.Username,
			Message: "username contains invalid characters (spaces, tabs, newlines, or colons)",
			Code:    ValidationCodeInvalidUsername,
		}
	}

	// Validate password
	if c.Password == "" {
		return &ValidationError{
			Field:   "password",
			Value:   "<empty>",
			Message: "password cannot be empty",
			Code:    ValidationCodeEmptyPassword,
		}
	}

	// Validate timeout
	if c.Timeout <= 0 {
		return &ValidationError{
			Field:   "timeout",
			Value:   c.Timeout.String(),
			Message: "timeout must be positive",
			Code:    ValidationCodeInvalidTimeout,
		}
	}

	// Timeout should be reasonable (not too short, not too long)
	if c.Timeout < 5*time.Second {
		return &ValidationError{
			Field:   "timeout",
			Value:   c.Timeout.String(),
			Message: "timeout should be at least 5 seconds for reliable operation",
			Code:    ValidationCodeTimeoutTooShort,
		}
	}
	if c.Timeout > 300*time.Second {
		return &ValidationError{
			Field:   "timeout",
			Value:   c.Timeout.String(),
			Message: "timeout should not exceed 5 minutes to avoid hanging",
			Code:    ValidationCodeTimeoutTooLong,
		}
	}

	// Validate max retries
	if c.MaxRetries < 0 {
		return &ValidationError{
			Field:   "max_retries",
			Value:   fmt.Sprintf("%d", c.MaxRetries),
			Message: "max retries cannot be negative",
			Code:    ValidationCodeInvalidRetries,
		}
	}
	if c.MaxRetries > 10 {
		return &ValidationError{
			Field:   "max_retries",
			Value:   fmt.Sprintf("%d", c.MaxRetries),
			Message: "max retries should not exceed 10 to avoid excessive retry attempts",
			Code:    ValidationCodeRetriesTooHigh,
		}
	}

	return nil
}

// ValidateQuick performs basic validation without comprehensive checks for faster execution.
// Returns structured errors for consistency with the main Validate method.
func (c *Config) ValidateQuick() error {
	if c == nil {
		return &ValidationError{
			Field:   "config",
			Value:   "nil",
			Message: "config cannot be nil",
			Code:    ValidationCodeNilConfig,
		}
	}

	if c.URL == "" {
		return &ValidationError{
			Field:   "url",
			Value:   c.URL,
			Message: "URL cannot be empty",
			Code:    ValidationCodeEmptyURL,
		}
	}

	if c.Username == "" {
		return &ValidationError{
			Field:   "username",
			Value:   c.Username,
			Message: "username cannot be empty",
			Code:    ValidationCodeEmptyUsername,
		}
	}

	if c.Password == "" {
		return &ValidationError{
			Field:   "password",
			Value:   "<empty>",
			Message: "password cannot be empty",
			Code:    ValidationCodeEmptyPassword,
		}
	}

	return nil
}

// ClearValidationCache clears the global validation cache.
// Useful for testing or when validation rules change.
func ClearValidationCache() {
	globalValidationCache.mu.Lock()
	defer globalValidationCache.mu.Unlock()

	globalValidationCache.cache = make(map[string]bool)
}

// ValidateWithSummary performs comprehensive validation and returns a structured summary.
// Useful for UI applications where detailed validation feedback is needed.
func (c *Config) ValidateWithSummary() (*ValidationSummary, error) {
	if err := c.Validate(); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			return &ValidationSummary{
				IsValid: false,
				Field:   validationErr.Field,
				Code:    validationErr.Code,
				Message: validationErr.Message,
				Value:   validationErr.Value,
			}, err
		}
		return &ValidationSummary{
			IsValid: false,
			Message: err.Error(),
		}, err
	}

	return &ValidationSummary{
		IsValid:      true,
		Message:      "Configuration is valid",
		SecureStatus: c.GetConnectionSummary(),
	}, nil
}

// ValidationSummary provides a structured summary of validation results.
type ValidationSummary struct {
	IsValid      bool   `json:"is_valid"`
	Field        string `json:"field,omitempty"`
	Code         string `json:"code,omitempty"`
	Message      string `json:"message"`
	Value        string `json:"value,omitempty"`
	SecureStatus string `json:"secure_status,omitempty"`
}

// maskSensitive masks sensitive information for logging with improved security.
func maskSensitive(value string) string {
	if value == "" {
		return "<empty>"
	}
	if len(value) <= 2 {
		return strings.Repeat("*", len(value))
	}
	// Show first and last character only, with at most 3 stars in between
	maskLen := len(value) - 2
	if maskLen > 3 {
		maskLen = 3
	}
	return value[:1] + strings.Repeat("*", maskLen) + value[len(value)-1:]
}
