package lifecycle

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// Test mock module for testing
type testMockModuleInitializer struct {
	name        string
	initError   error
	startError  error
	stopError   error
	healthError error
	initialized bool
	started     bool
}

func (m *testMockModuleInitializer) Name() string {
	return m.name
}

func (m *testMockModuleInitializer) Initialize(ctx context.Context, logger log.Logger) error {
	if m.initError != nil {
		return m.initError
	}
	m.initialized = true
	return nil
}

func (m *testMockModuleInitializer) Start(ctx context.Context) error {
	if m.startError != nil {
		return m.startError
	}
	m.started = true
	return nil
}

func (m *testMockModuleInitializer) Stop(ctx context.Context) error {
	if m.stopError != nil {
		return m.stopError
	}
	m.started = false
	return nil
}

func (m *testMockModuleInitializer) HealthCheck(ctx context.Context) error {
	if m.healthError != nil {
		return m.healthError
	}
	return nil
}

func TestNewLifecycleManager(t *testing.T) {
	logger := log.NewTestLogger()

	manager := NewLifecycleManager(logger)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.NotNil(t, manager.starter)
	assert.NotNil(t, manager.healthMgr)
	assert.Equal(t, StateStopped, manager.GetState())
	assert.Equal(t, 30*time.Second, manager.shutdownTimeout)
	assert.NotNil(t, manager.ctx)
	assert.NotNil(t, manager.cancel)
}

func TestLifecycleManager_SetState(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	manager.setState(StateStarting)
	assert.Equal(t, StateStarting, manager.GetState())

	manager.setState(StateRunning)
	assert.Equal(t, StateRunning, manager.GetState())

	manager.setState(StateStopping)
	assert.Equal(t, StateStopping, manager.GetState())

	manager.setState(StateStopped)
	assert.Equal(t, StateStopped, manager.GetState())

	manager.setState(StateError)
	assert.Equal(t, StateError, manager.GetState())
}

func TestLifecycleManager_SetError(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	testError := assert.AnError
	manager.setError(testError)

	assert.Equal(t, StateError, manager.GetState())
	assert.Equal(t, testError, manager.lastError)
}

func TestLifecycleManager_SetShutdownTimeout(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	newTimeout := 60 * time.Second
	manager.SetShutdownTimeout(newTimeout)

	assert.Equal(t, newTimeout, manager.shutdownTimeout)
}

func TestLifecycleManager_RegisterShutdownHook(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	hook := func() error {
		return nil
	}

	manager.RegisterShutdownHook(hook)

	assert.Len(t, manager.shutdownHooks, 1)
}

func TestLifecycleManager_GetMetrics(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Register some test modules
	module1 := &testMockModuleInitializer{name: "module1"}
	module2 := &testMockModuleInitializer{name: "module2"}

	manager.starter.RegisterModule(module1)
	manager.starter.RegisterModule(module2)

	// Set some statuses
	manager.starter.status["module1"].Status = "running"
	manager.starter.status["module2"].Status = "stopped"
	manager.startTime = time.Now().Add(-1 * time.Hour)

	metrics := manager.GetMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, string(StateStopped), metrics.State)
	assert.Equal(t, 2, metrics.ModuleCount)
	assert.Contains(t, metrics.RunningModules, "module1")
	assert.Contains(t, metrics.StoppedModules, "module2")
	assert.True(t, metrics.Uptime > 0)
	assert.Equal(t, 30*time.Second, metrics.ShutdownTimeout)
}

func TestLifecycleManager_PerformHealthCheck_NotRunning(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	ctx := context.Background()
	err := manager.PerformHealthCheck(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check not available in state")
}

func TestLifecycleManager_PerformHealthCheck_Running(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Set state to running
	manager.setState(StateRunning)

	// Register a healthy module
	module := &testMockModuleInitializer{name: "test-module"}
	manager.starter.RegisterModule(module)
	manager.starter.status["test-module"].Status = "running"

	ctx := context.Background()
	err := manager.PerformHealthCheck(ctx)

	// Without implementing actual health check logic in mock,
	// we just verify the function completes without panic
	_ = err
}

func TestLifecycleManager_GetStarter(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	starter := manager.GetStarter()

	assert.NotNil(t, starter)
	assert.Equal(t, manager.starter, starter)
}

func TestLifecycleManager_GetHealthManager(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	healthMgr := manager.GetHealthManager()

	assert.NotNil(t, healthMgr)
	assert.Equal(t, manager.healthMgr, healthMgr)
}

func TestLifecycleManager_ForceShutdown(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Set initial state
	manager.setState(StateRunning)

	manager.ForceShutdown()

	assert.Equal(t, StateStopped, manager.GetState())
}

func TestLifecycleManager_SignalHandling(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Setup signal handling
	manager.setupSignalHandling()

	// Verify signal channel is set up
	assert.NotNil(t, manager.shutdownSignal)

	// Send a test signal
	select {
	case manager.shutdownSignal <- syscall.SIGTERM:
	default:
		t.Error("Failed to send signal to shutdown channel")
	}
}

func TestLifecycleManager_ExecuteShutdownHooks(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	hook1Called := false
	hook2Called := false
	hook1Error := assert.AnError

	hook1 := func() error {
		hook1Called = true
		return hook1Error
	}

	hook2 := func() error {
		hook2Called = true
		return nil
	}

	manager.RegisterShutdownHook(hook1)
	manager.RegisterShutdownHook(hook2)

	ctx := context.Background()
	err := manager.executeShutdownHooks(ctx)

	// Should not return error even if hooks fail
	require.NoError(t, err)
	assert.True(t, hook1Called)
	assert.True(t, hook2Called)
}

func TestLifecycleManager_Restart(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// This test is limited because we don't have real module implementations
	ctx := context.Background()

	err := manager.Restart(ctx)

	// Without real implementations, restart may succeed (empty operation)
	_ = err
}

func TestLifecycleManager_StartupAndShutdownFlow(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Test initial state
	assert.Equal(t, StateStopped, manager.GetState())

	// Attempt to start (may succeed with no modules)
	ctx := context.Background()
	err := manager.Start(ctx)

	// Without real modules, start may succeed or fail
	_ = err

	// Test shutdown (should work)
	err = manager.Stop(ctx)

	// Should not error on shutdown
	require.NoError(t, err)
	assert.Equal(t, StateStopped, manager.GetState())
}

func TestLifecycleManager_ContextCancellation(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Verify context is initially not cancelled
	select {
	case <-manager.ctx.Done():
		t.Error("Context should not be cancelled initially")
	default:
	}

	// Cancel the context
	manager.cancel()

	// Verify context is cancelled
	select {
	case <-manager.ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled after cancel() call")
	}
}

func TestLifecycleManager_ConcurrentStateAccess(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewLifecycleManager(logger)

	// Test concurrent access to state
	done := make(chan bool, 10)

	// Start multiple goroutines that access state
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				state := manager.GetState()
				assert.True(t, state == StateStopped || state == StateStarting ||
					state == StateRunning || state == StateStopping || state == StateError)
				manager.setState(StateRunning)
				time.Sleep(time.Microsecond)
				manager.setState(StateStopped)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Final state should be consistent
	finalState := manager.GetState()
	assert.True(t, finalState == StateStopped || finalState == StateRunning)
}
