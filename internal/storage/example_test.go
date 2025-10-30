package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// Example demonstrates basic usage of the storage module.
func Example() {
	// Create a temporary directory for this example
	tmpDir, _ := os.MkdirTemp("", "storage-example-*")
	defer os.RemoveAll(tmpDir)

	// Create logger
	logger := log.NewTestLogger()

	// Create storage manager with custom config
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}
	manager, err := NewFileStorageManager(config, logger)
	if err != nil {
		fmt.Printf("failed to create manager: %v\n", err)
		return
	}

	// Write device data
	data := &PowerData{
		Timestamp: time.Date(2024, 10, 30, 10, 0, 0, 0, time.UTC).UnixMilli(),
		EnergyWH:  1234.5,
	}
	err = manager.Write("device-001", data)
	if err != nil {
		fmt.Printf("failed to write: %v\n", err)
		return
	}

	// Read device data
	retrieved, err := manager.Read("device-001")
	if err != nil {
		fmt.Printf("failed to read: %v\n", err)
		return
	}

	fmt.Printf("Energy: %.1f WH\n", retrieved.EnergyWH)
	// Output: Energy: 1234.5 WH
}

// Example_defaultConfiguration shows how to use default configuration.
func Example_defaultConfiguration() {
	// Use default configuration
	config := DefaultConfig()
	logger := log.NewTestLogger()

	// Validate configuration before use
	if err := config.Validate(); err != nil {
		fmt.Printf("invalid config: %v\n", err)
		return
	}

	manager, _ := NewFileStorageManager(config, logger)

	fmt.Printf("Data directory: %s\n", config.DataDir)
	fmt.Printf("File permissions: %o\n", config.FilePermissions)
	fmt.Printf("Manager created: %v\n", manager != nil)

	// Output:
	// Data directory: ./data
	// File permissions: 644
	// Manager created: true
}

// Example_errorHandling demonstrates error handling patterns.
func Example_errorHandling() {
	tmpDir, _ := os.MkdirTemp("", "storage-error-example-*")
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}
	manager, _ := NewFileStorageManager(config, logger)

	// Try to write invalid data - negative timestamp
	invalidData1 := &PowerData{
		Timestamp: -1,
		EnergyWH:  100,
	}
	err := manager.Write("device-001", invalidData1)
	if err != nil {
		fmt.Println("Invalid timestamp rejected")
	}

	// Try to write invalid data - negative energy
	invalidData2 := &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  -100,
	}
	err = manager.Write("device-002", invalidData2)
	if err != nil {
		fmt.Println("Negative energy rejected")
	}

	// Try invalid device ID
	validData := &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  100,
	}
	err = manager.Write("../invalid", validData)
	if err != nil {
		fmt.Println("Invalid device ID rejected")
	}

	// Output:
	// Invalid timestamp rejected
	// Negative energy rejected
	// Invalid device ID rejected
}

// Example_multipleDevices shows how to manage multiple devices.
func Example_multipleDevices() {
	tmpDir, _ := os.MkdirTemp("", "storage-multi-example-*")
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}
	manager, _ := NewFileStorageManager(config, logger)

	timestamp := time.Date(2024, 10, 30, 10, 0, 0, 0, time.UTC).UnixMilli()

	// Write data for multiple devices
	devices := []struct {
		id     string
		energy float64
	}{
		{"device-001", 1234.5},
		{"device-002", 2345.6},
		{"device-003", 3456.7},
	}

	for _, device := range devices {
		data := &PowerData{
			Timestamp: timestamp,
			EnergyWH:  device.energy,
		}
		manager.Write(device.id, data)
	}

	// Read back and display
	for _, device := range devices {
		data, _ := manager.Read(device.id)
		fmt.Printf("%s: %.1f WH\n", device.id, data.EnergyWH)
	}

	// Check files were created
	files, _ := filepath.Glob(filepath.Join(tmpDir, "*.txt"))
	fmt.Printf("Total files: %d\n", len(files))

	// Output:
	// device-001: 1234.5 WH
	// device-002: 2345.6 WH
	// device-003: 3456.7 WH
	// Total files: 3
}

// Example_dataValidation shows PowerData validation.
func Example_dataValidation() {
	// Valid data
	validData := &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  1234.5,
	}
	if err := validData.Validate(); err == nil {
		fmt.Println("Valid data passed")
	}

	// Invalid: negative timestamp
	invalidData1 := &PowerData{
		Timestamp: -1,
		EnergyWH:  100,
	}
	if err := invalidData1.Validate(); err != nil {
		fmt.Println("Negative timestamp rejected")
	}

	// Invalid: negative energy
	invalidData2 := &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  -100,
	}
	if err := invalidData2.Validate(); err != nil {
		fmt.Println("Negative energy rejected")
	}

	// Output:
	// Valid data passed
	// Negative timestamp rejected
	// Negative energy rejected
}

// Example_updateData shows how to update existing device data.
func Example_updateData() {
	tmpDir, _ := os.MkdirTemp("", "storage-update-example-*")
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}
	manager, _ := NewFileStorageManager(config, logger)

	// Write initial data
	initialData := &PowerData{
		Timestamp: time.Date(2024, 10, 30, 10, 0, 0, 0, time.UTC).UnixMilli(),
		EnergyWH:  1000.0,
	}
	manager.Write("device-001", initialData)

	// Read and display
	data, _ := manager.Read("device-001")
	fmt.Printf("Initial: %.1f WH\n", data.EnergyWH)

	// Update with new value
	updatedData := &PowerData{
		Timestamp: time.Date(2024, 10, 30, 11, 0, 0, 0, time.UTC).UnixMilli(),
		EnergyWH:  1500.0,
	}
	manager.Write("device-001", updatedData)

	// Read again
	data, _ = manager.Read("device-001")
	fmt.Printf("Updated: %.1f WH\n", data.EnergyWH)

	// Output:
	// Initial: 1000.0 WH
	// Updated: 1500.0 WH
}
