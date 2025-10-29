package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BuildEncoder 根据配置构建编码器
func BuildEncoder(config *Config) zapcore.Encoder {
	if err := config.Validate(); err != nil {
		// 如果配置无效，使用默认配置
		config = DefaultConfig()
	}

	encoderConfig := buildEncoderConfig(config)

	switch config.Format {
	case "json":
		return zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		return zapcore.NewConsoleEncoder(encoderConfig)
	default:
		// 默认使用 console 格式
		return zapcore.NewConsoleEncoder(encoderConfig)
	}
}

// buildEncoderConfig 构建编码器配置
func buildEncoderConfig(config *Config) zapcore.EncoderConfig {
	var encoderConfig zapcore.EncoderConfig

	if config.Development {
		// 开发环境使用彩色输出和更详细的格式
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	} else {
		// 生产环境使用标准格式
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	return encoderConfig
}

// BuildJSONEncoder 构建 JSON 格式编码器
func BuildJSONEncoder(development bool) zapcore.Encoder {
	encoderConfig := buildEncoderConfigByMode(development)
	return zapcore.NewJSONEncoder(encoderConfig)
}

// BuildConsoleEncoder 构建 Console 格式编码器
func BuildConsoleEncoder(development bool) zapcore.Encoder {
	encoderConfig := buildEncoderConfigByMode(development)
	return zapcore.NewConsoleEncoder(encoderConfig)
}

// buildEncoderConfigByMode 根据开发模式构建编码器配置
func buildEncoderConfigByMode(development bool) zapcore.EncoderConfig {
	if development {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		return encoderConfig
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return encoderConfig
}

// BuildProductionEncoder 构建生产环境编码器（JSON格式）
func BuildProductionEncoder() zapcore.Encoder {
	return BuildJSONEncoder(false)
}

// BuildDevelopmentEncoder 构建开发环境编码器（Console格式，带颜色）
func BuildDevelopmentEncoder() zapcore.Encoder {
	return BuildConsoleEncoder(true)
}
