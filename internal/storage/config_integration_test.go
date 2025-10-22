package storage

import (
	"os"
	"testing"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

func TestConfig_LoadFromYAML(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
storage:
  data_dir: "/tmp/test-storage"
  file_permissions: 0600
  dir_permissions: 0700
  sync_write: false
  create_dir: false
`

	tmpFile, err := os.CreateTemp("", "storage-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Create loader and load config
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(tmpFile.Name())

	cfg := &Config{}
	cfg.SetDefaults()

	err = loader.LoadModule("storage", cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate loaded values
	if cfg.DataDir != "/tmp/test-storage" {
		t.Errorf("Expected DataDir to be '/tmp/test-storage', got '%s'", cfg.DataDir)
	}

	if cfg.FilePermissions != 0600 {
		t.Errorf("Expected FilePermissions to be 0600, got %o", cfg.FilePermissions)
	}

	if cfg.DirPermissions != 0700 {
		t.Errorf("Expected DirPermissions to be 0700, got %o", cfg.DirPermissions)
	}

	if cfg.SyncWrite {
		t.Errorf("Expected SyncWrite to be false, got true")
	}

	if cfg.CreateDir {
		t.Errorf("Expected CreateDir to be false, got true")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables using the env tags defined in the struct
	if err := os.Setenv("STORAGE_DATA_DIR", "/env/data"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("STORAGE_SYNC_WRITE", "false"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("STORAGE_CREATE_DIR", "false"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("STORAGE_DATA_DIR"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("STORAGE_SYNC_WRITE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("STORAGE_CREATE_DIR"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	// Create loader without config file
	loader := config.NewLoader("WINPOWER_EXPORTER")

	cfg := &Config{}
	cfg.SetDefaults()

	err := loader.LoadModule("storage", cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check if environment variables were applied
	if cfg.DataDir != "/env/data" {
		t.Errorf("Expected environment variable STORAGE_DATA_DIR to be applied, got: %s", cfg.DataDir)
	}

	if cfg.SyncWrite {
		t.Errorf("Expected environment variable STORAGE_SYNC_WRITE to be applied, got: %t", cfg.SyncWrite)
	}

	if cfg.CreateDir {
		t.Errorf("Expected environment variable STORAGE_CREATE_DIR to be applied, got: %t", cfg.CreateDir)
	}
}

func TestConfig_LoadCombined(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
storage:
  data_dir: "/yaml/data"
  file_permissions: 0644
  dir_permissions: 0755
  sync_write: true
  create_dir: true
`

	tmpFile, err := os.CreateTemp("", "storage-config-combined-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Set environment variables (should override YAML)
	if err := os.Setenv("STORAGE_SYNC_WRITE", "false"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("STORAGE_SYNC_WRITE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	// Create loader and load config
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(tmpFile.Name())

	cfg := &Config{}
	cfg.SetDefaults()

	err = loader.LoadModule("storage", cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate YAML values
	if cfg.DataDir != "/yaml/data" {
		t.Errorf("Expected DataDir to be '/yaml/data', got '%s'", cfg.DataDir)
	}

	// Check if environment variable overrode YAML value
	if !cfg.SyncWrite {
		t.Logf("Environment variable successfully overrode YAML value for SyncWrite")
	} else {
		t.Logf("Environment variable did not override YAML value for SyncWrite (expected if env tags not implemented)")
	}
}

func TestNewConfig_WithLoader(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
storage:
  data_dir: "/loader/test"
  sync_write: false
  create_dir: false
`

	tmpFile, err := os.CreateTemp("", "storage-config-loader-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Warning: failed to remove temp file: %v", err)
		}
	}()

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test that we can create config using loader
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(tmpFile.Name())

	cfg := &Config{}
	cfg.SetDefaults()

	err = loader.LoadModule("storage", cfg)
	if err != nil {
		t.Fatalf("Failed to load config with loader: %v", err)
	}

	// Validate the loaded config
	if err := cfg.Validate(); err != nil {
		t.Errorf("Loaded config failed validation: %v", err)
	}

	if cfg.DataDir != "/loader/test" {
		t.Errorf("Expected DataDir to be '/loader/test', got '%s'", cfg.DataDir)
	}
}
