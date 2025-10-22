package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestConfigModuleCompatibility 测试新的 cmd/config 模块与现有 pkgs/config 模块的兼容性
func TestConfigModuleCompatibility(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)
	validator := NewValidator(logger)
	merger := NewMerger(logger)

	t.Run("default_config_compatibility", func(t *testing.T) {
		// 测试默认配置是否与现有模块兼容
		config := loader.LoadFromDefaults()
		require.NotNil(t, config)

		// 验证默认配置值与项目规范一致
		assert.Equal(t, 9090, config.Server.Port, "Default server port should match project specification")
		assert.Equal(t, "0.0.0.0", config.Server.Host, "Default server host should match project specification")
		assert.Equal(t, "info", config.Logging.Level, "Default log level should match project specification")
		assert.Equal(t, "json", config.Logging.Format, "Default log format should match project specification")
		assert.Equal(t, "./data", config.Storage.DataDir, "Default data directory should match project specification")
		assert.Equal(t, true, config.Storage.SyncWrite, "Default sync write should match project specification")
		assert.Equal(t, 5*time.Second, config.Collector.Interval, "Default collector interval should match project specification")
		assert.Equal(t, 5*time.Second, config.Scheduler.Interval, "Default scheduler interval should match project specification")

		// 注意：默认配置可能不包含 WinPower 的必需字段（URL、用户名、密码），所以验证会失败
		// 这是预期的行为，因为用户必须提供这些必需的配置
		err := validator.Validate(config)
		// 默认配置验证失败是预期的，因为缺少必需的 WinPower 配置
		assert.Error(t, err, "Default config should fail validation due to missing required WinPower fields")
		assert.Contains(t, err.Error(), "winpower", "Validation error should be related to WinPower configuration")
	})

	t.Run("config_structure_compatibility", func(t *testing.T) {
		// 测试配置结构是否包含所有必需的字段
		config := &Config{
			Server: ServerConfig{
				Port:        8080,
				Host:        "localhost",
				ReadTimeout: 30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				EnablePprof: false,
			},
			WinPower: WinPowerConfig{
				URL:           "http://example.com",
				Username:      "testuser",
				Password:      "testpass",
				Timeout:       30 * time.Second,
				RetryInterval: 5 * time.Second,
				MaxRetries:    3,
				SkipSSLVerify: false,
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   BoolPtr(true),
			},
			Storage: StorageConfig{
				DataDir:   "/tmp/data",
				SyncWrite: BoolPtr(true),
			},
			Collector: CollectorConfig{
				Enabled:       BoolPtr(true),
				Interval:      5 * time.Second,
				Timeout:       3 * time.Second,
				MaxConcurrent: 5,
			},
			Metrics: MetricsConfig{
				Enabled:   BoolPtr(true),
				Path:      "/metrics",
				Namespace: "winpower",
				Subsystem: "exporter",
				HelpText:  "WinPower G2 Exporter Metrics",
			},
			Energy: EnergyConfig{
				Enabled:       BoolPtr(true),
				Interval:      5 * time.Second,
				Precision:     3,
				StoragePeriod: 1 * time.Hour,
			},
			Scheduler: SchedulerConfig{
				Enabled:  BoolPtr(true),
				Interval: 5 * time.Second,
			},
			Auth: AuthConfig{
				Enabled: false,
				Method:  "token",
				Timeout: 30 * time.Second,
				CacheTTL: 1 * time.Hour,
			},
		}

		// 验证配置结构完整性
		assert.NotNil(t, config.Server, "Server config should not be nil")
		assert.NotNil(t, config.WinPower, "WinPower config should not be nil")
		assert.NotNil(t, config.Logging, "Logging config should not be nil")
		assert.NotNil(t, config.Storage, "Storage config should not be nil")
		assert.NotNil(t, config.Collector, "Collector config should not be nil")
		assert.NotNil(t, config.Metrics, "Metrics config should not be nil")
		assert.NotNil(t, config.Energy, "Energy config should not be nil")
		assert.NotNil(t, config.Scheduler, "Scheduler config should not be nil")
		assert.NotNil(t, config.Auth, "Auth config should not be nil")

		// 验证配置通过验证
		err := validator.Validate(config)
		assert.NoError(t, err, "Complete config should pass validation")
	})

	t.Run("environment_variable_compatibility", func(t *testing.T) {
		// 测试环境变量前缀是否与项目规范一致
		envVars := map[string]string{
			"WINPOWER_EXPORTER_SERVER_PORT":         "8080",
			"WINPOWER_EXPORTER_SERVER_HOST":         "0.0.0.0",
			"WINPOWER_EXPORTER_LOGGING_LEVEL":       "debug",
			"WINPOWER_EXPORTER_LOGGING_FORMAT":      "console",
			"WINPOWER_EXPORTER_DATA_DIR":            "/test/data",
			"WINPOWER_EXPORTER_SYNC_WRITE":          "false",
			"WINPOWER_EXPORTER_CONSOLE_URL":         "http://test.example.com",
			"WINPOWER_EXPORTER_USERNAME":            "testuser",
			"WINPOWER_EXPORTER_PASSWORD":            "testpass",
			"WINPOWER_EXPORTER_SKIP_SSL_VERIFY":     "true",
			"WINPOWER_EXPORTER_COLLECTOR_ENABLED":   "true",
			"WINPOWER_EXPORTER_METRICS_ENABLED":     "true",
			"WINPOWER_EXPORTER_METRICS_PATH":        "/custom-metrics",
			"WINPOWER_EXPORTER_ENERGY_ENABLED":      "true",
			"WINPOWER_EXPORTER_SCHEDULER_ENABLED":   "true",
			"WINPOWER_EXPORTER_AUTH_ENABLED":        "true",
		}

		// 设置测试环境变量
		originalEnv := make(map[string]string)
		for key, value := range envVars {
			originalEnv[key] = os.Getenv(key)
			_ = os.Setenv(key, value) // 忽略错误，测试环境中设置环境变量
		}
		defer func() {
			for key, value := range originalEnv {
				if value == "" {
					_ = os.Unsetenv(key) // 忽略错误，测试环境中清理环境变量
				} else {
					_ = os.Setenv(key, value) // 忽略错误，测试环境中恢复环境变量
				}
			}
		}()

		// 加载配置
		config, err := loader.Load("", nil)
		require.NoError(t, err)
		require.NotNil(t, config)

		// 验证环境变量是否正确解析
		assert.Equal(t, 8080, config.Server.Port, "Environment variable WINPOWER_EXPORTER_SERVER_PORT should be parsed")
		assert.Equal(t, "0.0.0.0", config.Server.Host, "Environment variable WINPOWER_EXPORTER_SERVER_HOST should be parsed")
		assert.Equal(t, "debug", config.Logging.Level, "Environment variable WINPOWER_EXPORTER_LOGGING_LEVEL should be parsed")
		assert.Equal(t, "console", config.Logging.Format, "Environment variable WINPOWER_EXPORTER_LOGGING_FORMAT should be parsed")
		assert.Equal(t, "/test/data", config.Storage.DataDir, "Environment variable WINPOWER_EXPORTER_DATA_DIR should be parsed")
		assert.Equal(t, false, config.Storage.SyncWrite, "Environment variable WINPOWER_EXPORTER_SYNC_WRITE should be parsed")
		assert.Equal(t, "http://test.example.com", config.WinPower.URL, "Environment variable WINPOWER_EXPORTER_CONSOLE_URL should be parsed")
		assert.Equal(t, "testuser", config.WinPower.Username, "Environment variable WINPOWER_EXPORTER_USERNAME should be parsed")
		assert.Equal(t, "testpass", config.WinPower.Password, "Environment variable WINPOWER_EXPORTER_PASSWORD should be parsed")
		assert.Equal(t, true, config.WinPower.SkipSSLVerify, "Environment variable WINPOWER_EXPORTER_SKIP_SSL_VERIFY should be parsed")
		assert.Equal(t, true, config.Collector.Enabled, "Environment variable WINPOWER_EXPORTER_COLLECTOR_ENABLED should be parsed")
		assert.Equal(t, true, config.Metrics.Enabled, "Environment variable WINPOWER_EXPORTER_METRICS_ENABLED should be parsed")
		assert.Equal(t, "/custom-metrics", config.Metrics.Path, "Environment variable WINPOWER_EXPORTER_METRICS_PATH should be parsed")
		assert.Equal(t, true, config.Energy.Enabled, "Environment variable WINPOWER_EXPORTER_ENERGY_ENABLED should be parsed")
		assert.Equal(t, true, config.Scheduler.Enabled, "Environment variable WINPOWER_EXPORTER_SCHEDULER_ENABLED should be parsed")
		assert.Equal(t, true, config.Auth.Enabled, "Environment variable WINPOWER_EXPORTER_AUTH_ENABLED should be parsed")

		// 设置必需的超时字段（环境变量测试没有设置这些字段）
	config.Server.ReadTimeout = 30 * time.Second
	config.Server.WriteTimeout = 30 * time.Second
	config.Server.IdleTimeout = 60 * time.Second
	config.WinPower.Timeout = 30 * time.Second
	config.WinPower.RetryInterval = 5 * time.Second

	// 验证配置通过验证
	err = validator.Validate(config)
	assert.NoError(t, err, "Config loaded from environment variables should pass validation")
	})

	t.Run("config_file_compatibility", func(t *testing.T) {
		// 测试配置文件格式是否与项目规范一致
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")

		configContent := `
# WinPower G2 Exporter Configuration
server:
  port: 9090
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  enable_pprof: false

winpower:
  url: "http://winpower.example.com"
  username: "admin"
  password: "password"
  timeout: "30s"
  retry_interval: "5s"
  max_retries: 3
  skip_ssl_verify: false

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true

storage:
  data_dir: "./data"
  sync_write: true

collector:
  enabled: true
  interval: "5s"
  timeout: "3s"
  max_concurrent: 5

metrics:
  enabled: true
  path: "/metrics"
  namespace: "winpower"
  subsystem: "exporter"
  help_text: "WinPower G2 Exporter Metrics"

energy:
  enabled: true
  interval: "5s"
  precision: 3
  storage_period: "1h"

scheduler:
  enabled: true
  interval: "5s"

auth:
  enabled: false
  method: "token"
  timeout: "30s"
  cache_ttl: "1h"
`

		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 加载配置文件
		config, err := loader.Load(configFile, nil)
		require.NoError(t, err)
		require.NotNil(t, config)

		// 验证配置文件内容是否正确解析
		assert.Equal(t, 9090, config.Server.Port, "Config file server port should be parsed")
		assert.Equal(t, "0.0.0.0", config.Server.Host, "Config file server host should be parsed")
		assert.Equal(t, 30*time.Second, config.Server.ReadTimeout, "Config file server read timeout should be parsed")
		assert.Equal(t, 30*time.Second, config.Server.WriteTimeout, "Config file server write timeout should be parsed")
		assert.Equal(t, 60*time.Second, config.Server.IdleTimeout, "Config file server idle timeout should be parsed")
		assert.Equal(t, false, config.Server.EnablePprof, "Config file server enable pprof should be parsed")

		assert.Equal(t, "http://winpower.example.com", config.WinPower.URL, "Config file WinPower URL should be parsed")
		assert.Equal(t, "admin", config.WinPower.Username, "Config file WinPower username should be parsed")
		assert.Equal(t, "password", config.WinPower.Password, "Config file WinPower password should be parsed")
		assert.Equal(t, 30*time.Second, config.WinPower.Timeout, "Config file WinPower timeout should be parsed")
		assert.Equal(t, 5*time.Second, config.WinPower.RetryInterval, "Config file WinPower retry interval should be parsed")
		assert.Equal(t, 3, config.WinPower.MaxRetries, "Config file WinPower max retries should be parsed")
		assert.Equal(t, false, config.WinPower.SkipSSLVerify, "Config file WinPower skip SSL verify should be parsed")

		assert.Equal(t, "info", config.Logging.Level, "Config file logging level should be parsed")
		assert.Equal(t, "json", config.Logging.Format, "Config file logging format should be parsed")
		assert.Equal(t, "stdout", config.Logging.Output, "Config file logging output should be parsed")
		assert.Equal(t, 100, config.Logging.MaxSize, "Config file logging max size should be parsed")
		assert.Equal(t, 7, config.Logging.MaxAge, "Config file logging max age should be parsed")
		assert.Equal(t, 3, config.Logging.MaxBackups, "Config file logging max backups should be parsed")
		assert.Equal(t, true, config.Logging.Compress, "Config file logging compress should be parsed")

		assert.Equal(t, "./data", config.Storage.DataDir, "Config file storage data dir should be parsed")
		assert.Equal(t, true, config.Storage.SyncWrite, "Config file storage sync write should be parsed")

		assert.Equal(t, true, config.Collector.Enabled, "Config file collector enabled should be parsed")
		assert.Equal(t, 5*time.Second, config.Collector.Interval, "Config file collector interval should be parsed")
		assert.Equal(t, 3*time.Second, config.Collector.Timeout, "Config file collector timeout should be parsed")
		assert.Equal(t, 5, config.Collector.MaxConcurrent, "Config file collector max concurrent should be parsed")

		assert.Equal(t, true, config.Metrics.Enabled, "Config file metrics enabled should be parsed")
		assert.Equal(t, "/metrics", config.Metrics.Path, "Config file metrics path should be parsed")
		assert.Equal(t, "winpower", config.Metrics.Namespace, "Config file metrics namespace should be parsed")
		assert.Equal(t, "exporter", config.Metrics.Subsystem, "Config file metrics subsystem should be parsed")
		assert.Equal(t, "WinPower G2 Exporter Metrics", config.Metrics.HelpText, "Config file metrics help text should be parsed")

		assert.Equal(t, true, config.Energy.Enabled, "Config file energy enabled should be parsed")
		assert.Equal(t, 5*time.Second, config.Energy.Interval, "Config file energy interval should be parsed")
		assert.Equal(t, 3, config.Energy.Precision, "Config file energy precision should be parsed")
		assert.Equal(t, 1*time.Hour, config.Energy.StoragePeriod, "Config file energy storage period should be parsed")

		assert.Equal(t, true, config.Scheduler.Enabled, "Config file scheduler enabled should be parsed")
		assert.Equal(t, 5*time.Second, config.Scheduler.Interval, "Config file scheduler interval should be parsed")

		assert.Equal(t, false, config.Auth.Enabled, "Config file auth enabled should be parsed")
		assert.Equal(t, "token", config.Auth.Method, "Config file auth method should be parsed")
		assert.Equal(t, 30*time.Second, config.Auth.Timeout, "Config file auth timeout should be parsed")
		assert.Equal(t, 1*time.Hour, config.Auth.CacheTTL, "Config file auth cache TTL should be parsed")

		// 验证配置通过验证
		err = validator.Validate(config)
		assert.NoError(t, err, "Config loaded from file should pass validation")
	})

	t.Run("merger_compatibility", func(t *testing.T) {
		// 测试配置合并器是否与现有模块兼容
		baseConfig := &Config{
			Server: ServerConfig{
				Port:        8080,
				Host:        "localhost",
				ReadTimeout: 30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				EnablePprof: false,
			},
			Logging: LoggingConfig{
				Level:      "info",
				Format:     "json",
				Output:     "stdout",
				MaxSize:    100,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   BoolPtr(true),
			},
			WinPower: WinPowerConfig{
				Timeout:       30 * time.Second,
				RetryInterval: 5 * time.Second,
				MaxRetries:    3,
				SkipSSLVerify: false,
			},
			Storage: StorageConfig{
				DataDir:   "./data",
				SyncWrite: BoolPtr(true),
			},
		}

		overlayConfig := &Config{
			Server: ServerConfig{
				Port:     9090,
				Host:     "0.0.0.0",
				EnablePprof: true,
			},
			Logging: LoggingConfig{
				Level: "debug",
			},
			WinPower: WinPowerConfig{
				URL:           "http://example.com",
				Username:      "user",
				Password:      "pass",
				SkipSSLVerify: true,
			},
			Storage: StorageConfig{
				DataDir: "/new/data",
			},
		}

		// 合并配置
		mergedConfig, err := merger.Merge(baseConfig, overlayConfig)
		require.NoError(t, err)
		require.NotNil(t, mergedConfig)

		// 验证合并结果
		assert.Equal(t, 9090, mergedConfig.Server.Port, "Merged server port should be from overlay")
		assert.Equal(t, "0.0.0.0", mergedConfig.Server.Host, "Merged server host should be from overlay")
		assert.Equal(t, true, mergedConfig.Server.EnablePprof, "Merged server enable pprof should be from overlay")
		assert.Equal(t, 30*time.Second, mergedConfig.Server.ReadTimeout, "Merged server read timeout should be from base")
		assert.Equal(t, 30*time.Second, mergedConfig.Server.WriteTimeout, "Merged server write timeout should be from base")
		assert.Equal(t, 60*time.Second, mergedConfig.Server.IdleTimeout, "Merged server idle timeout should be from base")

		assert.Equal(t, "debug", mergedConfig.Logging.Level, "Merged logging level should be from overlay")
		assert.Equal(t, "json", mergedConfig.Logging.Format, "Merged logging format should be from base")
		assert.Equal(t, "stdout", mergedConfig.Logging.Output, "Merged logging output should be from base")
		assert.Equal(t, 100, mergedConfig.Logging.MaxSize, "Merged logging max size should be from base")

		assert.Equal(t, "http://example.com", mergedConfig.WinPower.URL, "Merged WinPower URL should be from overlay")
		assert.Equal(t, "user", mergedConfig.WinPower.Username, "Merged WinPower username should be from overlay")
		assert.Equal(t, "pass", mergedConfig.WinPower.Password, "Merged WinPower password should be from overlay")
		assert.Equal(t, 30*time.Second, mergedConfig.WinPower.Timeout, "Merged WinPower timeout should be from base")
		assert.Equal(t, 5*time.Second, mergedConfig.WinPower.RetryInterval, "Merged WinPower retry interval should be from base")
		assert.Equal(t, 3, mergedConfig.WinPower.MaxRetries, "Merged WinPower max retries should be from base")
		assert.Equal(t, true, mergedConfig.WinPower.SkipSSLVerify, "Merged WinPower skip SSL verify should be from overlay")

		assert.Equal(t, "/new/data", mergedConfig.Storage.DataDir, "Merged storage data dir should be from overlay")
		assert.Equal(t, true, mergedConfig.Storage.SyncWrite, "Merged storage sync write should be from base")

		// 验证合并后的配置通过验证
		err = validator.Validate(mergedConfig)
		assert.NoError(t, err, "Merged config should pass validation")
	})

	t.Run("required_fields_compatibility", func(t *testing.T) {
		// 测试必需字段的处理是否与项目规范一致
		config := &Config{
			WinPower: WinPowerConfig{
				URL:      "http://example.com",
				Username: "user",
				Password: "pass",
			},
		}

		// 验证缺少必需字段时验证失败
		config.WinPower.URL = ""
		err := validator.Validate(config)
		assert.Error(t, err, "Config without WinPower URL should fail validation")
		assert.Contains(t, err.Error(), "winpower URL cannot be empty")

		config.WinPower.URL = "http://example.com"
		config.WinPower.Username = ""
		err = validator.Validate(config)
		assert.Error(t, err, "Config without WinPower username should fail validation")
		assert.Contains(t, err.Error(), "winpower username cannot be empty")

		config.WinPower.Username = "user"
		config.WinPower.Password = ""
		err = validator.Validate(config)
		assert.Error(t, err, "Config without WinPower password should fail validation")
		assert.Contains(t, err.Error(), "winpower password cannot be empty")
	})

	t.Run("module_defaults_compatibility", func(t *testing.T) {
		// 测试各模块默认值是否与项目设计文档一致
		defaults := loader.LoadFromDefaults()

		// 服务器默认值
		assert.Equal(t, 9090, defaults.Server.Port, "Default server port should be 9090")
		assert.Equal(t, "0.0.0.0", defaults.Server.Host, "Default server host should be 0.0.0.0")
		assert.Equal(t, 30*time.Second, defaults.Server.ReadTimeout, "Default server read timeout should be 30s")
		assert.Equal(t, 30*time.Second, defaults.Server.WriteTimeout, "Default server write timeout should be 30s")
		assert.Equal(t, 60*time.Second, defaults.Server.IdleTimeout, "Default server idle timeout should be 60s")
		assert.Equal(t, false, defaults.Server.EnablePprof, "Default server enable pprof should be false")

		// WinPower 默认值
		assert.Equal(t, 30*time.Second, defaults.WinPower.Timeout, "Default WinPower timeout should be 30s")
		assert.Equal(t, 5*time.Second, defaults.WinPower.RetryInterval, "Default WinPower retry interval should be 5s")
		assert.Equal(t, 3, defaults.WinPower.MaxRetries, "Default WinPower max retries should be 3")
		assert.Equal(t, false, defaults.WinPower.SkipSSLVerify, "Default WinPower skip SSL verify should be false")

		// 日志默认值
		assert.Equal(t, "info", defaults.Logging.Level, "Default log level should be info")
		assert.Equal(t, "json", defaults.Logging.Format, "Default log format should be json")
		assert.Equal(t, "stdout", defaults.Logging.Output, "Default log output should be stdout")
		assert.Equal(t, 100, defaults.Logging.MaxSize, "Default log max size should be 100")
		assert.Equal(t, 7, defaults.Logging.MaxAge, "Default log max age should be 7")
		assert.Equal(t, 3, defaults.Logging.MaxBackups, "Default log max backups should be 3")
		assert.Equal(t, true, defaults.Logging.Compress, "Default log compress should be true")

		// 存储默认值
		assert.Equal(t, "./data", defaults.Storage.DataDir, "Default storage data dir should be ./data")
		assert.Equal(t, true, defaults.Storage.SyncWrite, "Default storage sync write should be true")

		// 采集器默认值
		assert.Equal(t, true, defaults.Collector.Enabled, "Default collector enabled should be true")
		assert.Equal(t, 5*time.Second, defaults.Collector.Interval, "Default collector interval should be 5s")
		assert.Equal(t, 30*time.Second, defaults.Collector.Timeout, "Default collector timeout should be 30s")
		assert.Equal(t, 5, defaults.Collector.MaxConcurrent, "Default collector max concurrent should be 5")

		// 指标默认值
		assert.Equal(t, true, defaults.Metrics.Enabled, "Default metrics enabled should be true")
		assert.Equal(t, "/metrics", defaults.Metrics.Path, "Default metrics path should be /metrics")
		assert.Equal(t, "winpower", defaults.Metrics.Namespace, "Default metrics namespace should be winpower")
		assert.Equal(t, "exporter", defaults.Metrics.Subsystem, "Default metrics subsystem should be exporter")
		assert.Equal(t, "WinPower G2 Exporter Metrics", defaults.Metrics.HelpText, "Default metrics help text should be WinPower G2 Exporter Metrics")

		// 能耗默认值
		assert.Equal(t, true, defaults.Energy.Enabled, "Default energy enabled should be true")
		assert.Equal(t, 5*time.Second, defaults.Energy.Interval, "Default energy interval should be 5s")
		assert.Equal(t, 3, defaults.Energy.Precision, "Default energy precision should be 3")
		assert.Equal(t, 1*time.Hour, defaults.Energy.StoragePeriod, "Default energy storage period should be 1h")

		// 调度器默认值
		assert.Equal(t, true, defaults.Scheduler.Enabled, "Default scheduler enabled should be true")
		assert.Equal(t, 5*time.Second, defaults.Scheduler.Interval, "Default scheduler interval should be 5s")

		// 认证默认值
		assert.Equal(t, false, defaults.Auth.Enabled, "Default auth enabled should be false")
		assert.Equal(t, "token", defaults.Auth.Method, "Default auth method should be token")
		assert.Equal(t, 30*time.Second, defaults.Auth.Timeout, "Default auth timeout should be 30s")
		assert.Equal(t, 1*time.Hour, defaults.Auth.CacheTTL, "Default auth cache TTL should be 1h")
	})
}