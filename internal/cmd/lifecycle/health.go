package lifecycle

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Component    string        `json:"component"`
	Status       HealthStatus  `json:"status"`
	Message      string        `json:"message"`
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	Details      interface{}   `json:"details,omitempty"`
}

// HealthChecker defines the interface for health checkable components
type HealthChecker interface {
	// Name returns the name of the component
	Name() string

	// HealthCheck performs a health check on the component
	HealthCheck(ctx context.Context) *HealthCheckResult
}

// SystemHealthCheck represents system-level health information
type SystemHealthCheck struct {
	Uptime         time.Duration `json:"uptime"`
	MemoryUsage    MemoryStats   `json:"memory_usage"`
	GoroutineCount int           `json:"goroutine_count"`
	GCStats        GCStats       `json:"gc_stats"`
}

// MemoryStats represents memory usage statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc"`       // bytes allocated and still in use
	TotalAlloc uint64 `json:"total_alloc"` // bytes allocated (even if freed)
	Sys        uint64 `json:"sys"`         // bytes obtained from system
	NumGC      uint32 `json:"num_gc"`      // number of GC runs
}

// GCStats represents garbage collection statistics
type GCStats struct {
	NumGC         uint32    `json:"num_gc"`
	NumForcedGC   uint32    `json:"num_forced_gc"`
	GCCPUFraction float64   `json:"gc_cpu_fraction"`
	LastGC        time.Time `json:"last_gc"`
	NextGC        uint64    `json:"next_gc"`
	PauseTotalNs  uint64    `json:"pause_total_ns"`
	PauseNs       []uint64  `json:"pause_ns"`
}

// HealthManager manages health checks for all application components
type HealthManager struct {
	checkers        map[string]HealthChecker
	results         map[string]*HealthCheckResult
	logger          log.Logger
	mutex           sync.RWMutex
	checkInterval   time.Duration
	timeout         time.Duration
	lastSystemCheck time.Time
	systemHealth    *SystemHealthCheck
}

// NewHealthManager creates a new health manager
func NewHealthManager(logger log.Logger) *HealthManager {
	return &HealthManager{
		checkers:      make(map[string]HealthChecker),
		results:       make(map[string]*HealthCheckResult),
		logger:        logger,
		checkInterval: 30 * time.Second,
		timeout:       10 * time.Second,
	}
}

// RegisterChecker registers a health checker for a component
func (hm *HealthManager) RegisterChecker(checker HealthChecker) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.checkers[checker.Name()] = checker
	hm.results[checker.Name()] = &HealthCheckResult{
		Component: checker.Name(),
		Status:    HealthStatusUnknown,
		Message:   "Not checked yet",
		LastCheck: time.Time{},
	}

	hm.logger.Info("Health checker registered", zap.String("component", checker.Name()))
}

// UnregisterChecker removes a health checker
func (hm *HealthManager) UnregisterChecker(name string) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	delete(hm.checkers, name)
	delete(hm.results, name)

	hm.logger.Info("Health checker unregistered", zap.String("component", name))
}

// PerformHealthCheck performs a health check on a specific component
func (hm *HealthManager) PerformHealthCheck(ctx context.Context, componentName string) (*HealthCheckResult, error) {
	hm.mutex.RLock()
	checker, exists := hm.checkers[componentName]
	hm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("health checker not found for component: %s", componentName)
	}

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, hm.timeout)
	defer cancel()

	// Perform the health check
	startTime := time.Now()
	result := checker.HealthCheck(checkCtx)
	responseTime := time.Since(startTime)

	// Update result with timing information
	result.LastCheck = startTime
	result.ResponseTime = responseTime

	// Store the result
	hm.mutex.Lock()
	hm.results[componentName] = result
	hm.mutex.Unlock()

	// Log the result
	hm.logger.Info("Health check completed",
		zap.String("component", componentName),
		zap.String("status", string(result.Status)),
		zap.String("message", result.Message),
		zap.Duration("response_time", responseTime),
	)

	return result, nil
}

// PerformAllHealthChecks performs health checks on all registered components
func (hm *HealthManager) PerformAllHealthChecks(ctx context.Context) map[string]*HealthCheckResult {
	hm.mutex.RLock()
	checkers := make(map[string]HealthChecker)
	for name, checker := range hm.checkers {
		checkers[name] = checker
	}
	hm.mutex.RUnlock()

	results := make(map[string]*HealthCheckResult)

	// Perform health checks concurrently
	var wg sync.WaitGroup
	resultsMutex := sync.Mutex{}

	for name, checker := range checkers {
		wg.Add(1)
		go func(componentName string, healthChecker HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, hm.timeout)
			defer cancel()

			startTime := time.Now()
			result := healthChecker.HealthCheck(checkCtx)
			responseTime := time.Since(startTime)

			result.LastCheck = startTime
			result.ResponseTime = responseTime

			resultsMutex.Lock()
			results[componentName] = result
			resultsMutex.Unlock()

			hm.logger.Debug("Health check completed",
				zap.String("component", componentName),
				zap.String("status", string(result.Status)),
				zap.Duration("response_time", responseTime),
			)
		}(name, checker)
	}

	wg.Wait()

	// Update stored results
	hm.mutex.Lock()
	for name, result := range results {
		hm.results[name] = result
	}
	hm.mutex.Unlock()

	return results
}

// PerformSystemHealthCheck performs system-level health checks
func (hm *HealthManager) PerformSystemHealthCheck(ctx context.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get GC stats
	gcStats := GCStats{
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
		LastGC:        time.Unix(0, int64(m.LastGC)),
		NextGC:        m.NextGC,
		PauseTotalNs:  m.PauseTotalNs,
	}

	// Copy recent pause times (up to last 256)
	pauseCount := len(m.PauseNs)
	if pauseCount > 256 {
		pauseCount = 256
	}
	gcStats.PauseNs = make([]uint64, pauseCount)
	copy(gcStats.PauseNs, m.PauseNs[:pauseCount])

	systemHealth := &SystemHealthCheck{
		Uptime: time.Since(hm.lastSystemCheck),
		MemoryUsage: MemoryStats{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
		},
		GoroutineCount: runtime.NumGoroutine(),
		GCStats:        gcStats,
	}

	hm.mutex.Lock()
	hm.systemHealth = systemHealth
	hm.lastSystemCheck = time.Now()
	hm.mutex.Unlock()

	hm.logger.Debug("System health check completed",
		zap.Uint64("alloc_bytes", m.Alloc),
		zap.Uint64("sys_bytes", m.Sys),
		zap.Int("goroutine_count", runtime.NumGoroutine()),
		zap.Uint32("num_gc", m.NumGC),
	)

	return nil
}

// GetHealthStatus returns the current health status of all components
func (hm *HealthManager) GetHealthStatus() map[string]*HealthCheckResult {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	// Create a copy to avoid race conditions
	results := make(map[string]*HealthCheckResult)
	for name, result := range hm.results {
		resultCopy := *result
		results[name] = &resultCopy
	}

	return results
}

// GetSystemHealth returns the current system health information
func (hm *HealthManager) GetSystemHealth() *SystemHealthCheck {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	if hm.systemHealth == nil {
		return nil
	}

	// Return a copy to avoid race conditions
	systemHealthCopy := *hm.systemHealth
	return &systemHealthCopy
}

// IsHealthy returns true if all components are healthy
func (hm *HealthManager) IsHealthy() bool {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	for _, result := range hm.results {
		if result.Status != HealthStatusHealthy {
			return false
		}
	}

	return true
}

// GetUnhealthyComponents returns a list of unhealthy components
func (hm *HealthManager) GetUnhealthyComponents() []string {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var unhealthy []string
	for name, result := range hm.results {
		if result.Status != HealthStatusHealthy {
			unhealthy = append(unhealthy, name)
		}
	}

	return unhealthy
}

// SetCheckInterval sets the health check interval
func (hm *HealthManager) SetCheckInterval(interval time.Duration) {
	hm.checkInterval = interval
	hm.logger.Info("Health check interval updated", zap.Duration("interval", interval))
}

// SetTimeout sets the health check timeout
func (hm *HealthManager) SetTimeout(timeout time.Duration) {
	hm.timeout = timeout
	hm.logger.Info("Health check timeout updated", zap.Duration("timeout", timeout))
}

// GetCheckInterval returns the current health check interval
func (hm *HealthManager) GetCheckInterval() time.Duration {
	return hm.checkInterval
}

// GetTimeout returns the current health check timeout
func (hm *HealthManager) GetTimeout() time.Duration {
	return hm.timeout
}

// GetRegisteredCheckers returns a list of registered checker names
func (hm *HealthManager) GetRegisteredCheckers() []string {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var names []string
	for name := range hm.checkers {
		names = append(names, name)
	}

	return names
}

// ClearResults clears all stored health check results
func (hm *HealthManager) ClearResults() {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	for name := range hm.results {
		hm.results[name] = &HealthCheckResult{
			Component: name,
			Status:    HealthStatusUnknown,
			Message:   "Not checked yet",
			LastCheck: time.Time{},
		}
	}

	hm.logger.Info("Health check results cleared")
}

// HealthSummary provides a summary of the overall health status
type HealthSummary struct {
	OverallStatus       HealthStatus                  `json:"overall_status"`
	TotalComponents     int                           `json:"total_components"`
	HealthyComponents   int                           `json:"healthy_components"`
	UnhealthyComponents []string                      `json:"unhealthy_components"`
	LastCheck           time.Time                     `json:"last_check"`
	SystemHealth        *SystemHealthCheck            `json:"system_health,omitempty"`
	ComponentResults    map[string]*HealthCheckResult `json:"component_results"`
}

// GetHealthSummary returns a comprehensive health summary
func (hm *HealthManager) GetHealthSummary() *HealthSummary {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	summary := &HealthSummary{
		TotalComponents:  len(hm.results),
		ComponentResults: make(map[string]*HealthCheckResult),
	}

	// Determine overall status and count healthy components
	healthyCount := 0
	var lastCheck time.Time

	for name, result := range hm.results {
		// Copy result to avoid race conditions
		resultCopy := *result
		summary.ComponentResults[name] = &resultCopy

		if result.Status == HealthStatusHealthy {
			healthyCount++
		} else {
			// Any non-healthy status (including Unknown and Unhealthy) is considered unhealthy
			summary.UnhealthyComponents = append(summary.UnhealthyComponents, name)
		}

		if result.LastCheck.After(lastCheck) {
			lastCheck = result.LastCheck
		}
	}

	summary.HealthyComponents = healthyCount
	summary.LastCheck = lastCheck

	// Determine overall status
	if healthyCount == summary.TotalComponents && summary.TotalComponents > 0 {
		summary.OverallStatus = HealthStatusHealthy
	} else if len(summary.UnhealthyComponents) > 0 {
		summary.OverallStatus = HealthStatusUnhealthy
	} else {
		summary.OverallStatus = HealthStatusUnknown
	}

	// Include system health if available
	if hm.systemHealth != nil {
		systemHealthCopy := *hm.systemHealth
		summary.SystemHealth = &systemHealthCopy
	}

	return summary
}
