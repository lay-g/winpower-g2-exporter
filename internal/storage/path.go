package storage

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateDeviceID checks if a device ID is valid.
// Device IDs must be non-empty and not contain path separators or relative path components.
func validateDeviceID(deviceID string) error {
	if deviceID == "" {
		return fmt.Errorf("%w: device ID cannot be empty", ErrInvalidDeviceID)
	}

	// Check for path separators
	if strings.Contains(deviceID, "/") || strings.Contains(deviceID, "\\") {
		return fmt.Errorf("%w: device ID cannot contain path separators", ErrInvalidDeviceID)
	}

	// Check for relative path components
	if deviceID == "." || deviceID == ".." {
		return fmt.Errorf("%w: device ID cannot be a relative path component", ErrInvalidDeviceID)
	}

	// Check for leading/trailing dots (hidden files or relative paths)
	if strings.HasPrefix(deviceID, ".") {
		return fmt.Errorf("%w: device ID cannot start with a dot", ErrInvalidDeviceID)
	}

	return nil
}

// buildFilePath constructs a safe file path for a device.
// It validates the device ID and ensures the path stays within the data directory.
func buildFilePath(dataDir, deviceID string) (string, error) {
	if err := validateDeviceID(deviceID); err != nil {
		return "", err
	}

	// Construct the file path
	fileName := deviceID + ".txt"
	filePath := filepath.Join(dataDir, fileName)

	// Clean the path to resolve any relative components
	filePath = filepath.Clean(filePath)
	dataDir = filepath.Clean(dataDir)

	// Ensure the file path is within the data directory
	// This prevents path traversal attacks
	relPath, err := filepath.Rel(dataDir, filePath)
	if err != nil {
		return "", fmt.Errorf("%w: failed to validate file path", ErrInvalidDeviceID)
	}

	// Check if the relative path tries to escape the data directory
	if strings.HasPrefix(relPath, "..") || strings.HasPrefix(relPath, string(filepath.Separator)) {
		return "", fmt.Errorf("%w: device ID would escape data directory", ErrInvalidDeviceID)
	}

	return filePath, nil
}
