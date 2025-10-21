package log

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Field 定义日志字段类型
type Field = zapcore.Field

// Logger 定义日志接口
type Logger interface {
	// 基础日志方法
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	// 创建子日志器
	With(fields ...Field) Logger

	// 上下文日志
	WithContext(ctx context.Context) Logger

	// 同步缓冲区
	Sync() error
}

// zapLoggerImpl 实现 Logger 接口
type zapLoggerImpl struct {
	zap *zap.Logger
}

// NewLogger 创建新的日志器
func NewLogger(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid log config: %w", err)
	}

	zapConfig, err := buildZapConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build zap config: %w", err)
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	return &zapLoggerImpl{zap: zapLogger}, nil
}

// Debug 输出调试级别日志
func (l *zapLoggerImpl) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, l.convertFields(fields)...)
}

// Info 输出信息级别日志
func (l *zapLoggerImpl) Info(msg string, fields ...Field) {
	l.zap.Info(msg, l.convertFields(fields)...)
}

// Warn 输出警告级别日志
func (l *zapLoggerImpl) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, l.convertFields(fields)...)
}

// Error 输出错误级别日志
func (l *zapLoggerImpl) Error(msg string, fields ...Field) {
	l.zap.Error(msg, l.convertFields(fields)...)
}

// Fatal 输出致命错误级别日志
func (l *zapLoggerImpl) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, l.convertFields(fields)...)
}

// With 创建带有额外字段的子日志器
func (l *zapLoggerImpl) With(fields ...Field) Logger {
	return &zapLoggerImpl{zap: l.zap.With(l.convertFields(fields)...)}
}

// WithContext 创建带有上下文字段的日志器
func (l *zapLoggerImpl) WithContext(ctx context.Context) Logger {
	fields := extractContextFields(ctx)
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// Sync 同步缓冲区
func (l *zapLoggerImpl) Sync() error {
	return l.zap.Sync()
}

// convertFields 转换字段类型
func (l *zapLoggerImpl) convertFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}

	zapFields := make([]zap.Field, len(fields))
	copy(zapFields, fields)
	return zapFields
}

// 字段构造函数

// String 创建字符串字段
func String(key, value string) Field {
	return zap.String(key, value)
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Float64 创建浮点数字段
func Float64(key string, value float64) Field {
	return zap.Float64(key, value)
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

// Error 创建错误字段
func Error(err error) Field {
	return zap.Error(err)
}

// Any 创建任意类型字段
func Any(key string, value any) Field {
	return zap.Any(key, value)
}

// Duration 创建时间间隔字段
func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

// Time 创建时间字段
func Time(key string, value time.Time) Field {
	return zap.Time(key, value)
}

// Object 创建对象字段
func Object(key string, value zapcore.ObjectMarshaler) Field {
	return zap.Object(key, value)
}

// Array 创建数组字段
func Array(key string, value zapcore.ArrayMarshaler) Field {
	return zap.Array(key, value)
}

// Namespace 创建命名空间字段
func Namespace(key string) Field {
	return zap.Namespace(key)
}

// Stack 创建堆栈跟踪字段
func Stack(key string) Field {
	return zap.Stack(key)
}

// StackSkip 创建跳过指定层数的堆栈跟踪字段
func StackSkip(key string, skip int) Field {
	return zap.StackSkip(key, skip)
}

// Binary 创建二进制字段
func Binary(key string, value []byte) Field {
	return zap.Binary(key, value)
}

// ByteString 创建字节字符串字段
func ByteString(key string, value []byte) Field {
	return zap.ByteString(key, value)
}

// Complex128 创建128位复数字段
func Complex128(key string, value complex128) Field {
	return zap.Complex128(key, value)
}

// Complex64 创建64位复数字段
func Complex64(key string, value complex64) Field {
	return zap.Complex64(key, value)
}

// Uintptr 创建指针字段
func Uintptr(key string, value uintptr) Field {
	return zap.Uintptr(key, value)
}

// Reflect 创建反射字段
func Reflect(key string, value any) Field {
	return zap.Reflect(key, value)
}
