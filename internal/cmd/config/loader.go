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
	Server ServerConfig `yaml:"server" json:"server" mapstructure:"server"`

	// WinPower 配置
	WinPower WinPowerConfig `yaml:"winpower" json:"winpower" mapstructure:"winpower"`

	// Logging 配置
	Logging LoggingConfig `yaml:"logging" json:"logging" mapstructure:"logging"`

	// Storage 配置
	Storage StorageConfig `yaml:"storage" json:"storage" mapstructure:"storage"`

	// Collector 配置
	Collector CollectorConfig `yaml:"collector" json:"collector" mapstructure:"collector"`

	// Metrics 配置
	Metrics MetricsConfig `yaml:"metrics" json:"metrics" mapstructure:"metrics"`

	// Energy 配置
	Energy EnergyConfig `yaml:"energy" json:"energy" mapstructure:"energy"`

	// Scheduler 配置
	Scheduler SchedulerConfig `yaml:"scheduler" json:"scheduler" mapstructure:"scheduler"`

	// 认证配置
	Auth AuthConfig `yaml:"auth" json:"auth" mapstructure:"auth"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port" json:"port" mapstructure:"port" env:"WINPOWER_EXPORTER_PORT" default:"9090"`
	Host         string        `yaml:"host" json:"host" mapstructure:"host" env:"WINPOWER_EXPORTER_HOST" default:"0.0.0.0"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout" default:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout" mapstructure:"idle_timeout" default:"60s"`
	EnablePprof  bool          `yaml:"enable_pprof" json:"enable_pprof" mapstructure:"enable_pprof" default:"false"`
}

// WinPowerConfig WinPower连接配置
type WinPowerConfig struct {
	URL           string        `yaml:"url" json:"url" mapstructure:"url" env:"WINPOWER_EXPORTER_CONSOLE_URL"`
	Username      string        `yaml:"username" json:"username" mapstructure:"username" env:"WINPOWER_EXPORTER_USERNAME"`
	Password      string        `yaml:"password" json:"password" mapstructure:"password" env:"WINPOWER_EXPORTER_PASSWORD"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout" default:"30s"`
	RetryInterval time.Duration `yaml:"retry_interval" json:"retry_interval" mapstructure:"retry_interval" default:"5s"`
	MaxRetries    int           `yaml:"max_retries" json:"max_retries" mapstructure:"max_retries" default:"3"`
	SkipSSLVerify bool          `yaml:"skip_ssl_verify" json:"skip_ssl_verify" mapstructure:"skip_ssl_verify" env:"WINPOWER_EXPORTER_SKIP_SSL_VERIFY" default:"false"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level" json:"level" mapstructure:"level" env:"WINPOWER_EXPORTER_LOG_LEVEL" default:"info"`
	Format     string `yaml:"format" json:"format" mapstructure:"format" default:"json"`
	Output     string `yaml:"output" json:"output" mapstructure:"output" default:"stdout"`
	Filename   string `yaml:"filename" json:"filename" mapstructure:"filename" default:""`
	MaxSize    int    `yaml:"max_size" json:"max_size" mapstructure:"max_size" default:"100"`
	MaxAge     int    `yaml:"max_age" json:"max_age" mapstructure:"max_age" default:"7"`
	MaxBackups int    `yaml:"max_backups" json:"max_backups" mapstructure:"max_backups" default:"3"`
	Compress   *bool  `yaml:"compress" json:"compress" mapstructure:"compress" default:"true"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir   string `yaml:"data_dir" json:"data_dir" mapstructure:"data_dir" env:"WINPOWER_EXPORTER_DATA_DIR" default:"./data"`
	SyncWrite *bool  `yaml:"sync_write" json:"sync_write" mapstructure:"sync_write" env:"WINPOWER_EXPORTER_SYNC_WRITE" default:"true"`
}

// CollectorConfig 数据采集配置
type CollectorConfig struct {
	Enabled       *bool         `yaml:"enabled" json:"enabled" mapstructure:"enabled" default:"true"`
	Interval      time.Duration `yaml:"interval" json:"interval" mapstructure:"interval" default:"5s"`
	Timeout       time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout" default:"30s"`
	MaxConcurrent int           `yaml:"max_concurrent" json:"max_concurrent" mapstructure:"max_concurrent" default:"5"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled   *bool  `yaml:"enabled" json:"enabled" mapstructure:"enabled" default:"true"`
	Path      string `yaml:"path" json:"path" mapstructure:"path" default:"/metrics"`
	Namespace string `yaml:"namespace" json:"namespace" mapstructure:"namespace" default:"winpower"`
	Subsystem string `yaml:"subsystem" json:"subsystem" mapstructure:"subsystem" default:"exporter"`
	HelpText  string `yaml:"help_text" json:"help_text" mapstructure:"help_text" default:"WinPower G2 Exporter Metrics"`
}

// EnergyConfig 能耗计算配置
type EnergyConfig struct {
	Enabled       *bool         `yaml:"enabled" json:"enabled" mapstructure:"enabled" default:"true"`
	Interval      time.Duration `yaml:"interval" json:"interval" mapstructure:"interval" default:"5s"`
	Precision     int           `yaml:"precision" json:"precision" mapstructure:"precision" default:"3"`
	StoragePeriod time.Duration `yaml:"storage_period" json:"storage_period" mapstructure:"storage_period" default:"1h"`
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	Enabled  *bool         `yaml:"enabled" json:"enabled" mapstructure:"enabled" default:"true"`
	Interval time.Duration `yaml:"interval" json:"interval" mapstructure:"interval" default:"5s"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled  bool          `yaml:"enabled" json:"enabled" mapstructure:"enabled" default:"false"`
	Method   string        `yaml:"method" json:"method" mapstructure:"method" default:"token"`
	TokenURL string        `yaml:"token_url" json:"token_url" mapstructure:"token_url"`
	Username string        `yaml:"username" json:"username" mapstructure:"username"`
	Password string        `yaml:"password" json:"password" mapstructure:"password"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout" default:"30s"`
	CacheTTL time.Duration `yaml:"cache_ttl" json:"cache_ttl" mapstructure:"cache_ttl" default:"1h"`
}

// Loader 配置加载器实现
type Loader struct {
	v          *viper.Viper
	logger     *zap.Logger
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

	// 6. 应用硬编码默认值(确保未设置的字段有默认值，但不覆盖 bool 字段)
	l.applyHardcodedDefaults(&config, false)

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

	// Viper的SetDefault不会自动填充到unmarshal的结构中
	// 需要手动设置默认值，包括 bool 字段
	l.applyHardcodedDefaults(&config, true)
	l.postProcess(&config)
	return &config
}

// applyHardcodedDefaults 应用硬编码的默认值
// 这是必需的，因为Viper的SetDefault不会自动填充到unmarshal的结构中
// onlyForDefaults 参数：true 表示这是 LoadFromDefaults 调用，会设置 bool 默认值
//
//	false 表示这是 Load 调用，不设置 bool 默认值（避免覆盖）
func (l *Loader) applyHardcodedDefaults(config *Config, onlyForDefaults bool) {
	// Server 配置
	if config.Server.Port == 0 {
		config.Server.Port = 9090
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 30 * time.Second
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = 60 * time.Second
	}

	// WinPower 配置
	if config.WinPower.Timeout == 0 {
		config.WinPower.Timeout = 30 * time.Second
	}
	if config.WinPower.RetryInterval == 0 {
		config.WinPower.RetryInterval = 5 * time.Second
	}
	if config.WinPower.MaxRetries == 0 {
		config.WinPower.MaxRetries = 3
	}

	// Logging 配置
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}
	if config.Logging.Output == "" {
		config.Logging.Output = "stdout"
	}
	if config.Logging.MaxSize == 0 {
		config.Logging.MaxSize = 100
	}
	if config.Logging.MaxAge == 0 {
		config.Logging.MaxAge = 7
	}
	if config.Logging.MaxBackups == 0 {
		config.Logging.MaxBackups = 3
	}
	// Bool 字段：只在 LoadFromDefaults 时设置默认值
	if onlyForDefaults && config.Logging.Compress == nil {
		config.Logging.Compress = BoolPtr(true)
	}

	// Storage 配置
	if config.Storage.DataDir == "" {
		config.Storage.DataDir = "./data"
	}
	// Bool 字段：只在 LoadFromDefaults 时设置默认值
	if onlyForDefaults && config.Storage.SyncWrite == nil {
		config.Storage.SyncWrite = BoolPtr(true)
	}

	// Collector 配置
	// Bool 字段：只在 LoadFromDefaults 时设置默认值
	if onlyForDefaults && config.Collector.Enabled == nil {
		config.Collector.Enabled = BoolPtr(true)
	}
	if config.Collector.Interval == 0 {
		config.Collector.Interval = 5 * time.Second
	}
	if config.Collector.Timeout == 0 {
		config.Collector.Timeout = 30 * time.Second
	}
	if config.Collector.MaxConcurrent == 0 {
		config.Collector.MaxConcurrent = 5
	}

	// Metrics 配置
	// Bool 字段：只在 LoadFromDefaults 时设置默认值
	if onlyForDefaults && config.Metrics.Enabled == nil {
		config.Metrics.Enabled = BoolPtr(true)
	}
	if config.Metrics.Path == "" {
		config.Metrics.Path = "/metrics"
	}
	if config.Metrics.Namespace == "" {
		config.Metrics.Namespace = "winpower"
	}
	if config.Metrics.Subsystem == "" {
		config.Metrics.Subsystem = "exporter"
	}
	if config.Metrics.HelpText == "" {
		config.Metrics.HelpText = "WinPower G2 Exporter Metrics"
	}

	// Energy 配置
	// Bool 字段：只在 LoadFromDefaults 时设置默认值
	if onlyForDefaults && config.Energy.Enabled == nil {
		config.Energy.Enabled = BoolPtr(true)
	}
	if config.Energy.Interval == 0 {
		config.Energy.Interval = 5 * time.Second
	}
	if config.Energy.Precision == 0 {
		config.Energy.Precision = 3
	}
	if config.Energy.StoragePeriod == 0 {
		config.Energy.StoragePeriod = 1 * time.Hour
	}

	// Scheduler 配置
	// Bool 字段：只在 LoadFromDefaults 时设置默认值
	if onlyForDefaults && config.Scheduler.Enabled == nil {
		config.Scheduler.Enabled = BoolPtr(true)
	}
	if config.Scheduler.Interval == 0 {
		config.Scheduler.Interval = 5 * time.Second
	}

	// Auth 配置
	// Enabled 默认为 false，无需设置
	if config.Auth.Method == "" {
		config.Auth.Method = "token"
	}
	if config.Auth.Timeout == 0 {
		config.Auth.Timeout = 30 * time.Second
	}
	if config.Auth.CacheTTL == 0 {
		config.Auth.CacheTTL = 1 * time.Hour
	}
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
		"storage.data_dir":         "WINPOWER_EXPORTER_DATA_DIR",
		"server.port":              "WINPOWER_EXPORTER_SERVER_PORT",
		"server.host":              "WINPOWER_EXPORTER_SERVER_HOST",
		"logging.level":            "WINPOWER_EXPORTER_LOG_LEVEL",
		"logging.format":           "WINPOWER_EXPORTER_LOG_FORMAT",
		"winpower.url":             "WINPOWER_EXPORTER_CONSOLE_URL",
		"winpower.username":        "WINPOWER_EXPORTER_USERNAME",
		"winpower.password":        "WINPOWER_EXPORTER_PASSWORD",
		"winpower.skip_ssl_verify": "WINPOWER_EXPORTER_SKIP_SSL_VERIFY",
		"storage.sync_write":       "WINPOWER_EXPORTER_SYNC_WRITE",
		"collector.enabled":        "WINPOWER_EXPORTER_COLLECTOR_ENABLED",
		"metrics.enabled":          "WINPOWER_EXPORTER_METRICS_ENABLED",
		"metrics.path":             "WINPOWER_EXPORTER_METRICS_PATH",
		"energy.enabled":           "WINPOWER_EXPORTER_ENERGY_ENABLED",
		"scheduler.enabled":        "WINPOWER_EXPORTER_SCHEDULER_ENABLED",
		"auth.enabled":             "WINPOWER_EXPORTER_AUTH_ENABLED",
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
			config.Storage.SyncWrite = BoolPtr(boolVal)
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
			config.Collector.Enabled = BoolPtr(boolVal)
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_METRICS_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Metrics.Enabled = BoolPtr(boolVal)
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_METRICS_PATH"); val != "" {
		config.Metrics.Path = val
	}
	if val := os.Getenv("WINPOWER_EXPORTER_ENERGY_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Energy.Enabled = BoolPtr(boolVal)
		}
	}
	if val := os.Getenv("WINPOWER_EXPORTER_SCHEDULER_ENABLED"); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			config.Scheduler.Enabled = BoolPtr(boolVal)
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
