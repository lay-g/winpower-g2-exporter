package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// MockTokenManager is a mock implementation of TokenManager for testing
type MockTokenManager struct {
	mu sync.RWMutex

	// Mock data
	mockToken     string
	mockDeviceID  string
	mockExpiresAt time.Time
	mockError     error

	// Call tracking
	getTokenCalls int
	lastContext   context.Context

	// Configurable behavior
	ShouldRefresh bool
	FailAfterN    int
}

// NewMockTokenManager creates a new mock token manager
func NewMockTokenManager() *MockTokenManager {
	return &MockTokenManager{
		mockToken:     "mock-token-default",
		mockDeviceID:  "mock-device-default",
		mockExpiresAt: time.Now().Add(1 * time.Hour),
	}
}

// SetMockToken sets the token to return
func (m *MockTokenManager) SetMockToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockToken = token
}

// SetMockDeviceID sets the device ID to return
func (m *MockTokenManager) SetMockDeviceID(deviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockDeviceID = deviceID
}

// SetMockExpiresAt sets the expiration time
func (m *MockTokenManager) SetMockExpiresAt(expiresAt time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockExpiresAt = expiresAt
}

// SetMockError sets the error to return
func (m *MockTokenManager) SetMockError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockError = err
}

// GetToken mocks the TokenManager GetToken method
func (m *MockTokenManager) GetToken(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getTokenCalls++
	m.lastContext = ctx

	// Check for configured failure
	if m.FailAfterN > 0 && m.getTokenCalls >= m.FailAfterN {
		return "", &winpower.AuthenticationError{Message: "mock token failure"}
	}

	// Return mock error if set
	if m.mockError != nil {
		return "", m.mockError
	}

	return m.mockToken, nil
}

// GetCachedToken returns the cached token information
func (m *MockTokenManager) GetCachedToken() *winpower.TokenCache {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.mockToken == "" {
		return nil
	}

	return &winpower.TokenCache{
		Token:     m.mockToken,
		ExpiresAt: m.mockExpiresAt,
		DeviceID:  m.mockDeviceID,
	}
}

// GetExpiresAt returns when the cached token expires
func (m *MockTokenManager) GetExpiresAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mockExpiresAt
}

// GetDeviceID returns the device ID associated with the cached token
func (m *MockTokenManager) GetDeviceID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mockDeviceID
}

// IsValid returns whether the cached token is still valid
func (m *MockTokenManager) IsValid() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mockToken != "" && time.Now().Before(m.mockExpiresAt)
}

// ClearCache clears the cached token
func (m *MockTokenManager) ClearCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mockToken = ""
	m.mockDeviceID = ""
	m.mockExpiresAt = time.Time{}
}

// GetTokenCalls returns the number of GetToken calls
func (m *MockTokenManager) GetTokenCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getTokenCalls
}

// GetLastContext returns the last context used in GetToken
func (m *MockTokenManager) GetLastContext() context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastContext
}

// Reset resets all counters and state
func (m *MockTokenManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getTokenCalls = 0
	m.lastContext = nil
	m.mockError = nil
	m.ShouldRefresh = false
	m.FailAfterN = 0
	m.mockToken = "mock-token-default"
	m.mockDeviceID = "mock-device-default"
	m.mockExpiresAt = time.Now().Add(1 * time.Hour)
}
