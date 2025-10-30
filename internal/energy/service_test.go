package energy

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/energy/mocks"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
)

func TestNewEnergyService(t *testing.T) {
	logger := log.NewTestLogger()
	mockStorage := mocks.NewMockStorage()

	t.Run("Success", func(t *testing.T) {
		service := NewEnergyService(mockStorage, logger)
		if service == nil {
			t.Fatal("Expected non-nil service")
		}
		if service.storage == nil {
			t.Error("Expected non-nil storage")
		}
		if service.logger == nil {
			t.Error("Expected non-nil logger")
		}
		if service.stats == nil {
			t.Error("Expected non-nil stats")
		}
	})

	t.Run("Panic on nil storage", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil storage")
			}
		}()
		NewEnergyService(nil, logger)
	})

	t.Run("Panic on nil logger", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil logger")
			}
		}()
		NewEnergyService(mockStorage, nil)
	})
}

func TestEnergyService_Calculate(t *testing.T) {
	logger := log.NewTestLogger()

	t.Run("First calculation starts from 0", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		deviceID := "ups-001"
		power := 500.0

		energy, err := service.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if energy != 0 {
			t.Errorf("Expected energy = 0, got %v", energy)
		}

		// Verify data was saved
		data := mockStorage.GetData()
		if _, exists := data[deviceID]; !exists {
			t.Error("Expected data to be saved")
		}
	})

	t.Run("Sequential calculations accumulate energy", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		deviceID := "ups-001"
		power := 1000.0 // 1000W

		// First calculation
		energy1, err := service.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if energy1 != 0 {
			t.Errorf("Expected energy1 = 0, got %v", energy1)
		}

		// Wait 100ms for time interval
		time.Sleep(100 * time.Millisecond)

		// Second calculation
		energy2, err := service.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Energy should have increased
		if energy2 <= energy1 {
			t.Errorf("Expected energy2 > energy1, got energy1=%v, energy2=%v", energy1, energy2)
		}

		// Verify energy is positive
		if energy2 < 0 {
			t.Errorf("Expected positive energy, got %v", energy2)
		}
	})

	t.Run("Negative power decreases energy", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		deviceID := "ups-002"
		positivePower := 1000.0
		negativePower := -500.0

		// First calculation with positive power
		_, err := service.Calculate(deviceID, positivePower)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Second calculation with positive power to accumulate energy
		energy1, err := service.Calculate(deviceID, positivePower)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Third calculation with negative power
		energy2, err := service.Calculate(deviceID, negativePower)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Energy should have decreased
		if energy2 >= energy1 {
			t.Errorf("Expected energy2 < energy1, got energy1=%v, energy2=%v", energy1, energy2)
		}
	})

	t.Run("Zero power maintains energy", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		deviceID := "ups-003"

		// First calculation with positive power
		_, err := service.Calculate(deviceID, 1000.0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Get current energy
		energy1, err := service.Calculate(deviceID, 1000.0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Calculate with zero power
		energy2, err := service.Calculate(deviceID, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Energy should remain the same
		if energy2 != energy1 {
			t.Errorf("Expected energy2 = energy1, got energy1=%v, energy2=%v", energy1, energy2)
		}
	})

	t.Run("Invalid device ID", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		_, err := service.Calculate("", 100.0)
		if err == nil {
			t.Error("Expected error for empty device ID")
		}
		if !errors.Is(err, ErrInvalidDeviceID) {
			t.Errorf("Expected ErrInvalidDeviceID, got %v", err)
		}
	})

	t.Run("Storage read error", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		readErr := errors.New("read error")
		mockStorage.ReadFunc = func(deviceID string) (*storage.PowerData, error) {
			return nil, readErr
		}

		service := NewEnergyService(mockStorage, logger)

		_, err := service.Calculate("ups-001", 100.0)
		if err == nil {
			t.Error("Expected error from storage")
		}
		if !errors.Is(err, ErrStorageRead) {
			t.Errorf("Expected ErrStorageRead wrapper, got %v", err)
		}
	})

	t.Run("Storage write error", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		writeErr := errors.New("write error")
		mockStorage.WriteFunc = func(deviceID string, data *storage.PowerData) error {
			return writeErr
		}

		service := NewEnergyService(mockStorage, logger)

		_, err := service.Calculate("ups-001", 100.0)
		if err == nil {
			t.Error("Expected error from storage")
		}
		if !errors.Is(err, ErrStorageWrite) {
			t.Errorf("Expected ErrStorageWrite wrapper, got %v", err)
		}
	})
}

func TestEnergyService_Get(t *testing.T) {
	logger := log.NewTestLogger()

	t.Run("Success", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		deviceID := "ups-001"
		expectedEnergy := 123.45

		// Write test data
		err := mockStorage.Write(deviceID, &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  expectedEnergy,
		})
		if err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}

		// Get energy
		energy, err := service.Get(deviceID)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if energy != expectedEnergy {
			t.Errorf("Expected energy = %v, got %v", expectedEnergy, energy)
		}
	})

	t.Run("Device not found", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		_, err := service.Get("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent device")
		}
		if !errors.Is(err, ErrStorageRead) {
			t.Errorf("Expected ErrStorageRead wrapper, got %v", err)
		}
	})

	t.Run("Invalid device ID", func(t *testing.T) {
		mockStorage := mocks.NewMockStorage()
		service := NewEnergyService(mockStorage, logger)

		_, err := service.Get("")
		if err == nil {
			t.Error("Expected error for empty device ID")
		}
		if !errors.Is(err, ErrInvalidDeviceID) {
			t.Errorf("Expected ErrInvalidDeviceID, got %v", err)
		}
	})
}

func TestEnergyService_GetStats(t *testing.T) {
	logger := log.NewTestLogger()
	mockStorage := mocks.NewMockStorage()
	service := NewEnergyService(mockStorage, logger)

	stats := service.GetStats()
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Initial stats
	if stats.GetTotalCalculations() != 0 {
		t.Errorf("Expected TotalCalculations = 0, got %v", stats.GetTotalCalculations())
	}
	if stats.GetTotalErrors() != 0 {
		t.Errorf("Expected TotalErrors = 0, got %v", stats.GetTotalErrors())
	}

	// Perform calculation to update stats
	_, _ = service.Calculate("ups-001", 100.0)

	// Check updated stats
	if stats.GetTotalCalculations() != 1 {
		t.Errorf("Expected TotalCalculations = 1, got %v", stats.GetTotalCalculations())
	}
}

func TestEnergyService_ConcurrentAccess(t *testing.T) {
	logger := log.NewTestLogger()
	mockStorage := mocks.NewMockStorage()
	service := NewEnergyService(mockStorage, logger)

	deviceID := "ups-001"
	power := 1000.0
	goroutines := 10
	iterationsPerGoroutine := 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				_, _ = service.Calculate(deviceID, power)
				time.Sleep(time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Check stats
	stats := service.GetStats()
	totalCalculations := stats.GetTotalCalculations()
	expectedTotal := int64(goroutines * iterationsPerGoroutine)

	if totalCalculations != expectedTotal {
		t.Errorf("Expected TotalCalculations = %v, got %v", expectedTotal, totalCalculations)
	}

	// Verify final energy is positive
	energy, err := service.Get(deviceID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if energy < 0 {
		t.Errorf("Expected positive energy, got %v", energy)
	}
}

func TestEnergyService_SequentialCalculations(t *testing.T) {
	logger := log.NewTestLogger()
	mockStorage := mocks.NewMockStorage()
	service := NewEnergyService(mockStorage, logger)

	deviceID := "ups-001"
	powers := []float64{1000.0, 800.0, 1200.0, 900.0}
	delay := 100 * time.Millisecond

	var previousEnergy float64

	for i, power := range powers {
		if i > 0 {
			time.Sleep(delay)
		}

		energy, err := service.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Calculation %d failed: %v", i, err)
		}

		// Energy should increase with positive power (except first iteration which is 0)
		if i > 0 {
			if energy <= previousEnergy {
				t.Errorf("Iteration %d: Expected energy > %v, got %v", i, previousEnergy, energy)
			}
		}

		previousEnergy = energy

		t.Logf("Iteration %d: Power=%vW, Energy=%vWh", i, power, energy)
	}
}
