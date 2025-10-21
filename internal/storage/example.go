package storage

import (
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the storage module.
// This example shows basic operations for storing and retrieving power data.
func ExampleUsage() {
	// Create a storage manager with default configuration
	manager, err := DefaultStorageManager()
	if err != nil {
		log.Fatalf("Failed to create storage manager: %v", err)
	}

	// Define a device ID and some test data
	deviceID := "device-001"
	powerData := NewPowerData(1500.75) // 1500.75 Wh

	// Write power data for the device
	fmt.Printf("Writing power data for device %s: %.2f Wh\n", deviceID, powerData.EnergyWH)
	if err := manager.Write(deviceID, powerData); err != nil {
		log.Fatalf("Failed to write power data: %v", err)
	}

	// Read power data back
	readData, err := manager.Read(deviceID)
	if err != nil {
		log.Fatalf("Failed to read power data: %v", err)
	}

	fmt.Printf("Read power data for device %s: %.2f Wh (timestamp: %d)\n",
		deviceID, readData.EnergyWH, readData.Timestamp)

	// Check if file exists
	if manager.FileExists(deviceID) {
		fmt.Printf("Data file exists for device %s\n", deviceID)
	}

	// Try reading from a non-existent device (should return initialized data)
	nonExistentDevice := "device-999"
	initData, err := manager.Read(nonExistentDevice)
	if err != nil {
		log.Fatalf("Failed to read initialized data: %v", err)
	}

	fmt.Printf("Initialized data for new device %s: %.2f Wh\n", nonExistentDevice, initData.EnergyWH)
}

// ExampleWithCustomConfig demonstrates how to use the storage module with custom configuration.
func ExampleWithCustomConfig() {
	// Create custom configuration
	config := NewConfigWithPath("./custom-data")
	config.SyncWrite = true
	config.FilePermissions = 0600 // More restrictive permissions
	config.CreateDir = true

	// Create storage manager with custom config
	manager, err := NewFileStorageManager(config)
	if err != nil {
		log.Fatalf("Failed to create storage manager: %v", err)
	}

	fmt.Printf("Created storage manager with data directory: %s\n", manager.GetDataDir())

	// Use the manager as needed
	deviceID := "device-002"
	powerData := NewPowerData(2500.50)

	if err := manager.Write(deviceID, powerData); err != nil {
		log.Fatalf("Failed to write power data: %v", err)
	}

	fmt.Printf("Successfully wrote data to custom directory for device %s\n", deviceID)
}

// ExampleMock demonstrates how to use the MockStorageManager for testing.
// TODO: Fix this example - NewMockStorageManager is in mock_test.go and not accessible here
/*
func ExampleMock() {
	// Create a mock storage manager for testing
	mock := NewMockStorageManager()

	deviceID := "test-device"
	testData := &PowerData{
		Timestamp: 1694678400000,
		EnergyWH:  1000.0,
	}

	// Write data using mock
	if err := mock.Write(deviceID, testData); err != nil {
		log.Fatalf("Failed to write to mock: %v", err)
	}

	// Read data back
	readData, err := mock.Read(deviceID)
	if err != nil {
		log.Fatalf("Failed to read from mock: %v", err)
	}

	fmt.Printf("Mock test - Read data: %.2f Wh\n", readData.EnergyWH)

	// Verify mock was called
	if mock.WasWriteCalled(deviceID) {
		fmt.Printf("Write was called for device %s\n", deviceID)
	}

	if mock.WasReadCalled(deviceID) {
		fmt.Printf("Read was called for device %s\n", deviceID)
	}

	// Check call counts
	fmt.Printf("Total write calls: %d\n", mock.GetWriteCallCount())
	fmt.Printf("Total read calls: %d\n", mock.GetReadCallCount())
}
*/
