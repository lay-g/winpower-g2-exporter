package lifecycle

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// mockHealthChecker implements HealthChecker for testing
type mockHealthChecker struct {
	name        string
	status      HealthStatus
	message     string
	responseTime time.Duration
	error       error
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func (m *mockHealthChecker) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := &HealthCheckResult{
		Component:    m.name,
		Status:       m.status,
		Message:      m.message,
		ResponseTime: m.responseTime,
		LastCheck:    time.Now(),
	}

	if m.error != nil {
		result.Status = HealthStatusUnhealthy
		result.Message = m.error.Error()
	}

	return result
}

func TestNewHealthManager(t *testing.T) {
	logger := log.NewTestLogger()

	manager := NewHealthManager(logger)

	assert.NotNil(t, manager)
	assert.Equal(t, logger, manager.logger)
	assert.NotNil(t, manager.checkers)
	assert.NotNil(t, manager.results)
	assert.Equal(t, 30*time.Second, manager.checkInterval)
	assert.Equal(t, 10*time.Second, manager.timeout)
}

func TestHealthManager_RegisterChecker(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker := &mockHealthChecker{
		name:    "test-component",
		status:  HealthStatusHealthy,
		message: "All good",
	}

	manager.RegisterChecker(checker)

	assert.Contains(t, manager.checkers, "test-component")
	assert.Equal(t, checker, manager.checkers["test-component"])
	assert.Contains(t, manager.results, "test-component")
	assert.Equal(t, HealthStatusUnknown, manager.results["test-component"].Status)
	assert.Equal(t, "Not checked yet", manager.results["test-component"].Message)
}

func TestHealthManager_UnregisterChecker(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker := &mockHealthChecker{
		name: "test-component",
	}

	manager.RegisterChecker(checker)
	assert.Contains(t, manager.checkers, "test-component")

	manager.UnregisterChecker("test-component")
	assert.NotContains(t, manager.checkers, "test-component")
	assert.NotContains(t, manager.results, "test-component")
}

func TestHealthManager_PerformHealthCheck_Success(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker := &mockHealthChecker{
		name:         "test-component",
		status:       HealthStatusHealthy,
		message:      "Component is healthy",
		responseTime: 50 * time.Millisecond,
	}

	manager.RegisterChecker(checker)

	ctx := context.Background()
	result, err := manager.PerformHealthCheck(ctx, "test-component")

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-component", result.Component)
	assert.Equal(t, HealthStatusHealthy, result.Status)
	assert.Equal(t, "Component is healthy", result.Message)
	assert.True(t, result.ResponseTime > 0)
	assert.True(t, result.LastCheck.After(time.Time{}))

	// Verify result is stored
	storedResult := manager.results["test-component"]
	assert.Equal(t, result.Component, storedResult.Component)
	assert.Equal(t, result.Status, storedResult.Status)
	assert.Equal(t, result.Message, storedResult.Message)
	assert.Equal(t, result.LastCheck, storedResult.LastCheck)
}

func TestHealthManager_PerformHealthCheck_NotFound(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	ctx := context.Background()
	result, err := manager.PerformHealthCheck(ctx, "non-existent-component")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "health checker not found")
}

func TestHealthManager_PerformHealthCheck_Timeout(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	// Set a very short timeout for testing
	manager.SetTimeout(10 * time.Millisecond)

	checker := &mockHealthChecker{
		name: "slow-component",
	}

	manager.RegisterChecker(checker)

	ctx := context.Background()
	result, err := manager.PerformHealthCheck(ctx, "slow-component")

	// Should complete successfully
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestHealthManager_PerformAllHealthChecks(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker1 := &mockHealthChecker{
		name:    "component1",
		status:  HealthStatusHealthy,
		message: "Component 1 is healthy",
	}

	checker2 := &mockHealthChecker{
		name:    "component2",
		status:  HealthStatusUnhealthy,
		message: "Component 2 has issues",
	}

	checker3 := &mockHealthChecker{
		name:    "component3",
		status:  HealthStatusHealthy,
		message: "Component 3 is healthy",
	}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)
	manager.RegisterChecker(checker3)

	ctx := context.Background()
	results := manager.PerformAllHealthChecks(ctx)

	assert.Len(t, results, 3)
	assert.Contains(t, results, "component1")
	assert.Contains(t, results, "component2")
	assert.Contains(t, results, "component3")

	assert.Equal(t, HealthStatusHealthy, results["component1"].Status)
	assert.Equal(t, HealthStatusUnhealthy, results["component2"].Status)
	assert.Equal(t, HealthStatusHealthy, results["component3"].Status)

	// Verify results are stored
	assert.Len(t, manager.results, 3)
}

func TestHealthManager_PerformSystemHealthCheck(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	ctx := context.Background()
	err := manager.PerformSystemHealthCheck(ctx)

	require.NoError(t, err)

	// Verify system health is recorded
	assert.NotNil(t, manager.systemHealth)
	assert.True(t, !manager.lastSystemCheck.IsZero())

	systemHealth := manager.GetSystemHealth()
	assert.NotNil(t, systemHealth)
	assert.True(t, systemHealth.GoroutineCount >= 0)
	assert.True(t, systemHealth.MemoryUsage.Alloc >= 0)
}

func TestHealthManager_GetHealthStatus(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker1 := &mockHealthChecker{
		name:    "component1",
		status:  HealthStatusHealthy,
		message: "Component 1 is healthy",
	}

	checker2 := &mockHealthChecker{
		name:    "component2",
		status:  HealthStatusUnhealthy,
		message: "Component 2 has issues",
	}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)

	// Perform health checks to populate results
	ctx := context.Background()
	manager.PerformHealthCheck(ctx, "component1")
	manager.PerformHealthCheck(ctx, "component2")

	status := manager.GetHealthStatus()

	assert.Len(t, status, 2)
	assert.Contains(t, status, "component1")
	assert.Contains(t, status, "component2")

	// Verify returned results are copies (not references to internal data)
	status["component1"].Message = "Modified message"
	assert.NotEqual(t, "Modified message", manager.results["component1"].Message)
}

func TestHealthManager_IsHealthy(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	// Empty manager should be considered healthy
	assert.True(t, manager.IsHealthy())

	checker1 := &mockHealthChecker{
		name:    "component1",
		status:  HealthStatusHealthy,
		message: "Component 1 is healthy",
	}

	checker2 := &mockHealthChecker{
		name:    "component2",
		status:  HealthStatusHealthy,
		message: "Component 2 is healthy",
	}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)

	// Set up results manually
	manager.results["component1"] = &HealthCheckResult{
		Component: "component1",
		Status:    HealthStatusHealthy,
	}
	manager.results["component2"] = &HealthCheckResult{
		Component: "component2",
		Status:    HealthStatusHealthy,
	}

	assert.True(t, manager.IsHealthy())

	// Make one component unhealthy
	manager.results["component2"].Status = HealthStatusUnhealthy
	assert.False(t, manager.IsHealthy())

	// Make one component unknown
	manager.results["component2"].Status = HealthStatusUnknown
	assert.False(t, manager.IsHealthy())
}

func TestHealthManager_GetUnhealthyComponents(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker1 := &mockHealthChecker{
		name:    "component1",
		status:  HealthStatusHealthy,
		message: "Component 1 is healthy",
	}

	checker2 := &mockHealthChecker{
		name:    "component2",
		status:  HealthStatusUnhealthy,
		message: "Component 2 has issues",
	}

	checker3 := &mockHealthChecker{
		name:    "component3",
		status:  HealthStatusUnknown,
		message: "Component 3 status unknown",
	}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)
	manager.RegisterChecker(checker3)

	// Set up results manually
	manager.results["component1"] = &HealthCheckResult{
		Component: "component1",
		Status:    HealthStatusHealthy,
	}
	manager.results["component2"] = &HealthCheckResult{
		Component: "component2",
		Status:    HealthStatusUnhealthy,
	}
	manager.results["component3"] = &HealthCheckResult{
		Component: "component3",
		Status:    HealthStatusUnknown,
	}

	unhealthy := manager.GetUnhealthyComponents()

	assert.Len(t, unhealthy, 2)
	assert.Contains(t, unhealthy, "component2")
	assert.Contains(t, unhealthy, "component3")
	assert.NotContains(t, unhealthy, "component1")
}

func TestHealthManager_SetCheckInterval(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	newInterval := 60 * time.Second
	manager.SetCheckInterval(newInterval)

	assert.Equal(t, newInterval, manager.checkInterval)
}

func TestHealthManager_SetTimeout(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	newTimeout := 20 * time.Second
	manager.SetTimeout(newTimeout)

	assert.Equal(t, newTimeout, manager.timeout)
}

func TestHealthManager_GetRegisteredCheckers(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker1 := &mockHealthChecker{name: "component1"}
	checker2 := &mockHealthChecker{name: "component2"}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)

	names := manager.GetRegisteredCheckers()

	assert.Len(t, names, 2)
	assert.Contains(t, names, "component1")
	assert.Contains(t, names, "component2")
}

func TestHealthManager_ClearResults(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker := &mockHealthChecker{
		name:    "test-component",
		status:  HealthStatusHealthy,
		message: "Component is healthy",
	}

	manager.RegisterChecker(checker)

	// Set up a result
	ctx := context.Background()
	manager.PerformHealthCheck(ctx, "test-component")

	// Verify result exists
	assert.NotEqual(t, "Not checked yet", manager.results["test-component"].Message)

	// Clear results
	manager.ClearResults()

	// Verify results are cleared
	assert.Equal(t, "Not checked yet", manager.results["test-component"].Message)
	assert.Equal(t, HealthStatusUnknown, manager.results["test-component"].Status)
	assert.True(t, manager.results["test-component"].LastCheck.IsZero())
}

func TestHealthManager_GetHealthSummary(t *testing.T) {
	logger := log.NewTestLogger()
	manager := NewHealthManager(logger)

	checker1 := &mockHealthChecker{
		name:    "component1",
		status:  HealthStatusHealthy,
		message: "Component 1 is healthy",
	}

	checker2 := &mockHealthChecker{
		name:    "component2",
		status:  HealthStatusUnhealthy,
		message: "Component 2 has issues",
	}

	checker3 := &mockHealthChecker{
		name:    "component3",
		status:  HealthStatusUnknown,
		message: "Component 3 status unknown",
	}

	manager.RegisterChecker(checker1)
	manager.RegisterChecker(checker2)
	manager.RegisterChecker(checker3)

	// Set up results manually with timestamps
	now := time.Now()
	manager.results["component1"] = &HealthCheckResult{
		Component: "component1",
		Status:    HealthStatusHealthy,
		Message:   "Component 1 is healthy",
		LastCheck: now,
	}
	manager.results["component2"] = &HealthCheckResult{
		Component: "component2",
		Status:    HealthStatusUnhealthy,
		Message:   "Component 2 has issues",
		LastCheck: now.Add(-1 * time.Minute),
	}
	manager.results["component3"] = &HealthCheckResult{
		Component: "component3",
		Status:    HealthStatusUnknown,
		Message:   "Component 3 status unknown",
		LastCheck: now.Add(-2 * time.Minute),
	}

	// Set up system health
	manager.systemHealth = &SystemHealthCheck{
		Uptime:         1 * time.Hour,
		GoroutineCount: 42,
		MemoryUsage: MemoryStats{
			Alloc:      1024 * 1024,
			TotalAlloc: 2 * 1024 * 1024,
			Sys:        4 * 1024 * 1024,
			NumGC:      10,
		},
	}

	summary := manager.GetHealthSummary()

	assert.NotNil(t, summary)
	assert.Equal(t, HealthStatusUnhealthy, summary.OverallStatus)
	assert.Equal(t, 3, summary.TotalComponents)
	assert.Equal(t, 1, summary.HealthyComponents)
	assert.Len(t, summary.UnhealthyComponents, 2)
	assert.Contains(t, summary.UnhealthyComponents, "component2")
	assert.Contains(t, summary.UnhealthyComponents, "component3")
	assert.Equal(t, now, summary.LastCheck)
	assert.NotNil(t, summary.SystemHealth)
	assert.Len(t, summary.ComponentResults, 3)

	// Verify component results are included
	assert.Contains(t, summary.ComponentResults, "component1")
	assert.Contains(t, summary.ComponentResults, "component2")
	assert.Contains(t, summary.ComponentResults, "component3")
}