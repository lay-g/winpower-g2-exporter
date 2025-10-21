package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultScheduler_StartNormalFlow tests Start() method normal startup flow
func TestDefaultScheduler_StartNormalFlow(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)
	require.NotNil(t, scheduler)

	// Initially scheduler should not be running
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running initially")

	// Create context with cancellation for cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	assert.NoError(t, err, "Start should not return error")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running after Start")

	// Wait a bit to ensure scheduler is actually running
	time.Sleep(100 * time.Millisecond)

	// Verify that scheduler is still running
	assert.True(t, scheduler.IsRunning(), "Scheduler should still be running")

	// Clean up by stopping the scheduler
	err = scheduler.Stop()
	assert.NoError(t, err, "Stop should not return error")
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after Stop")
}

// TestDefaultScheduler_StartDuplicateHandling tests Start() method duplicate startup handling
func TestDefaultScheduler_StartDuplicateHandling(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context with cancellation for cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler first time
	err = scheduler.Start(ctx)
	assert.NoError(t, err, "First Start should succeed")
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running after first Start")

	// Try to start scheduler again (should not error)
	err = scheduler.Start(ctx)
	assert.NoError(t, err, "Second Start should not error (should be no-op)")
	assert.True(t, scheduler.IsRunning(), "Scheduler should still be running")

	// Clean up
	err = scheduler.Stop()
	assert.NoError(t, err, "Stop should succeed")
}

// TestDefaultScheduler_StartConcurrentSafety tests Start() method concurrent startup safety
func TestDefaultScheduler_StartConcurrentSafety(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create context with cancellation for cleanup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Number of goroutines trying to start scheduler concurrently
	const numGoroutines = 10

	// Use WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Start multiple goroutines trying to start scheduler
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := scheduler.Start(ctx)
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

	// Verify that scheduler is running
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running after concurrent start attempts")

	// Clean up
	err = scheduler.Stop()
	assert.NoError(t, err, "Stop should succeed")
}

// TestDefaultScheduler_StartErrorHandling tests Start() method error handling
func TestDefaultScheduler_StartErrorHandling(t *testing.T) {
	// Test with cancelled context
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Start with cancelled context
	err = scheduler.Start(ctx)
	// This might or might not error depending on implementation
	// The important thing is that scheduler handles gracefully

	// Verify scheduler state is consistent
	if err == nil {
		// If start succeeded, scheduler should be running
		assert.True(t, scheduler.IsRunning(), "If start succeeded, scheduler should be running")
		// Clean up
		_ = scheduler.Stop()
	} else {
		// If start failed, scheduler should not be running
		assert.False(t, scheduler.IsRunning(), "If start failed, scheduler should not be running")
	}
}
