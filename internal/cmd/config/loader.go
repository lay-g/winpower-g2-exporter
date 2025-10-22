package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// LoadConfig 定义配置加载的接口
type LoadConfig interface {
	// Load 从多个源加载配置
	Load(configPath string, flags *pflag.FlagSet) (*Config, error)

	// LoadFromDefaults 仅加载默认配置
	LoadFromDefaults() *Config

	// GetConfigPath 获取当前使用的配置文件路径
	GetConfigPath() string
}

// Config 统一配置结构
type Config struct {
	// Server 配置
	Server ServerConfig `yaml:"server" json:"server"`

	// WinPower 配置
	WinPower WinPowerConfig `yaml:"winpower" json:"winpower"`

	// Logging 配置
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// Storage 配置
	Storage StorageConfig `yaml:"storage" json:"storage"`

	// Collector 配置
	Collector CollectorConfig `yaml:"collector" json:"collector"`

	// Metrics 配置
	Metrics MetricsConfig `yaml:"metrics" json:"metrics"`

	// Energy 配置
	Energy EnergyConfig `yaml:"energy" json:"energy"`

	// Scheduler 配置
	Scheduler SchedulerConfig `yaml:"scheduler" json:"scheduler"`

	// 认证配置
	Auth AuthConfig `yaml:"auth" json:"auth"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Port        int           `yaml:"port" json:"port" env:"WINPOWER_EXPORTER_PORT" default:"9090"`
	Host        string        `yaml:"host" json:"host" env:"WINPOWER_EXPORTER_HOST" default:"0.0.0.0"`
	ReadTimeout time.Duration `yaml:"read_timeout" json:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" default:"30s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" json:"idle_timeout" default:"60s"`
	EnablePprof bool          `yaml:"enable_pprof" json:"enable_pprof" default:"false"`
}

// WinPowerConfig WinPower连接配置
type WinPowerConfig struct {
	URL            string        `yaml:"url" json:"url" env:"WINPOWER_EXPORTER_CONSOLE_URL"`
	Username       string        `yaml:"username" json:"username" env:"WINPOWER_EXPORTER_USERNAME"`
	Password       string        `yaml:"password" json:"password" env:"WINPOWER_EXPORTER_PASSWORD"`
	Timeout        time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	RetryInterval  time.Duration `yaml:"retry_interval" json:"retry_interval" default:"5s"`
	MaxRetries     int           `yaml:"max_retries" json:"max_retries" default:"3"`
	SkipSSLVerify  bool          `yaml:"skip_ssl_verify" json:"skip_ssl_verify" env:"WINPOWER_EXPORTER_SKIP_SSL_VERIFY" default:"false"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level" env:"WINPOWER_EXPORTER_LOG_LEVEL" default:"info"`
	Format     string `yaml:"format" json:"format" default:"json"`
	Output     string `yaml:"output" json:"output" default:"stdout"`
	Filename   string `yaml:"filename" json:"filename" default:""`
	MaxSize    int    `yaml:"max_size" json:"max_size" default:"100"`
	MaxAge     int    `yaml:"max_age" json:"max_age" default:"7"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups" default:"3"`
	Compress   bool   `yaml:"compress" json:"compress" default:"true"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir   string `yaml:"data_dir" json:"data_dir" env:"WINPOWER_EXPORTER_DATA_DIR" default:"./data"`
	SyncWrite bool   `yaml:"sync_write" json:"sync_write" env:"WINPOWER_EXPORTER_SYNC_WRITE" default:"true"`
}

// CollectorConfig 数据采集配置
type CollectorConfig struct {
	Enabled       bool          `yaml:"enabled" json:"enabled" default:"true"`
	Interval      time.Duration `yaml:"interval" json:"interval" default:"5s"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	MaxConcurrent int           `yaml:"max_concurrent" json:"max_concurrent" default:"5"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled" default:"true"`
	Path        string `yaml:"path" json:"path" default:"/metrics"`
	Namespace   string `yaml:"namespace" json:"namespace" default:"winpower"`
	Subsystem   string `yaml:"subsystem" json:"subsystem" default:"exporter"`
	HelpText    string `yaml:"help_text" json:"help_text" default:"WinPower G2 Exporter Metrics"`
}

// EnergyConfig 能耗计算配置
type EnergyConfig struct {
	Enabled       bool          `yaml:"enabled" json:"enabled" default:"true"`
	Interval      time.Duration `yaml:"interval" json:"interval" default:"5s"`
	Precision     int           `yaml:"precision" json:"precision" default:"3"`
	StoragePeriod time.Duration `yaml:"storage_period" json:"storage_period" default:"1h"`
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Enabled  bool          `yaml:"enabled" json:"enabled" default:"true"`
	Interval time.Duration `yaml:"interval" json:"interval" default:"5s"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled    bool          `yaml:"enabled" json:"enabled" default:"false"`
	Method     string        `yaml:"method" json:"method" default:"token"`
	TokenURL   string        `yaml:"token_url" json:"token_url"`
	Username   string        `yaml:"username" json:"username"`
	Password   string        `yaml:"password" json:"password"`
	Timeout    time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
	CacheTTL   time.Duration `yaml:"cache_ttl" json:"cache_ttl" default:"1h"`
}

// Loader 配置加载器实现
type Loader struct {
	v         *viper.Viper
	logger    *zap.Logger
	configPath string
}

// NewLoader 创建新的配置加载器
func NewLoader(logger *zap.Logger) *Loader {
	v := viper.New()

	return &Loader{
		v:      v,
		logger: logger,
	}
}

// Load 从多个源加载配置，按照优先级合并：
// 1. 默认值（最低优先级）
// 2. 配置文件
// 3. 环境变量（WINPOWER_EXPORTER_ 前缀）
// 4. 命令行参数（最高优先级）
func (l *Loader) Load(configPath string, flags *pflag.FlagSet) (*Config, error) {
	l.configPath = configPath

	// 1. 设置默认值
	l.setDefaults()

	// 2. 绑定环境变量（在加载配置文件之前绑定）
	if err := l.bindEnvVars(); err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// 3. 加载配置文件
	if err := l.loadConfigFile(configPath); err != nil {
		l.logger.Warn("Failed to load config file, using defaults", zap.Error(err))
	}

	// 4. 绑定命令行参数
	if err := l.bindFlags(flags); err != nil {
		return nil, fmt.Errorf("failed to bind command line flags: %w", err)
	}

	// 5. 解析到配置结构
	var config Config
	if err := l.v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. 手动处理环境变量覆盖（确保环境变量优先级）
	if err := l.applyEnvOverrides(&config); err != nil {
		return nil, fmt.Errorf("failed to apply environment variable overrides: %w", err)
	}

	// 7. 后处理和验证
	l.postProcess(&config)

	l.logger.Info("Configuration loaded successfully",
		zap.String("config_file", configPath),
		zap.String("server_host", config.Server.Host),
		zap.Int("server_port", config.Server.Port),
		zap.String("log_level", config.Logging.Level),
		zap.String("data_dir", config.Storage.DataDir))

	return &config, nil
}

// LoadFromDefaults 仅加载默认配置
func (l *Loader) LoadFromDefaults() *Config {
	l.setDefaults()

	var config Config
	if err := l.v.Unmarshal(&config); err != nil {
		l.logger.Error("Failed to unmarshal default config", zap.Error(err))
		return nil
	}

	l.postProcess(&config)
	return &config
}

// GetConfigPath 获取当前使用的配置文件路径
func (l *Loader) GetConfigPath() string {
	return l.configPath
}

// setDefaults 设置默认配置值
func (l *Loader) setDefaults() {
	// Server 配置
	l.v.SetDefault("server.port", 9090)
	l.v.SetDefault("server.host", "0.0.0.0")
	l.v.SetDefault("server.read_timeout", "30s")
	l.v.SetDefault("server.write_timeout", "30s")
	l.v.SetDefault("server.idle_timeout", "60s")
	l.v.SetDefault("server.enable_pprof", false)

	// WinPower 配置
	l.v.SetDefault("winpower.timeout", "30s")
	l.v.SetDefault("winpower.retry_interval", "5s")
	l.v.SetDefault("winpower.max_retries", 3)
	l.v.SetDefault("winpower.skip_ssl_verify", false)

	// Logging 配置
	l.v.SetDefault("logging.level", "info")
	l.v.SetDefault("logging.format", "json")
	l.v.SetDefault("logging.output", "stdout")
	l.v.SetDefault("logging.max_size", 100)
	l.v.SetDefault("logging.max_age", 7)
	l.v.SetDefault("logging.max_backups", 3)
	l.v.SetDefault("logging.compress", true)

	// Storage 配置
	l.v.SetDefault("storage.data_dir", "./data")
	l.v.SetDefault("storage.sync_write", true)

	// WinPower 配置 - 设置一些基本的默认值（用户仍然需要提供 URL、用户名和密码）
	l.v.SetDefault("winpower.url", "")
	l.v.SetDefault("winpower.username", "")
	l.v.SetDefault("winpower.password", "")

	// Collector 配置
	l.v.SetDefault("collector.enabled", true)
	l.v.SetDefault("collector.interval", "5s")
	l.v.SetDefault("collector.timeout", "30s")
	l.v.SetDefault("collector.max_concurrent", 5)

	// Metrics 配置
	l.v.SetDefault("metrics.enabled", true)
	l.v.SetDefault("metrics.path", "/metrics")
	l.v.SetDefault("metrics.namespace", "winpower")
	l.v.SetDefault("metrics.subsystem", "exporter")
	l.v.SetDefault("metrics.help_text", "WinPower G2 Exporter Metrics")

	// Energy 配置
	l.v.SetDefault("energy.enabled", true)
	l.v.SetDefault("energy.interval", "5s")
	l.v.SetDefault("energy.precision", 3)
	l.v.SetDefault("energy.storage_period", "1h")

	// Scheduler 配置
	l.v.SetDefault("scheduler.enabled", true)
	l.v.SetDefault("scheduler.interval", "5s")

	// Auth 配置
	l.v.SetDefault("auth.enabled", false)
	l.v.SetDefault("auth.method", "token")
	l.v.SetDefault("auth.timeout", "30s")
	l.v.SetDefault("auth.cache_ttl", "1h")
}

// loadConfigFile 加载配置文件
func (l *Loader) loadConfigFile(configPath string) error {
	if configPath == "" {
		return nil
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		l.logger.Info("Config file does not exist, using defaults", zap.String("path", configPath))
		return nil
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for config file: %w", err)
	}

	// 设置配置文件路径和名称
	l.v.SetConfigFile(absPath)

	// 根据文件扩展名设置配置格式
	ext := filepath.Ext(configPath)
	switch strings.ToLower(ext) {
	case ".yaml", ".yml":
		l.v.SetConfigType("yaml")
	case ".json":
		l.v.SetConfigType("json")
	case ".toml":
		l.v.SetConfigType("toml")
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	// 读取配置文件
	if err := l.v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	l.logger.Info("Config file loaded successfully",
		zap.String("path", absPath),
		zap.String("format", l.v.ConfigFileUsed()))

	return nil
}

// bindEnvVars 绑定环境变量
func (l *Loader) bindEnvVars() error {
	// 设置环境变量前缀
	l.v.SetEnvPrefix("WINPOWER_EXPORTER")

	// 设置环境变量键名替换器（将 . 替换为 _）
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 自动绑定环境变量
	l.v.AutomaticEnv()

	// 手动绑定一些可能无法自动绑定的环境变量
	envMappings := map[string]string{
		"storage.data_dir":     "WINPOWER_EXPORTER_DATA_DIR",
		"server.port":          "WINPOWER_EXPORTER_SERVER_PORT",
		"server.host":          "WINPOWER_EXPORTER_SERVER_HOST",
		"logging.level":        "WINPOWER_EXPORTER_LOG_LEVEL",
		"logging.format":       "WINPOWER_EXPORTER_LOG_FORMAT",
		"winpower.url":         "WINPOWER_EXPORTER_CONSOLE_URL",
		"winpower.username":    "WINPOWER_EXPORTER_USERNAME",
		"winpower.password":    "WINPOWER_EXPORTER_PASSWORD",
		"winpower.skip_ssl_verify": "WINPOWER_EXPORTER_SKIP_SSL_VERIFY",
		"storage.sync_write":   "WINPOWER_EXPORTER_SYNC_WRITE",
		"collector.enabled":    "WINPOWER_EXPORTER_COLLECTOR_ENABLED",
		"metrics.enabled":      "WINPOWER_EXPORTER_METRICS_ENABLED",
		"metrics.path":         "WINPOWER_EXPORTER_METRICS_PATH",
		"energy.enabled":       "WINPOWER_EXPORTER_ENERGY_ENABLED",
		"scheduler.enabled":    "WINPOWER_EXPORTER_SCHEDULER_ENABLED",
		"auth.enabled":         "WINPOWER_EXPORTER_AUTH_ENABLED",
	}

	for key, envVar := range envMappings {
		if err := l.v.BindEnv(key, envVar); err != nil {
			return fmt.Errorf("failed to bind env var %s to %s: %w", envVar, key, err)
		}
	}

	return nil
}

// bindFlags 绑定命令行参数
func (l *Loader) bindFlags(flags *pflag.FlagSet) error {
	if flags == nil {
		return nil
	}

	// 绑定所有命令行参数到 viper
	flags.VisitAll(func(flag *pflag.Flag) {
		key := flag.Name
		if err := l.v.BindPFlag(key, flags.Lookup(key)); err != nil {
			l.logger.Warn("Failed to bind flag", zap.String("flag", key), zap.Error(err))
		}
	})

	return nil
}

// applyEnvOverrides 手动应用环境变量覆盖
func (l *Loader) applyEnvOverrides(config *Config) error {
	// 手动检查和设置环境变量值
	if val := os.Getenv("WINPOWER_EXPORTER_DATA_DIR"); val != "" {
		config.Storage.DataDir = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_SYNC_WRITE"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Storage.SyncWrite = boolVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_SERVER_PORT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			config.Server.Port = intVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_SERVER_HOST"); val != "" {
		config.Server.Host = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_LOG_LEVEL"); val != "" {
		config.Logging.Level = strings.ToLower(val)
	}
	if val := os.Getenv("WINPOWER_EXPORTER_LOG_FORMAT"); val != "" {
		config.Logging.Format = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_CONSOLE_URL"); val != "" {
		config.WinPower.URL = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_USERNAME"); val != "" {
		config.WinPower.Username = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_PASSWORD"); val != "" {
		config.WinPower.Password = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_SKIP_SSL_VERIFY"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.WinPower.SkipSSLVerify = boolVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_COLLECTOR_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Collector.Enabled = boolVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_METRICS_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Metrics.Enabled = boolVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_METRICS_PATH"); val != "" {
		config.Metrics.Path = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_ENERGY_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Energy.Enabled = boolVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_SCHEDULER_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Scheduler.Enabled = boolVal
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_AUTH_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Auth.Enabled = boolVal
		}
	}

	return nil
}

// postProcess 对配置进行后处理
func (l *Loader) postProcess(config *Config) {
	// 确保数据目录存在
	if config.Storage.DataDir != "" {
		if err := os.MkdirAll(config.Storage.DataDir, 0755); err != nil {
			l.logger.Warn("Failed to create data directory",
				zap.String("dir", config.Storage.DataDir),
				zap.Error(err))
		}
	}

	// 规范化日志级别
	config.Logging.Level = strings.ToLower(config.Logging.Level)

	// 规范化主机地址
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}

	// 验证端口号
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		l.logger.Warn("Invalid port number, using default",
			zap.Int("port", config.Server.Port))
		config.Server.Port = 9090
	}

	// 验证必需的 WinPower 配置
	if config.WinPower.URL == "" {
		l.logger.Error("WinPower URL is required")
	}
	if config.WinPower.Username == "" {
		l.logger.Error("WinPower username is required")
	}
	if config.WinPower.Password == "" {
		l.logger.Error("WinPower password is required")
	}
}