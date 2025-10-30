package collector

import (
	"errors"
	"testing"
	"time"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{
			name:     "WinPowerCollection error type",
			errType:  ErrorTypeWinPowerCollection,
			expected: "WinPowerCollection",
		},
		{
			name:     "EnergyCalculation error type",
			errType:  ErrorTypeEnergyCalculation,
			expected: "EnergyCalculation",
		},
		{
			name:     "DataConversion error type",
			errType:  ErrorTypeDataConversion,
			expected: "DataConversion",
		},
		{
			name:     "Unknown error type",
			errType:  ErrorType(999),
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errType.String()
			if result != tt.expected {
				t.Errorf("ErrorType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCollectionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CollectionError
		expected string
	}{
		{
			name: "Error with device ID",
			err: &CollectionError{
				Type:     ErrorTypeEnergyCalculation,
				DeviceID: "device123",
				Message:  "calculation failed",
			},
			expected: "[EnergyCalculation] Device device123: calculation failed",
		},
		{
			name: "Error without device ID",
			err: &CollectionError{
				Type:    ErrorTypeWinPowerCollection,
				Message: "connection timeout",
			},
			expected: "[WinPowerCollection] connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("CollectionError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCollectionError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &CollectionError{
		Type:    ErrorTypeDataConversion,
		Message: "conversion failed",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("CollectionError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestNewCollectionError(t *testing.T) {
	cause := errors.New("test cause")
	deviceID := "device123"
	message := "test message"
	errType := ErrorTypeEnergyCalculation

	before := time.Now()
	err := NewCollectionError(errType, deviceID, message, cause)
	after := time.Now()

	if err.Type != errType {
		t.Errorf("Type = %v, want %v", err.Type, errType)
	}
	if err.DeviceID != deviceID {
		t.Errorf("DeviceID = %v, want %v", err.DeviceID, deviceID)
	}
	if err.Message != message {
		t.Errorf("Message = %v, want %v", err.Message, message)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
	if err.Timestamp.Before(before) || err.Timestamp.After(after) {
		t.Errorf("Timestamp not in expected range")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrWinPowerCollection",
			err:  ErrWinPowerCollection,
			want: "winpower data collection failed",
		},
		{
			name: "ErrEnergyCalculation",
			err:  ErrEnergyCalculation,
			want: "energy calculation failed",
		},
		{
			name: "ErrDataConversion",
			err:  ErrDataConversion,
			want: "data conversion failed",
		},
		{
			name: "ErrInvalidContext",
			err:  ErrInvalidContext,
			want: "invalid context",
		},
		{
			name: "ErrNilDependency",
			err:  ErrNilDependency,
			want: "nil dependency provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error message = %v, want %v", tt.err.Error(), tt.want)
			}
		})
	}
}
