package storage

// StorageManager defines the interface for storage operations.
// This interface provides abstraction for reading and writing power data.
type StorageManager interface {
	// Write persists power data for a specific device.
	// The operation should be atomic to prevent data corruption.
	Write(deviceID string, data *PowerData) error

	// Read retrieves power data for a specific device.
	// Returns initialized data with zero values for new devices.
	Read(deviceID string) (*PowerData, error)
}

// FileWriterInterface defines the interface for file write operations.
type FileWriterInterface interface {
	// WriteDeviceFile writes power data to a device-specific file.
	WriteDeviceFile(deviceID string, data *PowerData) error
}

// FileReaderInterface defines the interface for file read operations.
type FileReaderInterface interface {
	// ReadDeviceFile reads the raw content of a device file.
	ReadDeviceFile(deviceID string) ([]byte, error)

	// ParseData parses file content into a PowerData structure.
	ParseData(content []byte) (*PowerData, error)

	// ReadAndParse combines reading and parsing operations for convenience.
	ReadAndParse(deviceID string) (*PowerData, error)

	// FileExists checks if a device file exists.
	FileExists(deviceID string) bool
}
