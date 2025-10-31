package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
)

// mockCollector implements scheduler.CollectorInterface for integration testing
type mockCollector struct {
	mu                 sync.Mutex
	callCount          int
	lastCollectionTime time.Time
	shouldFail         bool
	failureCount       int
}

func (m *mockCollector) CollectDeviceData(ctx context.Context) (*scheduler.CollectionResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callCount++
	m.lastCollectionTime = time.Now()

	// Simulate failure for testing error recovery
	if m.shouldFail && m.failureCount > 0 {
		m.failureCount--
		return nil, context.DeadlineExceeded
	}

	return &scheduler.CollectionResult{
		Success:     true,
		DeviceCount: 3,
	}, nil
}

func (m *mockCollector) getCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *mockCollector) getLastCollectionTime() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastCollectionTime
}

// mockLogger implements scheduler.Logger for integration testing
type mockLogger struct {
	mu    sync.Mutex
	logs  []string
	debug bool
}

func (l *mockLogger) Info(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, "INFO: "+msg)
}

func (l *mockLogger) Error(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, "ERROR: "+msg)
}

func (l *mockLogger) Warn(msg string, fields ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, "WARN: "+msg)
}

func (l *mockLogger) Debug(msg string, fields ...interface{}) {
	if !l.debug {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, "DEBUG: "+msg)
}

func (l *mockLogger) hasLog(prefix string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, log := range l.logs {
		if len(log) >= len(prefix) && log[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// TestSchedulerIntegration tests the complete lifecycle of the scheduler
func TestSchedulerIntegration(t *testing.T) {
	t.Run("complete lifecycle with multiple collections", func(t *testing.T) {
		config := &scheduler.Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &mockCollector{}
		logger := &mockLogger{}

		// Create scheduler
		sched, err := scheduler.NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		// Start scheduler
		ctx := context.Background()
		if err := sched.Start(ctx); err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}

		// Verify scheduler is running
		if !sched.IsRunning() {
			t.Error("Scheduler should be running after Start()")
		}

		// Wait for multiple collection cycles
		time.Sleep(3500 * time.Millisecond)

		// Verify collections occurred
		callCount := collector.getCallCount()
		if callCount < 3 {
			t.Errorf("Expected at least 3 collections, got %d", callCount)
		}

		// Verify last collection time is recent
		lastTime := collector.getLastCollectionTime()
		if time.Since(lastTime) > 2*time.Second {
			t.Errorf("Last collection time is too old: %v", lastTime)
		}

		// Stop scheduler
		if err := sched.Stop(ctx); err != nil {
			t.Fatalf("Failed to stop scheduler: %v", err)
		}

		// Verify scheduler is not running
		if sched.IsRunning() {
			t.Error("Scheduler should not be running after Stop()")
		}

		// Verify logs
		if !logger.hasLog("INFO: scheduler started") {
			t.Error("Expected 'scheduler started' log")
		}
		if !logger.hasLog("INFO: collection completed") {
			t.Error("Expected 'collection completed' log")
		}
		if !logger.hasLog("INFO: scheduler stopped") {
			t.Error("Expected 'scheduler stopped' log")
		}
	})

	t.Run("error recovery and continued operation", func(t *testing.T) {
		config := &scheduler.Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &mockCollector{
			shouldFail:   true,
			failureCount: 2, // First 2 calls will fail
		}
		logger := &mockLogger{}

		sched, err := scheduler.NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		ctx := context.Background()
		if err := sched.Start(ctx); err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}

		// Wait for failures and recovery
		time.Sleep(4500 * time.Millisecond)

		// Verify scheduler recovered and continued
		callCount := collector.getCallCount()
		if callCount < 4 {
			t.Errorf("Expected at least 4 collection attempts (including failures), got %d", callCount)
		}

		// Verify error was logged
		if !logger.hasLog("ERROR: collection failed") {
			t.Error("Expected 'collection failed' log")
		}

		// Verify successful collection after recovery
		if !logger.hasLog("INFO: collection completed") {
			t.Error("Expected 'collection completed' log after recovery")
		}

		// Clean up
		_ = sched.Stop(ctx)
	})

	t.Run("rapid start and stop", func(t *testing.T) {
		config := scheduler.DefaultConfig()
		collector := &mockCollector{}
		logger := &mockLogger{}

		sched, err := scheduler.NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		ctx := context.Background()

		// Start
		if err := sched.Start(ctx); err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}

		// Immediately stop
		if err := sched.Stop(ctx); err != nil {
			t.Fatalf("Failed to stop scheduler: %v", err)
		}

		// Verify clean shutdown
		if sched.IsRunning() {
			t.Error("Scheduler should not be running after rapid stop")
		}
	})

	t.Run("multiple start/stop cycles", func(t *testing.T) {
		config := &scheduler.Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}
		collector := &mockCollector{}
		logger := &mockLogger{}

		sched, err := scheduler.NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		ctx := context.Background()

		for i := 0; i < 3; i++ {
			// Start
			if err := sched.Start(ctx); err != nil {
				t.Fatalf("Cycle %d: Failed to start scheduler: %v", i, err)
			}

			// Run for a bit
			time.Sleep(1500 * time.Millisecond)

			// Stop
			if err := sched.Stop(ctx); err != nil {
				t.Fatalf("Cycle %d: Failed to stop scheduler: %v", i, err)
			}

			// Verify not running
			if sched.IsRunning() {
				t.Errorf("Cycle %d: Scheduler should not be running", i)
			}
		}

		// Verify multiple collections occurred across all cycles
		callCount := collector.getCallCount()
		if callCount < 3 {
			t.Errorf("Expected at least 3 collections across cycles, got %d", callCount)
		}
	})
}

// TestSchedulerPerformance tests the scheduler's performance characteristics
func TestSchedulerPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	t.Run("interval precision", func(t *testing.T) {
		config := &scheduler.Config{
			CollectionInterval:      1 * time.Second,
			GracefulShutdownTimeout: 5 * time.Second,
		}

		collector := &mockCollector{}
		logger := &mockLogger{}

		sched, err := scheduler.NewDefaultScheduler(config, collector, logger)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		ctx := context.Background()
		if err := sched.Start(ctx); err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}

		// Measure interval precision over 10 seconds
		startTime := time.Now()
		time.Sleep(10 * time.Second)
		duration := time.Since(startTime)

		callCount := collector.getCallCount()
		expectedCalls := int(duration.Seconds())

		// Allow some tolerance for timing variations
		if callCount < expectedCalls-1 || callCount > expectedCalls+1 {
			t.Errorf("Interval precision issue: expected ~%d calls, got %d in %v",
				expectedCalls, callCount, duration)
		}

		_ = sched.Stop(ctx)
	})
}
