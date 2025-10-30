// Package mocks provides mock implementations for testing
package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// MockWinPowerClient is a mock implementation of winpower.Client interface
type MockWinPowerClient struct {
	mu sync.RWMutex

	// Mock data
	mockData     []winpower.ParsedDeviceData
	mockError    error
	callCount    int
	lastCallTime time.Time
	connected    bool

	// Mock statistics
	totalCalls         int
	successfulCalls    int
	failedCalls        int
	totalDevices       int
	lastCollectionTime time.Time

	// Configurable behavior
	ShouldFailAfterNCalls int           // Fail after N successful calls (0 = never fail)
	FailureError          error         // Error to return when failing
	CallDelay             time.Duration // Delay each call to simulate network latency
}

// NewMockWinPowerClient creates a new mock client with default settings
func NewMockWinPowerClient() *MockWinPowerClient {
	return &MockWinPowerClient{
		connected:          true,
		mockData:           []winpower.ParsedDeviceData{},
		lastCollectionTime: time.Now(),
	}
}

// SetMockData sets the data to be returned by CollectDeviceData
func (m *MockWinPowerClient) SetMockData(data []winpower.ParsedDeviceData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockData = data
}

// SetMockError sets the error to be returned by CollectDeviceData
func (m *MockWinPowerClient) SetMockError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockError = err
}

// SetConnected sets the connection status
func (m *MockWinPowerClient) SetConnected(connected bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = connected
}

// CollectDeviceData implements the Client interface
func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulate network delay if configured
	if m.CallDelay > 0 {
		time.Sleep(m.CallDelay)
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.callCount++
	m.totalCalls++
	m.lastCallTime = time.Now()

	// Check if should fail based on configuration
	if m.ShouldFailAfterNCalls > 0 && m.successfulCalls >= m.ShouldFailAfterNCalls {
		m.failedCalls++
		if m.FailureError != nil {
			return nil, m.FailureError
		}
		return nil, m.mockError
	}

	// Return mock error if set
	if m.mockError != nil {
		m.failedCalls++
		return nil, m.mockError
	}

	// Return mock data
	m.successfulCalls++
	m.totalDevices += len(m.mockData)
	m.lastCollectionTime = time.Now()
	return m.mockData, nil
}

// GetConnectionStatus implements the Client interface
func (m *MockWinPowerClient) GetConnectionStatus() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// GetLastCollectionTime implements the Client interface
func (m *MockWinPowerClient) GetLastCollectionTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCollectionTime
}

// GetStatistics returns mock statistics
func (m *MockWinPowerClient) GetStatistics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return map[string]interface{}{
		"total_calls":          m.totalCalls,
		"successful_calls":     m.successfulCalls,
		"failed_calls":         m.failedCalls,
		"total_devices":        m.totalDevices,
		"last_collection_time": m.lastCollectionTime,
	}
}

// Close implements the Client interface
func (m *MockWinPowerClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

// GetCallCount returns the number of times CollectDeviceData was called
func (m *MockWinPowerClient) GetCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

// GetLastCallTime returns the last time CollectDeviceData was called
func (m *MockWinPowerClient) GetLastCallTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastCallTime
}

// Reset resets all counters and state
func (m *MockWinPowerClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount = 0
	m.totalCalls = 0
	m.successfulCalls = 0
	m.failedCalls = 0
	m.totalDevices = 0
	m.mockError = nil
	m.connected = true
	m.ShouldFailAfterNCalls = 0
	m.FailureError = nil
	m.CallDelay = 0
}
