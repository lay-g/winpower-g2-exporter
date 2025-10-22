# WinPower G2 Exporter Configuration Guide

This guide explains how to configure the WinPower G2 Exporter using multiple configuration sources with a clear priority system.

## Configuration Priority System

The exporter loads configuration from multiple sources in the following priority order (highest to lowest):

1. **Command Line Flags** (highest priority)
2. **Environment Variables** (`WINPOWER_EXPORTER_*` prefix)
3. **Configuration Files** (YAML format)
4. **Default Values** (lowest priority)

Higher priority sources override lower priority sources, providing flexible configuration management.

## Configuration Files

### File Locations

The exporter automatically looks for configuration files in these locations (in order):

1. `config.yaml` (current directory)
2. `config.yml` (current directory)
3. `/etc/winpower-exporter/config.yaml`
4. `~/.config/winpower-exporter/config.yaml`

### Configuration File Format

See [`config.yaml.example`](./config.yaml.example) for a complete, annotated configuration example with all available options.

## Environment Variables

All environment variables use the `WINPOWER_EXPORTER_` prefix and override YAML configuration file settings.

### Environment Variable Naming Convention

Environment variables follow this pattern:
```
WINPOWER_EXPORTER_<MODULE>_<FIELD_NAME>
```

Where:
- `<MODULE>` is the configuration module (STORAGE, WINPOWER, ENERGY, SERVER, SCHEDULER)
- `<FIELD_NAME>` is the field name in uppercase

### Complete Environment Variable Reference

#### Storage Configuration
- `WINPOWER_EXPORTER_STORAGE_DATA_DIR` - Data directory path
- `WINPOWER_EXPORTER_STORAGE_CREATE_DIR` - Create directory if missing (true/false)
- `WINPOWER_EXPORTER_STORAGE_SYNC_WRITE` - Synchronous writes for data safety (true/false)
- `WINPOWER_EXPORTER_STORAGE_FILE_PERMISSIONS` - File permissions in octal (e.g., 0644)
- `WINPOWER_EXPORTER_STORAGE_DIR_PERMISSIONS` - Directory permissions in octal (e.g., 0755)

#### WinPower Connection
- `WINPOWER_EXPORTER_WINPOWER_URL` - WinPower API URL (REQUIRED)
- `WINPOWER_EXPORTER_WINPOWER_USERNAME` - API username (REQUIRED)
- `WINPOWER_EXPORTER_WINPOWER_PASSWORD` - API password (REQUIRED)
- `WINPOWER_EXPORTER_WINPOWER_TIMEOUT` - HTTP request timeout (e.g., 30s, 1m)
- `WINPOWER_EXPORTER_WINPOWER_MAX_RETRIES` - Maximum retry attempts (integer)
- `WINPOWER_EXPORTER_WINPOWER_SKIP_TLS_VERIFY` - Skip TLS verification (true/false)

#### Energy Calculation
- `WINPOWER_EXPORTER_ENERGY_PRECISION` - Calculation precision in watt-hours (float)
- `WINPOWER_EXPORTER_ENERGY_ENABLE_STATS` - Enable statistics collection (true/false)
- `WINPOWER_EXPORTER_ENERGY_MAX_CALCULATION_TIME` - Max calculation time in nanoseconds
- `WINPOWER_EXPORTER_ENERGY_NEGATIVE_POWER_ALLOWED` - Allow negative power for energy feedback (true/false)

#### Scheduler Configuration
- `WINPOWER_EXPORTER_SCHEDULER_COLLECTION_INTERVAL` - Collection interval (fixed at 5s)
- `WINPOWER_EXPORTER_SCHEDULER_GRACEFUL_SHUTDOWN_TIMEOUT` - Graceful shutdown timeout

#### HTTP Server Configuration
- `WINPOWER_EXPORTER_SERVER_PORT` - HTTP server port (integer)
- `WINPOWER_EXPORTER_SERVER_HOST` - Listen address (e.g., 0.0.0.0, 127.0.0.1)
- `WINPOWER_EXPORTER_SERVER_MODE` - Server mode (debug/release/test)
- `WINPOWER_EXPORTER_SERVER_READ_TIMEOUT` - HTTP read timeout (e.g., 30s)
- `WINPOWER_EXPORTER_SERVER_WRITE_TIMEOUT` - HTTP write timeout (e.g., 30s)
- `WINPOWER_EXPORTER_SERVER_IDLE_TIMEOUT` - Connection idle timeout (e.g., 60s)
- `WINPOWER_EXPORTER_SERVER_ENABLE_PPROF` - Enable pprof endpoints (true/false)
- `WINPOWER_EXPORTER_SERVER_ENABLE_CORS` - Enable CORS headers (true/false)
- `WINPOWER_EXPORTER_SERVER_ENABLE_RATE_LIMIT` - Enable rate limiting (true/false)
- `WINPOWER_EXPORTER_SERVER_RATE_LIMIT_RPS` - Rate limit requests per second (if enabled)

## Command Line Options

The exporter supports comprehensive command line options for all configuration parameters:

### Basic Options
- `-config string` - Path to configuration file (YAML)
- `-log-level string` - Log level (debug, info, warn, error)
- `-version` - Show version information
- `-help` - Show help information

### Module-Specific Options

#### WinPower Module
- `-winpower-url string` - WinPower server URL
- `-winpower-username string` - WinPower username
- `-winpower-password string` - WinPower password
- `-winpower-timeout duration` - HTTP request timeout
- `-winpower-max-retries int` - Maximum retry attempts
- `-winpower-skip-ssl-verify` - Skip TLS certificate verification

#### Storage Module
- `-storage-data-dir string` - Storage data directory path
- `-storage-sync-write` - Enable synchronous writes

#### Server Module
- `-port int` - Server port (default: 9090)
- `-host string` - Server host (default: 0.0.0.0)
- `-server-mode string` - Server mode (debug/release/test)

#### Scheduler Module
- `-scheduler-interval duration` - Collection interval
- `-energy-precision float` - Energy calculation precision

## Usage Examples

### Basic Usage with Configuration File

```bash
# Use default configuration file location
./winpower-g2-exporter

# Use custom configuration file
./winpower-g2-exporter -config /etc/winpower-exporter/config.yaml
```

### Environment Variables Only

```bash
# Set required WinPower connection parameters
export WINPOWER_EXPORTER_WINPOWER_URL="https://winpower.example.com:8443"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="admin"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="your-secure-password"

# Optional server configuration
export WINPOWER_EXPORTER_SERVER_PORT="9090"
export WINPOWER_EXPORTER_STORAGE_DATA_DIR="/var/lib/winpower-exporter"

# Start the exporter
./winpower-g2-exporter
```

### Mixed Configuration (File + Environment Overrides)

```bash
# Base configuration in config.yaml
cat > config.yaml << EOF
storage:
  data_dir: "./data"
  sync_write: true

winpower:
  url: "https://winpower-dev.example.com:8443"
  username: "admin"
  password: "dev-password"
  timeout: "30s"

server:
  port: 9090
  host: "0.0.0.0"
  mode: "release"

scheduler:
  collection_interval: "5s"

energy:
  precision: 0.01
  enable_stats: true
EOF

# Override production-specific settings with environment variables
export WINPOWER_EXPORTER_WINPOWER_URL="https://winpower-prod.example.com:8443"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="prod-password"
export WINPOWER_EXPORTER_SERVER_PORT="8080"
export WINPOWER_EXPORTER_STORAGE_DATA_DIR="/var/lib/winpower-exporter"

# Start with mixed configuration
./winpower-g2-exporter -config config.yaml
```

### Command Line Overrides

```bash
# Override specific settings with CLI flags
./winpower-g2-exporter \
  -config config.yaml \
  -winpower-url "https://winpower.example.com:8443" \
  -winpower-username "admin" \
  -winpower-password "secret" \
  -port 8080 \
  -log-level debug
```

### Docker Deployment

```dockerfile
# Dockerfile example
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY winpower-g2-exporter /usr/local/bin/
COPY config.yaml.example /etc/winpower-exporter/config.yaml

# Set environment variables for container deployment
ENV WINPOWER_EXPORTER_WINPOWER_URL="https://winpower.example.com:8443"
ENV WINPOWER_EXPORTER_WINPOWER_USERNAME="admin"
ENV WINPOWER_EXPORTER_WINPOWER_PASSWORD="docker-password"
ENV WINPOWER_EXPORTER_STORAGE_DATA_DIR="/data"

EXPOSE 9090
CMD ["winpower-g2-exporter", "-config", "/etc/winpower-exporter/config.yaml"]
```

```bash
# Docker run with environment variables
docker run -d \
  --name winpower-exporter \
  -p 9090:9090 \
  -e WINPOWER_EXPORTER_WINPOWER_URL="https://winpower.example.com:8443" \
  -e WINPOWER_EXPORTER_WINPOWER_USERNAME="admin" \
  -e WINPOWER_EXPORTER_WINPOWER_PASSWORD="secure-password" \
  -v /var/lib/winpower-exporter:/data \
  winpower-g2-exporter:latest
```

### Kubernetes Deployment

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: winpower-exporter-config
data:
  config.yaml: |
    storage:
      data_dir: "/data"
      sync_write: true
    winpower:
      url: "https://winpower.example.com:8443"
      timeout: "30s"
      max_retries: 3
    server:
      port: 9090
      host: "0.0.0.0"
      mode: "release"
    scheduler:
      collection_interval: "5s"
    energy:
      precision: 0.01
      enable_stats: true

---
# secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: winpower-exporter-secrets
type: Opaque
data:
  username: YWRtaW4=  # base64 encoded "admin"
  password: c2VjdXJlLXBhc3N3b3Jk  # base64 encoded password

---
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: winpower-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: winpower-exporter
  template:
    metadata:
      labels:
        app: winpower-exporter
    spec:
      containers:
      - name: winpower-exporter
        image: winpower-g2-exporter:latest
        ports:
        - containerPort: 9090
        env:
        - name: WINPOWER_EXPORTER_WINPOWER_USERNAME
          valueFrom:
            secretKeyRef:
              name: winpower-exporter-secrets
              key: username
        - name: WINPOWER_EXPORTER_WINPOWER_PASSWORD
          valueFrom:
            secretKeyRef:
              name: winpower-exporter-secrets
              key: password
        volumeMounts:
        - name: config
          mountPath: /etc/winpower-exporter
        - name: data
          mountPath: /data
        args: ["-config", "/etc/winpower-exporter/config.yaml"]
      volumes:
      - name: config
        configMap:
          name: winpower-exporter-config
      - name: data
        persistentVolumeClaim:
          claimName: winpower-exporter-data
```

## Configuration Validation

### Built-in Validation

The exporter includes comprehensive configuration validation:

```bash
# Validate configuration with debug logging
./winpower-g2-exporter -config config.yaml -log-level debug

# Check specific configuration module
./winpower-g2-exporter -config config.yaml -log-level debug 2>&1 | grep -E "(storage|winpower|energy|server|scheduler) config"
```

### Common Validation Errors and Solutions

#### Required Fields Missing
```
Error: winpower config validation failed: URL is required
```
**Solution**: Ensure `WINPOWER_EXPORTER_WINPOWER_URL` or `winpower.url` is set.

#### Invalid Time Durations
```
Error: failed to convert env value for timeout: time: invalid duration "30"
```
**Solution**: Use proper duration format (e.g., "30s", "1m", "2h").

#### Invalid File Permissions
```
Error: invalid file mode '644', expected octal format (e.g., 0644)
```
**Solution**: Use octal format with leading zero for permissions.

## Security Best Practices

### 1. Authentication Security
- **Change Default Credentials**: Never use default usernames/passwords in production
- **Use Strong Passwords**: Generate complex passwords for WinPower authentication
- **Environment Variable Security**: Use secret management systems for sensitive data

### 2. Network Security
- **Use HTTPS**: Always use HTTPS URLs for WinPower connections
- **TLS Certificates**: Use valid certificates; only use `skip_tls_verify` for development
- **Network Segmentation**: Deploy in isolated network segments
- **Firewall Rules**: Restrict access to metrics endpoint (`/metrics`)

### 3. File System Security
```bash
# Secure configuration file permissions
chmod 600 config.yaml
chown app:app config.yaml

# Secure data directory permissions
chmod 750 /var/lib/winpower-exporter
chown app:app /var/lib/winpower-exporter
```

### 4. Container Security
```dockerfile
# Run as non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup
USER appuser
```

## Troubleshooting

### Configuration Loading Issues

#### Problem: Configuration file not found
```bash
Error: failed to read config file: Config File "config" Not Found in "[/etc/winpower-exporter]"
```
**Solution**: Check file path and permissions:
```bash
ls -la config.yaml
file config.yaml  # Verify it's valid YAML
```

#### Problem: Environment variables not working
```bash
# Debug environment variable loading
env | grep WINPOWER_EXPORTER
./winpower-g2-exporter -log-level debug 2>&1 | grep -i env
```

### Connection Issues

#### Problem: WinPower connection failed
```bash
# Test connectivity
curl -k -u "admin:password" https://winpower.example.com:8443/api/health

# Debug with verbose logging
./winpower-g2-exporter -log-level debug -winpower-skip-ssl-verify
```

#### Problem: SSL/TLS certificate errors
```
Error: x509: certificate signed by unknown authority
```
**Solution**: For development, use `-winpower-skip-ssl-verify` or install proper certificates.

### Performance Issues

#### Problem: High memory usage
- Check `max_calculation_time` in energy configuration
- Monitor data directory size
- Review collection interval settings

#### Problem: Slow metrics response
- Enable HTTP timeouts
- Check rate limiting settings
- Monitor system resources

## Advanced Configuration Scenarios

### 1. Multi-Environment Configuration

```bash
# Development
./winpower-g2-exporter \
  -config config.dev.yaml \
  -log-level debug \
  -server-mode debug

# Production
export WINPOWER_EXPORTER_LOG_LEVEL="info"
export WINPOWER_EXPORTER_SERVER_MODE="release"
./winpower-g2-exporter -config config.prod.yaml
```

### 2. Configuration Templates

Create templates for different deployment scenarios:
- `config.docker.yaml` - Container deployments
- `config.k8s.yaml` - Kubernetes deployments
- `config.prod.yaml` - Production environments
- `config.dev.yaml` - Development environments

### 3. Configuration Migration

When migrating from older versions:

```bash
# Backup existing configuration
cp config.yaml config.yaml.backup

# Validate new configuration format
./winpower-g2-exporter -config config.yaml -log-level debug

# Test with dry run (if supported)
./winpower-g2-exporter -config config.yaml -dry-run
```

## Monitoring and Observability

### Configuration Metrics

The exporter exposes configuration-related metrics:

```bash
# View configuration metrics
curl http://localhost:9090/metrics | grep winpower_exporter_config
```

### Logging Configuration

```bash
# JSON structured logging
export WINPOWER_EXPORTER_LOG_LEVEL="info"
export WINPOWER_EXPORTER_LOG_FORMAT="json"

# Console logging for development
export WINPOWER_EXPORTER_LOG_LEVEL="debug"
export WINPOWER_EXPORTER_LOG_FORMAT="console"
```

## Support and Help

For additional help:

1. **Check logs**: Use `-log-level debug` for detailed troubleshooting
2. **Validate configuration**: Test with minimal configuration first
3. **Review examples**: See `config.yaml.example` for complete options
4. **Check dependencies**: Ensure Go version and system requirements are met
5. **Community support**: Check GitHub issues and documentation

```bash
# Get help with all available options
./winpower-g2-exporter -help

# Check version information
./winpower-g2-exporter -version
```