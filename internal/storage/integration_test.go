package storage

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestTempDir creates a temporary directory for testing.
func setupTestTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	return tempDir
}

// cleanupTestTempDir removes the temporary directory and all its contents.
func cleanupTestTempDir(t *testing.T, tempDir string) {
	if err := os.RemoveAll(tempDir); err != nil {
		t.Errorf("Failed to clean up temp directory '%s': %v", tempDir, err)
	}
}

func TestFileStorageManager_Integration(t *testing.T) {
	tempDir := setupTestTempDir(t)
	defer cleanupTestTempDir(t, tempDir)

	// Create config with temp directory
	config := NewConfigWithPath(tempDir)
	config.SyncWrite = true
	config.CreateDir = true

	// Create storage manager
	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	deviceID := "test-device-1"
	testData := &PowerData{
		Timestamp: 1694678400000,
		EnergyWH:  1500.75,
	}

	// Test Write
	t.Run("Write", func(t *testing.T) {
		err := manager.Write(deviceID, testData)
		if err != nil {
			t.Errorf("Write() failed: %v", err)
		}

		// Verify file exists
		filePath := manager.GetDeviceFilePath(deviceID)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file '%s' to exist after write", filePath)
		}
	})

	// Test Read
	t.Run("Read", func(t *testing.T) {
		readData, err := manager.Read(deviceID)
		if err != nil {
			t.Errorf("Read() failed: %v", err)
		}

		if readData.Timestamp != testData.Timestamp {
			t.Errorf("Expected Timestamp %d, got %d", testData.Timestamp, readData.Timestamp)
		}

		if readData.EnergyWH != testData.EnergyWH {
			t.Errorf("Expected EnergyWH %f, got %f", testData.EnergyWH, readData.EnergyWH)
		}
	})

	// Test Read non-existent device
	t.Run("ReadNonExistent", func(t *testing.T) {
		nonExistentDevice := "non-existent-device"
		readData, err := manager.Read(nonExistentDevice)
		if err != nil {
			t.Errorf("Read() for non-existent device failed: %v", err)
		}

		if readData == nil {
			t.Error("Expected initialized data for non-existent device, got nil")
			return
		}

		if readData.EnergyWH != 0 {
			t.Errorf("Expected zero energy for non-existent device, got %+v", readData)
		}
	})

	// Test FileExists
	t.Run("FileExists", func(t *testing.T) {
		if !manager.FileExists(deviceID) {
			t.Errorf("Expected FileExists() to return true for device '%s'", deviceID)
		}

		if manager.FileExists("non-existent-device") {
			t.Error("Expected FileExists() to return false for non-existent device")
		}
	})

	// Test ValidateDeviceData
	t.Run("ValidateDeviceData", func(t *testing.T) {
		err := manager.ValidateDeviceData(deviceID)
		if err != nil {
			t.Errorf("ValidateDeviceData() failed: %v", err)
		}

		// Validate non-existent device should not error
		err = manager.ValidateDeviceData("non-existent-device")
		if err != nil {
			t.Errorf("ValidateDeviceData() for non-existent device failed: %v", err)
		}
	})
}

func TestFileStorageManager_MultipleDevices(t *testing.T) {
	tempDir := setupTestTempDir(t)
	defer cleanupTestTempDir(t, tempDir)

	config := NewConfigWithPath(tempDir)
	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	devices := []string{"device-1", "device-2", "device-3"}
	testData := map[string]*PowerData{
		"device-1": {Timestamp: 1694678400000, EnergyWH: 1000.0},
		"device-2": {Timestamp: 1694678401000, EnergyWH: 2000.0},
		"device-3": {Timestamp: 1694678402000, EnergyWH: -500.5},
	}

	// Write data for all devices
	for _, deviceID := range devices {
		data := testData[deviceID]
		err := manager.Write(deviceID, data)
		if err != nil {
			t.Errorf("Write() failed for device '%s': %v", deviceID, err)
		}
	}

	// Read and verify data for all devices
	for _, deviceID := range devices {
		readData, err := manager.Read(deviceID)
		if err != nil {
			t.Errorf("Read() failed for device '%s': %v", deviceID, err)
		}

		expected := testData[deviceID]
		if readData.Timestamp != expected.Timestamp {
			t.Errorf("Device '%s': Expected Timestamp %d, got %d", deviceID, expected.Timestamp, readData.Timestamp)
		}

		if readData.EnergyWH != expected.EnergyWH {
			t.Errorf("Device '%s': Expected EnergyWH %f, got %f", deviceID, expected.EnergyWH, readData.EnergyWH)
		}
	}

	// Verify all files exist
	for _, deviceID := range devices {
		if !manager.FileExists(deviceID) {
			t.Errorf("Expected file to exist for device '%s'", deviceID)
		}
	}
}

func TestFileStorageManager_Persistence(t *testing.T) {
	tempDir := setupTestTempDir(t)
	defer cleanupTestTempDir(t, tempDir)

	deviceID := "persistent-device"
	originalData := &PowerData{
		Timestamp: 1694678400000,
		EnergyWH:  1500.75,
	}

	// First storage manager - write data
	config1 := NewConfigWithPath(tempDir)
	manager1, err := NewFileStorageManager(config1)
	if err != nil {
		t.Fatalf("Failed to create first storage manager: %v", err)
	}

	err = manager1.Write(deviceID, originalData)
	if err != nil {
		t.Fatalf("First Write() failed: %v", err)
	}

	// Create new storage manager - read data
	config2 := NewConfigWithPath(tempDir)
	manager2, err := NewFileStorageManager(config2)
	if err != nil {
		t.Fatalf("Failed to create second storage manager: %v", err)
	}

	readData, err := manager2.Read(deviceID)
	if err != nil {
		t.Fatalf("Second Read() failed: %v", err)
	}

	if readData.Timestamp != originalData.Timestamp {
		t.Errorf("Expected Timestamp %d, got %d", originalData.Timestamp, readData.Timestamp)
	}

	if readData.EnergyWH != originalData.EnergyWH {
		t.Errorf("Expected EnergyWH %f, got %f", originalData.EnergyWH, readData.EnergyWH)
	}
}

func TestFileStorageManager_ErrorHandling(t *testing.T) {
	tempDir := setupTestTempDir(t)
	defer cleanupTestTempDir(t, tempDir)

	// Test with invalid config (empty data dir)
	invalidConfig := &Config{DataDir: ""}
	_, err := NewFileStorageManager(invalidConfig)
	if err == nil {
		t.Error("Expected NewFileStorageManager to fail with invalid config")
	}

	// Test with nil config
	_, err = NewFileStorageManager(nil)
	if err == nil {
		t.Error("Expected NewFileStorageManager to fail with nil config")
	}

	// Test writing to invalid device ID
	validConfig := NewConfigWithPath(tempDir)
	manager, err := NewFileStorageManager(validConfig)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	err = manager.Write("", &PowerData{Timestamp: 123, EnergyWH: 100})
	if err == nil {
		t.Error("Expected Write() to fail with empty device ID")
	}

	err = manager.Write("valid-device", nil)
	if err == nil {
		t.Error("Expected Write() to fail with nil data")
	}

	// Test reading with invalid device ID
	_, err = manager.Read("")
	if err == nil {
		t.Error("Expected Read() to fail with empty device ID")
	}
}

func TestFileStorageManager_FileFormat(t *testing.T) {
	tempDir := setupTestTempDir(t)
	defer cleanupTestTempDir(t, tempDir)

	config := NewConfigWithPath(tempDir)
	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("Failed to create storage manager: %v", err)
	}

	deviceID := "format-test-device"
	testData := &PowerData{
		Timestamp: 1694678400000,
		EnergyWH:  1500.75,
	}

	// Write data
	err = manager.Write(deviceID, testData)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Read raw file content
	content, err := manager.ReadRaw(deviceID)
	if err != nil {
		t.Fatalf("ReadRaw() failed: %v", err)
	}

	expectedContent := "1694678400000\n1500.750000\n"

	if string(content) != expectedContent {
		t.Errorf("Expected file content '%s', got '%s'", expectedContent, string(content))
	}

	// Verify file path format
	expectedPath := filepath.Join(tempDir, deviceID+".txt")
	actualPath := manager.GetDeviceFilePath(deviceID)
	if actualPath != expectedPath {
		t.Errorf("Expected file path '%s', got '%s'", expectedPath, actualPath)
	}
}
