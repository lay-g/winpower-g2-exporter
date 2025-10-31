package scheduler

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.CollectionInterval != 5*time.Second {
		t.Errorf("expected CollectionInterval to be 5s, got %v", config.CollectionInterval)
	}

	if config.GracefulShutdownTimeout != 5*time.Second {
		t.Errorf("expected GracefulShutdownTimeout to be 5s, got %v", config.GracefulShutdownTimeout)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "valid custom config",
			config: &Config{
				CollectionInterval:      10 * time.Second,
				GracefulShutdownTimeout: 3 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "zero collection interval",
			config: &Config{
				CollectionInterval:      0,
				GracefulShutdownTimeout: 5 * time.Second,
			},
			wantErr: true,
			errMsg:  "collection_interval must be positive",
		},
		{
			name: "negative collection interval",
			config: &Config{
				CollectionInterval:      -1 * time.Second,
				GracefulShutdownTimeout: 5 * time.Second,
			},
			wantErr: true,
			errMsg:  "collection_interval must be positive",
		},
		{
			name: "zero graceful shutdown timeout",
			config: &Config{
				CollectionInterval:      5 * time.Second,
				GracefulShutdownTimeout: 0,
			},
			wantErr: true,
			errMsg:  "graceful_shutdown_timeout must be positive",
		},
		{
			name: "negative graceful shutdown timeout",
			config: &Config{
				CollectionInterval:      5 * time.Second,
				GracefulShutdownTimeout: -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "graceful_shutdown_timeout must be positive",
		},
		{
			name: "collection interval too small",
			config: &Config{
				CollectionInterval:      500 * time.Millisecond,
				GracefulShutdownTimeout: 5 * time.Second,
			},
			wantErr: true,
			errMsg:  "collection_interval must be at least 1s",
		},
		{
			name: "collection interval too large",
			config: &Config{
				CollectionInterval:      2 * time.Hour,
				GracefulShutdownTimeout: 5 * time.Second,
			},
			wantErr: true,
			errMsg:  "collection_interval must not exceed 1h0m0s",
		},
		{
			name: "minimum valid collection interval",
			config: &Config{
				CollectionInterval:      1 * time.Second,
				GracefulShutdownTimeout: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "maximum valid collection interval",
			config: &Config{
				CollectionInterval:      1 * time.Hour,
				GracefulShutdownTimeout: 1 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}
