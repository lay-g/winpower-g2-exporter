package cmd

import "time"

// CLIArgs represents command line arguments that can override configuration
type CLIArgs struct {
	// Basic flags
	Config     string
	LogLevel   string
	Port       int
	DataDir    string
	SkipSSLVerify bool
	SyncWrite  bool

	// WinPower module flags
	WinPowerURL           string
	WinPowerUsername      string
	WinPowerPassword      string
	WinPowerTimeout       time.Duration
	WinPowerMaxRetries    int
	WinPowerSkipTLSVerify bool

	// Server module flags
	ServerHost string
	ServerMode string

	// Scheduler module flags
	SchedulerInterval time.Duration

	// Energy module flags
	EnergyPrecision float64

	// Lifecycle flags
	ShutdownTimeout int // Shutdown timeout in seconds
	EnableForceShutdown bool // Enable force shutdown on SIGQUIT
}

// DefaultCLIArgs returns the default values for CLI arguments
func DefaultCLIArgs() *CLIArgs {
	return &CLIArgs{
		Config:          "./config.yaml",
		LogLevel:        "info",
		Port:            9090,
		DataDir:         "./data",
		SkipSSLVerify:   false,
		SyncWrite:       true,
		WinPowerTimeout: 30 * time.Second,
		WinPowerMaxRetries: 3,
		ServerHost:      "0.0.0.0",
		ServerMode:      "release",
		SchedulerInterval: 5 * time.Second,
		EnergyPrecision: 0.01,
		ShutdownTimeout: 30, // 30 seconds default
		EnableForceShutdown: true, // Enable force shutdown by default
	}
}

// HasOverrides checks if any CLI arguments are set to non-default values
func (c *CLIArgs) HasOverrides() bool {
	defaultArgs := DefaultCLIArgs()

	return c.Config != defaultArgs.Config ||
		c.LogLevel != defaultArgs.LogLevel ||
		c.Port != defaultArgs.Port ||
		c.DataDir != defaultArgs.DataDir ||
		c.SkipSSLVerify != defaultArgs.SkipSSLVerify ||
		c.SyncWrite != defaultArgs.SyncWrite ||
		c.WinPowerURL != "" ||
		c.WinPowerUsername != "" ||
		c.WinPowerPassword != "" ||
		c.WinPowerTimeout != defaultArgs.WinPowerTimeout ||
		c.WinPowerMaxRetries != defaultArgs.WinPowerMaxRetries ||
		c.WinPowerSkipTLSVerify != defaultArgs.WinPowerSkipTLSVerify ||
		c.ServerHost != defaultArgs.ServerHost ||
		c.ServerMode != defaultArgs.ServerMode ||
		c.SchedulerInterval != defaultArgs.SchedulerInterval ||
		c.EnergyPrecision != defaultArgs.EnergyPrecision ||
		c.ShutdownTimeout != defaultArgs.ShutdownTimeout ||
		c.EnableForceShutdown != defaultArgs.EnableForceShutdown
}