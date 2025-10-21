package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_StructureFields tests the Config struct fields definition
func TestConfig_StructureFields(t *testing.T) {
	// Create a config instance
	config := &Config{}

	// Test that the struct has the expected fields
	assert.NotNil(t, config, "Config struct should be created")

	// Test that the struct has the expected fields by setting values
	config.CollectionInterval = 10 * time.Second
	config.GracefulShutdownTimeout = 60 * time.Second

	assert.Equal(t, 10*time.Second, config.CollectionInterval, "CollectionInterval field should be settable")
	assert.Equal(t, 60*time.Second, config.GracefulShutdownTimeout, "GracefulShutdownTimeout field should be settable")
}

// TestDefaultConfig_ReturnsCorrectDefaults tests DefaultConfig function returns correct default values
func TestDefaultConfig_ReturnsCorrectDefaults(t *testing.T) {
	// Call DefaultConfig
	config := DefaultConfig()

	// Test that default config is not nil
	require.NotNil(t, config, "DefaultConfig should return a non-nil config")

	// Test default values
	assert.Equal(t, 5*time.Second, config.CollectionInterval, "Default collection interval should be 5 seconds")
	assert.Equal(t, 30*time.Second, config.GracefulShutdownTimeout, "Default graceful shutdown timeout should be 30 seconds")
}

// TestConfig_Validate_Scenarios tests Config.Validate method with various validation scenarios
func TestConfig_Validate_Scenarios(t *testing.T) {
	testCases := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				CollectionInterval:      5 * time.Second,
				GracefulShutdownTimeout: 30 * time.Second,
			},
			expectError: false,
		},
		{
			name: "zero collection interval",
			config: &Config{
				CollectionInterval:      0,
				GracefulShutdownTimeout: 30 * time.Second,
			},
			expectError: true,
			errorMsg:    "collection interval cannot be zero or negative",
		},
		{
			name: "negative collection interval",
			config: &Config{
				CollectionInterval:      -5 * time.Second,
				GracefulShutdownTimeout: 30 * time.Second,
			},
			expectError: true,
			errorMsg:    "collection interval cannot be zero or negative",
		},
		{
			name: "zero graceful shutdown timeout",
			config: &Config{
				CollectionInterval:      5 * time.Second,
				GracefulShutdownTimeout: 0,
			},
			expectError: true,
			errorMsg:    "graceful shutdown timeout cannot be zero or negative",
		},
		{
			name: "negative graceful shutdown timeout",
			config: &Config{
				CollectionInterval:      5 * time.Second,
				GracefulShutdownTimeout: -30 * time.Second,
			},
			expectError: true,
			errorMsg:    "graceful shutdown timeout cannot be zero or negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()

			if tc.expectError {
				assert.Error(t, err, "Expected validation error")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestSchedulerModuleConfig_Registration tests SchedulerModuleConfig registration functionality
func TestSchedulerModuleConfig_Registration(t *testing.T) {
	// Test that the module config is registered properly
	moduleConfig := GetModuleConfig()
	require.NotNil(t, moduleConfig, "Module config should be registered")
	require.NotNil(t, moduleConfig.Config, "Module config should have a Config")

	// Test that the config is properly initialized with defaults
	assert.Equal(t, 5*time.Second, moduleConfig.Config.CollectionInterval, "Module config should have default collection interval")
	assert.Equal(t, 30*time.Second, moduleConfig.Config.GracefulShutdownTimeout, "Module config should have default graceful shutdown timeout")

	// Test that we can set a new module config
	newConfig := &SchedulerModuleConfig{
		Config: &Config{
			CollectionInterval:      10 * time.Second,
			GracefulShutdownTimeout: 60 * time.Second,
		},
	}
	SetModuleConfig(newConfig)

	// Verify the new config was set
	updatedModuleConfig := GetModuleConfig()
	assert.Equal(t, 10*time.Second, updatedModuleConfig.Config.CollectionInterval, "Module config should be updated")
	assert.Equal(t, 60*time.Second, updatedModuleConfig.Config.GracefulShutdownTimeout, "Module config should be updated")
}

// MockWinPowerClient is a mock implementation of WinPower client for testing
type MockWinPowerClient struct {
	collectDataFunc func(ctx context.Context) error
	isConnectedFunc func() bool
}

func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) error {
	if m.collectDataFunc != nil {
		return m.collectDataFunc(ctx)
	}
	return nil
}

func (m *MockWinPowerClient) IsConnected() bool {
	if m.isConnectedFunc != nil {
		return m.isConnectedFunc()
	}
	return true
}

// MockLogger is a mock implementation of Logger for testing
type MockLogger struct {
	logMessages []string
}

func (m *MockLogger) Debug(msg string, fields ...log.Field) {
	m.logMessages = append(m.logMessages, "DEBUG: "+msg)
}

func (m *MockLogger) Info(msg string, fields ...log.Field) {
	m.logMessages = append(m.logMessages, "INFO: "+msg)
}

func (m *MockLogger) Warn(msg string, fields ...log.Field) {
	m.logMessages = append(m.logMessages, "WARN: "+msg)
}

func (m *MockLogger) Error(msg string, fields ...log.Field) {
	m.logMessages = append(m.logMessages, "ERROR: "+msg)
}

func (m *MockLogger) Fatal(msg string, fields ...log.Field) {
	m.logMessages = append(m.logMessages, "FATAL: "+msg)
}

func (m *MockLogger) With(fields ...log.Field) log.Logger {
	return m
}

func (m *MockLogger) WithContext(ctx context.Context) log.Logger {
	return m
}

func (m *MockLogger) Sync() error {
	return nil
}

// GetLogMessages returns the logged messages for testing
func (m *MockLogger) GetLogMessages() []string {
	return m.logMessages
}

// ClearLogMessages clears the logged messages
func (m *MockLogger) ClearLogMessages() {
	m.logMessages = nil
}

// TestDefaultScheduler_BasicFields tests DefaultScheduler struct basic fields
func TestDefaultScheduler_BasicFields(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler := &DefaultScheduler{
		config:    config,
		client:    client,
		logger:    logger,
		isRunning: false,
	}

	// Test that scheduler has the expected fields
	assert.NotNil(t, scheduler, "DefaultScheduler should be created")
	assert.Equal(t, config, scheduler.config, "Config field should be set correctly")
	assert.Equal(t, client, scheduler.client, "Client field should be set correctly")
	assert.Equal(t, logger, scheduler.logger, "Logger field should be set correctly")
	assert.False(t, scheduler.isRunning, "Initial isRunning state should be false")
}

// TestDefaultScheduler_InterfaceCompliance tests DefaultScheduler implements Scheduler interface
func TestDefaultScheduler_InterfaceCompliance(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Create scheduler
	scheduler := &DefaultScheduler{
		config:    config,
		client:    client,
		logger:    logger,
		isRunning: false,
	}

	// Test that DefaultScheduler implements Scheduler interface
	var _ Scheduler = scheduler

	// Test that scheduler has the required methods
	assert.NotNil(t, scheduler.Start, "Scheduler should have Start method")
	assert.NotNil(t, scheduler.Stop, "Scheduler should have Stop method")
	assert.NotNil(t, scheduler.IsRunning, "Scheduler should have IsRunning method")
}

// TestNewDefaultScheduler_ValidConstruction tests NewDefaultScheduler normal construction scenarios
func TestNewDefaultScheduler_ValidConstruction(t *testing.T) {
	// Create test dependencies
	config := DefaultConfig()
	client := &MockWinPowerClient{}
	logger := &MockLogger{}

	// Test constructor with valid parameters
	scheduler, err := NewDefaultScheduler(config, client, logger)

	// Verify constructor succeeded
	require.NoError(t, err, "NewDefaultScheduler should not return error with valid parameters")
	require.NotNil(t, scheduler, "NewDefaultScheduler should return a non-nil scheduler")

	// Verify the scheduler has correct dependencies
	assert.Equal(t, config, scheduler.config, "Scheduler should have correct config")
	assert.Equal(t, client, scheduler.client, "Scheduler should have correct client")
	assert.Equal(t, logger, scheduler.logger, "Scheduler should have correct logger")
	assert.False(t, scheduler.isRunning, "Scheduler should not be running initially")

	// Test constructor with default config and custom client/logger
	customClient := &MockWinPowerClient{}
	customLogger := &MockLogger{}

	scheduler2, err := NewDefaultScheduler(DefaultConfig(), customClient, customLogger)
	require.NoError(t, err, "NewDefaultScheduler should work with default config")
	require.NotNil(t, scheduler2, "NewDefaultScheduler should return valid scheduler")
	assert.Equal(t, customClient, scheduler2.client, "Should use custom client")
	assert.Equal(t, customLogger, scheduler2.logger, "Should use custom logger")
}

// TestNewDefaultScheduler_InvalidParameters tests NewDefaultScheduler invalid parameter handling
func TestNewDefaultScheduler_InvalidParameters(t *testing.T) {
	// Create valid dependencies
	validConfig := DefaultConfig()
	validClient := &MockWinPowerClient{}
	validLogger := &MockLogger{}

	testCases := []struct {
		name        string
		config      *Config
		client      WinPowerClient
		logger      log.Logger
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil config",
			config:      nil,
			client:      validClient,
			logger:      validLogger,
			expectError: true,
			errorMsg:    "config cannot be nil",
		},
		{
			name:        "nil client",
			config:      validConfig,
			client:      nil,
			logger:      validLogger,
			expectError: true,
			errorMsg:    "client cannot be nil",
		},
		{
			name:        "nil logger",
			config:      validConfig,
			client:      validClient,
			logger:      nil,
			expectError: true,
			errorMsg:    "logger cannot be nil",
		},
		{
			name: "invalid config (zero collection interval)",
			config: &Config{
				CollectionInterval:      0,
				GracefulShutdownTimeout: 30 * time.Second,
			},
			client:      validClient,
			logger:      validLogger,
			expectError: true,
			errorMsg:    "collection interval cannot be zero or negative",
		},
		{
			name: "invalid config (negative graceful shutdown timeout)",
			config: &Config{
				CollectionInterval:      5 * time.Second,
				GracefulShutdownTimeout: -30 * time.Second,
			},
			client:      validClient,
			logger:      validLogger,
			expectError: true,
			errorMsg:    "graceful shutdown timeout cannot be zero or negative",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheduler, err := NewDefaultScheduler(tc.config, tc.client, tc.logger)

			if tc.expectError {
				assert.Error(t, err, "Expected constructor to return error")
				assert.Nil(t, scheduler, "Expected constructor to return nil scheduler on error")
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error")
				assert.NotNil(t, scheduler, "Expected valid scheduler")
			}
		})
	}
}
