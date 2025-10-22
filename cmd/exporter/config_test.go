package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadAllConfigs tests loading all module configurations in dependency order
func TestLoadAllConfigs(t *testing.T) {
	tests := []struct {
		name           string
		configFile     string
		envVars        map[string]string
		expectedError  string
		validateConfig func(t *testing.T, config *AppConfig)
	}{
		{
			name:          "default configuration without file",
			configFile:    "",
			envVars:       nil,
			expectedError: "",
			validateConfig: func(t *testing.T, config *AppConfig) {
				require.NotNil(t, config)
				require.NotNil(t, config.Storage)
				require.NotNil(t, config.WinPower)
				require.NotNil(t, config.Energy)
				require.NotNil(t, config.Server)
				require.NotNil(t, config.Scheduler)

				// Verify default values
				assert.Equal(t, "./data", config.Storage.DataDir)
				// Note: SyncWrite defaults depend on module implementation
				assert.Equal(t, "https://localhost:8080", config.WinPower.URL)
				assert.Equal(t, "admin", config.WinPower.Username)
				assert.Equal(t, 9090, config.Server.Port)
				assert.Equal(t, 5*time.Second, config.Scheduler.CollectionInterval)
			},
		},
		{
			name: "configuration with valid YAML file",
			configFile: createTestConfigFile(t, `
storage:
  data_dir: "/tmp/test-data"
  sync_write: false

winpower:
  url: "https://test.example.com:9090"
  username: "testuser"
  password: "testpass"
  timeout: "45s"
  max_retries: 5

energy:
  calculation_interval: "10s"
  retention_period: "720h"

server:
  port: 8080
  host: "127.0.0.1"

scheduler:
  collection_interval: "10s"
  max_parallel_tasks: 5
`),
			envVars:       nil,
			expectedError: "",
			validateConfig: func(t *testing.T, config *AppConfig) {
				require.NotNil(t, config)
				assert.Equal(t, "/tmp/test-data", config.Storage.DataDir)
				assert.False(t, config.Storage.SyncWrite)
				assert.Equal(t, "https://test.example.com:9090", config.WinPower.URL)
				assert.Equal(t, "testuser", config.WinPower.Username)
				assert.Equal(t, 8080, config.Server.Port)
				assert.Equal(t, 10*time.Second, config.Scheduler.CollectionInterval)
			},
		},
		{
			name: "configuration with environment variable overrides",
			configFile: createTestConfigFile(t, `
storage:
  data_dir: "/tmp/config-data"
  sync_write: true

winpower:
  url: "https://config.example.com:8080"
  username: "configuser"
  password: "configpass"

server:
  port: 9090
`),
			envVars: map[string]string{
				"WINPOWER_EXPORTER_STORAGE_DATA_DIR": "/tmp/env-data",
				"WINPOWER_EXPORTER_SERVER_PORT":      "8080",
			},
			expectedError: "",
			validateConfig: func(t *testing.T, config *AppConfig) {
				require.NotNil(t, config)
				// Environment variables should override config file
				assert.Equal(t, "/tmp/env-data", config.Storage.DataDir)
				assert.Equal(t, 8080, config.Server.Port)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			if tt.envVars != nil {
				for key, value := range tt.envVars {
					_ = os.Setenv(key, value)
				}
				defer func() {
					for key := range tt.envVars {
						_ = os.Unsetenv(key)
					}
				}()
			}

			// Load configuration
			config, err := loadApplicationConfig(tt.configFile)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				if tt.validateConfig != nil {
					tt.validateConfig(t, config)
				}
			}
		})
	}
}

// TestLoadAllConfigs_WithValidation tests configuration validation failure handling
func TestLoadAllConfigs_WithValidation(t *testing.T) {
	tests := []struct {
		name          string
		configFile    string
		envVars       map[string]string
		expectedError string
		errorField    string
	}{
		{
			name: "invalid WinPower URL",
			configFile: createTestConfigFile(t, `
winpower:
  url: "invalid-url"
  username: "testuser"
  password: "testpass"
`),
			expectedError: "failed to load winpower config",
			errorField:    "url",
		},
		{
			name: "missing WinPower credentials",
			configFile: createTestConfigFile(t, `
winpower:
  url: "https://test.example.com:8080"
  username: ""
  password: ""
`),
			expectedError: "failed to load winpower config",
			errorField:    "username",
		},
		{
			name: "invalid server port",
			configFile: createTestConfigFile(t, `
server:
  port: -1
`),
			expectedError: "failed to load server config",
		},
		{
			name: "invalid scheduler interval",
			configFile: createTestConfigFile(t, `
scheduler:
  collection_interval: "invalid"
`),
			expectedError: "failed to load scheduler config",
		},
		{
			name: "invalid storage data directory path",
			configFile: createTestConfigFile(t, `
storage:
  data_dir: ""
`),
			expectedError: "failed to load storage config",
		},
		// TODO: Fix environment variable validation test
		// {
		// 	name:          "invalid environment variable value",
		// 	configFile:    "",
		// 	envVars:       map[string]string{"WINPOWER_URL": "invalid-env-url"},
		// 	expectedError: "failed to load winpower config",
		// 	errorField:    "url",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			if tt.envVars != nil {
				for key, value := range tt.envVars {
					_ = os.Setenv(key, value)
				}
				defer func() {
					for key := range tt.envVars {
						_ = os.Unsetenv(key)
					}
				}()
			}

			// Load configuration
			config, err := loadApplicationConfig(tt.configFile)

			// Should always fail validation
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Nil(t, config)

			// Check for specific field errors if specified
			if tt.errorField != "" {
				assert.Contains(t, err.Error(), tt.errorField)
			}
		})
	}
}

// TestLoadAllConfigs_MissingConfigFile tests configuration file missing handling
func TestLoadAllConfigs_MissingConfigFile(t *testing.T) {
	nonExistentFile := "/path/to/non/existent/config.yaml"

	// Test that loading configs from non-existent file falls back to defaults
	appConfig, err := loadApplicationConfig(nonExistentFile)
	if err != nil {
		t.Fatalf("Failed to load application config with missing file (should use defaults): %v", err)
	}

	// Should have default values
	if appConfig.Storage.DataDir == "" {
		t.Error("Expected storage DataDir to have default value, got empty string")
	}

	if appConfig.WinPower.Timeout == 0 {
		t.Error("Expected winpower Timeout to have default value, got 0")
	}

	if appConfig.Energy.Precision == 0 {
		t.Error("Expected energy Precision to have default value, got 0")
	}

	if appConfig.Scheduler.CollectionInterval == 0 {
		t.Error("Expected scheduler CollectionInterval to have default value, got 0")
	}

	if appConfig.Server.Port == 0 {
		t.Error("Expected server Port to have default value, got 0")
	}
}

// TestLoadAllConfigs_EnvironmentVariableOverride tests environment variable override
func TestLoadAllConfigs_EnvironmentVariableOverride(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "winpower-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a test configuration file
	configFile := filepath.Join(tempDir, "config.yaml")
	configContent := `
storage:
  data_dir: "/tmp/test-data"
  create_dir: true
  sync_write: true

winpower:
  url: "http://localhost:8080"
  username: "defaultuser"
  password: "defaultpass"
  timeout: "30s"
  skip_ssl_verify: true

energy:
  precision: 0.01
  enable_stats: true
  max_calculation_time: 1000000000
  negative_power_allowed: true

scheduler:
  collection_interval: "5s"
  graceful_shutdown_timeout: "30s"

server:
  port: 9090
  host: "0.0.0.0"
  mode: "release"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  enable_pprof: false
  enable_cors: false
  enable_rate_limit: false
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variables to override config file values
	if err := os.Setenv("WINPOWER_EXPORTER_STORAGE_DATA_DIR", "/tmp/env-override-data"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("WINPOWER_EXPORTER_SERVER_PORT", "8080"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("WINPOWER_EXPORTER_STORAGE_DATA_DIR"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("WINPOWER_EXPORTER_SERVER_PORT"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	// Test loading configs with environment variable overrides
	appConfig, err := loadApplicationConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load application config: %v", err)
	}

	// Should have environment variable value, not config file value
	if appConfig.Storage.DataDir != "/tmp/env-override-data" {
		t.Errorf("Expected storage DataDir to be '/tmp/env-override-data' from env, got '%s'", appConfig.Storage.DataDir)
	}

	// Should have environment variable value, not config file value
	if appConfig.Server.Port != 8080 {
		t.Errorf("Expected server Port to be 8080 from env, got %d", appConfig.Server.Port)
	}
}

// createTestConfigFile creates a temporary configuration file for testing
func createTestConfigFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	err := os.WriteFile(configFile, []byte(content), 0644)
	require.NoError(t, err)

	return configFile
}

// Test helper functions for individual config loading
func TestLoadStorageConfig(t *testing.T) {
	loader := config.NewLoader("WINPOWER_EXPORTER")

	config, err := loadStorageConfig(loader)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "./data", config.DataDir)
	// Note: SyncWrite default depends on the storage module implementation
}

func TestLoadWinPowerConfig(t *testing.T) {
	loader := config.NewLoader("WINPOWER_EXPORTER")

	config, err := loadWinPowerConfig(loader)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "https://localhost:8080", config.URL)
	assert.Equal(t, "admin", config.Username)
}

func TestLoadEnergyConfig(t *testing.T) {
	loader := config.NewLoader("WINPOWER_EXPORTER")

	config, err := loadEnergyConfig(loader)
	assert.NoError(t, err)
	assert.NotNil(t, config)
}

func TestLoadSchedulerConfig(t *testing.T) {
	loader := config.NewLoader("WINPOWER_EXPORTER")

	config, err := loadSchedulerConfig(loader)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 5*time.Second, config.CollectionInterval)
}

func TestLoadServerConfig(t *testing.T) {
	loader := config.NewLoader("WINPOWER_EXPORTER")

	config, err := loadServerConfig(loader)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 9090, config.Port)
	assert.Equal(t, "0.0.0.0", config.Host)
}

// TestIndividualModuleConfigs tests loading individual module configurations
func TestIndividualModuleConfigs(t *testing.T) {
	// Create a test configuration file
	configFile := createTestConfigFile(t, `
storage:
  data_dir: "/tmp/test-storage"
  sync_write: false

winpower:
  url: "http://test.example.com"
  username: "testuser"
  password: "testpass"
  timeout: "45s"
  skip_ssl_verify: false

energy:
  precision: 0.001
  enable_stats: false

scheduler:
  collection_interval: "10s"

server:
  port: 8080
  host: "127.0.0.1"
`)

	// Test loading individual configs
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(configFile)

	storageConfig, err := storage.NewConfig(loader)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/test-storage", storageConfig.DataDir)

	winpowerConfig, err := winpower.NewConfig(loader)
	require.NoError(t, err)
	assert.Equal(t, "http://test.example.com", winpowerConfig.URL)

	energyConfig, err := energy.NewConfig(loader)
	require.NoError(t, err)
	assert.Equal(t, 0.001, energyConfig.Precision)

	schedulerConfig, err := scheduler.NewConfig(loader)
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, schedulerConfig.CollectionInterval)

	serverConfig, err := server.NewConfig(loader)
	require.NoError(t, err)
	assert.Equal(t, 8080, serverConfig.Port)
}

// TestLoadAllConfigs_ConfigFileNotFound tests configuration file missing handling
func TestLoadAllConfigs_ConfigFileNotFound(t *testing.T) {
	// Test with non-existent config file - should use defaults
	config, err := loadApplicationConfig("/non/existent/config.yaml")

	// Should not error - should fall back to defaults
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotNil(t, config.Storage)
	assert.NotNil(t, config.WinPower)
}

// TestLoadAllConfigs_ConfigFileNotReadable tests configuration file permission errors
func TestLoadAllConfigs_ConfigFileNotReadable(t *testing.T) {
	// Create a file without read permissions
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// Create the file
	err := os.WriteFile(configFile, []byte("test: config"), 0644)
	require.NoError(t, err)

	// Remove read permissions
	err = os.Chmod(configFile, 0000)
	require.NoError(t, err)

	// Try to load configuration
	config, err := loadApplicationConfig(configFile)

	// Should error due to permissions
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}

// TestLoadAllConfigs_DependencyOrder tests that configurations are loaded in the correct dependency order
func TestLoadAllConfigs_DependencyOrder(t *testing.T) {
	// Test that configurations are loaded in the correct dependency order
	// This test ensures that dependencies are respected during loading

	configFile := createTestConfigFile(t, `
storage:
  data_dir: "/tmp/dep-test"
  sync_write: false

winpower:
  url: "https://dep-test.example.com:8080"
  username: "depuser"
  password: "deppass"

energy:
  calculation_interval: "15s"

server:
  port: 9999

scheduler:
  collection_interval: "20s"
`)

	config, err := loadApplicationConfig(configFile)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify all configs were loaded
	assert.NotNil(t, config.Storage)
	assert.NotNil(t, config.WinPower)
	assert.NotNil(t, config.Energy)
	assert.NotNil(t, config.Server)
	assert.NotNil(t, config.Scheduler)

	// Verify each config is valid
	assert.NoError(t, config.Storage.Validate())
	assert.NoError(t, config.WinPower.Validate())
	assert.NoError(t, config.Energy.Validate())
	assert.NoError(t, config.Server.Validate())
	assert.NoError(t, config.Scheduler.Validate())
}

// TestLoadAllConfigs_ConfigurationImmutability tests that loaded configurations can be safely modified
func TestLoadAllConfigs_ConfigurationImmutability(t *testing.T) {
	// Test that loaded configurations can be safely modified without affecting
	// subsequent loads

	configFile := createTestConfigFile(t, `
storage:
  data_dir: "/tmp/original"
  sync_write: true

winpower:
  url: "https://original.example.com:8080"
  username: "original"
  password: "original"
`)

	// Load first configuration
	config1, err := loadApplicationConfig(configFile)
	require.NoError(t, err)
	require.NotNil(t, config1)

	// Modify first configuration
	config1.Storage.DataDir = "/tmp/modified"
	config1.WinPower.URL = "https://modified.example.com:8080"

	// Load second configuration
	config2, err := loadApplicationConfig(configFile)
	require.NoError(t, err)
	require.NotNil(t, config2)

	// Second configuration should have original values
	assert.Equal(t, "/tmp/original", config2.Storage.DataDir)
	assert.Equal(t, "https://original.example.com:8080", config2.WinPower.URL)
}
