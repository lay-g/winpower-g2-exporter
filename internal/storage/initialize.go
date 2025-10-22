package storage

import (
	"fmt"
	"os"
)

// InitializeOptions contains options for storage module initialization.
type InitializeOptions struct {
	// Config specifies the storage configuration. Required.
	Config *Config

	// ValidateOnInit enables validation of the data directory during initialization.
	ValidateOnInit bool
}

// NewInitializeOptions creates initialization options with the given config.
func NewInitializeOptions(config *Config) *InitializeOptions {
	return &InitializeOptions{
		Config:         config,
		ValidateOnInit: true,
	}
}

// Initialize initializes the storage module with the given options.
// This function handles directory creation, validation, and configuration setup.
// The config parameter must be provided by the caller - the storage module
// does not handle configuration loading from any sources.
func Initialize(opts *InitializeOptions) (*FileStorageManager, error) {
	if opts == nil {
		return nil, fmt.Errorf("initialization options cannot be nil")
	}
	if opts.Config == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}

	// Clone configuration to avoid modifying the original
	configCloned := opts.Config.Clone()
	config, ok := configCloned.(*Config)
	if !ok {
		return nil, fmt.Errorf("failed to clone configuration: unexpected type")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid storage configuration: %w", err)
	}

	// Create storage manager
	manager, err := NewFileStorageManager(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage manager: %w", err)
	}

	// Validate data directory if requested
	if opts.ValidateOnInit {
		if err := ValidateDataDirectory(config.DataDir); err != nil {
			return nil, fmt.Errorf("data directory validation failed: %w", err)
		}
	}

	return manager, nil
}

// InitializeWithConfig initializes the storage module with the given configuration.
// This is the recommended way to initialize the storage module.
func InitializeWithConfig(config *Config) (*FileStorageManager, error) {
	return Initialize(NewInitializeOptions(config))
}

// ValidateDataDirectory validates that the data directory exists and is accessible.
func ValidateDataDirectory(dataDir string) error {
	if dataDir == "" {
		return fmt.Errorf("data directory path is empty")
	}

	// Check if directory exists
	info, err := os.Stat(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("data directory '%s' does not exist", dataDir)
		}
		return fmt.Errorf("failed to access data directory '%s': %w", dataDir, err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("path '%s' exists but is not a directory", dataDir)
	}

	// Check if directory is writable
	testFile := fmt.Sprintf("%s/.write_test", dataDir)
	file, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("data directory '%s' is not writable: %w", dataDir, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close test file: %w", err)
	}
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to remove test file: %w", err)
	}

	return nil
}

// EnsureDirectoryExists ensures the data directory exists with proper permissions.
func EnsureDirectoryExists(dataDir string, dirPermissions os.FileMode) error {
	if dataDir == "" {
		return fmt.Errorf("data directory path is empty")
	}

	// Check if directory already exists
	info, err := os.Stat(dataDir)
	if err == nil && info.IsDir() {
		// Directory exists, check permissions
		if info.Mode().Perm() != dirPermissions {
			// Try to update permissions
			if err := os.Chmod(dataDir, dirPermissions); err != nil {
				return fmt.Errorf("failed to update directory permissions: %w", err)
			}
		}
		return nil
	}

	// Create directory with proper permissions
	if err := os.MkdirAll(dataDir, dirPermissions); err != nil {
		return fmt.Errorf("failed to create data directory '%s': %w", dataDir, err)
	}

	return nil
}

// GetStorageInfo returns information about the storage module configuration.
func GetStorageInfo(manager *FileStorageManager) map[string]interface{} {
	if manager == nil {
		return map[string]interface{}{
			"initialized": false,
			"error":       "manager is nil",
		}
	}

	config := manager.GetConfig()
	return map[string]interface{}{
		"initialized":      true,
		"data_dir":         config.DataDir,
		"sync_write":       config.SyncWrite,
		"create_dir":       config.CreateDir,
		"file_permissions": fmt.Sprintf("%#o", config.FilePermissions),
		"dir_permissions":  fmt.Sprintf("%#o", config.DirPermissions),
		"module_version":   ModuleVersion,
	}
}

// CheckStorageHealth checks the health of the storage system.
func CheckStorageHealth(manager *FileStorageManager) error {
	if manager == nil {
		return fmt.Errorf("storage manager is nil")
	}

	config := manager.GetConfig()

	// Check if data directory is accessible
	if err := ValidateDataDirectory(config.DataDir); err != nil {
		return fmt.Errorf("storage health check failed: %w", err)
	}

	return nil
}

// Cleanup performs cleanup operations for the storage module.
func Cleanup(manager *FileStorageManager) error {
	if manager == nil {
		return nil
	}

	// Currently, there's nothing specific to clean up
	// This function is provided for future extensibility
	return nil
}
