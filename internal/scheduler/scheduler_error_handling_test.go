package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorHandling_Categorization tests the error categorization functionality
func TestErrorHandling_Categorization(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "unknown",
		},
		{
			name:     "network connection refused",
			err:      errors.New("connection refused"),
			expected: "network_error",
		},
		{
			name:     "network timeout",
			err:      errors.New("read timeout"),
			expected: "network_error",
		},
		{
			name:     "authentication unauthorized",
			err:      errors.New("401 unauthorized"),
			expected: "auth_error",
		},
		{
			name:     "authentication forbidden",
			err:      errors.New("403 forbidden"),
			expected: "auth_error",
		},
		{
			name:     "context cancelled",
			err:      context.Canceled,
			expected: "context_cancelled",
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: "timeout_error",
		},
		{
			name:     "JSON parse error",
			err:      errors.New("invalid JSON format"),
			expected: "data_error",
		},
		{
			name:     "rate limit exceeded",
			err:      errors.New("rate limit exceeded"),
			expected: "rate_limit_error",
		},
		{
			name:     "unknown error",
			err:      errors.New("some unknown error"),
			expected: "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineErrorType(tt.err)
			assert.Equal(t, tt.expected, result, "Error categorization should match expected")
		})
	}
}

// TestErrorHandling_CollectionID tests the collection ID generation
func TestErrorHandling_CollectionID(t *testing.T) {
	// Test that collection IDs are unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateCollectionID()
		assert.False(t, ids[id], "Collection ID should be unique: %s", id)
		ids[id] = true
		assert.Contains(t, id, "collection-", "Collection ID should have proper prefix")
	}
}

// TestErrorHandling_DataCollectionWithDifferentErrors tests data collection with various error types
func TestErrorHandling_DataCollectionWithDifferentErrors(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	tests := []struct {
		name             string
		collectError     error
		expectedLogCount int
	}{
		{
			name:             "network error",
			collectError:     errors.New("connection refused"),
			expectedLogCount: 2, // One error log + one warning from handleCollectionError
		},
		{
			name:             "authentication error",
			collectError:     errors.New("401 unauthorized"),
			expectedLogCount: 2, // One error log + one error from handleCollectionError
		},
		{
			name:             "timeout error",
			collectError:     context.DeadlineExceeded,
			expectedLogCount: 2, // One error log + one warning from handleCollectionError
		},
		{
			name:             "unknown error",
			collectError:     errors.New("random error"),
			expectedLogCount: 2, // One error log + one error from handleCollectionError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectCallCount := 0
			client := &MockWinPowerClient{
				collectDataFunc: func(ctx context.Context) error {
					collectCallCount++
					return tt.collectError
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

			// Clear existing log messages
			logger.ClearLogMessages()

			err = scheduler.Start(ctx)
			require.NoError(t, err)

			// Wait for collection to occur
			time.Sleep(100 * time.Millisecond)

			err = scheduler.Stop()
			require.NoError(t, err)

			// Verify collection was attempted
			assert.Greater(t, collectCallCount, 0, "Collection should have been attempted")

			// Verify error logs were recorded
			logMessages := logger.GetLogMessages()
			errorRelatedLogs := 0
			for _, msg := range logMessages {
				if msg == "ERROR: data collection failed" ||
					msg == "WARN: network connectivity issue detected" ||
					msg == "ERROR: authentication failure detected" ||
					msg == "WARN: data collection timeout" ||
					msg == "ERROR: data parsing or format error" ||
					msg == "WARN: rate limit exceeded" ||
					msg == "ERROR: unexpected error during data collection" {
					errorRelatedLogs++
				}
			}
			assert.GreaterOrEqual(t, errorRelatedLogs, 1, "Should have logged error-related messages")
		})
	}
}

// TestErrorHandling_PanicRecovery tests panic recovery during data collection
func TestErrorHandling_PanicRecovery(t *testing.T) {
	config := &Config{
		CollectionInterval:      50 * time.Millisecond,
		GracefulShutdownTimeout: 1 * time.Second,
	}

	panicTriggered := false
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			if !panicTriggered {
				panicTriggered = true
				panic("test panic during collection")
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

	// Clear existing log messages
	logger.ClearLogMessages()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for panic to occur and be recovered
	time.Sleep(100 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify panic was logged
	logMessages := logger.GetLogMessages()
	panicLogged := false
	for _, msg := range logMessages {
		if msg == "ERROR: panic recovered in data collection" {
			panicLogged = true
			break
		}
	}
	assert.True(t, panicLogged, "Panic should have been logged")
	assert.True(t, panicTriggered, "Panic should have been triggered")
}

// TestErrorHandling_ContextCancellationDuringCollection tests context cancellation during collection
func TestErrorHandling_ContextCancellationDuringCollection(t *testing.T) {
	config := &Config{
		CollectionInterval:      10 * time.Millisecond, // Very frequent for testing
		GracefulShutdownTimeout: 1 * time.Second,
	}

	collectionStarted := false
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectionStarted = true
			// Simulate a collection that can be cancelled
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(50 * time.Millisecond):
				return nil
			}
		},
		isConnectedFunc: func() bool {
			return true
		},
	}

	logger := &MockLogger{}

	scheduler, err := NewDefaultScheduler(config, client, logger)
	require.NoError(t, err)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for collection to start and context to be cancelled
	time.Sleep(75 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify collection was attempted
	assert.True(t, collectionStarted, "Collection should have been started")

	// Verify context cancellation was logged
	logMessages := logger.GetLogMessages()
	cancellationLogged := false
	for _, msg := range logMessages {
		if msg == "DEBUG: data collection skipped due to context cancellation" ||
			msg == "DEBUG: data collection completed but context cancelled" ||
			msg == "INFO: scheduler collection loop stopped due to context cancellation" {
			cancellationLogged = true
			break
		}
	}
	assert.True(t, cancellationLogged, "Context cancellation should have been logged")
}

// TestErrorHandling_MemoryPressure tests behavior under memory pressure scenarios
func TestErrorHandling_MemoryPressure(t *testing.T) {
	config := &Config{
		CollectionInterval:      10 * time.Millisecond,
		GracefulShutdownTimeout: 500 * time.Millisecond,
	}

	// Create a client that simulates memory pressure by allocating large amounts
	collectionCount := 0
	client := &MockWinPowerClient{
		collectDataFunc: func(ctx context.Context) error {
			collectionCount++
			// Simulate memory pressure by allocating large buffers
			// Note: This is just for testing, actual memory pressure handling
			// would be more sophisticated in a real scenario
			if collectionCount%10 == 0 {
				// Every 10th collection, simulate a memory-related error
				return errors.New("out of memory")
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

	err = scheduler.Start(ctx)
	require.NoError(t, err)

	// Wait for collections to occur
	time.Sleep(150 * time.Millisecond)

	err = scheduler.Stop()
	require.NoError(t, err)

	// Verify collections occurred
	assert.Greater(t, collectionCount, 0, "Collections should have occurred")

	// Verify memory-related errors were logged
	logMessages := logger.GetLogMessages()
	errorLogged := false
	for _, msg := range logMessages {
		if msg == "ERROR: data collection failed" {
			errorLogged = true
			break
		}
	}
	assert.True(t, errorLogged, "Memory-related errors should have been logged")
}
