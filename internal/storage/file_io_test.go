package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func TestFileWriter_Write(t *testing.T) {
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

	writer := NewFileWriter(config, logger)

	tests := []struct {
		name     string
		deviceID string
		data     *PowerData
		wantErr  bool
	}{
		{
			name:     "write valid data",
			deviceID: "device1",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  1000.50,
			},
			wantErr: false,
		},
		{
			name:     "write another device",
			deviceID: "device2",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  2000.75,
			},
			wantErr: false,
		},
		{
			name:     "invalid device ID",
			deviceID: "",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  100.0,
			},
			wantErr: true,
		},
		{
			name:     "invalid data - negative energy",
			deviceID: "device3",
			data: &PowerData{
				Timestamp: time.Now().UnixMilli(),
				EnergyWH:  -100.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writer.Write(tt.deviceID, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Write() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("Write() error = %v, want nil", err)
					return
				}

				// Verify file was created
				filePath := filepath.Join(tmpDir, tt.deviceID+".txt")
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("file was not created: %s", filePath)
				}
			}
		})
	}
}

func TestFileReader_Read(t *testing.T) {
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

	reader := NewFileReader(config, logger)
	writer := NewFileWriter(config, logger)

	// Setup: Write some test data
	testData := &PowerData{
		Timestamp: 1694678400000,
		EnergyWH:  15000.50,
	}
	if err := writer.Write("device1", testData); err != nil {
		t.Fatalf("failed to write test data: %v", err)
	}

	// Create a corrupted file
	corruptedFile := filepath.Join(tmpDir, "corrupted.txt")
	if err := os.WriteFile(corruptedFile, []byte("invalid\n"), 0644); err != nil {
		t.Fatalf("failed to create corrupted file: %v", err)
	}

	tests := []struct {
		name     string
		deviceID string
		wantErr  bool
		validate func(t *testing.T, data *PowerData)
	}{
		{
			name:     "read existing device",
			deviceID: "device1",
			wantErr:  false,
			validate: func(t *testing.T, data *PowerData) {
				if data.Timestamp != testData.Timestamp {
					t.Errorf("Timestamp = %v, want %v", data.Timestamp, testData.Timestamp)
				}
				if data.EnergyWH != testData.EnergyWH {
					t.Errorf("EnergyWH = %v, want %v", data.EnergyWH, testData.EnergyWH)
				}
			},
		},
		{
			name:     "read non-existent device returns default",
			deviceID: "new-device",
			wantErr:  false,
			validate: func(t *testing.T, data *PowerData) {
				if data.EnergyWH != 0.0 {
					t.Errorf("EnergyWH = %v, want 0.0", data.EnergyWH)
				}
				// Timestamp should be recent (within last second)
				now := time.Now().UnixMilli()
				if data.Timestamp > now || data.Timestamp < now-1000 {
					t.Errorf("Timestamp = %v, want recent timestamp near %v", data.Timestamp, now)
				}
			},
		},
		{
			name:     "read corrupted file",
			deviceID: "corrupted",
			wantErr:  true,
		},
		{
			name:     "invalid device ID",
			deviceID: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := reader.Read(tt.deviceID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Read() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("Read() error = %v, want nil", err)
					return
				}
				if tt.validate != nil {
					tt.validate(t, data)
				}
			}
		})
	}
}

func TestFileWriter_FileReader_Integration(t *testing.T) {
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

	writer := NewFileWriter(config, logger)
	reader := NewFileReader(config, logger)

	// Test write-read cycle for multiple devices
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
				EnergyWH:  0.0,
			},
		},
	}

	// Write all devices
	for _, device := range devices {
		if err := writer.Write(device.id, device.data); err != nil {
			t.Fatalf("failed to write device %s: %v", device.id, err)
		}
	}

	// Read and verify all devices
	for _, device := range devices {
		data, err := reader.Read(device.id)
		if err != nil {
			t.Fatalf("failed to read device %s: %v", device.id, err)
		}

		if data.Timestamp != device.data.Timestamp {
			t.Errorf("device %s: Timestamp = %v, want %v", device.id, data.Timestamp, device.data.Timestamp)
		}
		if data.EnergyWH != device.data.EnergyWH {
			t.Errorf("device %s: EnergyWH = %v, want %v", device.id, data.EnergyWH, device.data.EnergyWH)
		}
	}

	// Test updating a device
	updatedData := &PowerData{
		Timestamp: time.Now().UnixMilli() + 1000,
		EnergyWH:  1500.75,
	}
	if err := writer.Write("device1", updatedData); err != nil {
		t.Fatalf("failed to update device1: %v", err)
	}

	// Verify update
	data, err := reader.Read("device1")
	if err != nil {
		t.Fatalf("failed to read updated device1: %v", err)
	}
	if data.Timestamp != updatedData.Timestamp || data.EnergyWH != updatedData.EnergyWH {
		t.Errorf("device1 not updated correctly: got (%v, %v), want (%v, %v)",
			data.Timestamp, data.EnergyWH, updatedData.Timestamp, updatedData.EnergyWH)
	}
}
