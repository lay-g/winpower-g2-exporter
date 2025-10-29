package log

import (
	"fmt"
	"strings"
)

// Config 日志配置结构
type Config struct {
	// Level 日志级别: debug, info, warn, error, fatal
	Level string `json:"level" yaml:"level" mapstructure:"level"`

	// Format 输出格式: json, console
	Format string `json:"format" yaml:"format" mapstructure:"format"`

	// Output 输出目标: stdout, stderr, file, both
	Output string `json:"output" yaml:"output" mapstructure:"output"`

	// FilePath 文件输出路径（当 Output 为 file 或 both 时使用）
	FilePath string `json:"file_path" yaml:"file_path" mapstructure:"file_path"`

	// MaxSize 单个日志文件最大大小（MB）
	MaxSize int `json:"max_size" yaml:"max_size" mapstructure:"max_size"`

	// MaxAge 日志文件最大保留天数
	MaxAge int `json:"max_age" yaml:"max_age" mapstructure:"max_age"`

	// MaxBackups 最大保留的旧日志文件数
	MaxBackups int `json:"max_backups" yaml:"max_backups" mapstructure:"max_backups"`

	// Compress 是否压缩旧日志文件
	Compress bool `json:"compress" yaml:"compress" mapstructure:"compress"`

	// Development 是否为开发模式
	Development bool `json:"development" yaml:"development" mapstructure:"development"`

	// EnableCaller 是否记录调用位置（文件名和行号）
	EnableCaller bool `json:"enable_caller" yaml:"enable_caller" mapstructure:"enable_caller"`

	// EnableStacktrace 是否记录堆栈跟踪
	EnableStacktrace bool `json:"enable_stacktrace" yaml:"enable_stacktrace" mapstructure:"enable_stacktrace"`
}

// DefaultConfig 返回生产环境的默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:            "info",
		Format:           "json",
		Output:           "stdout",
		FilePath:         "",
		MaxSize:          100,  // 100 MB
		MaxAge:           30,   // 30 days
		MaxBackups:       10,   // 10 files
		Compress:         true, // 压缩旧日志
		Development:      false,
		EnableCaller:     false,
		EnableStacktrace: false, // 仅 error 级别及以上启用
	}
}

// DevelopmentDefaults 返回开发环境的默认配置
func DevelopmentDefaults() *Config {
	return &Config{
		Level:            "debug",
		Format:           "console",
		Output:           "stdout",
		FilePath:         "",
		MaxSize:          100,
		MaxAge:           7, // 7 days
		MaxBackups:       3, // 3 files
		Compress:         false,
		Development:      true,
		EnableCaller:     true,
		EnableStacktrace: true, // error 和 fatal 级别启用
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证日志级别
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	level := strings.ToLower(c.Level)
	if !validLevels[level] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error, fatal)", c.Level)
	}
	c.Level = level // 标准化为小写

	// 验证输出格式
	validFormats := map[string]bool{
		"json":    true,
		"console": true,
	}
	format := strings.ToLower(c.Format)
	if !validFormats[format] {
		return fmt.Errorf("invalid log format: %s (must be one of: json, console)", c.Format)
	}
	c.Format = format // 标准化为小写

	// 验证输出目标
	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
		"both":   true,
	}
	output := strings.ToLower(c.Output)
	if !validOutputs[output] {
		return fmt.Errorf("invalid log output: %s (must be one of: stdout, stderr, file, both)", c.Output)
	}
	c.Output = output // 标准化为小写

	// 验证文件输出路径
	if (c.Output == "file" || c.Output == "both") && c.FilePath == "" {
		return fmt.Errorf("file_path is required when output is 'file' or 'both'")
	}

	// 验证文件大小和备份配置
	if c.MaxSize < 0 {
		return fmt.Errorf("max_size must be non-negative")
	}
	if c.MaxAge < 0 {
		return fmt.Errorf("max_age must be non-negative")
	}
	if c.MaxBackups < 0 {
		return fmt.Errorf("max_backups must be non-negative")
	}

	return nil
}
