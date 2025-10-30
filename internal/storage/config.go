package storage

import (
	"fmt"
	"os"
)

// Config holds configuration for the storage module.
//
// This configuration defines where device data files are stored and
// what permissions they should have. Use DefaultConfig() for sensible
// defaults, or create a custom Config for specific requirements.
//
// Example:
//
//	config := &storage.Config{
//	    DataDir:         "/var/lib/winpower/data",
//	    FilePermissions: 0640,
//	}
//	if err := config.Validate(); err != nil {
//	    log.Fatal(err)
//	}
type Config struct {
	// DataDir is the directory where device data files are stored
	DataDir string `json:"data_dir" yaml:"data_dir" mapstructure:"data_dir"`

	// FilePermissions defines the permission bits for created files (e.g., 0644)
	FilePermissions os.FileMode `json:"file_permissions" yaml:"file_permissions" mapstructure:"file_permissions"`
}

// DefaultConfig returns a Config with sensible default values.
//
// The default configuration uses:
//   - DataDir: "./data" (relative to current working directory)
//   - FilePermissions: 0644 (owner read/write, group/others read-only)
//
// This is suitable for development and testing. For production, consider
// using an absolute path and more restrictive permissions.
//
// Example:
//
//	config := storage.DefaultConfig()
//	config.DataDir = "/var/lib/winpower/data"  // Override for production
//	manager, err := storage.NewFileStorageManager(config, logger)
func DefaultConfig() *Config {
	return &Config{
		DataDir:         "./data",
		FilePermissions: 0644,
	}
}

// Validate checks if the configuration is valid.
//
// Validation rules:
//   - Config must not be nil
//   - DataDir must not be empty
//   - FilePermissions must be between 0 and 0777 (valid Unix permissions)
//
// Returns an error if any validation rule is violated.
//
// Example:
//
//	config := &storage.Config{
//	    DataDir:         "/var/lib/winpower/data",
//	    FilePermissions: 0644,
//	}
//	if err := config.Validate(); err != nil {
//	    log.Fatalf("invalid config: %v", err)
//	}
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if c.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}

	// Validate file permissions (must be a valid Unix permission)
	if c.FilePermissions > 0777 {
		return fmt.Errorf("file permissions must be a valid Unix permission (0-0777)")
	}

	return nil
}
