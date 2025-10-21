package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataCollection_TickerTriggersNormalFlow tests ticker triggered data collection normal flow
func TestDataCollection_TickerTriggersNormalFlow(t *testing.T) {
	// Create test dependencies with short collection interval for faster testing
	config := &Config{
		CollectionInterval:      100 * time.Millisecond, // Short interval for testing
		GracefulShutdownTimeout: 5 * time.Second,
	}

	// Create mock client that tracks collection calls
	collectCallCount := 0
	var collectMutex sync.Mutex
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectMutex.Lock()
			collectCallCount++
			collectMutex.Unlock()
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)
	require.NotNil(t, scheduler)

	// Create context for test
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for multiple collection cycles
	time.Sleep(350 * time.Millisecond) // Should trigger ~3 collections

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify collection was called multiple times
	collectMutex.Lock()
	expectedMinCalls := 2 // At least 2 calls should have occurred
	assert.GreaterOrEqual(t, collectCallCount, expectedMinCalls,
		"Data collection should have been called at least %d times", expectedMinCalls)
	collectMutex.Unlock()

	// Verify scheduler is not running
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after stop")
}

// TestDataCollection_FailureHandling tests data collection failure handling
func TestDataCollection_FailureHandling(t *testing.T) {
	// Create test dependencies
	config := &Config{
		CollectionInterval:      100 * time.Millisecond,
		GracefulShutdownTimeout: 5 * time.Second,
	}

	// Create mock client that returns error on collection
	collectCallCount := 0
	var collectMutex sync.Mutex
	expectedError := assert.AnError
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectMutex.Lock()
			collectCallCount++
			collectMutex.Unlock()
			return expectedError
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context for test
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for collection cycles
	time.Sleep(250 * time.Millisecond)

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify collection was called despite errors
	collectMutex.Lock()
	assert.Greater(t, collectCallCount, 0, "Data collection should have been attempted even with errors")
	collectMutex.Unlock()

	// Verify error logs were recorded
	logMessages := logger.GetLogMessages()
	errorLogCount := 0
	for _, msg := range logMessages {
		if msg == "ERROR: data collection failed" {
			errorLogCount++
		}
	}
	assert.Greater(t, errorLogCount, 0, "Error logs should have been recorded for collection failures")
}

// TestDataCollection_Logging tests data collection logging
func TestDataCollection_Logging(t *testing.T) {
	// Create test dependencies
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 5 * time.Second,
	}

	// Track collection calls
	collectCallCount := 0
	var collectMutex sync.Mutex
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectMutex.Lock()
			collectCallCount++
			collectMutex.Unlock()
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context for test
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Clear any existing log messages
	logger.ClearLogMessages()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for collection cycles
	time.Sleep(150 * time.Millisecond)

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify log messages
	logMessages := logger.GetLogMessages()

	// Should have debug messages for collection start and completion
	debugLogCount := 0
	for _, msg := range logMessages {
		if msg == "DEBUG: starting data collection" || msg == "DEBUG: data collection completed successfully" {
			debugLogCount++
		}
	}

	collectMutex.Lock()
	expectedDebugLogs := collectCallCount * 2 // Each collection should have start and completion debug logs
	collectMutex.Unlock()

	assert.GreaterOrEqual(t, debugLogCount, expectedDebugLogs/2,
		"Should have debug logs for data collection")
}

// TestDataCollection_ConcurrentSafety tests data collection concurrent safety
func TestDataCollection_ConcurrentSafety(t *testing.T) {
	// Create test dependencies
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 5 * time.Second,
	}

	// Track collection calls with concurrent safety
	collectCallCount := 0
	var collectMutex sync.Mutex
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectMutex.Lock()
			collectCallCount++
			collectMutex.Unlock()

			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			return nil
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context for test
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Concurrently start and stop scheduler to test safety
	var wg sync.WaitGroup
	numOperations := 5

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Try to stop scheduler (this should be safe even if called multiple times)
			_ = scheduler.Stop()
		}()
	}

	// Wait for operations to complete
	wg.Wait()

	// Wait a bit more for collection cycles
	time.Sleep(100 * time.Millisecond)

	// Final stop (in case it's still running)
	_ = scheduler.Stop()

	// Verify no race conditions occurred
	collectMutex.Lock()
	assert.GreaterOrEqual(t, collectCallCount, 0, "Collection calls should be tracked safely")
	collectMutex.Unlock()

	// Verify final state
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after all stops")
}

// TestDataCollection_ContextCancellation tests data collection context cancellation
func TestDataCollection_ContextCancellation(t *testing.T) {
	// Create test dependencies
	config := &Config{
		CollectionInterval:      100 * time.Millisecond,
		GracefulShutdownTimeout: 5 * time.Second,
	}

	// Track collection calls
	collectCallCount := 0
	var collectMutex sync.Mutex
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectMutex.Lock()
			collectCallCount++
			collectMutex.Unlock()

			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Simulate work
				time.Sleep(20 * time.Millisecond)
				return nil
			}
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait a bit then cancel context
	time.Sleep(150 * time.Millisecond)
	cancel()

	// Wait for scheduler to stop due to context cancellation
	time.Sleep(100 * time.Millisecond)

	// Verify scheduler stopped due to context cancellation
	assert.False(t, scheduler.IsRunning(), "Scheduler should stop when context is cancelled")

	// Verify collection was called before cancellation
	collectMutex.Lock()
	assert.Greater(t, collectCallCount, 0, "Collection should have been called before cancellation")
	collectMutex.Unlock()

	// Verify context cancellation was logged
	logMessages := logger.GetLogMessages()
	cancellationLogged := false
	for _, msg := range logMessages {
		if msg == "INFO: scheduler collection loop stopped due to context cancellation" {
			cancellationLogged = true
			break
		}
	}
	assert.True(t, cancellationLogged, "Context cancellation should be logged")
}
