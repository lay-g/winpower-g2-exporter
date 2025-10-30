package storage

import (
	"os"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func TestNewFileStorageManager(t *testing.T) {
	logger := log.NewTestLogger()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DataDir:         "./data",
				FilePermissions: 0644,
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "invalid config - empty data dir",
			config: &Config{
				DataDir:         "",
				FilePermissions: 0644,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewFileStorageManager(tt.config, logger)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFileStorageManager() error = nil, want error")
				}
				if manager != nil {
					t.Errorf("NewFileStorageManager() manager = %v, want nil", manager)
				}
			} else {
				if err != nil {
					t.Errorf("NewFileStorageManager() error = %v, want nil", err)
				}
				if manager == nil {
					t.Errorf("NewFileStorageManager() manager = nil, want non-nil")
				}
			}
		})
	}
}

func TestFileStorageManager_Write_Read(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}

	manager, err := NewFileStorageManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	// Test data
	now := time.Now().UnixMilli()
	testData := &PowerData{
		Timestamp: now,
		EnergyWH:  1500.75,
	}

	// Test Write
	err = manager.Write("device1", testData)
	if err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Test Read
	readData, err := manager.Read("device1")
	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	if readData.Timestamp != testData.Timestamp {
		t.Errorf("Read() Timestamp = %v, want %v", readData.Timestamp, testData.Timestamp)
	}
	if readData.EnergyWH != testData.EnergyWH {
		t.Errorf("Read() EnergyWH = %v, want %v", readData.EnergyWH, testData.EnergyWH)
	}
}

func TestFileStorageManager_Read_NonExistent(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}

	manager, err := NewFileStorageManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	// Read non-existent device
	data, err := manager.Read("non-existent-device")
	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	// Should return default data
	if data.EnergyWH != 0.0 {
		t.Errorf("Read() EnergyWH = %v, want 0.0", data.EnergyWH)
	}

	// Timestamp should be recent
	now := time.Now().UnixMilli()
	if data.Timestamp > now || data.Timestamp < now-1000 {
		t.Errorf("Read() Timestamp = %v, want recent timestamp", data.Timestamp)
	}
}

func TestFileStorageManager_MultiDevice(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}

	manager, err := NewFileStorageManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	// Test multiple devices
	devices := []struct {
		id   string
		data *PowerData
	}{
		{
			id: "device1",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  1000.0,
			},
		},
		{
			id: "device2",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  2000.5,
			},
		},
		{
			id: "device3",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  3000.25,
			},
		},
	}

	// Write all devices
	for _, device := range devices {
		if err := manager.Write(device.id, device.data); err != nil {
			t.Fatalf("Write(%s) error = %v, want nil", device.id, err)
		}
	}

	// Read and verify all devices
	for _, device := range devices {
		data, err := manager.Read(device.id)
		if err != nil {
			t.Fatalf("Read(%s) error = %v, want nil", device.id, err)
		}

		if data.Timestamp != device.data.Timestamp {
			t.Errorf("Read(%s) Timestamp = %v, want %v", device.id, data.Timestamp, device.data.Timestamp)
		}
		if data.EnergyWH != device.data.EnergyWH {
			t.Errorf("Read(%s) EnergyWH = %v, want %v", device.id, data.EnergyWH, device.data.EnergyWH)
		}
	}
}

func TestFileStorageManager_Update(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	logger := log.NewTestLogger()
	config := &Config{
		DataDir:         tmpDir,
		FilePermissions: 0644,
	}

	manager, err := NewFileStorageManager(config, logger)
	if err != nil {
		t.Fatalf("failed to create storage manager: %v", err)
	}

	deviceID := "device1"

	// Initial write
	initialData := &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  1000.0,
	}
	if err := manager.Write(deviceID, initialData); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Update with new data
	updatedData := &PowerData{
		Timestamp: time.Now().UnixMilli() + 1000,
		EnergyWH:  1500.5,
	}
	if err := manager.Write(deviceID, updatedData); err != nil {
		t.Fatalf("Write() error = %v, want nil", err)
	}

	// Read and verify updated data
	data, err := manager.Read(deviceID)
	if err != nil {
		t.Fatalf("Read() error = %v, want nil", err)
	}

	if data.Timestamp != updatedData.Timestamp {
		t.Errorf("Read() Timestamp = %v, want %v", data.Timestamp, updatedData.Timestamp)
	}
	if data.EnergyWH != updatedData.EnergyWH {
		t.Errorf("Read() EnergyWH = %v, want %v", data.EnergyWH, updatedData.EnergyWH)
	}
}
