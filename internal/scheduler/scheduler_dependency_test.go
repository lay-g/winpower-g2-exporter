package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDefaultScheduler_DependencyInjection tests NewDefaultScheduler dependency injection verification
func TestNewDefaultScheduler_DependencyInjection(t *testing.T) {
	// Create custom test dependencies that we can verify
	mockClient := &MockWinPowerClient{
		isConnectedFunc: func() bool { return false }, // Custom behavior
		collectDataFunc: func(ctx context.Context) error {
			return assert.AnError // Custom error for testing
		},
	}

	mockLogger := &MockLogger{}

	// Create custom config
	customConfig := &Config{
		CollectionInterval:      10 * time.Second,
		GracefulShutdownTimeout: 60 * time.Second,
	}
	require.NoError(t, customConfig.Validate(), "Custom config should be valid")

	// Create scheduler with injected dependencies
	scheduler, err := NewDefaultScheduler(customConfig, mockClient, mockLogger)
	require.NoError(t, err, "Should create scheduler successfully")
	require.NotNil(t, scheduler, "Scheduler should not be nil")

	// Verify config injection
	assert.Equal(t, customConfig, scheduler.config, "Should inject custom config")
	assert.Equal(t, 10*time.Second, scheduler.config.CollectionInterval, "Should have custom collection interval")
	assert.Equal(t, 60*time.Second, scheduler.config.GracefulShutdownTimeout, "Should have custom timeout")

	// Verify client injection
	assert.Equal(t, mockClient, scheduler.client, "Should inject custom client")

	// Test that the injected client actually behaves as expected
	assert.False(t, scheduler.client.IsConnected(), "Custom client should return false for IsConnected")

	// Verify logger injection
	assert.Equal(t, mockLogger, scheduler.logger, "Should inject custom logger")

	// Test that logger methods can be called through injected dependency
	scheduler.logger.Info("Test message")
	logMessages := mockLogger.GetLogMessages()
	assert.Contains(t, logMessages, "INFO: Test message", "Injected logger should capture messages")

	// Test interface compliance - the injected scheduler should still implement Scheduler interface
	var _ Scheduler = scheduler
}
