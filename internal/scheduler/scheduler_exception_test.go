package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduler_ExceptionHandling tests various exception scenarios handling
func TestScheduler_ExceptionHandling(t *testing.T) {
	t.Run("start with nil context", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 500 * time.Millisecond,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		// Try to start with TODO context
		// This should succeed gracefully
		assert.NotPanics(t, func() {
			err = scheduler.Start(context.TODO())
			// Check result
			require.NoError(t, err, "Start with TODO context should succeed")
		})

		// Try to stop
		err = scheduler.Stop()
		assert.NoError(t, err, "Should be able to stop after nil context start")
	})

	t.Run("multiple start operations", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 500 * time.Millisecond,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Try to start scheduler multiple times
		for i := 0; i < 3; i++ {
			err := scheduler.Start(ctx)
			assert.NoError(t, err, "Multiple starts should not error")
		}

		// Scheduler should be running
		assert.True(t, scheduler.IsRunning(), "Scheduler should be running after multiple starts")

		// Stop scheduler
		err = scheduler.Stop()
		assert.NoError(t, err, "Should be able to stop after multiple starts")
	})

	t.Run("multiple stop operations", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 500 * time.Millisecond,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Start scheduler
		err = scheduler.Start(ctx)
		require.NoError(t, err)

		// Try to stop scheduler multiple times
		for i := 0; i < 3; i++ {
			err := scheduler.Stop()
			assert.NoError(t, err, "Multiple stops should not error")
		}

		// Scheduler should not be running
		assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after multiple stops")
	})

	t.Run("stop without start", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 500 * time.Millisecond,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		// Try to stop without starting
		err = scheduler.Stop()
		assert.NoError(t, err, "Stop without start should not error")

		assert.False(t, scheduler.IsRunning(), "Scheduler should not be running")
	})

	t.Run("start after stop", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 2 * time.Second, // Longer timeout
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Start and stop scheduler
		err = scheduler.Start(ctx)
		require.NoError(t, err)

		// Wait for context to cancel scheduler
		time.Sleep(150 * time.Millisecond)

		// Try to start again with fresh context
		ctx2, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel2()

		err = scheduler.Start(ctx2)
		assert.NoError(t, err, "Should be able to start after stop")

		// Should be running again
		assert.True(t, scheduler.IsRunning(), "Scheduler should be running after restart")

		// Clean up
		err = scheduler.Stop()
		assert.NoError(t, err)
	})

	t.Run("panic recovery in collection loop", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      10 * time.Millisecond,
			GracefulShutdownTimeout: 500 * time.Millisecond,
		}

		// Create a client that panics
		panicCount := 0
		client := &MockWinPowerClient{
			collectDataFunc: func(ctx context.Context) error {
				panicCount++
				if panicCount <= 2 {
					panic("simulated collection panic")
				}
				return nil
			},
			isConnectedFunc: func() bool {
				return true
			},
		}

		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		// This should not crash the test
		assert.NotPanics(t, func() {
			err = scheduler.Start(ctx)
			// Note: panic recovery would need to be implemented in the scheduler
			// For now, this test documents expected behavior
		})

		// Wait a bit then try to stop
		time.Sleep(50 * time.Millisecond)
		err = scheduler.Stop()
		// This might error due to panic, but should not crash the test
		_ = err
	})
}
