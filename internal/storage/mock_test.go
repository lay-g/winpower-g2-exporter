package storage

import (
	"fmt"
	"sync"
)

// MockStorageManager is a mock implementation of StorageManager for testing.
type MockStorageManager struct {
	// Test data storage
	data map[string]*PowerData

	// Call tracking
	writeCalls []WriteCall
	readCalls  []ReadCall
	callsMutex sync.Mutex

	// Behavior configuration
	writeError     error
	readError      error
	autoInitialize bool

	// Configuration
	config *Config
}

// WriteCall represents a recorded write call.
type WriteCall struct {
	DeviceID string
	Data     *PowerData
}

// ReadCall represents a recorded read call.
type ReadCall struct {
	DeviceID string
}

// NewMockStorageManager creates a new mock storage manager.
func NewMockStorageManager() *MockStorageManager {
	return &MockStorageManager{
		data:           make(map[string]*PowerData),
		writeCalls:     make([]WriteCall, 0),
		readCalls:      make([]ReadCall, 0),
		autoInitialize: true,
		config:         NewConfigWithDefaults(),
	}
}

// NewMockStorageManagerWithConfig creates a new mock storage manager with specific config.
func NewMockStorageManagerWithConfig(config *Config) *MockStorageManager {
	configClone := config.Clone()
	clonedConfig, ok := configClone.(*Config)
	if !ok {
		// Fall back to a default config if cloning fails
		clonedConfig = NewConfigWithDefaults()
	}
	return &MockStorageManager{
		data:           make(map[string]*PowerData),
		writeCalls:     make([]WriteCall, 0),
		readCalls:      make([]ReadCall, 0),
		autoInitialize: true,
		config:         clonedConfig,
	}
}

// Write stores power data for a device.
func (m *MockStorageManager) Write(deviceID string, data *PowerData) error {
	m.callsMutex.Lock()
	m.writeCalls = append(m.writeCalls, WriteCall{
		DeviceID: deviceID,
		Data:     data,
	})
	m.callsMutex.Unlock()

	if m.writeError != nil {
		return m.writeError
	}

	if data == nil {
		return fmt.Errorf("data cannot be nil")
	}

	m.data[deviceID] = data
	return nil
}

// Read retrieves power data for a device.
func (m *MockStorageManager) Read(deviceID string) (*PowerData, error) {
	m.callsMutex.Lock()
	m.readCalls = append(m.readCalls, ReadCall{
		DeviceID: deviceID,
	})
	m.callsMutex.Unlock()

	if m.readError != nil {
		return nil, m.readError
	}

	if data, exists := m.data[deviceID]; exists {
		return data, nil
	}

	if m.autoInitialize {
		return NewPowerData(0), nil
	}

	return nil, fmt.Errorf("no data found for device '%s'", deviceID)
}

// SetData sets predefined data for a device.
func (m *MockStorageManager) SetData(deviceID string, data *PowerData) {
	m.data[deviceID] = data
}

// GetData returns the stored data for a device.
func (m *MockStorageManager) GetData(deviceID string) (*PowerData, bool) {
	data, exists := m.data[deviceID]
	return data, exists
}

// GetAllData returns all stored data.
func (m *MockStorageManager) GetAllData() map[string]*PowerData {
	result := make(map[string]*PowerData)
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

// ClearData removes all stored data.
func (m *MockStorageManager) ClearData() {
	m.data = make(map[string]*PowerData)
}

// SetWriteError configures the mock to return an error on Write calls.
func (m *MockStorageManager) SetWriteError(err error) {
	m.writeError = err
}

// SetReadError configures the mock to return an error on Read calls.
func (m *MockStorageManager) SetReadError(err error) {
	m.readError = err
}

// SetAutoInitialize configures whether the mock should return initialized data for unknown devices.
func (m *MockStorageManager) SetAutoInitialize(autoInit bool) {
	m.autoInitialize = autoInit
}

// GetWriteCalls returns all recorded write calls.
func (m *MockStorageManager) GetWriteCalls() []WriteCall {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	calls := make([]WriteCall, len(m.writeCalls))
	copy(calls, m.writeCalls)
	return calls
}

// GetReadCalls returns all recorded read calls.
func (m *MockStorageManager) GetReadCalls() []ReadCall {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	calls := make([]ReadCall, len(m.readCalls))
	copy(calls, m.readCalls)
	return calls
}

// GetWriteCallCount returns the number of write calls made.
func (m *MockStorageManager) GetWriteCallCount() int {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()
	return len(m.writeCalls)
}

// GetReadCallCount returns the number of read calls made.
func (m *MockStorageManager) GetReadCallCount() int {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()
	return len(m.readCalls)
}

// ResetCallHistory clears all recorded calls.
func (m *MockStorageManager) ResetCallHistory() {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	m.writeCalls = make([]WriteCall, 0)
	m.readCalls = make([]ReadCall, 0)
}

// WasWriteCalled checks if Write was called for a specific device.
func (m *MockStorageManager) WasWriteCalled(deviceID string) bool {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	for _, call := range m.writeCalls {
		if call.DeviceID == deviceID {
			return true
		}
	}
	return false
}

// WasReadCalled checks if Read was called for a specific device.
func (m *MockStorageManager) WasReadCalled(deviceID string) bool {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	for _, call := range m.readCalls {
		if call.DeviceID == deviceID {
			return true
		}
	}
	return false
}

// GetLastWriteCall returns the last write call made.
func (m *MockStorageManager) GetLastWriteCall() (*WriteCall, bool) {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	if len(m.writeCalls) == 0 {
		return nil, false
	}

	return &m.writeCalls[len(m.writeCalls)-1], true
}

// GetLastReadCall returns the last read call made.
func (m *MockStorageManager) GetLastReadCall() (*ReadCall, bool) {
	m.callsMutex.Lock()
	defer m.callsMutex.Unlock()

	if len(m.readCalls) == 0 {
		return nil, false
	}

	return &m.readCalls[len(m.readCalls)-1], true
}

// GetConfig returns the mock's configuration.
func (m *MockStorageManager) GetConfig() *Config {
	configClone := m.config.Clone()
	clonedConfig, ok := configClone.(*Config)
	if !ok {
		// This should never happen, but return a new config if it does
		return NewConfigWithDefaults()
	}
	return clonedConfig
}

// GetConfig is a method to satisfy the same interface as FileStorageManager.
func (m *MockStorageManager) GetDataDir() string {
	return m.config.DataDir
}
