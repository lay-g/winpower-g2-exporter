package test

import (
	"os"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// TestCompleteConfigIntegration tests the complete configuration loading process
// for all modules using both YAML configuration and environment variables.
func TestCompleteConfigIntegration(t *testing.T) {
	// Create a comprehensive YAML configuration
	yamlContent := `
storage:
  data_dir: "/tmp/test-integration"
  file_permissions: 0640
  dir_permissions: 0750
  sync_write: false
  create_dir: true

winpower:
  url: "https://winpower.test.com:8080"
  username: "testuser"
  password: "testpass123"
  timeout: "30s"
  max_retries: 3
  skip_tls_verify: false

energy:
  precision: 0.005
  enable_stats: true
  max_calculation_time: 2000000000
  negative_power_allowed: false

server:
  host: "127.0.0.1"
  port: 9091
  mode: "release"
  read_timeout: "10s"
  write_timeout: "10s"
  idle_timeout: "60s"
  debug_mode: false
  test_mode: false
  pprof_enabled: false
  cors_enabled: true
  rate_limit_enabled: false
  rate_limit_rps: 100
  trusted_proxies: []

scheduler:
  collection_interval: 10s
  graceful_shutdown_timeout: 30s
`

	tmpFile, err := os.CreateTemp("", "complete-config-*.yaml")
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

	// Create loader and load all configurations
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(tmpFile.Name())

	// Test Storage module
	storageConfig, err := storage.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load storage config: %v", err)
	}

	if storageConfig.DataDir != "/tmp/test-integration" {
		t.Errorf("Expected storage DataDir to be '/tmp/test-integration', got '%s'", storageConfig.DataDir)
	}

	if storageConfig.FilePermissions != 0640 {
		t.Errorf("Expected storage FilePermissions to be 0640, got %o", storageConfig.FilePermissions)
	}

	// Test WinPower module
	winpowerConfig, err := winpower.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load winpower config: %v", err)
	}

	if winpowerConfig.URL != "https://winpower.test.com:8080" {
		t.Errorf("Expected winpower URL to be 'https://winpower.test.com:8080', got '%s'", winpowerConfig.URL)
	}

	if winpowerConfig.Username != "testuser" {
		t.Errorf("Expected winpower Username to be 'testuser', got '%s'", winpowerConfig.Username)
	}

	// Test Energy module
	energyConfig, err := energy.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load energy config: %v", err)
	}

	if energyConfig.Precision != 0.005 {
		t.Errorf("Expected energy Precision to be 0.005, got %f", energyConfig.Precision)
	}

	if !energyConfig.EnableStats {
		t.Errorf("Expected energy EnableStats to be true, got false")
	}

	// Test Server module
	serverConfig, err := server.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load server config: %v", err)
	}

	if serverConfig.Host != "127.0.0.1" {
		t.Errorf("Expected server Host to be '127.0.0.1', got '%s'", serverConfig.Host)
	}

	if serverConfig.Port != 9091 {
		t.Errorf("Expected server Port to be 9091, got %d", serverConfig.Port)
	}

	// Test Scheduler module
	schedulerConfig, err := scheduler.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load scheduler config: %v", err)
	}

	if schedulerConfig.CollectionInterval != 10*time.Second {
		t.Errorf("Expected scheduler CollectionInterval to be 10s, got %v", schedulerConfig.CollectionInterval)
	}

	if schedulerConfig.GracefulShutdownTimeout != 30*time.Second {
		t.Errorf("Expected scheduler GracefulShutdownTimeout to be 30s, got %v", schedulerConfig.GracefulShutdownTimeout)
	}
}

// TestEnvironmentVariableOverrides tests that environment variables can override YAML values
func TestEnvironmentVariableOverrides(t *testing.T) {
	// Create a basic YAML configuration
	yamlContent := `
storage:
  data_dir: "/yaml/data"
  sync_write: true

winpower:
  url: "https://yaml.example.com"
  username: "yamluser"

energy:
  precision: 0.01
  enable_stats: false

server:
  port: 9090
  debug_mode: false

scheduler:
  collection_interval: 5s
`

	tmpFile, err := os.CreateTemp("", "env-override-*.yaml")
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

	// Set environment variables that should override YAML values
	envVars := map[string]string{
		"STORAGE_DATA_DIR":              "/env/data",
		"STORAGE_SYNC_WRITE":            "false",
		"WINPOWER_URL":                  "https://env.example.com",
		"WINPOWER_USERNAME":             "envuser",
		"ENERGY_PRECISION":              "0.001",
		"ENERGY_ENABLE_STATS":           "true",
		"SERVER_PORT":                   "9091",
		"SERVER_ENABLE_PPROF":           "true",
		"SCHEDULER_COLLECTION_INTERVAL": "10s",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set env var %s: %v", key, err)
		}
		defer func(k string) {
			if err := os.Unsetenv(k); err != nil {
				t.Logf("Warning: failed to unset env var %s: %v", k, err)
			}
		}(key)
	}

	// Create loader and load configurations
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(tmpFile.Name())

	// Test that environment variables override YAML values
	storageConfig, err := storage.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load storage config: %v", err)
	}

	if storageConfig.DataDir != "/env/data" {
		t.Errorf("Expected storage DataDir to be overridden to '/env/data', got '%s'", storageConfig.DataDir)
	}

	if storageConfig.SyncWrite {
		t.Errorf("Expected storage SyncWrite to be overridden to false, got true")
	}

	winpowerConfig, err := winpower.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load winpower config: %v", err)
	}

	if winpowerConfig.URL != "https://env.example.com" {
		t.Errorf("Expected winpower URL to be overridden to 'https://env.example.com', got '%s'", winpowerConfig.URL)
	}

	energyConfig, err := energy.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load energy config: %v", err)
	}

	if energyConfig.Precision != 0.001 {
		t.Errorf("Expected energy Precision to be overridden to 0.001, got %f", energyConfig.Precision)
	}

	if !energyConfig.EnableStats {
		t.Errorf("Expected energy EnableStats to be overridden to true, got false")
	}

	serverConfig, err := server.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load server config: %v", err)
	}

	if serverConfig.Port != 9091 {
		t.Errorf("Expected server Port to be overridden to 9091, got %d", serverConfig.Port)
	}

	if !serverConfig.EnablePprof {
		t.Errorf("Expected server EnablePprof to be overridden to true, got false")
	}

	schedulerConfig, err := scheduler.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load scheduler config: %v", err)
	}

	if schedulerConfig.CollectionInterval != 10*time.Second {
		t.Errorf("Expected scheduler CollectionInterval to be overridden to 10s, got %v", schedulerConfig.CollectionInterval)
	}
}

// TestConfigurationValidation tests that all configurations can be validated
func TestConfigurationValidation(t *testing.T) {
	// Create loader without config file (will use defaults)
	loader := config.NewLoader("WINPOWER_EXPORTER")

	// Test that all default configurations are valid
	storageConfig, err := storage.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load storage config: %v", err)
	}

	if err := storageConfig.Validate(); err != nil {
		t.Errorf("Default storage config should be valid: %v", err)
	}

	winpowerConfig, err := winpower.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load winpower config: %v", err)
	}

	// Note: Default winpower config should pass validation because it has default values
	if err := winpowerConfig.Validate(); err != nil {
		t.Errorf("Default winpower config should be valid with defaults: %v", err)
	}

	energyConfig, err := energy.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load energy config: %v", err)
	}

	if err := energyConfig.Validate(); err != nil {
		t.Errorf("Default energy config should be valid: %v", err)
	}

	serverConfig, err := server.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load server config: %v", err)
	}

	if err := serverConfig.Validate(); err != nil {
		t.Errorf("Default server config should be valid: %v", err)
	}

	schedulerConfig, err := scheduler.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load scheduler config: %v", err)
	}

	if err := schedulerConfig.Validate(); err != nil {
		t.Errorf("Default scheduler config should be valid: %v", err)
	}
}

// TestConfigurationClone tests that all configurations can be cloned properly
func TestConfigurationClone(t *testing.T) {
	loader := config.NewLoader("WINPOWER_EXPORTER")

	// Test Storage config cloning
	storageConfig, err := storage.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load storage config: %v", err)
	}

	storageClone := storageConfig.Clone()
	if storageClone == nil {
		t.Fatal("Storage config clone returned nil")
	}

	if storageClone.(*storage.Config).DataDir != storageConfig.DataDir {
		t.Error("Storage config clone does not match original")
	}

	// Test Energy config cloning
	energyConfig, err := energy.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load energy config: %v", err)
	}

	energyClone := energyConfig.Clone()
	if energyClone == nil {
		t.Fatal("Energy config clone returned nil")
	}

	if energyClone.(*energy.Config).Precision != energyConfig.Precision {
		t.Error("Energy config clone does not match original")
	}

	// Test Server config cloning
	serverConfig, err := server.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load server config: %v", err)
	}

	serverClone := serverConfig.Clone()
	if serverClone == nil {
		t.Fatal("Server config clone returned nil")
	}

	if serverClone.(*server.Config).Port != serverConfig.Port {
		t.Error("Server config clone does not match original")
	}

	// Test Scheduler config cloning
	schedulerConfig, err := scheduler.NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load scheduler config: %v", err)
	}

	schedulerClone := schedulerConfig.Clone()
	if schedulerClone == nil {
		t.Fatal("Scheduler config clone returned nil")
	}

	if schedulerClone.(*scheduler.Config).CollectionInterval != schedulerConfig.CollectionInterval {
		t.Error("Scheduler config clone does not match original")
	}
}
