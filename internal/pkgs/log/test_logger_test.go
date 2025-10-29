package log

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewTestLogger(t *testing.T) {
	logger := NewTestLogger()
	if logger == nil {
		t.Fatal("NewTestLogger returned nil")
	}

	if logger.Count() != 0 {
		t.Errorf("New test logger should have 0 entries, got %d", logger.Count())
	}
}

func TestNewNoopLogger(t *testing.T) {
	logger := NewNoopLogger()
	if logger == nil {
		t.Fatal("NewNoopLogger returned nil")
	}

	// Noop logger should not panic when called
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")

	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

func TestTestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(Logger)
		expected zapcore.Level
	}{
		{
			name:     "Debug level",
			logFunc:  func(l Logger) { l.Debug("debug message") },
			expected: zapcore.DebugLevel,
		},
		{
			name:     "Info level",
			logFunc:  func(l Logger) { l.Info("info message") },
			expected: zapcore.InfoLevel,
		},
		{
			name:     "Warn level",
			logFunc:  func(l Logger) { l.Warn("warn message") },
			expected: zapcore.WarnLevel,
		},
		{
			name:     "Error level",
			logFunc:  func(l Logger) { l.Error("error message") },
			expected: zapcore.ErrorLevel,
		},
		{
			name:     "Fatal level",
			logFunc:  func(l Logger) { l.Fatal("fatal message") },
			expected: zapcore.FatalLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewTestLogger()
			tt.logFunc(logger)

			entries := logger.Entries()
			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(entries))
			}

			if entries[0].Level != tt.expected {
				t.Errorf("Expected level %v, got %v", tt.expected, entries[0].Level)
			}
		})
	}
}

func TestTestLogger_WithFields(t *testing.T) {
	logger := NewTestLogger()
	logger.Info("test message",
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	)

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if len(entries[0].Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(entries[0].Fields))
	}
}

func TestTestLogger_With(t *testing.T) {
	logger := NewTestLogger()
	childLogger := logger.With(zap.String("module", "test"))

	childLogger.Info("test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entries[0].Message)
	}
}

func TestTestLogger_WithContext(t *testing.T) {
	logger := NewTestLogger()
	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Context == nil {
		t.Error("Expected context to be set")
	}
}

func TestTestLogger_Entries(t *testing.T) {
	logger := NewTestLogger()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	entries := logger.Entries()
	if len(entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entries))
	}
}

func TestTestLogger_EntriesByLevel(t *testing.T) {
	logger := NewTestLogger()

	logger.Debug("debug")
	logger.Info("info1")
	logger.Info("info2")
	logger.Warn("warn")

	infoEntries := logger.EntriesByLevel(zapcore.InfoLevel)
	if len(infoEntries) != 2 {
		t.Errorf("Expected 2 info entries, got %d", len(infoEntries))
	}

	debugEntries := logger.EntriesByLevel(zapcore.DebugLevel)
	if len(debugEntries) != 1 {
		t.Errorf("Expected 1 debug entry, got %d", len(debugEntries))
	}
}

func TestTestLogger_EntriesByMessage(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message")
	logger.Info("test message")
	logger.Info("different message")

	entries := logger.EntriesByMessage("test message")
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with 'test message', got %d", len(entries))
	}

	entries = logger.EntriesByMessage("different message")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry with 'different message', got %d", len(entries))
	}
}

func TestTestLogger_EntriesByField(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("message1", zap.String("key", "value1"))
	logger.Info("message2", zap.String("key", "value1"))
	logger.Info("message3", zap.String("key", "value2"))

	entries := logger.EntriesByField("key", "value1")
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with key=value1, got %d", len(entries))
	}

	entries = logger.EntriesByField("key", "value2")
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry with key=value2, got %d", len(entries))
	}
}

func TestTestLogger_Clear(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test1")
	logger.Info("test2")

	if logger.Count() != 2 {
		t.Errorf("Expected 2 entries before clear, got %d", logger.Count())
	}

	logger.Clear()

	if logger.Count() != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", logger.Count())
	}
}

func TestTestLogger_Count(t *testing.T) {
	logger := NewTestLogger()

	if logger.Count() != 0 {
		t.Errorf("Expected 0 entries initially, got %d", logger.Count())
	}

	logger.Info("test1")
	if logger.Count() != 1 {
		t.Errorf("Expected 1 entry, got %d", logger.Count())
	}

	logger.Info("test2")
	if logger.Count() != 2 {
		t.Errorf("Expected 2 entries, got %d", logger.Count())
	}
}

func TestTestLogger_HasEntry(t *testing.T) {
	logger := NewTestLogger()

	logger.Info("test message",
		zap.String("key1", "value1"),
		zap.Int("key2", 42),
	)

	// Test exact match
	hasEntry := logger.HasEntry(zapcore.InfoLevel, "test message", map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})
	if !hasEntry {
		t.Error("Expected to find matching entry")
	}

	// Test wrong level
	hasEntry = logger.HasEntry(zapcore.DebugLevel, "test message", map[string]interface{}{
		"key1": "value1",
	})
	if hasEntry {
		t.Error("Should not find entry with wrong level")
	}

	// Test wrong message
	hasEntry = logger.HasEntry(zapcore.InfoLevel, "wrong message", map[string]interface{}{
		"key1": "value1",
	})
	if hasEntry {
		t.Error("Should not find entry with wrong message")
	}

	// Test wrong field value
	hasEntry = logger.HasEntry(zapcore.InfoLevel, "test message", map[string]interface{}{
		"key1": "wrong_value",
	})
	if hasEntry {
		t.Error("Should not find entry with wrong field value")
	}

	// Test empty fields map (should match)
	hasEntry = logger.HasEntry(zapcore.InfoLevel, "test message", map[string]interface{}{})
	if !hasEntry {
		t.Error("Expected to find entry when fields map is empty")
	}
}

func TestTestLogger_Sync(t *testing.T) {
	logger := NewTestLogger()

	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

func TestTestLogger_Core(t *testing.T) {
	logger := NewTestLogger()
	core := logger.Core()

	if core == nil {
		t.Error("Core() returned nil")
	}
}

func TestNewLogCapture(t *testing.T) {
	capture := NewLogCapture()
	if capture == nil {
		t.Fatal("NewLogCapture returned nil")
	}

	entries := capture.Entries()
	if len(entries) != 0 {
		t.Errorf("New capture should have 0 entries, got %d", len(entries))
	}
}

func TestLogCapture_Capture(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	logger.Info("test message", zap.String("key", "value"))

	entries := capture.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entries[0].Message)
	}

	if entries[0].Level != zapcore.InfoLevel {
		t.Errorf("Expected level Info, got %v", entries[0].Level)
	}
}

func TestLogCapture_WithContext(t *testing.T) {
	capture := NewLogCapture()
	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	logger := capture.WithContext(ctx)
	logger.Info("test message")

	entries := capture.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Context == nil {
		t.Error("Expected context to be set")
	}
}

func TestLogCapture_Clear(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	logger.Info("test1")
	logger.Info("test2")

	entries := capture.Entries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries before clear, got %d", len(entries))
	}

	capture.Clear()

	entries = capture.Entries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
}

func TestCaptureLogger_With(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	childLogger := logger.With(zap.String("module", "test"))
	childLogger.Info("test message", zap.String("key", "value"))

	entries := capture.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	// Should have both the "module" field from With() and "key" field from Info()
	if len(entries[0].Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(entries[0].Fields))
	}
}

func TestCaptureLogger_WithContext(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	ctx := context.WithValue(context.Background(), "test_key", "test_value")
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	entries := capture.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Context == nil {
		t.Error("Expected context to be set")
	}
}

func TestCaptureLogger_AllLevels(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
	logger.Fatal("fatal")

	entries := capture.Entries()
	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}

	// Verify all levels are captured
	levels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.FatalLevel,
	}

	for i, expectedLevel := range levels {
		if entries[i].Level != expectedLevel {
			t.Errorf("Entry %d: expected level %v, got %v", i, expectedLevel, entries[i].Level)
		}
	}
}

func TestCaptureLogger_Sync(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() returned error: %v", err)
	}
}

func TestCaptureLogger_Core(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	core := logger.Core()
	if core == nil {
		t.Error("Core() returned nil")
	}
}

func TestHasField(t *testing.T) {
	tests := []struct {
		name     string
		fields   []Field
		key      string
		value    interface{}
		expected bool
	}{
		{
			name: "String field match",
			fields: []Field{
				zap.String("key", "value"),
			},
			key:      "key",
			value:    "value",
			expected: true,
		},
		{
			name: "String field no match",
			fields: []Field{
				zap.String("key", "value"),
			},
			key:      "key",
			value:    "wrong",
			expected: false,
		},
		{
			name: "Int field match",
			fields: []Field{
				zap.Int("key", 42),
			},
			key:      "key",
			value:    42,
			expected: true,
		},
		{
			name: "Int64 field match",
			fields: []Field{
				zap.Int64("key", 42),
			},
			key:      "key",
			value:    int64(42),
			expected: true,
		},
		{
			name: "Bool field match true",
			fields: []Field{
				zap.Bool("key", true),
			},
			key:      "key",
			value:    true,
			expected: true,
		},
		{
			name: "Bool field match false",
			fields: []Field{
				zap.Bool("key", false),
			},
			key:      "key",
			value:    false,
			expected: true,
		},
		{
			name:     "Empty fields",
			fields:   []Field{},
			key:      "key",
			value:    "value",
			expected: false,
		},
		{
			name: "Key not found",
			fields: []Field{
				zap.String("other", "value"),
			},
			key:      "key",
			value:    "value",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasField(tt.fields, tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("hasField() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   []Field
		expected map[string]interface{}
		want     bool
	}{
		{
			name: "All fields match",
			fields: []Field{
				zap.String("key1", "value1"),
				zap.Int("key2", 42),
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			want: true,
		},
		{
			name: "Partial match fails",
			fields: []Field{
				zap.String("key1", "value1"),
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			want: false,
		},
		{
			name: "Empty expected matches anything",
			fields: []Field{
				zap.String("key1", "value1"),
			},
			expected: map[string]interface{}{},
			want:     true,
		},
		{
			name:   "No fields no expected",
			fields: []Field{},
			expected: map[string]interface{}{
				"key1": "value1",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchFields(tt.fields, tt.expected)
			if result != tt.want {
				t.Errorf("matchFields() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestTestLogger_ConcurrentAccess(t *testing.T) {
	logger := NewTestLogger()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent message", zap.Int("id", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries were captured
	if logger.Count() != 10 {
		t.Errorf("Expected 10 entries, got %d", logger.Count())
	}
}

func TestLogCapture_ConcurrentAccess(t *testing.T) {
	capture := NewLogCapture()
	logger := capture.Capture()

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent message", zap.Int("id", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all entries were captured
	entries := capture.Entries()
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries, got %d", len(entries))
	}
}
