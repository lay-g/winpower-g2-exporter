package storage

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("NewConfig() returned nil")
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

	clone := original.Clone()

	if clone == nil {
		t.Fatal("Clone() returned nil")
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
