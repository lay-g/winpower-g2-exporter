package winpower

import (
	"fmt"
	"time"
)

// ErrorType represents different types of errors that can occur in the WinPower module
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown error
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeConfig represents a configuration error
	ErrorTypeConfig
	// ErrorTypeConnection represents a connection-related error
	ErrorTypeConnection
	// ErrorTypeAuth represents an authentication error
	ErrorTypeAuth
	// ErrorTypeRequest represents a request-related error
	ErrorTypeRequest
	// ErrorTypeParse represents a data parsing error
	ErrorTypeParse
	// ErrorTypeValidation represents a data validation error
	ErrorTypeValidation
	// ErrorTypeTimeout represents a timeout error
	ErrorTypeTimeout
	// ErrorTypeRateLimit represents a rate limiting error
	ErrorTypeRateLimit
	// ErrorTypeInternal represents an internal system error
	ErrorTypeInternal
)

// String returns the string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeConfig:
		return "config"
	case ErrorTypeConnection:
		return "connection"
	case ErrorTypeAuth:
		return "auth"
	case ErrorTypeRequest:
		return "request"
	case ErrorTypeParse:
		return "parse"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeTimeout:
		return "timeout"
	case ErrorTypeRateLimit:
		return "rate_limit"
	case ErrorTypeInternal:
		return "internal"
	default:
		return "unknown"
	}
}

// ErrorSeverity represents the severity level of an error
type ErrorSeverity int

const (
	// ErrorSeverityUnknown represents an unknown error severity
	ErrorSeverityUnknown ErrorSeverity = iota
	// ErrorSeverityLow represents a low severity error (e.g., temporary network issue)
	ErrorSeverityLow
	// ErrorSeverityMedium represents a medium severity error (e.g., authentication failure)
	ErrorSeverityMedium
	// ErrorSeverityHigh represents a high severity error (e.g., configuration error)
	ErrorSeverityHigh
	// ErrorSeverityCritical represents a critical error (e.g., system failure)
	ErrorSeverityCritical
)

// String returns the string representation of the error severity
func (es ErrorSeverity) String() string {
	switch es {
	case ErrorSeverityLow:
		return "low"
	case ErrorSeverityMedium:
		return "medium"
	case ErrorSeverityHigh:
		return "high"
	case ErrorSeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Error represents a structured error in the WinPower module
type Error struct {
	Type      ErrorType              `json:"type"`
	Severity  ErrorSeverity          `json:"severity"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Operation string                 `json:"operation"`
	DeviceID  string                 `json:"device_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Cause     error                  `json:"cause,omitempty"`
	Retryable bool                   `json:"retryable"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.DeviceID != "" {
		return fmt.Sprintf("[%s:%s] %s operation '%s' failed for device '%s': %s",
			e.Type, e.Severity, e.Code, e.Operation, e.DeviceID, e.Message)
	}
	return fmt.Sprintf("[%s:%s] %s operation '%s' failed: %s",
		e.Type, e.Severity, e.Code, e.Operation, e.Message)
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *Error) WithContext(key string, value interface{}) *Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause adds a cause to the error
func (e *Error) WithCause(cause error) *Error {
	e.Cause = cause
	return e
}

// WithDeviceID sets the device ID for the error
func (e *Error) WithDeviceID(deviceID string) *Error {
	e.DeviceID = deviceID
	return e
}

// NewError creates a new WinPower error
func NewError(errorType ErrorType, severity ErrorSeverity, code, operation, message string) *Error {
	return &Error{
		Type:      errorType,
		Severity:  severity,
		Code:      code,
		Operation: operation,
		Message:   message,
		Timestamp: time.Now(),
		Retryable: determineRetryability(errorType, severity),
	}
}

// determineRetryability determines if an error is retryable based on type and severity
func determineRetryability(errorType ErrorType, severity ErrorSeverity) bool {
	switch errorType {
	case ErrorTypeConnection, ErrorTypeTimeout, ErrorTypeRateLimit, ErrorTypeConfig:
		return true
	case ErrorTypeAuth, ErrorTypeValidation, ErrorTypeParse:
		return false
	case ErrorTypeRequest, ErrorTypeInternal:
		return severity != ErrorSeverityCritical
	default:
		return false
	}
}

// Common error constructors
func NewConfigError(operation, message string) *Error {
	return NewError(ErrorTypeConfig, ErrorSeverityHigh, "CONFIG_ERROR", operation, message)
}

func NewConnectionError(operation, message string) *Error {
	return NewError(ErrorTypeConnection, ErrorSeverityMedium, "CONNECTION_ERROR", operation, message)
}

func NewAuthError(operation, message string) *Error {
	return NewError(ErrorTypeAuth, ErrorSeverityMedium, "AUTH_ERROR", operation, message)
}

func NewRequestError(operation, message string) *Error {
	return NewError(ErrorTypeRequest, ErrorSeverityLow, "REQUEST_ERROR", operation, message)
}

func NewParseError(operation, message string) *Error {
	return NewError(ErrorTypeParse, ErrorSeverityMedium, "PARSE_ERROR", operation, message)
}

func NewValidationError(operation, message string) *Error {
	return NewError(ErrorTypeValidation, ErrorSeverityMedium, "VALIDATION_ERROR", operation, message)
}

func NewTimeoutError(operation, message string) *Error {
	return NewError(ErrorTypeTimeout, ErrorSeverityMedium, "TIMEOUT_ERROR", operation, message)
}

func NewRateLimitError(operation, message string) *Error {
	return NewError(ErrorTypeRateLimit, ErrorSeverityLow, "RATE_LIMIT_ERROR", operation, message)
}

func NewInternalError(operation, message string) *Error {
	return NewError(ErrorTypeInternal, ErrorSeverityCritical, "INTERNAL_ERROR", operation, message)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if wErr, ok := err.(*Error); ok {
		return wErr.Retryable
	}
	return false
}

// GetErrorType returns the type of a WinPower error
func GetErrorType(err error) ErrorType {
	if wErr, ok := err.(*Error); ok {
		return wErr.Type
	}
	return ErrorTypeUnknown
}

// GetErrorSeverity returns the severity of a WinPower error
func GetErrorSeverity(err error) ErrorSeverity {
	if wErr, ok := err.(*Error); ok {
		return wErr.Severity
	}
	return ErrorSeverityUnknown
}

// GetDeviceID returns the device ID from a WinPower error if available
func GetDeviceID(err error) string {
	if wErr, ok := err.(*Error); ok {
		return wErr.DeviceID
	}
	return ""
}
