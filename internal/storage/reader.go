package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileReader implements the FileReaderInterface for reading device data files.
type FileReader struct {
	config *Config
}

// NewFileReader creates a new file reader instance.
func NewFileReader(config *Config) *FileReader {
	return &FileReader{
		config: config,
	}
}

// NewFileReaderWithPath creates a new file reader instance with a specific data directory.
// This is a convenience function that creates a config with the specified data directory.
func NewFileReaderWithPath(dataDir string) *FileReader {
	config := NewConfigWithPath(dataDir)
	return NewFileReader(config)
}

// NewFileReaderWithConfig creates a new file reader with configuration (alias for NewFileReader).
func NewFileReaderWithConfig(config *Config) *FileReader {
	return NewFileReader(config)
}

// ReadDeviceFile reads the raw content of a device file.
func (fr *FileReader) ReadDeviceFile(deviceID string) ([]byte, error) {
	if deviceID == "" {
		return nil, NewStorageError("read", deviceID, "", fmt.Errorf("device ID cannot be empty"))
	}

	filePath := fr.getDeviceFilePath(deviceID)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, NewStorageError("read", deviceID, filePath, ErrFileNotFound)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewStorageError("read", deviceID, filePath, fmt.Errorf("failed to read file: %w", err))
	}

	return content, nil
}

// ParseData parses file content into a PowerData structure.
func (fr *FileReader) ParseData(content []byte) (*PowerData, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("file content is empty")
	}

	// Convert to string and split by lines
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	// Validate file format (should have exactly 2 lines)
	if len(lines) != 2 {
		return nil, fmt.Errorf("%w: expected 2 lines, got %d", ErrInvalidFormat, len(lines))
	}

	// Parse timestamp (first line)
	timestamp, err := strconv.ParseInt(strings.TrimSpace(lines[0]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid timestamp format: %w", ErrInvalidFormat, err)
	}

	// Parse energy value (second line)
	energyValue, err := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid energy value format: %w", ErrInvalidFormat, err)
	}

	// Create PowerData structure
	data := &PowerData{
		Timestamp: timestamp,
		EnergyWH:  energyValue,
	}

	// Validate parsed data
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("parsed data validation failed: %w", err)
	}

	return data, nil
}

// ReadAndParse combines reading and parsing operations for convenience.
func (fr *FileReader) ReadAndParse(deviceID string) (*PowerData, error) {
	content, err := fr.ReadDeviceFile(deviceID)
	if err != nil {
		// Check if file doesn't exist - return initialized data
		if IsFileNotFoundError(err) {
			return fr.createInitializedData(), nil
		}
		return nil, err
	}

	data, err := fr.ParseData(content)
	if err != nil {
		// If parsing fails, return initialized data and log the error
		return fr.createInitializedData(), fmt.Errorf("failed to parse file content: %w", err)
	}

	return data, nil
}

// getDeviceFilePath returns the full path for a device file.
func (fr *FileReader) getDeviceFilePath(deviceID string) string {
	filename := fmt.Sprintf("%s.txt", deviceID)
	return filepath.Join(fr.config.DataDir, filename)
}

// createInitializedData creates initialized PowerData for new devices.
func (fr *FileReader) createInitializedData() *PowerData {
	return NewPowerData(0) // Initialize with zero energy
}

// FileExists checks if a device file exists.
func (fr *FileReader) FileExists(deviceID string) bool {
	filePath := fr.getDeviceFilePath(deviceID)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// ReadLines reads file content as lines for debugging purposes.
func (fr *FileReader) ReadLines(deviceID string) ([]string, error) {
	content, err := fr.ReadDeviceFile(deviceID)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file content: %w", err)
	}

	return lines, nil
}

// IsFileNotFoundError checks if an error is a file not found error.
func IsFileNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var storageErr *StorageError
	if AsStorageError(err, &storageErr) {
		return storageErr.Cause == ErrFileNotFound
	}

	return os.IsNotExist(err)
}

// AsStorageError attempts to cast an error to StorageError.
func AsStorageError(err error, target **StorageError) bool {
	if err == nil {
		return false
	}

	if se, ok := err.(*StorageError); ok {
		*target = se
		return true
	}

	return false
}
