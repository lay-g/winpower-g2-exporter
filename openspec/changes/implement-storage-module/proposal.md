# Storage Module Implementation Proposal

## Why
WinPower G2 Exporter currently lacks data persistence capabilities, causing energy data to be lost on service restart and preventing cumulative energy monitoring.

## What Changes
- Add new storage module in `internal/storage/` with file-based persistence
- Implement distributed module configuration with storage.Config structure
- Add StorageManager interface and PowerData structure
- Integrate storage module with energy module for data persistence
- Add comprehensive test coverage and mock implementations
- Update storage configuration to follow new modular config design
- **Remove all configuration loading logic from storage module** (优化任务)
- **Ensure storage module only defines Config struct without handling config sources**
- **Simplify storage module to pure module design with external configuration**

## Design Principles
- **Simplicity**: Simple file-based design, easy to understand and maintain
- **Reliability**: Basic file read/write functionality ensuring data persistence
- **Multi-device support**: Each device uses independent files for data storage
- **Atomicity**: File write operations have atomicity guarantees
- **Modular configuration**: Storage module owns its configuration structure
- **Testability**: Clean interface abstractions support unit and integration testing

## Impact
- Affected specs: storage-module (new capability)
- Affected code:
  - `internal/storage/` (new module with own config)
  - `internal/energy/` (integration)
  - `internal/pkgs/config/` (config loader integration)
  - `cmd/exporter/` (configuration initialization)