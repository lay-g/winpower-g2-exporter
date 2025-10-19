package log

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewTestLogger(t *testing.T) {
	logger := NewTestLogger()

	if logger == nil {
		t.Error("Expected test logger to be created")
		return
	}

	if !logger.Empty() {
		t.Error("Expected new test logger to be empty")
	}

	if logger.level != DebugLevel {
		t.Errorf("Expected default level to be %s, got %s", DebugLevel, logger.level)
	}
}

func TestNewTestLoggerWithT(t *testing.T) {
	logger := NewTestLoggerWithT(t)

	if logger == nil {
		t.Error("Expected test logger to be created")
	}

	logger.Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")
}

func TestTestLoggerBasicLogging(t *testing.T) {
	logger := NewTestLogger()

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	AssertLogCount(t, logger, 4)
	AssertLogContains(t, logger, "debug message")
	AssertLogContains(t, logger, "info message")
	AssertLogContains(t, logger, "warn message")
	AssertLogContains(t, logger, "error message")
}

func TestTestLoggerLevelFiltering(t *testing.T) {
	logger := NewTestLogger()
	logger.SetLevel(InfoLevel)

	logger.Debug("debug message") // Should be filtered
	logger.Info("info message")   // Should be included
	logger.Warn("warn message")   // Should be included
	logger.Error("error message") // Should be included

	AssertLogCount(t, logger, 3)
	AssertLogContains(t, logger, "info message")
	AssertLogContains(t, logger, "warn message")
	AssertLogContains(t, logger, "error message")

	// Debug should not be present
	if logger.CountByLevel(DebugLevel) != 0 {
		t.Error("debug level count should be 0")
	}
}

func TestTestLoggerWith(t *testing.T) {
	logger := NewTestLogger()

	childLogger := logger.With(String("component", "test"))
	childLogger.Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}
}

func TestTestLoggerWithContext(t *testing.T) {
	logger := NewTestLogger()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")
}

func TestTestLoggerEntries(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("message 1")
	logger.Error("message 2")

	entries := logger.Entries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	if entries[0].Message != "message 1" {
		t.Errorf("Expected first message to be 'message 1', got '%s'", entries[0].Message)
	}

	if entries[1].Message != "message 2" {
		t.Errorf("Expected second message to be 'message 2', got '%s'", entries[1].Message)
	}

	if entries[1].Level != ErrorLevel {
		t.Errorf("Expected second entry level to be %s, got %s", ErrorLevel, entries[1].Level)
	}
}

func TestTestLoggerEntriesByLevel(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("info message 1")
	logger.Info("info message 2")
	logger.Error("error message")
	logger.Warn("warn message")

	infoEntries := logger.EntriesByLevel(InfoLevel)
	if len(infoEntries) != 2 {
		t.Errorf("Expected 2 info entries, got %d", len(infoEntries))
	}

	errorEntries := logger.EntriesByLevel(ErrorLevel)
	if len(errorEntries) != 1 {
		t.Errorf("Expected 1 error entry, got %d", len(errorEntries))
	}

	warnEntries := logger.EntriesByLevel(WarnLevel)
	if len(warnEntries) != 1 {
		t.Errorf("Expected 1 warn entry, got %d", len(warnEntries))
	}

	debugEntries := logger.EntriesByLevel(DebugLevel)
	if len(debugEntries) != 0 {
		t.Errorf("Expected 0 debug entries, got %d", len(debugEntries))
	}
}

func TestTestLoggerEntriesByMessage(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message 1")
	logger.Info("another test message")
	logger.Info("different message")
	logger.Error("test error")

	testEntries := logger.EntriesByMessage("test")
	if len(testEntries) != 3 {
		t.Errorf("Expected 3 entries containing 'test', got %d", len(testEntries))
	}

	differentEntries := logger.EntriesByMessage("different")
	if len(differentEntries) != 1 {
		t.Errorf("Expected 1 entry containing 'different', got %d", len(differentEntries))
	}

	nonExistentEntries := logger.EntriesByMessage("nonexistent")
	if len(nonExistentEntries) != 0 {
		t.Errorf("Expected 0 entries containing 'nonexistent', got %d", len(nonExistentEntries))
	}
}

func TestTestLoggerHasMethods(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message")
	logger.Error("error message")

	if !logger.HasMessage("test") {
		t.Error("Expected logger to have message containing 'test'")
	}

	if !logger.HasMessage("test message") {
		t.Error("Expected logger to have exact message")
	}

	if logger.HasMessage("nonexistent") {
		t.Error("Expected logger to not have nonexistent message")
	}

	if !logger.HasLevel(InfoLevel) {
		t.Error("Expected logger to have info level")
	}

	if !logger.HasLevel(ErrorLevel) {
		t.Error("Expected logger to have error level")
	}

	if logger.HasLevel(DebugLevel) {
		t.Error("Expected logger to not have debug level")
	}
}

func TestTestLoggerCount(t *testing.T) {
	logger := NewTestLogger()

	if logger.Count() != 0 {
		t.Errorf("Expected empty logger count to be 0, got %d", logger.Count())
	}

	logger.Info("message 1")
	if logger.Count() != 1 {
		t.Errorf("Expected count to be 1, got %d", logger.Count())
	}

	logger.Info("message 2")
	if logger.Count() != 2 {
		t.Errorf("Expected count to be 2, got %d", logger.Count())
	}

	if logger.CountByLevel(InfoLevel) != 2 {
		t.Errorf("Expected info level count to be 2, got %d", logger.CountByLevel(InfoLevel))
	}

	if logger.CountByLevel(ErrorLevel) != 0 {
		t.Errorf("Expected error level count to be 0, got %d", logger.CountByLevel(ErrorLevel))
	}
}

func TestTestLoggerClear(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("message 1")
	logger.Info("message 2")

	if logger.Count() != 2 {
		t.Errorf("Expected count to be 2 before clear, got %d", logger.Count())
	}

	logger.Clear()

	if !logger.Empty() {
		t.Error("Expected logger to be empty after clear")
	}

	if logger.Count() != 0 {
		t.Errorf("Expected count to be 0 after clear, got %d", logger.Count())
	}
}

func TestTestLoggerFirstLastEntry(t *testing.T) {
	logger := NewTestLogger()

	// Empty logger
	if logger.FirstEntry() != nil {
		t.Error("Expected FirstEntry() to return nil for empty logger")
	}
	if logger.LastEntry() != nil {
		t.Error("Expected LastEntry() to return nil for empty logger")
	}

	// Single entry
	logger.Info("single message")
	if logger.FirstEntry() == nil {
		t.Error("Expected FirstEntry() to not return nil for single entry")
	}
	if logger.LastEntry() == nil {
		t.Error("Expected LastEntry() to not return nil for single entry")
	}
	if logger.FirstEntry() != logger.LastEntry() {
		t.Error("Expected FirstEntry() and LastEntry() to be same for single entry")
	}

	// Multiple entries
	logger.Info("first message")
	logger.Info("last message")

	first := logger.FirstEntry()
	last := logger.LastEntry()

	if first.Message != "single message" {
		t.Errorf("Expected first message to be 'single message', got '%s'", first.Message)
	}

	if last.Message != "last message" {
		t.Errorf("Expected last message to be 'last message', got '%s'", last.Message)
	}
}

func TestTestLoggerSetLevel(t *testing.T) {
	logger := NewTestLogger()

	logger.SetLevel(WarnLevel)

	logger.Debug("debug message") // Should be filtered
	logger.Info("info message")   // Should be filtered
	logger.Warn("warn message")   // Should be included
	logger.Error("error message") // Should be included

	AssertLogCount(t, logger, 2)
	AssertLogContains(t, logger, "warn message")
	AssertLogContains(t, logger, "error message")
}

func TestLogCapture(t *testing.T) {
	// 创建一个简单的日志捕获测试，避免全局状态冲突
	logger := NewTestLogger()

	// 直接使用测试日志器测试日志捕获功能
	logger.Info("captured message 1")
	logger.Error("captured message 2")

	// 验证日志被正确捕获
	AssertLogCount(t, logger, 2)
	AssertLogContains(t, logger, "captured message 1")
	AssertLogContains(t, logger, "captured message 2")

	// 测试Entries方法
	entries := logger.Entries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 captured entries, got %d", len(entries))
	}

	// 验证条目内容
	if entries[0].Message != "captured message 1" {
		t.Errorf("Expected first message to be 'captured message 1', got '%s'", entries[0].Message)
	}
	if entries[1].Message != "captured message 2" {
		t.Errorf("Expected second message to be 'captured message 2', got '%s'", entries[1].Message)
	}
}

func TestLogCaptureClear(t *testing.T) {
	// 测试日志捕获的清除功能，避免全局状态冲突
	logger := NewTestLogger()

	// 添加一些日志条目
	logger.Info("message 1")
	logger.Info("message 2")

	if len(logger.Entries()) != 2 {
		t.Errorf("Expected 2 entries before clear, got %d", len(logger.Entries()))
	}

	// 清除日志
	logger.Clear()

	if len(logger.Entries()) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(logger.Entries()))
	}
}

func TestNewNoopLogger(t *testing.T) {
	logger := NewNoopLogger()
	if logger == nil {
		t.Error("Expected noop logger to be created")
	}

	// 所有操作都应该是空操作
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.Fatal("fatal")

	childLogger := logger.With(String("test", "value"))
	if childLogger == nil {
		t.Error("Expected noop logger With() to return a logger")
	}

	ctx := context.Background()
	contextLogger := logger.WithContext(ctx)
	if contextLogger == nil {
		t.Error("Expected noop logger WithContext() to return a logger")
	}

	err := logger.Sync()
	if err != nil {
		t.Errorf("Expected no error from noop logger Sync(), got %v", err)
	}
}

func TestTestLoggerConcurrency(t *testing.T) {
	logger := NewTestLogger()
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Info("message", Int("id", id))
		}(i)
	}

	wg.Wait()

	AssertLogCount(t, logger, 10)

	// 并发读取
	entries := logger.Entries()
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries after concurrent writes, got %d", len(entries))
	}
}

func TestTestLoggerFieldHandling(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("message",
		String("string_field", "value"),
		Int("int_field", 42),
		Bool("bool_field", true),
		Error(&testErrorType{msg: "test error"}),
	)

	AssertLogCount(t, logger, 1)

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(entry.Fields))
	}
}

// 测试用的错误类型
type testErrorType struct {
	msg string
}

func (e *testErrorType) Error() string {
	return e.msg
}

func TestTestLoggerTimeFields(t *testing.T) {
	logger := NewTestLogger()

	start := time.Now()
	logger.Info("timestamp test")
	end := time.Now()

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Time.Before(start) || entry.Time.After(end) {
		t.Errorf("Entry time %v is not within expected range [%v, %v]", entry.Time, start, end)
	}
}

func TestTestLoggerMessageValidation(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("")
	logger.Info("message with spaces")
	logger.Info("message with\nnewlines")
	logger.Info("message with\ttabs")

	AssertLogCount(t, logger, 4)

	entries := logger.Entries()
	if len(entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(entries))
	}

	// 验证消息原样保存
	expectedMessages := []string{"", "message with spaces", "message with\nnewlines", "message with\ttabs"}
	for i, expected := range expectedMessages {
		if entries[i].Message != expected {
			t.Errorf("Entry %d message mismatch, expected '%s', got '%s'", i, expected, entries[i].Message)
		}
	}
}
