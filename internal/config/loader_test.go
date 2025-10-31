package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	require.NotNil(t, loader)
	require.NotNil(t, loader.viper)
	assert.NotEmpty(t, loader.searchPaths)
	assert.Contains(t, loader.searchPaths, ".")
	assert.Contains(t, loader.searchPaths, "./config")
}

func TestLoader_Load_WithDefaults(t *testing.T) {
	loader := NewLoader()
	cfg, err := loader.Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values are applied
	assert.NotNil(t, cfg.Server)
	assert.NotNil(t, cfg.WinPower)
	assert.NotNil(t, cfg.Storage)
	assert.NotNil(t, cfg.Scheduler)
	assert.NotNil(t, cfg.Logging)
}

func TestLoader_Load_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("WINPOWER_EXPORTER_SERVER_PORT", "9999")
	os.Setenv("WINPOWER_EXPORTER_LOGGING_LEVEL", "debug")
	defer os.Unsetenv("WINPOWER_EXPORTER_SERVER_PORT")
	defer os.Unsetenv("WINPOWER_EXPORTER_LOGGING_LEVEL")

	loader := NewLoader()
	cfg, err := loader.Load()

	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify environment variables override defaults
	assert.Equal(t, 9999, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
}

func TestLoader_Get(t *testing.T) {
	loader := NewLoader()
	loader.Set("test.key", "test_value")

	value := loader.Get("test.key")
	assert.Equal(t, "test_value", value)
}

func TestLoader_GetString(t *testing.T) {
	loader := NewLoader()
	loader.Set("test.string", "hello")

	value := loader.GetString("test.string")
	assert.Equal(t, "hello", value)
}

func TestLoader_GetInt(t *testing.T) {
	loader := NewLoader()
	loader.Set("test.int", 42)

	value := loader.GetInt("test.int")
	assert.Equal(t, 42, value)
}

func TestLoader_GetBool(t *testing.T) {
	loader := NewLoader()
	loader.Set("test.bool", true)

	value := loader.GetBool("test.bool")
	assert.True(t, value)
}

func TestLoader_GetStringSlice(t *testing.T) {
	loader := NewLoader()
	expected := []string{"a", "b", "c"}
	loader.Set("test.slice", expected)

	value := loader.GetStringSlice("test.slice")
	assert.Equal(t, expected, value)
}

func TestLoader_Set(t *testing.T) {
	loader := NewLoader()
	loader.Set("test.key", "value")

	assert.True(t, loader.IsSet("test.key"))
	assert.Equal(t, "value", loader.GetString("test.key"))
}

func TestLoader_IsSet(t *testing.T) {
	loader := NewLoader()

	// Test unset key
	assert.False(t, loader.IsSet("nonexistent.key"))

	// Test set key
	loader.Set("test.key", "value")
	assert.True(t, loader.IsSet("test.key"))
}

func TestLoader_Validate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Loader)
		wantErr bool
	}{
		{
			name: "valid config with defaults and required fields",
			setup: func(l *Loader) {
				// WinPower requires base_url, username, password
				l.Set("winpower.base_url", "https://test.com")
				l.Set("winpower.username", "user")
				l.Set("winpower.password", "pass")
			},
			wantErr: false,
		},
		{
			name: "valid config with custom values",
			setup: func(l *Loader) {
				l.Set("server.port", 9090)
				l.Set("winpower.base_url", "https://test.com")
				l.Set("winpower.username", "user")
				l.Set("winpower.password", "pass")
			},
			wantErr: false,
		},
		{
			name: "invalid server port",
			setup: func(l *Loader) {
				l.Set("server.port", -1)
				l.Set("winpower.base_url", "https://test.com")
				l.Set("winpower.username", "user")
				l.Set("winpower.password", "pass")
			},
			wantErr: true,
		},
		{
			name: "invalid winpower config",
			setup: func(l *Loader) {
				l.Set("winpower.base_url", "invalid-url")
				l.Set("winpower.username", "user")
				l.Set("winpower.password", "pass")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			tt.setup(loader)

			err := loader.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoader_SearchPaths(t *testing.T) {
	loader := NewLoader()

	// Verify search paths are set correctly
	assert.Greater(t, len(loader.searchPaths), 0)

	// Check that common paths are included
	foundWorkingDir := false
	foundConfigDir := false

	for _, path := range loader.searchPaths {
		if path == "." {
			foundWorkingDir = true
		}
		if path == "./config" {
			foundConfigDir = true
		}
	}

	assert.True(t, foundWorkingDir, "working directory should be in search paths")
	assert.True(t, foundConfigDir, "config directory should be in search paths")
}
