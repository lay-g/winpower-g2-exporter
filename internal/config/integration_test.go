package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_LoadValidConfig tests loading a valid config file
func TestIntegration_LoadValidConfig(t *testing.T) {
	// Create a loader and set config file explicitly
	loader := NewLoader()
	configPath := filepath.Join("fixtures", "valid_config.yaml")
	loader.viper.SetConfigFile(configPath)

	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify loaded values
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, "https://winpower.example.com", cfg.WinPower.BaseURL)
	assert.Equal(t, "testuser", cfg.WinPower.Username)
	assert.Equal(t, "./data", cfg.Storage.DataDir)
	assert.Equal(t, "info", cfg.Logging.Level)

	// Validate the config
	assert.NoError(t, cfg.Validate())
}

// TestIntegration_LoadPartialConfig tests loading a partial config with defaults
func TestIntegration_LoadPartialConfig(t *testing.T) {
	loader := NewLoader()
	configPath := filepath.Join("fixtures", "partial_config.yaml")
	loader.viper.SetConfigFile(configPath)

	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify partial values override defaults
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "debug", cfg.Server.Mode)

	// Verify defaults are used for unspecified values
	assert.NotZero(t, cfg.Server.ReadTimeout)
	assert.NotZero(t, cfg.Server.WriteTimeout)
	assert.NotNil(t, cfg.Storage)
	assert.NotNil(t, cfg.Scheduler)

	// Validate the config
	assert.NoError(t, cfg.Validate())
}

// TestIntegration_LoadInvalidConfig tests loading an invalid config
func TestIntegration_LoadInvalidConfig(t *testing.T) {
	loader := NewLoader()
	configPath := filepath.Join("fixtures", "invalid_config.yaml")
	loader.viper.SetConfigFile(configPath)

	cfg, err := loader.Load()
	require.NoError(t, err) // Loading should succeed
	require.NotNil(t, cfg)

	// Validation should fail
	err = cfg.Validate()
	assert.Error(t, err)
}

// TestIntegration_ConfigFileNotFound tests behavior when config file is not found
func TestIntegration_ConfigFileNotFound(t *testing.T) {
	loader := NewLoader()
	// Don't set a config file, let it search in default paths

	cfg, err := loader.Load()
	require.NoError(t, err) // Should use defaults when file not found
	require.NotNil(t, cfg)

	// Verify defaults are used
	assert.NotNil(t, cfg.Server)
	assert.NotNil(t, cfg.WinPower)
	assert.NotNil(t, cfg.Storage)
	assert.NotNil(t, cfg.Scheduler)
	assert.NotNil(t, cfg.Logging)
}

// TestIntegration_EnvironmentVariableOverride tests env vars override config file
func TestIntegration_EnvironmentVariableOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("WINPOWER_EXPORTER_SERVER_PORT", "7777")
	os.Setenv("WINPOWER_EXPORTER_LOGGING_LEVEL", "error")
	defer os.Unsetenv("WINPOWER_EXPORTER_SERVER_PORT")
	defer os.Unsetenv("WINPOWER_EXPORTER_LOGGING_LEVEL")

	loader := NewLoader()
	configPath := filepath.Join("fixtures", "valid_config.yaml")
	loader.viper.SetConfigFile(configPath)

	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variables override config file
	assert.Equal(t, 7777, cfg.Server.Port)
	assert.Equal(t, "error", cfg.Logging.Level)

	// Verify other values from config file are preserved
	assert.Equal(t, "https://winpower.example.com", cfg.WinPower.BaseURL)
}

// TestIntegration_MultiSourceConfig tests config loading from multiple sources
func TestIntegration_MultiSourceConfig(t *testing.T) {
	// Set environment variable
	os.Setenv("WINPOWER_EXPORTER_SERVER_PORT", "8888")
	defer os.Unsetenv("WINPOWER_EXPORTER_SERVER_PORT")

	loader := NewLoader()
	configPath := filepath.Join("fixtures", "partial_config.yaml")
	loader.viper.SetConfigFile(configPath)

	// Programmatically set a value (highest priority)
	loader.Set("logging.level", "fatal")

	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify priority: programmatic > env > file > defaults
	assert.Equal(t, 8888, cfg.Server.Port)        // From environment
	assert.Equal(t, "127.0.0.1", cfg.Server.Host) // From file
	assert.Equal(t, "fatal", cfg.Logging.Level)   // Programmatically set
}

// TestIntegration_ConfigValidation tests complete validation flow
func TestIntegration_ConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		envVars    map[string]string
		wantErr    bool
	}{
		{
			name:       "valid config file",
			configFile: "valid_config.yaml",
			wantErr:    false,
		},
		{
			name:       "partial config with defaults",
			configFile: "partial_config.yaml",
			wantErr:    false,
		},
		{
			name:       "invalid config file",
			configFile: "invalid_config.yaml",
			wantErr:    true,
		},
		{
			name:       "valid config with env override",
			configFile: "valid_config.yaml",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_SERVER_PORT": "9999",
			},
			wantErr: false,
		},
		{
			name:       "invalid config via env override",
			configFile: "valid_config.yaml",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_SERVER_PORT": "-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			loader := NewLoader()
			if tt.configFile != "" {
				configPath := filepath.Join("fixtures", tt.configFile)
				loader.viper.SetConfigFile(configPath)
			}

			cfg, err := loader.Load()
			require.NoError(t, err)
			require.NotNil(t, cfg)

			err = cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIntegration_ConfigFileSearch tests automatic config file search
func TestIntegration_ConfigFileSearch(t *testing.T) {
	// Save current directory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(cwd)

	// Create a temporary directory with a config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 5555
winpower:
  base_url: "https://test.com"
  username: "test"
  password: "test"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Change to temp directory
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create loader - should find config.yaml in current directory
	loader := NewLoader()
	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify it loaded the file
	assert.Equal(t, 5555, cfg.Server.Port)
	assert.Equal(t, "https://test.com", cfg.WinPower.BaseURL)
}

// TestIntegration_AllModulesValidate tests that all modules validate correctly
func TestIntegration_AllModulesValidate(t *testing.T) {
	loader := NewLoader()
	configPath := filepath.Join("fixtures", "valid_config.yaml")
	loader.viper.SetConfigFile(configPath)

	cfg, err := loader.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Test each module validates independently
	require.NotNil(t, cfg.Server)
	assert.NoError(t, cfg.Server.Validate())

	require.NotNil(t, cfg.WinPower)
	assert.NoError(t, cfg.WinPower.Validate())

	require.NotNil(t, cfg.Storage)
	assert.NoError(t, cfg.Storage.Validate())

	require.NotNil(t, cfg.Scheduler)
	assert.NoError(t, cfg.Scheduler.Validate())

	require.NotNil(t, cfg.Logging)
	assert.NoError(t, cfg.Logging.Validate())

	// Test complete config validates
	assert.NoError(t, cfg.Validate())
}

// TestIntegration_LoaderGetMethods tests all getter methods work after loading
func TestIntegration_LoaderGetMethods(t *testing.T) {
	loader := NewLoader()
	configPath := filepath.Join("fixtures", "valid_config.yaml")
	loader.viper.SetConfigFile(configPath)

	_, err := loader.Load()
	require.NoError(t, err)

	// Test Get methods
	assert.Equal(t, 9090, loader.GetInt("server.port"))
	assert.Equal(t, "0.0.0.0", loader.GetString("server.host"))
	assert.Equal(t, false, loader.GetBool("server.enable_pprof"))
	assert.True(t, loader.IsSet("server.port"))
	assert.False(t, loader.IsSet("nonexistent.key"))
}

// TestIntegration_ViperConfigFileNotFound tests handling of viper.ConfigFileNotFoundError
func TestIntegration_ViperConfigFileNotFound(t *testing.T) {
	loader := NewLoader()
	// Don't set a config file, let viper search in default paths (won't find any)

	cfg, err := loader.Load()
	// Should not return error when config file is not found during search
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should use defaults
	assert.NotNil(t, cfg.Server)
	assert.NotNil(t, cfg.Logging)
}

// TestIntegration_MalformedYAML tests handling of malformed YAML
func TestIntegration_MalformedYAML(t *testing.T) {
	// Create a temporary malformed YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "malformed.yaml")

	malformedYAML := `
server:
  port: 9090
  invalid: [unclosed array
`
	err := os.WriteFile(configPath, []byte(malformedYAML), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	loader.viper.SetConfigFile(configPath)

	_, err = loader.Load()
	// Should return error for malformed YAML
	assert.Error(t, err)
	var configErr *ConfigError
	assert.ErrorAs(t, err, &configErr)
}
