package storage

import (
	"path/filepath"
	"testing"
)

func TestValidateDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty device ID",
			deviceID: "",
			wantErr:  true,
			errMsg:   "device ID cannot be empty",
		},
		{
			name:     "device ID with forward slash",
			deviceID: "device/1",
			wantErr:  true,
			errMsg:   "device ID cannot contain path separators",
		},
		{
			name:     "device ID with backslash",
			deviceID: "device\\1",
			wantErr:  true,
			errMsg:   "device ID cannot contain path separators",
		},
		{
			name:     "device ID is single dot",
			deviceID: ".",
			wantErr:  true,
			errMsg:   "device ID cannot be a relative path component",
		},
		{
			name:     "device ID is double dot",
			deviceID: "..",
			wantErr:  true,
			errMsg:   "device ID cannot be a relative path component",
		},
		{
			name:     "device ID starts with dot",
			deviceID: ".hidden",
			wantErr:  true,
			errMsg:   "device ID cannot start with a dot",
		},
		{
			name:     "valid simple device ID",
			deviceID: "device1",
			wantErr:  false,
		},
		{
			name:     "valid device ID with hyphen",
			deviceID: "device-1",
			wantErr:  false,
		},
		{
			name:     "valid device ID with underscore",
			deviceID: "device_1",
			wantErr:  false,
		},
		{
			name:     "valid device ID with numbers",
			deviceID: "device123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDeviceID(tt.deviceID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateDeviceID() error = nil, want error containing %q", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("validateDeviceID() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateDeviceID() error = %v, want nil", err)
				}
			}
		})
	}
}

func TestBuildFilePath(t *testing.T) {
	tests := []struct {
		name     string
		dataDir  string
		deviceID string
		wantErr  bool
		errMsg   string
		validate func(t *testing.T, path string)
	}{
		{
			name:     "empty device ID",
			dataDir:  "./data",
			deviceID: "",
			wantErr:  true,
			errMsg:   "device ID cannot be empty",
		},
		{
			name:     "device ID with path separator",
			dataDir:  "./data",
			deviceID: "device/1",
			wantErr:  true,
			errMsg:   "device ID cannot contain path separators",
		},
		{
			name:     "valid device ID",
			dataDir:  "./data",
			deviceID: "device1",
			wantErr:  false,
			validate: func(t *testing.T, path string) {
				expected := filepath.Join("data", "device1.txt")
				if path != expected {
					t.Errorf("buildFilePath() = %v, want %v", path, expected)
				}
			},
		},
		{
			name:     "valid device ID with absolute path",
			dataDir:  "/var/lib/exporter",
			deviceID: "device123",
			wantErr:  false,
			validate: func(t *testing.T, path string) {
				expected := filepath.Join("/var/lib/exporter", "device123.txt")
				if path != expected {
					t.Errorf("buildFilePath() = %v, want %v", path, expected)
				}
			},
		},
		{
			name:     "device ID attempts path traversal",
			dataDir:  "./data",
			deviceID: "..",
			wantErr:  true,
			errMsg:   "device ID cannot be a relative path component",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := buildFilePath(tt.dataDir, tt.deviceID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("buildFilePath() error = nil, want error containing %q", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("buildFilePath() error = %v, want error containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("buildFilePath() error = %v, want nil", err)
					return
				}
				if tt.validate != nil {
					tt.validate(t, path)
				}
			}
		})
	}
}
