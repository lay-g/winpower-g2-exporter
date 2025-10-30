package storage

import (
	"errors"
	"testing"
)

func TestStorageError_Error(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		path      string
		err       error
		want      string
	}{
		{
			name:      "error with path",
			operation: "read",
			path:      "/data/device1.txt",
			err:       errors.New("file not found"),
			want:      "storage read failed for /data/device1.txt: file not found",
		},
		{
			name:      "error without path",
			operation: "write",
			path:      "",
			err:       errors.New("permission denied"),
			want:      "storage write failed: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &StorageError{
				Operation: tt.operation,
				Path:      tt.path,
				Err:       tt.err,
			}

			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorageError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	storageErr := &StorageError{
		Operation: "read",
		Path:      "/data/test.txt",
		Err:       originalErr,
	}

	unwrapped := storageErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestNewStorageError(t *testing.T) {
	operation := "write"
	path := "/data/device1.txt"
	err := errors.New("disk full")

	storageErr := NewStorageError(operation, path, err)

	if storageErr.Operation != operation {
		t.Errorf("Operation = %v, want %v", storageErr.Operation, operation)
	}
	if storageErr.Path != path {
		t.Errorf("Path = %v, want %v", storageErr.Path, path)
	}
	if storageErr.Err != err {
		t.Errorf("Err = %v, want %v", storageErr.Err, err)
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrFileNotFound",
			err:  ErrFileNotFound,
			want: "device file not found",
		},
		{
			name: "ErrInvalidFormat",
			err:  ErrInvalidFormat,
			want: "invalid file format",
		},
		{
			name: "ErrInvalidDeviceID",
			err:  ErrInvalidDeviceID,
			want: "invalid device ID",
		},
		{
			name: "ErrInvalidData",
			err:  ErrInvalidData,
			want: "invalid data",
		},
		{
			name: "ErrPermissionDenied",
			err:  ErrPermissionDenied,
			want: "permission denied",
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
