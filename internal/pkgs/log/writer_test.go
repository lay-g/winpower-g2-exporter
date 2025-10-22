package log

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestNewRotatingFileWriter(t *testing.T) {
	tempFile := t.TempDir() + "/test.log"
	writer := NewRotatingFileWriter(tempFile, 1, 1, 1, true)

	if writer == nil {
		t.Error("Expected rotating file writer to be created")
	}

	// 写入一些数据
	data := []byte("test message\n")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Expected no error writing to file, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// 验证文件存在
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Expected log file to exist")
	}
}

func TestMultiWriter(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	writer := NewMultiWriter(&buf1, &buf2)

	data := []byte("test message")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Expected no error writing to multi writer, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	if buf1.String() != "test message" {
		t.Errorf("Expected buffer 1 to contain 'test message', got '%s'", buf1.String())
	}

	if buf2.String() != "test message" {
		t.Errorf("Expected buffer 2 to contain 'test message', got '%s'", buf2.String())
	}
}

func TestMultiWriterAddRemove(t *testing.T) {
	var buf1, buf2, buf3 bytes.Buffer
	writer := NewMultiWriter(&buf1, &buf2)

	// 初始写入
	_, _ = writer.Write([]byte("message 1"))

	// 添加新的写入器
	writer.AddWriter(&buf3)
	_, _ = writer.Write([]byte("message 2"))

	// 移除一个写入器
	writer.RemoveWriter(&buf1)
	_, _ = writer.Write([]byte("message 3"))

	expectedBuf1 := "message 1message 2"
	expectedBuf2 := "message 1message 2message 3"
	expectedBuf3 := "message 2message 3"

	if buf1.String() != expectedBuf1 {
		t.Errorf("Expected buffer 1 to contain '%s', got '%s'", expectedBuf1, buf1.String())
	}

	if buf2.String() != expectedBuf2 {
		t.Errorf("Expected buffer 2 to contain '%s', got '%s'", expectedBuf2, buf2.String())
	}

	if buf3.String() != expectedBuf3 {
		t.Errorf("Expected buffer 3 to contain '%s', got '%s'", expectedBuf3, buf3.String())
	}
}

func TestMultiWriterWithNil(t *testing.T) {
	var buf bytes.Buffer
	writer := NewMultiWriter(&buf, nil, &buf)

	data := []byte("test message")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Expected no error with nil writer, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// 应该写入到非nil的写入器两次
	expected := "test messagetest message"
	if buf.String() != expected {
		t.Errorf("Expected buffer to contain '%s', got '%s'", expected, buf.String())
	}
}

func TestSafeWriter(t *testing.T) {
	var buf bytes.Buffer
	errorCalled := false

	onError := func(err error) {
		errorCalled = true
	}

	writer := NewSafeWriter(&buf, onError)

	data := []byte("test message")
	n, err := writer.Write(data)
	if err != nil {
		t.Errorf("Expected no error writing to safe writer, got %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	if buf.String() != "test message" {
		t.Errorf("Expected buffer to contain 'test message', got '%s'", buf.String())
	}

	if errorCalled {
		t.Error("Error handler should not be called on successful write")
	}
}

func TestSafeWriterWithError(t *testing.T) {
	// 创建一个会返回错误的写入器
	errorWriter := &errorWriter{err: &writerTestError{msg: "write error"}}
	errorCalled := false

	onError := func(err error) {
		errorCalled = true
	}

	writer := NewSafeWriter(errorWriter, onError)

	data := []byte("test message")
	_, err := writer.Write(data)
	if err == nil {
		t.Error("Expected error from safe writer")
	}

	if !errorCalled {
		t.Error("Error handler should be called on write error")
	}
}

func TestSafeWriterSetGet(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	writer := NewSafeWriter(&buf1, nil)

	if writer.GetWriter() != &buf1 {
		t.Error("Expected GetWriter() to return original writer")
	}

	writer.SetWriter(&buf2)
	if writer.GetWriter() != &buf2 {
		t.Error("Expected GetWriter() to return new writer after SetWriter()")
	}
}

func TestCreateWriter(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "stdout output",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: StdoutOutput,
			},
			wantErr: false,
		},
		{
			name: "stderr output",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: StderrOutput,
			},
			wantErr: false,
		},
		{
			name: "file output",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   FileOutput,
				Filename: t.TempDir() + "/test.log",
			},
			wantErr: false,
		},
		{
			name: "both output",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   BothOutput,
				Filename: t.TempDir() + "/test.log",
			},
			wantErr: false,
		},
		{
			name: "file output without filename",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: FileOutput,
			},
			wantErr: true,
		},
		{
			name: "both output without filename",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: BothOutput,
			},
			wantErr: true,
		},
		{
			name: "invalid config",
			config: &Config{
				Level:  InfoLevel,
				Format: JSONFormat,
				Output: Output("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer, err := CreateWriter(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateWriter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && writer == nil {
				t.Error("CreateWriter() returned nil writer")
			}
		})
	}
}

func TestCreateWriterBothOutput(t *testing.T) {
	withSilentZap(t, func() {
		tempFile := t.TempDir() + "/test.log"
		config := &Config{
			Level:    InfoLevel,
			Format:   JSONFormat,
			Output:   BothOutput,
			Filename: tempFile,
		}

		writer, err := CreateWriter(config)
		if err != nil {
			t.Fatalf("Expected no error creating both output writer, got %v", err)
		}

		data := []byte("test message\n")
		_, err = writer.Write(data)
		if err != nil {
			t.Errorf("Expected no error writing to both output writer, got %v", err)
		}

		// 验证文件被写入
		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Errorf("Expected no error reading log file, got %v", err)
		}

		if string(content) != "test message\n" {
			t.Errorf("Expected file content to be 'test message\\n', got '%s'", string(content))
		}
	})
}

func TestMultiWriterConcurrency(t *testing.T) {
	var buf bytes.Buffer
	writer := NewMultiWriter(&buf)

	var wg sync.WaitGroup
	messageCount := 100

	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			message := fmt.Sprintf("message %d\n", id)
			_, _ = writer.Write([]byte(message))
		}(i)
	}

	wg.Wait()

	content := buf.String()
	lines := strings.Split(content, "\n")

	// 移除最后的空行
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) != messageCount {
		t.Errorf("Expected %d lines, got %d", messageCount, len(lines))
	}
}

func TestMultiWriterClose(t *testing.T) {
	tempFile := t.TempDir() + "/test.log"
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	writer := NewMultiWriter(file)
	err = writer.Close()
	if err != nil {
		t.Errorf("Expected no error closing multi writer, got %v", err)
	}

	// 验证文件已关闭
	_, err = file.WriteString("test")
	if err == nil {
		t.Error("Expected error writing to closed file")
	}
}

// 测试用的错误写入器
type errorWriter struct {
	err error
}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, ew.err
}

// 测试用的错误类型
type writerTestError struct {
	msg string
}

func (e *writerTestError) Error() string {
	return e.msg
}

func TestSafeWriterNilWriter(t *testing.T) {
	var lastErr error
	errorCalled := false

	onError := func(err error) {
		lastErr = err
		errorCalled = true
	}

	writer := NewSafeWriter(nil, onError)

	data := []byte("test message")
	_, err := writer.Write(data)
	if err == nil {
		t.Error("Expected error writing to nil writer")
	}

	if !errorCalled {
		t.Error("Error handler should be called for nil writer")
	}

	if lastErr == nil {
		t.Error("Expected last error to be set for nil writer")
	}
}
