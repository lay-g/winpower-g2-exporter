package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduler_WaitGroupIssue is a simple test to debug WaitGroup issue
func TestScheduler_WaitGroupIssue(t *testing.T) {
	config := &Config{
		CollectionInterval:      100 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
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

	// Create context that will be cancelled after 200ms
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start scheduler
	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for context to be cancelled
	time.Sleep(250 * time.Millisecond)

	// Check if scheduler automatically stopped due to context cancellation
	assert.False(t, scheduler.IsRunning(), "Scheduler should have stopped due to context cancellation")

	// Try to stop scheduler - this should not block
	stopStart := time.Now()
	err = scheduler.Stop()
	stopDuration := time.Since(stopStart)

	t.Logf("Stop took: %v", stopDuration)
	t.Logf("Stop error: %v", err)

	// Stop should succeed quickly
	assert.NoError(t, err, "Stop should succeed")
	assert.Less(t, stopDuration, 100*time.Millisecond, "Stop should complete quickly")
}
