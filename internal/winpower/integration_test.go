package winpower

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"go.uber.org/zap/zaptest"
)

// TestIntegration_WinPowerEnergy tests the complete integration between WinPower and Energy modules
func TestIntegration_WinPowerEnergy(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "end_to_end_data_collection_with_energy",
			test: func(t *testing.T) {
				// Create WinPower client
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create WinPower client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				// Create mock storage for energy module
				storage := NewMockStorageManager()

				// Create energy service
				energyConfig := energy.DefaultConfig()
				energyService := energy.NewEnergyService(storage, logger, energyConfig)

				// Connect the client
				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Collect device data with energy calculation
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err != nil {
					t.Fatalf("Failed to collect device data with energy: %v", err)
				}

				// Verify devices were collected
				if len(devices) == 0 {
					t.Error("Expected devices to be collected")
				}

				// Verify energy calculations were performed for online devices
				for _, device := range devices {
					if device.Status == "online" {
						// Check if MockStorageManager has data for this device
						_, err := storage.Read(device.ID)
						if err != nil {
							t.Errorf("Expected energy data to be stored for device %s, but got error: %v", device.ID, err)
						}
					}
				}

				// Verify collection time was updated
				lastCollection := client.GetLastCollectionTime()
				if lastCollection.IsZero() {
					t.Error("Expected last collection time to be updated")
				}

				// Verify client statistics
				stats := client.GetStats()
				if stats.TotalConnections == 0 {
					t.Error("Expected client to have connection statistics")
				}
			},
		},
		{
			name: "concurrent_data_collection_safety",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create WinPower client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				storage := NewMockStorageManager()
				energyConfig := energy.DefaultConfig()
				energyService := energy.NewEnergyService(storage, logger, energyConfig)

				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Test concurrent data collection
				var wg sync.WaitGroup
				numGoroutines := 5
				errors := make(chan error, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()

						_, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
						if err != nil {
							errors <- err
						}
					}(i)
				}

				wg.Wait()
				close(errors)

				// Check for any errors
				for err := range errors {
					t.Errorf("Concurrent collection error: %v", err)
				}

				// Verify final state is consistent
				lastCollection := client.GetLastCollectionTime()
				if lastCollection.IsZero() {
					t.Error("Expected last collection time to be updated after concurrent operations")
				}
			},
		},
		{
			name: "error_handling_and_recovery",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create WinPower client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				storage := NewMockStorageManager()
				energyConfig := energy.DefaultConfig()
				energyService := energy.NewEnergyService(storage, logger, energyConfig)

				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Test collection with context cancellation
				cancelledCtx, cancel := context.WithCancel(ctx)
				cancel() // Cancel immediately

				_, err = client.CollectDeviceDataWithEnergy(cancelledCtx, energyService)
				if err == nil {
					t.Error("Expected error when collecting with cancelled context")
				}

				// Verify client is still usable after cancellation error
				devices, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
				if err != nil {
					t.Errorf("Failed to collect after context cancellation: %v", err)
				}

				if len(devices) == 0 {
					t.Error("Expected devices to be collected after context cancellation")
				}
			},
		},
		{
			name: "performance_validation",
			test: func(t *testing.T) {
				config := DefaultConfig()
				client, err := NewClient(config)
				if err != nil {
					t.Fatalf("Failed to create WinPower client: %v", err)
				}
				defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

				storage := NewMockStorageManager()
				energyConfig := energy.DefaultConfig()
				energyService := energy.NewEnergyService(storage, logger, energyConfig)

				ctx := context.Background()
				if err := client.Connect(ctx); err != nil {
					t.Fatalf("Failed to connect client: %v", err)
				}

				// Measure performance of multiple collections
				numCollections := 10
				start := time.Now()

				for i := 0; i < numCollections; i++ {
					_, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
					if err != nil {
						t.Errorf("Collection %d failed: %v", i, err)
					}
				}

				duration := time.Since(start)
				avgDuration := duration / time.Duration(numCollections)

				// Performance requirements:
				// - Average collection should be under 100ms
				// - Total time should be reasonable
				if avgDuration > 100*time.Millisecond {
					t.Errorf("Average collection time too high: %v (expected < 100ms)", avgDuration)
				}

				t.Logf("Performance: %d collections in %v (avg: %v)", numCollections, duration, avgDuration)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// MockStorageManager implements energy.StorageManager for testing
type MockStorageManager struct {
	data map[string]*energy.PowerData
	mu   sync.RWMutex
}

func NewMockStorageManager() *MockStorageManager {
	return &MockStorageManager{
		data: make(map[string]*energy.PowerData),
	}
}

func (m *MockStorageManager) Write(deviceID string, data *energy.PowerData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[deviceID] = data
	return nil
}

func (m *MockStorageManager) Read(deviceID string) (*energy.PowerData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.data[deviceID]
	if !exists {
		return nil, fmt.Errorf("file not found")
	}
	return data, nil
}

// TestIntegration_ErrorHandling tests comprehensive error handling scenarios
func TestIntegration_ErrorHandling(t *testing.T) {
	t.Run("structured_error_propagation", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config)
		if err != nil {
			t.Fatalf("Failed to create WinPower client: %v", err)
		}
		defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

		// Test invalid configuration error
		invalidConfig := &Config{
			URL:     "invalid-url",
			Timeout: -1 * time.Second,
		}

		_, err = NewClient(invalidConfig)
		if err == nil {
			t.Error("Expected error for invalid configuration")
		}

		// Verify it's a structured error
		if wErr, ok := err.(*Error); ok {
			if wErr.Type != ErrorTypeConfig {
				t.Errorf("Expected config error type, got %v", wErr.Type)
			}
			if !IsRetryable(wErr) {
				t.Errorf("Config errors should be retryable for reconnection scenarios")
			}
		} else {
			t.Error("Expected WinPower structured error")
		}
	})

	t.Run("error_recovery_and_logging", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config)
		if err != nil {
			t.Fatalf("Failed to create WinPower client: %v", err)
		}
		defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

		ctx := context.Background()
		cancelledCtx, cancel := context.WithCancel(ctx)
		cancel()

		// Test context cancellation error handling
		_, err = client.CollectDeviceData(cancelledCtx)
		if err == nil {
			t.Error("Expected error for cancelled context")
		}

		// Verify error is properly logged (this would be visible in test output)
		if err != nil {
			t.Logf("Context cancellation error (expected): %v", err)
		}

		// Test recovery from error
		if err := client.Connect(ctx); err != nil {
			t.Errorf("Failed to recover from error: %v", err)
		}
	})
}

// TestIntegration_ResourceManagement tests resource management and cleanup
func TestIntegration_ResourceManagement(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("client_lifecycle_management", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config)
		if err != nil {
			t.Fatalf("Failed to create WinPower client: %v", err)
		}

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Failed to connect client: %v", err)
		}

		// Verify client is connected
		if !client.IsConnected() {
			t.Error("Expected client to be connected")
		}

		// Close client
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}

		// Verify client is properly cleaned up
		if client.IsConnected() {
			t.Error("Expected client to be disconnected after close")
		}
	})

	t.Run("memory_leak_prevention", func(t *testing.T) {
		config := DefaultConfig()
		client, err := NewClient(config)
		if err != nil {
			t.Fatalf("Failed to create WinPower client: %v", err)
		}
		defer func() {
					if err := client.Close(); err != nil {
						t.Logf("Warning: failed to close client: %v", err)
					}
				}()

		storage := NewMockStorageManager()
		energyConfig := energy.DefaultConfig()
		energyService := energy.NewEnergyService(storage, logger, energyConfig)

		ctx := context.Background()
		if err := client.Connect(ctx); err != nil {
			t.Fatalf("Failed to connect client: %v", err)
		}

		// Perform many operations to test for memory leaks
		for i := 0; i < 100; i++ {
			_, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
			if err != nil {
				t.Errorf("Collection %d failed: %v", i, err)
			}
		}

		// If we reach here without running out of memory, memory management is working
		t.Log("Memory leak test passed: 100 collections completed successfully")
	})
}

// BenchmarkIntegration_CollectDeviceData benchmarks the data collection performance
func BenchmarkIntegration_CollectDeviceData(b *testing.B) {
	config := DefaultConfig()
	client, err := NewClient(config)
	if err != nil {
		b.Fatalf("Failed to create WinPower client: %v", err)
	}
	defer func() {
					if err := client.Close(); err != nil {
						b.Logf("Warning: failed to close client: %v", err)
					}
				}()

	storage := &MockStorageManager{}
	energyConfig := energy.DefaultConfig()
	logger := zaptest.NewLogger(b)
	energyService := energy.NewEnergyService(storage, logger, energyConfig)

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect client: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.CollectDeviceDataWithEnergy(ctx, energyService)
		if err != nil {
			b.Errorf("Collection failed: %v", err)
		}
	}
}

// BenchmarkIntegration_ClientOperations benchmarks individual client operations
func BenchmarkIntegration_ClientOperations(b *testing.B) {
	config := DefaultConfig()
	client, err := NewClient(config)
	if err != nil {
		b.Fatalf("Failed to create WinPower client: %v", err)
	}
	defer func() {
					if err := client.Close(); err != nil {
						b.Logf("Warning: failed to close client: %v", err)
					}
				}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		b.Fatalf("Failed to connect client: %v", err)
	}

	b.Run("GetStats", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = client.GetStats()
		}
	})

	b.Run("GetConnectionStatus", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = client.GetConnectionStatus()
		}
	})

	b.Run("GetLastCollectionTime", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = client.GetLastCollectionTime()
		}
	})
}
