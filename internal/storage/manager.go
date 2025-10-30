package storage

import (
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// FileStorageManager implements the StorageManager interface using file-based storage.
//
// It coordinates file reading and writing operations through FileReader and FileWriter
// components. Each device's data is stored in a separate text file.
//
// The manager provides:
//   - Data validation before writes
//   - Comprehensive logging for all operations
//   - Atomic file operations to prevent corruption
//   - Default data for non-existent devices
type FileStorageManager struct {
	config *Config
	reader FileReader
	writer FileWriter
	logger log.Logger
}

// NewFileStorageManager creates a new FileStorageManager with the given configuration.
//
// This function:
//   - Validates the configuration
//   - Initializes FileReader and FileWriter components
//   - Returns a fully initialized storage manager
//
// Parameters:
//   - config: Storage configuration (must not be nil and must pass Validate())
//   - logger: Logger instance for operation logging (must not be nil)
//
// Returns:
//   - StorageManager: Initialized storage manager ready for use
//   - error: Configuration validation error, if any
//
// The returned manager is safe to use immediately. The data directory will be
// created automatically on the first write operation if it doesn't exist.
//
// Example:
//
//	config := storage.DefaultConfig()
//	config.DataDir = "/var/lib/winpower/data"
//	manager, err := storage.NewFileStorageManager(config, logger)
//	if err != nil {
//	    log.Fatalf("failed to create storage manager: %v", err)
//	}
func NewFileStorageManager(config *Config, logger log.Logger) (StorageManager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	reader := NewFileReader(config, logger)
	writer := NewFileWriter(config, logger)

	return &FileStorageManager{
		config: config,
		reader: reader,
		writer: writer,
		logger: logger,
	}, nil
}

// Write stores power data for a device.
//
// This method:
//   - Validates the device ID and power data
//   - Writes data atomically to prevent corruption
//   - Creates the data directory if it doesn't exist
//   - Logs the operation (debug for attempt, info for success, error for failure)
//
// Parameters:
//   - deviceID: Unique identifier for the device (must be valid per validateDeviceID)
//   - data: Power data to store (must pass PowerData.Validate())
//
// Returns:
//   - error: Validation error, I/O error, or nil on success
//
// The write operation is atomic (uses temp file + rename) to ensure that files
// are either fully written or not written at all, even if the process crashes.
//
// Example:
//
//	data := &storage.PowerData{
//	    Timestamp: time.Now().UnixMilli(),
//	    EnergyWH:  1234.5,
//	}
//	if err := manager.Write("device-001", data); err != nil {
//	    log.Printf("failed to write: %v", err)
//	}
func (m *FileStorageManager) Write(deviceID string, data *PowerData) error {
	m.logger.Debug("writing device data",
		log.String("device_id", deviceID),
		log.Int64("timestamp", data.Timestamp),
		log.Float64("energy_wh", data.EnergyWH))

	if err := m.writer.Write(deviceID, data); err != nil {
		m.logger.Error("failed to write device data",
			log.String("device_id", deviceID),
			log.Err(err))
		return err
	}

	m.logger.Info("device data written successfully",
		log.String("device_id", deviceID))

	return nil
}

// Read retrieves power data for a device.
//
// This method:
//   - Validates the device ID
//   - Reads data from the device's file
//   - Returns default data (zero values) if the file doesn't exist
//   - Parses and validates the file content
//   - Logs the operation (debug for attempt and success, error for failure)
//
// Parameters:
//   - deviceID: Unique identifier for the device (must be valid per validateDeviceID)
//
// Returns:
//   - *PowerData: Retrieved power data, or default data if file doesn't exist
//   - error: Validation error, I/O error, parse error, or nil on success
//
// For new devices (file doesn't exist), this method returns default data with
// zero values instead of an error. This simplifies initialization logic in
// the caller.
//
// Example:
//
//	data, err := manager.Read("device-001")
//	if err != nil {
//	    log.Printf("failed to read: %v", err)
//	    return
//	}
//	fmt.Printf("Energy: %.2f WH at timestamp %d\n",
//	    data.EnergyWH, data.Timestamp)
func (m *FileStorageManager) Read(deviceID string) (*PowerData, error) {
	m.logger.Debug("reading device data",
		log.String("device_id", deviceID))

	data, err := m.reader.Read(deviceID)
	if err != nil {
		m.logger.Error("failed to read device data",
			log.String("device_id", deviceID),
			log.Err(err))
		return nil, err
	}

	m.logger.Debug("device data read successfully",
		log.String("device_id", deviceID),
		log.Int64("timestamp", data.Timestamp),
		log.Float64("energy_wh", data.EnergyWH))

	return data, nil
}
