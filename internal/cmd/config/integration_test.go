package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestConfigPriority 测试配置优先级：命令行参数 > 环境变量 > 配置文件 > 默认值
func TestConfigPriority(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	// 配置文件内容（优先级：第3级）
	configContent := `
server:
  port: 8080
  host: "config-host"
  read_timeout: "25s"
  write_timeout: "25s"
  idle_timeout: "55s"
  enable_pprof: true

logging:
  level: "debug"
  format: "console"
  output: "file"
  filename: "/tmp/config.log"
  max_size: 150
  max_age: 10
  max_backups: 5
  compress: false

winpower:
  url: "http://config.example.com"
  username: "configuser"
  password: "configpass"
  timeout: "45s"
  retry_interval: "8s"
  max_retries: 5
  skip_ssl_verify: true

storage:
  data_dir: "/config/data"
  sync_write: false

collector:
  enabled: false
  interval: "8s"
  timeout: "6s"
  max_concurrent: 8

metrics:
  enabled: false
  path: "/config-metrics"
  namespace: "config"
  subsystem: "config-exporter"
  help_text: "Config Metrics"

energy:
  enabled: false
  interval: "8s"
  precision: 5
  storage_period: "2h"

scheduler:
  enabled: false
  interval: "8s"

auth:
  enabled: true
  method: "basic"
  token_url: "http://config.example.com/token"
  username: "configauthuser"
  password: "configauthpass"
  timeout: "45s"
  cache_ttl: "2h"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// 设置环境变量（优先级：第2级）
	envVars := map[string]string{
		"WINPOWER_EXPORTER_SERVER_PORT":       "9000",
		"WINPOWER_EXPORTER_SERVER_HOST":       "env-host",
		"WINPOWER_EXPORTER_SERVER_ENABLE_PPROF": "false",
		"WINPOWER_EXPORTER_LOGGING_LEVEL":     "warn",
		"WINPOWER_EXPORTER_LOGGING_FORMAT":    "json",
		"WINPOWER_EXPORTER_LOGGING_MAX_SIZE":  "200",
		"WINPOWER_EXPORTER_CONSOLE_URL":       "http://env.example.com",
		"WINPOWER_EXPORTER_USERNAME":          "envuser",
		"WINPOWER_EXPORTER_PASSWORD":          "envpass",
		"WINPOWER_EXPORTER_SKIP_SSL_VERIFY":   "false",
		"WINPOWER_EXPORTER_DATA_DIR":          "/env/data",
		"WINPOWER_EXPORTER_SYNC_WRITE":        "true",
		"WINPOWER_EXPORTER_COLLECTOR_ENABLED": "true",
		"WINPOWER_EXPORTER_METRICS_ENABLED":   "true",
		"WINPOWER_EXPORTER_METRICS_PATH":      "/env-metrics",
		"WINPOWER_EXPORTER_ENERGY_ENABLED":    "true",
		"WINPOWER_EXPORTER_SCHEDULER_ENABLED": "true",
		"WINPOWER_EXPORTER_AUTH_ENABLED":      "false",
	}

	// 保存原始环境变量并设置测试环境变量
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

	// 创建命令行参数（优先级：第1级 - 最高优先级）
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("server.port", 0, "Server port")
	flags.String("server.host", "", "Server host")
	flags.Bool("server.enable_pprof", false, "Enable pprof")
	flags.String("logging.level", "", "Log level")
	flags.String("logging.format", "", "Log format")
	flags.String("logging.max_size", "", "Log max size")
	flags.String("storage.data_dir", "", "Data directory")
	flags.String("winpower.url", "", "WinPower URL")
	flags.String("winpower.username", "", "WinPower username")
	flags.String("winpower.password", "", "WinPower password")
	flags.Bool("collector.enabled", false, "Collector enabled")
	flags.Bool("metrics.enabled", false, "Metrics enabled")
	flags.String("metrics.path", "", "Metrics path")
	flags.Bool("energy.enabled", false, "Energy enabled")
	flags.Bool("scheduler.enabled", false, "Scheduler enabled")
	flags.Bool("auth.enabled", false, "Auth enabled")

	// 解析测试命令行参数
	args := []string{
		"--server.port=7000",
		"--server.host=cli-host",
		"--server.enable_pprof=true",
		"--logging.level=error",
		"--logging.max_size=300",
		"--storage.data_dir=/cli/data",
		"--winpower.url=http://cli.example.com",
		"--winpower.username=cliuser",
		"--winpower.password=clipass",
		"--collector.enabled=false",
		"--metrics.enabled=false",
		"--metrics.path=/cli-metrics",
		"--energy.enabled=false",
		"--scheduler.enabled=false",
		"--auth.enabled=true",
	}

	err = flags.Parse(args)
	require.NoError(t, err)

	// 加载配置（优先级：命令行参数 > 环境变量 > 配置文件 > 默认值）
	config, err := loader.Load(configFile, flags)
	require.NoError(t, err)
	require.NotNil(t, config)

	// 验证配置优先级
	t.Run("verify server config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, 7000, config.Server.Port, "Server port should use CLI flag (highest priority)")
		assert.Equal(t, "cli-host", config.Server.Host, "Server host should use CLI flag (highest priority)")
		assert.Equal(t, true, config.Server.EnablePprof, "Server enable_pprof should use CLI flag (highest priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, 25*time.Second, config.Server.ReadTimeout, "Server read_timeout should use config file value")
		assert.Equal(t, 25*time.Second, config.Server.WriteTimeout, "Server write_timeout should use config file value")
		assert.Equal(t, 55*time.Second, config.Server.IdleTimeout, "Server idle_timeout should use config file value")
	})

	t.Run("verify logging config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, "error", config.Logging.Level, "Log level should use CLI flag (highest priority)")
		assert.Equal(t, 300, config.Logging.MaxSize, "Log max_size should use CLI flag (highest priority)")

		// 环境变量 > 配置文件 > 默认值（命令行参数未设置）
		assert.Equal(t, "json", config.Logging.Format, "Log format should use environment variable (medium priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, "file", config.Logging.Output, "Log output should use config file value")
		assert.Equal(t, "/tmp/config.log", config.Logging.Filename, "Log filename should use config file value")
		assert.Equal(t, 10, config.Logging.MaxAge, "Log max_age should use config file value")
		assert.Equal(t, 5, config.Logging.MaxBackups, "Log max_backups should use config file value")
		assert.Equal(t, false, config.Logging.Compress, "Log compress should use config file value")
	})

	t.Run("verify WinPower config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, "http://cli.example.com", config.WinPower.URL, "WinPower URL should use CLI flag (highest priority)")
		assert.Equal(t, "cliuser", config.WinPower.Username, "WinPower username should use CLI flag (highest priority)")
		assert.Equal(t, "clipass", config.WinPower.Password, "WinPower password should use CLI flag (highest priority)")

		// 环境变量 > 配置文件 > 默认值（命令行参数未设置）
		assert.Equal(t, false, config.WinPower.SkipSSLVerify, "WinPower skip_ssl_verify should use environment variable (medium priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, 45*time.Second, config.WinPower.Timeout, "WinPower timeout should use config file value")
		assert.Equal(t, 8*time.Second, config.WinPower.RetryInterval, "WinPower retry_interval should use config file value")
		assert.Equal(t, 5, config.WinPower.MaxRetries, "WinPower max_retries should use config file value")
	})

	t.Run("verify storage config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, "/cli/data", config.Storage.DataDir, "Data directory should use CLI flag (highest priority)")

		// 环境变量 > 配置文件 > 默认值（命令行参数未设置）
		assert.Equal(t, true, config.Storage.SyncWrite, "Sync write should use environment variable (medium priority)")
	})

	t.Run("verify collector config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, false, config.Collector.Enabled, "Collector enabled should use CLI flag (highest priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, 8*time.Second, config.Collector.Interval, "Collector interval should use config file value")
		assert.Equal(t, 6*time.Second, config.Collector.Timeout, "Collector timeout should use config file value")
		assert.Equal(t, 8, config.Collector.MaxConcurrent, "Collector max_concurrent should use config file value")
	})

	t.Run("verify metrics config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, false, config.Metrics.Enabled, "Metrics enabled should use CLI flag (highest priority)")
		assert.Equal(t, "/cli-metrics", config.Metrics.Path, "Metrics path should use CLI flag (highest priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, "config", config.Metrics.Namespace, "Metrics namespace should use config file value")
		assert.Equal(t, "config-exporter", config.Metrics.Subsystem, "Metrics subsystem should use config file value")
		assert.Equal(t, "Config Metrics", config.Metrics.HelpText, "Metrics help_text should use config file value")
	})

	t.Run("verify energy config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, false, config.Energy.Enabled, "Energy enabled should use CLI flag (highest priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, 8*time.Second, config.Energy.Interval, "Energy interval should use config file value")
		assert.Equal(t, 5, config.Energy.Precision, "Energy precision should use config file value")
		assert.Equal(t, 2*time.Hour, config.Energy.StoragePeriod, "Energy storage_period should use config file value")
	})

	t.Run("verify scheduler config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, false, config.Scheduler.Enabled, "Scheduler enabled should use CLI flag (highest priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, 8*time.Second, config.Scheduler.Interval, "Scheduler interval should use config file value")
	})

	t.Run("verify auth config priority", func(t *testing.T) {
		// 命令行参数 > 环境变量 > 配置文件 > 默认值
		assert.Equal(t, true, config.Auth.Enabled, "Auth enabled should use CLI flag (highest priority)")

		// 配置文件 > 默认值（环境变量和命令行参数未设置）
		assert.Equal(t, "basic", config.Auth.Method, "Auth method should use config file value")
		assert.Equal(t, "http://config.example.com/token", config.Auth.TokenURL, "Auth token_url should use config file value")
		assert.Equal(t, "configauthuser", config.Auth.Username, "Auth username should use config file value")
		assert.Equal(t, "configauthpass", config.Auth.Password, "Auth password should use config file value")
		assert.Equal(t, 45*time.Second, config.Auth.Timeout, "Auth timeout should use config file value")
		assert.Equal(t, 2*time.Hour, config.Auth.CacheTTL, "Auth cache_ttl should use config file value")
	})

	t.Run("verify default values for unset fields", func(t *testing.T) {
		// 验证未被任何配置源设置的字段使用默认值
		// （在这个测试中，所有字段都已被某种配置源设置）
		// 这里我们添加一个额外的验证来确保配置加载正确
		assert.NotNil(t, config, "Config should not be nil")
		assert.True(t, config.Server.Port > 0, "Server port should be set")
		assert.NotEmpty(t, config.Server.Host, "Server host should be set")
		assert.NotEmpty(t, config.Logging.Level, "Log level should be set")
		assert.NotEmpty(t, config.Storage.DataDir, "Data directory should be set")
	})
}

// TestConfigPriorityWithPartialSources 测试部分配置源的情况
func TestConfigPriorityWithPartialSources(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	testCases := []struct {
		name          string
		hasConfigFile bool
		hasEnvVars    bool
		hasFlags      bool
		description   string
	}{
		{
			name:          "only_defaults",
			hasConfigFile: false,
			hasEnvVars:    false,
			hasFlags:      false,
			description:   "仅使用默认配置",
		},
		{
			name:          "config_file_only",
			hasConfigFile: true,
			hasEnvVars:    false,
			hasFlags:      false,
			description:   "仅使用配置文件和默认配置",
		},
		{
			name:          "env_vars_only",
			hasConfigFile: false,
			hasEnvVars:    true,
			hasFlags:      false,
			description:   "仅使用环境变量和默认配置",
		},
		{
			name:          "flags_only",
			hasConfigFile: false,
			hasEnvVars:    false,
			hasFlags:      true,
			description:   "仅使用命令行参数和默认配置",
		},
		{
			name:          "config_and_env",
			hasConfigFile: true,
			hasEnvVars:    true,
			hasFlags:      false,
			description:   "使用配置文件、环境变量和默认配置",
		},
		{
			name:          "env_and_flags",
			hasConfigFile: false,
			hasEnvVars:    true,
			hasFlags:      true,
			description:   "使用环境变量、命令行参数和默认配置",
		},
		{
			name:          "all_sources",
			hasConfigFile: true,
			hasEnvVars:    true,
			hasFlags:      true,
			description:   "使用所有配置源",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var configFile string
			var envVars map[string]string
			var flags *pflag.FlagSet

			// 设置配置文件
			if tc.hasConfigFile {
				tempDir := t.TempDir()
				configFile = filepath.Join(tempDir, "config.yaml")
				configContent := `
server:
  port: 8080
  host: "config-host"
logging:
  level: "debug"
winpower:
  url: "http://config.example.com"
  username: "configuser"
  password: "configpass"
storage:
  data_dir: "/config/data"
`
				err := os.WriteFile(configFile, []byte(configContent), 0644)
				require.NoError(t, err)
			}

			// 设置环境变量
			if tc.hasEnvVars {
				envVars = map[string]string{
					"WINPOWER_EXPORTER_SERVER_PORT": "9000",
					"WINPOWER_EXPORTER_LOGGING_LEVEL": "warn",
					"WINPOWER_EXPORTER_CONSOLE_URL":   "http://env.example.com",
					"WINPOWER_EXPORTER_USERNAME":      "envuser",
					"WINPOWER_EXPORTER_PASSWORD":      "envpass",
					"WINPOWER_EXPORTER_DATA_DIR":      "/env/data",
				}

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
			}

			// 设置命令行参数
			if tc.hasFlags {
				flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.Int("server.port", 0, "Server port")
				flags.String("server.host", "", "Server host")
				flags.String("logging.level", "", "Log level")
				flags.String("storage.data_dir", "", "Data directory")
				flags.String("winpower.url", "", "WinPower URL")
				flags.String("winpower.username", "", "WinPower username")
				flags.String("winpower.password", "", "WinPower password")

				args := []string{
					"--server.port=7000",
					"--server.host=cli-host",
					"--storage.data_dir=/cli/data",
					"--winpower.url=http://cli.example.com",
					"--winpower.username=cliuser",
					"--winpower.password=clipass",
				}

				err := flags.Parse(args)
				require.NoError(t, err)
			}

			// 加载配置
			config, err := loader.Load(configFile, flags)
			require.NoError(t, err)
			require.NotNil(t, config)

			// 根据测试用例验证配置值
			switch tc.name {
			case "only_defaults":
				assert.Equal(t, 9090, config.Server.Port, "Should use default port")
				assert.Equal(t, "0.0.0.0", config.Server.Host, "Should use default host")
				assert.Equal(t, "info", config.Logging.Level, "Should use default log level")
				assert.Equal(t, "./data", config.Storage.DataDir, "Should use default data directory")

			case "config_file_only":
				assert.Equal(t, 8080, config.Server.Port, "Should use config file port")
				assert.Equal(t, "config-host", config.Server.Host, "Should use config file host")
				assert.Equal(t, "debug", config.Logging.Level, "Should use config file log level")
				assert.Equal(t, "/config/data", config.Storage.DataDir, "Should use config file data directory")
				assert.Equal(t, "http://config.example.com", config.WinPower.URL, "Should use config file WinPower URL")

			case "env_vars_only":
				assert.Equal(t, 9000, config.Server.Port, "Should use environment variable port")
				assert.Equal(t, "0.0.0.0", config.Server.Host, "Should use default host")
				assert.Equal(t, "warn", config.Logging.Level, "Should use environment variable log level")
				assert.Equal(t, "/env/data", config.Storage.DataDir, "Should use environment variable data directory")
				assert.Equal(t, "http://env.example.com", config.WinPower.URL, "Should use environment variable WinPower URL")

			case "flags_only":
				assert.Equal(t, 7000, config.Server.Port, "Should use CLI flag port")
				assert.Equal(t, "cli-host", config.Server.Host, "Should use CLI flag host")
				assert.Equal(t, "info", config.Logging.Level, "Should use default log level")
				assert.Equal(t, "/cli/data", config.Storage.DataDir, "Should use CLI flag data directory")
				assert.Equal(t, "http://cli.example.com", config.WinPower.URL, "Should use CLI flag WinPower URL")

			case "config_and_env":
				assert.Equal(t, 9000, config.Server.Port, "Should use environment variable port (higher priority than config file)")
				assert.Equal(t, "config-host", config.Server.Host, "Should use config file host (environment variable not set)")
				assert.Equal(t, "warn", config.Logging.Level, "Should use environment variable log level (higher priority than config file)")
				assert.Equal(t, "/env/data", config.Storage.DataDir, "Should use environment variable data directory (higher priority than config file)")
				assert.Equal(t, "http://config.example.com", config.WinPower.URL, "Should use config file WinPower URL (environment variable not set)")

			case "env_and_flags":
				assert.Equal(t, 7000, config.Server.Port, "Should use CLI flag port (highest priority)")
				assert.Equal(t, "cli-host", config.Server.Host, "Should use CLI flag host")
				assert.Equal(t, "warn", config.Logging.Level, "Should use environment variable log level (CLI flag not set)")
				assert.Equal(t, "/cli/data", config.Storage.DataDir, "Should use CLI flag data directory (highest priority)")
				assert.Equal(t, "http://cli.example.com", config.WinPower.URL, "Should use CLI flag WinPower URL (highest priority)")

			case "all_sources":
				assert.Equal(t, 7000, config.Server.Port, "Should use CLI flag port (highest priority)")
				assert.Equal(t, "cli-host", config.Server.Host, "Should use CLI flag host (highest priority)")
				assert.Equal(t, "warn", config.Logging.Level, "Should use environment variable log level (CLI flag not set)")
				assert.Equal(t, "/cli/data", config.Storage.DataDir, "Should use CLI flag data directory (highest priority)")
				assert.Equal(t, "http://cli.example.com", config.WinPower.URL, "Should use CLI flag WinPower URL (highest priority)")
			}

			t.Logf("Test case '%s' (%s) passed: port=%d, host=%s, log_level=%s, data_dir=%s",
				tc.name, tc.description, config.Server.Port, config.Server.Host,
				config.Logging.Level, config.Storage.DataDir)
		})
	}
}

// TestConfigPriorityWithComplexScenario 测试复杂场景下的配置优先级
func TestConfigPriorityWithComplexScenario(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 创建多个配置文件来模拟复杂场景
	tempDir := t.TempDir()
	baseConfigFile := filepath.Join(tempDir, "base.yaml")

	// 基础配置文件
	baseConfigContent := `
server:
  port: 8080
  host: "base-host"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "60s"
  enable_pprof: false

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true

winpower:
  url: "http://base.example.com"
  username: "baseuser"
  password: "basepass"
  timeout: "30s"
  retry_interval: "5s"
  max_retries: 3
  skip_ssl_verify: false

storage:
  data_dir: "/base/data"
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

	err := os.WriteFile(baseConfigFile, []byte(baseConfigContent), 0644)
	require.NoError(t, err)

	// 模拟分阶段配置叠加的场景
	scenarios := []struct {
		name        string
		envVars     map[string]string
		cliArgs     []string
		description string
	}{
		{
			name: "stage1_base_only",
			envVars: map[string]string{},
			cliArgs: []string{},
			description: "仅基础配置",
		},
		{
			name: "stage2_with_env",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_SERVER_PORT":     "9000",
				"WINPOWER_EXPORTER_LOGGING_LEVEL":   "debug",
				"WINPOWER_EXPORTER_CONSOLE_URL":     "http://stage2.example.com",
				"WINPOWER_EXPORTER_USERNAME":        "stage2user",
				"WINPOWER_EXPORTER_PASSWORD":        "stage2pass",
				"WINPOWER_EXPORTER_SKIP_SSL_VERIFY": "true",
				"WINPOWER_EXPORTER_DATA_DIR":        "/stage2/data",
			},
			cliArgs: []string{},
			description: "基础配置 + 环境变量",
		},
		{
			name: "stage3_with_cli",
			envVars: map[string]string{
				"WINPOWER_EXPORTER_SERVER_PORT":     "9000",
				"WINPOWER_EXPORTER_LOGGING_LEVEL":   "debug",
				"WINPOWER_EXPORTER_CONSOLE_URL":     "http://stage2.example.com",
				"WINPOWER_EXPORTER_USERNAME":        "stage2user",
				"WINPOWER_EXPORTER_PASSWORD":        "stage2pass",
				"WINPOWER_EXPORTER_SKIP_SSL_VERIFY": "true",
				"WINPOWER_EXPORTER_DATA_DIR":        "/stage2/data",
			},
			cliArgs: []string{
				"--server.port=7000",
				"--server.host=stage3-host",
				"--server.enable_pprof=true",
				"--logging.format=console",
				"--storage.data_dir=/stage3/data",
				"--winpower.url=http://stage3.example.com",
				"--winpower.username=stage3user",
				"--winpower.password=stage3pass",
				"--collector.enabled=false",
				"--metrics.enabled=false",
				"--energy.enabled=false",
				"--scheduler.enabled=false",
				"--auth.enabled=true",
			},
			description: "基础配置 + 环境变量 + 命令行参数",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// 设置环境变量
			originalEnv := make(map[string]string)
			for key, value := range scenario.envVars {
				originalEnv[key] = os.Getenv(key)
				os.Setenv(key, value)
			}
			defer func() {
				for key, value := range originalEnv {
					if value == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, value)
					}
				}
			}()

			// 设置命令行参数
			var flags *pflag.FlagSet
			if len(scenario.cliArgs) > 0 {
				flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.Int("server.port", 0, "Server port")
				flags.String("server.host", "", "Server host")
				flags.Bool("server.enable_pprof", false, "Enable pprof")
				flags.String("logging.level", "", "Log level")
				flags.String("logging.format", "", "Log format")
				flags.String("storage.data_dir", "", "Data directory")
				flags.String("winpower.url", "", "WinPower URL")
				flags.String("winpower.username", "", "WinPower username")
				flags.String("winpower.password", "", "WinPower password")
				flags.Bool("collector.enabled", false, "Collector enabled")
				flags.Bool("metrics.enabled", false, "Metrics enabled")
				flags.Bool("energy.enabled", false, "Energy enabled")
				flags.Bool("scheduler.enabled", false, "Scheduler enabled")
				flags.Bool("auth.enabled", false, "Auth enabled")

				err := flags.Parse(scenario.cliArgs)
				require.NoError(t, err)
			}

			// 加载配置
			config, err := loader.Load(baseConfigFile, flags)
			require.NoError(t, err)
			require.NotNil(t, config)

			// 根据场景验证配置
			switch scenario.name {
			case "stage1_base_only":
				assert.Equal(t, 8080, config.Server.Port, "Stage 1: Should use base config port")
				assert.Equal(t, "base-host", config.Server.Host, "Stage 1: Should use base config host")
				assert.Equal(t, "info", config.Logging.Level, "Stage 1: Should use base config log level")
				assert.Equal(t, "json", config.Logging.Format, "Stage 1: Should use base config log format")
				assert.Equal(t, "/base/data", config.Storage.DataDir, "Stage 1: Should use base config data directory")
				assert.Equal(t, "http://base.example.com", config.WinPower.URL, "Stage 1: Should use base config WinPower URL")
				assert.Equal(t, true, config.Collector.Enabled, "Stage 1: Should use base config collector enabled")
				assert.Equal(t, true, config.Metrics.Enabled, "Stage 1: Should use base config metrics enabled")
				assert.Equal(t, true, config.Energy.Enabled, "Stage 1: Should use base config energy enabled")
				assert.Equal(t, true, config.Scheduler.Enabled, "Stage 1: Should use base config scheduler enabled")
				assert.Equal(t, false, config.Auth.Enabled, "Stage 1: Should use base config auth enabled")

			case "stage2_with_env":
				assert.Equal(t, 9000, config.Server.Port, "Stage 2: Should use environment variable port")
				assert.Equal(t, "base-host", config.Server.Host, "Stage 2: Should use base config host (env not set)")
				assert.Equal(t, "debug", config.Logging.Level, "Stage 2: Should use environment variable log level")
				assert.Equal(t, "json", config.Logging.Format, "Stage 2: Should use base config log format (env not set)")
				assert.Equal(t, "/stage2/data", config.Storage.DataDir, "Stage 2: Should use environment variable data directory")
				assert.Equal(t, "http://stage2.example.com", config.WinPower.URL, "Stage 2: Should use environment variable WinPower URL")
				assert.Equal(t, "stage2user", config.WinPower.Username, "Stage 2: Should use environment variable WinPower username")
				assert.Equal(t, "stage2pass", config.WinPower.Password, "Stage 2: Should use environment variable WinPower password")
				assert.Equal(t, true, config.WinPower.SkipSSLVerify, "Stage 2: Should use environment variable skip SSL verify")
				assert.Equal(t, true, config.Collector.Enabled, "Stage 2: Should use base config collector enabled")
				assert.Equal(t, true, config.Metrics.Enabled, "Stage 2: Should use base config metrics enabled")

			case "stage3_with_cli":
				assert.Equal(t, 7000, config.Server.Port, "Stage 3: Should use CLI flag port (highest priority)")
				assert.Equal(t, "stage3-host", config.Server.Host, "Stage 3: Should use CLI flag host (highest priority)")
				assert.Equal(t, true, config.Server.EnablePprof, "Stage 3: Should use CLI flag enable_pprof (highest priority)")
				assert.Equal(t, "debug", config.Logging.Level, "Stage 3: Should use environment variable log level (CLI not set)")
				assert.Equal(t, "console", config.Logging.Format, "Stage 3: Should use CLI flag log format (highest priority)")
				assert.Equal(t, "/stage3/data", config.Storage.DataDir, "Stage 3: Should use CLI flag data directory (highest priority)")
				assert.Equal(t, "http://stage3.example.com", config.WinPower.URL, "Stage 3: Should use CLI flag WinPower URL (highest priority)")
				assert.Equal(t, "stage3user", config.WinPower.Username, "Stage 3: Should use CLI flag WinPower username (highest priority)")
				assert.Equal(t, "stage3pass", config.WinPower.Password, "Stage 3: Should use CLI flag WinPower password (highest priority)")
				assert.Equal(t, true, config.WinPower.SkipSSLVerify, "Stage 3: Should use environment variable skip SSL verify (CLI not set)")
				assert.Equal(t, false, config.Collector.Enabled, "Stage 3: Should use CLI flag collector enabled (highest priority)")
				assert.Equal(t, false, config.Metrics.Enabled, "Stage 3: Should use CLI flag metrics enabled (highest priority)")
				assert.Equal(t, false, config.Energy.Enabled, "Stage 3: Should use CLI flag energy enabled (highest priority)")
				assert.Equal(t, false, config.Scheduler.Enabled, "Stage 3: Should use CLI flag scheduler enabled (highest priority)")
				assert.Equal(t, true, config.Auth.Enabled, "Stage 3: Should use CLI flag auth enabled (highest priority)")
			}

			t.Logf("Scenario '%s' (%s) verified successfully", scenario.name, scenario.description)
			t.Logf("  Final config: port=%d, host=%s, log_level=%s, data_dir=%s, winpower_url=%s",
				config.Server.Port, config.Server.Host, config.Logging.Level,
				config.Storage.DataDir, config.WinPower.URL)
		})
	}
}

// TestConfigPriorityEdgeCases 测试配置优先级的边界情况
func TestConfigPriorityEdgeCases(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	t.Run("empty_string_values", func(t *testing.T) {
		// 测试空字符串值的优先级处理
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		configContent := `
server:
  host: "config-host"
  port: 8080
logging:
  level: "debug"
  format: ""
`
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 设置环境变量，包含空字符串值
		envVars := map[string]string{
			"WINPOWER_EXPORTER_SERVER_HOST":     "",      // 空字符串
			"WINPOWER_EXPORTER_LOGGING_LEVEL":   "warn",  // 非空字符串
			"WINPOWER_EXPORTER_LOGGING_FORMAT":  "json",  // 非空字符串
		}

		originalEnv := make(map[string]string)
		for key, value := range envVars {
			originalEnv[key] = os.Getenv(key)
			os.Setenv(key, value)
		}
		defer func() {
			for key, value := range originalEnv {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
		}()

		// 设置命令行参数，包含空字符串值
		flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
		flags.String("server.host", "", "Server host")
		flags.String("logging.level", "", "Log level")
		flags.String("logging.format", "", "Log format")

		args := []string{
			"--server.host=",      // 显式设置空字符串
			"--logging.level",     // 未设置值
			"--logging.format",    // 未设置值
		}

		err = flags.Parse(args)
		require.NoError(t, err)

		config, err := loader.Load(configFile, flags)
		require.NoError(t, err)

		// 验证空字符串值的处理
		assert.Equal(t, "", config.Server.Host, "Empty string from CLI flag should override config file")
		assert.Equal(t, "warn", config.Logging.Level, "Environment variable should override config file")
		assert.Equal(t, "json", config.Logging.Format, "Environment variable should override config file")
		assert.Equal(t, 8080, config.Server.Port, "Config file value should be preserved when not overridden")
	})

	t.Run("zero_values", func(t *testing.T) {
		// 测试零值的优先级处理
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		configContent := `
server:
  port: 8080
  enable_pprof: true
collector:
  max_concurrent: 5
`
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 设置环境变量，包含零值
		envVars := map[string]string{
			"WINPOWER_EXPORTER_SERVER_PORT":       "0",     // 零值
			"WINPOWER_EXPORTER_SERVER_ENABLE_PPROF": "false", // 布尔零值
		}

		originalEnv := make(map[string]string)
		for key, value := range envVars {
			originalEnv[key] = os.Getenv(key)
			os.Setenv(key, value)
		}
		defer func() {
			for key, value := range originalEnv {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
		}()

		config, err := loader.Load(configFile, nil)
		require.NoError(t, err)

		// 验证零值处理
		assert.Equal(t, 0, config.Server.Port, "Zero value from environment should override config file")
		assert.Equal(t, false, config.Server.EnablePprof, "Boolean zero value from environment should override config file")
		assert.Equal(t, 5, config.Collector.MaxConcurrent, "Config file value should be preserved when not overridden")
	})

	t.Run("invalid_values_in_higher_priority", func(t *testing.T) {
		// 测试高优先级配置源中的无效值
		tempDir := t.TempDir()
		configFile := filepath.Join(tempDir, "config.yaml")
		configContent := `
server:
  port: 8080
  host: "valid-host"
logging:
  level: "info"
`
		err := os.WriteFile(configFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 设置包含无效值的环境变量
		envVars := map[string]string{
			"WINPOWER_EXPORTER_SERVER_PORT":     "invalid", // 无效端口号
			"WINPOWER_EXPORTER_LOGGING_LEVEL":   "invalid", // 无效日志级别
		}

		originalEnv := make(map[string]string)
		for key, value := range envVars {
			originalEnv[key] = os.Getenv(key)
			os.Setenv(key, value)
		}
		defer func() {
			for key, value := range originalEnv {
				if value == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, value)
				}
			}
		}()

		// Viper 会尝试转换无效值，但可能会失败或产生意外结果
		// 这个测试主要验证加载过程不会崩溃
		config, err := loader.Load(configFile, nil)
		require.NoError(t, err, "Loading should not fail even with invalid environment values")

		// 验证配置加载的行为
		t.Logf("Loaded config with invalid env values: port=%d, log_level=%s",
			config.Server.Port, config.Logging.Level)
	})
}