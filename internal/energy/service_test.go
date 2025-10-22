package energy

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMockStorage implements StorageManager for testing
type TestMockStorage struct {
	data     map[string]*PowerData
	readErr  error
	writeErr error
}

func (m *TestMockStorage) Write(deviceID string, data *PowerData) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.data == nil {
		m.data = make(map[string]*PowerData)
	}
	m.data[deviceID] = data
	return nil
}

func (m *TestMockStorage) Read(deviceID string) (*PowerData, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	if data, exists := m.data[deviceID]; exists {
		return data, nil
	}
	return nil, nil // No data for new devices
}

func NewTestMockStorage() *TestMockStorage {
	return &TestMockStorage{
		data: make(map[string]*PowerData),
	}
}

// TestNewEnergyService tests the constructor
func TestNewEnergyService(t *testing.T) {
	logger := newTestLogger(t)
	storage := NewTestMockStorage()

	t.Run("valid inputs", func(t *testing.T) {
		config := DefaultConfig()
		service := NewEnergyService(storage, logger, config)

		assert.NotNil(t, service)
		assert.Equal(t, storage, service.storage)
		assert.Equal(t, logger, service.logger)
		assert.Equal(t, config, service.config)
		assert.NotNil(t, service.stats)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		service := NewEnergyService(storage, logger, nil)

		assert.NotNil(t, service)
		assert.NotNil(t, service.config)
		assert.Equal(t, 0.01, service.config.Precision)
		assert.True(t, service.config.EnableStats)
	})

	t.Run("nil logger uses no-op", func(t *testing.T) {
		service := NewEnergyService(storage, nil, DefaultConfig())

		assert.NotNil(t, service)
		assert.NotNil(t, service.logger)
	})

	t.Run("nil storage", func(t *testing.T) {
		service := NewEnergyService(nil, logger, DefaultConfig())

		assert.NotNil(t, service)
		assert.Nil(t, service.storage)
	})
}

// TestEnergyService_Calculate tests the Calculate method
func TestEnergyService_Calculate(t *testing.T) {
	logger := newTestLogger(t)

	t.Run("valid new device", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		deviceID := "test-device"
		power := 500.0

		totalEnergy, err := service.Calculate(deviceID, power)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, totalEnergy) // First calculation should start from 0

		// Verify data was saved to storage
		savedData, err := storage.Read(deviceID)
		assert.NoError(t, err)
		assert.NotNil(t, savedData)
		assert.Equal(t, 0.0, savedData.EnergyWH)
	})

	t.Run("valid existing device", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		deviceID := "test-device"

		// Set up historical data
		historicalTime := time.Now().Add(-5 * time.Minute)
		_ = storage.Write(deviceID, &PowerData{
			Timestamp: historicalTime.UnixMilli(),
			EnergyWH:  1000.0,
		})

		power := 600.0
		totalEnergy, err := service.Calculate(deviceID, power)

		assert.NoError(t, err)
		assert.Greater(t, totalEnergy, 1000.0) // Should be greater than historical

		// Expected: 1000 + (600W * 5min/60min) = 1000 + 50 = 1050
		expectedEnergy := 1050.0
		assert.InDelta(t, expectedEnergy, totalEnergy, 1.0) // Allow small variance
	})

	t.Run("zero power", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		deviceID := "test-device"
		power := 0.0

		totalEnergy, err := service.Calculate(deviceID, power)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, totalEnergy)
	})

	t.Run("negative power allowed", func(t *testing.T) {
		config := DefaultConfig()
		config.NegativePowerAllowed = true

		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, config)

		deviceID := "test-device"
		power := -100.0

		totalEnergy, err := service.Calculate(deviceID, power)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, totalEnergy) // First calculation
	})

	t.Run("negative power not allowed", func(t *testing.T) {
		config := DefaultConfig()
		config.NegativePowerAllowed = false

		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, config)

		deviceID := "test-device"
		power := -100.0

		totalEnergy, err := service.Calculate(deviceID, power)

		assert.Error(t, err)
		assert.Equal(t, 0.0, totalEnergy)
		assert.Contains(t, err.Error(), "negative power is not allowed")
	})

	t.Run("empty device ID", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		totalEnergy, err := service.Calculate("", 500.0)

		assert.Error(t, err)
		assert.Equal(t, 0.0, totalEnergy)
		assert.Contains(t, err.Error(), "device ID cannot be empty")
	})

	t.Run("NaN power", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		totalEnergy, err := service.Calculate("test-device", math.NaN())

		assert.Error(t, err)
		assert.Equal(t, 0.0, totalEnergy)
		assert.Contains(t, err.Error(), "power value must be a finite number")
	})

	t.Run("Inf power", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		totalEnergy, err := service.Calculate("test-device", math.Inf(1))

		assert.Error(t, err)
		assert.Equal(t, 0.0, totalEnergy)
		assert.Contains(t, err.Error(), "power value must be a finite number")
	})

	t.Run("storage read error", func(t *testing.T) {
		storage := NewTestMockStorage()
		storage.readErr = assert.AnError

		service := NewEnergyService(storage, logger, DefaultConfig())

		totalEnergy, err := service.Calculate("test-device", 500.0)

		assert.Error(t, err)
		assert.Equal(t, 0.0, totalEnergy)
		assert.Contains(t, err.Error(), "failed to load history data")
	})

	t.Run("storage write error", func(t *testing.T) {
		storage := NewTestMockStorage()
		storage.writeErr = assert.AnError

		service := NewEnergyService(storage, logger, DefaultConfig())

		totalEnergy, err := service.Calculate("test-device", 500.0)

		assert.Error(t, err)
		assert.Equal(t, 0.0, totalEnergy)
		assert.Contains(t, err.Error(), "failed to save data")
	})
}

// TestEnergyService_Get tests the Get method
func TestEnergyService_Get(t *testing.T) {
	logger := newTestLogger(t)

	t.Run("existing device", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		deviceID := "test-device"
		expectedEnergy := 1500.75

		// Set up storage data
		_ = storage.Write(deviceID, &PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  expectedEnergy,
		})

		energy, err := service.Get(deviceID)

		assert.NoError(t, err)
		assert.Equal(t, expectedEnergy, energy)
	})

	t.Run("non-existing device", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		energy, err := service.Get("non-existing-device")

		assert.NoError(t, err)
		assert.Equal(t, 0.0, energy) // Should return 0 for new devices
	})

	t.Run("storage read error", func(t *testing.T) {
		storage := NewTestMockStorage()
		storage.readErr = assert.AnError

		service := NewEnergyService(storage, logger, DefaultConfig())

		energy, err := service.Get("test-device")

		assert.Error(t, err)
		assert.Equal(t, 0.0, energy)
		assert.Contains(t, err.Error(), "failed to read from storage")
	})

	t.Run("empty device ID", func(t *testing.T) {
		storage := NewTestMockStorage()
		service := NewEnergyService(storage, logger, DefaultConfig())

		energy, err := service.Get("")

		assert.Error(t, err)
		assert.Equal(t, 0.0, energy)
		assert.Contains(t, err.Error(), "device ID cannot be empty")
	})
}

// TestEnergyService_GetStats tests the GetStats method
func TestEnergyService_GetStats(t *testing.T) {
	logger := newTestLogger(t)
	storage := NewTestMockStorage()

	t.Run("stats enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableStats = true

		service := NewEnergyService(storage, logger, config)

		// Perform some calculations
		_, _ = service.Calculate("device1", 500.0)
		_, _ = service.Calculate("device2", 600.0)

		stats := service.GetStats()

		assert.Equal(t, int64(2), stats.TotalCalculations)
		assert.Equal(t, int64(0), stats.TotalErrors)
		assert.Greater(t, stats.LastUpdateTime, int64(0))
		assert.Greater(t, stats.AvgCalculationTime, int64(0))
	})

	t.Run("stats disabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableStats = false

		service := NewEnergyService(storage, logger, config)

		// Perform calculations
		_, _ = service.Calculate("device1", 500.0)

		stats := service.GetStats()

		assert.Equal(t, int64(0), stats.TotalCalculations)
		assert.Equal(t, int64(0), stats.TotalErrors)
		assert.Equal(t, int64(0), stats.LastUpdateTime)
		assert.Equal(t, int64(0), stats.AvgCalculationTime)
	})
}

// TestEnergyService_ConcurrentAccess tests concurrent access safety
func TestEnergyService_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	logger := newTestLogger(t)
	storage := NewTestMockStorage()
	service := NewEnergyService(storage, logger, DefaultConfig())

	const numGoroutines = 10
	const numCalculationsPerGoroutine = 20

	errors := make(chan error, numGoroutines*numCalculationsPerGoroutine)

	// Start multiple goroutines performing calculations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < numCalculationsPerGoroutine; j++ {
				deviceID := t.Name() + "-device-" + string(rune('A'+goroutineID%5))
				power := float64(100 + j)

				_, err := service.Calculate(deviceID, power)
				if err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(1 * time.Second)
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent calculation error: %v", err)
	}

	// Verify statistics
	stats := service.GetStats()
	expectedCalculations := int64(numGoroutines * numCalculationsPerGoroutine)
	assert.Equal(t, expectedCalculations, stats.TotalCalculations)
	assert.Equal(t, int64(0), stats.TotalErrors)
}
