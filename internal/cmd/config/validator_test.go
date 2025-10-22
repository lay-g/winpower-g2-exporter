package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestConfigValidator_Validate(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid config", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Host:        "0.0.0.0",
				Port:        9090,
				ReadTimeout: 30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
			},
			WinPower: WinPowerConfig{
				URL:           "http://example.com",
				Username:      "testuser",
				Password:      "testpass",
				Timeout:       30 * time.Second,
				RetryInterval: 5 * time.Second,
				MaxRetries:    3,
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   true,
			},
			Storage: StorageConfig{
				DataDir:   t.TempDir(),
				SyncWrite: true,
			},
			Collector: CollectorConfig{
				Enabled:       true,
				Interval:      5 * time.Second,
				Timeout:       3 * time.Second,
				MaxConcurrent: 5,
			},
			Metrics: MetricsConfig{
				Enabled:   true,
				Path:      "/metrics",
				Namespace: "winpower",
				Subsystem: "exporter",
				HelpText:  "Test metrics",
			},
			Energy: EnergyConfig{
				Enabled:       true,
				Interval:      5 * time.Second,
				Precision:     3,
				StoragePeriod: 1 * time.Hour,
			},
			Scheduler: SchedulerConfig{
				Enabled:  true,
				Interval: 5 * time.Second,
			},
			Auth: AuthConfig{
				Enabled: false,
			},
		}

		err := validator.Validate(config)
		assert.NoError(t, err, "Valid config should pass validation")
	})

	t.Run("nil config", func(t *testing.T) {
		err := validator.Validate(nil)
		assert.Error(t, err, "Nil config should fail validation")
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

func TestConfigValidator_ValidateServer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid server config", func(t *testing.T) {
		config := &ServerConfig{
			Host:        "0.0.0.0",
			Port:        9090,
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := validator.ValidateServer(config)
		assert.NoError(t, err, "Valid server config should pass validation")
	})

	t.Run("invalid host", func(t *testing.T) {
		config := &ServerConfig{
			Host:        "",
			Port:        9090,
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := validator.ValidateServer(config)
		assert.Error(t, err, "Empty host should fail validation")
		assert.Contains(t, err.Error(), "server host cannot be empty")
	})

	t.Run("invalid port", func(t *testing.T) {
		config := &ServerConfig{
			Host:        "0.0.0.0",
			Port:        0,
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := validator.ValidateServer(config)
		assert.Error(t, err, "Invalid port should fail validation")
		assert.Contains(t, err.Error(), "server port must be between 1 and 65535")
	})

	t.Run("negative timeout", func(t *testing.T) {
		config := &ServerConfig{
			Host:        "0.0.0.0",
			Port:        9090,
			ReadTimeout: -1 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := validator.ValidateServer(config)
		assert.Error(t, err, "Negative timeout should fail validation")
		assert.Contains(t, err.Error(), "server read timeout must be positive")
	})

	t.Run("localhost host", func(t *testing.T) {
		config := &ServerConfig{
			Host:        "localhost",
			Port:        9090,
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := validator.ValidateServer(config)
		assert.NoError(t, err, "Localhost host should pass validation")
	})

	t.Run("IP address host", func(t *testing.T) {
		config := &ServerConfig{
			Host:        "192.168.1.100",
			Port:        9090,
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := validator.ValidateServer(config)
		assert.NoError(t, err, "IP address host should pass validation")
	})
}

func TestConfigValidator_ValidateWinPower(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid WinPower config", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "http://example.com",
			Username:      "testuser",
			Password:      "testpass",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.NoError(t, err, "Valid WinPower config should pass validation")
	})

	t.Run("empty URL", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "",
			Username:      "testuser",
			Password:      "testpass",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.Error(t, err, "Empty URL should fail validation")
		assert.Contains(t, err.Error(), "winpower URL cannot be empty")
	})

	t.Run("invalid URL format", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "not-a-url",
			Username:      "testuser",
			Password:      "testpass",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.Error(t, err, "Invalid URL format should fail validation")
		assert.Contains(t, err.Error(), "invalid winpower URL format")
	})

	t.Run("empty username", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "http://example.com",
			Username:      "",
			Password:      "testpass",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.Error(t, err, "Empty username should fail validation")
		assert.Contains(t, err.Error(), "winpower username cannot be empty")
	})

	t.Run("empty password", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "http://example.com",
			Username:      "testuser",
			Password:      "",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.Error(t, err, "Empty password should fail validation")
		assert.Contains(t, err.Error(), "winpower password cannot be empty")
	})

	t.Run("negative timeout", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "http://example.com",
			Username:      "testuser",
			Password:      "testpass",
			Timeout:       -1 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.Error(t, err, "Negative timeout should fail validation")
		assert.Contains(t, err.Error(), "winpower timeout must be positive")
	})

	t.Run("too long username", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "http://example.com",
			Username:      string(make([]byte, 300)), // 300 characters
			Password:      "testpass",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.Error(t, err, "Too long username should fail validation")
		assert.Contains(t, err.Error(), "winpower username too long")
	})

	t.Run("HTTPS URL", func(t *testing.T) {
		config := &WinPowerConfig{
			URL:           "https://secure.example.com",
			Username:      "testuser",
			Password:      "testpass",
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
		}

		err := validator.ValidateWinPower(config)
		assert.NoError(t, err, "HTTPS URL should pass validation")
	})
}

func TestConfigValidator_ValidateLogging(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid logging config", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.NoError(t, err, "Valid logging config should pass validation")
	})

	t.Run("invalid log level", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "invalid",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.Error(t, err, "Invalid log level should fail validation")
		assert.Contains(t, err.Error(), "invalid log level")
	})

	t.Run("invalid log format", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "info",
			Format:     "invalid",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.Error(t, err, "Invalid log format should fail validation")
		assert.Contains(t, err.Error(), "invalid log format")
	})

	t.Run("invalid log output", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "invalid",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.Error(t, err, "Invalid log output should fail validation")
		assert.Contains(t, err.Error(), "invalid log output")
	})

	t.Run("file output without filename", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "file",
			Filename:   "",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.Error(t, err, "File output without filename should fail validation")
		assert.Contains(t, err.Error(), "log filename cannot be empty")
	})

	t.Run("file output with valid filename", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "file",
			Filename:   tempDir + "/test.log",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.NoError(t, err, "File output with valid filename should pass validation")
	})

	t.Run("negative max size", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    -1,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.Error(t, err, "Negative max size should fail validation")
		assert.Contains(t, err.Error(), "log max size must be positive")
	})

	t.Run("case insensitive log level", func(t *testing.T) {
		config := &LoggingConfig{
			Level:      "DEBUG",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		err := validator.ValidateLogging(config)
		assert.Error(t, err, "Uppercase log level should fail validation")
		assert.Contains(t, err.Error(), "invalid log level")
	})
}

func TestConfigValidator_ValidateStorage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid storage config", func(t *testing.T) {
		config := &StorageConfig{
			DataDir:   t.TempDir(),
			SyncWrite: true,
		}

		err := validator.ValidateStorage(config)
		assert.NoError(t, err, "Valid storage config should pass validation")
	})

	t.Run("empty data dir", func(t *testing.T) {
		config := &StorageConfig{
			DataDir:   "",
			SyncWrite: true,
		}

		err := validator.ValidateStorage(config)
		assert.Error(t, err, "Empty data dir should fail validation")
		assert.Contains(t, err.Error(), "storage data directory cannot be empty")
	})

	t.Run("non-existent data dir", func(t *testing.T) {
		config := &StorageConfig{
			DataDir:   "/non/existent/path",
			SyncWrite: true,
		}

		err := validator.ValidateStorage(config)
		assert.Error(t, err, "Non-existent data dir should fail validation")
		assert.Contains(t, err.Error(), "cannot create data directory")
	})

	t.Run("relative data dir", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &StorageConfig{
			DataDir:   tempDir + "/relative",
			SyncWrite: true,
		}

		err := validator.ValidateStorage(config)
		assert.NoError(t, err, "Relative data dir should be converted to absolute path")
		assert.True(t, config.DataDir[0] == '/', "Data dir should be converted to absolute path")
	})

	t.Run("unwritable data dir", func(t *testing.T) {
		// 创建一个只读目录
		tempDir := t.TempDir()
		readOnlyDir := tempDir + "/readonly"
		err := os.Mkdir(readOnlyDir, 0444)
		require.NoError(t, err)
		defer os.Chmod(readOnlyDir, 0755) // 清理时恢复权限

		config := &StorageConfig{
			DataDir:   readOnlyDir + "/subdir",
			SyncWrite: true,
		}

		err = validator.ValidateStorage(config)
		assert.Error(t, err, "Unwritable data dir should fail validation")
		assert.Contains(t, err.Error(), "cannot create data directory")
	})
}

func TestConfigValidator_ValidateCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid collector config", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Timeout:       3 * time.Second,
			MaxConcurrent: 5,
		}

		err := validator.ValidateCollector(config)
		assert.NoError(t, err, "Valid collector config should pass validation")
	})

	t.Run("negative interval", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      -1 * time.Second,
			Timeout:       3 * time.Second,
			MaxConcurrent: 5,
		}

		err := validator.ValidateCollector(config)
		assert.Error(t, err, "Negative interval should fail validation")
		assert.Contains(t, err.Error(), "collector interval must be positive")
	})

	t.Run("negative timeout", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Timeout:       -1 * time.Second,
			MaxConcurrent: 5,
		}

		err := validator.ValidateCollector(config)
		assert.Error(t, err, "Negative timeout should fail validation")
		assert.Contains(t, err.Error(), "collector timeout must be positive")
	})

	t.Run("timeout greater than interval", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      3 * time.Second,
			Timeout:       5 * time.Second,
			MaxConcurrent: 5,
		}

		err := validator.ValidateCollector(config)
		assert.Error(t, err, "Timeout greater than interval should fail validation")
		assert.Contains(t, err.Error(), "collector timeout")
		assert.Contains(t, err.Error(), "must be less than interval")
	})

	t.Run("negative max concurrent", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Timeout:       3 * time.Second,
			MaxConcurrent: -1,
		}

		err := validator.ValidateCollector(config)
		assert.Error(t, err, "Negative max concurrent should fail validation")
		assert.Contains(t, err.Error(), "collector max concurrent must be positive")
	})

	t.Run("zero max concurrent", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Timeout:       3 * time.Second,
			MaxConcurrent: 0,
		}

		err := validator.ValidateCollector(config)
		assert.Error(t, err, "Zero max concurrent should fail validation")
		assert.Contains(t, err.Error(), "collector max concurrent must be positive")
	})

	t.Run("very short interval", func(t *testing.T) {
		config := &CollectorConfig{
			Enabled:       true,
			Interval:      100 * time.Millisecond,
			Timeout:       50 * time.Millisecond,
			MaxConcurrent: 5,
		}

		err := validator.ValidateCollector(config)
		assert.NoError(t, err, "Very short interval should pass validation but may log warning")
	})
}

func TestConfigValidator_ValidateMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid metrics config", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "winpower",
			Subsystem: "exporter",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.NoError(t, err, "Valid metrics config should pass validation")
	})

	t.Run("empty path", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "",
			Namespace: "winpower",
			Subsystem: "exporter",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.Error(t, err, "Empty path should fail validation")
		assert.Contains(t, err.Error(), "metrics path cannot be empty")
	})

	t.Run("path without leading slash", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "metrics",
			Namespace: "winpower",
			Subsystem: "exporter",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.Error(t, err, "Path without leading slash should fail validation")
		assert.Contains(t, err.Error(), "must start with '/'")
	})

	t.Run("empty namespace", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "",
			Subsystem: "exporter",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.Error(t, err, "Empty namespace should fail validation")
		assert.Contains(t, err.Error(), "metrics namespace cannot be empty")
	})

	t.Run("invalid namespace format", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "invalid-namespace",
			Subsystem: "exporter",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.Error(t, err, "Invalid namespace format should fail validation")
		assert.Contains(t, err.Error(), "invalid metrics namespace format")
	})

	t.Run("empty help text", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "winpower",
			Subsystem: "exporter",
			HelpText:  "",
		}

		err := validator.ValidateMetrics(config)
		assert.Error(t, err, "Empty help text should fail validation")
		assert.Contains(t, err.Error(), "metrics help text cannot be empty")
	})

	t.Run("valid subsystem", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "winpower",
			Subsystem: "valid_subsystem",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.NoError(t, err, "Valid subsystem should pass validation")
	})

	t.Run("invalid subsystem format", func(t *testing.T) {
		config := &MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "winpower",
			Subsystem: "invalid-subsystem",
			HelpText:  "Test metrics",
		}

		err := validator.ValidateMetrics(config)
		assert.Error(t, err, "Invalid subsystem format should fail validation")
		assert.Contains(t, err.Error(), "invalid metrics subsystem format")
	})
}

func TestConfigValidator_ValidateEnergy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid energy config", func(t *testing.T) {
		config := &EnergyConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Precision:     3,
			StoragePeriod: 1 * time.Hour,
		}

		err := validator.ValidateEnergy(config)
		assert.NoError(t, err, "Valid energy config should pass validation")
	})

	t.Run("negative interval", func(t *testing.T) {
		config := &EnergyConfig{
			Enabled:       true,
			Interval:      -1 * time.Second,
			Precision:     3,
			StoragePeriod: 1 * time.Hour,
		}

		err := validator.ValidateEnergy(config)
		assert.Error(t, err, "Negative interval should fail validation")
		assert.Contains(t, err.Error(), "energy interval must be positive")
	})

	t.Run("negative precision", func(t *testing.T) {
		config := &EnergyConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Precision:     -1,
			StoragePeriod: 1 * time.Hour,
		}

		err := validator.ValidateEnergy(config)
		assert.Error(t, err, "Negative precision should fail validation")
		assert.Contains(t, err.Error(), "energy precision must be between 0 and 10")
	})

	t.Run("precision too high", func(t *testing.T) {
		config := &EnergyConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Precision:     15,
			StoragePeriod: 1 * time.Hour,
		}

		err := validator.ValidateEnergy(config)
		assert.Error(t, err, "Precision too high should fail validation")
		assert.Contains(t, err.Error(), "energy precision must be between 0 and 10")
	})

	t.Run("negative storage period", func(t *testing.T) {
		config := &EnergyConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Precision:     3,
			StoragePeriod: -1 * time.Hour,
		}

		err := validator.ValidateEnergy(config)
		assert.Error(t, err, "Negative storage period should fail validation")
		assert.Contains(t, err.Error(), "energy storage period must be positive")
	})

	t.Run("storage period shorter than interval", func(t *testing.T) {
		config := &EnergyConfig{
			Enabled:       true,
			Interval:      5 * time.Minute,
			Precision:     3,
			StoragePeriod: 1 * time.Minute,
		}

		err := validator.ValidateEnergy(config)
		assert.NoError(t, err, "Storage period shorter than interval should pass validation but may log warning")
	})
}

func TestConfigValidator_ValidateScheduler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("valid scheduler config", func(t *testing.T) {
		config := &SchedulerConfig{
			Enabled:  true,
			Interval: 5 * time.Second,
		}

		err := validator.ValidateScheduler(config)
		assert.NoError(t, err, "Valid scheduler config should pass validation")
	})

	t.Run("negative interval", func(t *testing.T) {
		config := &SchedulerConfig{
			Enabled:  true,
			Interval: -1 * time.Second,
		}

		err := validator.ValidateScheduler(config)
		assert.Error(t, err, "Negative interval should fail validation")
		assert.Contains(t, err.Error(), "scheduler interval must be positive")
	})

	t.Run("very short interval", func(t *testing.T) {
		config := &SchedulerConfig{
			Enabled:  true,
			Interval: 100 * time.Millisecond,
		}

		err := validator.ValidateScheduler(config)
		assert.NoError(t, err, "Very short interval should pass validation but may log warning")
	})
}

func TestConfigValidator_ValidateAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("disabled auth config", func(t *testing.T) {
		config := &AuthConfig{
			Enabled: false,
		}

		err := validator.ValidateAuth(config)
		assert.NoError(t, err, "Disabled auth config should pass validation")
	})

	t.Run("valid token auth config", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "token",
			TokenURL: "http://example.com/token",
			Username: "testuser",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.NoError(t, err, "Valid token auth config should pass validation")
	})

	t.Run("valid basic auth config", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "basic",
			Username: "testuser",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.NoError(t, err, "Valid basic auth config should pass validation")
	})

	t.Run("valid oauth2 auth config", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "oauth2",
			TokenURL: "https://example.com/oauth2/token",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.NoError(t, err, "Valid oauth2 auth config should pass validation")
	})

	t.Run("invalid auth method", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "invalid",
			Username: "testuser",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Invalid auth method should fail validation")
		assert.Contains(t, err.Error(), "invalid auth method")
	})

	t.Run("token auth without token URL", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "token",
			Username: "testuser",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Token auth without token URL should fail validation")
		assert.Contains(t, err.Error(), "token URL cannot be empty")
	})

	t.Run("token auth without username", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "token",
			TokenURL: "http://example.com/token",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Token auth without username should fail validation")
		assert.Contains(t, err.Error(), "username cannot be empty")
	})

	t.Run("token auth without password", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "token",
			TokenURL: "http://example.com/token",
			Username: "testuser",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Token auth without password should fail validation")
		assert.Contains(t, err.Error(), "password cannot be empty")
	})

	t.Run("basic auth without username", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "basic",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Basic auth without username should fail validation")
		assert.Contains(t, err.Error(), "username cannot be empty")
	})

	t.Run("negative timeout", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "basic",
			Username: "testuser",
			Password: "testpass",
			Timeout:  -1 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Negative timeout should fail validation")
		assert.Contains(t, err.Error(), "auth timeout must be positive")
	})

	t.Run("negative cache TTL", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "basic",
			Username: "testuser",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: -1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Negative cache TTL should fail validation")
		assert.Contains(t, err.Error(), "auth cache TTL must be positive")
	})

	t.Run("invalid token URL format", func(t *testing.T) {
		config := &AuthConfig{
			Enabled:  true,
			Method:   "token",
			TokenURL: "not-a-url",
			Username: "testuser",
			Password: "testpass",
			Timeout:  30 * time.Second,
			CacheTTL: 1 * time.Hour,
		}

		err := validator.ValidateAuth(config)
		assert.Error(t, err, "Invalid token URL format should fail validation")
		assert.Contains(t, err.Error(), "invalid token URL format")
	})
}

func TestConfigValidator_contains(t *testing.T) {
	logger := zaptest.NewLogger(t)
	validator := NewValidator(logger)

	t.Run("item exists", func(t *testing.T) {
		slice := []string{"apple", "banana", "orange"}
		result := validator.contains(slice, "banana")
		assert.True(t, result, "Should return true for existing item")
	})

	t.Run("item does not exist", func(t *testing.T) {
		slice := []string{"apple", "banana", "orange"}
		result := validator.contains(slice, "grape")
		assert.False(t, result, "Should return false for non-existing item")
	})

	t.Run("empty slice", func(t *testing.T) {
		slice := []string{}
		result := validator.contains(slice, "apple")
		assert.False(t, result, "Should return false for empty slice")
	})

	t.Run("case sensitive", func(t *testing.T) {
		slice := []string{"Apple", "Banana"}
		result := validator.contains(slice, "apple")
		assert.False(t, result, "Should be case sensitive")
	})
}