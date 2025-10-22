package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/cmd"
	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// AppConfig represents the complete application configuration
type AppConfig struct {
	Storage   *storage.Config   `yaml:"storage"`
	WinPower  *winpower.Config  `yaml:"winpower"`
	Energy    *energy.Config    `yaml:"energy"`
	Server    *server.Config    `yaml:"server"`
	Scheduler *scheduler.Config `yaml:"scheduler"`
}

// Application represents the main application structure
type Application struct {
	logger    log.Logger
	config    *AppConfig
	args      *cmd.CLIArgs
	shutdown  chan os.Signal
	startTime time.Time
}

func main() {
	ctx := context.Background()

	// Create root command
	rootCmd := cmd.NewRootCommand()

	// Execute command
	if err := rootCmd.Execute(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run starts the application with the given CLI arguments
func Run(ctx context.Context, cliArgs *cmd.CLIArgs) error {
	// Initialize logger
	logConfig := log.DefaultConfig()
	logConfig.Level = log.Level(cliArgs.LogLevel)
	logger, err := log.NewLogger(logConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting WinPower G2 Exporter",
		zap.String("version", cmd.ApplicationVersion),
		zap.String("config_file", cliArgs.Config),
		zap.String("log_level", cliArgs.LogLevel),
	)

	// Load application configuration
	logger.Info("Loading application configuration",
		zap.String("config_file", cliArgs.Config),
		zap.String("config_source", determineConfigSource(cliArgs.Config)),
	)

	appConfig, err := loadApplicationConfig(cliArgs.Config)
	if err != nil {
		logger.Fatal("Failed to load configuration",
			zap.Error(err),
			zap.String("config_file", cliArgs.Config),
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
	if cliArgs.HasOverrides() {
		logger.Info("Applying CLI parameter overrides")
		applyCLIOverrides(appConfig, cliArgs)
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
		args:      cliArgs,
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

	return nil
}

// loadApplicationConfig loads the complete application configuration
func loadApplicationConfig(configFile string) (*AppConfig, error) {
	// Determine config file path
	if configFile == "" || configFile == "./config.yaml" {
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
func applyCLIOverrides(config *AppConfig, args *cmd.CLIArgs) {
	// Apply WinPower overrides
	if args.WinPowerURL != "" {
		config.WinPower.URL = args.WinPowerURL
	}
	if args.WinPowerUsername != "" {
		config.WinPower.Username = args.WinPowerUsername
	}
	if args.WinPowerPassword != "" {
		config.WinPower.Password = args.WinPowerPassword
	}
	if args.WinPowerTimeout > 0 {
		config.WinPower.Timeout = args.WinPowerTimeout
	}
	if args.WinPowerMaxRetries > 0 {
		config.WinPower.MaxRetries = args.WinPowerMaxRetries
	}
	if args.WinPowerSkipTLSVerify {
		config.WinPower.SkipTLSVerify = args.WinPowerSkipTLSVerify
	}

	// Apply Storage overrides
	if args.DataDir != "" {
		config.Storage.DataDir = args.DataDir
	}
	// Note: boolean flags override regardless of value
	config.Storage.SyncWrite = args.SyncWrite

	// Apply Server overrides
	if args.Port > 0 {
		config.Server.Port = args.Port
	}
	if args.ServerHost != "" {
		config.Server.Host = args.ServerHost
	}
	if args.ServerMode != "" {
		config.Server.Mode = args.ServerMode
	}

	// Apply Scheduler overrides
	if args.SchedulerInterval > 0 {
		config.Scheduler.CollectionInterval = args.SchedulerInterval
	}

	// Apply Energy overrides
	if args.EnergyPrecision > 0 {
		config.Energy.Precision = args.EnergyPrecision
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