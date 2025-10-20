package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultStorageManager(t *testing.T) {
	manager, err := DefaultStorageManager()
	if err != nil {
		t.Fatalf("DefaultStorageManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("DefaultStorageManager() returned nil")
	}

	// Test basic functionality
	deviceID := "test-device"
	data := NewPowerData(100.0)

	err = manager.Write(deviceID, data)
	if err != nil {
		t.Errorf("Manager Write() error = %v", err)
	}

	readData, err := manager.Read(deviceID)
	if err != nil {
		t.Errorf("Manager Read() error = %v", err)
	}

	if readData.EnergyWH != data.EnergyWH {
		t.Error("Manager read/write consistency failed")
	}
}

func TestStorageManagerWithPath(t *testing.T) {
	tempDir := t.TempDir()

	manager, err := StorageManagerWithPath(tempDir)
	if err != nil {
		t.Fatalf("StorageManagerWithPath() error = %v", err)
	}

	if manager == nil {
		t.Fatal("StorageManagerWithPath() returned nil")
	}

	// Verify data directory
	if manager.GetDataDir() != tempDir {
		t.Errorf("GetDataDir() = %s, expected %s", manager.GetDataDir(), tempDir)
	}

	// Test functionality
	deviceID := "path-test-device"
	data := NewPowerData(200.0)

	err = manager.Write(deviceID, data)
	if err != nil {
		t.Errorf("Manager Write() error = %v", err)
	}

	readData, err := manager.Read(deviceID)
	if err != nil {
		t.Errorf("Manager Read() error = %v", err)
	}

	if readData.EnergyWH != data.EnergyWH {
		t.Error("Manager read/write consistency failed")
	}
}

func TestNewManager(t *testing.T) {
	config := NewConfig()
	config.DataDir = t.TempDir()

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	// Test functionality
	deviceID := "new-manager-test"
	data := NewPowerData(300.0)

	err = manager.Write(deviceID, data)
	if err != nil {
		t.Errorf("Manager Write() error = %v", err)
	}

	readData, err := manager.Read(deviceID)
	if err != nil {
		t.Errorf("Manager Read() error = %v", err)
	}

	if readData.EnergyWH != data.EnergyWH {
		t.Error("Manager read/write consistency failed")
	}
}

func TestFileStorageManager_ListDeviceFiles(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("NewFileStorageManager() error = %v", err)
	}

	// Initially no files
	files, err := manager.ListDeviceFiles()
	if err != nil {
		t.Fatalf("ListDeviceFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("ListDeviceFiles() returned %d files, expected 0", len(files))
	}

	// Create some device files
	deviceIDs := []string{"device1", "device2", "device3"}
	for i, deviceID := range deviceIDs {
		data := NewPowerData(float64((i + 1) * 100))
		err := manager.Write(deviceID, data)
		if err != nil {
			t.Fatalf("Write() error for %s: %v", deviceID, err)
		}
	}

	// List files again
	files, err = manager.ListDeviceFiles()
	if err != nil {
		t.Fatalf("ListDeviceFiles() error = %v", err)
	}
	if len(files) != len(deviceIDs) {
		t.Errorf("ListDeviceFiles() returned %d files, expected %d", len(files), len(deviceIDs))
	}

	// Verify file names
	expectedFiles := make(map[string]bool)
	for _, deviceID := range deviceIDs {
		expectedFiles[filepath.Join(tempDir, deviceID+".txt")] = true
	}

	for _, file := range files {
		if !expectedFiles[file] {
			t.Errorf("Unexpected file in list: %s", file)
		}
	}
}

func TestFileStorageManager_ListDeviceFiles_WithNonDataFiles(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("NewFileStorageManager() error = %v", err)
	}

	// Create device files
	data := NewPowerData(100.0)
	err = manager.Write("device1", data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Create non-device files in the same directory
	nonDeviceFiles := []string{"readme.txt", "config.yaml", "data.json", "backup"}
	for _, filename := range nonDeviceFiles {
		filePath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// List files - should only return .txt files that match device pattern
	files, err := manager.ListDeviceFiles()
	if err != nil {
		t.Fatalf("ListDeviceFiles() error = %v", err)
	}

	// Should only have the device file, not other .txt files
	if len(files) != 1 {
		t.Errorf("ListDeviceFiles() returned %d files, expected 1 (only device files)", len(files))
	}

	expectedDeviceFile := filepath.Join(tempDir, "device1.txt")
	if len(files) == 1 && files[0] != expectedDeviceFile {
		t.Errorf("Expected device file %s, got %s", expectedDeviceFile, files[0])
	}
}

func TestFileStorageManager_ListDeviceFiles_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("NewFileStorageManager() error = %v", err)
	}

	// List files in empty directory
	files, err := manager.ListDeviceFiles()
	if err != nil {
		t.Fatalf("ListDeviceFiles() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("ListDeviceFiles() returned %d files in empty directory, expected 0", len(files))
	}
}

func TestFileStorageManager_ListDeviceFiles_NonExistentDirectory(t *testing.T) {
	config := NewConfig()
	config.DataDir = "/non/existent/directory"
	config.CreateDir = false // Don't create directory

	manager, err := NewFileStorageManager(config)
	if err == nil {
		t.Error("NewFileStorageManager() should return error for non-existent directory")
	}

	// Even if manager creation succeeded somehow, listing should fail
	if manager != nil {
		files, err := manager.ListDeviceFiles()
		if err == nil {
			t.Error("ListDeviceFiles() should return error for non-existent directory")
		}
		if files != nil {
			t.Errorf("ListDeviceFiles() should return nil files for non-existent directory, got %v", files)
		}
	}
}

func TestFileStorageManager_GetConfig(t *testing.T) {
	config := NewConfig()
	config.DataDir = t.TempDir()
	config.SyncWrite = false
	config.FilePermissions = 0600
	config.DirPermissions = 0700

	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("NewFileStorageManager() error = %v", err)
	}

	retrievedConfig := manager.GetConfig()
	if retrievedConfig == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Verify values
	if retrievedConfig.DataDir != config.DataDir {
		t.Errorf("GetConfig() DataDir = %s, expected %s", retrievedConfig.DataDir, config.DataDir)
	}

	if retrievedConfig.SyncWrite != config.SyncWrite {
		t.Errorf("GetConfig() SyncWrite = %v, expected %v", retrievedConfig.SyncWrite, config.SyncWrite)
	}

	if retrievedConfig.FilePermissions != config.FilePermissions {
		t.Errorf("GetConfig() FilePermissions = %o, expected %o", retrievedConfig.FilePermissions, config.FilePermissions)
	}

	if retrievedConfig.DirPermissions != config.DirPermissions {
		t.Errorf("GetConfig() DirPermissions = %o, expected %o", retrievedConfig.DirPermissions, config.DirPermissions)
	}

	// Verify it returns a copy, not the original
	retrievedConfig.DataDir = "/modified/path"
	if manager.GetConfig().DataDir == "/modified/path" {
		t.Error("GetConfig() should return a copy, not the original config")
	}
}

func TestFileStorageManager_GetDataDir(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("NewFileStorageManager() error = %v", err)
	}

	dataDir := manager.GetDataDir()
	if dataDir != tempDir {
		t.Errorf("GetDataDir() = %s, expected %s", dataDir, tempDir)
	}
}

func TestFileStorageManager_GetDeviceFilePath(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := NewFileStorageManager(config)
	if err != nil {
		t.Fatalf("NewFileStorageManager() error = %v", err)
	}

	testCases := []struct {
		deviceID string
		expected string
	}{
		{"device1", tempDir + "/device1.txt"},
		{"device-with-dash", tempDir + "/device-with-dash.txt"},
		{"device_with_underscore", tempDir + "/device_with_underscore.txt"},
		{"device123", tempDir + "/device123.txt"},
	}

	for _, tc := range testCases {
		t.Run(tc.deviceID, func(t *testing.T) {
			result := manager.GetDeviceFilePath(tc.deviceID)
			if result != tc.expected {
				t.Errorf("GetDeviceFilePath(%s) = %s, expected %s", tc.deviceID, result, tc.expected)
			}
		})
	}
}

func TestCreateInitializedData(t *testing.T) {
	data := CreateInitializedData()
	if data == nil {
		t.Fatal("CreateInitializedData() returned nil")
	}

	if !data.IsZero() {
		t.Error("CreateInitializedData() should return zero data")
	}

	if data.Timestamp != 0 {
		t.Errorf("CreateInitializedData() Timestamp = %d, expected 0", data.Timestamp)
	}

	if data.EnergyWH != 0 {
		t.Errorf("CreateInitializedData() EnergyWH = %f, expected 0", data.EnergyWH)
	}
}

func TestIsZeroData(t *testing.T) {
	// Test with zero data
	zeroData := &PowerData{Timestamp: 0, EnergyWH: 0}
	if !IsZeroData(zeroData) {
		t.Error("IsZeroData() should return true for zero data")
	}

	// Test with non-zero timestamp
	nonZeroTimestamp := &PowerData{Timestamp: 123, EnergyWH: 0}
	if IsZeroData(nonZeroTimestamp) {
		t.Error("IsZeroData() should return false for non-zero timestamp")
	}

	// Test with non-zero energy
	nonZeroEnergy := &PowerData{Timestamp: 0, EnergyWH: 123.45}
	if IsZeroData(nonZeroEnergy) {
		t.Error("IsZeroData() should return false for non-zero energy")
	}

	// Test with nil data
	if IsZeroData(nil) {
		t.Error("IsZeroData() should return false for nil data")
	}
}
