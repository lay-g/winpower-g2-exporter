package log

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	// 全局日志器实例
	globalLogger atomic.Value // *zapLogger
	globalMu     sync.RWMutex
	// 全局配置
	globalConfig atomic.Value // *Config
)

// 全局初始化控制
var (
	globalInitialized int32
)

// Init 初始化全局日志器
func Init(config *Config) error {
	if config == nil {
		config = DefaultConfig()
	}

	logger, err := NewLogger(config)
	if err != nil {
		return fmt.Errorf("failed to create global logger: %w", err)
	}

	globalLogger.Store(logger)
	globalConfig.Store(config.Clone())
	atomic.StoreInt32(&globalInitialized, 1)

	return nil
}

// InitDevelopment 初始化开发环境全局日志器
func InitDevelopment() error {
	config := DevelopmentDefaults()
	return Init(config)
}

// ResetGlobal 重置全局日志器（主要用于测试）
func ResetGlobal() {
	globalLogger.Store((*zapLoggerImpl)(nil))
	globalConfig.Store((*Config)(nil))
	atomic.StoreInt32(&globalInitialized, 0)
}

// Default 获取默认的全局日志器
func Default() Logger {
	if atomic.LoadInt32(&globalInitialized) == 0 {
		// 如果未初始化，使用默认配置初始化
		if err := Init(DefaultConfig()); err != nil {
			// 如果初始化失败，创建一个最基本的日志器
			logger, _ := NewLogger(DefaultConfig())
			globalLogger.Store(logger)
			atomic.StoreInt32(&globalInitialized, 1)
		}
	}

	if logger := globalLogger.Load(); logger != nil {
		if impl := logger.(*zapLoggerImpl); impl != nil {
			return impl
		}
	}

	// 如果加载失败，创建临时日志器
	logger, _ := NewLogger(DefaultConfig())
	globalLogger.Store(logger)
	return logger
}

// GetGlobalConfig 获取全局配置
func GetGlobalConfig() *Config {
	if config := globalConfig.Load(); config != nil {
		if cfg := config.(*Config); cfg != nil {
			return cfg.Clone()
		}
	}
	return DefaultConfig()
}

// 全局日志函数
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

func GlobalError(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// With 创建带有额外字段的全局日志器
func With(fields ...Field) Logger {
	return Default().With(fields...)
}

// WithContext 创建带有上下文字段的全局日志器
func WithContext(ctx context.Context) Logger {
	return Default().WithContext(ctx)
}

// Sync 同步全局日志器缓冲区
func Sync() error {
	if atomic.LoadInt32(&globalInitialized) == 0 {
		return nil
	}

	if logger := globalLogger.Load(); logger != nil {
		return logger.(Logger).Sync()
	}

	return nil
}

// IsInitialized 检查全局日志器是否已初始化
func IsInitialized() bool {
	return atomic.LoadInt32(&globalInitialized) == 1
}

// Reconfigure 重新配置全局日志器
func Reconfigure(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// 创建新的日志器
	newLogger, err := NewLogger(config)
	if err != nil {
		return fmt.Errorf("failed to create new logger: %w", err)
	}

	// 同步旧的日志器
	if oldLogger := globalLogger.Load(); oldLogger != nil {
		_ = oldLogger.(Logger).Sync()
	}

	// 更新全局日志器和配置
	globalLogger.Store(newLogger)
	globalConfig.Store(config.Clone())

	return nil
}

// GetLevel 获取当前全局日志级别
func GetLevel() Level {
	if config := globalConfig.Load(); config != nil {
		if cfg, ok := config.(*Config); ok && cfg != nil {
			return cfg.Level
		}
	}
	return InfoLevel
}

// SetLevel 设置全局日志级别
func SetLevel(level Level) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	config := DefaultConfig()
	config.Level = level
	return Reconfigure(config)
}

// SetOutput 设置全局日志输出目标
func SetOutput(output Output) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	config := DefaultConfig()
	config.Output = output
	return Reconfigure(config)
}

// SetFormat 设置全局日志格式
func SetFormat(format Format) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	config := DefaultConfig()
	config.Format = format
	return Reconfigure(config)
}

// SetFile 设置全局日志文件配置
func SetFile(filename string, maxSize, maxAge, maxBackups int, compress bool) error {
	globalMu.Lock()
	defer globalMu.Unlock()

	config := DefaultConfig()
	config.SetFile(filename, maxSize, maxAge, maxBackups, compress)
	return Reconfigure(config)
}
