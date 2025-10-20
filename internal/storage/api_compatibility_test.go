package storage

import (
	"testing"
)

func TestAPICompatibility_NewFileWriterWithPath(t *testing.T) {
	tempDir := t.TempDir()

	// Test the new convenience function
	writer := NewFileWriterWithPath(tempDir)
	if writer == nil {
		t.Fatal("NewFileWriterWithPath() returned nil")
	}

	// Test that it works
	deviceID := "test-device"
	data := NewPowerData(100.0)

	err := writer.WriteDeviceFile(deviceID, data)
	if err != nil {
		t.Errorf("WriteDeviceFile() error = %v", err)
	}
}

func TestAPICompatibility_NewFileReaderWithPath(t *testing.T) {
	tempDir := t.TempDir()

	// First write some test data
	writer := NewFileWriterWithPath(tempDir)
	deviceID := "test-device"
	data := NewPowerData(100.0)
	err := writer.WriteDeviceFile(deviceID, data)
	if err != nil {
		t.Fatalf("WriteDeviceFile() error = %v", err)
	}

	// Test the new convenience function
	reader := NewFileReaderWithPath(tempDir)
	if reader == nil {
		t.Fatal("NewFileReaderWithPath() returned nil")
	}

	// Test that it works
	readData, err := reader.ReadAndParse(deviceID)
	if err != nil {
		t.Errorf("ReadAndParse() error = %v", err)
	}

	if readData.EnergyWH != data.EnergyWH {
		t.Error("Read data doesn't match written data")
	}
}

func TestAPICompatibility_NewFileWriterWithConfig(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir
	config.SyncWrite = false

	// Test the new convenience function
	writer := NewFileWriterWithConfig(config)
	if writer == nil {
		t.Fatal("NewFileWriterWithConfig() returned nil")
	}

	// Test that it works
	deviceID := "test-device"
	data := NewPowerData(100.0)

	err := writer.WriteDeviceFile(deviceID, data)
	if err != nil {
		t.Errorf("WriteDeviceFile() error = %v", err)
	}
}

func TestAPICompatibility_NewFileReaderWithConfig(t *testing.T) {
	tempDir := t.TempDir()
	config := NewConfig()
	config.DataDir = tempDir

	// First write some test data
	writer := NewFileWriterWithConfig(config)
	deviceID := "test-device"
	data := NewPowerData(100.0)
	err := writer.WriteDeviceFile(deviceID, data)
	if err != nil {
		t.Fatalf("WriteDeviceFile() error = %v", err)
	}

	// Test the new convenience function
	reader := NewFileReaderWithConfig(config)
	if reader == nil {
		t.Fatal("NewFileReaderWithConfig() returned nil")
	}

	// Test that it works
	readData, err := reader.ReadAndParse(deviceID)
	if err != nil {
		t.Errorf("ReadAndParse() error = %v", err)
	}

	if readData.EnergyWH != data.EnergyWH {
		t.Error("Read data doesn't match written data")
	}
}

func TestAPICompatibility_ModuleConstants(t *testing.T) {
	// Test that module constants are properly defined
	if ModuleVersion == "" {
		t.Error("ModuleVersion should not be empty")
	}

	if ModuleName == "" {
		t.Error("ModuleName should not be empty")
	}

	if ModuleName != "storage" {
		t.Errorf("ModuleName = %s, expected 'storage'", ModuleName)
	}
}

func TestAPICompatibility_NewPowerData(t *testing.T) {
	// Test that NewPowerData only accepts energy value
	data := NewPowerData(100.5)

	if data == nil {
		t.Fatal("NewPowerData() returned nil")
	}

	if data.EnergyWH != 100.5 {
		t.Errorf("NewPowerData() EnergyWH = %f, expected 100.5", data.EnergyWH)
	}

	// Timestamp should be automatically set to current time
	if data.Timestamp == 0 {
		t.Error("NewPowerData() should set timestamp to current time")
	}
}

func TestAPICompatibility_ConfigValidation(t *testing.T) {
	config := NewConfig()

	// Test that default configuration is valid
	if err := config.Validate(); err != nil {
		t.Errorf("Default config should be valid: %v", err)
	}

	// Test that we can clone a config
	cloned := config.Clone()
	// Clone() method should never return nil for valid config
	if cloned.DataDir != config.DataDir {
		t.Error("Cloned config should have same DataDir")
	}

	// Test String method
	configStr := config.String()
	if configStr == "" {
		t.Error("String() should not return empty string")
	}
}

func TestAPICompatibility_FilePermissions(t *testing.T) {
	config := NewConfig()

	// Test default file permissions
	if config.FilePermissions == 0 {
		t.Error("Default FilePermissions should not be 0")
	}

	// Test default directory permissions
	if config.DirPermissions == 0 {
		t.Error("Default DirPermissions should not be 0")
	}

	// Test that permissions are reasonable
	if config.FilePermissions != 0644 {
		t.Errorf("Default FilePermissions = %o, expected 0644", config.FilePermissions)
	}

	if config.DirPermissions != 0755 {
		t.Errorf("Default DirPermissions = %o, expected 0755", config.DirPermissions)
	}
}

func TestAPICompatibility_InitializeFunctions(t *testing.T) {
	tempDir := t.TempDir()

	// Test InitializeWithConfig
	config := NewConfig()
	config.DataDir = tempDir

	manager, err := InitializeWithConfig(config)
	if err != nil {
		t.Fatalf("InitializeWithConfig() error = %v", err)
	}
	if manager == nil {
		t.Fatal("InitializeWithConfig() returned nil")
	}

	// Verify the data directory is correct
	if manager.GetDataDir() != tempDir {
		t.Errorf("InitializeWithConfig() GetDataDir() = %s, expected %s", manager.GetDataDir(), tempDir)
	}

	// Test Initialize with options
	opts := NewInitializeOptions(config)
	opts.ValidateOnInit = false

	manager, err = Initialize(opts)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if manager == nil {
		t.Fatal("Initialize() returned nil")
	}

	// Test error cases
	_, err = Initialize(nil)
	if err == nil {
		t.Error("Initialize(nil) should return error")
	}

	invalidOpts := &InitializeOptions{Config: nil}
	_, err = Initialize(invalidOpts)
	if err == nil {
		t.Error("Initialize with nil config should return error")
	}
}

func TestAPICompatibility_ConfigSetDefaults(t *testing.T) {
	config := &Config{}

	// Test SetDefaults method
	config.SetDefaults()

	// Verify defaults are set
	if config.DataDir == "" {
		t.Error("SetDefaults() should set DataDir")
	}

	if config.FilePermissions == 0 {
		t.Error("SetDefaults() should set FilePermissions")
	}

	if config.DirPermissions == 0 {
		t.Error("SetDefaults() should set DirPermissions")
	}

	// Test that NewConfig sets boolean defaults
	newConfig := NewConfig()
	if !newConfig.SyncWrite {
		t.Error("NewConfig() should set SyncWrite to true")
	}

	if !newConfig.CreateDir {
		t.Error("NewConfig() should set CreateDir to true")
	}
}
