package log

import (
	"fmt"
	"io"
	"os"
	"sync"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// NewRotatingFileWriter 创建带轮转的文件写入器
func NewRotatingFileWriter(filename string, maxSize, maxAge, maxBackups int, compress bool) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackups,
		Compress:   compress,
	}
}

// MultiWriter 多目标写入器，支持同时写入多个目标
type MultiWriter struct {
	writers []io.Writer
	mu      sync.RWMutex
}

// NewMultiWriter 创建多目标写入器
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write 实现io.Writer接口
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	// 第一个写入器的长度作为返回值
	var firstN int
	var firstErr error

	for i, writer := range mw.writers {
		if writer == nil {
			continue
		}

		n, err := writer.Write(p)
		if i == 0 {
			firstN = n
			firstErr = err
		}

		// 如果有任何一个写入器出错，记录但继续写入其他写入器
		if err != nil {
			// 这里可以添加错误日志，但为了避免循环依赖，暂时不做处理
			continue
		}
	}

	return firstN, firstErr
}

// AddWriter 添加写入器
func (mw *MultiWriter) AddWriter(writer io.Writer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	mw.writers = append(mw.writers, writer)
}

// RemoveWriter 移除写入器
func (mw *MultiWriter) RemoveWriter(writer io.Writer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	for i, w := range mw.writers {
		if w == writer {
			mw.writers = append(mw.writers[:i], mw.writers[i+1:]...)
			break
		}
	}
}

// Close 关闭所有写入器（如果支持关闭操作）
func (mw *MultiWriter) Close() error {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	var lastErr error
	for _, writer := range mw.writers {
		if closer, ok := writer.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}

// CreateWriter 根据配置创建写入器
func CreateWriter(config *Config) (io.Writer, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	switch config.Output {
	case StdoutOutput:
		return os.Stdout, nil
	case StderrOutput:
		return os.Stderr, nil
	case FileOutput:
		return NewRotatingFileWriter(
			config.Filename,
			config.MaxSize,
			config.MaxAge,
			config.MaxBackups,
			config.Compress,
		), nil
	case BothOutput:
		// 创建文件写入器
		fileWriter := NewRotatingFileWriter(
			config.Filename,
			config.MaxSize,
			config.MaxAge,
			config.MaxBackups,
			config.Compress,
		)

		// 创建多目标写入器，同时输出到stdout和文件
		return NewMultiWriter(os.Stdout, fileWriter), nil
	default:
		return nil, fmt.Errorf("unsupported output type: %s", config.Output)
	}
}

// SafeWriter 安全写入器，用于处理写入错误
type SafeWriter struct {
	writer  io.Writer
	mu      sync.RWMutex
	onError func(error)
}

// NewSafeWriter 创建安全写入器
func NewSafeWriter(writer io.Writer, onError func(error)) *SafeWriter {
	return &SafeWriter{
		writer:  writer,
		onError: onError,
	}
}

// Write 实现io.Writer接口
func (sw *SafeWriter) Write(p []byte) (n int, err error) {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	if sw.writer == nil {
		err = fmt.Errorf("writer is nil")
		if sw.onError != nil {
			sw.onError(err)
		}
		return 0, err
	}

	n, err = sw.writer.Write(p)
	if err != nil && sw.onError != nil {
		sw.onError(err)
	}

	return n, err
}

// SetWriter 设置底层写入器
func (sw *SafeWriter) SetWriter(writer io.Writer) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.writer = writer
}

// GetWriter 获取底层写入器
func (sw *SafeWriter) GetWriter() io.Writer {
	sw.mu.RLock()
	defer sw.mu.RUnlock()
	return sw.writer
}
