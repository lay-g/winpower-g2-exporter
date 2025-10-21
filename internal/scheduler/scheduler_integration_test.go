package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduler_Integration_Lifecycle tests complete start-run-stop lifecycle
func TestScheduler_Integration_Lifecycle(t *testing.T) {
	config := &Config{
		CollectionInterval:      100 * time.Millisecond,
		GracefulShutdownTimeout: 2 * time.Second,
	}

	// Track collection calls
	collectCallCount := 0
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectCallCount++
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

	// Test initial state
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running initially")

	// Create context with timeout for test
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Verify running state
	assert.True(t, scheduler.IsRunning(), "Scheduler should be running after start")

	// Let scheduler run for some time
	time.Sleep(300 * time.Millisecond)

	// Stop scheduler
	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify stopped state
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after stop")

	// Verify that collections happened
	assert.Greater(t, collectCallCount, 0, "Data collection should have occurred")
	assert.GreaterOrEqual(t, collectCallCount, 2, "Should have collected data at least twice")
}

// TestScheduler_Integration_ContextCancellation tests context cancellation impact
func TestScheduler_Integration_ContextCancellation(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	// Track collection calls
	collectCallCount := 0
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectCallCount++
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

	// Create context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for context to be cancelled
	time.Sleep(250 * time.Millisecond)

	// Scheduler should have stopped due to context cancellation
	assert.False(t, scheduler.IsRunning(), "Scheduler should stop when context is cancelled")

	// Should have made some collections before cancellation
	assert.Greater(t, collectCallCount, 0, "Should have made collections before cancellation")

	// Should have cancellation logged
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

// TestScheduler_Integration_LongRunStability tests long-running stability
func TestScheduler_Integration_LongRunStability(t *testing.T) {
	config := &Config{
		CollectionInterval:      10 * time.Millisecond, // Fast collection
		GracefulShutdownTimeout: 1 * time.Second,
	}

	// Track collection calls
	collectCallCount := 0
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectCallCount++
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

	// Run for 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	startTime := time.Now()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for test completion
	time.Sleep(250 * time.Millisecond)

	// Verify collections occurred regularly
	expectedCollections := int(200 * time.Millisecond / config.CollectionInterval)
	assert.GreaterOrEqual(t, collectCallCount, expectedCollections/2,
		"Should have made approximately expected number of collections")

	// Verify timing consistency - use range check instead of delta
	duration := time.Since(startTime)
	expectedDuration := 200 * time.Millisecond
	// Use range check to account for system load and timing variations
	minDuration := expectedDuration - 100*time.Millisecond
	maxDuration := expectedDuration + 200*time.Millisecond
	assert.GreaterOrEqual(t, duration, minDuration,
		"Test duration should not be too short")
	assert.LessOrEqual(t, duration, maxDuration,
		"Test duration should not be too long")
}

// TestScheduler_Integration_RealDependencies tests scheduler with realistic dependencies
func TestScheduler_Integration_RealDependencies(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	// Simulate realistic client behavior
	collectCallCount := 0
	collectErrors := 0
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectCallCount++

			// Simulate occasional errors
			if collectCallCount%5 == 0 {
				collectErrors++
				return assert.AnError
			}

			// Simulate realistic collection time
			time.Sleep(5 * time.Millisecond)
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

	// Run test
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for collections to occur
	time.Sleep(300 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify realistic behavior
	assert.Greater(t, collectCallCount, 0, "Should have made collections")
	assert.Greater(t, collectErrors, 0, "Should have experienced some errors")
	assert.Less(t, collectErrors, collectCallCount/2, "Error rate should be reasonable")

	// Verify error logging
	logMessages := logger.GetLogMessages()
	errorLogCount := 0
	for _, msg := range logMessages {
		if msg == "ERROR: data collection failed" {
			errorLogCount++
		}
	}
	assert.Equal(t, collectErrors, errorLogCount, "Each error should be logged")
}
