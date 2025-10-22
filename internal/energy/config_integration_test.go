package energy

import (
	"os"
	"testing"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

func TestConfig_LoadFromYAML(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
energy:
  precision: 0.001
  enable_stats: false
  max_calculation_time: 2000000000
  negative_power_allowed: false
`

	tmpFile, err := os.CreateTemp("", "energy-config-*.yaml")
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

	cfg, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate loaded values
	if cfg.Precision != 0.001 {
		t.Errorf("Expected Precision to be 0.001, got %f", cfg.Precision)
	}

	if cfg.EnableStats {
		t.Errorf("Expected EnableStats to be false, got true")
	}

	if cfg.MaxCalculationTime != 2000000000 {
		t.Errorf("Expected MaxCalculationTime to be 2000000000, got %d", cfg.MaxCalculationTime)
	}

	if cfg.NegativePowerAllowed {
		t.Errorf("Expected NegativePowerAllowed to be false, got true")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("ENERGY_PRECISION", "0.005"); err != nil {
		t.Fatalf("Failed to set ENERGY_PRECISION env var: %v", err)
	}
	if err := os.Setenv("ENERGY_ENABLE_STATS", "false"); err != nil {
		t.Fatalf("Failed to set ENERGY_ENABLE_STATS env var: %v", err)
	}
	if err := os.Setenv("ENERGY_MAX_CALCULATION_TIME", "5000000000"); err != nil {
		t.Fatalf("Failed to set ENERGY_MAX_CALCULATION_TIME env var: %v", err)
	}
	if err := os.Setenv("ENERGY_NEGATIVE_POWER_ALLOWED", "false"); err != nil {
		t.Fatalf("Failed to set ENERGY_NEGATIVE_POWER_ALLOWED env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ENERGY_PRECISION"); err != nil {
			t.Logf("Warning: failed to unset ENERGY_PRECISION: %v", err)
		}
		if err := os.Unsetenv("ENERGY_ENABLE_STATS"); err != nil {
			t.Logf("Warning: failed to unset ENERGY_ENABLE_STATS: %v", err)
		}
		if err := os.Unsetenv("ENERGY_MAX_CALCULATION_TIME"); err != nil {
			t.Logf("Warning: failed to unset ENERGY_MAX_CALCULATION_TIME: %v", err)
		}
		if err := os.Unsetenv("ENERGY_NEGATIVE_POWER_ALLOWED"); err != nil {
			t.Logf("Warning: failed to unset ENERGY_NEGATIVE_POWER_ALLOWED: %v", err)
		}
	}()

	// Create loader without config file
	loader := config.NewLoader("WINPOWER_EXPORTER")

	cfg, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check if environment variables were applied
	if cfg.Precision != 0.005 {
		t.Errorf("Expected environment variable ENERGY_PRECISION to be applied, got: %f", cfg.Precision)
	}

	if cfg.EnableStats {
		t.Errorf("Expected environment variable ENERGY_ENABLE_STATS to be applied, got: %t", cfg.EnableStats)
	}

	if cfg.MaxCalculationTime != 5000000000 {
		t.Errorf("Expected environment variable ENERGY_MAX_CALCULATION_TIME to be applied, got: %d", cfg.MaxCalculationTime)
	}

	if cfg.NegativePowerAllowed {
		t.Errorf("Expected environment variable ENERGY_NEGATIVE_POWER_ALLOWED to be applied, got: %t", cfg.NegativePowerAllowed)
	}
}

func TestConfig_LoadCombined(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
energy:
  precision: 0.01
  enable_stats: true
  max_calculation_time: 1000000000
  negative_power_allowed: true
`

	tmpFile, err := os.CreateTemp("", "energy-config-combined-*.yaml")
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
	if err := os.Setenv("ENERGY_ENABLE_STATS", "false"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("ENERGY_ENABLE_STATS"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	// Create loader and load config
	loader := config.NewLoader("WINPOWER_EXPORTER")
	loader.SetConfigFile(tmpFile.Name())

	cfg, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Validate YAML values
	if cfg.Precision != 0.01 {
		t.Errorf("Expected Precision to be 0.01, got %f", cfg.Precision)
	}

	// Check if environment variable overrode YAML value
	if !cfg.EnableStats {
		t.Logf("Environment variable successfully overrode YAML value for EnableStats")
	} else {
		t.Errorf("Expected environment variable to override YAML value for EnableStats")
	}
}

func TestNewConfig_WithLoader(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
energy:
  precision: 0.02
  enable_stats: false
  negative_power_allowed: false
`

	tmpFile, err := os.CreateTemp("", "energy-config-loader-*.yaml")
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

	cfg, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load config with loader: %v", err)
	}

	// Validate the loaded config
	if err := cfg.Validate(); err != nil {
		t.Errorf("Loaded config failed validation: %v", err)
	}

	if cfg.Precision != 0.02 {
		t.Errorf("Expected Precision to be 0.02, got %f", cfg.Precision)
	}

	if cfg.EnableStats {
		t.Errorf("Expected EnableStats to be false, got true")
	}

	if cfg.NegativePowerAllowed {
		t.Errorf("Expected NegativePowerAllowed to be false, got true")
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	// Create loader without config file
	loader := config.NewLoader("WINPOWER_EXPORTER")

	cfg, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check default values
	defaultCfg := DefaultConfig()
	if cfg.Precision != defaultCfg.Precision {
		t.Errorf("Expected default Precision to be %f, got %f", defaultCfg.Precision, cfg.Precision)
	}

	if cfg.EnableStats != defaultCfg.EnableStats {
		t.Errorf("Expected default EnableStats to be %t, got %t", defaultCfg.EnableStats, cfg.EnableStats)
	}

	if cfg.MaxCalculationTime != defaultCfg.MaxCalculationTime {
		t.Errorf("Expected default MaxCalculationTime to be %d, got %d", defaultCfg.MaxCalculationTime, cfg.MaxCalculationTime)
	}

	if cfg.NegativePowerAllowed != defaultCfg.NegativePowerAllowed {
		t.Errorf("Expected default NegativePowerAllowed to be %t, got %t", defaultCfg.NegativePowerAllowed, cfg.NegativePowerAllowed)
	}
}
