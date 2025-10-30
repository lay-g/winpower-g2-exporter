// Package storage provides file-based persistent storage for device power data.
//
// The storage module stores accumulated energy values for each device in individual
// text files. Each file contains two lines: timestamp (Unix milliseconds) and
// energy value (watt-hours). This simple format ensures easy debugging and
// manual inspection when needed.
//
// # Basic Usage
//
// Create a storage manager with default configuration:
//
//	config := storage.DefaultConfig()
//	manager, err := storage.NewFileStorageManager(config, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Write device data:
//
//	data := &storage.PowerData{
//	    Timestamp: time.Now().UnixMilli(),
//	    EnergyWH:  1234.5,
//	}
//	err := manager.Write("device-001", data)
//	if err != nil {
//	    log.Printf("failed to write: %v", err)
//	}
//
// Read device data:
//
//	data, err := manager.Read("device-001")
//	if err != nil {
//	    log.Printf("failed to read: %v", err)
//	}
//	fmt.Printf("Energy: %.2f WH at %d\n", data.EnergyWH, data.Timestamp)
//
// # File Format
//
// Each device's data is stored in a text file with the following format:
//
//	<timestamp in Unix milliseconds>
//	<energy in watt-hours>
//
// Example file content (./data/device-001.txt):
//
//	1698758400000
//	1234.5
//
// Files are stored in the configured DataDir directory with the naming pattern:
// <device-id>.txt. The device ID is validated to prevent path traversal attacks.
//
// # Configuration
//
// The storage module can be configured with custom settings:
//
//	config := &storage.Config{
//	    DataDir:         "/var/lib/winpower/data",
//	    FilePermissions: 0644,
//	}
//	if err := config.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//	manager, err := storage.NewFileStorageManager(config, logger)
//
// # Thread Safety
//
// The storage module uses atomic file operations (write to temp file + rename)
// to prevent data corruption during writes. This ensures that files are either
// fully written or not written at all, even if the process crashes mid-write.
//
// However, concurrent writes to the same device from multiple goroutines may
// result in data races. Callers should serialize writes to the same device ID
// if concurrent access is needed. The scheduler module handles this coordination
// in the main application.
//
// # Error Handling
//
// All operations return typed errors that can be inspected using standard
// error handling patterns:
//
//	data, err := manager.Read("device-001")
//	if errors.Is(err, storage.ErrFileNotFound) {
//	    // File doesn't exist - this is expected for new devices
//	    // Read() returns default data with zero values in this case
//	}
//
// For detailed error information, unwrap the StorageError:
//
//	var storageErr *storage.StorageError
//	if errors.As(err, &storageErr) {
//	    fmt.Printf("Operation: %s, Path: %s, Cause: %v\n",
//	        storageErr.Operation, storageErr.Path, storageErr.Err)
//	}
//
// Common errors:
//   - ErrFileNotFound: Device file doesn't exist (returns default data)
//   - ErrInvalidFormat: File content is corrupted or malformed
//   - ErrInvalidDeviceID: Device ID contains invalid characters
//   - ErrInvalidData: PowerData validation failed
//   - ErrPermissionDenied: Insufficient permissions to read/write files
//
// # Data Validation
//
// PowerData is validated before writing to ensure data integrity:
//   - Timestamp must be positive and not more than 24 hours in the future
//   - Energy value must be finite (not NaN or Inf) and non-negative
//
// Validation happens automatically in Write() operations. You can also
// validate data explicitly:
//
//	data := &storage.PowerData{Timestamp: -1, EnergyWH: 100}
//	if err := data.Validate(); err != nil {
//	    log.Printf("invalid data: %v", err)
//	}
//
// # Integration
//
// The storage module integrates with the logging module for operation tracing:
//   - Debug logs for all read/write operations
//   - Info logs for successful writes
//   - Error logs for all failures
//
// All log messages include the device ID and relevant data fields for
// troubleshooting.
package storage
