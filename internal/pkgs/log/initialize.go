package log

import (
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLoggerWithWriter 创建带有自定义写入器的日志器
func NewLoggerWithWriter(config *Config, writer io.Writer) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid log config: %w", err)
	}

	// 创建自定义的 zap 核心
	core := zapcore.NewCore(
		createEncoder(config),
		zapcore.AddSync(writer),
		parseLevelToZapLevel(config.Level),
	)

	// 添加调用者信息
	var options []zap.Option
	if config.EnableCaller {
		options = append(options, zap.AddCaller())
	}

	// 添加堆栈跟踪（仅对错误级别及以上）
	if config.Level == DebugLevel || config.Level == ErrorLevel {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	zapLogger := zap.New(core, options...)

	return &zapLoggerImpl{zap: zapLogger}, nil
}

// BuildLoggerWithRotation 创建带轮转功能的日志器
func BuildLoggerWithRotation(config *Config) (Logger, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 确保文件输出配置正确
	if config.Output != FileOutput && config.Output != BothOutput {
		return nil, fmt.Errorf("rotation is only available for file output")
	}

	if config.Filename == "" {
		return nil, fmt.Errorf("filename is required for rotation")
	}

	return NewLogger(config)
}

// createEncoder 创建编码器
func createEncoder(config *Config) zapcore.Encoder {
	var encoderConfig zapcore.EncoderConfig

	if config.Format == ConsoleFormat {
		encoderConfig = buildConsoleEncoderConfig(config)
	} else {
		encoderConfig = buildJSONEncoderConfig(config)
	}

	if config.Format == ConsoleFormat {
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// parseLevelToZapLevel 解析日志级别到zap级别
func parseLevelToZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewDevelopmentLogger 创建开发环境日志器
func NewDevelopmentLogger() (Logger, error) {
	config := DevelopmentDefaults()
	return NewLogger(config)
}

// NewProductionLogger 创建生产环境日志器
func NewProductionLogger() (Logger, error) {
	config := DefaultConfig()
	return NewLogger(config)
}

// NewProductionLoggerWithConfig 创建带自定义配置的生产环境日志器
func NewProductionLoggerWithConfig(filename string, level Level) (Logger, error) {
	config := DefaultConfig()
	config.Level = level
	config.Output = FileOutput
	config.Filename = filename
	return NewLogger(config)
}

// NewConsoleLogger 创建控制台日志器
func NewConsoleLogger(level Level, enableColor bool) (Logger, error) {
	config := &Config{
		Level:        level,
		Format:       ConsoleFormat,
		Output:       StdoutOutput,
		EnableColor:  enableColor,
		EnableCaller: level == DebugLevel,
	}
	return NewLogger(config)
}

// NewFileLogger 创建文件日志器
func NewFileLogger(filename string, level Level, maxSize, maxAge, maxBackups int, compress bool) (Logger, error) {
	config := &Config{
		Level:      level,
		Format:     JSONFormat,
		Output:     FileOutput,
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   compress,
	}
	return NewLogger(config)
}

// NewMultiOutputLogger 创建多输出日志器
func NewMultiOutputLogger(filename string, level Level, maxSize, maxAge, maxBackups int, compress bool) (Logger, error) {
	config := &Config{
		Level:      level,
		Format:     JSONFormat,
		Output:     BothOutput,
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   compress,
	}
	return NewLogger(config)
}

// CreateLoggerForTesting 创建测试专用日志器
func CreateLoggerForTesting() Logger {
	return NewTestLogger()
}

// CreateNoopLogger 创建无操作日志器
func CreateNoopLogger() Logger {
	return NewNoopLogger()
}

// InitializeFromEnv 从环境变量初始化日志器
func InitializeFromEnv() error {
	config := DefaultConfig()

	// 从环境变量读取配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = Level(level)
	}

	if format := os.Getenv("LOG_FORMAT"); format != "" {
		config.Format = Format(format)
	}

	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = Output(output)
	}

	if filename := os.Getenv("LOG_FILE"); filename != "" {
		config.Filename = filename
		config.Output = FileOutput
	}

	// 初始化全局日志器
	return Init(config)
}
