package winpower

import (
	"errors"
	"testing"
	"time"
)

// TestError_Error tests the Error method
func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "error_without_device_id",
			test: func(t *testing.T) {
				err := NewError(ErrorTypeConnection, ErrorSeverityMedium, "CONN_001", "connect", "Connection failed")
				expected := "[connection:medium] CONN_001 operation 'connect' failed: Connection failed"
				if got := err.Error(); got != expected {
					t.Errorf("Error.Error() = %v, want %v", got, expected)
				}
			},
		},
		{
			name: "error_with_device_id",
			test: func(t *testing.T) {
				err := NewError(ErrorTypeConnection, ErrorSeverityMedium, "CONN_001", "connect", "Connection failed")
				err.DeviceID = "device-123"
				expected := "[connection:medium] CONN_001 operation 'connect' failed for device 'device-123': Connection failed"
				if got := err.Error(); got != expected {
					t.Errorf("Error.Error() = %v, want %v", got, expected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestErrorType_String tests the ErrorType String method
func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{"config", ErrorTypeConfig, "config"},
		{"connection", ErrorTypeConnection, "connection"},
		{"auth", ErrorTypeAuth, "auth"},
		{"request", ErrorTypeRequest, "request"},
		{"parse", ErrorTypeParse, "parse"},
		{"validation", ErrorTypeValidation, "validation"},
		{"timeout", ErrorTypeTimeout, "timeout"},
		{"rate_limit", ErrorTypeRateLimit, "rate_limit"},
		{"internal", ErrorTypeInternal, "internal"},
		{"unknown", ErrorTypeUnknown, "unknown"},
		{"invalid", ErrorType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestErrorSeverity_String tests the ErrorSeverity String method
func TestErrorSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		expected string
	}{
		{"low", ErrorSeverityLow, "low"},
		{"medium", ErrorSeverityMedium, "medium"},
		{"high", ErrorSeverityHigh, "high"},
		{"critical", ErrorSeverityCritical, "critical"},
		{"unknown", ErrorSeverityUnknown, "unknown"},
		{"invalid", ErrorSeverity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.severity.String(); got != tt.expected {
				t.Errorf("ErrorSeverity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestNewError tests the NewError constructor
func TestNewError(t *testing.T) {
	t.Run("creates_error_with_correct_fields", func(t *testing.T) {
		before := time.Now()
		err := NewError(ErrorTypeAuth, ErrorSeverityHigh, "AUTH_001", "login", "Invalid credentials")
		after := time.Now()

		if err.Type != ErrorTypeAuth {
			t.Errorf("NewError() Type = %v, want %v", err.Type, ErrorTypeAuth)
		}

		if err.Severity != ErrorSeverityHigh {
			t.Errorf("NewError() Severity = %v, want %v", err.Severity, ErrorSeverityHigh)
		}

		if err.Code != "AUTH_001" {
			t.Errorf("NewError() Code = %v, want %v", err.Code, "AUTH_001")
		}

		if err.Operation != "login" {
			t.Errorf("NewError() Operation = %v, want %v", err.Operation, "login")
		}

		if err.Message != "Invalid credentials" {
			t.Errorf("NewError() Message = %v, want %v", err.Message, "Invalid credentials")
		}

		if err.Timestamp.Before(before) || err.Timestamp.After(after) {
			t.Errorf("NewError() Timestamp = %v, should be between %v and %v", err.Timestamp, before, after)
		}

		if err.DeviceID != "" {
			t.Errorf("NewError() DeviceID = %v, want empty string", err.DeviceID)
		}
	})
}

// TestError_Wrappers tests the WithX methods
func TestError_Wrappers(t *testing.T) {
	err := NewError(ErrorTypeRequest, ErrorSeverityLow, "REQ_001", "fetch", "Request failed")

	t.Run("WithContext", func(t *testing.T) {
		result := err.WithContext("attempt", 3)
		if result.Context["attempt"] != 3 {
			t.Errorf("WithContext() = %v, want %v", result.Context["attempt"], 3)
		}
	})

	t.Run("WithCause", func(t *testing.T) {
		cause := errors.New("network error")
		result := err.WithCause(cause)
		if result.Cause != cause {
			t.Errorf("WithCause() = %v, want %v", result.Cause, cause)
		}
	})

	t.Run("WithDeviceID", func(t *testing.T) {
		deviceID := "device-456"
		result := err.WithDeviceID(deviceID)
		if result.DeviceID != deviceID {
			t.Errorf("WithDeviceID() = %v, want %v", result.DeviceID, deviceID)
		}
	})
}

// TestCommonErrorConstructors tests the common error constructors
func TestCommonErrorConstructors(t *testing.T) {
	tests := []struct {
		name             string
		constructor      func(string, string) *Error
		expectedType     ErrorType
		expectedSeverity ErrorSeverity
		expectedCode     string
	}{
		{
			name:             "NewConfigError",
			constructor:      NewConfigError,
			expectedType:     ErrorTypeConfig,
			expectedSeverity: ErrorSeverityHigh,
			expectedCode:     "CONFIG_ERROR",
		},
		{
			name:             "NewConnectionError",
			constructor:      NewConnectionError,
			expectedType:     ErrorTypeConnection,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "CONNECTION_ERROR",
		},
		{
			name:             "NewAuthError",
			constructor:      NewAuthError,
			expectedType:     ErrorTypeAuth,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "AUTH_ERROR",
		},
		{
			name:             "NewRequestError",
			constructor:      NewRequestError,
			expectedType:     ErrorTypeRequest,
			expectedSeverity: ErrorSeverityLow,
			expectedCode:     "REQUEST_ERROR",
		},
		{
			name:             "NewParseError",
			constructor:      NewParseError,
			expectedType:     ErrorTypeParse,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "PARSE_ERROR",
		},
		{
			name:             "NewValidationError",
			constructor:      NewValidationError,
			expectedType:     ErrorTypeValidation,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "VALIDATION_ERROR",
		},
		{
			name:             "NewTimeoutError",
			constructor:      NewTimeoutError,
			expectedType:     ErrorTypeTimeout,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "TIMEOUT_ERROR",
		},
		{
			name:             "NewRateLimitError",
			constructor:      NewRateLimitError,
			expectedType:     ErrorTypeRateLimit,
			expectedSeverity: ErrorSeverityLow,
			expectedCode:     "RATE_LIMIT_ERROR",
		},
		{
			name:             "NewInternalError",
			constructor:      NewInternalError,
			expectedType:     ErrorTypeInternal,
			expectedSeverity: ErrorSeverityCritical,
			expectedCode:     "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test_op", "test message")
			if err.Type != tt.expectedType {
				t.Errorf("%s() Type = %v, want %v", tt.name, err.Type, tt.expectedType)
			}
			if err.Severity != tt.expectedSeverity {
				t.Errorf("%s() Severity = %v, want %v", tt.name, err.Severity, tt.expectedSeverity)
			}
			if err.Code != tt.expectedCode {
				t.Errorf("%s() Code = %v, want %v", tt.name, err.Code, tt.expectedCode)
			}
			if err.Operation != "test_op" {
				t.Errorf("%s() Operation = %v, want %v", tt.name, err.Operation, "test_op")
			}
			if err.Message != "test message" {
				t.Errorf("%s() Message = %v, want %v", tt.name, err.Message, "test message")
			}
		})
	}
}

// TestError_HelperFunctions tests error helper functions
func TestError_HelperFunctions(t *testing.T) {
	// Test IsRetryable
	t.Run("IsRetryable", func(t *testing.T) {
		retryableErr := NewConnectionError("connect", "connection failed")
		nonRetryableErr := NewConfigError("validate", "invalid config")

		if !IsRetryable(retryableErr) {
			t.Errorf("IsRetryable() = false for connection error, want true")
		}

		if !IsRetryable(nonRetryableErr) {
			t.Errorf("IsRetryable() = false for config error, want true")
		}

		if IsRetryable(errors.New("generic error")) {
			t.Errorf("IsRetryable() = true for generic error, want false")
		}
	})

	// Test GetErrorType
	t.Run("GetErrorType", func(t *testing.T) {
		authErr := NewAuthError("login", "login failed")
		if GetErrorType(authErr) != ErrorTypeAuth {
			t.Errorf("GetErrorType() = %v, want %v", GetErrorType(authErr), ErrorTypeAuth)
		}

		if GetErrorType(errors.New("generic error")) != ErrorTypeUnknown {
			t.Errorf("GetErrorType() = %v, want %v", GetErrorType(errors.New("generic error")), ErrorTypeUnknown)
		}
	})

	// Test GetErrorSeverity
	t.Run("GetErrorSeverity", func(t *testing.T) {
		criticalErr := NewInternalError("startup", "system failure")
		if GetErrorSeverity(criticalErr) != ErrorSeverityCritical {
			t.Errorf("GetErrorSeverity() = %v, want %v", GetErrorSeverity(criticalErr), ErrorSeverityCritical)
		}

		if GetErrorSeverity(errors.New("generic error")) != ErrorSeverityUnknown {
			t.Errorf("GetErrorSeverity() = %v, want %v", GetErrorSeverity(errors.New("generic error")), ErrorSeverityUnknown)
		}
	})

	// Test GetDeviceID
	t.Run("GetDeviceID", func(t *testing.T) {
		deviceErr := NewRequestError("collect", "collection failed")
		_ = deviceErr.WithDeviceID("device-789")

		if GetDeviceID(deviceErr) != "device-789" {
			t.Errorf("GetDeviceID() = %v, want %v", GetDeviceID(deviceErr), "device-789")
		}

		if GetDeviceID(errors.New("generic error")) != "" {
			t.Errorf("GetDeviceID() = %v, want empty string", GetDeviceID(errors.New("generic error")))
		}
	})
}

// TestError_Unwrap tests the Unwrap method for error chaining
func TestError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewRequestError("fetch", "fetch failed").WithCause(cause)

	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}

	// Test with no cause
	errWithoutCause := NewRequestError("fetch", "fetch failed")
	if errWithoutCause.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", errWithoutCause.Unwrap())
	}
}
