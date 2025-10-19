package log

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEndToEndLogging(t *testing.T) {
	// 使用测试日志器进行端到端测试
	logger := NewTestLogger()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithUserID(ctx, "user-456")
	ctx = WithAction(ctx, "test-action")

	logger.WithContext(ctx).Info("test message",
		String("component", "test"),
		Int("retry_count", 3),
		Bool("success", true),
	)

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != InfoLevel {
		t.Errorf("Expected level %s, got %s", InfoLevel, entry.Level)
	}

	// 验证上下文字段
	fields := entry.Fields
	expectedFields := []string{"request_id", "user_id", "action", "component", "retry_count", "success"}

	for _, expectedField := range expectedFields {
		found := false
		for _, field := range fields {
			if field.Key == expectedField {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected field '%s' not found in log entry", expectedField)
		}
	}
}

func TestFileOutputIntegration(t *testing.T) {
	tempFile := t.TempDir() + "/test.log"

	config := &Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		Output:     FileOutput,
		Filename:   tempFile,
		MaxSize:    1,
		MaxAge:     1,
		MaxBackups: 1,
		Compress:   false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("test message 1")
	logger.Error("test error", Error(&integrationTestError{msg: "test error message"}))

	// 等待文件写入
	time.Sleep(100 * time.Millisecond)

	// 读取文件内容
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "test message 1") {
		t.Errorf("Log file does not contain expected message 'test message 1'")
	}
	if !strings.Contains(contentStr, "test error message") {
		t.Errorf("Log file does not contain expected error message")
	}

	// 验证JSON格式
	lines := strings.Split(strings.TrimSpace(contentStr), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var logEntry map[string]interface{}
		if err := json.Unmarshal([]byte(line), &logEntry); err != nil {
			t.Errorf("Failed to parse JSON log entry: %v", err)
		}

		if logEntry["message"] == nil {
			t.Error("Log entry missing message field")
		}
		if logEntry["level"] == nil {
			t.Error("Log entry missing level field")
		}
		if logEntry["timestamp"] == nil {
			t.Error("Log entry missing timestamp field")
		}
	}
}

func TestConsoleOutputIntegration(t *testing.T) {
	var buf bytes.Buffer

	config := &Config{
		Level:        DebugLevel,
		Format:       ConsoleFormat,
		Output:       StdoutOutput,
		EnableColor:  false,
		EnableCaller: true,
	}

	// 重置标准输出到缓冲区
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, err := NewLogger(config)
	if err != nil {
		os.Stdout = originalStdout
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Debug("debug message")
	logger.Info("info message")

	// 恢复标准输出
	_ = w.Close()
	os.Stdout = originalStdout

	// 读取捕获的输出
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "debug message") {
		t.Error("Console output does not contain debug message")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Console output does not contain info message")
	}
	if !strings.Contains(output, "DEBUG") {
		t.Error("Console output does not contain DEBUG level")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Console output does not contain INFO level")
	}
}

func TestBothOutputIntegration(t *testing.T) {
	tempFile := t.TempDir() + "/test.log"
	var buf bytes.Buffer

	config := &Config{
		Level:    InfoLevel,
		Format:   JSONFormat,
		Output:   BothOutput,
		Filename: tempFile,
	}

	// 重置标准输出到缓冲区
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, err := NewLogger(config)
	if err != nil {
		os.Stdout = originalStdout
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("both output test")

	// 恢复标准输出
	_ = w.Close()
	os.Stdout = originalStdout

	// 读取标准输出
	_, _ = buf.ReadFrom(r)
	stdoutContent := buf.String()

	// 读取文件内容
	fileContent, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// 验证两个输出都包含消息
	if !strings.Contains(stdoutContent, "both output test") {
		t.Error("Stdout does not contain expected message")
	}
	if !strings.Contains(string(fileContent), "both output test") {
		t.Error("File does not contain expected message")
	}
}

func TestGlobalLoggerIntegration(t *testing.T) {
	// 重置全局状态
	ResetGlobal()
	defer ResetGlobal()

	// 初始化全局日志器
	config := &Config{
		Level:  InfoLevel,
		Format: JSONFormat,
		Output: StdoutOutput,
	}

	err := Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize global logger: %v", err)
	}

	// 使用全局日志函数
	Info("global info message")
	GlobalError("global error message", String("error_type", "test"))

	// 验证日志器已初始化
	if !IsInitialized() {
		t.Error("Global logger should be initialized")
	}

	// 测试重新配置
	newConfig := &Config{
		Level:  DebugLevel,
		Format: ConsoleFormat,
		Output: StdoutOutput,
	}

	err = Reconfigure(newConfig)
	if err != nil {
		t.Errorf("Failed to reconfigure global logger: %v", err)
	}

	if GetLevel() != DebugLevel {
		t.Errorf("Expected global level to be %s, got %s", DebugLevel, GetLevel())
	}
}

func TestContextLoggingIntegration(t *testing.T) {
	logger := NewTestLogger()

	// 创建带有多个上下文值的context
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-12345")
	ctx = WithTraceID(ctx, "trace-67890")
	ctx = WithUserID(ctx, "user-abc123")
	ctx = WithService(ctx, "test-service")
	ctx = WithAction(ctx, "integration-test")
	ctx = WithDuration(ctx, 150*time.Millisecond)

	// 将日志器添加到上下文中
	ctx = WithLogger(ctx, logger)

	// 测试全局上下文日志函数
	InfoWithContext(ctx, "context test message",
		String("component", "integration"),
		Int("attempt", 1),
	)

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "context test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	fields := entry.Fields

	// 验证所有上下文字段都被提取
	expectedContextFields := []string{
		"request_id", "trace_id", "user_id", "service", "action", "duration",
	}

	for _, expectedField := range expectedContextFields {
		found := false
		for _, field := range fields {
			if field.Key == expectedField {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected context field '%s' not found", expectedField)
		}
	}
}

func TestErrorHandlingIntegration(t *testing.T) {
	logger := NewTestLogger()

	// 测试各种错误情况
	testErr := &integrationTestError{msg: "integration test error"}

	logger.Error("error with error field", Error(testErr))
	logger.Warn("warning without error")
	logger.Info("info with nil error", Error(nil))

	AssertLogCount(t, logger, 3)

	entries := logger.Entries()
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// 验证错误日志包含错误字段
	errorEntry := entries[0]
	hasErrorField := false
	for _, field := range errorEntry.Fields {
		if field.Key == "error" {
			hasErrorField = true
			break
		}
	}
	if !hasErrorField {
		t.Error("Error entry should have error field")
	}
}

func TestPerformanceIntegration(t *testing.T) {
	logger := NewTestLogger()

	// 性能测试：大量日志写入
	messageCount := 1000
	start := time.Now()

	for i := 0; i < messageCount; i++ {
		logger.Info("performance test message",
			Int("iteration", i),
			String("component", "performance"),
		)
	}

	duration := time.Since(start)
	entriesPerSecond := float64(messageCount) / duration.Seconds()

	t.Logf("Logged %d entries in %v (%.2f entries/sec)",
		messageCount, duration, entriesPerSecond)

	AssertLogCount(t, logger, messageCount)

	// 基本性能断言（根据实际情况调整阈值）
	if entriesPerSecond < 1000 {
		t.Errorf("Performance below threshold: %.2f entries/sec (expected >= 1000)", entriesPerSecond)
	}
}

func TestLogRotationIntegration(t *testing.T) {
	tempFile := t.TempDir() + "/rotation-test.log"

	config := &Config{
		Level:      InfoLevel,
		Format:     JSONFormat,
		Output:     FileOutput,
		Filename:   tempFile,
		MaxSize:    1, // 1MB - 很小以便测试轮转
		MaxAge:     1,
		MaxBackups: 2,
		Compress:   false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer func() { _ = logger.Sync() }()

	// 写入大量日志以触发轮转
	largeMessage := strings.Repeat("This is a large message to trigger log rotation. ", 100)

	for i := 0; i < 100; i++ {
		logger.Info(largeMessage, Int("iteration", i))
	}

	// 验证原始文件存在
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Original log file should exist")
	}

	// 轮转是异步的，可能需要等待
	time.Sleep(100 * time.Millisecond)
}

func TestConcurrentLoggingIntegration(t *testing.T) {
	logger := NewTestLogger()

	numGoroutines := 10
	messagesPerGoroutine := 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("concurrent message",
					Int("goroutine_id", goroutineID),
					Int("message_id", j),
				)
			}
		}(i)
	}

	wg.Wait()

	expectedTotal := numGoroutines * messagesPerGoroutine
	AssertLogCount(t, logger, expectedTotal)

	// 验证没有重复或丢失的日志
	entries := logger.Entries()
	messageCounts := make(map[string]int)

	for _, entry := range entries {
		key := entry.Message
		messageCounts[key]++
	}

	if len(messageCounts) != 1 {
		t.Errorf("Expected 1 unique message, got %d", len(messageCounts))
	}

	if messageCounts["concurrent message"] != expectedTotal {
		t.Errorf("Expected %d concurrent messages, got %d", expectedTotal, messageCounts["concurrent message"])
	}
}

// 测试用的错误类型
type integrationTestError struct {
	msg string
}

func (e *integrationTestError) Error() string {
	return e.msg
}
