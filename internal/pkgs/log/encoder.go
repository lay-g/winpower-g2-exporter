package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// buildZapConfig 构建 zap 配置
func buildZapConfig(config *Config) (*zap.Config, error) {
	zapConfig := zap.NewProductionConfig()

	// 设置日志级别
	level, err := parseLevel(config.Level)
	if err != nil {
		return nil, err
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// 设置编码器配置
	if config.Format == ConsoleFormat {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig = buildConsoleEncoderConfig(config)
	} else {
		zapConfig.Encoding = "json"
		zapConfig.EncoderConfig = buildJSONEncoderConfig(config)
	}

	// 设置输出路径
	switch config.Output {
	case StdoutOutput:
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	case StderrOutput:
		zapConfig.OutputPaths = []string{"stderr"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	case FileOutput:
		zapConfig.OutputPaths = []string{config.Filename}
		zapConfig.ErrorOutputPaths = []string{config.Filename}
	case BothOutput:
		zapConfig.OutputPaths = []string{"stdout", config.Filename}
		zapConfig.ErrorOutputPaths = []string{"stderr", config.Filename}
	}

	// 设置调用者信息
	zapConfig.DisableCaller = !config.EnableCaller

	// 设置堆栈跟踪
	zapConfig.DisableStacktrace = config.Level != DebugLevel && config.Level != ErrorLevel

	return &zapConfig, nil
}

// parseLevel 解析日志级别
func parseLevel(level Level) (zapcore.Level, error) {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel, nil
	case InfoLevel:
		return zapcore.InfoLevel, nil
	case WarnLevel:
		return zapcore.WarnLevel, nil
	case ErrorLevel:
		return zapcore.ErrorLevel, nil
	case FatalLevel:
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, &ErrInvalidLevel{level: string(level)}
	}
}

// ErrInvalidLevel 无效日志级别错误
type ErrInvalidLevel struct {
	level string
}

func (e *ErrInvalidLevel) Error() string {
	return "invalid log level: " + e.level
}

// buildJSONEncoderConfig 构建 JSON 编码器配置
func buildJSONEncoderConfig(config *Config) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel,
		EncodeTime:     encodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 开发模式下增加更多详细信息
	if config.Level == DebugLevel {
		encoderConfig.FunctionKey = "function"
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	}

	return encoderConfig
}

// buildConsoleEncoderConfig 构建 Console 编码器配置
func buildConsoleEncoderConfig(config *Config) zapcore.EncoderConfig {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeConsoleLevel,
		EncodeTime:     encodeConsoleTime,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   encodeConsoleCaller,
	}

	// 开发模式下显示函数名
	if config.Level == DebugLevel {
		encoderConfig.FunctionKey = "F"
	}

	return encoderConfig
}

// encodeLevel 编码日志级别（JSON格式）
func encodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(level.String())
}

// encodeTime 编码时间（JSON格式）
func encodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
}

// encodeConsoleLevel 编码日志级别（Console格式）
func encodeConsoleLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var s string
	switch level {
	case zapcore.DebugLevel:
		s = "DEBUG"
	case zapcore.InfoLevel:
		s = "INFO"
	case zapcore.WarnLevel:
		s = "WARN"
	case zapcore.ErrorLevel:
		s = "ERROR"
	case zapcore.FatalLevel:
		s = "FATAL"
	default:
		s = level.String()
	}
	enc.AppendString(s)
}

// encodeConsoleTime 编码时间（Console格式）
func encodeConsoleTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

// encodeConsoleCaller 编码调用者信息（Console格式）
func encodeConsoleCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	if !caller.Defined {
		enc.AppendString("")
		return
	}
	enc.AppendString(caller.TrimmedPath())
}
