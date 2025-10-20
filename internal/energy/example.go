package energy

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ExampleUsage demonstrates how to use the energy module.
// This example shows the basic usage patterns for energy calculation.
func ExampleUsage() {
	// Create a logger
	logger, _ := zap.NewDevelopment(
		zap.IncreaseLevel(zapcore.InfoLevel),
	)
	defer func() {
		_ = logger.Sync()
	}()

	// Create energy configuration
	config := DefaultConfig()
	config.Precision = 0.001 // Higher precision
	config.EnableStats = true

	// Create mock storage for demonstration
	storage := &MockStorage{
		data: make(map[string]*PowerData),
	}

	// Create energy service
	energyService := NewEnergyService(storage, logger, config)

	// Example 1: First calculation for a new device
	deviceID := "ups-001"
	power := 500.0 // 500 watts

	totalEnergy, err := energyService.Calculate(deviceID, power)
	if err != nil {
		logger.Error("Failed to calculate energy", zap.Error(err))
		return
	}

	logger.Info("Energy calculation completed",
		zap.String("device_id", deviceID),
		zap.Float64("total_energy", totalEnergy),
		zap.Float64("current_power", power))

	// Example 2: Subsequent calculation
	time.Sleep(100 * time.Millisecond) // Wait a bit
	power = 600.0                      // Increased power

	totalEnergy, err = energyService.Calculate(deviceID, power)
	if err != nil {
		logger.Error("Failed to calculate energy", zap.Error(err))
		return
	}

	logger.Info("Energy calculation completed",
		zap.String("device_id", deviceID),
		zap.Float64("total_energy", totalEnergy),
		zap.Float64("current_power", power))

	// Example 3: Query current energy data
	currentEnergy, err := energyService.Get(deviceID)
	if err != nil {
		logger.Error("Failed to get energy data", zap.Error(err))
		return
	}

	logger.Info("Current energy data",
		zap.String("device_id", deviceID),
		zap.Float64("current_energy", currentEnergy))

	// Example 4: Get statistics
	stats := energyService.GetStats()
	logger.Info("Energy service statistics",
		zap.Int64("total_calculations", stats.TotalCalculations),
		zap.Int64("total_errors", stats.TotalErrors),
		zap.Duration("avg_calculation_time", time.Duration(stats.AvgCalculationTime)))
}

// MockStorage provides a simple in-memory storage implementation for testing and examples.
type MockStorage struct {
	data map[string]*PowerData
}

// Write stores power data for a device.
func (m *MockStorage) Write(deviceID string, data *PowerData) error {
	if m.data == nil {
		m.data = make(map[string]*PowerData)
	}
	m.data[deviceID] = data
	return nil
}

// Read retrieves power data for a device.
func (m *MockStorage) Read(deviceID string) (*PowerData, error) {
	if data, exists := m.data[deviceID]; exists {
		return data, nil
	}
	return nil, nil // No data for new devices
}

// NewMockStorage creates a new mock storage instance.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]*PowerData),
	}
}
