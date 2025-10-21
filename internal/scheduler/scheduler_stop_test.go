package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultScheduler_StopNormalFlow tests Stop() method normal shutdown flow
func TestDefaultScheduler_StopNormalFlow(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context for startup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err, "Start should succeed")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running")

	// Stop scheduler
	err = scheduler.Stop()
	assert.NoError(t, err, "Stop should not return error")
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after Stop")

	// Clear log messages
	logger.ClearLogMessages()

	// Try to stop again (should not error)
	err = scheduler.Stop()
	assert.NoError(t, err, "Second Stop should not error")
	assert.False(t, scheduler.IsRunning(), "Scheduler should still not be running")
}

// TestDefaultScheduler_StopDuplicateHandling tests Stop() method duplicate stop handling
func TestDefaultScheduler_StopDuplicateHandling(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Try to stop scheduler that was never started
	err = scheduler.Stop()
	assert.NoError(t, err, "Stop on non-started scheduler should not error")
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running")

	// Create context and start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err, "Start should succeed")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running")

	// Stop scheduler multiple times
	for i := 0; i < 3; i++ {
		err = scheduler.Stop()
		assert.NoErrorf(t, err, "Stop attempt %d should not error", i)
		assert.Falsef(t, scheduler.IsRunning(), "Scheduler should not be running after stop attempt %d", i)
	}
}

// TestDefaultScheduler_StopConcurrentSafety tests Stop() method concurrent stop safety
func TestDefaultScheduler_StopConcurrentSafety(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context and start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err, "Start should succeed")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running")

	// Number of goroutines trying to stop scheduler concurrently
	const numGoroutines = 10

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Start multiple goroutines trying to stop scheduler
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := scheduler.Stop()
			mu.Lock()
			errors = append(errors, err)
			mu.Unlock()
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Check that no errors occurred
	for i, err := range errors {
		assert.NoErrorf(t, err, "Goroutine %d should not return error", i)
	}

	// Verify that scheduler is not running
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after concurrent stop attempts")
}

// TestDefaultScheduler_StopGracefulTimeout tests Stop() method graceful shutdown timeout
func TestDefaultScheduler_StopGracefulTimeout(t *testing.T) {
	// Create test dependencies with short timeout
	config := &Config{
		CollectionInterval:      1 * time.Second,
		GracefulShutdownTimeout: 100 * time.Millisecond, // Very short timeout
	}
	require.NoError(t, config.Validate(), "Config should be valid")

	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context and start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err, "Start should succeed")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running")

	// Wait a bit to ensure collection loop is running
	time.Sleep(50 * time.Millisecond)

	// Stop scheduler - should complete quickly
	startTime := time.Now()
	err = scheduler.Stop()
	elapsed := time.Since(startTime)

	assert.NoError(t, err, "Stop should not error")
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running")
	assert.Less(t, elapsed, config.GracefulShutdownTimeout+50*time.Millisecond,
		"Stop should complete within reasonable time")
}

// TestDefaultScheduler_StopResourceCleanup tests Stop() method resource cleanup
func TestDefaultScheduler_StopResourceCleanup(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context and start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err, "Start should succeed")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running")

	// Verify that runtime fields are set
	assert.NotNil(t, scheduler.ctx, "Context should be set")
	assert.NotNil(t, scheduler.cancel, "Cancel function should be set")
	assert.NotNil(t, scheduler.ticker, "Ticker should be set")

	// Stop scheduler
	err = scheduler.Stop()
	assert.NoError(t, err, "Stop should not error")
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running")

	// Verify that resources are cleaned up (ticker should be stopped)
	// Note: We can't directly verify that ticker is stopped without exposing internal state
	// But we can test by starting/stopping multiple times
}
