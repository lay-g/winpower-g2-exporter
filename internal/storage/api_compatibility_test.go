package storage

import (
	"os"
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

func TestAPICompatibility_EnvironmentVariables(t *testing.T) {
	// Test that environment variable names are consistent
	helpText := GetEnvironmentHelp()

	expectedVars := []string{
		"STORAGE_DATA_DIR",
		"STORAGE_SYNC_WRITE",
		"STORAGE_CREATE_DIR",
		"STORAGE_FILE_PERMISSIONS",
		"STORAGE_DIR_PERMISSIONS",
	}

	for _, envVar := range expectedVars {
		if !containsAPI(helpText, envVar) {
			t.Errorf("Environment variable %s not found in help text", envVar)
		}
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

func TestAPICompatibility_ConfigLoaderInterface(t *testing.T) {
	loader := NewMockConfigLoader()

	// Test all interface methods are implemented
	loader.Set("test_string", "value")
	if value, found := loader.LoadString("test_string"); !found || value != "value" {
		t.Error("LoadString() not working correctly")
	}

	loader.Set("test_int", 42)
	if value, found := loader.LoadInt("test_int"); !found || value != 42 {
		t.Error("LoadInt() not working correctly")
	}

	loader.Set("test_bool", true)
	if value, found := loader.LoadBool("test_bool"); !found || !value {
		t.Error("LoadBool() not working correctly")
	}

	loader.Set("test_slice", []string{"a", "b", "c"})
	if value, found := loader.LoadStringSlice("test_slice"); !found || len(value) != 3 {
		t.Error("LoadStringSlice() not working correctly")
	}
}

func TestAPICompatibility_ParseFileMode(t *testing.T) {
	testCases := []struct {
		input    string
		expected os.FileMode
		hasError bool
	}{
		{"644", 0644, false},
		{"0755", 0755, false},
		{"0600", 0600, false},
		{"777", 0777, false},
		{"", 0000, true},  // Invalid
		{"invalid", 0000, true},  // Invalid
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseFileMode(tc.input)

			if tc.hasError && err == nil {
				t.Errorf("parseFileMode(%s) should return error", tc.input)
			}

			if !tc.hasError {
				if err != nil {
					t.Errorf("parseFileMode(%s) should not return error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("parseFileMode(%s) = %o, expected %o", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestAPICompatibility_InitializeFunctions(t *testing.T) {
	tempDir := t.TempDir()

	// Test InitializeWithDefaults
	manager, err := InitializeWithDefaults()
	if err != nil {
		t.Fatalf("InitializeWithDefaults() error = %v", err)
	}
	if manager == nil {
		t.Fatal("InitializeWithDefaults() returned nil")
	}

	// Test InitializeWithPath
	manager, err = InitializeWithPath(tempDir)
	if err != nil {
		t.Fatalf("InitializeWithPath() error = %v", err)
	}
	if manager == nil {
		t.Fatal("InitializeWithPath() returned nil")
	}

	// Verify the data directory is correct
	if manager.GetDataDir() != tempDir {
		t.Errorf("InitializeWithPath() GetDataDir() = %s, expected %s", manager.GetDataDir(), tempDir)
	}

	// Test InitializeWithLoader
	loader := NewMockConfigLoader()
	loader.Set("data_dir", tempDir)

	manager, err = InitializeWithLoader(loader)
	if err != nil {
		t.Fatalf("InitializeWithLoader() error = %v", err)
	}
	if manager == nil {
		t.Fatal("InitializeWithLoader() returned nil")
	}

	// Verify the data directory from loader
	if manager.GetDataDir() != tempDir {
		t.Errorf("InitializeWithLoader() GetDataDir() = %s, expected %s", manager.GetDataDir(), tempDir)
	}
}

func TestAPICompatibility_ConfigOperations(t *testing.T) {
	config := NewConfig()

	// Test ApplyEnvironmentOverrides method exists
	config.ApplyEnvironmentOverrides()

	// Test that the method doesn't panic
	if config == nil {
		t.Error("ApplyEnvironmentOverrides() should not panic")
	}
}

// Helper function for string contains check
func containsAPI(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsSubstringAPI(s, substr)))
}

func containsSubstringAPI(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}