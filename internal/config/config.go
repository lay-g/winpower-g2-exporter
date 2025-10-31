package config

import (
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// ConfigValidator 配置验证接口
// 各功能模块的配置结构体需要实现此接口以提供自定义验证逻辑
type ConfigValidator interface {
	Validate() error
}

// ConfigManager 配置管理器接口
type ConfigManager interface {
	// Load 加载配置
	Load() (*Config, error)

	// Get 获取配置值
	Get(key string) interface{}

	// GetString 获取字符串配置值
	GetString(key string) string

	// GetInt 获取整数配置值
	GetInt(key string) int

	// GetBool 获取布尔配置值
	GetBool(key string) bool

	// GetStringSlice 获取字符串切片配置值
	GetStringSlice(key string) []string

	// Set 设置配置值
	Set(key string, value interface{})

	// IsSet 检查配置是否已设置
	IsSet(key string) bool

	// Validate 验证完整配置
	Validate() error
}

// Config 顶层配置结构体
// 引用各模块的配置结构体
type Config struct {
	// Server 服务器配置
	Server *server.Config `yaml:"server" mapstructure:"server"`

	// WinPower WinPower连接配置
	WinPower *winpower.Config `yaml:"winpower" mapstructure:"winpower"`

	// Storage 存储配置
	Storage *storage.Config `yaml:"storage" mapstructure:"storage"`

	// Scheduler 调度器配置
	Scheduler *scheduler.Config `yaml:"scheduler" mapstructure:"scheduler"`

	// Logging 日志配置
	Logging *log.Config `yaml:"logging" mapstructure:"logging"`
}

// Validate 验证完整配置
func (c *Config) Validate() error {
	// 验证各个模块的配置（跳过 nil 配置）
	if c.Server != nil {
		if err := c.Server.Validate(); err != nil {
			return &ConfigError{
				Message: "server validation failed",
				Err:     err,
			}
		}
	}

	if c.WinPower != nil {
		if err := c.WinPower.Validate(); err != nil {
			return &ConfigError{
				Message: "winpower validation failed",
				Err:     err,
			}
		}
	}

	if c.Storage != nil {
		if err := c.Storage.Validate(); err != nil {
			return &ConfigError{
				Message: "storage validation failed",
				Err:     err,
			}
		}
	}

	if c.Scheduler != nil {
		if err := c.Scheduler.Validate(); err != nil {
			return &ConfigError{
				Message: "scheduler validation failed",
				Err:     err,
			}
		}
	}

	if c.Logging != nil {
		if err := c.Logging.Validate(); err != nil {
			return &ConfigError{
				Message: "logging validation failed",
				Err:     err,
			}
		}
	}

	return nil
}
