# Storage Module Implementation Proposal

## Why

The WinPower G2 Exporter currently lacks persistent storage for accumulated energy data, which is essential for providing continuous metrics across application restarts. Without a storage module, energy calculations reset to zero each time the application starts, breaking data continuity and making long-term energy monitoring unreliable.

## Summary

Implement the storage module for WinPower G2 Exporter to provide file-based persistence for accumulated energy data. This module will handle reading and writing power data for multiple devices using individual files.

## Background

The storage module is a critical component of the WinPower G2 Exporter architecture. It provides persistent storage for accumulated energy values, ensuring data continuity across application restarts. Based on the design document in `docs/design/storage.md`, this module will implement a simple, reliable file-based storage system.

## Problem Statement

Currently, the WinPower G2 Exporter lacks persistent storage for energy calculations. Without a storage module:
- Accumulated energy data is lost on application restart
- No historical continuity in energy metrics
- Energy calculations reset to zero each time the application starts

## Solution Overview

Implement a file-based storage module that:
- Stores energy data per device in individual files
- Provides atomic read/write operations
- Handles device file creation and management
- Integrates with the existing logging module
- Follows TDD principles with comprehensive testing

## Scope

### In Scope
- Core storage interfaces and implementations
- File-based persistence for PowerData structures
- Configuration management for storage settings
- Comprehensive unit and integration tests
- Error handling and validation
- Integration with logging module

### Out of Scope
- Data backup mechanisms
- Automatic data cleanup
- Concurrent access optimization (beyond basic file locking)
- Database storage alternatives
- Data compression or encryption

## Architecture

The storage module will consist of:
- **StorageManager**: Primary interface for storage operations
- **FileStorageManager**: Core implementation using file I/O
- **FileWriter/Reader**: Specialized components for file operations
- **Config**: Configuration structure for storage settings
- **PowerData**: Data structure for energy information

## Dependencies

- Go 1.25+
- internal/pkgs/log (logging module)
- os package for file operations
- testing package for TDD implementation

## Success Criteria

1. Successfully store and retrieve PowerData for multiple devices
2. Handle file creation, writing, and reading atomically
3. Provide proper error handling and logging
4. Achieve >80% test coverage
5. Pass all static analysis checks (`make lint`)
6. Integrate cleanly with existing module architecture

## Risks and Mitigations

- **File permission issues**: Configurable file permissions with validation
- **Disk space exhaustion**: Proper error handling and logging
- **Data corruption**: Atomic write operations and validation
- **Concurrent access**: Simple file locking mechanisms

## Implementation Timeline

Estimated 3-5 days for full implementation including testing, based on TDD approach.

## Related Documentation

- `docs/design/storage.md` - Detailed design specification
- `openspec/project.md` - Project conventions and patterns
- `internal/pkgs/log/` - Logging module for integration reference