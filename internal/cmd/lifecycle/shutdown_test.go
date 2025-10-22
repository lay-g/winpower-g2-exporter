package lifecycle

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// MockLogger implements the log.Logger interface for testing
type mockLogger struct {
	logs []string
	mu   sync.Mutex
}

func (m *mockLogger) Debug(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockLogger) Info(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockLogger) Warn(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockLogger) Error(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockLogger) Fatal(msg string, fields ...log.Field) {
	m.log(msg)
}

func (m *mockLogger) With(fields ...log.Field) log.Logger {
	return m
}

func (m *mockLogger) WithContext(ctx context.Context) log.Logger {
	return m
}

func (m *mockLogger) Sync() error {
	return nil
}

func (m *mockLogger) log(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) GetLogs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	logs := make([]string, len(m.logs))
	copy(logs, m.logs)
	return logs
}

func (m *mockLogger) ClearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = []string{}
}

func TestDefaultShutdownConfig(t *testing.T) {
	config := DefaultShutdownConfig()

	assert.Equal(t, 30*time.Second, config.DefaultTimeout)
	assert.Equal(t, 10*time.Second, config.PhaseTimeouts[PhasePreShutdown])
	assert.Equal(t, 60*time.Second, config.PhaseTimeouts[PhaseMainShutdown])
	assert.Equal(t, 10*time.Second, config.PhaseTimeouts[PhasePostShutdown])
	assert.False(t, config.ParallelExecution)
	assert.Equal(t, 5*time.Second, config.ForceKillAfter)
	assert.True(t, config.EnableGraceful)
	assert.True(t, config.LogShutdownSteps)
}

func TestNewShutdownManager(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()

	sm := NewShutdownManager(logger, config)

	assert.NotNil(t, sm)
	assert.Equal(t, logger, sm.logger)
	assert.Equal(t, config, sm.config)
	assert.Empty(t, sm.steps)
	assert.Empty(t, sm.results)
	assert.False(t, sm.isShutdown)
	assert.True(t, sm.startTime.IsZero())
}

func TestNewShutdownManagerWithNilConfig(t *testing.T) {
	logger := &mockLogger{}

	sm := NewShutdownManager(logger, nil)

	assert.NotNil(t, sm)
	assert.NotNil(t, sm.config)
	assert.Equal(t, 30*time.Second, sm.config.DefaultTimeout)
}

func TestAddShutdownStep(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	step := ShutdownStep{
		Name:        "test-step",
		Phase:       PhaseMainShutdown,
		Timeout:     5 * time.Second,
		Required:    true,
		Priority:    1,
		ExecuteFunc: func(ctx context.Context) error { return nil },
	}

	sm.AddShutdownStep(step)

	assert.Equal(t, 1, sm.GetStepCount())
	assert.Contains(t, logger.GetLogs(), "Shutdown step added")
}

func TestAddShutdownSteps(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	steps := []ShutdownStep{
		{
			Name:        "step-1",
			Phase:       PhasePreShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error { return nil },
		},
		{
			Name:        "step-2",
			Phase:       PhaseMainShutdown,
			Timeout:     2 * time.Second,
			Required:    false,
			Priority:    2,
			ExecuteFunc: func(ctx context.Context) error { return nil },
		},
	}

	sm.AddShutdownSteps(steps)

	assert.Equal(t, 2, sm.GetStepCount())
}

func TestExecuteShutdownSuccess(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Add test steps
	executedOrder := make([]string, 0)
	var mu sync.Mutex

	steps := []ShutdownStep{
		{
			Name:        "pre-step",
			Phase:       PhasePreShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				mu.Lock()
				executedOrder = append(executedOrder, "pre-step")
				mu.Unlock()
				return nil
			},
		},
		{
			Name:        "main-step-1",
			Phase:       PhaseMainShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    2,
			ExecuteFunc: func(ctx context.Context) error {
				mu.Lock()
				executedOrder = append(executedOrder, "main-step-1")
				mu.Unlock()
				return nil
			},
		},
		{
			Name:        "main-step-2",
			Phase:       PhaseMainShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				mu.Lock()
				executedOrder = append(executedOrder, "main-step-2")
				mu.Unlock()
				return nil
			},
		},
		{
			Name:        "post-step",
			Phase:       PhasePostShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				mu.Lock()
				executedOrder = append(executedOrder, "post-step")
				mu.Unlock()
				return nil
			},
		},
	}

	sm.AddShutdownSteps(steps)

	ctx := context.Background()
	err := sm.ExecuteShutdown(ctx)

	assert.NoError(t, err)
	assert.True(t, sm.IsShutdown())
	assert.NotZero(t, sm.startTime)

	// Check execution order: pre, main (priority 1 then 2), post
	expectedOrder := []string{"pre-step", "main-step-2", "main-step-1", "post-step"}
	assert.Equal(t, expectedOrder, executedOrder)

	// Check results
	results := sm.GetResults()
	assert.Len(t, results, 4)
	for _, result := range results {
		assert.True(t, result.Success)
		assert.NoError(t, result.Error)
		assert.False(t, result.TimeoutHit)
		assert.False(t, result.Skipped)
	}
}

func TestExecuteShutdownWithRequiredStepFailure(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Add test steps with a required step that fails
	steps := []ShutdownStep{
		{
			Name:        "successful-step",
			Phase:       PhasePreShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error { return nil },
		},
		{
			Name:        "failing-step",
			Phase:       PhaseMainShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				return errors.New("step failed")
			},
		},
	}

	sm.AddShutdownSteps(steps)

	ctx := context.Background()
	err := sm.ExecuteShutdown(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required step 'failing-step' failed")

	// Check that the successful step still executed
	results := sm.GetResults()
	assert.Len(t, results, 2)
	assert.Equal(t, "successful-step", results[0].Step.Name)
	assert.True(t, results[0].Success)
	assert.Equal(t, "failing-step", results[1].Step.Name)
	assert.False(t, results[1].Success)
}

func TestExecuteShutdownWithOptionalStepFailure(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Add test steps with an optional step that fails
	steps := []ShutdownStep{
		{
			Name:        "successful-step",
			Phase:       PhasePreShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error { return nil },
		},
		{
			Name:        "failing-optional-step",
			Phase:       PhaseMainShutdown,
			Timeout:     1 * time.Second,
			Required:    false,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				return errors.New("optional step failed")
			},
		},
		{
			Name:        "another-successful-step",
			Phase:       PhasePostShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error { return nil },
		},
	}

	sm.AddShutdownSteps(steps)

	ctx := context.Background()
	err := sm.ExecuteShutdown(ctx)

	assert.NoError(t, err) // Should succeed because the failing step is optional

	// Check that all steps executed
	results := sm.GetResults()
	assert.Len(t, results, 3)
	assert.Equal(t, "successful-step", results[0].Step.Name)
	assert.True(t, results[0].Success)
	assert.Equal(t, "failing-optional-step", results[1].Step.Name)
	assert.False(t, results[1].Success)
	assert.Equal(t, "another-successful-step", results[2].Step.Name)
	assert.True(t, results[2].Success)
}

func TestExecuteShutdownWithTimeout(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Add a step that will timeout
	step := ShutdownStep{
		Name:        "timeout-step",
		Phase:       PhaseMainShutdown,
		Timeout:     100 * time.Millisecond,
		Required:    true,
		Priority:    1,
		ExecuteFunc: func(ctx context.Context) error {
			select {
			case <-time.After(200 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	sm.AddShutdownStep(step)

	ctx := context.Background()
	start := time.Now()
	err := sm.ExecuteShutdown(ctx)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.True(t, duration >= 100*time.Millisecond) // Should wait at least the timeout
	assert.True(t, duration < 500*time.Millisecond) // But not too much longer

	// Check results
	results := sm.GetResults()
	assert.Len(t, results, 1)
	assert.False(t, results[0].Success)
	assert.True(t, results[0].TimeoutHit)
}

func TestExecuteShutdownAlreadyInProgress(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Add a simple step
	step := ShutdownStep{
		Name:        "test-step",
		Phase:       PhaseMainShutdown,
		Timeout:     1 * time.Second,
		Required:    true,
		Priority:    1,
		ExecuteFunc: func(ctx context.Context) error { return nil },
	}

	sm.AddShutdownStep(step)

	// Mark as shutdown in progress
	sm.isShutdown = true

	ctx := context.Background()
	err := sm.ExecuteShutdown(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown already in progress")
}

func TestParallelExecution(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	config.ParallelExecution = true
	sm := NewShutdownManager(logger, config)

	// Add steps that can run in parallel
	var mu sync.Mutex
	executedOrder := make([]string, 0)

	steps := []ShutdownStep{
		{
			Name:        "step-1",
			Phase:       PhaseMainShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				mu.Lock()
				executedOrder = append(executedOrder, "step-1")
				mu.Unlock()
				return nil
			},
		},
		{
			Name:        "step-2",
			Phase:       PhaseMainShutdown,
			Timeout:     1 * time.Second,
			Required:    true,
			Priority:    1,
			ExecuteFunc: func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				mu.Lock()
				executedOrder = append(executedOrder, "step-2")
				mu.Unlock()
				return nil
			},
		},
	}

	sm.AddShutdownSteps(steps)

	ctx := context.Background()
	start := time.Now()
	err := sm.ExecuteShutdown(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	// Should complete in roughly the time of the longest step (100ms) plus overhead
	assert.True(t, duration < 200*time.Millisecond)

	// Both steps should have executed
	results := sm.GetResults()
	assert.Len(t, results, 2)
	for _, result := range results {
		assert.True(t, result.Success)
	}
}

func TestClearSteps(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Add some steps
	step := ShutdownStep{
		Name:        "test-step",
		Phase:       PhaseMainShutdown,
		Timeout:     1 * time.Second,
		Required:    true,
		Priority:    1,
		ExecuteFunc: func(ctx context.Context) error { return nil },
	}

	sm.AddShutdownStep(step)
	assert.Equal(t, 1, sm.GetStepCount())

	// Clear steps
	sm.ClearSteps()
	assert.Equal(t, 0, sm.GetStepCount())
	assert.Empty(t, sm.GetResults())
	assert.Contains(t, logger.GetLogs(), "All shutdown steps cleared")
}

func TestGetDuration(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Before shutdown starts
	assert.Equal(t, time.Duration(0), sm.GetDuration())

	// Add a simple step and execute shutdown
	step := ShutdownStep{
		Name:        "test-step",
		Phase:       PhaseMainShutdown,
		Timeout:     1 * time.Second,
		Required:    true,
		Priority:    1,
		ExecuteFunc: func(ctx context.Context) error { return nil },
	}

	sm.AddShutdownStep(step)

	ctx := context.Background()
	sm.ExecuteShutdown(ctx)

	// After shutdown
	assert.Greater(t, sm.GetDuration(), time.Duration(0))
}

func TestCancel(t *testing.T) {
	logger := &mockLogger{}
	config := DefaultShutdownConfig()
	sm := NewShutdownManager(logger, config)

	// Cancel the context
	sm.Cancel()

	// The context should be cancelled
	assert.Error(t, sm.ctx.Err())
	assert.Equal(t, context.Canceled, sm.ctx.Err())
}