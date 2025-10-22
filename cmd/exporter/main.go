package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

const (
	// ApplicationVersion represents the current version of the application
	ApplicationVersion = "1.0.0"
	// ApplicationName represents the name of the application
	ApplicationName = "winpower-g2-exporter"
)

// AppConfig represents the complete application configuration
type AppConfig struct {
	Storage   *storage.Config   `yaml:"storage"`
	WinPower  *winpower.Config  `yaml:"winpower"`
	Energy    *energy.Config    `yaml:"energy"`
	Server    *server.Config    `yaml:"server"`
	Scheduler *scheduler.Config `yaml:"scheduler"`
}

// AppConfigOverrides represents CLI parameter overrides for configuration
type AppConfigOverrides struct {
	WinPowerURL           string
	WinPowerUsername      string
	WinPowerPassword      string
	WinPowerTimeout       time.Duration
	WinPowerMaxRetries    int
	WinPowerSkipTLSVerify bool
	StorageDataDir        string
	StorageSyncWrite      bool
	ServerPort            int
	ServerHost            string
	ServerMode            string
	SchedulerInterval     time.Duration
	EnergyPrecision       float64
}

// Application represents the main application structure
type Application struct {
	logger    log.Logger
	config    *AppConfig
	shutdown  chan os.Signal
	startTime time.Time
}

func main() {
	// Define command line flags
	var (
		configFile = flag.String("config", "", "Path to configuration file (YAML)")
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		version    = flag.Bool("version", false, "Show version information")
		help       = flag.Bool("help", false, "Show help information")

		// WinPower module flags
		winpowerURL           = flag.String("winpower-url", "", "WinPower server URL (e.g., https://winpower.example.com:8080)")
		winpowerUsername      = flag.String("winpower-username", "", "WinPower username")
		winpowerPassword      = flag.String("winpower-password", "", "WinPower password")
		winpowerTimeout       = flag.Duration("winpower-timeout", 0, "WinPower request timeout (e.g., 30s)")
		winpowerMaxRetries    = flag.Int("winpower-max-retries", 0, "WinPower maximum retry attempts")
		winpowerSkipTLSVerify = flag.Bool("winpower-skip-ssl-verify", false, "Skip TLS certificate verification")

		// Storage module flags
		storageDataDir   = flag.String("storage-data-dir", "", "Storage data directory path")
		storageSyncWrite = flag.Bool("storage-sync-write", false, "Enable synchronous writes")

		// Server module flags
		serverPort = flag.Int("port", 0, "Server port (default: 9090)")
		serverHost = flag.String("host", "", "Server host (default: 0.0.0.0)")
		serverMode = flag.String("server-mode", "", "Server mode (debug, release, test)")

		// Scheduler module flags
		schedulerInterval = flag.Duration("scheduler-interval", 0, "Scheduler collection interval (e.g., 5s)")

		// Energy module flags
		energyPrecision = flag.Float64("energy-precision", 0, "Energy calculation precision (e.g., 0.01)")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	// Initialize logger
	logConfig := log.DefaultConfig()
	logConfig.Level = log.Level(*logLevel)
	logger, err := log.NewLogger(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting WinPower G2 Exporter",
		zap.String("version", ApplicationVersion),
		zap.String("config_file", *configFile),
		zap.String("log_level", *logLevel),
	)

	// Load application configuration
	logger.Info("Loading application configuration",
		zap.String("config_file", *configFile),
		zap.String("config_source", determineConfigSource(*configFile)),
	)

	appConfig, err := loadApplicationConfig(*configFile)
	if err != nil {
		logger.Fatal("Failed to load configuration",
			zap.Error(err),
			zap.String("config_file", *configFile),
		)
	}

	logger.Info("Configuration loaded successfully from file and environment variables",
		zap.String("storage", appConfig.Storage.String()),
		zap.String("winpower", appConfig.WinPower.String()),
		zap.String("energy", appConfig.Energy.String()),
		zap.String("server", appConfig.Server.String()),
		zap.String("scheduler", appConfig.Scheduler.String()),
	)

	// Apply CLI parameter overrides
	cliOverrides := AppConfigOverrides{
		WinPowerURL:           *winpowerURL,
		WinPowerUsername:      *winpowerUsername,
		WinPowerPassword:      *winpowerPassword,
		WinPowerTimeout:       *winpowerTimeout,
		WinPowerMaxRetries:    *winpowerMaxRetries,
		WinPowerSkipTLSVerify: *winpowerSkipTLSVerify,
		StorageDataDir:        *storageDataDir,
		StorageSyncWrite:      *storageSyncWrite,
		ServerPort:            *serverPort,
		ServerHost:            *serverHost,
		ServerMode:            *serverMode,
		SchedulerInterval:     *schedulerInterval,
		EnergyPrecision:       *energyPrecision,
	}

	if hasCLIOverrides(cliOverrides) {
		logger.Info("Applying CLI parameter overrides")
		applyCLIOverrides(appConfig, cliOverrides)
		logger.Info("Configuration updated with CLI overrides",
			zap.String("storage", appConfig.Storage.String()),
			zap.String("winpower", appConfig.WinPower.String()),
			zap.String("energy", appConfig.Energy.String()),
			zap.String("server", appConfig.Server.String()),
			zap.String("scheduler", appConfig.Scheduler.String()),
		)
	}

	// Re-validate configuration after CLI overrides
	if err := validateAllConfigs(appConfig); err != nil {
		logger.Fatal("Configuration validation failed after CLI overrides",
			zap.Error(err),
			zap.Any("final_config", appConfig),
		)
	}

	logger.Info("Final configuration validation passed",
		zap.String("winpower_url", appConfig.WinPower.URL),
		zap.String("winpower_username", appConfig.WinPower.Username),
		zap.Int("server_port", appConfig.Server.Port),
		zap.String("server_host", appConfig.Server.Host),
		zap.String("storage_data_dir", appConfig.Storage.DataDir),
		zap.Duration("scheduler_interval", appConfig.Scheduler.CollectionInterval),
	)

	// Create and run application
	app := &Application{
		logger:    logger,
		config:    appConfig,
		shutdown:  make(chan os.Signal, 1),
		startTime: time.Now(),
	}

	// Setup signal handling
	signal.Notify(app.shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Run the application
	if err := app.Run(); err != nil {
		logger.Fatal("Application failed", zap.Error(err))
	}

	logger.Info("Application stopped gracefully",
		zap.Duration("uptime", time.Since(app.startTime)),
	)
}

// loadApplicationConfig loads the complete application configuration
func loadApplicationConfig(configFile string) (*AppConfig, error) {
	// Determine config file path
	if configFile == "" {
		// Try default locations
		defaultPaths := []string{
			"config.yaml",
			"config.yml",
			"/etc/winpower-exporter/config.yaml",
			filepath.Join(os.Getenv("HOME"), ".config", "winpower-exporter", "config.yaml"),
		}

		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				configFile = path
				break
			}
		}
	}

	// Create config loader
	loader := config.NewLoader("WINPOWER_EXPORTER")
	if configFile != "" {
		loader.SetConfigFile(configFile)
	}

	// Load configurations in dependency order with error logging
	var storageConfig *storage.Config
	var winpowerConfig *winpower.Config
	var energyConfig *energy.Config
	var schedulerConfig *scheduler.Config
	var serverConfig *server.Config
	var err error

	// Load storage config (first dependency)
	if storageConfig, err = loadStorageConfig(loader); err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	// Load WinPower config
	if winpowerConfig, err = loadWinPowerConfig(loader); err != nil {
		return nil, fmt.Errorf("failed to load winpower config: %w", err)
	}

	// Load energy config (depends on storage)
	if energyConfig, err = loadEnergyConfig(loader); err != nil {
		return nil, fmt.Errorf("failed to load energy config: %w", err)
	}

	// Load scheduler config
	if schedulerConfig, err = loadSchedulerConfig(loader); err != nil {
		return nil, fmt.Errorf("failed to load scheduler config: %w", err)
	}

	// Load server config (last dependency)
	if serverConfig, err = loadServerConfig(loader); err != nil {
		return nil, fmt.Errorf("failed to load server config: %w", err)
	}

	return &AppConfig{
		Storage:   storageConfig,
		WinPower:  winpowerConfig,
		Energy:    energyConfig,
		Server:    serverConfig,
		Scheduler: schedulerConfig,
	}, nil
}

// loadStorageConfig loads storage module configuration
func loadStorageConfig(loader *config.Loader) (*storage.Config, error) {
	return storage.NewConfig(loader)
}

// loadWinPowerConfig loads WinPower module configuration
func loadWinPowerConfig(loader *config.Loader) (*winpower.Config, error) {
	return winpower.NewConfig(loader)
}

// loadEnergyConfig loads energy module configuration
func loadEnergyConfig(loader *config.Loader) (*energy.Config, error) {
	return energy.NewConfig(loader)
}

// loadSchedulerConfig loads scheduler module configuration
func loadSchedulerConfig(loader *config.Loader) (*scheduler.Config, error) {
	return scheduler.NewConfig(loader)
}

// loadServerConfig loads server module configuration
func loadServerConfig(loader *config.Loader) (*server.Config, error) {
	return server.NewConfig(loader)
}

// applyCLIOverrides applies CLI parameter overrides to the loaded configuration
func applyCLIOverrides(config *AppConfig, overrides AppConfigOverrides) {
	// Apply WinPower overrides
	if overrides.WinPowerURL != "" {
		config.WinPower.URL = overrides.WinPowerURL
	}
	if overrides.WinPowerUsername != "" {
		config.WinPower.Username = overrides.WinPowerUsername
	}
	if overrides.WinPowerPassword != "" {
		config.WinPower.Password = overrides.WinPowerPassword
	}
	if overrides.WinPowerTimeout > 0 {
		config.WinPower.Timeout = overrides.WinPowerTimeout
	}
	if overrides.WinPowerMaxRetries > 0 {
		config.WinPower.MaxRetries = overrides.WinPowerMaxRetries
	}
	if overrides.WinPowerSkipTLSVerify {
		config.WinPower.SkipTLSVerify = overrides.WinPowerSkipTLSVerify
	}

	// Apply Storage overrides
	if overrides.StorageDataDir != "" {
		config.Storage.DataDir = overrides.StorageDataDir
	}
	// Note: boolean flags override regardless of value
	config.Storage.SyncWrite = overrides.StorageSyncWrite

	// Apply Server overrides
	if overrides.ServerPort > 0 {
		config.Server.Port = overrides.ServerPort
	}
	if overrides.ServerHost != "" {
		config.Server.Host = overrides.ServerHost
	}
	if overrides.ServerMode != "" {
		config.Server.Mode = overrides.ServerMode
	}

	// Apply Scheduler overrides
	if overrides.SchedulerInterval > 0 {
		config.Scheduler.CollectionInterval = overrides.SchedulerInterval
	}

	// Apply Energy overrides
	if overrides.EnergyPrecision > 0 {
		config.Energy.Precision = overrides.EnergyPrecision
	}
}

// validateAllConfigs validates all module configurations
func validateAllConfigs(config *AppConfig) error {
	if err := config.Storage.Validate(); err != nil {
		return fmt.Errorf("storage config validation failed: %w", err)
	}
	if err := config.WinPower.Validate(); err != nil {
		return fmt.Errorf("winpower config validation failed: %w", err)
	}
	if err := config.Energy.Validate(); err != nil {
		return fmt.Errorf("energy config validation failed: %w", err)
	}
	if err := config.Scheduler.Validate(); err != nil {
		return fmt.Errorf("scheduler config validation failed: %w", err)
	}
	if err := config.Server.Validate(); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	return nil
}

// determineConfigSource determines the source of configuration for logging
func determineConfigSource(configFile string) string {
	if configFile != "" {
		return "file"
	}
	return "defaults_and_environment"
}

// hasCLIOverrides checks if any CLI parameters are set (non-zero/non-empty)
func hasCLIOverrides(overrides AppConfigOverrides) bool {
	return overrides.WinPowerURL != "" ||
		overrides.WinPowerUsername != "" ||
		overrides.WinPowerPassword != "" ||
		overrides.WinPowerTimeout > 0 ||
		overrides.WinPowerMaxRetries > 0 ||
		overrides.WinPowerSkipTLSVerify ||
		overrides.StorageDataDir != "" ||
		overrides.ServerPort > 0 ||
		overrides.ServerHost != "" ||
		overrides.ServerMode != "" ||
		overrides.SchedulerInterval > 0 ||
		overrides.EnergyPrecision > 0
}

// Run starts the application and blocks until shutdown signal
func (app *Application) Run() error {
	app.logger.Info("Starting application components")

	// TODO: Initialize and start components in dependency order
	// 1. Storage manager
	// 2. WinPower client
	// 3. Energy calculator
	// 4. Scheduler
	// 5. HTTP server

	app.logger.Info("Application started successfully")

	// Wait for shutdown signal
	<-app.shutdown
	app.logger.Info("Shutdown signal received")

	// TODO: Graceful shutdown of components

	return nil
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("%s version %s\n", ApplicationName, ApplicationVersion)
}

// showHelp displays help information
func showHelp() {
	fmt.Printf(`WinPower G2 Exporter

USAGE:
  %s [OPTIONS]

BASIC OPTIONS:
  -config string            Path to configuration file (YAML)
  -log-level string         Log level (debug, info, warn, error) [default: info]
  -version                  Show version information
  -help                     Show this help message

WINPOWER MODULE OPTIONS:
  -winpower-url string      WinPower server URL (e.g., https://winpower.example.com:8080)
  -winpower-username string WinPower username
  -winpower-password string WinPower password
  -winpower-timeout duration WinPower request timeout (e.g., 30s)
  -winpower-max-retries int WinPower maximum retry attempts [default: 3]
  -winpower-skip-ssl-verify Skip TLS certificate verification [default: false]

STORAGE MODULE OPTIONS:
  -storage-data-dir string  Storage data directory path [default: ./data]
  -storage-sync-write       Enable synchronous writes [default: true]

SERVER MODULE OPTIONS:
  -port int                 Server port [default: 9090]
  -host string              Server host [default: 0.0.0.0]
  -server-mode string       Server mode (debug, release, test) [default: release]

SCHEDULER MODULE OPTIONS:
  -scheduler-interval duration Scheduler collection interval [default: 5s]

ENERGY MODULE OPTIONS:
  -energy-precision float   Energy calculation precision [default: 0.01]

CONFIGURATION:
  The exporter loads configuration from multiple sources in priority order:
  1. Command line flags (highest priority)
  2. Environment variables (WINPOWER_EXPORTER_ prefix)
  3. Configuration files (YAML)
  4. Default values (lowest priority)

  Environment variables:
    WINPOWER_EXPORTER_WINPOWER_URL="https://winpower.example.com:8080"
    WINPOWER_EXPORTER_STORAGE_DATA_DIR="/var/lib/winpower-exporter"
    WINPOWER_EXPORTER_SERVER_PORT="9090"
    WINPOWER_EXPORTER_SCHEDULER_COLLECTION_INTERVAL="5s"

  Example configuration file:
    storage:
      data_dir: "./data"
      sync_write: true

    winpower:
      url: "https://winpower.example.com:8080"
      username: "admin"
      password: "secret"
      timeout: "30s"
      max_retries: 3
      skip_tls_verify: false

    server:
      port: 9090
      host: "0.0.0.0"
      mode: "release"

    scheduler:
      collection_interval: "5s"

    energy:
      precision: 0.01

EXAMPLES:
  # Start with default configuration
  %s

  # Start with custom configuration file
  %s -config /etc/winpower-exporter/config.yaml

  # Start with CLI parameters
  %s -winpower-url https://winpower.example.com:8080 -winpower-username admin -port 8080

  # Start with environment variables and custom config file
  WINPOWER_EXPORTER_WINPOWER_PASSWORD="secret" %s -config config.yaml

  # Start with debug logging and pprof enabled
  %s -log-level debug -server-mode debug

  # Show version information
  %s -version

  # Show help
  %s -help

For more information, see the documentation at:
  https://github.com/lay-g/winpower-g2-exporter
`, ApplicationName, ApplicationName, ApplicationName, ApplicationName, ApplicationName, ApplicationName, ApplicationName, ApplicationName)
}
