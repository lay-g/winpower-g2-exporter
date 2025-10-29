package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field 是类型安全的日志字段
type Field = zapcore.Field

// Logger 定义日志接口
type Logger interface {
	// Debug 记录 debug 级别日志
	Debug(msg string, fields ...Field)

	// Info 记录 info 级别日志
	Info(msg string, fields ...Field)

	// Warn 记录 warn 级别日志
	Warn(msg string, fields ...Field)

	// Error 记录 error 级别日志
	Error(msg string, fields ...Field)

	// Fatal 记录 fatal 级别日志并退出程序
	Fatal(msg string, fields ...Field)

	// With 创建一个带有额外字段的子日志器
	With(fields ...Field) Logger

	// WithContext 创建一个从上下文中提取字段的子日志器
	WithContext(ctx context.Context) Logger

	// Sync 刷新缓冲区
	Sync() error

	// Core 返回底层的 zapcore.Core（用于高级用法）
	Core() zapcore.Core
}

// zapLogger 是基于 zap 的 Logger 实现
type zapLogger struct {
	logger *zap.Logger
}

// 确保 zapLogger 实现了 Logger 接口
var _ Logger = (*zapLogger)(nil)

// Debug 实现 Logger.Debug
func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

// Info 实现 Logger.Info
func (l *zapLogger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

// Warn 实现 Logger.Warn
func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

// Error 实现 Logger.Error
func (l *zapLogger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

// Fatal 实现 Logger.Fatal
func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
}

// With 实现 Logger.With
func (l *zapLogger) With(fields ...Field) Logger {
	return &zapLogger{
		logger: l.logger.With(fields...),
	}
}

// WithContext 实现 Logger.WithContext
func (l *zapLogger) WithContext(ctx context.Context) Logger {
	fields := extractContextFields(ctx)
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// Sync 实现 Logger.Sync
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// Core 实现 Logger.Core
func (l *zapLogger) Core() zapcore.Core {
	return l.logger.Core()
}

// 全局日志器
var (
	globalLogger     Logger
	globalLoggerOnce sync.Once
	globalLoggerMu   sync.RWMutex
)

// Init 使用指定配置初始化全局日志器
func Init(config *Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	logger, err := NewLogger(config)
	if err != nil {
		return err
	}

	setGlobalLogger(logger)
	return nil
}

// InitDevelopment 使用开发环境默认配置初始化全局日志器
func InitDevelopment() error {
	return Init(DevelopmentDefaults())
}

// Default 返回全局日志器实例
func Default() Logger {
	globalLoggerMu.RLock()
	logger := globalLogger
	globalLoggerMu.RUnlock()

	if logger == nil {
		// 如果没有初始化，使用默认配置初始化一次
		globalLoggerOnce.Do(func() {
			_ = Init(DefaultConfig())
		})
		globalLoggerMu.RLock()
		logger = globalLogger
		globalLoggerMu.RUnlock()
	}

	return logger
}

// setGlobalLogger 设置全局日志器（内部使用）
func setGlobalLogger(logger Logger) {
	globalLoggerMu.Lock()
	globalLogger = logger
	globalLoggerMu.Unlock()
}

// ResetGlobal 重置全局日志器（仅用于测试）
func ResetGlobal() {
	globalLoggerMu.Lock()
	globalLogger = nil
	globalLoggerMu.Unlock()
	globalLoggerOnce = sync.Once{}
}

// NewLogger 使用指定配置创建新的日志器
func NewLogger(config *Config) (Logger, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 构建 encoder
	encoder := BuildEncoder(config)

	// 构建 writer
	writerCloser, err := BuildWriter(config)
	if err != nil {
		return nil, err
	}

	// 构建 core
	level := parseLevel(config.Level)
	core := zapcore.NewCore(encoder, writerCloser, level)

	// 构建选项
	opts := buildOptions(config)

	// 创建 zap logger
	zapLog := zap.New(core, opts...)

	return &zapLogger{logger: zapLog}, nil
}

// buildOptions 构建 zap 选项
func buildOptions(config *Config) []zap.Option {
	var opts []zap.Option

	if config.EnableCaller {
		opts = append(opts, zap.AddCaller())
		opts = append(opts, zap.AddCallerSkip(1)) // 跳过包装函数
	}

	if config.EnableStacktrace {
		// 在 error 和 fatal 级别启用堆栈跟踪
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	if config.Development {
		opts = append(opts, zap.Development())
	}

	return opts
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// 全局便捷函数

// Debug 使用全局日志器记录 debug 级别日志
func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// Info 使用全局日志器记录 info 级别日志
func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// Warn 使用全局日志器记录 warn 级别日志
func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// Error 使用全局日志器记录 error 级别日志
func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// Fatal 使用全局日志器记录 fatal 级别日志并退出程序
func Fatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}

// DebugContext 使用全局日志器和上下文记录 debug 级别日志
func DebugContext(ctx context.Context, msg string, fields ...Field) {
	Default().WithContext(ctx).Debug(msg, fields...)
}

// InfoContext 使用全局日志器和上下文记录 info 级别日志
func InfoContext(ctx context.Context, msg string, fields ...Field) {
	Default().WithContext(ctx).Info(msg, fields...)
}

// WarnContext 使用全局日志器和上下文记录 warn 级别日志
func WarnContext(ctx context.Context, msg string, fields ...Field) {
	Default().WithContext(ctx).Warn(msg, fields...)
}

// ErrorContext 使用全局日志器和上下文记录 error 级别日志
func ErrorContext(ctx context.Context, msg string, fields ...Field) {
	Default().WithContext(ctx).Error(msg, fields...)
}

// FatalContext 使用全局日志器和上下文记录 fatal 级别日志并退出程序
func FatalContext(ctx context.Context, msg string, fields ...Field) {
	Default().WithContext(ctx).Fatal(msg, fields...)
}

// 字段构造器（直接导出 zap 的字段构造器）

// String 创建一个字符串字段
func String(key, val string) Field {
	return zap.String(key, val)
}

// Int 创建一个整数字段
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int64 创建一个 64 位整数字段
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Float64 创建一个浮点数字段
func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

// Bool 创建一个布尔字段
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Duration 创建一个时间间隔字段
func Duration(key string, val interface{}) Field {
	// 使用 Any 来处理不同类型的 duration
	return zap.Any(key, val)
}

// Time 创建一个时间字段
func Time(key string, val interface{}) Field {
	return zap.Any(key, val)
}

// Err 创建一个错误字段（键名为 "error"）
func Err(err error) Field {
	return zap.Error(err)
}

// Any 创建一个任意类型字段（使用反射，性能较低）
func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}
