package config

import (
	"testing"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server:    server.DefaultConfig(),
				WinPower:  validWinPowerConfig(),
				Storage:   storage.DefaultConfig(),
				Scheduler: scheduler.DefaultConfig(),
				Logging:   log.DefaultConfig(),
			},
			wantErr: false,
		},
		{
			name: "invalid server config",
			config: &Config{
				Server: &server.Config{
					Port: -1, // Invalid port
					Host: "0.0.0.0",
				},
				WinPower:  validWinPowerConfig(),
				Storage:   storage.DefaultConfig(),
				Scheduler: scheduler.DefaultConfig(),
				Logging:   log.DefaultConfig(),
			},
			wantErr: true,
		},
		{
			name: "invalid winpower config",
			config: &Config{
				Server: server.DefaultConfig(),
				WinPower: &winpower.Config{
					BaseURL:  "", // Empty URL
					Username: "test",
					Password: "test",
				},
				Storage:   storage.DefaultConfig(),
				Scheduler: scheduler.DefaultConfig(),
				Logging:   log.DefaultConfig(),
			},
			wantErr: true,
		},
		{
			name: "invalid storage config",
			config: &Config{
				Server:   server.DefaultConfig(),
				WinPower: validWinPowerConfig(),
				Storage: &storage.Config{
					DataDir:         "", // Empty data dir
					FilePermissions: 0644,
				},
				Scheduler: scheduler.DefaultConfig(),
				Logging:   log.DefaultConfig(),
			},
			wantErr: true,
		},
		{
			name: "invalid scheduler config",
			config: &Config{
				Server:   server.DefaultConfig(),
				WinPower: validWinPowerConfig(),
				Storage:  storage.DefaultConfig(),
				Scheduler: &scheduler.Config{
					CollectionInterval: 0, // Invalid interval
				},
				Logging: log.DefaultConfig(),
			},
			wantErr: true,
		},
		{
			name: "invalid logging config",
			config: &Config{
				Server:    server.DefaultConfig(),
				WinPower:  validWinPowerConfig(),
				Storage:   storage.DefaultConfig(),
				Scheduler: scheduler.DefaultConfig(),
				Logging: &log.Config{
					Level:  "invalid", // Invalid log level
					Format: "json",
					Output: "stdout",
				},
			},
			wantErr: true,
		},
		{
			name: "nil module configs are allowed",
			config: &Config{
				Server:    nil,
				WinPower:  nil,
				Storage:   nil,
				Scheduler: nil,
				Logging:   nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigValidator_Interface(t *testing.T) {
	// Verify that all module configs implement ConfigValidator
	var _ ConfigValidator = (*server.Config)(nil)
	var _ ConfigValidator = (*winpower.Config)(nil)
	var _ ConfigValidator = (*storage.Config)(nil)
	var _ ConfigValidator = (*scheduler.Config)(nil)
	var _ ConfigValidator = (*log.Config)(nil)
}

func TestConfigManager_Interface(t *testing.T) {
	// Verify that Loader implements ConfigManager
	var _ ConfigManager = (*Loader)(nil)
}

// validWinPowerConfig returns a valid WinPower config for testing
func validWinPowerConfig() *winpower.Config {
	cfg := winpower.DefaultConfig()
	cfg.BaseURL = "https://winpower.example.com"
	cfg.Username = "test"
	cfg.Password = "test"
	return cfg
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		Server:    server.DefaultConfig(),
		WinPower:  validWinPowerConfig(),
		Storage:   storage.DefaultConfig(),
		Scheduler: scheduler.DefaultConfig(),
		Logging:   log.DefaultConfig(),
	}

	// Verify all fields are populated
	require.NotNil(t, cfg.Server)
	require.NotNil(t, cfg.WinPower)
	require.NotNil(t, cfg.Storage)
	require.NotNil(t, cfg.Scheduler)
	require.NotNil(t, cfg.Logging)

	// Verify all configs validate
	assert.NoError(t, cfg.Validate())
}
