package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileWriter implements the FileWriterInterface for writing device data files.
type FileWriter struct {
	config *Config
}

// NewFileWriter creates a new file writer instance.
func NewFileWriter(config *Config) *FileWriter {
	return &FileWriter{
		config: config,
	}
}

// NewFileWriterWithPath creates a new file writer instance with a specific data directory.
// This is a convenience function that creates a config with the specified data directory.
func NewFileWriterWithPath(dataDir string) *FileWriter {
	config := NewConfigWithPath(dataDir)
	return NewFileWriter(config)
}

// NewFileWriterWithConfig creates a new file writer with configuration (alias for NewFileWriter).
func NewFileWriterWithConfig(config *Config) *FileWriter {
	return NewFileWriter(config)
}

// WriteDeviceFile writes power data to a device-specific file.
// The operation includes atomic write guarantees and proper error handling.
func (fw *FileWriter) WriteDeviceFile(deviceID string, data *PowerData) error {
	if deviceID == "" {
		return NewStorageError("write", deviceID, "", fmt.Errorf("device ID cannot be empty"))
	}

	if data == nil {
		return NewStorageError("write", deviceID, "", fmt.Errorf("data cannot be nil"))
	}

	// Validate data before writing
	if err := data.Validate(); err != nil {
		return NewStorageError("write", deviceID, "", fmt.Errorf("invalid data: %w", err))
	}

	// Get the full file path
	filePath := fw.getDeviceFilePath(deviceID)

	// Ensure the directory exists
	if err := fw.ensureDirectoryExists(); err != nil {
		return NewStorageError("write", deviceID, filePath, err)
	}

	// Prepare file content
	content := fw.formatDataContent(data)

	// Write to temporary file first for atomic operation
	tempPath := filePath + ".tmp"
	if err := fw.writeFileContent(tempPath, content); err != nil {
		return NewStorageError("write", deviceID, tempPath, err)
	}

	// Rename temporary file to final file (atomic operation)
	if err := os.Rename(tempPath, filePath); err != nil {
		// Clean up temp file if rename fails
		_ = os.Remove(tempPath)
		return NewStorageError("write", deviceID, filePath, fmt.Errorf("failed to rename temp file: %w", err))
	}

	return nil
}

// ensureDirectoryExists creates the data directory if it doesn't exist.
func (fw *FileWriter) ensureDirectoryExists() error {
	if !fw.config.CreateDir {
		// Check if directory exists
		if _, err := os.Stat(fw.config.DataDir); os.IsNotExist(err) {
			return fmt.Errorf("data directory '%s' does not exist and create_dir is disabled", fw.config.DataDir)
		}
		return nil
	}

	// Create directory with proper permissions
	if err := os.MkdirAll(fw.config.DataDir, fw.config.DirPermissions); err != nil {
		return fmt.Errorf("failed to create data directory '%s': %w", fw.config.DataDir, err)
	}

	return nil
}

// getDeviceFilePath returns the full path for a device file.
func (fw *FileWriter) getDeviceFilePath(deviceID string) string {
	filename := fmt.Sprintf("%s.txt", deviceID)
	return filepath.Join(fw.config.DataDir, filename)
}

// formatDataContent formats power data into the file content format.
func (fw *FileWriter) formatDataContent(data *PowerData) string {
	// Format: timestamp on first line, energy value on second line
	return fmt.Sprintf("%d\n%f\n", data.Timestamp, data.EnergyWH)
}

// writeFileContent writes content to a file with proper permissions and sync options.
func (fw *FileWriter) writeFileContent(filePath, content string) error {
	// Create file with specified permissions
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fw.config.FilePermissions)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Write content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	// Sync to disk if enabled
	if fw.config.SyncWrite {
		if err := file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file to disk: %w", err)
		}
	}

	return nil
}

// WriteStringToFile is a utility method for writing string content to a file.
// This method is primarily used for testing and utilities.
func (fw *FileWriter) WriteStringToFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fw.config.FilePermissions)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}

	if fw.config.SyncWrite {
		if err := file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file to disk: %w", err)
		}
	}

	return nil
}
