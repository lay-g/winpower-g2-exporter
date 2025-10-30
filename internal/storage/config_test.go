package storage

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DataDir != "./data" {
		t.Errorf("DataDir = %v, want ./data", cfg.DataDir)
	}

	if cfg.FilePermissions != 0644 {
		t.Errorf("FilePermissions = %v, want 0644", cfg.FilePermissions)
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
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "empty data directory",
			config: &Config{
				DataDir:         "",
				FilePermissions: 0644,
			},
			wantErr: true,
			errMsg:  "data directory cannot be empty",
		},
		{
			name: "invalid file permissions",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 01000,
			},
			wantErr: true,
			errMsg:  "file permissions must be a valid Unix permission",
		},
		{
			name: "valid config",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0644,
			},
			wantErr: false,
		},
		{
			name: "valid config with different permissions",
			config: &Config{
				DataDir:         "/var/lib/exporter",
				FilePermissions: 0600,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errMsg)
					return
				}
				// Check if error message contains expected substring
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
