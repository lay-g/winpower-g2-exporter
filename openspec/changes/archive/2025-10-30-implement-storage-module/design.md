# Storage Module Design Document

## Overview

This document provides the architectural design rationale for the storage module implementation in the WinPower G2 Exporter project.

## Architectural Decisions

### 1. File-Based Storage Strategy

**Decision**: Use simple text files for data persistence rather than binary formats or databases.

**Rationale**:
- **Simplicity**: Text files are human-readable and easy to debug
- **Reliability**: Fewer moving parts compared to database systems
- **Portability**: Easy to backup, migrate, and inspect
- **Performance**: Sufficient for the low-volume data requirements
- **Maintenance**: No external dependencies or database management overhead

**Trade-offs**:
- Limited to basic read/write operations
- No complex querying capabilities
- Manual file management required

### 2. Per-Device File Organization

**Decision**: Each device gets its own `{device_id}.txt` file.

**Rationale**:
- **Isolation**: Device failures don't affect other devices
- **Scalability**: Easy to add/remove devices
- **Atomicity**: Simpler to ensure data integrity per device
- **Debugging**: Easy to inspect individual device data

**Trade-offs**:
- More files to manage at scale
- Potential filesystem overhead with many devices

### 3. Simple File Format

**Decision**: Two-line format with timestamp and energy value.

```
1694678400000
15000.50
```

**Rationale**:
- **Parsing simplicity**: Easy to read and validate
- **Human readable**: Can be inspected manually
- **Minimal overhead**: No complex serialization needed
- **Backward compatibility**: Simple format is stable

**Trade-offs**:
- Limited extensibility
- No metadata storage
- Fixed structure

### 4. Synchronous Write Operations

**Decision**: Always use synchronous file writes.

**Rationale**:
- **Data integrity**: Ensures data is written to disk
- **Reliability**: No data loss on application crash
- **Simplicity**: No buffering or caching complexity
- **Predictability**: Consistent behavior

**Trade-offs**:
- Potential performance impact
- I/O blocking during writes

### 5. Configuration-Driven Design

**Decision**: External configuration for storage parameters.

**Rationale**:
- **Flexibility**: Easy to adapt to different environments
- **Deployment**: Different configs for dev/staging/prod
- **Maintenance**: No code changes for configuration updates
- **Standards**: Follows project convention patterns

**Trade-offs**:
- Additional configuration management
- Validation complexity

## Interface Design Philosophy

### 1. Interface Segregation

```go
type StorageManager interface {
    Write(deviceID string, data *PowerData) error
    Read(deviceID string) (*PowerData, error)
}
```

**Rationale**: Small, focused interface that's easy to implement and test.

### 2. Dependency Injection

```go
func NewFileStorageManager(config *Config, logger logger.Logger) StorageManager
```

**Rationale**: Enables testing with mock dependencies and flexible configuration.

### 3. Error Handling Strategy

- **Typed errors**: Specific error types for different failure modes
- **Error wrapping**: Preserve context while allowing error inspection
- **Graceful degradation**: Return sensible defaults for missing files

## Testing Strategy

### 1. Test-Driven Development (TDD)

**Approach**: Write tests before implementation code.

**Benefits**:
- Clear requirements definition
- Immediate feedback
- Better API design
- Comprehensive coverage

### 2. Mock-Based Testing

**Approach**: Use mocks for file system operations in unit tests.

**Benefits**:
- Isolation from filesystem
- Predictable test environment
- Edge case simulation
- Fast execution

### 3. Integration Testing

**Approach**: Test with real file operations in temporary directories.

**Benefits**:
- Real-world validation
- Filesystem behavior verification
- Configuration testing
- End-to-end scenarios

## Security Considerations

### 1. File Permissions

- **Configurable permissions**: Allow deployment-specific settings
- **Default secure**: Reasonable default permissions (0644)
- **Validation**: Permission value validation

### 2. Path Validation

- **Input sanitization**: Validate device IDs in file paths
- **Path traversal prevention**: Restrict to configured directory
- **Error handling**: Graceful handling of invalid paths

## Monitoring and Observability

### 1. Logging Integration

- **Structured logging**: Use project's logging module
- **Operation tracking**: Log read/write operations
- **Error reporting**: Detailed error information
- **Performance metrics**: Operation timing

### 2. Error Visibility

- **Typed errors**: Clear error categories
- **Context preservation**: Error wrapping with context
- **Recovery guidance**: Helpful error messages

## Conclusion

This design prioritizes simplicity, reliability, and maintainability over complex features. The file-based approach is well-suited for the project's requirements and provides a solid foundation for future enhancements while meeting current needs effectively.