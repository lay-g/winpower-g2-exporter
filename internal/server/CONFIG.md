# Server Module Configuration Guide

This document provides comprehensive configuration guidance for the HTTP server module, including examples, best practices, and security considerations.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration Reference](#configuration-reference)
- [Environment Variables](#environment-variables)
- [Configuration Examples](#configuration-examples)
- [Security Considerations](#security-considerations)
- [Performance Tuning](#performance-tuning)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Basic Configuration

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/lay-g/winpower-g2-exporter/internal/server"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func main() {
    // Create logger
    logger := log.NewNoop()

    // Create services (mock examples)
    metricsService := &mockMetricsService{}
    healthService := &mockHealthService{}

    // Use default configuration
    config := server.DefaultConfig()

    // Create and start server
    srv := server.NewHTTPServer(config, logger, metricsService, healthService)

    // Start server in background
    go func() {
        if err := srv.Start(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // Graceful shutdown on interrupt
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Stop(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

## Configuration Reference

### Complete Configuration Schema

```yaml
# Server Configuration
port: 9090                          # TCP port (1-65535)
host: "0.0.0.0"                    # Bind address
mode: "release"                    # Gin mode: debug, release, test

# Timeouts (duration format: "30s", "1m", "1h")
read_timeout: "30s"                # Request reading timeout
write_timeout: "30s"               # Response writing timeout
idle_timeout: "60s"                # Keep-alive timeout

# Feature Flags
enable_pprof: false                # Enable /debug/pprof endpoints
enable_cors: false                 # Enable CORS middleware
enable_rate_limit: false           # Enable rate limiting
```

### Configuration Fields Details

| Field | Type | Default | Range/Values | Description |
|-------|------|---------|--------------|-------------|
| `port` | int | 9090 | 1-65535 | TCP port for HTTP server |
| `host` | string | "0.0.0.0" | Valid IP/hostname | Network interface to bind |
| `mode` | string | "release" | debug/release/test | Gin framework runtime mode |
| `read_timeout` | duration | "30s" | > 0s | Maximum request read time |
| `write_timeout` | duration | "30s" | > 0s | Maximum response write time |
| `idle_timeout` | duration | "60s" | > 0s | Keep-alive connection timeout |
| `enable_pprof` | bool | false | true/false | Enable pprof debugging endpoints |
| `enable_cors` | bool | false | true/false | Enable CORS middleware |
| `enable_rate_limit` | bool | false | true/false | Enable rate limiting middleware |

## Environment Variables

The server module supports configuration via environment variables with the `WINPOWER_EXPORTER_` prefix:

### Network Configuration

```bash
# Server listening configuration
WINPOWER_EXPORTER_PORT=9090
WINPOWER_EXPORTER_HOST=0.0.0.0
WINPOWER_EXPORTER_MODE=release
```

### Timeout Configuration

```bash
# Timeout settings (duration format)
WINPOWER_EXPORTER_READ_TIMEOUT=30s
WINPOWER_EXPORTER_WRITE_TIMEOUT=30s
WINPOWER_EXPORTER_IDLE_TIMEOUT=60s
```

### Feature Flags

```bash
# Enable/disable features
WINPOWER_EXPORTER_ENABLE_PPROF=false
WINPOWER_EXPORTER_ENABLE_CORS=false
WINPOWER_EXPORTER_ENABLE_RATE_LIMIT=false
```

### Complete Environment Setup

```bash
export WINPOWER_EXPORTER_PORT=9090
export WINPOWER_EXPORTER_HOST=0.0.0.0
export WINPOWER_EXPORTER_MODE=release
export WINPOWER_EXPORTER_READ_TIMEOUT=30s
export WINPOWER_EXPORTER_WRITE_TIMEOUT=30s
export WINPOWER_EXPORTER_IDLE_TIMEOUT=60s
export WINPOWER_EXPORTER_ENABLE_PPROF=false
export WINPOWER_EXPORTER_ENABLE_CORS=false
export WINPOWER_EXPORTER_ENABLE_RATE_LIMIT=false
```

## Configuration Examples

### Development Environment

```go
config := &server.Config{
    Port:         8080,                 // Development port
    Host:         "127.0.0.1",          // Localhost only
    Mode:         "debug",              // Debug mode with detailed logging
    ReadTimeout:  10 * time.Second,     // Shorter timeouts for development
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  30 * time.Second,
    EnablePprof:  true,                 // Enable debugging
    EnableCORS:   true,                 // Enable for frontend development
}
```

### Production Environment

```go
config := &server.Config{
    Port:         9090,                 // Standard Prometheus port
    Host:         "0.0.0.0",            // All interfaces
    Mode:         "release",            // Optimized performance
    ReadTimeout:  30 * time.Second,     // Production-ready timeouts
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
    EnablePprof:  false,                // Security: disable debugging
    EnableCORS:   false,                // Security: disable unless needed
    EnableRateLimit: true,              // Enable protection
}
```

### High-Security Environment

```go
config := &server.Config{
    Port:         9090,
    Host:         "127.0.0.1",          // Restrict to localhost
    Mode:         "release",
    ReadTimeout:  15 * time.Second,     // Shorter timeouts for security
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  30 * time.Second,
    EnablePprof:  false,                // No debugging in production
    EnableCORS:   false,                // No cross-origin access
    EnableRateLimit: true,              // Protection against abuse
}
```

### Container/Docker Environment

```go
config := &server.Config{
    Port:         9090,
    Host:         "0.0.0.0",            // All interfaces for container networking
    Mode:         "release",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  45 * time.Second,     // Shorter for container environments
    EnablePprof:  false,
    EnableCORS:   false,
    EnableRateLimit: false,             // Handle at reverse proxy level
}
```

### Configuration File Example

Create a `config.yaml` file:

```yaml
server:
  port: 9090
  host: "0.0.0.0"
  mode: "release"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  enable_pprof: false
  enable_cors: false
  enable_rate_limit: false
```

Load with your preferred configuration library:

```go
import "gopkg.in/yaml.v2"

// Load configuration from file
data, err := ioutil.ReadFile("config.yaml")
if err != nil {
    log.Fatal(err)
}

var config struct {
    Server server.Config `yaml:"server"`
}

if err := yaml.Unmarshal(data, &config); err != nil {
    log.Fatal(err)
}

if err := config.Server.Validate(); err != nil {
    log.Fatal(err)
}
```

## Security Considerations

### Network Security

1. **Host Binding**
   - Use `"127.0.0.1"` for local development
   - Use `"0.0.0.0"` only behind a reverse proxy
   - Avoid binding to specific IPs in cloud environments

2. **Port Selection**
   - Avoid privileged ports (< 1024) unless necessary
   - Use standard Prometheus port (9090) for metrics
   - Ensure firewall rules allow the chosen port

3. **Timeout Configuration**
   - Set appropriate timeouts to prevent attacks
   - Consider shorter timeouts for public-facing services
   - Balance between security and usability

### Feature Security

1. **Pprof Debugging**
   ```go
   // ❌ Insecure: Enable pprof in production
   config.EnablePprof = true

   // ✅ Secure: Disable pprof in production
   config.EnablePprof = false

   // ✅ Conditional: Enable only in development
   config.EnablePprof = os.Getenv("ENV") == "development"
   ```

2. **CORS Configuration**
   ```go
   // ❌ Insecure: Enable CORS for all origins
   config.EnableCORS = true

   // ✅ Secure: Disable unless explicitly needed
   config.EnableCORS = false

   // ✅ Conditional: Enable only for frontend
   config.EnableCORS = os.Getenv("ENABLE_CORS") == "true"
   ```

3. **Rate Limiting**
   ```go
   // ✅ Enable rate limiting for production
   config.EnableRateLimit = true

   // ✅ Disable behind reverse proxy with rate limiting
   config.EnableRateLimit = false
   ```

### Production Security Checklist

- [ ] Use `"release"` mode in production
- [ ] Set `EnablePprof = false`
- [ ] Set `EnableCORS = false` unless needed
- [ ] Enable rate limiting if not handled by reverse proxy
- [ ] Configure appropriate timeouts
- [ ] Use TLS termination at reverse proxy
- [ ] Implement proper logging and monitoring
- [ ] Regular security updates and patches

## Performance Tuning

### Timeout Optimization

```go
// High-throughput service
config := &server.Config{
    ReadTimeout:  10 * time.Second,   // Faster request processing
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  30 * time.Second,   // Shorter keep-alive
}

// Service with large responses
config := &server.Config{
    ReadTimeout:  60 * time.Second,   // Support large uploads
    WriteTimeout: 120 * time.Second,  // Support large downloads
    IdleTimeout:  300 * time.Second,  // Longer keep-alive
}
```

### Mode Selection

```go
// Development: Detailed logging and debugging
config.Mode = "debug"

// Production: Optimized performance
config.Mode = "release"

// Testing: Mocked dependencies
config.Mode = "test"
```

### Resource Optimization

```go
// Resource-constrained environment
config := &server.Config{
    IdleTimeout:  15 * time.Second,   // Shorter timeout to free resources
    EnableRateLimit: true,            // Prevent resource exhaustion
}
```

## Troubleshooting

### Common Issues

#### 1. Port Already in Use

**Error**: `bind: address already in use`

**Solutions**:
```bash
# Find process using port
lsof -i :9090

# Kill process (if appropriate)
kill -9 <PID>

# Or use different port
export WINPOWER_EXPORTER_PORT=9091
```

#### 2. Permission Denied

**Error**: `bind: permission denied`

**Solutions**:
```go
// Use non-privileged port (> 1024)
config.Port = 9090

// Or run with appropriate permissions (not recommended)
// sudo ./application
```

#### 3. Invalid Configuration

**Error**: `invalid port: 70000 (must be between 1 and 65535)`

**Solutions**:
```go
// Validate configuration before use
if err := config.Validate(); err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}

// Use valid port range
config.Port = 9090
```

#### 4. Connection Timeouts

**Symptoms**: Requests timing out frequently

**Solutions**:
```go
// Increase timeouts for slow clients or large payloads
config.ReadTimeout = 60 * time.Second
config.WriteTimeout = 60 * time.Second
```

#### 5. High Memory Usage

**Symptoms**: Memory usage increasing over time

**Solutions**:
```go
// Reduce idle timeout to free connections faster
config.IdleTimeout = 30 * time.Second

// Enable rate limiting to prevent abuse
config.EnableRateLimit = true
```

### Debug Configuration

```go
// Enable comprehensive debugging
config := &server.Config{
    Mode:        "debug",
    EnablePprof: true,
}

// Log configuration for debugging
log.Printf("Server configuration: %+v", config)

// Validate and report issues
if err := config.Validate(); err != nil {
    log.Printf("Configuration error: %v", err)
}
```

### Health Check Configuration

Ensure health check endpoint is accessible:

```bash
# Test health endpoint
curl http://localhost:9090/health

# Expected response
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z",
  "details": {...}
}
```

### Metrics Endpoint Configuration

Verify metrics endpoint is working:

```bash
# Test metrics endpoint
curl http://localhost:9090/metrics

# Expected Prometheus format output
# HELP winpower_exporter_up WinPower Exporter is up
# TYPE winpower_exporter_up gauge
winpower_exporter_up 1
```

## Best Practices

1. **Configuration Management**
   - Use environment variables for deployment-specific values
   - Store sensitive configuration in secure vaults
   - Validate configuration at startup
   - Log configuration (without sensitive data)

2. **Security First**
   - Start with most restrictive settings
   - Enable features only when needed
   - Regular security audits
   - Monitor for unusual activity

3. **Performance Optimization**
   - Tune timeouts for your use case
   - Monitor resource usage
   - Test under load
   - Use appropriate Gin mode

4. **Monitoring and Observability**
   - Enable structured logging
   - Monitor key metrics
   - Set up health checks
   - Implement alerting

5. **Deployment Considerations**
   - Use containerization
   - Implement graceful shutdown
   - Handle configuration changes
   - Plan for scalability