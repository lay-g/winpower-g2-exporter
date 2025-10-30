package storage

import (
	"errors"
	"fmt"
)

// Common error variables
var (
	// ErrFileNotFound indicates that a device file was not found
	ErrFileNotFound = errors.New("device file not found")

	// ErrInvalidFormat indicates that a device file has an invalid format
	ErrInvalidFormat = errors.New("invalid file format")

	// ErrInvalidDeviceID indicates that a device ID is invalid
	ErrInvalidDeviceID = errors.New("invalid device ID")

	// ErrInvalidData indicates that the provided data is invalid
	ErrInvalidData = errors.New("invalid data")

	// ErrPermissionDenied indicates that file operation was denied due to permissions
	ErrPermissionDenied = errors.New("permission denied")
)

// StorageError represents an error that occurred during storage operations.
type StorageError struct {
	// Operation is the operation that failed (e.g., "read", "write")
	Operation string

	// Path is the file path where the error occurred
	Path string

	// Err is the underlying error
	Err error
}

// Error implements the error interface.
func (e *StorageError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("storage %s failed for %s: %v", e.Operation, e.Path, e.Err)
	}
	return fmt.Sprintf("storage %s failed: %v", e.Operation, e.Err)
}

// Unwrap returns the underlying error.
func (e *StorageError) Unwrap() error {
	return e.Err
}

// NewStorageError creates a new StorageError.
func NewStorageError(operation, path string, err error) *StorageError {
	return &StorageError{
		Operation: operation,
		Path:      path,
		Err:       err,
	}
}
