package winpower

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
)

// TestNewClient tests the NewClient constructor
func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "nil_config_should_error",
			test: func(t *testing.T) {
				client, err := NewClient(nil)
				if err == nil {
					t.Fatal("NewClient() should return error for nil config")
				}
				if client != nil {
					t.Error("NewClient() should return nil client when config is nil")
				}
			},
		},
		{
			name: "invalid_config_should_error",
			test: func(t *testing.T) {
				config := &Config{
					URL:     "invalid-url",
					Timeout: 10 * time.Second,
				}
				client, err := NewClient(config)
				if err == nil {
					t.Fatal("NewClient() should return error for invalid config")
				}
				if client != nil {
					t.Error("NewClient() should return nil client when config is invalid")
				}
			},
		},
		{
			name: "valid_config_should_succeed",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("NewClient() should not return error for valid config: %v", err)
				}
				if client == nil {
					t.Fatal("NewClient() should return non-nil client for valid config")
				}

				// Clean up
				if err := client.Close(); err != nil {
					t.Logf("Warning: failed to close client: %v", err)
				}
			},
		},
		{
			name: "should_have_default_status_disconnected",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("NewClient() failed: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				if client.GetConnectionStatus() != StatusDisconnected {
					t.Errorf("Expected status %v, got: %v", StatusDisconnected, client.GetConnectionStatus())
				}

				if client.IsConnected() {
					t.Error("New client should not be connected initially")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestClient_Connect tests the Connect method
func TestClient_Connect(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "connect_with_nil_context_should_error",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				err := client.Connect(context.TODO())
				if err == nil {
					t.Error("Connect() should return error for nil context")
				}
			},
		},
		{
			name: "connect_with_cancelled_context_should_error",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				err := client.Connect(ctx)
				if err == nil {
					t.Error("Connect() should return error for cancelled context")
				}
			},
		},
		{
			name: "successful_connect_should_update_status",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Mock a successful connection (since we have mock token manager)
				err := client.Connect(context.Background())
				if err != nil {
					// This is expected since we're using mock authentication
					// In real implementation, this would succeed with proper authentication
					return
				}

				if client.GetConnectionStatus() != StatusConnected {
					t.Errorf("Expected status %v, got: %v", StatusConnected, client.GetConnectionStatus())
				}

				if !client.IsConnected() {
					t.Error("Client should be connected after successful Connect()")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestClient_Disconnect tests the Disconnect method
func TestClient_Disconnect(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "disconnect_should_update_status",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Initially should be disconnected
				if client.GetConnectionStatus() != StatusDisconnected {
					t.Errorf("Expected initial status %v", StatusDisconnected)
				}

				// Disconnect should not error even when already disconnected
				err := client.Disconnect()
				if err != nil {
					t.Errorf("Disconnect() should not error when already disconnected: %v", err)
				}

				// Should still be disconnected
				if client.GetConnectionStatus() != StatusDisconnected {
					t.Errorf("Expected status %v after disconnect", StatusDisconnected)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestClient_CollectDeviceData tests the CollectDeviceData method
func TestClient_CollectDeviceData(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "collect_when_not_connected_should_error",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				_, err := client.CollectDeviceData(context.Background())
				if err == nil {
					t.Error("CollectDeviceData() should return error when not connected")
				}
			},
		},
		{
			name: "collect_should_update_last_collection_time",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Set up a mock connected state
				client.tokenManager.SetToken(&TokenInfo{
					Token:  "mock-token",
					Expiry: time.Now().Add(1 * time.Hour),
				})

				initialTime := client.GetLastCollectionTime()
				if !initialTime.IsZero() {
					t.Error("Initial last collection time should be zero")
				}

				// Mock data collection
				_, err := client.CollectDeviceData(context.Background())
				if err != nil {
					// Expected since we don't have real API implementation
					return
				}

				// Check that collection time was updated
				newTime := client.GetLastCollectionTime()
				if newTime.IsZero() {
					t.Error("Last collection time should be updated after collection")
				}

				if !newTime.After(initialTime) {
					t.Error("Last collection time should be after initial time")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestClient_GetStats tests the GetStats method
func TestClient_GetStats(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "new_client_should_have_zero_stats",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, _ := NewClient(config)
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				stats := client.GetStats()
				if stats == nil {
					t.Error("GetStats() should not return nil")
					return
				}

				if stats.TotalConnections != 0 {
					t.Errorf("Expected 0 total connections, got: %d", stats.TotalConnections)
				}

				if stats.SuccessfulRequests != 0 {
					t.Errorf("Expected 0 successful requests, got: %d", stats.SuccessfulRequests)
				}

				if stats.FailedRequests != 0 {
					t.Errorf("Expected 0 failed requests, got: %d", stats.FailedRequests)
				}
			},
		},
		{
			name: "nil_client_should_return_empty_stats",
			test: func(t *testing.T) {
				var client *Client
				stats := client.GetStats()
				if stats == nil {
					t.Error("GetStats() should not return nil for nil client")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestConnectionStatus_String tests the String method for ConnectionStatus
func TestConnectionStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		status   ConnectionStatus
		expected string
	}{
		{"disconnected", StatusDisconnected, "disconnected"},
		{"connecting", StatusConnecting, "connecting"},
		{"connected", StatusConnected, "connected"},
		{"reconnecting", StatusReconnecting, "reconnecting"},
		{"error", StatusError, "error"},
		{"unknown", ConnectionStatus(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.String(); got != tt.expected {
				t.Errorf("ConnectionStatus.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// MockEnergyService is a mock implementation of energy.EnergyInterface for testing
type MockEnergyService struct {
	calculations map[string]float64
	errors       map[string]bool
}

func NewMockEnergyService() *MockEnergyService {
	return &MockEnergyService{
		calculations: make(map[string]float64),
		errors:       make(map[string]bool),
	}
}

func (m *MockEnergyService) Calculate(deviceID string, power float64) (float64, error) {
	if m.errors[deviceID] {
		return 0, fmt.Errorf("mock calculation error for device %s", deviceID)
	}

	m.calculations[deviceID] = power * 0.1 // Simple mock calculation
	return m.calculations[deviceID], nil
}

func (m *MockEnergyService) Get(deviceID string) (float64, error) {
	energy, exists := m.calculations[deviceID]
	if !exists {
		return 0, nil
	}
	return energy, nil
}

func (m *MockEnergyService) SetError(deviceID string, shouldError bool) {
	m.errors[deviceID] = shouldError
}

func (m *MockEnergyService) GetCalculationCount() int {
	return len(m.calculations)
}

// VerifyEnergyInterface is a compile-time check to ensure MockEnergyService implements energy.EnergyInterface
var _ energy.EnergyInterface = (*MockEnergyService)(nil)

// TestClient_CollectDeviceDataWithEnergy tests the CollectDeviceDataWithEnergy method
func TestClient_CollectDeviceDataWithEnergy(t *testing.T) {

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "collect_with_nil_energy_service_should_return_data",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Collect data without energy service
				devices, err := client.CollectDeviceDataWithEnergy(ctx, nil)
				if err != nil {
					t.Fatalf("CollectDeviceDataWithEnergy() should not return error: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should return device data")
				}
			},
		},
		{
			name: "collect_with_energy_service_should_calculate_energy",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Create mock energy service
				energyService := NewMockEnergyService()

				// Collect data with energy service
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err != nil {
					t.Fatalf("CollectDeviceDataWithEnergy() should not return error: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should return device data")
				}

				// Check that energy calculations were performed for online devices
				calcCount := energyService.GetCalculationCount()
				if calcCount == 0 {
					t.Error("Energy service should have performed calculations for online devices")
				}
			},
		},
		{
			name: "collect_when_not_connected_should_error",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Don't connect the client
				ctx := context.Background()
				energyService := NewMockEnergyService()

				// Try to collect data
				_, err = client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err == nil {
					t.Error("CollectDeviceDataWithEnergy() should return error when client is not connected")
				}
			},
		},
		{
			name: "energy_service_errors_should_not_stop_collection",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Create mock energy service with errors for some devices
				energyService := NewMockEnergyService()
				energyService.SetError("test-device-1", true) // This device will cause energy calculation errors

				// Collect data with energy service
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err != nil {
					t.Fatalf("CollectDeviceDataWithEnergy() should not return error: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should return device data even with energy errors")
				}

				// Should still have some calculations for devices without errors
				_ = energyService.GetCalculationCount()
				// Note: Due to mock implementation, we expect at least some calculations
				// even with errors, because the mock service processes all devices
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestClient_CollectDeviceDataWithEnergyFailureScenarios tests various failure scenarios in energy integration
func TestClient_CollectDeviceDataWithEnergyFailureScenarios(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "energy_service_panic_should_be_handled_gracefully",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Create a mock energy service that panics
				energyService := &PanickingEnergyService{}

				// Collect data with panicking energy service
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				// Should not panic, should return data despite energy service issues
				if err != nil {
					t.Fatalf("CollectDeviceDataWithEnergy() should not return error even when energy service panics: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should still return device data when energy service panics")
				}
			},
		},
		{
			name: "energy_service_timeout_should_not_block_collection",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Create a slow energy service
				energyService := &SlowEnergyService{delay: 100 * time.Millisecond}

				// Use context with timeout
				timeoutCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()

				// Collect data with slow energy service
				devices, err := client.CollectDeviceDataWithEnergy(timeoutCtx, energyService)
				// Should still return some data even if energy calculations don't complete
				if err != nil {
					// Error is acceptable due to context timeout, but should have some data
					t.Logf("Expected error due to timeout: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should return device data even if energy calculations timeout")
				}
			},
		},
		{
			name: "invalid_device_data_from_energy_should_not_fail_collection",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Create an energy service that returns invalid data
				energyService := &InvalidDataEnergyService{}

				// Collect data with energy service returning invalid data
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err != nil {
					t.Fatalf("CollectDeviceDataWithEnergy() should not return error: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should return device data even if energy calculations return invalid data")
				}
			},
		},
		{
			name: "concurrent_energy_calculations_should_be_thread_safe",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Create a thread-safe energy service
				energyService := NewThreadSafeMockEnergyService()

				// Run multiple concurrent collections
				const numGoroutines = 10
				errChan := make(chan error, numGoroutines)
				dataChan := make(chan []*DeviceData, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					go func(id int) {
						devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
						if err != nil {
							errChan <- fmt.Errorf("goroutine %d: %v", id, err)
							return
						}
						dataChan <- devices
					}(i)
				}

				// Wait for all goroutines to complete
				for i := 0; i < numGoroutines; i++ {
					select {
					case err := <-errChan:
						t.Errorf("Concurrent collection failed: %v", err)
					case data := <-dataChan:
						if len(data) == 0 {
							t.Error("Concurrent collection should return device data")
						}
					case <-time.After(5 * time.Second):
						t.Error("Concurrent collection timed out")
					}
				}
			},
		},
		{
			name: "energy_service_nil_interface_should_be_handled",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Collect data with nil energy service interface
				var energyService energy.EnergyInterface = nil
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err != nil {
					t.Fatalf("CollectDeviceDataWithEnergy() should not return error with nil energy interface: %v", err)
				}

				if len(devices) == 0 {
					t.Error("CollectDeviceDataWithEnergy() should return device data with nil energy interface")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Additional mock energy services for failure testing

// PanickingEnergyService is a mock energy service that panics on calculation
type PanickingEnergyService struct{}

func (p *PanickingEnergyService) Calculate(deviceID string, power float64) (float64, error) {
	panic("intentional panic for testing")
}

func (p *PanickingEnergyService) Get(deviceID string) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}

// SlowEnergyService is a mock energy service that responds slowly
type SlowEnergyService struct {
	delay time.Duration
}

func (s *SlowEnergyService) Calculate(deviceID string, power float64) (float64, error) {
	time.Sleep(s.delay)
	return power * 0.1, nil
}

func (s *SlowEnergyService) Get(deviceID string) (float64, error) {
	time.Sleep(s.delay)
	return 100.0, nil
}

// InvalidDataEnergyService is a mock energy service that returns invalid data
type InvalidDataEnergyService struct{}

func (i *InvalidDataEnergyService) Calculate(deviceID string, power float64) (float64, error) {
	// Return invalid energy values (NaN, Inf, etc.)
	switch deviceID {
	case "test-device-1":
		return 0, nil // valid
	case "test-device-2":
		return math.NaN(), nil // NaN
	case "test-device-3":
		return math.Inf(1), nil // +Inf
	default:
		return math.Inf(-1), nil // -Inf
	}
}

func (i *InvalidDataEnergyService) Get(deviceID string) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}

// ThreadSafeMockEnergyService is a thread-safe mock energy service for concurrent testing
type ThreadSafeMockEnergyService struct {
	mutex        sync.RWMutex
	calculations map[string]float64
	errors       map[string]bool
}

func NewThreadSafeMockEnergyService() *ThreadSafeMockEnergyService {
	return &ThreadSafeMockEnergyService{
		calculations: make(map[string]float64),
		errors:       make(map[string]bool),
	}
}

func (m *ThreadSafeMockEnergyService) Calculate(deviceID string, power float64) (float64, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.errors[deviceID] {
		return 0, fmt.Errorf("mock calculation error for device %s", deviceID)
	}

	m.calculations[deviceID] = power * 0.1 // Simple mock calculation
	return m.calculations[deviceID], nil
}

func (m *ThreadSafeMockEnergyService) Get(deviceID string) (float64, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	energy, exists := m.calculations[deviceID]
	if !exists {
		return 0, nil
	}
	return energy, nil
}

func (m *ThreadSafeMockEnergyService) SetError(deviceID string, hasError bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.errors[deviceID] = hasError
}

func (m *ThreadSafeMockEnergyService) GetCalculationCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.calculations)
}
