package storage

// DefaultStorageManager creates a storage manager with default configuration.
// This is the recommended way to create a storage manager for most use cases.
func DefaultStorageManager() (*FileStorageManager, error) {
	config := NewConfig()
	return NewFileStorageManager(config)
}

// StorageManagerWithPath creates a storage manager with a specific data directory.
// This is useful when you want to specify a custom storage location.
func StorageManagerWithPath(dataDir string) (*FileStorageManager, error) {
	config := NewConfigWithPath(dataDir)
	return NewFileStorageManager(config)
}

// NewManager is an alias for NewFileStorageManager for backward compatibility.
func NewManager(config *Config) (*FileStorageManager, error) {
	return NewFileStorageManager(config)
}
