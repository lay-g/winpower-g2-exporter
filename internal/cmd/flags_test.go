package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Add common flags
	cmd.Flags().StringP("config", "c", "./config.yaml", "Configuration file path")
	cmd.Flags().StringP("log-level", "l", "info", "Log level (debug|info|warn|error)")
	cmd.Flags().IntP("port", "p", 9090, "HTTP service port")
	cmd.Flags().String("data-dir", "./data", "Data directory path")
	cmd.Flags().Bool("skip-ssl-verify", false, "Skip SSL certificate verification")
	cmd.Flags().Bool("sync-write", true, "Enable synchronous writes")

	// Add server flags
	cmd.Flags().String("winpower-url", "", "WinPower server URL")
	cmd.Flags().String("winpower-username", "", "WinPower username")
	cmd.Flags().String("winpower-password", "", "WinPower password")
	cmd.Flags().Duration("winpower-timeout", 0, "WinPower request timeout")
	cmd.Flags().Int("winpower-max-retries", 0, "WinPower maximum retry attempts")
	cmd.Flags().Bool("winpower-skip-tls-verify", false, "Skip TLS verification for WinPower")
	cmd.Flags().String("host", "", "Server host address")
	cmd.Flags().String("server-mode", "", "Server mode (debug|release|test)")
	cmd.Flags().Duration("scheduler-interval", 0, "Scheduler collection interval")
	cmd.Flags().Float64("energy-precision", 0, "Energy calculation precision")

	return cmd
}

func TestFlagParser_ParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr bool
		validate    func(*testing.T, *CLIArgs)
	}{
		{
			name:        "Default flags",
			args:        []string{},
			expectedErr: false,
			validate: func(t *testing.T, args *CLIArgs) {
				defaultArgs := DefaultCLIArgs()
				if args.Config != defaultArgs.Config {
					t.Errorf("Expected config %s, got %s", defaultArgs.Config, args.Config)
				}
				if args.LogLevel != defaultArgs.LogLevel {
					t.Errorf("Expected log level %s, got %s", defaultArgs.LogLevel, args.LogLevel)
				}
				if args.Port != defaultArgs.Port {
					t.Errorf("Expected port %d, got %d", defaultArgs.Port, args.Port)
				}
			},
		},
		{
			name:        "Valid config and port",
			args:        []string{"--config", "/test/config.yaml", "--port", "8080"},
			expectedErr: false,
			validate: func(t *testing.T, args *CLIArgs) {
				if args.Config != "/test/config.yaml" {
					t.Errorf("Expected config '/test/config.yaml', got '%s'", args.Config)
				}
				if args.Port != 8080 {
					t.Errorf("Expected port 8080, got %d", args.Port)
				}
			},
		},
		{
			name:        "Short flags",
			args:        []string{"-c", "/short/config.yaml", "-l", "debug", "-p", "9000"},
			expectedErr: false,
			validate: func(t *testing.T, args *CLIArgs) {
				if args.Config != "/short/config.yaml" {
					t.Errorf("Expected config '/short/config.yaml', got '%s'", args.Config)
				}
				if args.LogLevel != "debug" {
					t.Errorf("Expected log level 'debug', got '%s'", args.LogLevel)
				}
				if args.Port != 9000 {
					t.Errorf("Expected port 9000, got %d", args.Port)
				}
			},
		},
		{
			name:        "WinPower flags",
			args:        []string{"--winpower-url", "https://winpower.test.com", "--winpower-username", "admin", "--winpower-password", "secret"},
			expectedErr: false,
			validate: func(t *testing.T, args *CLIArgs) {
				if args.WinPowerURL != "https://winpower.test.com" {
					t.Errorf("Expected WinPower URL 'https://winpower.test.com', got '%s'", args.WinPowerURL)
				}
				if args.WinPowerUsername != "admin" {
					t.Errorf("Expected WinPower username 'admin', got '%s'", args.WinPowerUsername)
				}
				if args.WinPowerPassword != "secret" {
					t.Errorf("Expected WinPower password 'secret', got '%s'", args.WinPowerPassword)
				}
			},
		},
		{
			name:        "Duration flags",
			args:        []string{"--winpower-timeout", "60s", "--scheduler-interval", "10s"},
			expectedErr: false,
			validate: func(t *testing.T, args *CLIArgs) {
				expectedWinPowerTimeout := 60 * time.Second
				if args.WinPowerTimeout != expectedWinPowerTimeout {
					t.Errorf("Expected WinPower timeout %v, got %v", expectedWinPowerTimeout, args.WinPowerTimeout)
				}
				expectedSchedulerInterval := 10 * time.Second
				if args.SchedulerInterval != expectedSchedulerInterval {
					t.Errorf("Expected scheduler interval %v, got %v", expectedSchedulerInterval, args.SchedulerInterval)
				}
			},
		},
		{
			name:        "Boolean flags",
			args:        []string{"--skip-ssl-verify", "--sync-write=false", "--winpower-skip-tls-verify"},
			expectedErr: false,
			validate: func(t *testing.T, args *CLIArgs) {
				if !args.SkipSSLVerify {
					t.Errorf("Expected SkipSSLVerify to be true")
				}
				if args.SyncWrite {
					t.Errorf("Expected SyncWrite to be false")
				}
				if !args.WinPowerSkipTLSVerify {
					t.Errorf("Expected WinPowerSkipTLSVerify to be true")
				}
			},
		},
		{
			name:        "Invalid port",
			args:        []string{"--port", "99999"},
			expectedErr: true,
		},
		{
			name:        "Invalid log level",
			args:        []string{"--log-level", "invalid"},
			expectedErr: true,
		},
		{
			name:        "Invalid server mode",
			args:        []string{"--server-mode", "invalid"},
			expectedErr: true,
		},
		{
			name:        "Invalid port range (negative)",
			args:        []string{"--port", "-1"},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestCommand()
			cmd.SetArgs(tt.args)

			// Parse flags
			if err := cmd.ParseFlags(tt.args); err != nil {
				if !tt.expectedErr {
					t.Errorf("Unexpected error parsing flags: %v", err)
				}
				return
			}

			// Create flag parser and parse
			parser := NewFlagParser(cmd)
			args, err := parser.ParseFlags()

			if tt.expectedErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validate != nil {
				tt.validate(t, args)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectedErr bool
	}{
		{
			name:        "Valid debug level",
			level:       "debug",
			expectedErr: false,
		},
		{
			name:        "Valid info level",
			level:       "info",
			expectedErr: false,
		},
		{
			name:        "Valid warn level",
			level:       "warn",
			expectedErr: false,
		},
		{
			name:        "Valid error level",
			level:       "error",
			expectedErr: false,
		},
		{
			name:        "Valid uppercase DEBUG",
			level:       "DEBUG",
			expectedErr: false,
		},
		{
			name:        "Valid mixed case Info",
			level:       "Info",
			expectedErr: false,
		},
		{
			name:        "Invalid level",
			level:       "invalid",
			expectedErr: true,
		},
		{
			name:        "Empty level",
			level:       "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogLevel(tt.level)
			if tt.expectedErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestValidateServerMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		expectedErr bool
	}{
		{
			name:        "Valid debug mode",
			mode:        "debug",
			expectedErr: false,
		},
		{
			name:        "Valid release mode",
			mode:        "release",
			expectedErr: false,
		},
		{
			name:        "Valid test mode",
			mode:        "test",
			expectedErr: false,
		},
		{
			name:        "Valid uppercase DEBUG",
			mode:        "DEBUG",
			expectedErr: false,
		},
		{
			name:        "Valid mixed case Release",
			mode:        "Release",
			expectedErr: false,
		},
		{
			name:        "Invalid mode",
			mode:        "invalid",
			expectedErr: true,
		},
		{
			name:        "Empty mode",
			mode:        "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServerMode(tt.mode)
			if tt.expectedErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetFlagHelp(t *testing.T) {
	help := GetFlagHelp()

	// Check that help contains expected sections
	expectedSections := []string{
		"FLAGS:",
		"WINPOWER MODULE FLAGS:",
		"SERVER MODULE FLAGS:",
		"SCHEDULER MODULE FLAGS:",
		"ENERGY MODULE FLAGS:",
		"--config",
		"--log-level",
		"--port",
		"--data-dir",
		"--winpower-url",
		"--winpower-username",
		"--winpower-password",
		"--host",
		"--server-mode",
		"--scheduler-interval",
		"--energy-precision",
	}

	for _, section := range expectedSections {
		if !strings.Contains(help, section) {
			t.Errorf("Expected help to contain '%s', but it was not found", section)
		}
	}

	// Check that help contains default values
	expectedDefaults := []string{
		"./config.yaml",
		"info",
		"9090",
		"./data",
		"false",
		"true",
		"0.0.0.0",
		"release",
		"0.01",
	}

	for _, defaultValue := range expectedDefaults {
		if !strings.Contains(help, defaultValue) {
			t.Errorf("Expected help to contain default value '%s', but it was not found", defaultValue)
		}
	}
}