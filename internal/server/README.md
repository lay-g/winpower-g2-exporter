# HTTP Server Module

This module provides a high-performance HTTP server implementation for the WinPower G2 Prometheus Exporter. It offers minimal and stable web interfaces with comprehensive middleware support and graceful lifecycle management.

## 🚀 Quick Start

```go
import (
    "context"
    "log"
    "time"

    "github.com/lay-g/winpower-g2-exporter/internal/server"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// Create server with default configuration
config := server.DefaultConfig()
srv := server.NewHTTPServer(config, logger, metricsService, healthService)

// Start server
go func() {
    if err := srv.Start(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// Graceful shutdown
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Stop(ctx)
```

## 📋 Features

- **High Performance**: Based on Gin framework with optimized middleware chain
- **Secure**: Security-first defaults with optional features
- **Observable**: Structured logging and comprehensive metrics
- **Production Ready**: Graceful shutdown with timeout control
- **Extensible**: Dependency injection and middleware customization
- **Developer Friendly**: Rich debugging and profiling options

## 🌐 Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Application health check (JSON) |
| `/metrics` | GET | Prometheus metrics (text/plain) |
| `/debug/pprof/*` | GET | Performance profiling (optional) |

## ⚙️ Configuration

### Basic Configuration

```go
config := &server.Config{
    Port:         9090,
    Host:         "0.0.0.0",
    Mode:         "release",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
    EnablePprof:  false,
    EnableCORS:   false,
    EnableRateLimit: false,
}
```

### Environment Variables

```bash
WINPOWER_EXPORTER_PORT=9090
WINPOWER_EXPORTER_HOST=0.0.0.0
WINPOWER_EXPORTER_MODE=release
WINPOWER_EXPORTER_READ_TIMEOUT=30s
WINPOWER_EXPORTER_WRITE_TIMEOUT=30s
WINPOWER_EXPORTER_IDLE_TIMEOUT=60s
WINPOWER_EXPORTER_ENABLE_PPROF=false
WINPOWER_EXPORTER_ENABLE_CORS=false
WINPOWER_EXPORTER_ENABLE_RATE_LIMIT=false
```

### Configuration Examples

```go
// Development
config := server.ExampleConfigDevelopment()

// Production
config := server.ExampleConfigProduction()

// High Security
config := server.ExampleConfigHighSecurity()

// Container/Docker
config := server.ExampleConfigContainer()

// High Throughput
config := server.ExampleConfigHighThroughput()
```

### Builder Pattern

```go
config, err := server.NewConfigurationBuilder().
    WithPort(8080).
    WithHost("127.0.0.1").
    WithDebugMode().
    WithPprof(true).
    WithSecurity().
    Build()
```

## 🔧 Interfaces

### MetricsService

```go
type MetricsService interface {
    Render(ctx context.Context) (string, error)
}
```

### HealthService

```go
type HealthService interface {
    Check(ctx context.Context) (status string, details map[string]any)
}
```

### Server

```go
type Server interface {
    Start() error
    Stop(ctx context.Context) error
}
```

## 📚 Documentation

- **[API Documentation](doc.go)** - Complete API reference with examples
- **[Configuration Guide](CONFIG.md)** - Comprehensive configuration options
- **[Examples](example_test.go)** - Usage examples and best practices
- **[Configuration Examples](examples.go)** - Predefined configurations for different scenarios

## 🛡️ Security

- Security-first defaults (debugging disabled in production)
- Configurable CORS support
- Rate limiting middleware
- Timeout protection against slow attacks
- Structured error handling without information leakage

## 🔍 Development

### Running Tests

```bash
# Unit tests
go test ./internal/server/...

# Tests with coverage
go test -cover ./internal/server/...

# Integration tests
go test -tags=integration ./internal/server/...

# Benchmark tests
go test -bench=. ./internal/server/...
```

### Debug Mode

```go
config := &server.Config{
    Mode:        "debug",
    EnablePprof:  true,
    EnableCORS:   true,
}
```

Access profiling endpoints:
- http://localhost:9090/debug/pprof/

## 🚀 Production Deployment

### Recommended Configuration

```go
config := &server.Config{
    Port:         9090,
    Host:         "0.0.0.0",
    Mode:         "release",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
    EnablePprof:  false,
    EnableCORS:   false,
    EnableRateLimit: true,
}
```

### Behind Reverse Proxy

```go
config := server.ExampleConfigBehindProxy()
```

### Docker Deployment

```dockerfile
FROM alpine:latest
COPY ./winpower-exporter /app/
EXPOSE 9090
CMD ["/app/winpower-exporter", "--port=9090"]
```

## 📊 Monitoring

### Health Check

```bash
curl http://localhost:9090/health
```

Response:
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z",
  "details": {
    "version": "1.0.0",
    "uptime": "2h30m45s"
  }
}
```

### Metrics

```bash
curl http://localhost:9090/metrics
```

Response (Prometheus format):
```
# HELP winpower_exporter_up WinPower Exporter is up
# TYPE winpower_exporter_up gauge
winpower_exporter_up 1

# HELP winpower_devices_total Total number of devices
# TYPE winpower_devices_total gauge
winpower_devices_total 5
```

## 🐛 Troubleshooting

### Common Issues

1. **Port already in use**
   ```bash
   lsof -i :9090
   kill -9 <PID>
   ```

2. **Permission denied**
   ```bash
   # Use non-privileged port
   config.Port = 9090
   ```

3. **Configuration validation error**
   ```go
   if err := config.Validate(); err != nil {
       log.Fatal(err)
   }
   ```

### Debug Configuration

```go
// Enable debugging
config.Mode = "debug"
config.EnablePprof = true

// Log configuration
log.Printf("Server config: %+v", config)
```

## 🤝 Contributing

1. Follow existing code style and patterns
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure all tests pass before submitting

## 📄 License

This module is part of the WinPower G2 Prometheus Exporter project.