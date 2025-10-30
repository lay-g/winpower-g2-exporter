package mocks

import (
	"encoding/json"
	"sync"

	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// MockDataParser is a mock implementation of DataParser for testing
type MockDataParser struct {
	mu sync.RWMutex

	// Mock data
	mockData  []winpower.ParsedDeviceData
	mockError error

	// Call tracking
	parseCalls int

	// Configurable behavior
	FailAfterN int
}

// NewMockDataParser creates a new mock data parser
func NewMockDataParser() *MockDataParser {
	return &MockDataParser{
		mockData: []winpower.ParsedDeviceData{},
	}
}

// SetMockData sets the data to return from ParseResponse
func (m *MockDataParser) SetMockData(data []winpower.ParsedDeviceData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockData = data
}

// SetMockError sets the error to return from ParseResponse
func (m *MockDataParser) SetMockError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockError = err
}

// ParseResponse mocks the DataParser ParseResponse method
func (m *MockDataParser) ParseResponse(resp *winpower.DeviceDataResponse) ([]winpower.ParsedDeviceData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.parseCalls++

	// Check for configured failure
	if m.FailAfterN > 0 && m.parseCalls >= m.FailAfterN {
		return nil, &winpower.ParseError{Message: "mock parse failure"}
	}

	// Return mock error if set
	if m.mockError != nil {
		return nil, m.mockError
	}

	// Check if response is nil
	if resp == nil {
		return nil, &winpower.ParseError{Message: "nil response"}
	}

	// Check for non-success code
	if resp.Code != "000000" {
		return nil, &winpower.ParseError{Message: "non-success code: " + resp.Code}
	}

	return m.mockData, nil
}

// GetParseCalls returns the number of ParseResponse calls
func (m *MockDataParser) GetParseCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.parseCalls
}

// Reset resets all counters and state
func (m *MockDataParser) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.parseCalls = 0
	m.mockError = nil
	m.FailAfterN = 0
	m.mockData = []winpower.ParsedDeviceData{}
}

// SetMockDataFromJSON sets mock data from JSON string
func (m *MockDataParser) SetMockDataFromJSON(jsonData string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var data []winpower.ParsedDeviceData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}
	m.mockData = data
	return nil
}

// AddMockDevice adds a single device to the mock data
func (m *MockDataParser) AddMockDevice(device winpower.ParsedDeviceData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockData = append(m.mockData, device)
}

// ClearMockData clears all mock data
func (m *MockDataParser) ClearMockData() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockData = []winpower.ParsedDeviceData{}
}
