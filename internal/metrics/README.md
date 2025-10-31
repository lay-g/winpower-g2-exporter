# Metrics Module

The Metrics module is responsible for managing and exposing Prometheus metrics for the WinPower G2 Exporter.

## Overview

This module serves as the metrics management center of the application, providing:

- **Prometheus Metrics Registry**: Manages all metric definitions and registrations
- **HTTP Handler**: Provides a standard Gin handler for the `/metrics` endpoint
- **Data Coordination**: Triggers data collection from the Collector module
- **Metric Categories**: Manages both exporter self-monitoring and device metrics

## Architecture

The Metrics module coordinates with the Collector module to fetch the latest device data and updates all metrics accordingly:

```
Prometheus Scrape → HandleMetrics() → Collector.CollectDeviceData()
                                     ↓
                          Update All Metrics
                                     ↓
                     Return Prometheus Format
```

## Metric Categories

### 1. Exporter Self-Monitoring Metrics

These metrics track the health and performance of the exporter itself:

- `winpower_exporter_up`: Exporter running status
- `winpower_exporter_requests_total`: Total HTTP requests
- `winpower_exporter_request_duration_seconds`: Request duration histogram
- `winpower_exporter_collection_duration_seconds`: Collection duration histogram
- `winpower_exporter_scrape_errors_total`: Total scrape errors
- `winpower_exporter_device_count`: Number of discovered devices
- `winpower_exporter_memory_bytes`: Memory usage (optional)

### 2. WinPower Connection Metrics

These metrics track the connection to the WinPower G2 system:

- `winpower_connection_status`: Connection status
- `winpower_auth_status`: Authentication status
- `winpower_api_response_time_seconds`: API response time histogram
- `winpower_token_expiry_seconds`: Token remaining validity
- `winpower_token_valid`: Token validity status

### 3. Device Metrics

These metrics are created dynamically for each discovered device:

**Status Metrics:**
- `winpower_device_connected`: Device connection status
- `winpower_device_last_update_timestamp`: Last update timestamp

**Electrical Parameters:**
- `winpower_device_input_voltage`: Input voltage
- `winpower_device_input_frequency`: Input frequency
- `winpower_device_output_voltage`: Output voltage
- `winpower_device_output_current`: Output current
- `winpower_device_output_frequency`: Output frequency

**Load and Power:**
- `winpower_device_load_percent`: Load percentage
- `winpower_device_load_total_watts`: Total load power (core metric)
- `winpower_power_watts`: Instantaneous power (same as load_total_watts)

**Battery:**
- `winpower_device_battery_charging`: Charging status
- `winpower_device_battery_capacity`: Battery capacity percentage
- `winpower_device_battery_remain_seconds`: Battery remaining time

**UPS Status:**
- `winpower_device_ups_temperature`: UPS temperature
- `winpower_device_ups_mode`: UPS operating mode
- `winpower_device_ups_status`: UPS status
- `winpower_device_ups_fault_code`: UPS fault code

**Energy:**
- `winpower_device_cumulative_energy`: Cumulative energy consumption (Wh)

## Usage

### Basic Usage

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/lay-g/winpower-g2-exporter/internal/metrics"
    "go.uber.org/zap"
)

// Create metrics service
logger := zap.NewProduction()
metricsService, err := metrics.NewMetricsService(collector, logger, nil)
if err != nil {
    log.Fatal(err)
}

// Setup HTTP server
router := gin.Default()
router.GET("/metrics", metricsService.HandleMetrics)
router.Run(":9090")
```

### Custom Configuration

```go
config := &metrics.MetricsConfig{
    Namespace:           "winpower",
    Subsystem:           "exporter",
    WinPowerHost:        "my-winpower-host",
    EnableMemoryMetrics: true,
}

metricsService, err := metrics.NewMetricsService(collector, logger, config)
```

## Labels

### Common Labels

All metrics use consistent labeling:

- **Exporter metrics**: `winpower_host`
- **Connection metrics**: `winpower_host`
- **Device metrics**: `winpower_host`, `device_id`, `device_name`, `device_type`
- **Fault metrics**: Additional `fault_code` label for aggregation

## Testing

The module includes comprehensive tests:

```bash
# Run unit tests
go test -v ./internal/metrics

# Run with coverage
go test -cover ./internal/metrics

# Run integration tests
go test -v ./internal/metrics/... -tags=integration
```

### Mock Collector

For testing, use the provided mock collector:

```go
import "github.com/lay-g/winpower-g2-exporter/internal/metrics/mocks"

mockCollector := mocks.NewMockCollectorWithDevices()
service, err := metrics.NewMetricsService(mockCollector, logger, nil)
```

## Error Handling

The module defines specific errors:

- `ErrCollectorNil`: Collector interface is nil
- `ErrLoggerNil`: Logger is nil
- `ErrCollectionFailed`: Data collection failed
- `ErrMetricsUpdateFailed`: Metrics update failed
- `ErrInvalidCollectionResult`: Collection result is invalid

## Performance Considerations

1. **Concurrent Access**: The module uses read-write locks to protect concurrent access to device metrics
2. **Dynamic Metrics**: Device metrics are created on-demand and cached for subsequent updates
3. **Memory Management**: Optional memory metrics can be enabled to monitor exporter memory usage
4. **Histogram Buckets**: Optimized bucket configurations for duration metrics

## Dependencies

- `github.com/prometheus/client_golang/prometheus`: Prometheus client library
- `github.com/gin-gonic/gin`: Gin web framework for HTTP handler
- `go.uber.org/zap`: Structured logging
- Internal dependencies: `collector` module for data collection interface

## Design Documents

For detailed design information, see:
- [Metrics Module Design](../../docs/design/metrics.md)
- [Architecture Overview](../../docs/design/architecture.md)
