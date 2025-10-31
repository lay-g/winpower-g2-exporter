package server

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 8080 {
		t.Errorf("DefaultConfig().Port = %d, want 8080", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("DefaultConfig().Host = %s, want 0.0.0.0", cfg.Host)
	}
	if cfg.Mode != "release" {
		t.Errorf("DefaultConfig().Mode = %s, want release", cfg.Mode)
	}
	if cfg.ReadTimeout != 10*time.Second {
		t.Errorf("DefaultConfig().ReadTimeout = %v, want 10s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 10*time.Second {
		t.Errorf("DefaultConfig().WriteTimeout = %v, want 10s", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != 60*time.Second {
		t.Errorf("DefaultConfig().IdleTimeout = %v, want 60s", cfg.IdleTimeout)
	}
	if cfg.EnablePprof != false {
		t.Errorf("DefaultConfig().EnablePprof = %v, want false", cfg.EnablePprof)
	}
	if cfg.ShutdownTimeout != 30*time.Second {
		t.Errorf("DefaultConfig().ShutdownTimeout = %v, want 30s", cfg.ShutdownTimeout)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Port:            0,
				Host:            "0.0.0.0",
				Mode:            "release",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Port:            65536,
				Host:            "0.0.0.0",
				Mode:            "release",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "empty host",
			config: &Config{
				Port:            8080,
				Host:            "",
				Mode:            "release",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			config: &Config{
				Port:            8080,
				Host:            "0.0.0.0",
				Mode:            "invalid",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid read timeout",
			config: &Config{
				Port:            8080,
				Host:            "0.0.0.0",
				Mode:            "release",
				ReadTimeout:     500 * time.Millisecond,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid write timeout",
			config: &Config{
				Port:            8080,
				Host:            "0.0.0.0",
				Mode:            "release",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    500 * time.Millisecond,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid idle timeout",
			config: &Config{
				Port:            8080,
				Host:            "0.0.0.0",
				Mode:            "release",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     500 * time.Millisecond,
				ShutdownTimeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid shutdown timeout",
			config: &Config{
				Port:            8080,
				Host:            "0.0.0.0",
				Mode:            "release",
				ReadTimeout:     10 * time.Second,
				WriteTimeout:    10 * time.Second,
				IdleTimeout:     60 * time.Second,
				ShutdownTimeout: 500 * time.Millisecond,
			},
			wantErr: true,
		},
		{
			name: "valid debug mode",
			config: &Config{
				Port:            8080,
				Host:            "127.0.0.1",
				Mode:            "debug",
				ReadTimeout:     5 * time.Second,
				WriteTimeout:    5 * time.Second,
				IdleTimeout:     30 * time.Second,
				EnablePprof:     true,
				ShutdownTimeout: 15 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "valid test mode",
			config: &Config{
				Port:            9090,
				Host:            "localhost",
				Mode:            "test",
				ReadTimeout:     1 * time.Second,
				WriteTimeout:    1 * time.Second,
				IdleTimeout:     1 * time.Second,
				ShutdownTimeout: 1 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != ErrInvalidConfig {
				t.Errorf("Config.Validate() error = %v, want %v", err, ErrInvalidConfig)
			}
		})
	}
}
