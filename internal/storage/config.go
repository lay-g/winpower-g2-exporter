package storage

import (
	"fmt"
	"os"
)

// Config represents the storage module configuration.
// This configuration controls file storage behavior and permissions.
type Config struct {
	// DataDir is the directory path where device data files are stored.
	DataDir string `yaml:"data_dir" json:"data_dir"`

	// FilePermissions specifies the file permissions for device data files.
	FilePermissions os.FileMode `yaml:"file_permissions" json:"file_permissions"`

	// DirPermissions specifies the directory permissions for the data directory.
	DirPermissions os.FileMode `yaml:"dir_permissions" json:"dir_permissions"`

	// SyncWrite enables synchronous writes to ensure data durability.
	SyncWrite bool `yaml:"sync_write" json:"sync_write"`

	// CreateDir enables automatic creation of the data directory.
	CreateDir bool `yaml:"create_dir" json:"create_dir"`
}

// NewConfig creates a new storage configuration with default values.
func NewConfig() *Config {
	cfg := &Config{}
	cfg.SetDefaults()
	// Set boolean defaults for new configurations
	cfg.SyncWrite = true
	cfg.CreateDir = true
	return cfg
}

// NewConfigWithPath creates a new storage configuration with a specific data directory.
func NewConfigWithPath(dataDir string) *Config {
	cfg := NewConfig()
	cfg.DataDir = dataDir
	return cfg
}

// SetDefaults sets the default values for the configuration.
func (c *Config) SetDefaults() {
	if c.DataDir == "" {
		c.DataDir = "./data"
	}
	if c.FilePermissions == 0 {
		c.FilePermissions = 0644 // -rw-r--r--
	}
	if c.DirPermissions == 0 {
		c.DirPermissions = 0755 // drwxr-xr-x
	}
	// Note: We don't set boolean defaults here as they should be explicitly set
	// This preserves the existing values when SetDefaults is called
}

// Validate checks if the configuration values are valid.
func (c *Config) Validate() error {
	if c.DataDir == "" {
		return fmt.Errorf("storage: data_dir is required")
	}

	if c.FilePermissions == 0 {
		return fmt.Errorf("storage: file_permissions must be specified")
	}

	if c.DirPermissions == 0 {
		return fmt.Errorf("storage: dir_permissions must be specified")
	}

	return nil
}

// String returns a string representation of the configuration.
// This method is used for logging and debugging.
func (c *Config) String() string {
	return fmt.Sprintf("StorageConfig{DataDir: %s, FilePermissions: %o, DirPermissions: %o, SyncWrite: %t, CreateDir: %t}",
		c.DataDir, c.FilePermissions, c.DirPermissions, c.SyncWrite, c.CreateDir)
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	return &Config{
		DataDir:         c.DataDir,
		FilePermissions: c.FilePermissions,
		DirPermissions:  c.DirPermissions,
		SyncWrite:       c.SyncWrite,
		CreateDir:       c.CreateDir,
	}
}
