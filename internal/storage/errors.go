package storage

import (
	"fmt"
)

// StorageError represents storage operation errors with context.
type StorageError struct {
	Operation string // The operation that failed (e.g., "write", "read")
	DeviceID  string // The device identifier
	Path      string // The file path involved
	Cause     error  // The underlying error
}

// Error implements the error interface.
func (se *StorageError) Error() string {
	if se.DeviceID != "" {
		return fmt.Sprintf("storage %s operation failed for device '%s' on path '%s': %v",
			se.Operation, se.DeviceID, se.Path, se.Cause)
	}
	return fmt.Sprintf("storage %s operation failed on path '%s': %v",
		se.Operation, se.Path, se.Cause)
}

// Unwrap returns the underlying cause for error handling.
func (se *StorageError) Unwrap() error {
	return se.Cause
}

// Predefined error types
var (
	// ErrFileNotFound indicates that the device file does not exist.
	ErrFileNotFound = fmt.Errorf("file not found")

	// ErrInvalidFormat indicates that the file format is invalid or corrupted.
	ErrInvalidFormat = fmt.Errorf("invalid file format")

	// ErrPermissionDenied indicates insufficient permissions for file operations.
	ErrPermissionDenied = fmt.Errorf("permission denied")

	// ErrInvalidData indicates invalid or nil data.
	ErrInvalidData = fmt.Errorf("invalid data")

	// ErrInvalidTimestamp indicates an invalid timestamp value.
	ErrInvalidTimestamp = fmt.Errorf("invalid timestamp")

	// ErrInvalidEnergyValue indicates an invalid energy value.
	ErrInvalidEnergyValue = fmt.Errorf("invalid energy value")

	// ErrDiskFull indicates insufficient disk space.
	ErrDiskFull = fmt.Errorf("disk full")
)

// NewStorageError creates a new storage error with context.
func NewStorageError(operation, deviceID, path string, cause error) *StorageError {
	return &StorageError{
		Operation: operation,
		DeviceID:  deviceID,
		Path:      path,
		Cause:     cause,
	}
}
