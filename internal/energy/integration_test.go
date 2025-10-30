package energy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
)

func TestEnergyService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "energy-integration-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := log.NewTestLogger()

	// Create real storage manager
	storageConfig := &storage.Config{
		DataDir:         tempDir,
		FilePermissions: 0644,
	}
	storageManager, err := storage.NewFileStorageManager(storageConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create energy service
	service := NewEnergyService(storageManager, logger)

	deviceID := "ups-integration-001"
	power := 1000.0 // 1000W

	t.Run("First calculation creates file", func(t *testing.T) {
		energy, err := service.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Calculate failed: %v", err)
		}

		if energy != 0 {
			t.Errorf("Expected energy = 0, got %v", energy)
		}

		// Verify file was created
		filePath := filepath.Join(tempDir, deviceID+".txt")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Expected file to be created")
		}
	})

	t.Run("Second calculation accumulates energy", func(t *testing.T) {
		// Wait for time interval
		time.Sleep(100 * time.Millisecond)

		energy, err := service.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Calculate failed: %v", err)
		}

		// Energy should have increased
		if energy <= 0 {
			t.Errorf("Expected energy > 0, got %v", energy)
		}

		t.Logf("Accumulated energy: %vWh", energy)
	})

	t.Run("Service restart recovers data", func(t *testing.T) {
		// Get current energy
		energy1, err := service.Get(deviceID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Create new service instance (simulate restart)
		newService := NewEnergyService(storageManager, logger)

		// Get energy from new service
		energy2, err := newService.Get(deviceID)
		if err != nil {
			t.Fatalf("Get failed after restart: %v", err)
		}

		// Energy should be the same
		if energy1 != energy2 {
			t.Errorf("Expected energy = %v after restart, got %v", energy1, energy2)
		}

		// Continue calculation with new service
		time.Sleep(100 * time.Millisecond)
		energy3, err := newService.Calculate(deviceID, power)
		if err != nil {
			t.Fatalf("Calculate failed after restart: %v", err)
		}

		// Energy should have increased from previous value
		if energy3 <= energy2 {
			t.Errorf("Expected energy3 > energy2, got energy2=%v, energy3=%v", energy2, energy3)
		}

		t.Logf("Energy after restart and calculation: %vWh", energy3)
	})
}

func TestEnergyService_DataConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "energy-consistency-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := log.NewTestLogger()

	// Create real storage manager
	storageConfig := &storage.Config{
		DataDir:         tempDir,
		FilePermissions: 0644,
	}
	storageManager, err := storage.NewFileStorageManager(storageConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create energy service
	service := NewEnergyService(storageManager, logger)

	deviceID := "ups-consistency-001"

	t.Run("Multiple calculations maintain consistency", func(t *testing.T) {
		powers := []float64{1000.0, 800.0, 1200.0, 900.0}
		delay := 100 * time.Millisecond

		var previousEnergy float64
		for i, power := range powers {
			if i > 0 {
				time.Sleep(delay)
			}

			energy, err := service.Calculate(deviceID, power)
			if err != nil {
				t.Fatalf("Calculate failed at iteration %d: %v", i, err)
			}

			// Verify monotonic increase (for positive power)
			if i > 0 && energy <= previousEnergy {
				t.Errorf("Iteration %d: Energy should increase, got prev=%v, curr=%v",
					i, previousEnergy, energy)
			}

			previousEnergy = energy
			t.Logf("Iteration %d: Power=%vW, Energy=%vWh", i, power, energy)
		}

		// Verify final energy matches stored data
		storedEnergy, err := service.Get(deviceID)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		if storedEnergy != previousEnergy {
			t.Errorf("Expected stored energy = %v, got %v", previousEnergy, storedEnergy)
		}
	})

	t.Run("Stats consistency", func(t *testing.T) {
		stats := service.GetStats()

		// Should have 4 calculations from previous test
		if stats.GetTotalCalculations() < 4 {
			t.Errorf("Expected at least 4 calculations, got %v", stats.GetTotalCalculations())
		}

		// No errors expected
		if stats.GetTotalErrors() != 0 {
			t.Errorf("Expected 0 errors, got %v", stats.GetTotalErrors())
		}

		// Last update time should be recent
		lastUpdate := stats.GetLastUpdateTime()
		if time.Since(lastUpdate) > 5*time.Second {
			t.Errorf("Last update time is too old: %v", lastUpdate)
		}
	})
}

func TestEnergyService_MultiDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "energy-multidevice-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := log.NewTestLogger()

	// Create real storage manager
	storageConfig := &storage.Config{
		DataDir:         tempDir,
		FilePermissions: 0644,
	}
	storageManager, err := storage.NewFileStorageManager(storageConfig, logger)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Create energy service
	service := NewEnergyService(storageManager, logger)

	devices := map[string]float64{
		"ups-001": 1000.0,
		"ups-002": 1500.0,
		"ups-003": 800.0,
	}

	t.Run("Multiple devices maintain separate energy values", func(t *testing.T) {
		// First calculation for all devices
		for deviceID, power := range devices {
			energy, err := service.Calculate(deviceID, power)
			if err != nil {
				t.Fatalf("Calculate failed for %s: %v", deviceID, err)
			}
			if energy != 0 {
				t.Errorf("Expected energy = 0 for %s, got %v", deviceID, energy)
			}
		}

		// Wait and calculate again
		time.Sleep(100 * time.Millisecond)

		energyValues := make(map[string]float64)
		for deviceID, power := range devices {
			energy, err := service.Calculate(deviceID, power)
			if err != nil {
				t.Fatalf("Calculate failed for %s: %v", deviceID, err)
			}
			energyValues[deviceID] = energy
			t.Logf("Device %s: Power=%vW, Energy=%vWh", deviceID, power, energy)
		}

		// Verify all devices have different energy values (based on different powers)
		if energyValues["ups-001"] == energyValues["ups-002"] {
			t.Error("Expected different energy values for different devices")
		}

		// Verify files were created for all devices
		for deviceID := range devices {
			filePath := filepath.Join(tempDir, deviceID+".txt")
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected file for device %s to exist", deviceID)
			}
		}
	})
}
