package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

// Config represents the storage module configuration.
// This configuration controls file storage behavior and permissions.
type Config struct {
	// DataDir is the directory path where device data files are stored.
	DataDir string `yaml:"data_dir" json:"data_dir" env:"STORAGE_DATA_DIR"`

	// FilePermissions specifies the file permissions for device data files.
	FilePermissions os.FileMode `yaml:"file_permissions" json:"file_permissions" env:"STORAGE_FILE_PERMISSIONS"`

	// DirPermissions specifies the directory permissions for the data directory.
	DirPermissions os.FileMode `yaml:"dir_permissions" json:"dir_permissions" env:"STORAGE_DIR_PERMISSIONS"`

	// SyncWrite enables synchronous writes to ensure data durability.
	SyncWrite bool `yaml:"sync_write" json:"sync_write" env:"STORAGE_SYNC_WRITE"`

	// CreateDir enables automatic creation of the data directory.
	CreateDir bool `yaml:"create_dir" json:"create_dir" env:"STORAGE_CREATE_DIR"`
}

// NewConfigWithDefaults creates a new storage configuration with default values.
// This function is maintained for backward compatibility.
func NewConfigWithDefaults() *Config {
	cfg := &Config{}
	cfg.SetDefaults()
	// Set boolean defaults for new configurations
	cfg.SyncWrite = true
	cfg.CreateDir = true
	return cfg
}

// NewConfigWithPath creates a new storage configuration with a specific data directory.
func NewConfigWithPath(dataDir string) *Config {
	cfg := NewConfigWithDefaults()
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

	// Validate data directory path format
	if len(c.DataDir) > 4096 {
		return fmt.Errorf("storage: data_dir path too long (max 4096 characters)")
	}

	// Check for invalid characters in path (basic validation)
	if strings.ContainsAny(c.DataDir, "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f") {
		return fmt.Errorf("storage: data_dir contains invalid control characters")
	}

	if c.FilePermissions == 0 {
		return fmt.Errorf("storage: file_permissions must be specified")
	}

	// Validate file permissions are within reasonable range
	if c.FilePermissions > 0777 {
		return fmt.Errorf("storage: file_permissions too permissive (max 0777)")
	}

	if c.DirPermissions == 0 {
		return fmt.Errorf("storage: dir_permissions must be specified")
	}

	// Validate directory permissions are within reasonable range
	if c.DirPermissions > 0777 {
		return fmt.Errorf("storage: dir_permissions too permissive (max 0777)")
	}

	// Validate directory permissions allow read access (at least for owner)
	if c.DirPermissions&0400 == 0 {
		return fmt.Errorf("storage: dir_permissions must allow read access for owner")
	}

	return nil
}

// String returns a string representation of the configuration.
// This method is used for logging and debugging.
func (c *Config) String() string {
	if c == nil {
		return "<nil>"
	}
	return fmt.Sprintf("StorageConfig{DataDir: %s, FilePermissions: %o, DirPermissions: %o, SyncWrite: %t, CreateDir: %t}",
		c.DataDir, c.FilePermissions, c.DirPermissions, c.SyncWrite, c.CreateDir)
}

// Clone creates a deep copy of the configuration.
func (c *Config) Clone() config.Config {
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

// NewConfig creates a new storage configuration using the config loader.
// This function loads configuration from YAML files and environment variables.
func NewConfig(loader *config.Loader) (*Config, error) {
	cfg := &Config{}
	cfg.SetDefaults()

	if err := loader.LoadModule("storage", cfg); err != nil {
		return nil, fmt.Errorf("failed to load storage config: %w", err)
	}

	// Validate the loaded configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid storage config: %w", err)
	}

	return cfg, nil
}
