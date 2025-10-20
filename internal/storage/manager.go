package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileStorageManager implements the StorageManager interface using file-based persistence.
// It provides atomic write operations and reliable read operations for device data.
type FileStorageManager struct {
	config *Config
	writer FileWriterInterface
	reader FileReaderInterface
}

// NewFileStorageManager creates a new file storage manager instance.
func NewFileStorageManager(config *Config) (*FileStorageManager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Check if data directory exists when CreateDir is false
	if !config.CreateDir {
		if _, err := os.Stat(config.DataDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("data directory does not exist and CreateDir is false: %s", config.DataDir)
		}
	}

	manager := &FileStorageManager{
		config: config.Clone(), // Create a copy to prevent external modifications
		writer: NewFileWriter(config),
		reader: NewFileReader(config),
	}

	// Initialize data directory if needed
	if config.CreateDir {
		writer := NewFileWriter(config)
		if err := writer.ensureDirectoryExists(); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	return manager, nil
}

// Write persists power data for a specific device with atomic guarantees.
func (fsm *FileStorageManager) Write(deviceID string, data *PowerData) error {
	if err := fsm.writer.WriteDeviceFile(deviceID, data); err != nil {
		return fmt.Errorf("failed to write data for device '%s': %w", deviceID, err)
	}
	return nil
}

// Read retrieves power data for a specific device.
// Returns initialized data with zero values for new devices.
func (fsm *FileStorageManager) Read(deviceID string) (*PowerData, error) {
	data, err := fsm.reader.ReadAndParse(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to read data for device '%s': %w", deviceID, err)
	}
	return data, nil
}

// ReadRaw reads the raw file content for a device.
// This method is useful for debugging and file inspection.
func (fsm *FileStorageManager) ReadRaw(deviceID string) ([]byte, error) {
	content, err := fsm.reader.ReadDeviceFile(deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to read raw data for device '%s': %w", deviceID, err)
	}
	return content, nil
}

// FileExists checks if a data file exists for the specified device.
func (fsm *FileStorageManager) FileExists(deviceID string) bool {
	return fsm.reader.FileExists(deviceID)
}

// GetConfig returns a copy of the current configuration.
func (fsm *FileStorageManager) GetConfig() *Config {
	return fsm.config.Clone()
}

// GetDataDir returns the data directory path.
func (fsm *FileStorageManager) GetDataDir() string {
	return fsm.config.DataDir
}

// ValidateDeviceData checks if existing device data is valid.
// This method can be used for data integrity checks.
func (fsm *FileStorageManager) ValidateDeviceData(deviceID string) error {
	if !fsm.FileExists(deviceID) {
		return nil // No file to validate
	}

	data, err := fsm.Read(deviceID)
	if err != nil {
		return fmt.Errorf("failed to read device data for validation: %w", err)
	}

	if err := data.Validate(); err != nil {
		return fmt.Errorf("device data validation failed: %w", err)
	}

	return nil
}

// ListDeviceFiles returns a list of all device files in the data directory.
// This method is useful for administrative purposes.
func (fsm *FileStorageManager) ListDeviceFiles() ([]string, error) {
	entries, err := os.ReadDir(fsm.config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	var deviceFiles []string
	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Check if file has .txt extension and represents a device file
		if filepath.Ext(entry.Name()) == ".txt" {
			// Check if the file follows the device naming pattern (deviceid.txt)
			// and is a valid device data file by trying to read it
			deviceID := entry.Name()[:len(entry.Name())-4] // Remove .txt extension
			if fsm.isValidDeviceFile(deviceID) {
				fullPath := filepath.Join(fsm.config.DataDir, entry.Name())
				deviceFiles = append(deviceFiles, fullPath)
			}
		}
	}

	return deviceFiles, nil
}

// isValidDeviceFile checks if a file is a valid device data file by attempting to read and parse it.
func (fsm *FileStorageManager) isValidDeviceFile(deviceID string) bool {
	_, err := fsm.reader.ReadAndParse(deviceID)
	return err == nil
}

// GetDeviceFilePath returns the full path for a device's data file.
func (fsm *FileStorageManager) GetDeviceFilePath(deviceID string) string {
	return fmt.Sprintf("%s/%s.txt", fsm.config.DataDir, deviceID)
}

// CreateInitializedData creates initialized PowerData for new devices.
// This is a utility method that can be used by other modules.
func CreateInitializedData() *PowerData {
	return NewPowerDataWithTimestamp(0, 0)
}

// IsZeroData checks if PowerData represents an uninitialized state.
// This is a standalone function with different behavior than the method.
func IsZeroData(data *PowerData) bool {
	if data == nil {
		return false
	}
	return data.Timestamp == 0 && data.EnergyWH == 0
}
