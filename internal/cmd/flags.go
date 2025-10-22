package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// FlagParser handles parsing and validation of command line flags
type FlagParser struct {
	cmd *cobra.Command
}

// NewFlagParser creates a new flag parser for the given command
func NewFlagParser(cmd *cobra.Command) *FlagParser {
	return &FlagParser{cmd: cmd}
}

// ParseFlags parses all flags and returns CLIArgs
func (fp *FlagParser) ParseFlags() (*CLIArgs, error) {
	args := DefaultCLIArgs()

	// Parse common flags
	if err := fp.parseCommonFlags(args); err != nil {
		return nil, fmt.Errorf("failed to parse common flags: %w", err)
	}

	// Parse server-specific flags
	if err := fp.parseServerFlags(args); err != nil {
		return nil, fmt.Errorf("failed to parse server flags: %w", err)
	}

	// Validate parsed arguments
	if err := fp.validateArgs(args); err != nil {
		return nil, fmt.Errorf("flag validation failed: %w", err)
	}

	return args, nil
}

// parseCommonFlags parses common flags that are shared across commands
func (fp *FlagParser) parseCommonFlags(args *CLIArgs) error {
	// Config file
	if config, err := fp.cmd.Flags().GetString("config"); err == nil {
		args.Config = config
	}

	// Log level
	if logLevel, err := fp.cmd.Flags().GetString("log-level"); err == nil {
		if err := validateLogLevel(logLevel); err != nil {
			return fmt.Errorf("invalid log-level: %w", err)
		}
		args.LogLevel = logLevel
	}

	// Port
	if port, err := fp.cmd.Flags().GetInt("port"); err == nil {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		args.Port = port
	}

	// Data directory
	if dataDir, err := fp.cmd.Flags().GetString("data-dir"); err == nil {
		args.DataDir = dataDir
	}

	// Skip SSL verify
	if skipSSLVerify, err := fp.cmd.Flags().GetBool("skip-ssl-verify"); err == nil {
		args.SkipSSLVerify = skipSSLVerify
	}

	// Sync write
	if syncWrite, err := fp.cmd.Flags().GetBool("sync-write"); err == nil {
		args.SyncWrite = syncWrite
	}

	return nil
}

// parseServerFlags parses server-specific flags
func (fp *FlagParser) parseServerFlags(args *CLIArgs) error {
	// WinPower module flags
	if winpowerURL, err := fp.cmd.Flags().GetString("winpower-url"); err == nil {
		args.WinPowerURL = winpowerURL
	}

	if winpowerUsername, err := fp.cmd.Flags().GetString("winpower-username"); err == nil {
		args.WinPowerUsername = winpowerUsername
	}

	if winpowerPassword, err := fp.cmd.Flags().GetString("winpower-password"); err == nil {
		args.WinPowerPassword = winpowerPassword
	}

	if winpowerTimeout, err := fp.cmd.Flags().GetDuration("winpower-timeout"); err == nil {
		if winpowerTimeout > 0 {
			args.WinPowerTimeout = winpowerTimeout
		}
	}

	if winpowerMaxRetries, err := fp.cmd.Flags().GetInt("winpower-max-retries"); err == nil {
		if winpowerMaxRetries > 0 {
			args.WinPowerMaxRetries = winpowerMaxRetries
		}
	}

	if winpowerSkipTLSVerify, err := fp.cmd.Flags().GetBool("winpower-skip-tls-verify"); err == nil {
		args.WinPowerSkipTLSVerify = winpowerSkipTLSVerify
	}

	// Server module flags
	if serverHost, err := fp.cmd.Flags().GetString("host"); err == nil {
		if serverHost != "" {
			args.ServerHost = serverHost
		}
	}

	if serverMode, err := fp.cmd.Flags().GetString("server-mode"); err == nil {
		if serverMode != "" {
			if err := validateServerMode(serverMode); err != nil {
				return fmt.Errorf("invalid server-mode: %w", err)
			}
			args.ServerMode = serverMode
		}
	}

	// Scheduler module flags
	if schedulerInterval, err := fp.cmd.Flags().GetDuration("scheduler-interval"); err == nil {
		if schedulerInterval > 0 {
			args.SchedulerInterval = schedulerInterval
		}
	}

	// Energy module flags
	if energyPrecision, err := fp.cmd.Flags().GetFloat64("energy-precision"); err == nil {
		if energyPrecision > 0 {
			args.EnergyPrecision = energyPrecision
		}
	}

	return nil
}

// validateArgs performs comprehensive validation of parsed arguments
func (fp *FlagParser) validateArgs(args *CLIArgs) error {
	// Validate log level
	if err := validateLogLevel(args.LogLevel); err != nil {
		return fmt.Errorf("log-level validation failed: %w", err)
	}

	// Validate server mode
	if err := validateServerMode(args.ServerMode); err != nil {
		return fmt.Errorf("server-mode validation failed: %w", err)
	}

	// Validate port range
	if args.Port < 1 || args.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", args.Port)
	}

	// Validate timeout values
	if args.WinPowerTimeout <= 0 {
		return fmt.Errorf("winpower-timeout must be positive, got: %v", args.WinPowerTimeout)
	}

	// Validate retry count
	if args.WinPowerMaxRetries < 0 {
		return fmt.Errorf("winpower-max-retries cannot be negative, got: %d", args.WinPowerMaxRetries)
	}

	// Validate scheduler interval
	if args.SchedulerInterval <= 0 {
		return fmt.Errorf("scheduler-interval must be positive, got: %v", args.SchedulerInterval)
	}

	// Validate energy precision
	if args.EnergyPrecision <= 0 || args.EnergyPrecision > 1 {
		return fmt.Errorf("energy-precision must be between 0 and 1, got: %f", args.EnergyPrecision)
	}

	return nil
}

// validateLogLevel validates the log level value
func validateLogLevel(level string) error {
	validLevels := []string{"debug", "info", "warn", "error"}
	for _, valid := range validLevels {
		if strings.EqualFold(level, valid) {
			return nil
		}
	}
	return fmt.Errorf("log level must be one of: %s", strings.Join(validLevels, ", "))
}

// validateServerMode validates the server mode value
func validateServerMode(mode string) error {
	validModes := []string{"debug", "release", "test"}
	for _, valid := range validModes {
		if strings.EqualFold(mode, valid) {
			return nil
		}
	}
	return fmt.Errorf("server mode must be one of: %s", strings.Join(validModes, ", "))
}

// GetFlagHelp returns help text for all supported flags
func GetFlagHelp() string {
	return `FLAGS:
  -c, --config string          Configuration file path (default "./config.yaml")
  -l, --log-level string       Log level (debug|info|warn|error) (default "info")
  -p, --port int              HTTP service port (default 9090)
      --data-dir string       Data directory path (default "./data")
      --skip-ssl-verify       Skip SSL certificate verification (default false)
      --sync-write            Enable synchronous writes (default true)

WINPOWER MODULE FLAGS:
      --winpower-url string      WinPower server URL
      --winpower-username string WinPower username
      --winpower-password string WinPower password
      --winpower-timeout duration WinPower request timeout (default 30s)
      --winpower-max-retries int WinPower maximum retry attempts (default 3)
      --winpower-skip-tls-verify Skip TLS verification for WinPower (default false)

SERVER MODULE FLAGS:
      --host string             Server host address (default "0.0.0.0")
      --server-mode string      Server mode (debug|release|test) (default "release")

SCHEDULER MODULE FLAGS:
      --scheduler-interval duration Scheduler collection interval (default 5s)

ENERGY MODULE FLAGS:
      --energy-precision float Energy calculation precision (default 0.01)
`
}