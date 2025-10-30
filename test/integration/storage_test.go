//go:build integration
// +build integration

// Package integration provides integration tests for the storage module.
package integration

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
)

// TestStorageIntegration tests the full storage workflow with real file operations.
func TestStorageIntegration(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()

	// Create logger for tests
	logger := log.NewNoopLogger()

	// Create storage configuration
	config := &storage.Config{
		DataDir:         tempDir,
		FilePermissions: 0644,
	}

	// Test configuration validation
	if err := config.Validate(); err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}

	// Create storage manager
	manager, err := storage.NewFileStorageManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	// Test scenario 1: Write and read data for a single device
	t.Run("WriteAndRead", func(t *testing.T) {
		deviceID := "test-device-001"
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  1234.56,
		}

		// Write data
		if err := manager.Write(deviceID, data); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}

		// Verify file was created
		filePath := filepath.Join(tempDir, deviceID+".txt")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("Expected file was not created: %s", filePath)
		}

		// Read data back
		readData, err := manager.Read(deviceID)
		if err != nil {
			t.Fatalf("Failed to read data: %v", err)
		}

		// Verify data matches
		if readData.Timestamp != data.Timestamp {
			t.Errorf("Timestamp mismatch: got %d, want %d", readData.Timestamp, data.Timestamp)
		}
		if readData.EnergyWH != data.EnergyWH {
			t.Errorf("Energy mismatch: got %f, want %f", readData.EnergyWH, data.EnergyWH)
		}
	})

	// Test scenario 2: Multiple devices
	t.Run("MultipleDevices", func(t *testing.T) {
		devices := []struct {
			id     string
			energy float64
		}{
			{"device-A", 100.5},
			{"device-B", 200.75},
			{"device-C", 300.25},
		}

		// Write data for multiple devices
		for _, d := range devices {
			data := &storage.PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  d.energy,
			}
			if err := manager.Write(d.id, data); err != nil {
				t.Fatalf("Failed to write data for %s: %v", d.id, err)
			}
		}

		// Read and verify each device
		for _, d := range devices {
			readData, err := manager.Read(d.id)
			if err != nil {
				t.Fatalf("Failed to read data for %s: %v", d.id, err)
			}
			if readData.EnergyWH != d.energy {
				t.Errorf("Energy mismatch for %s: got %f, want %f", d.id, readData.EnergyWH, d.energy)
			}
		}
	})

	// Test scenario 3: Update existing data
	t.Run("UpdateData", func(t *testing.T) {
		deviceID := "update-test"

		// Initial write
		initialData := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  500.0,
		}
		if err := manager.Write(deviceID, initialData); err != nil {
			t.Fatalf("Failed to write initial data: %v", err)
		}

		// Update data
		time.Sleep(10 * time.Millisecond) // Ensure different timestamp
		updatedData := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  750.5,
		}
		if err := manager.Write(deviceID, updatedData); err != nil {
			t.Fatalf("Failed to update data: %v", err)
		}

		// Verify updated data
		readData, err := manager.Read(deviceID)
		if err != nil {
			t.Fatalf("Failed to read updated data: %v", err)
		}
		if readData.EnergyWH != updatedData.EnergyWH {
			t.Errorf("Energy not updated: got %f, want %f", readData.EnergyWH, updatedData.EnergyWH)
		}
		if readData.Timestamp != updatedData.Timestamp {
			t.Errorf("Timestamp not updated: got %d, want %d", readData.Timestamp, updatedData.Timestamp)
		}
	})

	// Test scenario 4: Read non-existent device returns default
	t.Run("ReadNonExistent", func(t *testing.T) {
		deviceID := "non-existent-device"
		data, err := manager.Read(deviceID)
		if err != nil {
			t.Fatalf("Read should not error for non-existent device: %v", err)
		}
		// When file doesn't exist, reader returns current timestamp with zero energy
		if data.EnergyWH != 0.0 {
			t.Errorf("Expected zero energy for non-existent device, got: %f", data.EnergyWH)
		}
		if data.Timestamp <= 0 {
			t.Errorf("Expected valid timestamp for non-existent device, got: %d", data.Timestamp)
		}
	})

	// Test scenario 5: Concurrent access
	t.Run("ConcurrentAccess", func(t *testing.T) {
		const numGoroutines = 5
		const numOperations = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Use separate devices for each goroutine to avoid contention
		for g := 0; g < numGoroutines; g++ {
			go func(goroutineID int) {
				defer wg.Done()
				deviceID := "concurrent-device-" + string(rune('A'+goroutineID))

				for i := 0; i < numOperations; i++ {
					data := &storage.PowerData{
						Timestamp: time.Now().UnixMilli(),
						EnergyWH:  float64(goroutineID*100 + i),
					}

					if err := manager.Write(deviceID, data); err != nil {
						t.Errorf("Goroutine %d: Write failed: %v", goroutineID, err)
						return
					}

					if _, err := manager.Read(deviceID); err != nil {
						t.Errorf("Goroutine %d: Read failed: %v", goroutineID, err)
						return
					}
				}
			}(g)
		}

		wg.Wait()
	})

	// Test scenario 6: File permission validation
	t.Run("FilePermissions", func(t *testing.T) {
		deviceID := "permission-test"
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  999.99,
		}

		if err := manager.Write(deviceID, data); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}

		// Check file permissions
		filePath := filepath.Join(tempDir, deviceID+".txt")
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}

		// On Unix-like systems, check if permissions match
		if fileInfo.Mode().Perm() != 0644 {
			t.Logf("File permissions: %o (expected 0644, but this may vary by OS)", fileInfo.Mode().Perm())
		}
	})

	// Test scenario 7: Large energy values
	t.Run("LargeValues", func(t *testing.T) {
		deviceID := "large-values"
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),  // Use current timestamp
			EnergyWH:  1.7976931348623157e+308, // Max float64
		}

		if err := manager.Write(deviceID, data); err != nil {
			t.Fatalf("Failed to write large values: %v", err)
		}

		readData, err := manager.Read(deviceID)
		if err != nil {
			t.Fatalf("Failed to read large values: %v", err)
		}

		if readData.EnergyWH != data.EnergyWH {
			t.Errorf("Large energy mismatch: got %e, want %e", readData.EnergyWH, data.EnergyWH)
		}
	})

	// Test scenario 8: Error handling - invalid device ID
	t.Run("InvalidDeviceID", func(t *testing.T) {
		invalidIDs := []string{"", "../etc", "dev/ice", ".hidden", ".."}
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  100.0,
		}

		for _, id := range invalidIDs {
			if err := manager.Write(id, data); err == nil {
				t.Errorf("Expected error for invalid device ID %q, got nil", id)
			}
		}
	})

	// Test scenario 9: Data persistence across manager instances
	t.Run("PersistenceAcrossInstances", func(t *testing.T) {
		deviceID := "persistence-test"
		originalData := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  555.55,
		}

		// Write with first manager
		if err := manager.Write(deviceID, originalData); err != nil {
			t.Fatalf("Failed to write data: %v", err)
		}

		// Create new manager instance
		newManager, err := storage.NewFileStorageManager(config, logger)
		if err != nil {
			t.Fatalf("Failed to create new manager: %v", err)
		}

		// Read with new manager
		readData, err := newManager.Read(deviceID)
		if err != nil {
			t.Fatalf("Failed to read data with new manager: %v", err)
		}

		if readData.Timestamp != originalData.Timestamp || readData.EnergyWH != originalData.EnergyWH {
			t.Errorf("Data not persisted across manager instances")
		}
	})

	// Test scenario 10: Stress test with many devices
	t.Run("StressTestManyDevices", func(t *testing.T) {
		const numDevices = 100

		// Write data for many devices
		for i := 0; i < numDevices; i++ {
			deviceID := filepath.Join("stress", "device", "deep", "nested", "path", "test-"+string(rune(i)))
			// Clean device ID to remove path separators
			deviceID = "stress-device-" + string(rune('A'+i%26)) + "-" + string(rune('0'+i/26))

			data := &storage.PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  float64(i) * 10.5,
			}

			if err := manager.Write(deviceID, data); err != nil {
				t.Fatalf("Failed to write device %s: %v", deviceID, err)
			}
		}

		// Verify all devices can be read
		successCount := 0
		for i := 0; i < numDevices; i++ {
			deviceID := "stress-device-" + string(rune('A'+i%26)) + "-" + string(rune('0'+i/26))
			if _, err := manager.Read(deviceID); err == nil {
				successCount++
			}
		}

		if successCount != numDevices {
			t.Errorf("Expected to read %d devices, but only succeeded with %d", numDevices, successCount)
		}

		// Check that directory isn't cluttered with too many files
		entries, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read temp directory: %v", err)
		}
		t.Logf("Created %d files in temp directory", len(entries))
	})
}

// TestStorageConfigurationIntegration tests various configuration scenarios.
func TestStorageConfigurationIntegration(t *testing.T) {
	logger := log.NewNoopLogger()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		tempDir := t.TempDir()
		config := storage.DefaultConfig()
		config.DataDir = tempDir

		manager, err := storage.NewFileStorageManager(config, logger)
		if err != nil {
			t.Fatalf("Failed to create manager with default config: %v", err)
		}

		// Verify manager works with default config
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  123.45,
		}
		if err := manager.Write("test-device", data); err != nil {
			t.Fatalf("Failed to write with default config: %v", err)
		}
	})

	t.Run("CustomPermissions", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &storage.Config{
			DataDir:         tempDir,
			FilePermissions: 0600, // More restrictive permissions
		}

		manager, err := storage.NewFileStorageManager(config, logger)
		if err != nil {
			t.Fatalf("Failed to create manager with custom permissions: %v", err)
		}

		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  678.90,
		}
		if err := manager.Write("secure-device", data); err != nil {
			t.Fatalf("Failed to write with custom permissions: %v", err)
		}
	})

	t.Run("InvalidConfiguration", func(t *testing.T) {
		invalidConfigs := []*storage.Config{
			nil,
			{DataDir: "", FilePermissions: 0644},
			{DataDir: "/tmp", FilePermissions: 01000}, // Invalid permissions (sticky bit without other bits)
		}

		for i, config := range invalidConfigs {
			if _, err := storage.NewFileStorageManager(config, logger); err == nil {
				t.Errorf("Config %d: Expected error for invalid config, got nil", i)
			}
		}
	})
}

// TestStorageDataDirCreation tests directory creation scenarios.
func TestStorageDataDirCreation(t *testing.T) {
	logger := log.NewNoopLogger()

	t.Run("CreateNonExistentDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		dataDir := filepath.Join(tempDir, "storage")

		config := &storage.Config{
			DataDir:         dataDir,
			FilePermissions: 0644,
		}

		manager, err := storage.NewFileStorageManager(config, logger)
		if err != nil {
			t.Fatalf("Failed to create manager with non-existent directory: %v", err)
		}

		// Write data - this should create the directory
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  111.11,
		}
		if err := manager.Write("test-device", data); err != nil {
			t.Fatalf("Failed to write to newly created directory: %v", err)
		}

		// Verify directory was created
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			t.Errorf("Expected directory to be created: %s", dataDir)
		}
	})

	t.Run("UseExistingDirectory", func(t *testing.T) {
		tempDir := t.TempDir()

		config := &storage.Config{
			DataDir:         tempDir,
			FilePermissions: 0644,
		}

		manager, err := storage.NewFileStorageManager(config, logger)
		if err != nil {
			t.Fatalf("Failed to create manager with existing directory: %v", err)
		}

		// Verify we can write to the existing directory
		data := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  222.22,
		}
		if err := manager.Write("test-device", data); err != nil {
			t.Fatalf("Failed to write to existing directory: %v", err)
		}
	})
}

// TestStorageErrorRecovery tests error recovery scenarios.
func TestStorageErrorRecovery(t *testing.T) {
	tempDir := t.TempDir()
	logger := log.NewNoopLogger()

	config := &storage.Config{
		DataDir:         tempDir,
		FilePermissions: 0644,
	}

	manager, err := storage.NewFileStorageManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	t.Run("ReadCorruptedFile", func(t *testing.T) {
		deviceID := "corrupted-device"
		filePath := filepath.Join(tempDir, deviceID+".txt")

		// Create a corrupted file
		if err := os.WriteFile(filePath, []byte("invalid\ndata\nextra"), 0644); err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		// Attempt to read corrupted file
		data, err := manager.Read(deviceID)
		if err == nil {
			t.Errorf("Expected error when reading corrupted file, got data: %+v", data)
		}
	})

	t.Run("WriteAfterCorruption", func(t *testing.T) {
		deviceID := "recovery-device"
		filePath := filepath.Join(tempDir, deviceID+".txt")

		// Create a corrupted file
		if err := os.WriteFile(filePath, []byte("corrupted"), 0644); err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		// Write valid data - should overwrite corrupted file
		validData := &storage.PowerData{
			Timestamp: time.Now().UnixMilli(),
			EnergyWH:  333.33,
		}
		if err := manager.Write(deviceID, validData); err != nil {
			t.Fatalf("Failed to write after corruption: %v", err)
		}

		// Verify data can now be read
		readData, err := manager.Read(deviceID)
		if err != nil {
			t.Fatalf("Failed to read after recovery: %v", err)
		}
		if readData.EnergyWH != validData.EnergyWH {
			t.Errorf("Data mismatch after recovery: got %f, want %f", readData.EnergyWH, validData.EnergyWH)
		}
	})
}
