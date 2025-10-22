package log

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Level 定义日志级别类型
type Level string

// 支持的日志级别常量
const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	FatalLevel Level = "fatal"
)

// Format 定义日志输出格式
type Format string

// 支持的日志格式常量
const (
	JSONFormat    Format = "json"
	ConsoleFormat Format = "console"
)

// Output 定义日志输出目标
type Output string

// 支持的输出目标常量
const (
	StdoutOutput Output = "stdout"
	StderrOutput Output = "stderr"
	FileOutput   Output = "file"
	BothOutput   Output = "both"
)

// Config 日志配置结构
type Config struct {
	Level        Level  `json:"level" yaml:"level"`                           // 日志级别
	Format       Format `json:"format" yaml:"format"`                         // 日志格式
	Output       Output `json:"output" yaml:"output"`                         // 输出目标
	Filename     string `json:"filename,omitempty" yaml:"filename"`           // 日志文件路径
	MaxSize      int    `json:"max_size,omitempty" yaml:"max_size"`           // 单个日志文件最大大小（MB）
	MaxAge       int    `json:"max_age,omitempty" yaml:"max_age"`             // 日志文件保留天数
	MaxBackups   int    `json:"max_backups,omitempty" yaml:"max_backups"`     // 保留的旧日志文件数量
	Compress     bool   `json:"compress,omitempty" yaml:"compress"`           // 是否压缩旧日志文件
	EnableColor  bool   `json:"enable_color,omitempty" yaml:"enable_color"`   // 是否启用颜色输出（仅Console格式）
	EnableCaller bool   `json:"enable_caller,omitempty" yaml:"enable_caller"` // 是否显示调用者信息
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Level:        InfoLevel,
		Format:       JSONFormat,
		Output:       StdoutOutput,
		Filename:     "",
		MaxSize:      100, // 100MB
		MaxAge:       30,  // 30 days
		MaxBackups:   10,  // 10 files
		Compress:     true,
		EnableColor:  false,
		EnableCaller: false,
	}
}

// DevelopmentDefaults 返回开发环境默认配置
func DevelopmentDefaults() *Config {
	return &Config{
		Level:        DebugLevel,
		Format:       ConsoleFormat,
		Output:       StdoutOutput,
		Filename:     "",
		MaxSize:      100,
		MaxAge:       30,
		MaxBackups:   10,
		Compress:     false,
		EnableColor:  true,
		EnableCaller: true,
	}
}

// TestDefaults 返回测试环境默认配置
func TestDefaults() *Config {
	return &Config{
		Level:        ErrorLevel, // 只显示错误级别日志
		Format:       ConsoleFormat,
		Output:       StderrOutput, // 输出到stderr避免与测试输出混合
		Filename:     "",
		MaxSize:      100,
		MaxAge:       30,
		MaxBackups:   10,
		Compress:     false,
		EnableColor:  false, // 测试时禁用颜色
		EnableCaller: false, // 测试时禁用调用者信息
	}
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 验证日志级别
	switch c.Level {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel:
		// 有效级别
	default:
		return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error, fatal", c.Level)
	}

	// 验证日志格式
	switch c.Format {
	case JSONFormat, ConsoleFormat:
		// 有效格式
	default:
		return fmt.Errorf("invalid log format: %s, must be one of: json, console", c.Format)
	}

	// 验证输出目标
	switch c.Output {
	case StdoutOutput, StderrOutput, FileOutput, BothOutput:
		// 有效输出目标
	default:
		return fmt.Errorf("invalid log output: %s, must be one of: stdout, stderr, file, both", c.Output)
	}

	// 验证文件输出配置
	if c.Output == FileOutput || c.Output == BothOutput {
		if c.Filename == "" {
			return fmt.Errorf("filename is required when output is 'file' or 'both'")
		}

		// 确保目录存在
		dir := c.Filename[:strings.LastIndex(c.Filename, "/")]
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory: %w", err)
			}
		}
	}

	// 验证数值参数
	if c.MaxSize <= 0 {
		c.MaxSize = 100
	}
	if c.MaxAge <= 0 {
		c.MaxAge = 30
	}
	if c.MaxBackups <= 0 {
		c.MaxBackups = 10
	}

	// 在非Windows平台上默认启用颜色输出
	if !c.EnableColor && c.Format == ConsoleFormat && runtime.GOOS != "windows" {
		c.EnableColor = true
	}

	return nil
}

// Clone 克隆配置
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	clone := *c
	return &clone
}

// SetLevel 设置日志级别
func (c *Config) SetLevel(level Level) *Config {
	c.Level = level
	return c
}

// SetFormat 设置日志格式
func (c *Config) SetFormat(format Format) *Config {
	c.Format = format
	return c
}

// SetOutput 设置输出目标
func (c *Config) SetOutput(output Output) *Config {
	c.Output = output
	return c
}

// SetFile 设置日志文件配置
func (c *Config) SetFile(filename string, maxSize, maxAge, maxBackups int, compress bool) *Config {
	c.Filename = filename
	c.MaxSize = maxSize
	c.MaxAge = maxAge
	c.MaxBackups = maxBackups
	c.Compress = compress
	return c
}
