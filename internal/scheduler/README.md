# Scheduler Module

The scheduler module is responsible for triggering periodic data collection from the WinPower system. It runs on a fixed interval (default 5 seconds) and calls the WinPower client to collect device data.

## Architecture

The scheduler module follows the dependency injection pattern and uses interfaces to maintain loose coupling with other modules:

### Core Interfaces

- **Scheduler**: Main scheduler interface with Start(), Stop(), and IsRunning() methods
- **WinPowerClient**: Interface for interacting with the WinPower system
- **SchedulerFactory**: Factory interface for creating scheduler instances

### Configuration

The scheduler module configuration includes:

- **CollectionInterval**: Interval between data collection cycles (default: 5 seconds)
- **GracefulShutdownTimeout**: Timeout for graceful shutdown (default: 30 seconds)

## Phase 1 Implementation

Phase 1 includes the basic project structure and configuration module implementation with TDD approach:

### Completed Features

1. **Project Structure**: Complete directory structure with proper Go module organization
2. **Configuration Module**: Full TDD implementation of configuration management
3. **Interface Definitions**: Clear interfaces for all major components
4. **Factory Pattern**: Factory implementation for creating scheduler instances
5. **Mock Objects**: Complete mock implementations for testing

### Files Created

- `interface.go`: Interface definitions for all major components
- `config.go`: Configuration structures, validation, and module registration
- `scheduler.go`: Basic scheduler implementation (placeholder for future phases)
- `factory.go`: Factory implementation for creating schedulers
- `scheduler_test.go`: Comprehensive test suite with TDD approach

### Testing

The module includes comprehensive tests covering:

- Config structure field validation
- Default configuration values
- Configuration validation scenarios (valid/invalid inputs)
- Module configuration registration and management
- Mock implementations for WinPowerClient and Logger interfaces

## Usage

```go
// Create default config
config := scheduler.DefaultConfig()

// Create mock client and logger for testing
client := &MockWinPowerClient{}
logger := &MockLogger{}

// Create scheduler using factory
factory := scheduler.NewDefaultSchedulerFactory()
scheduler, err := factory.CreateScheduler(config, client, logger)
if err != nil {
    log.Fatal(err)
}

// Start scheduler
ctx := context.Background()
err = scheduler.Start(ctx)
if err != nil {
    log.Fatal(err)
}

// Stop scheduler
err = scheduler.Stop()
if err != nil {
    log.Fatal(err)
}
```

## Next Phases

Future phases will implement:

- Phase 2: Scheduler interface and basic structure implementation
- Phase 3: Constructor implementation with dependency injection
- Phase 4: Start functionality with concurrent safety
- Phase 5: Stop functionality with graceful shutdown
- Phase 6: Data collection logic integration
- Phase 7: Edge cases and error handling
- Phase 8: Integration tests
- Phase 9: Performance and resource testing
- Phase 10: Factory pattern implementation
- Phase 11: Documentation and code quality
- Phase 12: Final validation

## Dependencies

- `github.com/lay-g/winpower-g2-exporter/internal/pkgs/log`: Internal logging package
- Standard library: `context`, `time`, `fmt`, `errors`
- Testing: `github.com/stretchr/testify/assert`, `github.com/stretchr/testify/require`