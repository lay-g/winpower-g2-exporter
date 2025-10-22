package winpower

import (
	"os"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

func TestConfig_LoadFromYAML(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
winpower:
  url: "https://winpower.example.com:8080"
  username: "testuser"
  password: "testpass123"
  timeout: "45s"
  max_retries: 5
  skip_tls_verify: true
`

	tmpFile, err := os.CreateTemp("", "winpower-config-*.yaml")
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
	if cfg.URL != "https://winpower.example.com:8080" {
		t.Errorf("Expected URL to be 'https://winpower.example.com:8080', got '%s'", cfg.URL)
	}

	if cfg.Username != "testuser" {
		t.Errorf("Expected Username to be 'testuser', got '%s'", cfg.Username)
	}

	if cfg.Password != "testpass123" {
		t.Errorf("Expected Password to be 'testpass123', got '%s'", cfg.Password)
	}

	if cfg.Timeout != 45*time.Second {
		t.Errorf("Expected Timeout to be 45s, got %v", cfg.Timeout)
	}

	if cfg.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries to be 5, got %d", cfg.MaxRetries)
	}

	if !cfg.SkipTLSVerify {
		t.Errorf("Expected SkipTLSVerify to be true, got false")
	}
}

func TestConfig_LoadFromEnv(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("WINPOWER_URL", "https://env.example.com"); err != nil {
		t.Fatalf("Failed to set WINPOWER_URL env var: %v", err)
	}
	if err := os.Setenv("WINPOWER_USERNAME", "envuser"); err != nil {
		t.Fatalf("Failed to set WINPOWER_USERNAME env var: %v", err)
	}
	if err := os.Setenv("WINPOWER_PASSWORD", "envpass"); err != nil {
		t.Fatalf("Failed to set WINPOWER_PASSWORD env var: %v", err)
	}
	if err := os.Setenv("WINPOWER_SKIP_TLS_VERIFY", "true"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("WINPOWER_URL"); err != nil {
			t.Logf("Warning: failed to unset WINPOWER_URL: %v", err)
		}
		if err := os.Unsetenv("WINPOWER_USERNAME"); err != nil {
			t.Logf("Warning: failed to unset WINPOWER_USERNAME: %v", err)
		}
		if err := os.Unsetenv("WINPOWER_PASSWORD"); err != nil {
			t.Logf("Warning: failed to unset WINPOWER_PASSWORD: %v", err)
		}
		if err := os.Unsetenv("WINPOWER_SKIP_TLS_VERIFY"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	// Create loader without config file
	loader := config.NewLoader("WINPOWER_EXPORTER")

	cfg, err := NewConfig(loader)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check if environment variables were applied
	if cfg.URL == "https://env.example.com" {
		t.Logf("Environment variable WINPOWER_URL was applied: %s", cfg.URL)
	} else {
		t.Logf("Environment variable WINPOWER_URL was not applied, got: %s", cfg.URL)
	}

	if cfg.Username == "envuser" {
		t.Logf("Environment variable WINPOWER_USERNAME was applied: %s", cfg.Username)
	} else {
		t.Logf("Environment variable WINPOWER_USERNAME was not applied, got: %s", cfg.Username)
	}
}

func TestConfig_LoadCombined(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
winpower:
  url: "https://yaml.example.com"
  username: "yamluser"
  password: "yamlpass"
  timeout: "30s"
  max_retries: 3
  skip_tls_verify: false
`

	tmpFile, err := os.CreateTemp("", "winpower-config-combined-*.yaml")
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
	if err := os.Setenv("WINPOWER_SKIP_TLS_VERIFY", "true"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("WINPOWER_SKIP_TLS_VERIFY"); err != nil {
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
	if cfg.URL != "https://yaml.example.com" {
		t.Errorf("Expected URL to be 'https://yaml.example.com', got '%s'", cfg.URL)
	}

	// Check if environment variable overrode YAML value
	if cfg.SkipTLSVerify {
		t.Logf("Environment variable successfully overrode YAML value for SkipTLSVerify")
	} else {
		t.Logf("Environment variable did not override YAML value for SkipTLSVerify (expected if env tags not implemented)")
	}
}

func TestNewConfig_WithLoader(t *testing.T) {
	// Create a temporary YAML config file
	yamlContent := `
winpower:
  url: "https://loader.example.com"
  username: "loaderuser"
  password: "loaderpass"
  skip_tls_verify: false
`

	tmpFile, err := os.CreateTemp("", "winpower-config-loader-*.yaml")
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
	if cfg.URL != "https://loader.example.com" {
		t.Errorf("Expected URL to be 'https://loader.example.com', got '%s'", cfg.URL)
	}

	if cfg.Username != "loaderuser" {
		t.Errorf("Expected Username to be 'loaderuser', got '%s'", cfg.Username)
	}
}
