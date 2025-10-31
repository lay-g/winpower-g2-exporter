package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Loader 配置加载器
type Loader struct {
	viper       *viper.Viper
	flags       *pflag.FlagSet
	searchPaths []string
}

// NewLoader 创建新的配置加载器
func NewLoader() *Loader {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	// 设置环境变量前缀
	v.SetEnvPrefix("WINPOWER_EXPORTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 定义搜索路径
	searchPaths := []string{
		".",        // 工作目录
		"./config", // 工作目录/config
		filepath.Join(os.Getenv("HOME"), "config/winpower-exporter"), // 用户配置目录
		"/etc/winpower-exporter",                                     // 系统配置目录
	}

	for _, path := range searchPaths {
		v.AddConfigPath(path)
	}

	return &Loader{
		viper:       v,
		searchPaths: searchPaths,
	}
}

// Load 加载配置
func (l *Loader) Load() (*Config, error) {
	// 设置默认值
	l.setDefaults()

	// 绑定命令行参数
	if err := l.bindFlags(); err != nil {
		return nil, &ConfigError{
			Message: "failed to bind flags",
			Err:     err,
		}
	}

	// 读取配置文件
	if err := l.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, &ConfigError{
				Message: "failed to read config file",
				Err:     err,
			}
		}
		// 配置文件不存在是允许的，使用默认配置和环境变量
	}

	// 解析到配置结构体
	var config Config

	// Viper's Unmarshal doesn't properly merge defaults for nested structs,
	// so we need to initialize each module config with defaults first
	config.Server = &server.Config{}
	config.WinPower = &winpower.Config{}
	config.Storage = &storage.Config{}
	config.Scheduler = &scheduler.Config{}
	config.Logging = &log.Config{}

	// Use Unmarshal with custom decode hooks for time.Duration
	opts := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	)

	if err := l.viper.Unmarshal(&config, opts); err != nil {
		return nil, &ConfigError{
			Message: "failed to unmarshal config",
			Err:     err,
		}
	}

	// Viper may not have filled in all fields from defaults, so fill them manually
	// This ensures all duration fields get their default values if not specified
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = l.viper.GetDuration("server.read_timeout")
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = l.viper.GetDuration("server.write_timeout")
	}
	if config.Server.IdleTimeout == 0 {
		config.Server.IdleTimeout = l.viper.GetDuration("server.idle_timeout")
	}
	if config.Server.ShutdownTimeout == 0 {
		config.Server.ShutdownTimeout = l.viper.GetDuration("server.shutdown_timeout")
	}

	if config.WinPower.Timeout == 0 {
		config.WinPower.Timeout = l.viper.GetDuration("winpower.timeout")
	}
	if config.WinPower.RefreshThreshold == 0 {
		config.WinPower.RefreshThreshold = l.viper.GetDuration("winpower.refresh_threshold")
	}

	if config.Scheduler.CollectionInterval == 0 {
		config.Scheduler.CollectionInterval = l.viper.GetDuration("scheduler.collection_interval")
	}
	if config.Scheduler.GracefulShutdownTimeout == 0 {
		config.Scheduler.GracefulShutdownTimeout = l.viper.GetDuration("scheduler.graceful_shutdown_timeout")
	}

	return &config, nil
}

// Get 获取配置值
func (l *Loader) Get(key string) interface{} {
	return l.viper.Get(key)
}

// GetString 获取字符串配置值
func (l *Loader) GetString(key string) string {
	return l.viper.GetString(key)
}

// GetInt 获取整数配置值
func (l *Loader) GetInt(key string) int {
	return l.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (l *Loader) GetBool(key string) bool {
	return l.viper.GetBool(key)
}

// GetStringSlice 获取字符串切片配置值
func (l *Loader) GetStringSlice(key string) []string {
	return l.viper.GetStringSlice(key)
}

// Set 设置配置值
func (l *Loader) Set(key string, value interface{}) {
	l.viper.Set(key, value)
}

// IsSet 检查配置是否已设置
func (l *Loader) IsSet(key string) bool {
	return l.viper.IsSet(key)
}

// Validate 验证完整配置
func (l *Loader) Validate() error {
	config, err := l.Load()
	if err != nil {
		return err
	}
	return config.Validate()
}
