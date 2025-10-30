package mocks

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	mu sync.RWMutex

	// Mock responses
	loginResponse      *winpower.LoginResponse
	deviceDataResponse *winpower.DeviceDataResponse
	loginError         error
	deviceDataError    error

	// Call tracking
	loginCalls      int
	deviceDataCalls int
	lastToken       string

	// Configurable behavior
	LoginDelay      int // Fail login after N successful logins
	DeviceDataDelay int // Fail device data after N successful calls
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	mockLogin := &winpower.LoginResponse{
		Code:    "000000",
		Message: "OK",
	}
	mockLogin.Data.Token = "mock-token-123"
	mockLogin.Data.DeviceID = "mock-device-001"

	return &MockHTTPClient{
		loginResponse: mockLogin,
		deviceDataResponse: &winpower.DeviceDataResponse{
			Code:        "000000",
			Msg:         "OK",
			Total:       0,
			PageSize:    20,
			CurrentPage: 1,
			Data:        []winpower.DeviceInfo{},
		},
	}
}

// SetLoginResponse sets the mock login response
func (m *MockHTTPClient) SetLoginResponse(resp *winpower.LoginResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loginResponse = resp
}

// SetDeviceDataResponse sets the mock device data response
func (m *MockHTTPClient) SetDeviceDataResponse(resp *winpower.DeviceDataResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deviceDataResponse = resp
}

// SetLoginError sets the error to return for login attempts
func (m *MockHTTPClient) SetLoginError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loginError = err
}

// SetDeviceDataError sets the error to return for device data requests
func (m *MockHTTPClient) SetDeviceDataError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deviceDataError = err
}

// Login mocks the HTTPClient Login method
func (m *MockHTTPClient) Login(ctx context.Context, username, password string) (*winpower.LoginResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.loginCalls++

	// Check for error
	if m.loginError != nil {
		return nil, m.loginError
	}

	// Check for configured failure
	if m.LoginDelay > 0 && m.loginCalls >= m.LoginDelay {
		return nil, &winpower.AuthenticationError{Message: "mock login failure"}
	}

	return m.loginResponse, nil
}

// GetDeviceData mocks the HTTPClient GetDeviceData method
func (m *MockHTTPClient) GetDeviceData(ctx context.Context, token string) (*winpower.DeviceDataResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deviceDataCalls++
	m.lastToken = token

	// Check for error
	if m.deviceDataError != nil {
		return nil, m.deviceDataError
	}

	// Check for configured failure
	if m.DeviceDataDelay > 0 && m.deviceDataCalls >= m.DeviceDataDelay {
		return nil, &winpower.NetworkError{Message: "mock device data failure"}
	}

	return m.deviceDataResponse, nil
}

// Close mocks the HTTPClient Close method
func (m *MockHTTPClient) Close() error {
	return nil
}

// GetLoginCalls returns the number of login attempts
func (m *MockHTTPClient) GetLoginCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.loginCalls
}

// GetDeviceDataCalls returns the number of device data requests
func (m *MockHTTPClient) GetDeviceDataCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deviceDataCalls
}

// GetLastToken returns the last token used for device data request
func (m *MockHTTPClient) GetLastToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastToken
}

// Reset resets all counters and errors
func (m *MockHTTPClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loginCalls = 0
	m.deviceDataCalls = 0
	m.lastToken = ""
	m.loginError = nil
	m.deviceDataError = nil
	m.LoginDelay = 0
	m.DeviceDataDelay = 0
}

// SetDeviceDataFromJSON sets device data from JSON string
func (m *MockHTTPClient) SetDeviceDataFromJSON(jsonData string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var resp winpower.DeviceDataResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		return err
	}
	m.deviceDataResponse = &resp
	return nil
}

// SetLoginResponseFromJSON sets login response from JSON string
func (m *MockHTTPClient) SetLoginResponseFromJSON(jsonData string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var resp winpower.LoginResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		return err
	}
	m.loginResponse = &resp
	return nil
}
