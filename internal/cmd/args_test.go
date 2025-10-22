package cmd

import (
	"testing"
	"time"
)

func TestDefaultCLIArgs(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T, *CLIArgs)
	}{
		{
			name: "Default config path",
			test: func(t *testing.T, args *CLIArgs) {
				if args.Config != "./config.yaml" {
					t.Errorf("Expected default config to be './config.yaml', got '%s'", args.Config)
				}
			},
		},
		{
			name: "Default log level",
			test: func(t *testing.T, args *CLIArgs) {
				if args.LogLevel != "info" {
					t.Errorf("Expected default log level to be 'info', got '%s'", args.LogLevel)
				}
			},
		},
		{
			name: "Default port",
			test: func(t *testing.T, args *CLIArgs) {
				if args.Port != 9090 {
					t.Errorf("Expected default port to be 9090, got %d", args.Port)
				}
			},
		},
		{
			name: "Default data directory",
			test: func(t *testing.T, args *CLIArgs) {
				if args.DataDir != "./data" {
					t.Errorf("Expected default data directory to be './data', got '%s'", args.DataDir)
				}
			},
		},
		{
			name: "Default SSL verification",
			test: func(t *testing.T, args *CLIArgs) {
				if args.SkipSSLVerify != false {
					t.Errorf("Expected default SkipSSLVerify to be false, got %v", args.SkipSSLVerify)
				}
			},
		},
		{
			name: "Default sync write",
			test: func(t *testing.T, args *CLIArgs) {
				if args.SyncWrite != true {
					t.Errorf("Expected default SyncWrite to be true, got %v", args.SyncWrite)
				}
			},
		},
		{
			name: "Default WinPower timeout",
			test: func(t *testing.T, args *CLIArgs) {
				expectedTimeout := 30 * time.Second
				if args.WinPowerTimeout != expectedTimeout {
					t.Errorf("Expected default WinPowerTimeout to be %v, got %v", expectedTimeout, args.WinPowerTimeout)
				}
			},
		},
		{
			name: "Default server host",
			test: func(t *testing.T, args *CLIArgs) {
				if args.ServerHost != "0.0.0.0" {
					t.Errorf("Expected default ServerHost to be '0.0.0.0', got '%s'", args.ServerHost)
				}
			},
		},
		{
			name: "Default server mode",
			test: func(t *testing.T, args *CLIArgs) {
				if args.ServerMode != "release" {
					t.Errorf("Expected default ServerMode to be 'release', got '%s'", args.ServerMode)
				}
			},
		},
		{
			name: "Default scheduler interval",
			test: func(t *testing.T, args *CLIArgs) {
				expectedInterval := 5 * time.Second
				if args.SchedulerInterval != expectedInterval {
					t.Errorf("Expected default SchedulerInterval to be %v, got %v", expectedInterval, args.SchedulerInterval)
				}
			},
		},
		{
			name: "Default energy precision",
			test: func(t *testing.T, args *CLIArgs) {
				if args.EnergyPrecision != 0.01 {
					t.Errorf("Expected default EnergyPrecision to be 0.01, got %f", args.EnergyPrecision)
				}
			},
		},
	}

	args := DefaultCLIArgs()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, args)
		})
	}
}

func TestCLIArgs_HasOverrides(t *testing.T) {
	tests := []struct {
		name     string
		args     *CLIArgs
		expected bool
	}{
		{
			name:     "Default args has no overrides",
			args:     DefaultCLIArgs(),
			expected: false,
		},
		{
			name: "Config override",
			args: &CLIArgs{
				Config:          "/custom/config.yaml",
				LogLevel:        "info",
				Port:            9090,
				DataDir:         "./data",
				SkipSSLVerify:   false,
				SyncWrite:       true,
				WinPowerTimeout: 30 * time.Second,
				ServerHost:      "0.0.0.0",
				ServerMode:      "release",
				SchedulerInterval: 5 * time.Second,
				EnergyPrecision: 0.01,
			},
			expected: true,
		},
		{
			name: "Log level override",
			args: &CLIArgs{
				Config:          "./config.yaml",
				LogLevel:        "debug",
				Port:            9090,
				DataDir:         "./data",
				SkipSSLVerify:   false,
				SyncWrite:       true,
				WinPowerTimeout: 30 * time.Second,
				ServerHost:      "0.0.0.0",
				ServerMode:      "release",
				SchedulerInterval: 5 * time.Second,
				EnergyPrecision: 0.01,
			},
			expected: true,
		},
		{
			name: "Port override",
			args: &CLIArgs{
				Config:          "./config.yaml",
				LogLevel:        "info",
				Port:            8080,
				DataDir:         "./data",
				SkipSSLVerify:   false,
				SyncWrite:       true,
				WinPowerTimeout: 30 * time.Second,
				ServerHost:      "0.0.0.0",
				ServerMode:      "release",
				SchedulerInterval: 5 * time.Second,
				EnergyPrecision: 0.01,
			},
			expected: true,
		},
		{
			name: "Multiple overrides",
			args: &CLIArgs{
				Config:          "./config.yaml",
				LogLevel:        "debug",
				Port:            8080,
				DataDir:         "/custom/data",
				SkipSSLVerify:   true,
				SyncWrite:       false,
				WinPowerURL:     "https://example.com",
				WinPowerTimeout: 60 * time.Second,
				ServerHost:      "127.0.0.1",
				ServerMode:      "debug",
				SchedulerInterval: 10 * time.Second,
				EnergyPrecision: 0.001,
			},
			expected: true,
		},
		{
			name: "WinPower URL override only",
			args: &CLIArgs{
				Config:          "./config.yaml",
				LogLevel:        "info",
				Port:            9090,
				DataDir:         "./data",
				SkipSSLVerify:   false,
				SyncWrite:       true,
				WinPowerURL:     "https://winpower.example.com",
				WinPowerTimeout: 30 * time.Second,
				ServerHost:      "0.0.0.0",
				ServerMode:      "release",
				SchedulerInterval: 5 * time.Second,
				EnergyPrecision: 0.01,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.args.HasOverrides()
			if result != tt.expected {
				t.Errorf("CLIArgs.HasOverrides() = %v, expected %v", result, tt.expected)
			}
		})
	}
}