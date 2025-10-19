package log

import (
	"context"
	"errors"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name:    "development config",
			config:  DevelopmentDefaults(),
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "file output config",
			config: &Config{
				Level:    InfoLevel,
				Format:   JSONFormat,
				Output:   FileOutput,
				Filename: t.TempDir() + "/test.log",
			},
			wantErr: false,
		},
		{
			name: "invalid config",
			config: &Config{
				Level:  Level("invalid"),
				Format: JSONFormat,
				Output: StdoutOutput,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("NewLogger() returned nil logger")
			}
			if logger != nil {
				_ = logger.Sync()
			}
		})
	}
}

func TestZapLoggerBasicLogging(t *testing.T) {
	logger := NewTestLogger()

	testCases := []struct {
		level   Level
		message string
	}{
		{DebugLevel, "debug message"},
		{InfoLevel, "info message"},
		{WarnLevel, "warn message"},
		{ErrorLevel, "error message"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.level), func(t *testing.T) {
			logger.Clear()
			logger.SetLevel(DebugLevel)

			switch tc.level {
			case DebugLevel:
				logger.Debug(tc.message)
			case InfoLevel:
				logger.Info(tc.message)
			case WarnLevel:
				logger.Warn(tc.message)
			case ErrorLevel:
				logger.Error(tc.message)
			}

			AssertLogCount(t, logger, 1)
			AssertLogContains(t, logger, tc.message)
			AssertLogHasLevel(t, logger, tc.level)
		})
	}
}

func TestZapLoggerWithFields(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message",
		String("user_id", "12345"),
		Int("count", 42),
		Bool("active", true),
		Error(errors.New("test error")),
	)

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.Fields) != 4 {
		t.Errorf("Expected 4 fields, got %d", len(entry.Fields))
	}
}

func TestZapLoggerWith(t *testing.T) {
	logger := NewTestLogger()

	childLogger := logger.With(
		String("component", "test"),
		Int("version", 1),
	)

	childLogger.Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")
}

func TestZapLoggerWithContext(t *testing.T) {
	logger := NewTestLogger()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithUserID(ctx, "user-456")

	logger.WithContext(ctx).Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.Fields) < 2 {
		t.Errorf("Expected at least 2 fields from context, got %d", len(entry.Fields))
	}
}

func TestFieldConstructors(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message",
		String("string_key", "string_value"),
		Int("int_key", 42),
		Int64("int64_key", int64(1234567890)),
		Float64("float64_key", 3.14159),
		Bool("bool_key", true),
		Any("any_key", map[string]string{"nested": "value"}),
	)

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.Fields) != 6 {
		t.Errorf("Expected 6 fields, got %d", len(entry.Fields))
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	logger := NewTestLogger()

	logger.SetLevel(InfoLevel)

	logger.Debug("debug message") // Should be filtered out
	logger.Info("info message")   // Should be included
	logger.Warn("warn message")   // Should be included

	AssertLogCount(t, logger, 2)
	AssertLogContains(t, logger, "info message")
	AssertLogContains(t, logger, "warn message")

	// 验证debug消息被过滤掉了
	if logger.HasMessage("debug message") {
		t.Error("Debug message should be filtered out")
	}
}

func TestLoggerEmpty(t *testing.T) {
	logger := NewTestLogger()

	if !logger.Empty() {
		t.Error("Expected new logger to be empty")
	}

	logger.Info("test message")

	if logger.Empty() {
		t.Error("Expected logger with entries to not be empty")
	}
}

func TestLoggerFirstLastEntry(t *testing.T) {
	logger := NewTestLogger()

	// Empty logger should return nil
	if logger.FirstEntry() != nil {
		t.Error("Expected FirstEntry() to return nil for empty logger")
	}
	if logger.LastEntry() != nil {
		t.Error("Expected LastEntry() to return nil for empty logger")
	}

	logger.Info("first message")
	logger.Info("second message")

	first := logger.FirstEntry()
	last := logger.LastEntry()

	if first == nil || last == nil {
		t.Error("Expected entries to not be nil")
		return
	}

	if first.Message != "first message" {
		t.Errorf("Expected first message to be 'first message', got '%s'", first.Message)
	}

	if last.Message != "second message" {
		t.Errorf("Expected last message to be 'second message', got '%s'", last.Message)
	}
}

func TestLoggerClear(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message")
	AssertLogCount(t, logger, 1)

	logger.Clear()
	AssertLogEmpty(t, logger)
}

func TestLoggerCountByLevel(t *testing.T) {
	logger := NewTestLogger()

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Info("another info message")
	logger.Warn("warn message")
	logger.Error("error message")

	AssertLogCount(t, logger, 5)
	if logger.CountByLevel(DebugLevel) != 1 {
		t.Errorf("Expected 1 debug level entry, got %d", logger.CountByLevel(DebugLevel))
	}
	if logger.CountByLevel(InfoLevel) != 2 {
		t.Errorf("Expected 2 info level entries, got %d", logger.CountByLevel(InfoLevel))
	}
	if logger.CountByLevel(WarnLevel) != 1 {
		t.Errorf("Expected 1 warn level entry, got %d", logger.CountByLevel(WarnLevel))
	}
	if logger.CountByLevel(ErrorLevel) != 1 {
		t.Errorf("Expected 1 error level entry, got %d", logger.CountByLevel(ErrorLevel))
	}
}

func TestLoggerEntriesByMessage(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message")
	logger.Info("another test message")
	logger.Info("different message")

	entries := logger.EntriesByMessage("test")
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries containing 'test', got %d", len(entries))
	}

	entries = logger.EntriesByMessage("different")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry containing 'different', got %d", len(entries))
	}

	entries = logger.EntriesByMessage("nonexistent")
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries containing 'nonexistent', got %d", len(entries))
	}
}

func TestLoggerHasMessage(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message")

	if !logger.HasMessage("test") {
		t.Error("Expected logger to have message containing 'test'")
	}

	if !logger.HasMessage("test message") {
		t.Error("Expected logger to have exact message")
	}

	if logger.HasMessage("nonexistent") {
		t.Error("Expected logger to not have nonexistent message")
	}
}

func TestLoggerHasLevel(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("info message")
	logger.Error("error message")

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

func TestLoggerSetLevel(t *testing.T) {
	logger := NewTestLogger()

	logger.SetLevel(WarnLevel)

	if logger.level != WarnLevel {
		t.Errorf("Expected level to be set to %s, got %s", WarnLevel, logger.level)
	}
}
