package server

import "testing"

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrInvalidConfig",
			err:  ErrInvalidConfig,
			want: "invalid server configuration",
		},
		{
			name: "ErrServerNotStarted",
			err:  ErrServerNotStarted,
			want: "server not started",
		},
		{
			name: "ErrServerAlreadyRunning",
			err:  ErrServerAlreadyRunning,
			want: "server already running",
		},
		{
			name: "ErrMetricsServiceNil",
			err:  ErrMetricsServiceNil,
			want: "metrics service cannot be nil",
		},
		{
			name: "ErrHealthServiceNil",
			err:  ErrHealthServiceNil,
			want: "health service cannot be nil",
		},
		{
			name: "ErrLoggerNil",
			err:  ErrLoggerNil,
			want: "logger cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
