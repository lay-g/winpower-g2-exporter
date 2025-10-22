package cmd

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/cmd/lifecycle"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)



// ServerCommand implements the server subcommand
type ServerCommand struct {
	args          *CLIArgs
	lifecycleMgr  *lifecycle.LifecycleManager
	startTime     time.Time
}

// NewServerCommand creates a new server command instance
func NewServerCommand(args *CLIArgs) *ServerCommand {
	return &ServerCommand{
		args: args,
	}
}

// Name returns the command name
func (s *ServerCommand) Name() string {
	return "server"
}

// Description returns the command description
func (s *ServerCommand) Description() string {
	return "Start the HTTP server and begin data collection"
}

// Validate validates the server command arguments
func (s *ServerCommand) Validate(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("server command takes no arguments")
	}

	if s.args == nil {
		return fmt.Errorf("CLI arguments not initialized")
	}

	return nil
}

// Execute executes the server command
func (s *ServerCommand) Execute(ctx context.Context, args []string) error {
	if err := s.Validate(args); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	s.startTime = time.Now()

	// Initialize logger
	logger, err := s.initializeLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting WinPower G2 Exporter Server",
		zap.String("version", ApplicationVersion),
		zap.String("config_file", s.args.Config),
		zap.String("log_level", s.args.LogLevel),
	)

	// Load and validate configuration
	if err := s.loadConfiguration(logger); err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create enhanced shutdown and signal configurations
	shutdownConfig := lifecycle.DefaultShutdownConfig()
	signalConfig := lifecycle.DefaultSignalConfig()

	// Customize configurations based on command line arguments
	if s.args.ShutdownTimeout > 0 {
		shutdownConfig.DefaultTimeout = time.Duration(s.args.ShutdownTimeout) * time.Second
		shutdownConfig.PhaseTimeouts[lifecycle.PhaseMainShutdown] = time.Duration(s.args.ShutdownTimeout) * time.Second
	}

	// Configure force shutdown based on CLI arguments
	signalConfig.EnableForceShutdown = s.args.EnableForceShutdown

	// Create lifecycle manager with enhanced configurations
	s.lifecycleMgr = lifecycle.NewLifecycleManagerWithConfigs(logger, shutdownConfig, signalConfig)

	// Register shutdown hooks
	s.registerShutdownHooks(logger)

	// Start the application using lifecycle manager
	if err := s.lifecycleMgr.Start(ctx); err != nil {
		logger.Fatal("Failed to start application", zap.Error(err))
	}

	logger.Info("Application started successfully",
		zap.Duration("startup_time", time.Since(s.startTime)),
	)

	// Wait for shutdown signal
	if err := s.lifecycleMgr.WaitForShutdown(); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		return err
	}

	return nil
}

// initializeLogger initializes the application logger
func (s *ServerCommand) initializeLogger() (log.Logger, error) {
	logConfig := log.DefaultConfig()
	logConfig.Level = log.Level(s.args.LogLevel)
	return log.NewLogger(logConfig)
}

// loadConfiguration loads and validates the application configuration
func (s *ServerCommand) loadConfiguration(logger log.Logger) error {
	logger.Info("Loading application configuration",
		zap.String("config_file", s.args.Config),
		zap.String("config_source", s.determineConfigSource()),
	)

	if err := s.validateConfiguration(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info("Configuration loaded successfully")

	return nil
}

// validateConfiguration validates the basic configuration parameters
func (s *ServerCommand) validateConfiguration() error {
	// Basic validation of command line arguments
	if s.args.Port <= 0 || s.args.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", s.args.Port)
	}

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[s.args.LogLevel] {
		return fmt.Errorf("invalid log level: %s", s.args.LogLevel)
	}

	return nil
}

// determineConfigSource determines the source of configuration for logging
func (s *ServerCommand) determineConfigSource() string {
	if s.args.Config != "" && s.args.Config != "./config.yaml" {
		return "file"
	}
	return "defaults_and_environment"
}

// registerShutdownHooks registers shutdown hooks with the lifecycle manager
func (s *ServerCommand) registerShutdownHooks(logger log.Logger) {
	// Register cleanup hook for logger
	s.lifecycleMgr.RegisterShutdownHook(func() error {
		logger.Info("Syncing logger before shutdown")
		return logger.Sync()
	})

	// Register application metrics reporting hook
	s.lifecycleMgr.RegisterShutdownHook(func() error {
		if s.lifecycleMgr != nil {
			metrics := s.lifecycleMgr.GetMetrics()
			logger.Info("Application shutdown metrics",
				zap.Duration("total_uptime", metrics.Uptime),
				zap.Strings("running_modules", metrics.RunningModules),
				zap.Strings("stopped_modules", metrics.StoppedModules),
			)

			// Log enhanced shutdown results if available
			if s.lifecycleMgr.GetShutdownManager() != nil {
				shutdownResults := s.lifecycleMgr.GetShutdownResults()
				if len(shutdownResults) > 0 {
					logger.Info("Enhanced shutdown results",
						zap.Int("total_steps", len(shutdownResults)),
					)
					for _, result := range shutdownResults {
						if result.Success {
							logger.Debug("Shutdown step completed",
								zap.String("step", result.Step.Name),
								zap.Duration("duration", result.Duration),
							)
						} else {
							logger.Error("Shutdown step failed",
								zap.String("step", result.Step.Name),
								zap.Error(result.Error),
								zap.Bool("timeout_hit", result.TimeoutHit),
							)
						}
					}
				}
			}

			// Log signal statistics if available
			if s.lifecycleMgr.GetSignalManager() != nil {
				signalStats := s.lifecycleMgr.GetSignalStats()
				if len(signalStats) > 0 {
					logger.Info("Signal handling statistics",
						zap.Any("signals_received", signalStats),
					)
				}
			}
		}
		return nil
	})

	logger.Info("Enhanced shutdown hooks registered",
		zap.Bool("enhanced_shutdown_enabled", true),
		zap.Duration("shutdown_timeout", time.Duration(s.args.ShutdownTimeout)*time.Second),
		zap.Bool("force_shutdown_enabled", s.args.EnableForceShutdown),
	)
}