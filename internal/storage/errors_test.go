package storage

import (
	"errors"
	"testing"
)

func TestStorageError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *StorageError
		expected string
	}{
		{
			name: "error with device ID",
			err: &StorageError{
				Operation: "write",
				DeviceID:  "device123",
				Path:      "/tmp/device123.txt",
				Cause:     errors.New("permission denied"),
			},
			expected: "storage write operation failed for device 'device123' on path '/tmp/device123.txt': permission denied",
		},
		{
			name: "error without device ID",
			err: &StorageError{
				Operation: "read",
				DeviceID:  "",
				Path:      "/tmp/unknown.txt",
				Cause:     errors.New("file not found"),
			},
			expected: "storage read operation failed on path '/tmp/unknown.txt': file not found",
		},
		{
			name: "error with nil cause",
			err: &StorageError{
				Operation: "validate",
				DeviceID:  "device456",
				Path:      "/tmp/device456.txt",
				Cause:     nil,
			},
			expected: "storage validate operation failed for device 'device456' on path '/tmp/device456.txt': <nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.err.Error()
			if actual != tt.expected {
				t.Errorf("StorageError.Error() = %s, expected %s", actual, tt.expected)
			}
		})
	}
}

func TestStorageError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	err := &StorageError{
		Operation: "write",
		DeviceID:  "device123",
		Path:      "/tmp/device123.txt",
		Cause:     originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("StorageError.Unwrap() = %v, expected %v", unwrapped, originalErr)
	}

	// Test with nil cause
	errNilCause := &StorageError{
		Operation: "write",
		DeviceID:  "device123",
		Path:      "/tmp/device123.txt",
		Cause:     nil,
	}

	unwrappedNil := errNilCause.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("StorageError.Unwrap() with nil cause = %v, expected nil", unwrappedNil)
	}
}

func TestNewStorageError(t *testing.T) {
	operation := "write"
	deviceID := "device123"
	path := "/tmp/device123.txt"
	cause := errors.New("test error")

	err := NewStorageError(operation, deviceID, path, cause)

	if err.Operation != operation {
		t.Errorf("NewStorageError() Operation = %s, expected %s", err.Operation, operation)
	}

	if err.DeviceID != deviceID {
		t.Errorf("NewStorageError() DeviceID = %s, expected %s", err.DeviceID, deviceID)
	}

	if err.Path != path {
		t.Errorf("NewStorageError() Path = %s, expected %s", err.Path, path)
	}

	if err.Cause != cause {
		t.Errorf("NewStorageError() Cause = %v, expected %v", err.Cause, cause)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check func(error) bool
	}{
		{
			name: "ErrFileNotFound",
			err:  ErrFileNotFound,
			check: func(err error) bool {
				return err.Error() == "file not found"
			},
		},
		{
			name: "ErrInvalidFormat",
			err:  ErrInvalidFormat,
			check: func(err error) bool {
				return err.Error() == "invalid file format"
			},
		},
		{
			name: "ErrPermissionDenied",
			err:  ErrPermissionDenied,
			check: func(err error) bool {
				return err.Error() == "permission denied"
			},
		},
		{
			name: "ErrInvalidData",
			err:  ErrInvalidData,
			check: func(err error) bool {
				return err.Error() == "invalid data"
			},
		},
		{
			name: "ErrInvalidTimestamp",
			err:  ErrInvalidTimestamp,
			check: func(err error) bool {
				return err.Error() == "invalid timestamp"
			},
		},
		{
			name: "ErrInvalidEnergyValue",
			err:  ErrInvalidEnergyValue,
			check: func(err error) bool {
				return err.Error() == "invalid energy value"
			},
		},
		{
			name: "ErrDiskFull",
			err:  ErrDiskFull,
			check: func(err error) bool {
				return err.Error() == "disk full"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.err) {
				t.Errorf("Predefined error %s check failed", tt.name)
			}
		})
	}
}

func TestStorageError_ErrorHandling(t *testing.T) {
	// Test error wrapping and unwrapping
	originalErr := errors.New("disk full")
	storageErr := NewStorageError("write", "device123", "/tmp/device123.txt", originalErr)

	// Test errors.Is functionality
	if !errors.Is(storageErr, originalErr) {
		t.Error("errors.Is should return true for wrapped error")
	}

	// Test error message contains expected components
	errMsg := storageErr.Error()
	expectedComponents := []string{
		"write",
		"device123",
		"/tmp/device123.txt",
		"disk full",
	}

	for _, component := range expectedComponents {
		if !contains(errMsg, component) {
			t.Errorf("Error message should contain '%s': %s", component, errMsg)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
