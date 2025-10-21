package scheduler

import (
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultSchedulerFactory_Creation tests DefaultSchedulerFactory creation
func TestDefaultSchedulerFactory_Creation(t *testing.T) {
	// Create factory
	factory := NewDefaultSchedulerFactory()

	// Verify factory is created
	assert.NotNil(t, factory, "Factory should be created")

	// Verify factory implements SchedulerFactory interface
	var _ SchedulerFactory = factory
}

// TestDefaultSchedulerFactory_CreateScheduler_Success tests successful scheduler creation
func TestDefaultSchedulerFactory_CreateScheduler_Success(t *testing.T) {
	factory := NewDefaultSchedulerFactory()

	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler, err := factory.CreateScheduler(config, client, logger)

	// Verify success
	require.NoError(t, err, "Scheduler creation should succeed")
	require.NotNil(t, scheduler, "Scheduler should be created")

	// Verify scheduler implements Scheduler interface
	_ = scheduler

	// Verify scheduler is not running initially
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running initially")
}

// TestDefaultSchedulerFactory_CreateScheduler_WithNilConfig tests scheduler creation with nil config
func TestDefaultSchedulerFactory_CreateScheduler_WithNilConfig(t *testing.T) {
	factory := NewDefaultSchedulerFactory()

	// Create test dependencies
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler with nil config
	scheduler, err := factory.CreateScheduler(nil, client, logger)

	// Verify success (should use default config)
	require.NoError(t, err, "Scheduler creation with nil config should succeed")
	require.NotNil(t, scheduler, "Scheduler should be created")

	// Verify scheduler is not running initially
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running initially")
}

// TestDefaultSchedulerFactory_CreateScheduler_WithInvalidConfig tests scheduler creation with invalid config
func TestDefaultSchedulerFactory_CreateScheduler_WithInvalidConfig(t *testing.T) {
	factory := NewDefaultSchedulerFactory()

	// Create test dependencies
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create invalid config
	invalidConfig := &Config{
		CollectionInterval:      0, // Invalid
		GracefulShutdownTimeout: 30 * time.Second,
	}

	// Try to create scheduler
	scheduler, err := factory.CreateScheduler(invalidConfig, client, logger)

	// Verify failure
	assert.Error(t, err, "Scheduler creation with invalid config should fail")
	assert.Nil(t, scheduler, "Scheduler should not be created with invalid config")
	assert.Contains(t, err.Error(), "collection interval cannot be zero or negative",
		"Error should mention the specific validation failure")
}

// TestDefaultSchedulerFactory_CreateScheduler_WithNilDependencies tests scheduler creation with nil dependencies
func TestDefaultSchedulerFactory_CreateScheduler_WithNilDependencies(t *testing.T) {
	factory := NewDefaultSchedulerFactory()

	config := DefaultConfig()

	testCases := []struct {
		name        string
		client      WinPowerClient
		logger      log.Logger
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil client",
			client:      nil,
			logger:      &MockLogger{},
			expectError: true,
			errorMsg:    "client cannot be nil",
		},
		{
			name:        "nil logger",
			client:      &MockWinPowerClient{},
			logger:      nil,
			expectError: true,
			errorMsg:    "logger cannot be nil",
		},
		{
			name:        "both nil",
			client:      nil,
			logger:      nil,
			expectError: true,
			errorMsg:    "client cannot be nil", // First validation that fails
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Try to create scheduler
			scheduler, err := factory.CreateScheduler(config, tc.client, tc.logger)

			if tc.expectError {
				assert.Error(t, err, "Expected error for "+tc.name)
				assert.Nil(t, scheduler, "Scheduler should not be created for "+tc.name)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error should contain expected message")
				}
			} else {
				assert.NoError(t, err, "Should not error for "+tc.name)
				assert.NotNil(t, scheduler, "Scheduler should be created for "+tc.name)
			}
		})
	}
}

// TestCreateScheduler_ConvenienceFunction tests the convenience function
func TestCreateScheduler_ConvenienceFunction(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler using convenience function
	scheduler, err := CreateScheduler(config, client, logger)

	// Verify success
	require.NoError(t, err, "Convenience function should succeed")
	require.NotNil(t, scheduler, "Scheduler should be created")

	// Verify scheduler implements Scheduler interface
	_ = scheduler

	// Verify scheduler is not running initially
	assert.False(t, scheduler.IsRunning(), "Scheduler should not be running initially")
}

// TestDefaultFactory_GlobalInstance tests the global factory instance
func TestDefaultFactory_GlobalInstance(t *testing.T) {
	// Verify global factory is not nil
	assert.NotNil(t, DefaultFactory, "Default factory should not be nil")

	// Test using global factory
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	scheduler, err := DefaultFactory.CreateScheduler(config, client, logger)

	// Verify success
	require.NoError(t, err, "Global factory should succeed")
	require.NotNil(t, scheduler, "Scheduler should be created")
}
