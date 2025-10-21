package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScheduler_ConfigBoundaryValues tests configuration boundary values (minimum intervals, timeouts, etc.)
func TestScheduler_ConfigBoundaryValues(t *testing.T) {
	testCases := []struct {
		name                        string
		collectionInterval          time.Duration
		gracefulShutdownTimeout     time.Duration
		expectConstructorError      bool
		expectedConstructorErrorMsg string
		shouldStartSuccessfully     bool
	}{
		{
			name:                    "reasonable interval for testing",
			collectionInterval:      50 * time.Millisecond,
			gracefulShutdownTimeout: 2 * time.Second,
			expectConstructorError:  false,
			shouldStartSuccessfully: true,
		},
		{
			name:                    "large interval",
			collectionInterval:      100 * time.Millisecond,
			gracefulShutdownTimeout: 2 * time.Second,
			expectConstructorError:  false,
			shouldStartSuccessfully: true,
		},
		{
			name:                    "minimum timeout",
			collectionInterval:      10 * time.Millisecond,
			gracefulShutdownTimeout: 1 * time.Second,
			expectConstructorError:  false,
			shouldStartSuccessfully: true,
		},
		{
			name:                        "zero collection interval",
			collectionInterval:          0,
			gracefulShutdownTimeout:     30 * time.Second,
			expectConstructorError:      true,
			expectedConstructorErrorMsg: "collection interval cannot be zero or negative",
		},
		{
			name:                        "negative collection interval",
			collectionInterval:          -1 * time.Second,
			gracefulShutdownTimeout:     30 * time.Second,
			expectConstructorError:      true,
			expectedConstructorErrorMsg: "collection interval cannot be zero or negative",
		},
		{
			name:                        "zero graceful shutdown timeout",
			collectionInterval:          5 * time.Second,
			gracefulShutdownTimeout:     0,
			expectConstructorError:      true,
			expectedConstructorErrorMsg: "graceful shutdown timeout cannot be zero or negative",
		},
		{
			name:                        "negative graceful shutdown timeout",
			collectionInterval:          5 * time.Second,
			gracefulShutdownTimeout:     -30 * time.Second,
			expectConstructorError:      true,
			expectedConstructorErrorMsg: "graceful shutdown timeout cannot be zero or negative",
		},
		{
			name:                    "timeout shorter than interval",
			collectionInterval:      1 * time.Second,
			gracefulShutdownTimeout: 500 * time.Millisecond,
			expectConstructorError:  false,
			shouldStartSuccessfully: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create config
			config := &Config{
				CollectionInterval:      tc.collectionInterval,
				GracefulShutdownTimeout: tc.gracefulShutdownTimeout,
			}

			// Create mock dependencies
			client := &MockWinPowerClient{}
			logger := &MockLogger{}

			// Try to create scheduler
			scheduler, err := NewDefaultScheduler(config, client, logger)

			if tc.expectConstructorError {
				assert.Error(t, err, "Expected constructor error")
				assert.Nil(t, scheduler, "Expected nil scheduler on constructor error")
				if tc.expectedConstructorErrorMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedConstructorErrorMsg, "Error message should contain expected text")
				}
				return
			}

			// Constructor should succeed
			require.NoError(t, err, "Constructor should not fail")
			require.NotNil(t, scheduler, "Scheduler should be created")

			// Test starting if expected to succeed
			if tc.shouldStartSuccessfully {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				// Create context-aware mock client
				contextAwareClient := &MockWinPowerClient{
					collectDataFunc: func(ctx context.Context) error {
						// Check if context is cancelled
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

				// Replace the client with context-aware one
				scheduler, err = NewDefaultScheduler(config, contextAwareClient, logger)
				require.NoError(t, err)

				err = scheduler.Start(ctx)
				assert.NoError(t, err, "Scheduler should start successfully")

				// Wait a bit then stop
				time.Sleep(tc.collectionInterval / 2)
				err = scheduler.Stop()
				assert.NoError(t, err, "Scheduler should stop successfully")
			}
		})
	}
}

// TestScheduler_ExceptionScenarios tests various exception scenarios handling
func TestScheduler_ExceptionScenarios(t *testing.T) {
	t.Run("context already cancelled at start", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 1 * time.Second,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		// Create already cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Try to start with cancelled context
		err = scheduler.Start(ctx)
		assert.NoError(t, err, "Starting with cancelled context should not error immediately")

		// Wait a bit for the context cancellation to take effect
		time.Sleep(20 * time.Millisecond)

		// Scheduler should not be running
		assert.False(t, scheduler.IsRunning(), "Scheduler should not be running with cancelled context")
	})

	t.Run("multiple rapid start/stop cycles", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      10 * time.Millisecond,
			GracefulShutdownTimeout: 100 * time.Millisecond,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Perform multiple start/stop cycles rapidly
		for i := 0; i < 10; i++ {
			err = scheduler.Start(ctx)
			assert.NoError(t, err, "Start should succeed in cycle %d", i)

			// Brief pause
			time.Sleep(5 * time.Millisecond)

			err = scheduler.Stop()
			assert.NoError(t, err, "Stop should succeed in cycle %d", i)

			assert.False(t, scheduler.IsRunning(), "Scheduler should not be running after cycle %d", i)
		}
	})

	t.Run("client panic during collection", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      10 * time.Millisecond,
			GracefulShutdownTimeout: 100 * time.Millisecond,
		}

		// Create client that panics during collection
		client := &MockWinPowerClient{
			collectDataFunc: func(ctx context.Context) error {
				panic("simulated client panic")
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

		// Start scheduler (should not crash despite client panic)
		assert.NotPanics(t, func() {
			err = scheduler.Start(ctx)
			assert.NoError(t, err, "Start should succeed even if client might panic")
		})

		// Wait for collection attempts
		time.Sleep(50 * time.Millisecond)

		// Stop scheduler
		err = scheduler.Stop()
		assert.NoError(t, err, "Stop should succeed")
	})

	t.Run("nil context", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      50 * time.Millisecond,
			GracefulShutdownTimeout: 1 * time.Second,
		}

		client := &MockWinPowerClient{}
		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		// Try to start with TODO context (should handle gracefully)
		// Note: The behavior depends on implementation - we expect graceful handling
		assert.NotPanics(t, func() {
			err = scheduler.Start(context.TODO())
			// This should succeed gracefully
			require.NoError(t, err)
		})
	})
}

// TestScheduler_ResourceLimitationScenarios tests resource limitation scenarios
func TestScheduler_ResourceLimitationScenarios(t *testing.T) {
	t.Run("high frequency collection", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      5 * time.Millisecond, // High frequency but reasonable
			GracefulShutdownTimeout: 200 * time.Millisecond,
		}

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

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)

		// Wait for high frequency collection
		time.Sleep(50 * time.Millisecond)

		err = scheduler.Stop()
		assert.NoError(t, err)

		// Should have made several collection calls
		assert.Greater(t, collectCallCount, 5, "Should have made several collection calls at high frequency")
	})

	t.Run("slow collection operations", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      100 * time.Millisecond,
			GracefulShutdownTimeout: 1 * time.Second, // Increased timeout for slow operations
		}

		// Create client with slow operations
		client := &MockWinPowerClient{
			collectDataFunc: func(ctx context.Context) error {
				// Simulate slow collection
				time.Sleep(50 * time.Millisecond)
				return nil
			},
			isConnectedFunc: func() bool {
				return true
			},
		}

		logger := &MockLogger{}

		scheduler, err := NewDefaultScheduler(config, client, logger)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		err = scheduler.Start(ctx)
		require.NoError(t, err)

		// Wait for slow operations (reduced to avoid multiple collection cycles)
		time.Sleep(150 * time.Millisecond)

		err = scheduler.Stop()
		assert.NoError(t, err)

		// Scheduler should handle slow operations gracefully
		assert.False(t, scheduler.IsRunning(), "Scheduler should stop cleanly even with slow operations")
	})

	t.Run("concurrent scheduler instances", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      20 * time.Millisecond,
			GracefulShutdownTimeout: 200 * time.Millisecond,
		}

		// Create multiple scheduler instances
		numSchedulers := 5
		schedulers := make([]*DefaultScheduler, 0, numSchedulers)

		for i := 0; i < numSchedulers; i++ {
			client := &MockWinPowerClient{
				collectDataFunc: func(ctx context.Context) error {
					return nil
				},
				isConnectedFunc: func() bool {
					return true
				},
			}

			logger := &MockLogger{}

			scheduler, err := NewDefaultScheduler(config, client, logger)
			require.NoError(t, err)
			schedulers = append(schedulers, scheduler)
		}

		ctx := context.Background()

		// Start all schedulers concurrently
		for i, scheduler := range schedulers {
			err := scheduler.Start(ctx)
			assert.NoError(t, err, "Scheduler %d should start", i)
		}

		// Wait a bit
		time.Sleep(50 * time.Millisecond)

		// Stop all schedulers
		for i, scheduler := range schedulers {
			err := scheduler.Stop()
			assert.NoError(t, err, "Scheduler %d should stop", i)
		}

		// All schedulers should be stopped
		for i, scheduler := range schedulers {
			assert.False(t, scheduler.IsRunning(), "Scheduler %d should not be running", i)
		}
	})
}
