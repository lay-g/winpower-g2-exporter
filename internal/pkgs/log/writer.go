package log

import (
	"io"
	"os"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// WriterCloser 是一个可以被关闭的 WriteSyncer
type WriterCloser struct {
	zapcore.WriteSyncer
	closers []io.Closer
}

// Close 关闭所有相关的资源
func (w *WriterCloser) Close() error {
	var lastErr error
	for _, closer := range w.closers {
		if err := closer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// BuildWriter 根据配置构建写入器
func BuildWriter(config *Config) (*WriterCloser, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	var writers []zapcore.WriteSyncer
	var closers []io.Closer

	switch config.Output {
	case "stdout":
		writers = append(writers, zapcore.AddSync(os.Stdout))
	case "stderr":
		writers = append(writers, zapcore.AddSync(os.Stderr))
	case "file":
		if config.FilePath != "" {
			fileWriter := buildFileWriter(config)
			writers = append(writers, zapcore.AddSync(fileWriter))
			closers = append(closers, fileWriter)
		}
	case "both":
		writers = append(writers, zapcore.AddSync(os.Stdout))
		if config.FilePath != "" {
			fileWriter := buildFileWriter(config)
			writers = append(writers, zapcore.AddSync(fileWriter))
			closers = append(closers, fileWriter)
		}
	default:
		// 默认输出到 stdout
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	// 如果没有配置任何输出，默认使用 stdout
	if len(writers) == 0 {
		writers = append(writers, zapcore.AddSync(os.Stdout))
	}

	return &WriterCloser{
		WriteSyncer: zapcore.NewMultiWriteSyncer(writers...),
		closers:     closers,
	}, nil
}

// buildFileWriter 构建文件写入器
func buildFileWriter(config *Config) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxSize,    // MB
		MaxAge:     config.MaxAge,     // days
		MaxBackups: config.MaxBackups, // files
		Compress:   config.Compress,   // 是否压缩
		LocalTime:  true,              // 使用本地时间
	}
}

// BuildStdoutWriter 构建 stdout 写入器
func BuildStdoutWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}

// BuildStderrWriter 构建 stderr 写入器
func BuildStderrWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stderr)
}

// BuildFileWriterWithRotation 构建带轮转的文件写入器
func BuildFileWriterWithRotation(filePath string, maxSize, maxAge, maxBackups int, compress bool) io.WriteCloser {
	return &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   compress,
		LocalTime:  true,
	}
}
