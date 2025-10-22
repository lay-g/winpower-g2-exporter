package lifecycle

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// LifecycleState represents the current state of the application lifecycle
type LifecycleState string

const (
	StateStarting LifecycleState = "starting"
	StateRunning  LifecycleState = "running"
	StateStopping LifecycleState = "stopping"
	StateStopped  LifecycleState = "stopped"
	StateError    LifecycleState = "error"
)

// ShutdownHook represents a shutdown hook function
type ShutdownHook func() error

// LifecycleManager manages the application lifecycle including startup, shutdown, and health monitoring
type LifecycleManager struct {
	// Dependencies
	logger          log.Logger
	starter         *Starter
	healthMgr       *HealthManager
	shutdownMgr     *ShutdownManager
	signalMgr       *SignalManager

	// State management
	state      LifecycleState
	stateMutex sync.RWMutex
	startTime  time.Time

	// Shutdown handling (legacy, kept for compatibility)
	shutdownHooks   []ShutdownHook
	shutdownSignal  chan os.Signal
	shutdownTimeout time.Duration

	// Error handling
	lastError     error
	errorOccurred chan error

	// Wait group for graceful shutdown
	wg sync.WaitGroup

	// Context management
	ctx    context.Context
	cancel context.CancelFunc

	// Enhanced shutdown features
	useEnhancedShutdown bool
}

// LifecycleMetrics contains metrics about the application lifecycle
type LifecycleMetrics struct {
	StartTime         time.Time     `json:"start_time"`
	Uptime            time.Duration `json:"uptime"`
	State             string        `json:"state"`
	ModuleCount       int           `json:"module_count"`
	RunningModules    []string      `json:"running_modules"`
	StoppedModules    []string      `json:"stopped_modules"`
	ErrorModules      []string      `json:"error_modules"`
	LastError         string        `json:"last_error,omitempty"`
	ShutdownTimeout   time.Duration `json:"shutdown_timeout"`
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(logger log.Logger) *LifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	// Create shutdown and signal managers with default configs
	shutdownConfig := DefaultShutdownConfig()
	signalConfig := DefaultSignalConfig()

	return &LifecycleManager{
		logger:              logger,
		starter:             NewStarter(logger),
		healthMgr:           NewHealthManager(logger),
		shutdownMgr:         NewShutdownManager(logger, shutdownConfig),
		signalMgr:           NewSignalManager(logger, signalConfig),
		state:               StateStopped,
		shutdownHooks:       make([]ShutdownHook, 0),
		shutdownSignal:      make(chan os.Signal, 1),
		shutdownTimeout:     30 * time.Second,
		errorOccurred:       make(chan error, 1),
		ctx:                 ctx,
		cancel:              cancel,
		useEnhancedShutdown: true, // Enable enhanced shutdown by default
	}
}

// NewLifecycleManagerWithConfigs creates a new lifecycle manager with custom configurations
func NewLifecycleManagerWithConfigs(logger log.Logger, shutdownConfig *ShutdownConfig, signalConfig *SignalConfig) *LifecycleManager {
	ctx, cancel := context.WithCancel(context.Background())

	if shutdownConfig == nil {
		shutdownConfig = DefaultShutdownConfig()
	}
	if signalConfig == nil {
		signalConfig = DefaultSignalConfig()
	}

	return &LifecycleManager{
		logger:              logger,
		starter:             NewStarter(logger),
		healthMgr:           NewHealthManager(logger),
		shutdownMgr:         NewShutdownManager(logger, shutdownConfig),
		signalMgr:           NewSignalManager(logger, signalConfig),
		state:               StateStopped,
		shutdownHooks:       make([]ShutdownHook, 0),
		shutdownSignal:      make(chan os.Signal, 1),
		shutdownTimeout:     shutdownConfig.DefaultTimeout,
		errorOccurred:       make(chan error, 1),
		ctx:                 ctx,
		cancel:              cancel,
		useEnhancedShutdown: true,
	}
}

// Start starts the application and all its modules
func (lm *LifecycleManager) Start(ctx context.Context) error {
	lm.setState(StateStarting)
	lm.startTime = time.Now()

	lm.logger.Info("Starting application lifecycle manager",
		zap.Time("start_time", lm.startTime),
		zap.String("version", "unknown"),
		zap.Bool("enhanced_shutdown", lm.useEnhancedShutdown),
	)

	// Setup enhanced signal handling if enabled
	if lm.useEnhancedShutdown {
		if err := lm.setupEnhancedSignalHandling(); err != nil {
			lm.setError(fmt.Errorf("failed to setup enhanced signal handling: %w", err))
			return lm.lastError
		}
	} else {
		// Fallback to legacy signal handling
		lm.setupSignalHandling()
	}

	// Setup shutdown steps for enhanced shutdown
	if lm.useEnhancedShutdown {
		lm.setupShutdownSteps()
	}

	// Setup modules
	if err := lm.starter.SetupModules(ctx); err != nil {
		lm.setError(fmt.Errorf("failed to setup modules: %w", err))
		return lm.lastError
	}

	// Start all modules
	if err := lm.starter.StartModules(ctx); err != nil {
		lm.setError(fmt.Errorf("failed to start modules: %w", err))
		return lm.lastError
	}

	// Start health monitoring
	lm.wg.Add(1)
	go lm.healthMonitoringLoop()

	lm.setState(StateRunning)
	lm.logger.Info("Application started successfully",
		zap.Duration("startup_time", time.Since(lm.startTime)),
		zap.Strings("running_modules", lm.starter.getRunningModules()),
	)

	return nil
}

// Stop stops the application and all its modules
func (lm *LifecycleManager) Stop(ctx context.Context) error {
	lm.setState(StateStopping)

	lm.logger.Info("Stopping application lifecycle manager")

	// Cancel the context to signal all goroutines to stop
	lm.cancel()

	// Use enhanced shutdown if enabled
	if lm.useEnhancedShutdown {
		return lm.enhancedShutdown(ctx)
	}

	// Fallback to legacy shutdown
	return lm.legacyShutdown(ctx)
}

// WaitForShutdown waits for shutdown signals
func (lm *LifecycleManager) WaitForShutdown() error {
	lm.logger.Info("Waiting for shutdown signal")

	if lm.useEnhancedShutdown {
		// Use enhanced signal manager
		select {
		case err := <-lm.errorOccurred:
			lm.logger.Error("Fatal error occurred, initiating shutdown", zap.Error(err))
			return lm.Stop(lm.ctx)
		case <-lm.ctx.Done():
			lm.logger.Info("Context cancelled, initiating shutdown")
			return lm.Stop(context.Background())
		}
	} else {
		// Use legacy signal handling
		select {
		case sig := <-lm.shutdownSignal:
			lm.logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
			return lm.Stop(lm.ctx)
		case err := <-lm.errorOccurred:
			lm.logger.Error("Fatal error occurred, initiating shutdown", zap.Error(err))
			return lm.Stop(lm.ctx)
		case <-lm.ctx.Done():
			lm.logger.Info("Context cancelled, initiating shutdown")
			return lm.Stop(context.Background())
		}
	}
}

// RegisterShutdownHook registers a function to be called during shutdown
func (lm *LifecycleManager) RegisterShutdownHook(hook ShutdownHook) {
	lm.shutdownHooks = append(lm.shutdownHooks, hook)
	lm.logger.Debug("Shutdown hook registered", zap.Int("total_hooks", len(lm.shutdownHooks)))
}

// GetState returns the current lifecycle state
func (lm *LifecycleManager) GetState() LifecycleState {
	lm.stateMutex.RLock()
	defer lm.stateMutex.RUnlock()
	return lm.state
}

// GetMetrics returns current lifecycle metrics
func (lm *LifecycleManager) GetMetrics() *LifecycleMetrics {
	lm.stateMutex.RLock()
	defer lm.stateMutex.RUnlock()

	moduleStatuses := lm.starter.GetModuleStatus()
	var runningModules, stoppedModules, errorModules []string

	for name, status := range moduleStatuses {
		switch status.Status {
		case "running":
			runningModules = append(runningModules, name)
		case "stopped":
			stoppedModules = append(stoppedModules, name)
		case "error":
			errorModules = append(errorModules, name)
		}
	}

	lastError := ""
	if lm.lastError != nil {
		lastError = lm.lastError.Error()
	}

	return &LifecycleMetrics{
		StartTime:       lm.startTime,
		Uptime:          time.Since(lm.startTime),
		State:           string(lm.state),
		ModuleCount:     len(moduleStatuses),
		RunningModules:  runningModules,
		StoppedModules:  stoppedModules,
		ErrorModules:    errorModules,
		LastError:       lastError,
		ShutdownTimeout: lm.shutdownTimeout,
	}
}

// PerformHealthCheck performs a comprehensive health check
func (lm *LifecycleManager) PerformHealthCheck(ctx context.Context) error {
	// Check if we're in a valid state for health checks
	currentState := lm.GetState()
	if currentState != StateRunning {
		return fmt.Errorf("health check not available in state: %s", currentState)
	}

	// Perform module health checks
	if err := lm.starter.PerformHealthCheck(ctx); err != nil {
		return fmt.Errorf("module health check failed: %w", err)
	}

	// Perform system-level health checks
	if err := lm.healthMgr.PerformSystemHealthCheck(ctx); err != nil {
		return fmt.Errorf("system health check failed: %w", err)
	}

	return nil
}

// SetShutdownTimeout sets the shutdown timeout
func (lm *LifecycleManager) SetShutdownTimeout(timeout time.Duration) {
	lm.shutdownTimeout = timeout
	lm.logger.Info("Shutdown timeout updated", zap.Duration("timeout", timeout))
}

// setState updates the lifecycle state
func (lm *LifecycleManager) setState(state LifecycleState) {
	lm.stateMutex.Lock()
	defer lm.stateMutex.Unlock()

	oldState := lm.state
	lm.state = state

	lm.logger.Info("Lifecycle state changed",
		zap.String("old_state", string(oldState)),
		zap.String("new_state", string(state)),
	)
}

// setError sets the last error and notifies error channel
func (lm *LifecycleManager) setError(err error) {
	lm.lastError = err
	lm.setState(StateError)

	select {
	case lm.errorOccurred <- err:
	default:
		// Channel is full, error will be lost but that's acceptable
		lm.logger.Warn("Error channel full, dropping error", zap.Error(err))
	}
}

// setupSignalHandling configures signal handling for graceful shutdown
func (lm *LifecycleManager) setupSignalHandling() {
	signal.Notify(lm.shutdownSignal,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGQUIT, // Quit signal
	)
}

// executeShutdownHooks executes all registered shutdown hooks
func (lm *LifecycleManager) executeShutdownHooks(ctx context.Context) error {
	lm.logger.Info("Executing shutdown hooks", zap.Int("hook_count", len(lm.shutdownHooks)))

	for i, hook := range lm.shutdownHooks {
		hookStart := time.Now()

		if err := hook(); err != nil {
			lm.logger.Error("Shutdown hook failed",
				zap.Int("hook_index", i),
				zap.Error(err),
				zap.Duration("execution_time", time.Since(hookStart)),
			)
		} else {
			lm.logger.Debug("Shutdown hook executed successfully",
				zap.Int("hook_index", i),
				zap.Duration("execution_time", time.Since(hookStart)),
			)
		}
	}

	return nil
}

// healthMonitoringLoop runs the health monitoring loop
func (lm *LifecycleManager) healthMonitoringLoop() {
	defer lm.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // Health check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(lm.ctx, 10*time.Second)
			if err := lm.PerformHealthCheck(ctx); err != nil {
				lm.logger.Error("Health check failed", zap.Error(err))
				// Don't set error state for health check failures, just log them
			}
			cancel()

		case <-lm.ctx.Done():
			lm.logger.Info("Health monitoring loop stopping")
			return
		}
	}
}

// GetStarter returns the application starter for direct access
func (lm *LifecycleManager) GetStarter() *Starter {
	return lm.starter
}

// GetHealthManager returns the health manager for direct access
func (lm *LifecycleManager) GetHealthManager() *HealthManager {
	return lm.healthMgr
}

// Restart restarts the application by stopping and starting all modules
func (lm *LifecycleManager) Restart(ctx context.Context) error {
	lm.logger.Info("Restarting application")

	// Stop the application
	if err := lm.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop application during restart: %w", err)
	}

	// Wait a brief moment before starting
	time.Sleep(1 * time.Second)

	// Start the application
	if err := lm.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application during restart: %w", err)
	}

	lm.logger.Info("Application restarted successfully")
	return nil
}

// ForceShutdown forces an immediate shutdown without waiting for graceful shutdown
func (lm *LifecycleManager) ForceShutdown() {
	lm.logger.Warn("Force shutdown initiated")

	// Cancel context immediately
	lm.cancel()

	// Force state to stopped
	lm.setState(StateStopped)

	lm.logger.Info("Force shutdown completed")
}

// setupEnhancedSignalHandling sets up enhanced signal handling using the signal manager
func (lm *LifecycleManager) setupEnhancedSignalHandling() error {
	// Register signal handlers
	handlers := map[os.Signal]SignalHandler{
		syscall.SIGTERM: lm.handleGracefulShutdown,
		syscall.SIGINT:  lm.handleGracefulShutdown,
		syscall.SIGQUIT: lm.handleForceShutdown,
	}

	lm.signalMgr.RegisterMultipleHandlers(handlers)

	// Start listening for signals
	if err := lm.signalMgr.StartListening(); err != nil {
		return fmt.Errorf("failed to start signal manager: %w", err)
	}

	lm.logger.Info("Enhanced signal handling setup completed")
	return nil
}

// setupShutdownSteps sets up the shutdown steps for enhanced shutdown
func (lm *LifecycleManager) setupShutdownSteps() {
	// Pre-shutdown phase: stop accepting new requests
	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-accepting-requests",
		Phase:    PhasePreShutdown,
		Timeout:  5 * time.Second,
		Required: true,
		Priority: 1,
		ExecuteFunc: func(ctx context.Context) error {
			lm.logger.Info("Stopping to accept new requests")
			// TODO: Implement server stop accepting requests
			return nil
		},
	})

	// Main shutdown phase: stop modules in reverse order
	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-scheduler",
		Phase:    PhaseMainShutdown,
		Timeout:  10 * time.Second,
		Required: true,
		Priority: 1,
		ExecuteFunc: func(ctx context.Context) error {
			// TODO: Implement individual module stopping
			// For now, stop all modules
			return lm.starter.StopModules(ctx)
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-server",
		Phase:    PhaseMainShutdown,
		Timeout:  10 * time.Second,
		Required: true,
		Priority: 2,
		ExecuteFunc: func(ctx context.Context) error {
			// For now, this is a no-op since we stop all modules at once
			// In a more refined implementation, we would stop modules individually
			return nil
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-metrics",
		Phase:    PhaseMainShutdown,
		Timeout:  5 * time.Second,
		Required: true,
		Priority: 3,
		ExecuteFunc: func(ctx context.Context) error {
			// For now, this is a no-op since we stop all modules at once
			// In a more refined implementation, we would stop modules individually
			return nil
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-collector",
		Phase:    PhaseMainShutdown,
		Timeout:  10 * time.Second,
		Required: true,
		Priority: 4,
		ExecuteFunc: func(ctx context.Context) error {
			// For now, this is a no-op since we stop all modules at once
			// In a more refined implementation, we would stop modules individually
			return nil
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-energy",
		Phase:    PhaseMainShutdown,
		Timeout:  5 * time.Second,
		Required: true,
		Priority: 5,
		ExecuteFunc: func(ctx context.Context) error {
			// For now, this is a no-op since we stop all modules at once
			// In a more refined implementation, we would stop modules individually
			return nil
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-auth",
		Phase:    PhaseMainShutdown,
		Timeout:  5 * time.Second,
		Required: true,
		Priority: 6,
		ExecuteFunc: func(ctx context.Context) error {
			// For now, this is a no-op since we stop all modules at once
			// In a more refined implementation, we would stop modules individually
			return nil
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "stop-storage",
		Phase:    PhaseMainShutdown,
		Timeout:  10 * time.Second,
		Required: true,
		Priority: 7,
		ExecuteFunc: func(ctx context.Context) error {
			// For now, this is a no-op since we stop all modules at once
			// In a more refined implementation, we would stop modules individually
			return nil
		},
	})

	// Post-shutdown phase: cleanup resources
	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "execute-legacy-hooks",
		Phase:    PhasePostShutdown,
		Timeout:  10 * time.Second,
		Required: false,
		Priority: 1,
		ExecuteFunc: func(ctx context.Context) error {
			return lm.executeShutdownHooks(ctx)
		},
	})

	lm.shutdownMgr.AddShutdownStep(ShutdownStep{
		Name:     "cleanup-resources",
		Phase:    PhasePostShutdown,
		Timeout:  5 * time.Second,
		Required: true,
		Priority: 2,
		ExecuteFunc: func(ctx context.Context) error {
			lm.logger.Info("Cleaning up resources")
			// TODO: Implement resource cleanup
			return nil
		},
	})

	lm.logger.Info("Shutdown steps configured",
		zap.Int("total_steps", lm.shutdownMgr.GetStepCount()),
	)
}

// enhancedShutdown performs the enhanced shutdown process
func (lm *LifecycleManager) enhancedShutdown(ctx context.Context) error {
	lm.logger.Info("Starting enhanced shutdown process")

	// Stop signal manager first to prevent new signals
	if lm.signalMgr != nil && lm.signalMgr.IsListening() {
		lm.signalMgr.StopListening()
	}

	// Execute shutdown steps using shutdown manager
	if err := lm.shutdownMgr.ExecuteShutdown(ctx); err != nil {
		lm.logger.Error("Enhanced shutdown failed", zap.Error(err))
		lm.setError(err)
		return err
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		lm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		lm.logger.Info("All goroutines finished gracefully")
	case <-ctx.Done():
		lm.logger.Warn("Shutdown timeout reached during goroutine wait")
	}

	lm.setState(StateStopped)
	lm.logger.Info("Enhanced shutdown completed successfully",
		zap.Duration("uptime", time.Since(lm.startTime)),
		zap.Duration("shutdown_duration", lm.shutdownMgr.GetDuration()),
	)

	return nil
}

// legacyShutdown performs the legacy shutdown process
func (lm *LifecycleManager) legacyShutdown(ctx context.Context) error {
	lm.logger.Info("Starting legacy shutdown process")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, lm.shutdownTimeout)
	defer cancel()

	// Execute shutdown hooks
	if err := lm.executeShutdownHooks(shutdownCtx); err != nil {
		lm.logger.Error("Error executing shutdown hooks", zap.Error(err))
	}

	// Stop all modules
	if err := lm.starter.StopModules(shutdownCtx); err != nil {
		lm.logger.Error("Error stopping modules", zap.Error(err))
		lm.setError(err)
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		lm.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		lm.logger.Info("All goroutines finished gracefully")
	case <-shutdownCtx.Done():
		lm.logger.Warn("Shutdown timeout reached, forcing exit")
	}

	lm.setState(StateStopped)
	lm.logger.Info("Legacy shutdown completed successfully",
		zap.Duration("uptime", time.Since(lm.startTime)),
	)

	return nil
}

// handleGracefulShutdown handles graceful shutdown signals
func (lm *LifecycleManager) handleGracefulShutdown(ctx context.Context, sig os.Signal) error {
	lm.logger.Info("Graceful shutdown signal received",
		zap.String("signal", sig.String()),
	)

	// Initiate graceful shutdown
	go func() {
		if err := lm.Stop(ctx); err != nil {
			lm.logger.Error("Graceful shutdown failed", zap.Error(err))
		}
	}()

	return nil
}

// handleForceShutdown handles force shutdown signals
func (lm *LifecycleManager) handleForceShutdown(ctx context.Context, sig os.Signal) error {
	lm.logger.Warn("Force shutdown signal received",
		zap.String("signal", sig.String()),
	)

	// Initiate force shutdown
	go func() {
		lm.ForceShutdown()
	}()

	return nil
}

// GetShutdownManager returns the shutdown manager for direct access
func (lm *LifecycleManager) GetShutdownManager() *ShutdownManager {
	return lm.shutdownMgr
}

// GetSignalManager returns the signal manager for direct access
func (lm *LifecycleManager) GetSignalManager() *SignalManager {
	return lm.signalMgr
}

// SetUseEnhancedShutdown enables or disables enhanced shutdown
func (lm *LifecycleManager) SetUseEnhancedShutdown(enabled bool) {
	if lm.GetState() != StateStopped {
		lm.logger.Warn("Cannot change enhanced shutdown setting while application is running")
		return
	}

	lm.useEnhancedShutdown = enabled
	lm.logger.Info("Enhanced shutdown setting changed",
		zap.Bool("enabled", enabled),
	)
}

// GetSignalStats returns signal handling statistics
func (lm *LifecycleManager) GetSignalStats() map[os.Signal]int {
	if lm.signalMgr != nil {
		return lm.signalMgr.GetSignalStats()
	}
	return make(map[os.Signal]int)
}

// GetShutdownResults returns the results of the last shutdown process
func (lm *LifecycleManager) GetShutdownResults() []ShutdownResult {
	if lm.shutdownMgr != nil {
		return lm.shutdownMgr.GetResults()
	}
	return make([]ShutdownResult, 0)
}