# WinPower G2 Exporter

[![CI/CD Pipeline](https://github.com/lay-g/winpower-g2-exporter/actions/workflows/ci.yml/badge.svg)](https://github.com/lay-g/winpower-g2-exporter/actions/workflows/ci.yml)

A Go application for collecting data from WinPower G2 devices, calculating energy consumption, and exporting metrics in Prometheus format.

## Overview

WinPower G2 Exporter provides real-time device status monitoring and energy consumption statistics by periodically collecting device data from WinPower G2 systems. This exporter features a modular design that supports multi-device management and can reliably handle device connections, data collection, energy calculation, and metric exposure.

## AI Project Description

This project is developed with the assistance of Claude Code AI assistant.

## Key Features

- **Unified Data Collection**: Automatically collects WinPower device data every 5 seconds
- **Energy Calculation**: Precise cumulative energy calculation based on power data
- **Prometheus Compatible**: Provides standard `/metrics` endpoint for Prometheus scraping
- **Multi-Device Support**: Supports monitoring multiple WinPower devices simultaneously
- **Self-Authentication Management**: Automatic token refresh and authentication flow handling
- **Structured Logging**: High-performance structured logging based on zap
- **TDD Development**: Test-driven development ensuring code quality

## Architecture Design

### Modular Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    WinPower G2 Exporter                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │   Config    │  │   Logging   │  │      Storage         │   │
│  │ Config Mgmt │  │ Log Module  │  │   Storage Module     │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │  WinPower   │  │  Collector  │  │       Energy         │   │
│  │Data Collect │  │Data Coord   │  │ Energy Calc Module  │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │   Metrics   │  │   Server    │  │     Scheduler       │   │
│  │Metrics Mgmt │  │ HTTP Server │  │  Scheduler Module   │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  WinPower G2    │
                    │     System      │
                    └─────────────────┘
```

### Data Flow

1. **Scheduler**: Triggers data collection every 5 seconds
2. **WinPower Module**: Fetches device data from WinPower system
3. **Collector Module**: Coordinates data collection and energy calculation
4. **Energy Module**: Calculates cumulative energy based on power data
5. **Storage Module**: Persists cumulative energy values
6. **Metrics Module**: Updates Prometheus metrics
7. **Server Module**: Exposes metrics through HTTP endpoints

## Quick Start

### Prerequisites

- Go 1.25+
- WinPower G2 System
- Prometheus (optional, for metric collection)

### Installation

#### Build from Source

```bash
# Clone repository
git clone https://github.com/lay-g/winpower-g2-exporter.git
cd winpower-g2-exporter

# Build
make build

# Run
./bin/winpower-g2-exporter server
```

#### Using Docker

```bash
# Build image
make docker-build

# Run container
make docker-run
```

### Basic Configuration

Create a configuration file `config.yaml`:

```yaml
server:
  port: 9090
  host: "0.0.0.0"

winpower:
  base_url: "https://your-winpower-server.com"
  username: "admin"
  password: "your-password"
  timeout: 30s
  skip_ssl_verify: false

storage:
  data_dir: "./data"
  sync_write: true

scheduler:
  collection_interval: 5s

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Run Service

```bash
# Using configuration file
./winpower-g2-exporter server --config config.yaml

# Or using environment variables
export WINPOWER_EXPORTER_WINPOWER_BASE_URL="https://your-winpower-server.com"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="admin"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="your-password"
./winpower-g2-exporter server
```

### Verify Operation

Access the health check endpoint:
```bash
curl http://localhost:9090/health
```

Access the metrics endpoint:
```bash
curl http://localhost:9090/metrics
```

## Configuration

### Configuration Methods

WinPower G2 Exporter supports multiple configuration methods:

| Method        | Priority | Description                     |
|---------------|----------|---------------------------------|
| CLI Args      | Highest  | Parameters prefixed with `--`   |
| Environment   | Medium   | `WINPOWER_EXPORTER_` prefix     |
| Config File   | Lowest   | YAML format configuration file  |

### Required Configuration

| Parameter              | Environment Variable                          | Description                    | Example                           |
|------------------------|-----------------------------------------------|--------------------------------|-----------------------------------|
| `--winpower.url`       | `WINPOWER_EXPORTER_WINPOWER_URL`              | WinPower service address       | `https://winpower.example.com`    |
| `--winpower.username`  | `WINPOWER_EXPORTER_WINPOWER_USERNAME`         | Username                       | `admin`                           |
| `--winpower.password`  | `WINPOWER_EXPORTER_WINPOWER_PASSWORD`         | Password                       | `secret`                          |

### Optional Configuration

| Parameter                | Environment Variable                           | Default   | Description                    |
|-------------------------|------------------------------------------------|-----------|--------------------------------|
| `--port`                 | `WINPOWER_EXPORTER_PORT`                      | `9090`    | HTTP service port              |
| `--log-level`            | `WINPOWER_EXPORTER_LOGGING_LEVEL`             | `info`    | Log level (debug|info|warn|error) |
| `--skip-ssl-verify`      | `WINPOWER_EXPORTER_SKIP_SSL_VERIFY`           | `false`   | Skip SSL certificate verification |
| `--data-dir`             | `WINPOWER_EXPORTER_STORAGE_DATA_DIR`          | `./data`  | Data storage directory         |
| `--sync-write`           | `WINPOWER_EXPORTER_STORAGE_SYNC_WRITE`        | `true`    | Synchronous write mode         |

### Complete Configuration Example

```yaml
# config.yaml
server:
  port: 9090
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  enable_pprof: false

winpower:
  base_url: "https://winpower.example.com"
  username: "admin"
  password: "secret"
  timeout: 30s
  api_timeout: 10s
  skip_ssl_verify: false
  refresh_threshold: 5m

storage:
  data_dir: "./data"
  file_permissions: 0644

scheduler:
  collection_interval: 5s
  graceful_shutdown_timeout: 5s

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true
  enable_caller: false
  enable_stacktrace: false
```

## Metrics

### Exporter Self-Monitoring Metrics

| Metric Name                                    | Type      | Description                |
|-----------------------------------------------|-----------|----------------------------|
| `winpower_exporter_up`                        | Gauge     | Exporter running status    |
| `winpower_exporter_requests_total`            | Counter   | Total HTTP requests        |
| `winpower_exporter_collection_duration_seconds` | Histogram | Collection + calculation duration |
| `winpower_exporter_scrape_errors_total`       | Counter   | Total collection errors    |
| `winpower_exporter_device_count`              | Gauge     | Number of discovered devices|

### WinPower Connection Metrics

| Metric Name                             | Type      | Description                |
|----------------------------------------|-----------|----------------------------|
| `winpower_connection_status`            | Gauge     | WinPower connection status |
| `winpower_auth_status`                  | Gauge     | Authentication status      |
| `winpower_api_response_time_seconds`   | Histogram | API response latency       |
| `winpower_token_expiry_seconds`        | Gauge     | Token remaining validity   |

### Device Status Metrics

| Metric Name                            | Type  | Description                | Labels                               |
|---------------------------------------|-------|----------------------------|--------------------------------------|
| `winpower_device_connected`            | Gauge | Device connection status   | device_id, device_name, device_type  |
| `winpower_device_load_percent`         | Gauge | Device load percentage     | device_id, device_name, device_type  |
| `winpower_device_load_total_watts`     | Gauge | Total load active power (W)| device_id, device_name, device_type  |
| `winpower_device_cumulative_energy`    | Gauge | Cumulative energy (Wh)     | device_id, device_name, device_type  |

### Electrical Parameter Metrics

| Metric Name                            | Type  | Description            |
|---------------------------------------|-------|------------------------|
| `winpower_device_input_voltage`       | Gauge | Input voltage (V)      |
| `winpower_device_output_voltage`      | Gauge | Output voltage (V)     |
| `winpower_device_output_current`      | Gauge | Output current (A)     |
| `winpower_device_input_frequency`     | Gauge | Input frequency (Hz)   |
| `winpower_device_output_frequency`    | Gauge | Output frequency (Hz)  |

### Battery Parameter Metrics

| Metric Name                            | Type  | Description                    |
|---------------------------------------|-------|--------------------------------|
| `winpower_device_battery_charging`    | Gauge | Battery charging status        |
| `winpower_device_battery_capacity`    | Gauge | Battery capacity (%)           |
| `winpower_device_battery_remain_seconds` | Gauge | Battery remaining time (seconds)|
| `winpower_device_ups_temperature`     | Gauge | UPS temperature (°C)           |

## Deployment Guide

### Docker Deployment

```dockerfile
# Using pre-built image
docker run -d \
  --name winpower-exporter \
  -p 9090:9090 \
  -e WINPOWER_EXPORTER_WINPOWER_BASE_URL="https://winpower.example.com" \
  -e WINPOWER_EXPORTER_WINPOWER_USERNAME="admin" \
  -e WINPOWER_EXPORTER_WINPOWER_PASSWORD="secret" \
  -v $(pwd)/data:/app/data \
  winpower-g2-exporter:latest
```

### Docker Compose Deployment

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  winpower-exporter:
    image: winpower-g2-exporter:latest
    container_name: winpower-exporter
    restart: unless-stopped
    ports:
      - "9090:9090"
    environment:
      - WINPOWER_EXPORTER_WINPOWER_BASE_URL=https://winpower.example.com
      - WINPOWER_EXPORTER_WINPOWER_USERNAME=admin
      - WINPOWER_EXPORTER_WINPOWER_PASSWORD=${WINPOWER_PASSWORD}
      - WINPOWER_EXPORTER_LOGGING_LEVEL=info
      - WINPOWER_EXPORTER_LOGGING_FORMAT=json
    volumes:
      - ./data:/app/data
      - ./config.yaml:/app/config.yaml:ro
```

Create a `.env` file for sensitive information:

```bash
# .env
WINPOWER_PASSWORD=your-secret-password
```

Start the service:

```bash
# Start service
docker-compose up -d

# View logs
docker-compose logs -f winpower-exporter

# Stop service
docker-compose down

# Rebuild and start
docker-compose up -d --build
```

Access the service:

- **Metrics Endpoint**: http://localhost:9090/metrics
- **Health Check**: http://localhost:9090/health

### Grafana Dashboard

Recommended Grafana queries:

```promql
# Device total load power
winpower_device_load_total_watts

# Cumulative energy consumption
winpower_device_cumulative_energy

# Hourly energy consumption
increase(winpower_device_cumulative_energy[1h]) / 1000

# Device connection status
winpower_device_connected == 1

# 5-minute average power
avg_over_time(winpower_device_load_total_watts[5m])

# System health status
winpower_exporter_up == 1
```

## Development Guide

### Project Structure

```
winpower-g2-exporter/
├── cmd/                         # Application entry points
│   └── winpower-g2-exporter/
│       └── main.go               # Main program entry
├── internal/                    # Internal implementation packages
│   ├── cmd/                     # Command line tools implementation
│   ├── pkgs/                    # Internal common packages
│   │   └── log/                 # Logging module
│   ├── config/                  # Configuration management module
│   ├── storage/                 # Storage module
│   ├── winpower/                # WinPower module
│   ├── collector/               # Collector module
│   ├── energy/                  # Energy calculation module
│   ├── metrics/                 # Metrics module
│   ├── server/                  # HTTP service module
│   └── scheduler/               # Scheduler module
├── docs/                        # Project documentation
│   ├── design/                  # Design documents
│   └── examples/                # Usage examples
├── tests/                       # Test files
│   ├── integration/             # Integration tests
│   └── mocks/                   # Mock objects
├── deployments/                 # Deployment configurations
├── Dockerfile                   # Docker build file
├── Makefile                     # Build and development scripts
├── go.mod                       # Go module definition
└── README.md                    # Project documentation
```

### Build Commands

```bash
# Build for current platform
make build

# Build for Linux AMD64
make build-linux

# Build for all supported platforms
make build-all

# Clean build artifacts
make clean

# Format code
make fmt

# Static analysis
make lint

# Install dependencies
make deps

# Update dependencies
make update-deps

# Run tests
make test

# Test coverage
make test-coverage

# Integration tests
make test-integration

# Run all tests
make test-all

# Docker build
make docker-build

# Docker run
make docker-run
```

### Testing

```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Generate test coverage report
make test-coverage

# Run benchmarks
go test -bench=. ./...
```

### Code Standards

- Follow official Go code standards
- Use `gofmt` to format code
- Use `golangci-lint` for static analysis
- Write unit tests maintaining 80%+ test coverage
- Follow Test-Driven Development (TDD) principles

## Troubleshooting

### Common Issues

#### 1. WinPower Connection Failure

**Symptoms**: Logs show authentication failure or connection timeout

**Solutions**:
- Check if WinPower service address is correct
- Verify username and password are correct
- Confirm network connection and firewall settings
- If using self-signed certificates, set `skip_ssl_verify: true`

#### 2. Empty Device Data

**Symptoms**: Metrics endpoint returns data but device metrics are empty

**Solutions**:
- Check if there are devices online in WinPower system
- Verify user permissions are sufficient to access device data
- Check if WinPower API responses are normal

#### 3. Energy Calculation Anomalies

**Symptoms**: Cumulative energy values don't update or calculate incorrectly

**Solutions**:
- Check data directory permissions
- Verify power data is being retrieved normally
- Check Energy module logs for errors

### Log Analysis

```bash
# View error logs
jq 'select(.level == "error")' /var/log/winpower-exporter.log

# View collection duration
jq 'select(.msg == "Device data collected") | .duration_ms' /var/log/winpower-exporter.log

# View authentication status
jq 'select(.msg | contains("Token"))' /var/log/winpower-exporter.log
```

### Health Check

```bash
# Check service status
curl -f http://localhost:9090/health || echo "Service异常"

# Check metrics endpoint
curl -s http://localhost:9090/metrics | head -n 10

# Check process status
ps aux | grep winpower-exporter
```

## Contributing

1. Fork the project
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

### Commit Guidelines

- Use clear commit messages
- One commit per change
- Include relevant test cases
- Follow existing code style

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- Documentation: [Project Documentation](docs/)
- Issues: [GitHub Issues](https://github.com/lay-g/winpower-g2-exporter/issues)
- Discussions: [GitHub Discussions](https://github.com/lay-g/winpower-g2-exporter/discussions)