# WinPower Module Documentation

## Overview

The WinPower module is responsible for collecting device data from WinPower G2 systems. It provides a complete implementation for authentication, data collection, parsing, and error handling.

## Architecture

The module is organized into several key components:

```
WinPower Module
├── Client          - Main entry point for data collection
├── HTTPClient      - HTTP communication layer
├── TokenManager    - Authentication and token lifecycle management
├── DataParser      - Response parsing and data transformation
├── DataValidator   - Data validation and verification
└── Mocks           - Mock implementations for testing
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/lay-g/winpower-g2-exporter/internal/winpower"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/log as logpkg
)

func main() {
    // Create logger
    logger, _ := logpkg.NewLogger(&logpkg.Config{
        Level:  "info",
        Format: "json",
    })

    // Create configuration
    config := &winpower.Config{
        BaseURL:          "https://winpower.example.com",
        Username:         "admin",
        Password:         "password",
        Timeout:          15 * time.Second,
        SkipSSLVerify:    false,
        RefreshThreshold: 5 * time.Minute,
    }

    // Create client
    client, err := winpower.NewClient(config, logger)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Collect device data
    ctx := context.Background()
    data, err := client.CollectDeviceData(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // Process data
    for _, device := range data {
        log.Printf("Device %s: %.2fW\n",
            device.DeviceID,
            device.RealtimeData.LoadTotalWatt)
    }
}
```

## Configuration

### Config Structure

```go
type Config struct {
    BaseURL          string        // WinPower API base URL (required)
    Username         string        // Authentication username (required)
    Password         string        // Authentication password (required)
    Timeout          time.Duration // HTTP request timeout (default: 15s)
    SkipSSLVerify    bool          // Skip SSL certificate verification (default: false)
    RefreshThreshold time.Duration // Token refresh threshold (default: 5m)
}
```

### Configuration Examples

#### Production Configuration

```yaml
winpower:
  base_url: "https://winpower.example.com"
  username: "admin"
  password: "secure_password"
  timeout: 15s
  skip_ssl_verify: false
  refresh_threshold: 5m
```

#### Development/Testing Configuration

```yaml
winpower:
  base_url: "https://test.winpower.local"
  username: "test_user"
  password: "test_pass"
  timeout: 30s
  skip_ssl_verify: true  # Only for self-signed certificates in testing
  refresh_threshold: 3m
```

### Environment Variables

You can also configure using environment variables:

```bash
WINPOWER_BASE_URL="https://winpower.example.com"
WINPOWER_USERNAME="admin"
WINPOWER_PASSWORD="password"
WINPOWER_TIMEOUT="15s"
WINPOWER_SKIP_SSL_VERIFY="false"
WINPOWER_REFRESH_THRESHOLD="5m"
```

## API Reference

### Client Interface

The main interface for interacting with WinPower systems.

#### Methods

##### CollectDeviceData

Collects device data from WinPower system.

```go
func (c *Client) CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error)
```

**Parameters:**
- `ctx`: Context for cancellation and timeouts

**Returns:**
- `[]ParsedDeviceData`: Array of parsed device data
- `error`: Error if collection fails

**Example:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

data, err := client.CollectDeviceData(ctx)
if err != nil {
    // Handle error
    return err
}
```

##### GetConnectionStatus

Returns the current connection status.

```go
func (c *Client) GetConnectionStatus() bool
```

**Returns:**
- `bool`: true if connected, false otherwise

##### GetLastCollectionTime

Returns the timestamp of the last successful data collection.

```go
func (c *Client) GetLastCollectionTime() time.Time
```

**Returns:**
- `time.Time`: Timestamp of last collection

##### GetStatistics

Returns collection statistics.

```go
func (c *Client) GetStatistics() map[string]interface{}
```

**Returns:**
- `map[string]interface{}`: Statistics including total calls, successful calls, failed calls, etc.

##### Close

Closes the client and releases resources.

```go
func (c *Client) Close() error
```

**Returns:**
- `error`: Error if closing fails

### Data Structures

#### ParsedDeviceData

Represents parsed device information.

```go
type ParsedDeviceData struct {
    DeviceID     string       // Unique device identifier
    DeviceType   int          // Device type code
    Model        string       // Device model
    Alias        string       // Device alias/name
    Connected    bool         // Connection status
    RealtimeData RealtimeData // Real-time monitoring data
}
```

#### RealtimeData

Real-time monitoring data from the device.

```go
type RealtimeData struct {
    // Power measurements
    LoadTotalWatt float64 // Total load in watts
    LoadWatt1     float64 // Load phase 1 in watts
    LoadTotalVa   float64 // Total load in VA
    LoadVa1       float64 // Load VA phase 1

    // Voltage measurements
    InputVolt1  float64 // Input voltage phase 1
    OutputVolt1 float64 // Output voltage phase 1
    BatVoltP    float64 // Battery voltage

    // Current measurements
    OutputCurrent1 float64 // Output current phase 1

    // Frequency
    InputFreq  float64 // Input frequency (Hz)
    OutputFreq float64 // Output frequency (Hz)

    // Load percentage
    LoadPercent float64 // Load percentage (0-100)

    // Battery information
    BatCapacity    float64 // Battery capacity percentage (0-100)
    BatRemainTime  int     // Battery remaining time (minutes)
    IsCharging     bool    // Battery charging status
    BatteryStatus  int     // Battery status code

    // Temperature
    UpsTemperature float64 // UPS temperature (Celsius)

    // Status
    Mode      int    // Operating mode
    Status    int    // Device status
    FaultCode string // Fault code if any
}
```

## Error Handling

### Error Types

The module defines several error types for different scenarios:

#### AuthenticationError

Authentication-related errors.

```go
type AuthenticationError struct {
    Message string
    Err     error
}
```

**When it occurs:**
- Invalid credentials
- Token expired
- Authentication service unavailable

**Example handling:**
```go
data, err := client.CollectDeviceData(ctx)
if err != nil {
    if winpower.IsAuthenticationError(err) {
        // Handle authentication error
        log.Println("Authentication failed, please check credentials")
        return err
    }
}
```

#### NetworkError

Network-related errors.

```go
type NetworkError struct {
    Message string
    Err     error
}
```

**When it occurs:**
- Network connectivity issues
- Connection timeout
- DNS resolution failure

**Example handling:**
```go
data, err := client.CollectDeviceData(ctx)
if err != nil {
    if winpower.IsNetworkError(err) {
        // Handle network error
        log.Println("Network error, will retry later")
        return err
    }
}
```

#### ParseError

Data parsing errors.

```go
type ParseError struct {
    Message string
    Err     error
}
```

**When it occurs:**
- Invalid JSON response
- Unexpected data format
- Missing required fields

**Example handling:**
```go
data, err := client.CollectDeviceData(ctx)
if err != nil {
    if winpower.IsParseError(err) {
        // Handle parse error
        log.Println("Failed to parse response")
        return err
    }
}
```

### Error Checking Utilities

```go
// Check if error is authentication related
func IsAuthenticationError(err error) bool

// Check if error is network related
func IsNetworkError(err error) bool

// Check if error is parse related
func IsParseError(err error) bool

// Check if error is config related
func IsConfigError(err error) bool
```

## Advanced Usage

### Context and Cancellation

Always use context for cancellation and timeout control:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

data, err := client.CollectDeviceData(ctx)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("Collection timed out")
    }
}

// With cancellation
ctx, cancel := context.WithCancel(context.Background())

// Start collection in goroutine
go func() {
    data, err := client.CollectDeviceData(ctx)
    // Process data
}()

// Cancel if needed
cancel()
```

### Connection Status Monitoring

```go
// Check connection status before collection
if !client.GetConnectionStatus() {
    log.Println("Client not connected")
    // Handle disconnection
}

// Get last collection time
lastTime := client.GetLastCollectionTime()
if time.Since(lastTime) > 5*time.Minute {
    log.Println("No recent data collection")
}
```

### Statistics Monitoring

```go
stats := client.GetStatistics()
log.Printf("Total calls: %v\n", stats["total_calls"])
log.Printf("Successful: %v\n", stats["successful_calls"])
log.Printf("Failed: %v\n", stats["failed_calls"])
log.Printf("Total devices: %v\n", stats["total_devices"])
```

## Testing

### Using Mock Client

The module provides mock implementations for testing:

```go
import (
    "testing"
    "context"

    "github.com/lay-g/winpower-g2-exporter/internal/winpower"
    "github.com/lay-g/winpower-g2-exporter/internal/winpower/mocks"
)

func TestMyFunction(t *testing.T) {
    // Create mock client
    mockClient := mocks.NewMockWinPowerClient()

    // Configure mock data
    mockClient.SetMockData([]winpower.ParsedDeviceData{
        {
            DeviceID: "test-device-001",
            RealtimeData: winpower.RealtimeData{
                LoadTotalWatt: 1000.0,
            },
        },
    })

    // Test your function
    data, err := mockClient.CollectDeviceData(context.Background())
    if err != nil {
        t.Fatal(err)
    }

    if len(data) != 1 {
        t.Errorf("Expected 1 device, got %d", len(data))
    }
}
```

### Testing Error Scenarios

```go
func TestErrorHandling(t *testing.T) {
    mockClient := mocks.NewMockWinPowerClient()

    // Configure to return error
    mockClient.SetMockError(&winpower.NetworkError{
        Message: "connection refused",
    })

    // Test error handling
    _, err := mockClient.CollectDeviceData(context.Background())
    if err == nil {
        t.Error("Expected error, got nil")
    }

    if !winpower.IsNetworkError(err) {
        t.Error("Expected NetworkError")
    }
}
```

## Best Practices

### 1. Always Use Context

```go
// Good
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
data, err := client.CollectDeviceData(ctx)

// Bad
data, err := client.CollectDeviceData(context.Background())
```

### 2. Handle Errors Appropriately

```go
data, err := client.CollectDeviceData(ctx)
if err != nil {
    if winpower.IsAuthenticationError(err) {
        // Re-authenticate or alert
    } else if winpower.IsNetworkError(err) {
        // Retry with backoff
    } else {
        // Log and continue
    }
}
```

### 3. Close Client Properly

```go
client, err := winpower.NewClient(config, logger)
if err != nil {
    return err
}
defer client.Close()
```

### 4. Monitor Connection Status

```go
if !client.GetConnectionStatus() {
    log.Warn("WinPower client disconnected")
    // Attempt reconnection or alert
}
```

### 5. Use Appropriate Timeouts

```go
config := &winpower.Config{
    BaseURL:          "https://winpower.example.com",
    Timeout:          15 * time.Second,  // Reasonable timeout
    RefreshThreshold: 5 * time.Minute,   // Refresh before expiry
}
```

## Troubleshooting

### Common Issues

#### Issue: Authentication Failures

**Symptoms:**
- `AuthenticationError` returned
- Logs show "authentication failed"

**Solutions:**
1. Verify credentials are correct
2. Check WinPower system is accessible
3. Ensure account has necessary permissions

#### Issue: Connection Timeouts

**Symptoms:**
- `NetworkError` with "timeout" message
- Operations take too long

**Solutions:**
1. Increase timeout in configuration
2. Check network connectivity
3. Verify WinPower system is responsive

#### Issue: SSL Certificate Errors

**Symptoms:**
- `NetworkError` with "certificate" message
- Cannot connect to HTTPS endpoint

**Solutions:**
1. Install proper SSL certificates
2. Use `skip_ssl_verify: true` for self-signed certificates (testing only)
3. Configure proper certificate chain

#### Issue: Parse Errors

**Symptoms:**
- `ParseError` returned
- Logs show "failed to parse"

**Solutions:**
1. Verify WinPower API version compatibility
2. Check response format hasn't changed
3. Enable debug logging to inspect raw responses

### Debug Logging

Enable debug logging to troubleshoot issues:

```go
logger, _ := logpkg.NewLogger(&logpkg.Config{
    Level:  "debug",  // Enable debug level
    Format: "json",
})
```

This will log detailed information about:
- HTTP requests and responses
- Token management operations
- Data parsing steps
- Error details

## Performance Considerations

### Connection Reuse

The module automatically reuses HTTP connections:
- Single `http.Client` instance per WinPower client
- Automatic connection pooling
- Keep-Alive enabled by default

### Token Caching

Tokens are cached and reused:
- Automatic refresh before expiration
- Thread-safe token access
- Configurable refresh threshold

### Memory Usage

The module is designed for low memory footprint:
- No historical data caching
- Efficient JSON parsing
- Prompt garbage collection

## Security Considerations

### Credential Management

- Never hardcode credentials in source code
- Use environment variables or secure configuration management
- Rotate credentials regularly

### SSL/TLS

- Always use HTTPS in production
- Only use `skip_ssl_verify: true` for testing with self-signed certificates
- Keep certificates up to date

### Logging

- Passwords and tokens are automatically sanitized in logs
- Use structured logging for security auditing
- Configure log retention appropriately

## Migration Guide

### From Previous Versions

If migrating from an older version:

1. Update configuration format
2. Update error handling to use new error types
3. Update context usage
4. Test with new mock objects

## Support

For issues or questions:
- Check the troubleshooting section
- Review the examples in `examples/` directory
- Open an issue on GitHub

## License

This module is part of the WinPower G2 Exporter project.
See the main project LICENSE file for details.
