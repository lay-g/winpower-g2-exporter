// Package mocks provides mock implementations of WinPower module interfaces for testing.
//
// This package contains mock implementations of the following interfaces:
//   - MockWinPowerClient: Mock implementation of the Client interface
//   - MockHTTPClient: Mock implementation of the HTTPClient interface
//   - MockTokenManager: Mock implementation of the TokenManager interface
//   - MockDataParser: Mock implementation of the DataParser interface
//
// These mocks are designed to support various testing scenarios including:
//   - Normal operation testing
//   - Error condition testing
//   - Concurrent access testing
//   - Performance testing
//
// Example usage:
//
//	// Create a mock client
//	mockClient := mocks.NewMockWinPowerClient()
//
//	// Configure mock data
//	mockClient.SetMockData([]winpower.ParsedDeviceData{
//	    {DeviceID: "test-device-001", /* ... */},
//	})
//
//	// Use in tests
//	data, err := mockClient.CollectDeviceData(context.Background())
//
//	// Configure error scenarios
//	mockClient.SetMockError(errors.New("network error"))
//	data, err = mockClient.CollectDeviceData(context.Background())
//
//	// Verify call counts
//	callCount := mockClient.GetCallCount()
package mocks
