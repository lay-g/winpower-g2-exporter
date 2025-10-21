package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduler_StopDebugging is a debugging test to understand why stop is timing out
func TestScheduler_StopDebugging(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond, // Reasonable interval
		GracefulShutdownTimeout: 1 * time.Second,       // Longer timeout
	}

	// Create context-aware mock client
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			// Check if context is cancelled before doing work
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Simulate brief work
				time.Sleep(1 * time.Millisecond)
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

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait a bit for some collection cycles
	time.Sleep(100 * time.Millisecond)

	// Try to stop scheduler
	logger.ClearLogMessages() // Clear any existing messages
	stopStart := time.Now()
	err = scheduler.Stop()
	stopDuration := time.Since(stopStart)

	t.Logf("Stop took: %v", stopDuration)
	t.Logf("Stop error: %v", err)

	// Print log messages for debugging
	logMessages := logger.GetLogMessages()
	t.Logf("Log messages (%d):", len(logMessages))
	for i, msg := range logMessages {
		t.Logf("  [%d] %s", i, msg)
	}

	// Check if stop succeeded within reasonable time
	assert.NoError(t, err, "Stop should succeed without timeout")
	assert.Less(t, stopDuration, 500*time.Millisecond, "Stop should complete quickly")
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after stop")
}
