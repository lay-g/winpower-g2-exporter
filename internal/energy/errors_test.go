package energy

import (
	"errors"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrInvalidDeviceID",
			err:  ErrInvalidDeviceID,
			want: "invalid device ID: device ID cannot be empty",
		},
		{
			name: "ErrInvalidPower",
			err:  ErrInvalidPower,
			want: "invalid power value",
		},
		{
			name: "ErrStorageRead",
			err:  ErrStorageRead,
			want: "failed to read data from storage",
		},
		{
			name: "ErrStorageWrite",
			err:  ErrStorageWrite,
			want: "failed to write data to storage",
		},
		{
			name: "ErrCalculation",
			err:  ErrCalculation,
			want: "energy calculation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error message = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorWrapping(t *testing.T) {
	baseErr := errors.New("base error")

	tests := []struct {
		name     string
		err      error
		target   error
		shouldBe bool
	}{
		{
			name:     "Wrapped ErrStorageRead",
			err:      ErrStorageRead,
			target:   ErrStorageRead,
			shouldBe: true,
		},
		{
			name:     "Different error",
			err:      ErrStorageRead,
			target:   ErrStorageWrite,
			shouldBe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := errors.Is(tt.err, tt.target)
			if is != tt.shouldBe {
				t.Errorf("errors.Is() = %v, want %v", is, tt.shouldBe)
			}
		})
	}

	// Test wrapping
	t.Run("Wrapped error", func(t *testing.T) {
		wrappedErr := errors.Join(ErrStorageRead, baseErr)
		if !errors.Is(wrappedErr, ErrStorageRead) {
			t.Error("Wrapped error should match ErrStorageRead")
		}
	})
}
