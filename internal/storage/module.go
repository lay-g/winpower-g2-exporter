package storage

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Environment variable prefixes for storage configuration
const (
	EnvPrefix = "STORAGE_"
)

// Environment variable keys
const (
	EnvDataDir         = EnvPrefix + "DATA_DIR"
	EnvSyncWrite       = EnvPrefix + "SYNC_WRITE"
	EnvCreateDir       = EnvPrefix + "CREATE_DIR"
	EnvFilePermissions = EnvPrefix + "FILE_PERMISSIONS"
	EnvDirPermissions  = EnvPrefix + "DIR_PERMISSIONS"
)

// ConfigLoader defines the interface for loading configuration from external sources.
// This interface allows the storage module to integrate with distributed configuration systems.
type ConfigLoader interface {
	// LoadString loads a string configuration value by key
	LoadString(key string) (string, bool)

	// LoadInt loads an integer configuration value by key
	LoadInt(key string) (int, bool)

	// LoadBool loads a boolean configuration value by key
	LoadBool(key string) (bool, bool)

	// LoadStringSlice loads a string slice configuration value by key
	LoadStringSlice(key string) ([]string, bool)
}

// NewConfigFromLoader creates a storage configuration from a ConfigLoader.
// This function enables distributed module configuration support.
func NewConfigFromLoader(loader ConfigLoader) *Config {
	config := NewConfig()

	if loader == nil {
		// If no loader provided, use environment variables as fallback
		return NewConfigFromEnv()
	}

	// Load data directory
	if value, found := loader.LoadString("data_dir"); found {
		config.DataDir = value
	}

	// Load sync write setting
	if value, found := loader.LoadBool("sync_write"); found {
		config.SyncWrite = value
	}

	// Load create directory setting
	if value, found := loader.LoadBool("create_dir"); found {
		config.CreateDir = value
	}

	// Load file permissions
	if value, found := loader.LoadString("file_permissions"); found {
		if perm, err := parseFileMode(value); err == nil {
			config.FilePermissions = perm
		}
	}

	// Load directory permissions
	if value, found := loader.LoadString("dir_permissions"); found {
		if perm, err := parseFileMode(value); err == nil {
			config.DirPermissions = perm
		}
	}

	return config
}

// NewConfigFromEnv creates a storage configuration from environment variables.
// This provides environment-based configuration support.
func NewConfigFromEnv() *Config {
	config := NewConfig()

	// Load data directory
	if value := os.Getenv(EnvDataDir); value != "" {
		config.DataDir = value
	}

	// Load sync write setting
	if value := os.Getenv(EnvSyncWrite); value != "" {
		if sync, err := strconv.ParseBool(value); err == nil {
			config.SyncWrite = sync
		}
	}

	// Load create directory setting
	if value := os.Getenv(EnvCreateDir); value != "" {
		if create, err := strconv.ParseBool(value); err == nil {
			config.CreateDir = create
		}
	}

	// Load file permissions
	if value := os.Getenv(EnvFilePermissions); value != "" {
		if perm, err := parseFileMode(value); err == nil {
			config.FilePermissions = perm
		}
	}

	// Load directory permissions
	if value := os.Getenv(EnvDirPermissions); value != "" {
		if perm, err := parseFileMode(value); err == nil {
			config.DirPermissions = perm
		}
	}

	return config
}

// parseFileMode parses a file mode from string (supports octal format)
func parseFileMode(s string) (os.FileMode, error) {
	// Check for empty string
	if s == "" {
		return 0, fmt.Errorf("empty file mode string")
	}

	// Check for invalid characters (only digits allowed)
	for _, r := range s {
		if r < '0' || r > '7' {
			return 0, fmt.Errorf("invalid file mode '%s': contains non-octal characters", s)
		}
	}

	// Remove leading "0" if present
	s = strings.TrimPrefix(s, "0")
	if s == "" {
		s = "0"
	}

	// Parse as octal
	value, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid file mode '%s': %w", s, err)
	}

	return os.FileMode(value), nil
}

// ApplyEnvironmentOverrides applies environment variable overrides to existing config.
func (c *Config) ApplyEnvironmentOverrides() {
	// Load data directory
	if value := os.Getenv(EnvDataDir); value != "" {
		c.DataDir = value
	}

	// Load sync write setting
	if value := os.Getenv(EnvSyncWrite); value != "" {
		if sync, err := strconv.ParseBool(value); err == nil {
			c.SyncWrite = sync
		}
	}

	// Load create directory setting
	if value := os.Getenv(EnvCreateDir); value != "" {
		if create, err := strconv.ParseBool(value); err == nil {
			c.CreateDir = create
		}
	}

	// Load file permissions
	if value := os.Getenv(EnvFilePermissions); value != "" {
		if perm, err := parseFileMode(value); err == nil {
			c.FilePermissions = perm
		}
	}

	// Load directory permissions
	if value := os.Getenv(EnvDirPermissions); value != "" {
		if perm, err := parseFileMode(value); err == nil {
			c.DirPermissions = perm
		}
	}
}

// GetEnvironmentHelp returns help text for storage environment variables.
func GetEnvironmentHelp() string {
	return `Storage Module Environment Variables:
  STORAGE_DATA_DIR         Data directory path (default: "./data")
  STORAGE_SYNC_WRITE       Enable synchronous writes (default: "true")
  STORAGE_CREATE_DIR       Create data directory if missing (default: "true")
  STORAGE_FILE_PERMISSIONS File permissions in octal format (default: "0644")
  STORAGE_DIR_PERMISSIONS  Directory permissions in octal format (default: "0755")
`
}

// NewConfigFromLoader creates a storage manager using distributed configuration.
// This is the recommended factory function when using a configuration loader.
func NewFileStorageManagerFromLoader(loader ConfigLoader) (*FileStorageManager, error) {
	config := NewConfigFromLoader(loader)
	return NewFileStorageManager(config)
}

// NewConfigWithDefaults creates a configuration with both loader and defaults.
// The loader takes precedence over defaults, but environment variables override both.
func NewConfigWithDefaults(loader ConfigLoader, defaults *Config) *Config {
	var config *Config

	if loader != nil {
		config = NewConfigFromLoader(loader)
	} else {
		config = NewConfig()
	}

	// Apply defaults for missing values
	if defaults != nil {
		if config.DataDir == "" {
			config.DataDir = defaults.DataDir
		}
		if config.FilePermissions == 0 {
			config.FilePermissions = defaults.FilePermissions
		}
		if config.DirPermissions == 0 {
			config.DirPermissions = defaults.DirPermissions
		}
	}

	// Apply environment variable overrides
	config.ApplyEnvironmentOverrides()

	return config
}
