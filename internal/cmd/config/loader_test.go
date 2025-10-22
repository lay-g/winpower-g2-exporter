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

func TestLoader_LoadFromDefaults(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	config := loader.LoadFromDefaults()

	require.NotNil(t, config, "Default config should not be nil")

	// 验证默认服务器配置
	assert.Equal(t, 9090, config.Server.Port, "Default server port should be 9090")
	assert.Equal(t, "0.0.0.0", config.Server.Host, "Default server host should be 0.0.0.0")
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout, "Default read timeout should be 30s")
	assert.Equal(t, false, config.Server.EnablePprof, "Default pprof should be disabled")

	// 验证默认日志配置
	assert.Equal(t, "info", config.Logging.Level, "Default log level should be info")
	assert.Equal(t, "json", config.Logging.Format, "Default log format should be json")
	assert.Equal(t, "stdout", config.Logging.Output, "Default log output should be stdout")

	// 验证默认存储配置
	assert.Equal(t, "./data", config.Storage.DataDir, "Default data directory should be ./data")
	assert.NotNil(t, config.Storage.SyncWrite, "Default sync write should not be nil")
	assert.Equal(t, true, *config.Storage.SyncWrite, "Default sync write should be true")

	// 验证默认采集器配置
	assert.NotNil(t, config.Collector.Enabled, "Collector enabled should not be nil")
	assert.Equal(t, true, *config.Collector.Enabled, "Collector should be enabled by default")
	assert.Equal(t, 5*time.Second, config.Collector.Interval, "Default collector interval should be 5s")

	// 验证默认指标配置
	assert.NotNil(t, config.Metrics.Enabled, "Metrics enabled should not be nil")
	assert.Equal(t, true, *config.Metrics.Enabled, "Metrics should be enabled by default")
	assert.Equal(t, "/metrics", config.Metrics.Path, "Default metrics path should be /metrics")
	assert.Equal(t, "winpower", config.Metrics.Namespace, "Default metrics namespace should be winpower")
}

func TestLoader_LoadWithConfigFile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `
server:
  port: 8080
  host: "127.0.0.1"
  enable_pprof: true

logging:
  level: "debug"
  format: "console"

winpower:
  url: "http://test.example.com"
  username: "testuser"
  password: "testpass"
  timeout: "60s"

storage:
  data_dir: "/tmp/test-data"
  sync_write: false
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// 加载配置
	config, err := loader.Load(configFile, nil)
	require.NoError(t, err)
	require.NotNil(t, config)

	// 验证配置文件中的值被正确加载
	assert.Equal(t, 8080, config.Server.Port, "Server port should be loaded from config file")
	assert.Equal(t, "127.0.0.1", config.Server.Host, "Server host should be loaded from config file")
	assert.Equal(t, true, config.Server.EnablePprof, "Server pprof should be loaded from config file")

	assert.Equal(t, "debug", config.Logging.Level, "Log level should be loaded from config file")
	assert.Equal(t, "console", config.Logging.Format, "Log format should be loaded from config file")

	assert.Equal(t, "http://test.example.com", config.WinPower.URL, "WinPower URL should be loaded from config file")
	assert.Equal(t, "testuser", config.WinPower.Username, "WinPower username should be loaded from config file")
	assert.Equal(t, "testpass", config.WinPower.Password, "WinPower password should be loaded from config file")
	assert.Equal(t, 60*time.Second, config.WinPower.Timeout, "WinPower timeout should be loaded from config file")

	assert.Equal(t, "/tmp/test-data", config.Storage.DataDir, "Data directory should be loaded from config file")
	assert.Equal(t, false, config.Storage.SyncWrite, "Sync write should be loaded from config file")

	// 验证未在配置文件中指定的值使用默认值
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout, "Read timeout should use default value")
	assert.Equal(t, true, config.Collector.Enabled, "Collector should use default enabled value")
}

func TestLoader_LoadWithEnvVars(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 设置环境变量
	envVars := map[string]string{
		"WINPOWER_EXPORTER_SERVER_PORT":     "9000",
		"WINPOWER_EXPORTER_LOGGING_LEVEL":   "warn",
		"WINPOWER_EXPORTER_DATA_DIR":        "/custom/data",
		"WINPOWER_EXPORTER_CONSOLE_URL":     "http://env.example.com",
		"WINPOWER_EXPORTER_USERNAME":        "envuser",
		"WINPOWER_EXPORTER_PASSWORD":        "envpass",
		"WINPOWER_EXPORTER_SKIP_SSL_VERIFY": "true",
	}

	// 保存原始环境变量并设置测试环境变量
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

	// 加载配置
	config, err := loader.Load("", nil)
	require.NoError(t, err)
	require.NotNil(t, config)

	// 验证环境变量值被正确加载
	assert.Equal(t, 9000, config.Server.Port, "Server port should be loaded from environment variable")
	assert.Equal(t, "warn", config.Logging.Level, "Log level should be loaded from environment variable")
	assert.Equal(t, "/custom/data", config.Storage.DataDir, "Data directory should be loaded from environment variable")
	assert.Equal(t, "http://env.example.com", config.WinPower.URL, "WinPower URL should be loaded from environment variable")
	assert.Equal(t, "envuser", config.WinPower.Username, "WinPower username should be loaded from environment variable")
	assert.Equal(t, "envpass", config.WinPower.Password, "WinPower password should be loaded from environment variable")
	assert.Equal(t, true, config.WinPower.SkipSSLVerify, "Skip SSL verify should be loaded from environment variable")

	// 验证未在环境变量中指定的值使用默认值
	assert.Equal(t, "0.0.0.0", config.Server.Host, "Server host should use default value")
	assert.Equal(t, "json", config.Logging.Format, "Log format should use default value")
}

func TestLoader_LoadWithFlags(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 创建命令行参数
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("server.port", 0, "Server port")
	flags.String("server.host", "", "Server host")
	flags.Bool("server.enable_pprof", false, "Enable pprof")
	flags.String("logging.level", "", "Log level")
	flags.String("storage.data_dir", "", "Data directory")
	flags.String("winpower.url", "", "WinPower URL")
	flags.String("winpower.username", "", "WinPower username")
	flags.String("winpower.password", "", "WinPower password")

	// 解析测试参数
	args := []string{
		"--server.port=7000",
		"--server.host=192.168.1.100",
		"--server.enable_pprof=true",
		"--logging.level=error",
		"--storage.data_dir=/flag/data",
		"--winpower.url=http://flag.example.com",
		"--winpower.username=flaguser",
		"--winpower.password=flagpass",
	}

	err := flags.Parse(args)
	require.NoError(t, err)

	// 加载配置
	config, err := loader.Load("", flags)
	require.NoError(t, err)
	require.NotNil(t, config)

	// 验证命令行参数值被正确加载
	assert.Equal(t, 7000, config.Server.Port, "Server port should be loaded from command line flag")
	assert.Equal(t, "192.168.1.100", config.Server.Host, "Server host should be loaded from command line flag")
	assert.Equal(t, true, config.Server.EnablePprof, "Server pprof should be loaded from command line flag")
	assert.Equal(t, "error", config.Logging.Level, "Log level should be loaded from command line flag")
	assert.Equal(t, "/flag/data", config.Storage.DataDir, "Data directory should be loaded from command line flag")
	assert.Equal(t, "http://flag.example.com", config.WinPower.URL, "WinPower URL should be loaded from command line flag")
	assert.Equal(t, "flaguser", config.WinPower.Username, "WinPower username should be loaded from command line flag")
	assert.Equal(t, "flagpass", config.WinPower.Password, "WinPower password should be loaded from command line flag")

	// 验证未在命令行参数中指定的值使用默认值
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout, "Read timeout should use default value")
	assert.Equal(t, "json", config.Logging.Format, "Log format should use default value")
}

func TestLoader_ConfigPriority(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 创建临时配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `
server:
  port: 8080
  host: "config-host"
logging:
  level: "debug"
storage:
  data_dir: "/config/data"
winpower:
  url: "http://config.example.com"
  username: "configuser"
  password: "configpass"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// 设置环境变量
	envVars := map[string]string{
		"WINPOWER_EXPORTER_SERVER_PORT":   "9000",
		"WINPOWER_EXPORTER_LOGGING_LEVEL": "warn",
		"WINPOWER_EXPORTER_DATA_DIR":      "/env/data",
	}

	// 保存原始环境变量并设置测试环境变量
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

	// 创建命令行参数
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.Int("server.port", 0, "Server port")
	flags.String("logging.level", "", "Log level")
	flags.String("storage.data_dir", "", "Data directory")

	// 解析测试参数
	args := []string{
		"--server.port=7000",
		"--logging.level=error",
	}
	err = flags.Parse(args)
	require.NoError(t, err)

	// 加载配置（优先级：命令行参数 > 环境变量 > 配置文件 > 默认值）
	config, err := loader.Load(configFile, flags)
	require.NoError(t, err)
	require.NotNil(t, config)

	// 验证优先级：命令行参数 > 环境变量 > 配置文件 > 默认值
	assert.Equal(t, 7000, config.Server.Port, "Server port should use command line flag (highest priority)")
	assert.Equal(t, "config-host", config.Server.Host, "Server host should use config file value")
	assert.Equal(t, "error", config.Logging.Level, "Log level should use command line flag (highest priority)")
	assert.Equal(t, "/env/data", config.Storage.DataDir, "Data directory should use environment variable (medium priority)")
	assert.Equal(t, "http://config.example.com", config.WinPower.URL, "WinPower URL should use config file value")
	assert.Equal(t, "configuser", config.WinPower.Username, "WinPower username should use config file value")
	assert.Equal(t, "configpass", config.WinPower.Password, "WinPower password should use config file value")
}

func TestLoader_InvalidConfigFile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 创建无效的配置文件
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid.yaml")

	invalidContent := `
server:
  port: invalid_port_number
  host: []
`

	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	// 加载配置应该返回错误
	config, err := loader.Load(configFile, nil)
	assert.Error(t, err, "Loading invalid config should return error")
	assert.Nil(t, config, "Config should be nil when loading fails")
}

func TestLoader_NonExistentConfigFile(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 尝试加载不存在的配置文件
	nonExistentFile := "/tmp/non_existent_config.yaml"

	// 加载配置应该成功，使用默认值
	config, err := loader.Load(nonExistentFile, nil)
	assert.NoError(t, err, "Loading non-existent config should not return error")
	assert.NotNil(t, config, "Config should not be nil when using defaults")

	// 验证使用的是默认值
	assert.Equal(t, 9090, config.Server.Port, "Should use default server port")
	assert.Equal(t, "info", config.Logging.Level, "Should use default log level")
}

func TestLoader_GetConfigPath(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 测试初始状态
	assert.Equal(t, "", loader.GetConfigPath(), "Initial config path should be empty")

	// 测试设置配置文件路径后
	configPath := "/test/config.yaml"
	_, err := loader.Load(configPath, nil)
	assert.NoError(t, err)
	assert.Equal(t, configPath, loader.GetConfigPath(), "Config path should be set after loading")
}

func TestLoader_PostProcessing(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	// 测试后处理功能
	config := loader.LoadFromDefaults()
	require.NotNil(t, config)

	// 验证日志级别被规范化
	config.Logging.Level = "DEBUG"
	loader.postProcess(config)
	assert.Equal(t, "debug", config.Logging.Level, "Log level should be normalized to lowercase")

	// 验证空主机地址被设置默认值
	config.Server.Host = ""
	loader.postProcess(config)
	assert.Equal(t, "0.0.0.0", config.Server.Host, "Empty host should be set to default")

	// 测试无效端口号的修正
	config.Server.Port = -1
	loader.postProcess(config)
	assert.Equal(t, 9090, config.Server.Port, "Invalid port should be corrected to default")
}

func TestLoader_SupportedConfigFormats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	testCases := []struct {
		name     string
		content  string
		filename string
	}{
		{
			name:     "YAML format",
			content:  "server:\n  port: 8080",
			filename: "config.yaml",
		},
		{
			name:     "JSON format",
			content:  `{"server": {"port": 8080}}`,
			filename: "config.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configFile := filepath.Join(tempDir, tc.filename)

			err := os.WriteFile(configFile, []byte(tc.content), 0644)
			require.NoError(t, err)

			config, err := loader.Load(configFile, nil)
			assert.NoError(t, err, "Should load %s successfully", tc.name)
			assert.NotNil(t, config, "Config should not be nil for %s", tc.name)
		})
	}
}

func TestLoader_UnsupportedConfigFormat(t *testing.T) {
	logger := zaptest.NewLogger(t)
	loader := NewLoader(logger)

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.txt")

	err := os.WriteFile(configFile, []byte("server:\n  port: 8080"), 0644)
	require.NoError(t, err)

	// 加载不支持的格式应该返回错误
	config, err := loader.Load(configFile, nil)
	assert.Error(t, err, "Loading unsupported format should return error")
	assert.Nil(t, config, "Config should be nil when format is unsupported")
}
