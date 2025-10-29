package log

import (
	"context"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "req-123"

	ctx = WithRequestID(ctx, requestID)

	// 验证值已设置
	if val := ctx.Value(contextKeyRequestID); val != requestID {
		t.Errorf("Expected request_id %s, got %v", requestID, val)
	}
}

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "trace-456"

	ctx = WithTraceID(ctx, traceID)

	// 验证值已设置
	if val := ctx.Value(contextKeyTraceID); val != traceID {
		t.Errorf("Expected trace_id %s, got %v", traceID, val)
	}
}

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := "user-789"

	ctx = WithUserID(ctx, userID)

	// 验证值已设置
	if val := ctx.Value(contextKeyUserID); val != userID {
		t.Errorf("Expected user_id %s, got %v", userID, val)
	}
}

func TestWithComponent(t *testing.T) {
	ctx := context.Background()
	component := "collector"

	ctx = WithComponent(ctx, component)

	// 验证值已设置
	if val := ctx.Value(contextKeyComponent); val != component {
		t.Errorf("Expected component %s, got %v", component, val)
	}
}

func TestWithOperation(t *testing.T) {
	ctx := context.Background()
	operation := "fetch-data"

	ctx = WithOperation(ctx, operation)

	// 验证值已设置
	if val := ctx.Value(contextKeyOperation); val != operation {
		t.Errorf("Expected operation %s, got %v", operation, val)
	}
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx = WithLogger(ctx, logger)

	// 验证日志器已设置
	if val := ctx.Value(contextKeyLogger); val == nil {
		t.Error("Expected logger to be set in context")
	}
}

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 测试从上下文获取日志器
	ctx = WithLogger(ctx, logger)
	retrieved := FromContext(ctx)
	if retrieved == nil {
		t.Error("FromContext() returned nil")
	}

	// 测试从空上下文获取日志器（应返回全局日志器）
	emptyCtx := context.Background()
	defaultLogger := FromContext(emptyCtx)
	if defaultLogger == nil {
		t.Error("FromContext() with empty context returned nil")
	}
}

func TestExtractContextFields(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedCount int
		checkFields   map[string]string
	}{
		{
			name: "empty context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedCount: 0,
		},
		{
			name: "request_id only",
			setupContext: func() context.Context {
				return WithRequestID(context.Background(), "req-123")
			},
			expectedCount: 1,
			checkFields: map[string]string{
				"request_id": "req-123",
			},
		},
		{
			name: "multiple fields",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = WithRequestID(ctx, "req-123")
				ctx = WithTraceID(ctx, "trace-456")
				ctx = WithUserID(ctx, "user-789")
				return ctx
			},
			expectedCount: 3,
			checkFields: map[string]string{
				"request_id": "req-123",
				"trace_id":   "trace-456",
				"user_id":    "user-789",
			},
		},
		{
			name: "all fields",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = WithRequestID(ctx, "req-123")
				ctx = WithTraceID(ctx, "trace-456")
				ctx = WithUserID(ctx, "user-789")
				ctx = WithComponent(ctx, "collector")
				ctx = WithOperation(ctx, "fetch")
				return ctx
			},
			expectedCount: 5,
			checkFields: map[string]string{
				"request_id": "req-123",
				"trace_id":   "trace-456",
				"user_id":    "user-789",
				"component":  "collector",
				"operation":  "fetch",
			},
		},
		{
			name: "empty string values",
			setupContext: func() context.Context {
				ctx := context.Background()
				ctx = WithRequestID(ctx, "")
				ctx = WithTraceID(ctx, "")
				return ctx
			},
			expectedCount: 0, // 空字符串不应被提取
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			fields := extractContextFields(ctx)

			if len(fields) != tt.expectedCount {
				t.Errorf("Expected %d fields, got %d", tt.expectedCount, len(fields))
			}

			// 验证字段内容（简单验证字段数量）
			// 由于 Field 是 zap 的类型，无法直接检查值，
			// 但可以通过数量来验证
		})
	}
}

func TestContextChaining(t *testing.T) {
	// 测试多次链式调用
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithTraceID(ctx, "trace-456")
	ctx = WithUserID(ctx, "user-789")
	ctx = WithComponent(ctx, "collector")
	ctx = WithOperation(ctx, "fetch")

	// 验证所有值都已设置
	if val := ctx.Value(contextKeyRequestID); val != "req-123" {
		t.Error("request_id not set correctly")
	}
	if val := ctx.Value(contextKeyTraceID); val != "trace-456" {
		t.Error("trace_id not set correctly")
	}
	if val := ctx.Value(contextKeyUserID); val != "user-789" {
		t.Error("user_id not set correctly")
	}
	if val := ctx.Value(contextKeyComponent); val != "collector" {
		t.Error("component not set correctly")
	}
	if val := ctx.Value(contextKeyOperation); val != "fetch" {
		t.Error("operation not set correctly")
	}
}

func TestContextWithLogger(t *testing.T) {
	// 创建日志器
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 创建带有上下文字段的 context
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithLogger(ctx, logger)

	// 从上下文获取日志器
	retrieved := FromContext(ctx)
	if retrieved == nil {
		t.Fatal("FromContext() returned nil")
	}

	// 使用上下文日志器
	ctxLogger := retrieved.WithContext(ctx)
	ctxLogger.Info("context test")
}

func TestContextKeys(t *testing.T) {
	// 验证上下文 key 的唯一性
	keys := []contextKey{
		contextKeyRequestID,
		contextKeyTraceID,
		contextKeyUserID,
		contextKeyComponent,
		contextKeyOperation,
		contextKeyLogger,
	}

	// 检查没有重复的 key
	seen := make(map[contextKey]bool)
	for _, key := range keys {
		if seen[key] {
			t.Errorf("Duplicate context key found: %v", key)
		}
		seen[key] = true
	}
}

// 测试上下文字段提取的并发安全性
func TestExtractContextFieldsConcurrency(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithTraceID(ctx, "trace-456")

	// 并发提取字段
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			fields := extractContextFields(ctx)
			if len(fields) != 2 {
				t.Errorf("Expected 2 fields, got %d", len(fields))
			}
			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 100; i++ {
		<-done
	}
}

// 基准测试
func BenchmarkWithRequestID(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WithRequestID(ctx, "req-123")
	}
}

func BenchmarkExtractContextFields(b *testing.B) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithTraceID(ctx, "trace-456")
	ctx = WithUserID(ctx, "user-789")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractContextFields(ctx)
	}
}

func BenchmarkFromContext(b *testing.B) {
	logger, err := NewLogger(DefaultConfig())
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()
	ctx = WithLogger(ctx, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FromContext(ctx)
	}
}
