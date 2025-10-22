package lifecycle

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// ModuleInitializer defines the interface for module initialization
type ModuleInitializer interface {
	// Initialize initializes the module with the given context and configuration
	Initialize(ctx context.Context, logger log.Logger) error

	// Start starts the module
	Start(ctx context.Context) error

	// Stop stops the module
	Stop(ctx context.Context) error

	// Name returns the module name
	Name() string

	// HealthCheck performs a health check on the module
	HealthCheck(ctx context.Context) error
}

// mockModuleInitializer implements ModuleInitializer for testing and demonstration
type mockModuleInitializer struct {
	name        string
	initError   error
	startError  error
	stopError   error
	healthError error
	initialized bool
	started     bool
}

func (m *mockModuleInitializer) Name() string {
	return m.name
}

func (m *mockModuleInitializer) Initialize(ctx context.Context, logger log.Logger) error {
	if m.initError != nil {
		return m.initError
	}
	m.initialized = true
	return nil
}

func (m *mockModuleInitializer) Start(ctx context.Context) error {
	if m.startError != nil {
		return m.startError
	}
	m.started = true
	return nil
}

func (m *mockModuleInitializer) Stop(ctx context.Context) error {
	if m.stopError != nil {
		return m.stopError
	}
	m.started = false
	return nil
}

func (m *mockModuleInitializer) HealthCheck(ctx context.Context) error {
	if m.healthError != nil {
		return m.healthError
	}
	return nil
}

// ModuleStatus represents the status of a module
type ModuleStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"` // "initializing", "running", "stopped", "error"
	StartTime time.Time `json:"start_time"`
	Error     string    `json:"error,omitempty"`
}

// Starter is responsible for starting all application modules in the correct order
type Starter struct {
	modules     map[string]ModuleInitializer
	status      map[string]*ModuleStatus
	logger      log.Logger
	startupTime time.Time
}

// NewStarter creates a new application starter
func NewStarter(logger log.Logger) *Starter {
	return &Starter{
		modules: make(map[string]ModuleInitializer),
		status:  make(map[string]*ModuleStatus),
		logger:  logger,
	}
}

// RegisterModule registers a module for initialization
func (s *Starter) RegisterModule(initializer ModuleInitializer) {
	s.modules[initializer.Name()] = initializer
	s.status[initializer.Name()] = &ModuleStatus{
		Name:   initializer.Name(),
		Status: "registered",
	}
}

// StartModules initializes and starts all modules in the correct dependency order
func (s *Starter) StartModules(ctx context.Context) error {
	s.startupTime = time.Now()
	s.logger.Info("Starting application modules",
		zap.Time("startup_time", s.startupTime),
		zap.Int("module_count", len(s.modules)),
	)

	// Define module initialization order based on dependencies
	initializationOrder := []string{
		"storage",     // Storage must be initialized first
		"logging",     // Logging is already initialized, but included for completeness
		"auth",        // Authentication module (placeholder for future implementation)
		"energy",      // Energy calculation depends on storage
		"winpower",    // WinPower client
		"collector",   // Data collector (placeholder for future implementation)
		"metrics",     // Metrics registry
		"server",      // HTTP server
		"scheduler",   // Scheduler starts last
	}

	// Initialize modules in dependency order
	for _, moduleName := range initializationOrder {
		if initializer, exists := s.modules[moduleName]; exists {
			if err := s.initializeModule(ctx, initializer); err != nil {
				return fmt.Errorf("failed to initialize module %s: %w", moduleName, err)
			}
		}
	}

	s.logger.Info("All modules initialized successfully",
		zap.Duration("total_startup_time", time.Since(s.startupTime)),
		zap.Strings("initialized_modules", s.getInitializedModules()),
	)

	return nil
}

// initializeModule initializes a single module
func (s *Starter) initializeModule(ctx context.Context, initializer ModuleInitializer) error {
	moduleName := initializer.Name()
	startTime := time.Now()

	s.logger.Info("Initializing module", zap.String("module", moduleName))

	// Update status to initializing
	s.updateModuleStatus(moduleName, "initializing", startTime, "")

	// Initialize the module
	if err := initializer.Initialize(ctx, s.logger); err != nil {
		s.updateModuleStatus(moduleName, "error", startTime, err.Error())
		return fmt.Errorf("module initialization failed: %w", err)
	}

	// Start the module
	if err := initializer.Start(ctx); err != nil {
		s.updateModuleStatus(moduleName, "error", startTime, err.Error())
		return fmt.Errorf("module start failed: %w", err)
	}

	// Update status to running
	s.updateModuleStatus(moduleName, "running", startTime, "")

	s.logger.Info("Module initialized successfully",
		zap.String("module", moduleName),
		zap.Duration("initialization_time", time.Since(startTime)),
	)

	return nil
}

// StopModules stops all modules in reverse initialization order
func (s *Starter) StopModules(ctx context.Context) error {
	s.logger.Info("Stopping application modules")

	// Define module stop order (reverse of initialization)
	stopOrder := []string{
		"scheduler",   // Stop scheduler first
		"server",      // Stop HTTP server
		"metrics",     // Stop metrics registry
		"collector",   // Stop data collector
		"winpower",    // Stop WinPower client
		"energy",      // Stop energy calculation
		"auth",        // Stop authentication
		"logging",     // Stop logging (should be last)
		"storage",     // Stop storage (should be last)
	}

	// Stop modules in reverse order
	for _, moduleName := range stopOrder {
		if initializer, exists := s.modules[moduleName]; exists {
			if err := s.stopModule(ctx, initializer); err != nil {
				s.logger.Error("Failed to stop module",
					zap.String("module", moduleName),
					zap.Error(err),
				)
			}
		}
	}

	s.logger.Info("All modules stopped",
		zap.Duration("total_uptime", time.Since(s.startupTime)),
	)

	return nil
}

// stopModule stops a single module
func (s *Starter) stopModule(ctx context.Context, initializer ModuleInitializer) error {
	moduleName := initializer.Name()
	startTime := time.Now()

	s.logger.Info("Stopping module", zap.String("module", moduleName))

	// Update status to stopping
	s.updateModuleStatus(moduleName, "stopping", time.Time{}, "")

	// Stop the module with timeout
	stopCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := initializer.Stop(stopCtx); err != nil {
		s.updateModuleStatus(moduleName, "error", time.Time{}, err.Error())
		return fmt.Errorf("module stop failed: %w", err)
	}

	// Update status to stopped
	s.updateModuleStatus(moduleName, "stopped", time.Time{}, "")

	s.logger.Info("Module stopped successfully",
		zap.String("module", moduleName),
		zap.Duration("stop_time", time.Since(startTime)),
	)

	return nil
}

// GetModuleStatus returns the status of all modules
func (s *Starter) GetModuleStatus() map[string]*ModuleStatus {
	return s.status
}

// GetStartupMetrics returns startup metrics
func (s *Starter) GetStartupMetrics() map[string]interface{} {
	return map[string]interface{}{
		"startup_time":      s.startupTime,
		"uptime":           time.Since(s.startupTime),
		"module_count":     len(s.modules),
		"initialized_modules": s.getInitializedModules(),
		"running_modules":  s.getRunningModules(),
	}
}

// PerformHealthCheck performs health checks on all running modules
func (s *Starter) PerformHealthCheck(ctx context.Context) error {
	s.logger.Debug("Performing health checks on all modules")

	for moduleName, status := range s.status {
		if status.Status != "running" {
			continue
		}

		if initializer, exists := s.modules[moduleName]; exists {
			if err := initializer.HealthCheck(ctx); err != nil {
				s.logger.Error("Module health check failed",
					zap.String("module", moduleName),
					zap.Error(err),
				)
				return fmt.Errorf("health check failed for module %s: %w", moduleName, err)
			}
		}
	}

	s.logger.Debug("All module health checks passed")
	return nil
}

// updateModuleStatus updates the status of a module
func (s *Starter) updateModuleStatus(name, status string, startTime time.Time, errorMsg string) {
	if moduleStatus, exists := s.status[name]; exists {
		moduleStatus.Status = status
		moduleStatus.Error = errorMsg
		if !startTime.IsZero() {
			moduleStatus.StartTime = startTime
		}
	}
}

// getInitializedModules returns a list of initialized module names
func (s *Starter) getInitializedModules() []string {
	var modules []string
	for _, status := range s.status {
		if status.Status == "running" || status.Status == "stopped" {
			modules = append(modules, status.Name)
		}
	}
	return modules
}

// getRunningModules returns a list of running module names
func (s *Starter) getRunningModules() []string {
	var modules []string
	for _, status := range s.status {
		if status.Status == "running" {
			modules = append(modules, status.Name)
		}
	}
	return modules
}

// SetupModules creates and registers all application modules
func (s *Starter) SetupModules(ctx context.Context) error {
	// For now, we create mock modules to demonstrate the lifecycle management
	// In a real implementation, these would be the actual module implementations
	s.logger.Info("Setting up application modules")

	// Create mock modules for demonstration
	mockModules := []string{"storage", "energy", "winpower", "metrics", "server", "scheduler"}

	for _, moduleName := range mockModules {
		mockModule := &mockModuleInitializer{name: moduleName}
		s.RegisterModule(mockModule)
		s.logger.Info("Registered mock module", zap.String("module", moduleName))
	}

	s.logger.Info("All modules setup completed")
	return nil
}