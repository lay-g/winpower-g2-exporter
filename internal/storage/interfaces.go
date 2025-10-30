package storage

// StorageManager defines the interface for storage operations.
// It provides methods to read and write power data for devices.
type StorageManager interface {
	// Write stores power data for a device.
	// It validates the device ID and data before writing.
	// Returns an error if validation fails or write operation fails.
	Write(deviceID string, data *PowerData) error

	// Read retrieves power data for a device.
	// For new devices (file doesn't exist), it returns default initialized data.
	// Returns an error if the device ID is invalid or read operation fails.
	Read(deviceID string) (*PowerData, error)
}

// FileWriter defines the interface for writing device data to files.
type FileWriter interface {
	// Write writes power data for a device to its file.
	Write(deviceID string, data *PowerData) error
}

// FileReader defines the interface for reading device data from files.
type FileReader interface {
	// Read reads power data for a device from its file.
	// Returns default data if the file doesn't exist.
	Read(deviceID string) (*PowerData, error)
}
