package winpower

import (
	"errors"
	"fmt"
)

// Sentinel errors for common error conditions.
var (
	// ErrConnectionNotReady indicates the client is not ready for data collection.
	ErrConnectionNotReady = errors.New("winpower: connection not ready")

	// ErrAuthenticationFailed indicates authentication with WinPower failed.
	ErrAuthenticationFailed = errors.New("winpower: authentication failed")

	// ErrTokenExpired indicates the current token has expired.
	ErrTokenExpired = errors.New("winpower: token expired")

	// ErrInvalidResponse indicates the API response is invalid or malformed.
	ErrInvalidResponse = errors.New("winpower: invalid response")

	// ErrInvalidDeviceData indicates the device data is invalid or malformed.
	ErrInvalidDeviceData = errors.New("winpower: invalid device data")

	// ErrNetworkError indicates a network communication error.
	ErrNetworkError = errors.New("winpower: network error")

	// ErrParseError indicates a data parsing error.
	ErrParseError = errors.New("winpower: parse error")

	// ErrInvalidConfig indicates the configuration is invalid.
	ErrInvalidConfig = errors.New("winpower: invalid config")

	// ErrTimeout indicates the request timed out.
	ErrTimeout = errors.New("winpower: request timeout")
)

// AuthenticationError represents an authentication-related error.
type AuthenticationError struct {
	Message string
	Err     error
}

func (e *AuthenticationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("authentication error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("authentication error: %s", e.Message)
}

func (e *AuthenticationError) Unwrap() error {
	return e.Err
}

// NetworkError represents a network-related error.
type NetworkError struct {
	Message string
	Err     error
}

func (e *NetworkError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("network error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("network error: %s", e.Message)
}

func (e *NetworkError) Unwrap() error {
	return e.Err
}

// ParseError represents a data parsing error.
type ParseError struct {
	Field   string
	Message string
	Err     error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("parse error on field %q: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("parse error on field %q: %s", e.Field, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Field   string
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error on field %q: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("config error on field %q: %s", e.Field, e.Message)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// IsAuthenticationError checks if the error is an authentication error.
func IsAuthenticationError(err error) bool {
	var authErr *AuthenticationError
	return errors.As(err, &authErr) || errors.Is(err, ErrAuthenticationFailed) || errors.Is(err, ErrTokenExpired)
}

// IsNetworkError checks if the error is a network error.
func IsNetworkError(err error) bool {
	var netErr *NetworkError
	return errors.As(err, &netErr) || errors.Is(err, ErrNetworkError) || errors.Is(err, ErrTimeout)
}

// IsParseError checks if the error is a parse error.
func IsParseError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr) || errors.Is(err, ErrParseError)
}

// IsConfigError checks if the error is a configuration error.
func IsConfigError(err error) bool {
	var cfgErr *ConfigError
	return errors.As(err, &cfgErr) || errors.Is(err, ErrInvalidConfig)
}
