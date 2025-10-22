package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// ShutdownPhase represents a phase in the shutdown process
type ShutdownPhase string

const (
	PhasePreShutdown  ShutdownPhase = "pre_shutdown"
	PhaseMainShutdown ShutdownPhase = "main_shutdown"
	PhasePostShutdown ShutdownPhase = "post_shutdown"
)

// ShutdownStep represents a single step in the shutdown process
type ShutdownStep struct {
	Name        string
	Phase       ShutdownPhase
	Timeout     time.Duration
	Required    bool  // Whether this step must succeed
	Priority    int   // Lower numbers execute first within the same phase
	ExecuteFunc func(ctx context.Context) error
}

// ShutdownResult represents the result of a shutdown step
type ShutdownResult struct {
	Step         ShutdownStep
	Success      bool
	Error        error
	Duration     time.Duration
	TimeoutHit   bool
	Skipped      bool
	SkipReason   string
}

// ShutdownConfig contains configuration for the shutdown process
type ShutdownConfig struct {
	DefaultTimeout     time.Duration
	PhaseTimeouts      map[ShutdownPhase]time.Duration
	ParallelExecution  bool
	ForceKillAfter     time.Duration
	EnableGraceful     bool
	LogShutdownSteps   bool
}

// DefaultShutdownConfig returns a default shutdown configuration
func DefaultShutdownConfig() *ShutdownConfig {
	return &ShutdownConfig{
		DefaultTimeout: 30 * time.Second,
		PhaseTimeouts: map[ShutdownPhase]time.Duration{
			PhasePreShutdown:  10 * time.Second,
			PhaseMainShutdown: 60 * time.Second,
			PhasePostShutdown: 10 * time.Second,
		},
		ParallelExecution: false, // Execute sequentially for predictable order
		ForceKillAfter:    5 * time.Second,
		EnableGraceful:    true,
		LogShutdownSteps:  true,
	}
}

// ShutdownManager handles the graceful shutdown process
type ShutdownManager struct {
	// Dependencies
	logger log.Logger
	config *ShutdownConfig

	// State management
	steps      []ShutdownStep
	results    []ShutdownResult
	mutex      sync.RWMutex
	isShutdown bool
	startTime  time.Time

	// Context management
	ctx    context.Context
	cancel context.CancelFunc
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(logger log.Logger, config *ShutdownConfig) *ShutdownManager {
	if config == nil {
		config = DefaultShutdownConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ShutdownManager{
		logger:  logger,
		config:  config,
		steps:   make([]ShutdownStep, 0),
		results: make([]ShutdownResult, 0),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// AddShutdownStep adds a new step to the shutdown process
func (sm *ShutdownManager) AddShutdownStep(step ShutdownStep) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.steps = append(sm.steps, step)
	sm.logger.Debug("Shutdown step added",
		zap.String("step_name", step.Name),
		zap.String("phase", string(step.Phase)),
		zap.Duration("timeout", step.Timeout),
		zap.Bool("required", step.Required),
		zap.Int("priority", step.Priority),
	)
}

// AddShutdownSteps adds multiple steps to the shutdown process
func (sm *ShutdownManager) AddShutdownSteps(steps []ShutdownStep) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for _, step := range steps {
		sm.steps = append(sm.steps, step)
		sm.logger.Debug("Shutdown step added",
			zap.String("step_name", step.Name),
			zap.String("phase", string(step.Phase)),
			zap.Duration("timeout", step.Timeout),
			zap.Bool("required", step.Required),
			zap.Int("priority", step.Priority),
		)
	}
}

// ExecuteShutdown executes the complete shutdown process
func (sm *ShutdownManager) ExecuteShutdown(ctx context.Context) error {
	sm.mutex.Lock()
	if sm.isShutdown {
		sm.mutex.Unlock()
		return fmt.Errorf("shutdown already in progress")
	}
	sm.isShutdown = true
	sm.startTime = time.Now()
	sm.mutex.Unlock()

	sm.logger.Info("Starting graceful shutdown process")

	// Group steps by phase
	phases := []ShutdownPhase{PhasePreShutdown, PhaseMainShutdown, PhasePostShutdown}

	for _, phase := range phases {
		if err := sm.executePhase(ctx, phase); err != nil {
			sm.logger.Error("Shutdown phase failed",
				zap.String("phase", string(phase)),
				zap.Error(err),
			)
			return err
		}
	}

	// Wait for all goroutines to finish with final timeout
	if err := sm.waitForCompletion(ctx); err != nil {
		sm.logger.Error("Waiting for completion failed", zap.Error(err))
		return err
	}

	duration := time.Since(sm.startTime)
	sm.logger.Info("Graceful shutdown completed successfully",
		zap.Duration("total_duration", duration),
		zap.Int("total_steps", len(sm.results)),
	)

	return nil
}

// executePhase executes all steps in a specific phase
func (sm *ShutdownManager) executePhase(ctx context.Context, phase ShutdownPhase) error {
	sm.logger.Info("Executing shutdown phase",
		zap.String("phase", string(phase)),
	)

	// Get phase timeout
	phaseTimeout := sm.getPhaseTimeout(phase)
	phaseCtx, cancel := context.WithTimeout(ctx, phaseTimeout)
	defer cancel()

	// Filter steps for this phase
	phaseSteps := sm.getStepsByPhase(phase)
	if len(phaseSteps) == 0 {
		sm.logger.Debug("No steps found for phase", zap.String("phase", string(phase)))
		return nil
	}

	// Sort steps by priority (lower numbers first)
	sm.sortStepsByPriority(phaseSteps)

	// Execute steps
	if sm.config.ParallelExecution {
		return sm.executeStepsParallel(phaseCtx, phaseSteps)
	}
	return sm.executeStepsSequential(phaseCtx, phaseSteps)
}

// executeStepsSequential executes steps one by one in order
func (sm *ShutdownManager) executeStepsSequential(ctx context.Context, steps []ShutdownStep) error {
	for _, step := range steps {
		if err := sm.executeStep(ctx, step); err != nil {
			if step.Required {
				return fmt.Errorf("required step '%s' failed: %w", step.Name, err)
			}
			sm.logger.Warn("Non-required step failed, continuing",
				zap.String("step_name", step.Name),
				zap.Error(err),
			)
		}
	}
	return nil
}

// executeStepsParallel executes steps in parallel (if configured)
func (sm *ShutdownManager) executeStepsParallel(ctx context.Context, steps []ShutdownStep) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(steps))
	resultsChan := make(chan ShutdownResult, len(steps))

	for _, step := range steps {
		wg.Add(1)
		go func(s ShutdownStep) {
			defer wg.Done()
			result := ShutdownResult{
				Step:    s,
				Success: false,
			}

			stepStart := time.Now()
			err := s.ExecuteFunc(ctx)
			result.Duration = time.Since(stepStart)

			if err != nil {
				result.Error = err
				if ctx.Err() == context.DeadlineExceeded {
					result.TimeoutHit = true
				}
			} else {
				result.Success = true
			}

			resultsChan <- result
			if !result.Success && s.Required {
				errChan <- fmt.Errorf("required step '%s' failed: %w", s.Name, err)
			}
		}(step)
	}

	// Wait for all steps to complete
	wg.Wait()
	close(errChan)
	close(resultsChan)

	// Collect results
	for result := range resultsChan {
		sm.mutex.Lock()
		sm.results = append(sm.results, result)
		sm.mutex.Unlock()
	}

	// Check for errors in required steps
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// executeStep executes a single shutdown step
func (sm *ShutdownManager) executeStep(ctx context.Context, step ShutdownStep) error {
	result := ShutdownResult{
		Step:    step,
		Success: false,
	}

	sm.logger.Info("Executing shutdown step",
		zap.String("step_name", step.Name),
		zap.String("phase", string(step.Phase)),
		zap.Duration("timeout", step.Timeout),
	)

	stepStart := time.Now()

	// Create step context with timeout
	stepTimeout := step.Timeout
	if stepTimeout == 0 {
		stepTimeout = sm.config.DefaultTimeout
	}

	stepCtx, cancel := context.WithTimeout(ctx, stepTimeout)
	defer cancel()

	// Execute the step
	err := step.ExecuteFunc(stepCtx)
	result.Duration = time.Since(stepStart)

	if err != nil {
		result.Error = err
		if stepCtx.Err() == context.DeadlineExceeded {
			result.TimeoutHit = true
			sm.logger.Error("Shutdown step timed out",
				zap.String("step_name", step.Name),
				zap.Duration("timeout", stepTimeout),
				zap.Duration("actual_duration", result.Duration),
			)
		} else {
			sm.logger.Error("Shutdown step failed",
				zap.String("step_name", step.Name),
				zap.Duration("duration", result.Duration),
				zap.Error(err),
			)
		}
	} else {
		result.Success = true
		sm.logger.Info("Shutdown step completed successfully",
			zap.String("step_name", step.Name),
			zap.Duration("duration", result.Duration),
		)
	}

	// Store result
	sm.mutex.Lock()
	sm.results = append(sm.results, result)
	sm.mutex.Unlock()

	return err
}

// waitForCompletion waits for all operations to complete with final timeout
func (sm *ShutdownManager) waitForCompletion(ctx context.Context) error {
	if sm.config.ForceKillAfter > 0 {
		finalCtx, cancel := context.WithTimeout(ctx, sm.config.ForceKillAfter)
		defer cancel()

		sm.logger.Info("Waiting for final completion",
			zap.Duration("force_kill_after", sm.config.ForceKillAfter),
		)

		select {
		case <-finalCtx.Done():
			if finalCtx.Err() == context.DeadlineExceeded {
				sm.logger.Warn("Force kill timeout reached, terminating immediately")
				return fmt.Errorf("force kill timeout reached")
			}
			return finalCtx.Err()
		case <-sm.ctx.Done():
			sm.logger.Info("Context cancelled, final completion done")
			return nil
		}
	}

	return nil
}

// getStepsByPhase returns all steps for a specific phase
func (sm *ShutdownManager) getStepsByPhase(phase ShutdownPhase) []ShutdownStep {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var phaseSteps []ShutdownStep
	for _, step := range sm.steps {
		if step.Phase == phase {
			phaseSteps = append(phaseSteps, step)
		}
	}

	return phaseSteps
}

// sortStepsByPriority sorts steps by priority (lower numbers first)
func (sm *ShutdownManager) sortStepsByPriority(steps []ShutdownStep) {
	// Simple insertion sort since we expect small number of steps
	for i := 1; i < len(steps); i++ {
		key := steps[i]
		j := i - 1
		for j >= 0 && steps[j].Priority > key.Priority {
			steps[j+1] = steps[j]
			j--
		}
		steps[j+1] = key
	}
}

// getPhaseTimeout returns the timeout for a specific phase
func (sm *ShutdownManager) getPhaseTimeout(phase ShutdownPhase) time.Duration {
	if timeout, exists := sm.config.PhaseTimeouts[phase]; exists {
		return timeout
	}
	return sm.config.DefaultTimeout
}

// GetResults returns all shutdown results
func (sm *ShutdownManager) GetResults() []ShutdownResult {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Return a copy to prevent external modification
	results := make([]ShutdownResult, len(sm.results))
	copy(results, sm.results)
	return results
}

// IsShutdown returns whether shutdown is in progress
func (sm *ShutdownManager) IsShutdown() bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.isShutdown
}

// Cancel cancels the shutdown context
func (sm *ShutdownManager) Cancel() {
	sm.cancel()
}

// GetDuration returns the duration of the shutdown process
func (sm *ShutdownManager) GetDuration() time.Duration {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if sm.startTime.IsZero() {
		return 0
	}
	return time.Since(sm.startTime)
}

// GetStepCount returns the number of registered shutdown steps
func (sm *ShutdownManager) GetStepCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.steps)
}

// ClearSteps clears all registered shutdown steps
func (sm *ShutdownManager) ClearSteps() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.steps = make([]ShutdownStep, 0)
	sm.results = make([]ShutdownResult, 0)
	sm.logger.Debug("All shutdown steps cleared")
}