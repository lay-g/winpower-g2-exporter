package log

import (
	"context"
	"testing"
	"time"
)

func TestContextWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "req-12345"

	ctx = WithRequestID(ctx, requestID)

	retrieved := RequestIDFromContext(ctx)
	if retrieved != requestID {
		t.Errorf("Expected request ID %s, got %s", requestID, retrieved)
	}
}

func TestContextWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "trace-67890"

	ctx = WithTraceID(ctx, traceID)

	retrieved := TraceIDFromContext(ctx)
	if retrieved != traceID {
		t.Errorf("Expected trace ID %s, got %s", traceID, retrieved)
	}
}

func TestContextWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := "user-abc123"

	ctx = WithUserID(ctx, userID)

	retrieved := UserIDFromContext(ctx)
	if retrieved != userID {
		t.Errorf("Expected user ID %s, got %s", userID, retrieved)
	}
}

func TestContextWithMultipleValues(t *testing.T) {
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req-123")
	ctx = WithTraceID(ctx, "trace-456")
	ctx = WithUserID(ctx, "user-789")
	ctx = WithUsername(ctx, "testuser")
	ctx = WithUserAgent(ctx, "test-agent")
	ctx = WithService(ctx, "test-service")
	ctx = WithVersion(ctx, "1.0.0")
	ctx = WithDeviceID(ctx, "device-123")
	ctx = WithAction(ctx, "test-action")
	ctx = WithDuration(ctx, 5*time.Second)

	// 验证所有值都能正确提取
	if RequestIDFromContext(ctx) != "req-123" {
		t.Error("Failed to extract request ID")
	}
	if TraceIDFromContext(ctx) != "trace-456" {
		t.Error("Failed to extract trace ID")
	}
	if UserIDFromContext(ctx) != "user-789" {
		t.Error("Failed to extract user ID")
	}
}

func TestExtractContextFields(t *testing.T) {
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req-123")
	ctx = WithTraceID(ctx, "trace-456")
	ctx = WithUserID(ctx, "user-789")
	ctx = WithAction(ctx, "test-action")
	ctx = WithDuration(ctx, 100*time.Millisecond)

	fields := extractContextFields(ctx)

	// 验证字段数量（至少应该有我们设置的字段）
	if len(fields) < 5 {
		t.Errorf("Expected at least 5 fields, got %d", len(fields))
	}

	// 验证特定字段存在
	foundRequestID := false
	foundTraceID := false
	foundUserID := false
	foundAction := false
	foundDuration := false

	for _, field := range fields {
		switch field.Key {
		case RequestIDKey:
			foundRequestID = true
		case TraceIDKey:
			foundTraceID = true
		case UserIDKey:
			foundUserID = true
		case ActionKey:
			foundAction = true
		case DurationKey:
			foundDuration = true
		}
	}

	if !foundRequestID {
		t.Error("Request ID field not found")
	}
	if !foundTraceID {
		t.Error("Trace ID field not found")
	}
	if !foundUserID {
		t.Error("User ID field not found")
	}
	if !foundAction {
		t.Error("Action field not found")
	}
	if !foundDuration {
		t.Error("Duration field not found")
	}
}

func TestExtractContextFieldsEmpty(t *testing.T) {
	ctx := context.Background()
	fields := extractContextFields(ctx)

	if len(fields) != 0 {
		t.Errorf("Expected no fields from empty context, got %d", len(fields))
	}
}

func TestWithContext(t *testing.T) {
	logger := NewTestLogger()
	ctx := context.Background()

	ctx = WithRequestID(ctx, "req-123")
	ctx = WithUserID(ctx, "user-456")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	// 应该包含上下文字段
	if len(entry.Fields) < 2 {
		t.Errorf("Expected at least 2 fields from context, got %d", len(entry.Fields))
	}
}

func TestLoggerFromContext(t *testing.T) {
	ctx := context.Background()
	logger := NewTestLogger()

	// 测试没有日志器的上下文
	result := LoggerFromContext(ctx, logger)
	if result == nil {
		t.Error("Expected logger from default logger when context has none")
	}

	// 测试有日志器的上下文
	ctx = WithLogger(ctx, logger)
	result = LoggerFromContext(ctx, nil)
	if result == nil {
		t.Error("Expected logger from context")
	}

	// 测试上下文日志器优先级
	ctx = WithLogger(ctx, logger)
	result = LoggerFromContext(ctx, NewTestLogger())
	if result != logger {
		t.Error("Expected context logger to take priority over default logger")
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	logger := NewTestLogger()

	// 没有日志器的上下文
	result := FromContext(ctx)
	if result != nil {
		t.Error("Expected nil when no logger in context")
	}

	// 有日志器的上下文
	ctx = WithLogger(ctx, logger)
	result = FromContext(ctx)
	if result == nil {
		t.Error("Expected logger from context")
	}
}

func TestWithFields(t *testing.T) {
	ctx := context.Background()
	logger := NewTestLogger()

	// 添加字段到上下文
	ctx = WithFields(ctx, String("field1", "value1"), Int("field2", 42))

	// 再次添加字段
	ctx = WithFields(ctx, String("field3", "value3"))

	fields := FieldsFromContext(ctx)
	if len(fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(fields))
	}

	// 使用带字段的日志器
	logger.WithContext(ctx).Info("test message")

	AssertLogCount(t, logger, 1)
	AssertLogContains(t, logger, "test message")

	entries := logger.Entries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if len(entry.Fields) < 3 {
		t.Errorf("Expected at least 3 fields, got %d", len(entry.Fields))
	}
}

func TestContextLoggingFunctions(t *testing.T) {
	logger := NewTestLogger()
	ctx := context.Background()

	ctx = WithLogger(ctx, logger)
	ctx = WithRequestID(ctx, "req-123")

	DebugWithContext(ctx, "debug message")
	InfoWithContext(ctx, "info message")
	WarnWithContext(ctx, "warn message")
	ErrorWithContext(ctx, "error message")

	AssertLogCount(t, logger, 4)
	AssertLogContains(t, logger, "debug message")
	AssertLogContains(t, logger, "info message")
	AssertLogContains(t, logger, "warn message")
	AssertLogContains(t, logger, "error message")

	entries := logger.Entries()
	if len(entries) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(entries))
	}

	// 验证所有条目都有上下文字段
	for _, entry := range entries {
		if len(entry.Fields) == 0 {
			t.Errorf("Entry '%s' should have context fields", entry.Message)
		}
	}
}

func TestContextLoggingWithoutLogger(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	// 没有日志器的上下文应该不产生任何日志
	InfoWithContext(ctx, "test message")

	// 这里我们无法直接测试，因为没有全局日志器的情况下
	// 这些函数应该是空操作
}

func TestContextValues(t *testing.T) {
	ctx := context.Background()

	// 测试所有上下文值设置和获取函数
	testCases := []struct {
		name    string
		setFunc func(context.Context, string) context.Context
		getFunc func(context.Context) string
		value   string
	}{
		{"request_id", WithRequestID, RequestIDFromContext, "req-123"},
		{"trace_id", WithTraceID, TraceIDFromContext, "trace-456"},
		{"user_id", WithUserID, UserIDFromContext, "user-789"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx = tc.setFunc(ctx, tc.value)
			retrieved := tc.getFunc(ctx)
			if retrieved != tc.value {
				t.Errorf("Expected %s, got %s", tc.value, retrieved)
			}
		})
	}
}

func TestContextWithServiceInfo(t *testing.T) {
	ctx := context.Background()

	ctx = WithService(ctx, "test-service")
	ctx = WithVersion(ctx, "1.0.0")
	ctx = WithDeviceID(ctx, "device-123")
	ctx = WithAction(ctx, "test-action")
	ctx = WithDuration(ctx, 5*time.Second)

	fields := extractContextFields(ctx)

	// 验证服务相关字段
	foundService := false
	foundVersion := false
	foundDeviceID := false
	foundAction := false
	foundDuration := false

	for _, field := range fields {
		switch field.Key {
		case ServiceKey:
			foundService = true
		case VersionKey:
			foundVersion = true
		case DeviceIDKey:
			foundDeviceID = true
		case ActionKey:
			foundAction = true
		case DurationKey:
			foundDuration = true
		}
	}

	if !foundService {
		t.Error("Service field not found")
	}
	if !foundVersion {
		t.Error("Version field not found")
	}
	if !foundDeviceID {
		t.Error("Device ID field not found")
	}
	if !foundAction {
		t.Error("Action field not found")
	}
	if !foundDuration {
		t.Error("Duration field not found")
	}
}

func TestContextWithUserAgent(t *testing.T) {
	ctx := context.Background()
	userAgent := "Mozilla/5.0 (Test Browser)"

	ctx = WithUserAgent(ctx, userAgent)

	fields := extractContextFields(ctx)
	foundUserAgent := false

	for _, field := range fields {
		if field.Key == UserAgentKey {
			foundUserAgent = true
			break
		}
	}

	if !foundUserAgent {
		t.Error("User agent field not found")
	}
}
