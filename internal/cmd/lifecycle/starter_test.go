package lifecycle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func TestNewStarter(t *testing.T) {
		logger := log.NewTestLogger()

	starter := NewStarter(logger)

	assert.NotNil(t, starter)
	assert.Equal(t, logger, starter.logger)
	assert.NotNil(t, starter.modules)
	assert.NotNil(t, starter.status)
}

func TestStarter_RegisterModule(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module := &mockModuleInitializer{name: "test-module"}

	starter.RegisterModule(module)

	assert.Contains(t, starter.modules, "test-module")
	assert.Equal(t, module, starter.modules["test-module"])
	assert.Contains(t, starter.status, "test-module")
	assert.Equal(t, "registered", starter.status["test-module"].Status)
	assert.Equal(t, "test-module", starter.status["test-module"].Name)
}

func TestStarter_InitializeModule_Success(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module := &mockModuleInitializer{name: "test-module"}
	ctx := context.Background()

	// Register the module first to initialize status
	starter.RegisterModule(module)

	err := starter.initializeModule(ctx, module)

	require.NoError(t, err)
	assert.True(t, module.initialized)
	assert.True(t, module.started)
	assert.Equal(t, "running", starter.status["test-module"].Status)
}

func TestStarter_InitializeModule_InitError(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	initError := assert.AnError
	module := &mockModuleInitializer{
		name:      "test-module",
		initError: initError,
	}
	ctx := context.Background()

	// Register the module first to initialize status
	starter.RegisterModule(module)

	err := starter.initializeModule(ctx, module)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "module initialization failed")
	assert.False(t, module.initialized)
	assert.False(t, module.started)
	assert.Equal(t, "error", starter.status["test-module"].Status)
	assert.Equal(t, initError.Error(), starter.status["test-module"].Error)
}

func TestStarter_InitializeModule_StartError(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	startError := assert.AnError
	module := &mockModuleInitializer{
		name:       "test-module",
		startError: startError,
	}
	ctx := context.Background()

	// Register the module first to initialize status
	starter.RegisterModule(module)

	err := starter.initializeModule(ctx, module)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "module start failed")
	assert.True(t, module.initialized)
	assert.False(t, module.started)
	assert.Equal(t, "error", starter.status["test-module"].Status)
	assert.Equal(t, startError.Error(), starter.status["test-module"].Error)
}

func TestStarter_StopModules_Success(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	// Register and start modules with names that are in the stop order
	module1 := &mockModuleInitializer{name: "server"}
	module2 := &mockModuleInitializer{name: "scheduler"}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)

	// Mark modules as started
	module1.started = true
	module2.started = true
	starter.status["server"].Status = "running"
	starter.status["scheduler"].Status = "running"

	ctx := context.Background()

	err := starter.StopModules(ctx)

	require.NoError(t, err)
	assert.False(t, module1.started)
	assert.False(t, module2.started)
	assert.Equal(t, "stopped", starter.status["server"].Status)
	assert.Equal(t, "stopped", starter.status["scheduler"].Status)
}

func TestStarter_StopModules_StopError(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	// Register modules with names that are in the stop order
	module1 := &mockModuleInitializer{name: "server"}
	module2 := &mockModuleInitializer{
		name:      "scheduler",
		stopError: assert.AnError,
	}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)

	// Mark modules as started
	module1.started = true
	module2.started = true
	starter.status["server"].Status = "running"
	starter.status["scheduler"].Status = "running"

	ctx := context.Background()

	err := starter.StopModules(ctx)

	// Should not return error even if some modules fail to stop
	require.NoError(t, err)
	assert.False(t, module1.started)
	// module2 should still be started because its stop method failed
	assert.True(t, module2.started)
}

func TestStarter_PerformHealthCheck_Success(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module1 := &mockModuleInitializer{name: "module1"}
	module2 := &mockModuleInitializer{name: "module2"}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)

	// Mark modules as running
	starter.status["module1"].Status = "running"
	starter.status["module2"].Status = "running"
	module1.started = true
	module2.started = true

	ctx := context.Background()

	err := starter.PerformHealthCheck(ctx)

	require.NoError(t, err)
}

func TestStarter_PerformHealthCheck_HealthCheckError(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module1 := &mockModuleInitializer{name: "module1"}
	module2 := &testMockModuleInitializer{
		name:        "module2",
		healthError: assert.AnError,
	}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)

	// Mark modules as running
	starter.status["module1"].Status = "running"
	starter.status["module2"].Status = "running"
	module1.started = true
	module2.started = true

	ctx := context.Background()

	err := starter.PerformHealthCheck(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed for module module2")
}

func TestStarter_GetModuleStatus(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module := &mockModuleInitializer{name: "test-module"}
	starter.RegisterModule(module)

	status := starter.GetModuleStatus()

	assert.Contains(t, status, "test-module")
	assert.Equal(t, "test-module", status["test-module"].Name)
	assert.Equal(t, "registered", status["test-module"].Status)
}

func TestStarter_GetStartupMetrics(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module1 := &mockModuleInitializer{name: "module1"}
	module2 := &mockModuleInitializer{name: "module2"}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)

	// Mark modules as running
	starter.status["module1"].Status = "running"
	starter.status["module2"].Status = "running"

	metrics := starter.GetStartupMetrics()

	assert.NotNil(t, metrics)
	assert.Equal(t, 2, metrics["module_count"])
	assert.Contains(t, metrics, "startup_time")
	assert.Contains(t, metrics, "uptime")
	assert.Contains(t, metrics, "initialized_modules")
	assert.Contains(t, metrics, "running_modules")
}

func TestStarter_GetInitializedModules(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module1 := &mockModuleInitializer{name: "module1"}
	module2 := &mockModuleInitializer{name: "module2"}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)

	// Mark modules with different statuses
	starter.status["module1"].Status = "running"
	starter.status["module2"].Status = "stopped"

	modules := starter.getInitializedModules()

	assert.Len(t, modules, 2)
	assert.Contains(t, modules, "module1")
	assert.Contains(t, modules, "module2")
}

func TestStarter_GetRunningModules(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	module1 := &mockModuleInitializer{name: "module1"}
	module2 := &mockModuleInitializer{name: "module2"}
	module3 := &mockModuleInitializer{name: "module3"}

	starter.RegisterModule(module1)
	starter.RegisterModule(module2)
	starter.RegisterModule(module3)

	// Mark modules with different statuses
	starter.status["module1"].Status = "running"
	starter.status["module2"].Status = "stopped"
	starter.status["module3"].Status = "error"

	modules := starter.getRunningModules()

	assert.Len(t, modules, 1)
	assert.Contains(t, modules, "module1")
	assert.NotContains(t, modules, "module2")
	assert.NotContains(t, modules, "module3")
}

func TestStarter_SetupModules(t *testing.T) {
		logger := log.NewTestLogger()
	starter := NewStarter(logger)

	ctx := context.Background()

	err := starter.SetupModules(ctx)

	require.NoError(t, err)
	assert.Equal(t, 6, len(starter.modules))
	assert.Contains(t, starter.modules, "storage")
	assert.Contains(t, starter.modules, "energy")
	assert.Contains(t, starter.modules, "winpower")
	assert.Contains(t, starter.modules, "metrics")
	assert.Contains(t, starter.modules, "server")
	assert.Contains(t, starter.modules, "scheduler")
}