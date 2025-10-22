package storage

import (
	"os"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfigWithDefaults()

	if config == nil {
		t.Fatal("NewConfigWithDefaults() returned nil")
	}

	// Check default values
	if config.DataDir != "./data" {
		t.Errorf("Expected DataDir to be './data', got '%s'", config.DataDir)
	}

	if config.FilePermissions != 0644 {
		t.Errorf("Expected FilePermissions to be 0644, got %o", config.FilePermissions)
	}

	if config.DirPermissions != 0755 {
		t.Errorf("Expected DirPermissions to be 0755, got %o", config.DirPermissions)
	}

	if !config.SyncWrite {
		t.Errorf("Expected SyncWrite to be true, got false")
	}

	if !config.CreateDir {
		t.Errorf("Expected CreateDir to be true, got false")
	}
}

func TestNewConfigWithPath(t *testing.T) {
	customPath := "/tmp/test-data"
	config := NewConfigWithPath(customPath)

	if config.DataDir != customPath {
		t.Errorf("Expected DataDir to be '%s', got '%s'", customPath, config.DataDir)
	}

	// Other defaults should still be set
	if config.FilePermissions != 0644 {
		t.Errorf("Expected FilePermissions to be 0644, got %o", config.FilePermissions)
	}

	if !config.SyncWrite {
		t.Errorf("Expected SyncWrite to be true, got false")
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	config := &Config{}

	// Before SetDefaults
	if config.DataDir != "" {
		t.Errorf("Expected DataDir to be empty, got '%s'", config.DataDir)
	}

	config.SetDefaults()

	// After SetDefaults - check only fields that SetDefaults should set
	if config.DataDir == "" {
		t.Error("Expected DataDir to be set after SetDefaults()")
	}

	if config.FilePermissions == 0 {
		t.Error("Expected FilePermissions to be set after SetDefaults()")
	}

	if config.DirPermissions == 0 {
		t.Error("Expected DirPermissions to be set after SetDefaults()")
	}

	// Note: SyncWrite and CreateDir are not set by SetDefaults
	// They are set only in NewConfig() to preserve existing values
}

func TestConfig_SetDefaults_PreservesExistingValues(t *testing.T) {
	config := &Config{
		DataDir:         "/custom/path",
		FilePermissions: 0600,
		DirPermissions:  0700,
		SyncWrite:       false,
		CreateDir:       false,
	}

	config.SetDefaults()

	// Existing values should be preserved
	if config.DataDir != "/custom/path" {
		t.Errorf("Expected DataDir to remain '/custom/path', got '%s'", config.DataDir)
	}

	if config.FilePermissions != 0600 {
		t.Errorf("Expected FilePermissions to remain 0600, got %o", config.FilePermissions)
	}

	if config.DirPermissions != 0700 {
		t.Errorf("Expected DirPermissions to remain 0700, got %o", config.DirPermissions)
	}

	if config.SyncWrite {
		t.Error("Expected SyncWrite to remain false")
	}

	if config.CreateDir {
		t.Error("Expected CreateDir to remain false")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr bool
		errMsg      string
	}{
		{
			name: "valid config",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0644,
				DirPermissions:  0755,
				SyncWrite:       true,
				CreateDir:       true,
			},
			expectedErr: false,
		},
		{
			name: "empty data dir",
			config: &Config{
				DataDir:         "",
				FilePermissions: 0644,
				DirPermissions:  0755,
				SyncWrite:       true,
				CreateDir:       true,
			},
			expectedErr: true,
			errMsg:      "storage: data_dir is required",
		},
		{
			name: "zero file permissions",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0,
				DirPermissions:  0755,
				SyncWrite:       true,
				CreateDir:       true,
			},
			expectedErr: true,
			errMsg:      "storage: file_permissions must be specified",
		},
		{
			name: "zero dir permissions",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0644,
				DirPermissions:  0,
				SyncWrite:       true,
				CreateDir:       true,
			},
			expectedErr: true,
			errMsg:      "storage: dir_permissions must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if tt.expectedErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	config := &Config{
		DataDir:         "./data",
		FilePermissions: 0644,
		DirPermissions:  0755,
		SyncWrite:       true,
		CreateDir:       true,
	}

	str := config.String()
	expected := "StorageConfig{DataDir: ./data, FilePermissions: 644, DirPermissions: 755, SyncWrite: true, CreateDir: true}"

	if str != expected {
		t.Errorf("String() = %s, expected %s", str, expected)
	}
}

func TestConfig_Clone(t *testing.T) {
	original := &Config{
		DataDir:         "./data",
		FilePermissions: 0644,
		DirPermissions:  0755,
		SyncWrite:       true,
		CreateDir:       true,
	}

	cloned := original.Clone()

	if cloned == nil {
		t.Fatal("Clone() returned nil")
	}

	// Type assert to get back to *Config
	clone, ok := cloned.(*Config)
	if !ok {
		t.Fatal("Clone() did not return *Config type")
	}

	// Check that values are the same
	if clone.DataDir != original.DataDir {
		t.Errorf("Expected DataDir to be '%s', got '%s'", original.DataDir, clone.DataDir)
	}

	if clone.FilePermissions != original.FilePermissions {
		t.Errorf("Expected FilePermissions to be %o, got %o", original.FilePermissions, clone.FilePermissions)
	}

	if clone.DirPermissions != original.DirPermissions {
		t.Errorf("Expected DirPermissions to be %o, got %o", original.DirPermissions, clone.DirPermissions)
	}

	if clone.SyncWrite != original.SyncWrite {
		t.Errorf("Expected SyncWrite to be %v, got %v", original.SyncWrite, clone.SyncWrite)
	}

	if clone.CreateDir != original.CreateDir {
		t.Errorf("Expected CreateDir to be %v, got %v", original.CreateDir, clone.CreateDir)
	}

	// Check that they are different objects
	if clone == original {
		t.Error("Clone() returned the same object reference")
	}

	// Modify clone and check that original is not affected
	clone.DataDir = "/modified/path"
	if original.DataDir == "/modified/path" {
		t.Error("Modifying clone affected original")
	}
}

func TestConfig_Clone_Nil(t *testing.T) {
	var original *Config
	clone := original.Clone()

	if clone != nil {
		t.Errorf("Expected Clone() of nil to return nil, got %v", clone)
	}
}

func TestConfig_Validate_EmptyDataDir(t *testing.T) {
	config := &Config{
		DataDir:         "", // Empty data directory
		FilePermissions: 0644,
		DirPermissions:  0755,
		SyncWrite:       true,
		CreateDir:       true,
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for empty data directory")
	}

	expectedErr := "storage: data_dir is required"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestConfig_Validate_InvalidPermissions(t *testing.T) {
	tests := []struct {
		name          string
		filePerm      os.FileMode
		dirPerm       os.FileMode
		expectedError string
	}{
		{
			name:          "zero file permissions",
			filePerm:      0,
			dirPerm:       0755,
			expectedError: "storage: file_permissions must be specified",
		},
		{
			name:          "zero directory permissions",
			filePerm:      0644,
			dirPerm:       0,
			expectedError: "storage: dir_permissions must be specified",
		},
		{
			name:          "file permissions too permissive",
			filePerm:      01000, // Beyond 0777
			dirPerm:       0755,
			expectedError: "storage: file_permissions too permissive",
		},
		{
			name:          "directory permissions too permissive",
			filePerm:      0644,
			dirPerm:       01000, // Beyond 0777
			expectedError: "storage: dir_permissions too permissive",
		},
		{
			name:          "directory permissions without read access",
			filePerm:      0644,
			dirPerm:       0300, // Execute and write but no read
			expectedError: "storage: dir_permissions must allow read access",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				DataDir:         "./data",
				FilePermissions: tt.filePerm,
				DirPermissions:  tt.dirPerm,
				SyncWrite:       true,
				CreateDir:       true,
			}

			err := config.Validate()
			if err == nil {
				t.Errorf("Expected validation error for %s", tt.name)
			}

			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedError, err.Error())
			}
		})
	}
}

func TestConfig_String_SensitiveInfo(t *testing.T) {
	// Storage config doesn't have sensitive info, but test string representation
	// and ensure it provides useful debugging information
	config := &Config{
		DataDir:         "/path/to/data",
		FilePermissions: 0644,
		DirPermissions:  0755,
		SyncWrite:       true,
		CreateDir:       false,
	}

	str := config.String()

	// Should contain all configuration values
	expectedSubstrings := []string{
		"StorageConfig{",
		"DataDir: /path/to/data",
		"FilePermissions: 644",
		"DirPermissions: 755",
		"SyncWrite: true",
		"CreateDir: false",
		"}",
	}

	for _, substring := range expectedSubstrings {
		if !strings.Contains(str, substring) {
			t.Errorf("String() should contain '%s', got '%s'", substring, str)
		}
	}

	// Test with nil config
	var nilConfig *Config
	nilStr := nilConfig.String()
	if nilStr != "<nil>" {
		t.Errorf("Expected <nil> for nil config string, got '%s'", nilStr)
	}
}

func TestConfig_Validate_Enhanced(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectedErr bool
		errMsg      string
	}{
		{
			name: "path too long",
			config: &Config{
				DataDir:         strings.Repeat("a", 4097), // Exceeds 4096 character limit
				FilePermissions: 0644,
				DirPermissions:  0755,
			},
			expectedErr: true,
			errMsg:      "data_dir path too long",
		},
		{
			name: "path with control characters",
			config: &Config{
				DataDir:         "./data\x00\x01", // Contains control characters
				FilePermissions: 0644,
				DirPermissions:  0755,
			},
			expectedErr: true,
			errMsg:      "data_dir contains invalid control characters",
		},
		{
			name: "file permissions at max allowed",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0777, // Exactly the max allowed
				DirPermissions:  0755,
			},
			expectedErr: false, // Should be valid (exactly at max)
		},
		{
			name: "directory permissions at max allowed",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0644,
				DirPermissions:  0777, // Exactly the max allowed
			},
			expectedErr: false, // Should be valid (exactly at max)
		},
		{
			name: "valid config with max permissions",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0777,
				DirPermissions:  0777,
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.expectedErr {
				t.Errorf("Validate() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if tt.expectedErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}
