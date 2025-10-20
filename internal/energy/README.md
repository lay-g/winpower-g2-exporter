# Energy Module

The energy module provides power-to-energy conversion and accumulated energy tracking capabilities for WinPower G2 Exporter.

## Overview

This module implements a simple, single-threaded architecture for energy calculation:

- **Serial Execution**: All calculations are executed serially through a global lock to ensure data consistency
- **Storage Integration**: Uses the storage module for data persistence
- **Configurable**: Supports configuration for precision, statistics, and calculation parameters
- **UPS-Focused**: Optimized for UPS device energy monitoring

## Core Features

### Energy Calculation

The module calculates accumulated energy consumption using the formula:

```
Energy (Wh) = Power (W) × Time (h)
```

### Key Operations

- **Calculate(deviceID, power)**: Performs energy accumulation calculation
- **Get(deviceID)**: Retrieves current accumulated energy for a device
- **GetStats()**: Returns calculation statistics

## Usage

### Basic Usage

```go
import (
    "github.com/lay-g/winpower-g2-exporter/internal/energy"
    "github.com/lay-g/winpower-g2-exporter/internal/storage"
    "go.uber.org/zap"
)

// Create dependencies
logger := zap.NewProduction()
storage := storage.NewFileStorageManager(storage.DefaultConfig(), logger)

// Create energy service
energyService := energy.NewEnergyService(storage, logger, energy.DefaultConfig())

// Calculate energy for a device
deviceID := "ups-001"
power := 500.0 // 500 watts

totalEnergy, err := energyService.Calculate(deviceID, power)
if err != nil {
    logger.Error("Energy calculation failed", zap.Error(err))
    return
}

logger.Info("Energy calculation completed",
    zap.String("device_id", deviceID),
    zap.Float64("total_energy", totalEnergy))

// Get current energy data
currentEnergy, err := energyService.Get(deviceID)
if err != nil {
    logger.Error("Failed to get energy data", zap.Error(err))
    return
}

logger.Info("Current energy data",
    zap.String("device_id", deviceID),
    zap.Float64("current_energy", currentEnergy))
```

### Configuration

```go
// Create custom configuration
config := energy.DefaultConfig().
    WithPrecision(0.001).              // Higher precision
    WithStatsEnabled(true).            // Enable statistics
    WithMaxCalculationTime(2 * time.Second). // Timeout
    WithNegativePowerAllowed(true)     // Support energy feedback

energyService := energy.NewEnergyService(storage, logger, config)
```

## Configuration

### Configuration Parameters

| Parameter | Type | Default | Description |
|------------|------|---------|-------------|
| `Precision` | float64 | 0.01 | Calculation precision in Wh |
| `EnableStats` | bool | true | Enable statistics collection |
| `MaxCalculationTime` | int64 | 1s | Maximum calculation time (nanoseconds) |
| `NegativePowerAllowed` | bool | true | Allow negative power values |

### Example Configuration

```yaml
energy:
  precision: 0.001
  enable_stats: true
  max_calculation_time: 1000000000  # 1 second in nanoseconds
  negative_power_allowed: true
```

## Statistics

The module tracks basic calculation statistics when enabled:

- **TotalCalculations**: Total number of calculations performed
- **TotalErrors**: Total number of calculation errors
- **LastUpdateTime**: Timestamp of the last calculation
- **AvgCalculationTime**: Average calculation time in nanoseconds

```go
stats := energyService.GetStats()
logger.Info("Energy service statistics",
    zap.Int64("total_calculations", stats.TotalCalculations),
    zap.Int64("total_errors", stats.TotalErrors),
    zap.Duration("avg_calculation_time", time.Duration(stats.AvgCalculationTime)))
```

## Error Handling

The module provides comprehensive error handling with context:

### Error Types

- `ErrInvalidDeviceID`: Invalid device ID (empty or whitespace)
- `ErrInvalidPowerValue`: Invalid power value (NaN or Inf)
- `ErrNegativePowerNotAllowed`: Negative power when not allowed
- `ErrCalculationTimeout`: Calculation timeout
- `ErrStorageOperation`: Storage operation failure
- `ErrInvalidConfiguration`: Invalid configuration

### Error Handling Example

```go
totalEnergy, err := energyService.Calculate("device-001", 500.0)
if err != nil {
    if errors.Is(err, energy.ErrInvalidPowerValue) {
        logger.Error("Invalid power value provided")
    } else if errors.Is(err, energy.ErrNegativePowerNotAllowed) {
        logger.Error("Negative power not allowed by configuration")
    } else {
        logger.Error("Energy calculation failed", zap.Error(err))
    }
    return
}
```

## Performance

### Performance Characteristics

- **Single Threaded**: All calculations execute serially
- **Low Overhead**: Minimal computational complexity
- **Fast for Small Scale**: Optimized for <20 devices

### Performance Benchmarks

- **Single Calculation**: < 10ms average
- **Concurrent Access**: Safe but serialized
- **Memory Usage**: Minimal

## Integration

### With Storage Module

The energy module integrates seamlessly with the storage module:

```go
// Storage provides persistence for energy data
type StorageManager interface {
    Write(deviceID string, data *PowerData) error
    Read(deviceID string) (*PowerData, error)
}

// Energy module uses storage for data persistence
type PowerData struct {
    Timestamp int64   `json:"timestamp"` // Millisecond timestamp
    EnergyWH  float64 `json:"energy_wh"`  // Cumulative energy in Wh
}
```

### With Collector Module

The energy module is designed to be called by the collector module:

```go
// Collector module calls energy module
func (c *Collector) processDeviceData(device *DeviceData) {
    power := device.PowerInfo.LoadTotalWatt

    // Calculate energy
    totalEnergy, err := c.energyService.Calculate(device.DeviceID, power)
    if err != nil {
        c.logger.Error("Failed to calculate energy", zap.Error(err))
        return
    }

    // Update device data
    device.EnergyInfo.TotalEnergy = totalEnergy
    device.EnergyInfo.CurrentPower = power
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./internal/energy -v

# Run specific tests
go test ./internal/energy -run TestEnergyService_Calculate

# Run with coverage
go test ./internal/energy -cover

# Run benchmarks
go test ./internal/energy -bench=.
```

### Test Coverage

Current test coverage: **76.9%**

Target coverage: **≥90%**

## Architecture

### Design Principles

1. **Simplicity**: Minimal complexity, easy to understand and maintain
2. **Reliability**: Serial execution ensures data consistency
3. **Performance**: Optimized for the intended use case
4. **Testability**: Comprehensive test coverage with mocks

### Thread Safety

- **Global Lock**: All calculations are protected by a single RWMutex
- **Concurrent Reads**: Get operations allow concurrent reads
- **Atomic Operations**: Storage operations are atomic

## Limitations

1. **Single Threaded**: Not suitable for high-concurrency scenarios
2. **Device Scale**: Optimized for <20 devices
3. **Memory-based**: Statistics are kept in memory only
4. **Power Source**: Depends on external power readings

## Best Practices

1. **Error Handling**: Always check and handle errors appropriately
2. **Configuration**: Use appropriate precision for your use case
3. **Monitoring**: Monitor statistics for performance insights
4. **Logging**: Use structured logging for debugging
5. **Testing**: Write comprehensive tests for integration scenarios

## Future Enhancements

- **Distributed Lock**: Support for multi-instance deployments
- **Performance Optimization**: Batch processing for larger scales
- **Advanced Statistics**: More detailed metrics and analytics
- **Configuration Hot Reload**: Runtime configuration updates