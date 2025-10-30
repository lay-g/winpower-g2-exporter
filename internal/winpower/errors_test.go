package winpower

import (
	"errors"
	"testing"
)

func TestAuthenticationError(t *testing.T) {
	tests := []struct {
		name     string
		err      *AuthenticationError
		expected string
	}{
		{
			name: "with wrapped error",
			err: &AuthenticationError{
				Message: "invalid credentials",
				Err:     errors.New("unauthorized"),
			},
			expected: "authentication error: invalid credentials: unauthorized",
		},
		{
			name: "without wrapped error",
			err: &AuthenticationError{
				Message: "token expired",
			},
			expected: "authentication error: token expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAuthenticationError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("base error")
	err := &AuthenticationError{
		Message: "test",
		Err:     wrappedErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}

	err2 := &AuthenticationError{
		Message: "test",
	}

	if unwrapped := err2.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}
}

func TestNetworkError(t *testing.T) {
	tests := []struct {
		name     string
		err      *NetworkError
		expected string
	}{
		{
			name: "with wrapped error",
			err: &NetworkError{
				Message: "connection refused",
				Err:     errors.New("dial tcp failed"),
			},
			expected: "network error: connection refused: dial tcp failed",
		},
		{
			name: "without wrapped error",
			err: &NetworkError{
				Message: "timeout",
			},
			expected: "network error: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestNetworkError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("base error")
	err := &NetworkError{
		Message: "test",
		Err:     wrappedErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name: "with wrapped error",
			err: &ParseError{
				Field:   "loadTotalWatt",
				Message: "invalid format",
				Err:     errors.New("strconv error"),
			},
			expected: "parse error on field \"loadTotalWatt\": invalid format: strconv error",
		},
		{
			name: "without wrapped error",
			err: &ParseError{
				Field:   "deviceId",
				Message: "missing field",
			},
			expected: "parse error on field \"deviceId\": missing field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("base error")
	err := &ParseError{
		Field:   "test",
		Message: "test",
		Err:     wrappedErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name: "with wrapped error",
			err: &ConfigError{
				Field:   "base_url",
				Message: "invalid format",
				Err:     errors.New("parse error"),
			},
			expected: "config error on field \"base_url\": invalid format: parse error",
		},
		{
			name: "without wrapped error",
			err: &ConfigError{
				Field:   "timeout",
				Message: "must be positive",
			},
			expected: "config error on field \"timeout\": must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("base error")
	err := &ConfigError{
		Field:   "test",
		Message: "test",
		Err:     wrappedErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != wrappedErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestIsAuthenticationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "AuthenticationError struct",
			err:  &AuthenticationError{Message: "test"},
			want: true,
		},
		{
			name: "ErrAuthenticationFailed",
			err:  ErrAuthenticationFailed,
			want: true,
		},
		{
			name: "ErrTokenExpired",
			err:  ErrTokenExpired,
			want: true,
		},
		{
			name: "wrapped AuthenticationError",
			err:  errors.Join(ErrAuthenticationFailed, errors.New("other")),
			want: true,
		},
		{
			name: "NetworkError",
			err:  &NetworkError{Message: "test"},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthenticationError(tt.err); got != tt.want {
				t.Errorf("IsAuthenticationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNetworkError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "NetworkError struct",
			err:  &NetworkError{Message: "test"},
			want: true,
		},
		{
			name: "ErrNetworkError",
			err:  ErrNetworkError,
			want: true,
		},
		{
			name: "ErrTimeout",
			err:  ErrTimeout,
			want: true,
		},
		{
			name: "wrapped NetworkError",
			err:  errors.Join(ErrNetworkError, errors.New("other")),
			want: true,
		},
		{
			name: "AuthenticationError",
			err:  &AuthenticationError{Message: "test"},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNetworkError(tt.err); got != tt.want {
				t.Errorf("IsNetworkError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsParseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ParseError struct",
			err:  &ParseError{Field: "test", Message: "test"},
			want: true,
		},
		{
			name: "ErrParseError",
			err:  ErrParseError,
			want: true,
		},
		{
			name: "wrapped ParseError",
			err:  errors.Join(ErrParseError, errors.New("other")),
			want: true,
		},
		{
			name: "NetworkError",
			err:  &NetworkError{Message: "test"},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsParseError(tt.err); got != tt.want {
				t.Errorf("IsParseError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ConfigError struct",
			err:  &ConfigError{Field: "test", Message: "test"},
			want: true,
		},
		{
			name: "ErrInvalidConfig",
			err:  ErrInvalidConfig,
			want: true,
		},
		{
			name: "wrapped ConfigError",
			err:  errors.Join(ErrInvalidConfig, errors.New("other")),
			want: true,
		},
		{
			name: "NetworkError",
			err:  &NetworkError{Message: "test"},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("generic"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConfigError(tt.err); got != tt.want {
				t.Errorf("IsConfigError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are distinct
	sentinelErrors := []error{
		ErrConnectionNotReady,
		ErrAuthenticationFailed,
		ErrTokenExpired,
		ErrInvalidResponse,
		ErrNetworkError,
		ErrParseError,
		ErrInvalidConfig,
		ErrTimeout,
	}

	for i, err1 := range sentinelErrors {
		for j, err2 := range sentinelErrors {
			if i != j && err1 == err2 {
				t.Errorf("sentinel errors should be distinct: %v == %v", err1, err2)
			}
		}
	}

	// Test that each sentinel error has a non-empty message
	for _, err := range sentinelErrors {
		if err.Error() == "" {
			t.Errorf("sentinel error should have non-empty message: %v", err)
		}
	}
}
