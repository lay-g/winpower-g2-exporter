package scheduler

import (
	"errors"
	"testing"
)

func TestErrorValues(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrAlreadyRunning",
			err:  ErrAlreadyRunning,
			want: "scheduler is already running",
		},
		{
			name: "ErrNotRunning",
			err:  ErrNotRunning,
			want: "scheduler is not running",
		},
		{
			name: "ErrShutdownTimeout",
			err:  ErrShutdownTimeout,
			want: "scheduler shutdown timeout exceeded",
		},
		{
			name: "ErrNilCollector",
			err:  ErrNilCollector,
			want: "collector cannot be nil",
		},
		{
			name: "ErrNilLogger",
			err:  ErrNilLogger,
			want: "logger cannot be nil",
		},
		{
			name: "ErrNilConfig",
			err:  ErrNilConfig,
			want: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
				return
			}
			if tt.err.Error() != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

func TestErrorIdentity(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
	}{
		{
			name:   "ErrAlreadyRunning identity",
			err:    ErrAlreadyRunning,
			target: ErrAlreadyRunning,
		},
		{
			name:   "ErrNotRunning identity",
			err:    ErrNotRunning,
			target: ErrNotRunning,
		},
		{
			name:   "ErrShutdownTimeout identity",
			err:    ErrShutdownTimeout,
			target: ErrShutdownTimeout,
		},
		{
			name:   "ErrNilCollector identity",
			err:    ErrNilCollector,
			target: ErrNilCollector,
		},
		{
			name:   "ErrNilLogger identity",
			err:    ErrNilLogger,
			target: ErrNilLogger,
		},
		{
			name:   "ErrNilConfig identity",
			err:    ErrNilConfig,
			target: ErrNilConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.target) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.target)
			}
		})
	}
}
