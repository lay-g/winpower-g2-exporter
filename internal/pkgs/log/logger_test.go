package log

import (
	"bytes"
	"context"
	"os"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name:      "default config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name:      "development config",
			config:    DevelopmentDefaults(),
			wantError: false,
		},
		{
			name: "invalid config",
			config: &Config{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if logger == nil {
					t.Error("Expected logger but got nil")
				}
			}
		})
	}
}

func TestLoggerInterface(t *testing.T) {
	// 测试日志器是否实现了接口
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 测试所有方法是否可调用
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// 测试 With
	childLogger := logger.With(String("key", "value"))
	if childLogger == nil {
		t.Error("With() returned nil")
	}

	// 测试 WithContext
	ctx := context.Background()
	ctxLogger := logger.WithContext(ctx)
	if ctxLogger == nil {
		t.Error("WithContext() returned nil")
	}

	// 测试 Core
	core := logger.Core()
	if core == nil {
		t.Error("Core() returned nil")
	}

	// 测试 Sync
	if err := logger.Sync(); err != nil {
		// Sync 可能返回错误（如 stdout），不算失败
		t.Logf("Sync returned error (expected for stdout): %v", err)
	}
}

func TestLoggerWithFields(t *testing.T) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 测试各种字段类型
	logger.Info("test message",
		String("string", "value"),
		Int("int", 42),
		Int64("int64", 123456),
		Float64("float", 3.14),
		Bool("bool", true),
	)
}

func TestLoggerWithContext(t *testing.T) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 创建带有上下文字段的 context
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithTraceID(ctx, "trace-456")
	ctx = WithUserID(ctx, "user-789")

	// 测试上下文日志
	ctxLogger := logger.WithContext(ctx)
	ctxLogger.Info("context message")
}

func TestGlobalLogger(t *testing.T) {
	// 重置全局日志器
	ResetGlobal()

	// 测试默认初始化
	logger := Default()
	if logger == nil {
		t.Error("Default() returned nil")
	}

	// 测试使用全局日志器
	Debug("debug")
	Info("info")
	Warn("warn")
	Error("error")
}

func TestGlobalLoggerInit(t *testing.T) {
	// 重置全局日志器
	ResetGlobal()

	// 测试初始化
	config := DefaultConfig()
	err := Init(config)
	if err != nil {
		t.Errorf("Init() failed: %v", err)
	}

	// 验证全局日志器已设置
	logger := Default()
	if logger == nil {
		t.Error("Global logger not set after Init()")
	}
}

func TestInitDevelopment(t *testing.T) {
	// 重置全局日志器
	ResetGlobal()

	// 测试开发环境初始化
	err := InitDevelopment()
	if err != nil {
		t.Errorf("InitDevelopment() failed: %v", err)
	}

	// 验证全局日志器已设置
	logger := Default()
	if logger == nil {
		t.Error("Global logger not set after InitDevelopment()")
	}
}

func TestGlobalLoggerContext(t *testing.T) {
	// 重置全局日志器
	ResetGlobal()

	// 初始化
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("InitDevelopment() failed: %v", err)
	}

	// 创建上下文
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	// 测试上下文全局日志
	InfoContext(ctx, "context info")
	DebugContext(ctx, "context debug")
	WarnContext(ctx, "context warn")
	ErrorContext(ctx, "context error")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		level string
		want  zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"fatal", zapcore.FatalLevel},
		{"invalid", zapcore.InfoLevel}, // 默认为 info
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			got := parseLevel(tt.level)
			if got != tt.want {
				t.Errorf("parseLevel(%s) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestFieldConstructors(t *testing.T) {
	// 测试所有字段构造器
	_ = String("key", "value")
	_ = Int("key", 42)
	_ = Int64("key", 123)
	_ = Float64("key", 3.14)
	_ = Bool("key", true)
	_ = Any("key", struct{ Name string }{Name: "test"})
}

// 测试日志输出格式（通过捕获输出验证）
func TestLogOutput(t *testing.T) {
	// 创建一个临时的日志配置，输出到 stderr 以便捕获
	config := &Config{
		Level:            "info",
		Format:           "json",
		Output:           "stderr",
		Development:      false,
		EnableCaller:     false,
		EnableStacktrace: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 记录一条日志
	logger.Info("test message", String("key", "value"))

	// 注意：实际验证输出内容需要重定向 stderr，这里只是测试不崩溃
}

// TestLoggerConcurrency 测试并发安全性
func TestLoggerConcurrency(t *testing.T) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 并发写入日志
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				logger.Info("concurrent log", Int("goroutine", id), Int("iteration", j))
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestGlobalLoggerConcurrency 测试全局日志器并发安全性
func TestGlobalLoggerConcurrency(t *testing.T) {
	ResetGlobal()
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("InitDevelopment() failed: %v", err)
	}

	// 并发使用全局日志器
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				Info("concurrent global log", Int("goroutine", id), Int("iteration", j))
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// 测试不同日志级别
func TestLogLevels(t *testing.T) {
	levels := []struct {
		name   string
		level  string
		enable map[string]bool // 哪些级别应该被启用
	}{
		{
			name:  "debug level",
			level: "debug",
			enable: map[string]bool{
				"debug": true,
				"info":  true,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "info level",
			level: "info",
			enable: map[string]bool{
				"debug": false,
				"info":  true,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "warn level",
			level: "warn",
			enable: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "error level",
			level: "error",
			enable: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  false,
				"error": true,
			},
		},
	}

	for _, tt := range levels {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Level = tt.level

			logger, err := NewLogger(config)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			// 验证日志级别
			core := logger.Core()

			if tt.enable["debug"] != core.Enabled(zapcore.DebugLevel) {
				t.Errorf("Debug level enabled = %v, want %v", core.Enabled(zapcore.DebugLevel), tt.enable["debug"])
			}
			if tt.enable["info"] != core.Enabled(zapcore.InfoLevel) {
				t.Errorf("Info level enabled = %v, want %v", core.Enabled(zapcore.InfoLevel), tt.enable["info"])
			}
			if tt.enable["warn"] != core.Enabled(zapcore.WarnLevel) {
				t.Errorf("Warn level enabled = %v, want %v", core.Enabled(zapcore.WarnLevel), tt.enable["warn"])
			}
			if tt.enable["error"] != core.Enabled(zapcore.ErrorLevel) {
				t.Errorf("Error level enabled = %v, want %v", core.Enabled(zapcore.ErrorLevel), tt.enable["error"])
			}
		})
	}
}

// 测试实际的日志输出内容
func TestActualLogOutput(t *testing.T) {
	// 创建内存缓冲区
	var buf bytes.Buffer

	// 无法直接重定向 zap 的输出到 buffer，
	// 这里只是演示测试思路
	_ = buf
}

// 测试 JSON 格式输出
func TestJSONFormatOutput(t *testing.T) {
	config := &Config{
		Level:            "info",
		Format:           "json",
		Output:           "stdout",
		Development:      false,
		EnableCaller:     false,
		EnableStacktrace: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 记录日志（无法直接验证输出，但确保不崩溃）
	logger.Info("json message", String("key", "value"), Int("count", 42))
}

// 测试 Console 格式输出
func TestConsoleFormatOutput(t *testing.T) {
	config := &Config{
		Level:            "info",
		Format:           "console",
		Output:           "stdout",
		Development:      true,
		EnableCaller:     true,
		EnableStacktrace: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 记录日志（无法直接验证输出，但确保不崩溃）
	logger.Info("console message", String("key", "value"), Int("count", 42))
}

// 测试错误处理
func TestErrorField(t *testing.T) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 测试错误字段
	testErr := os.ErrNotExist
	logger.Error("error occurred", Err(testErr))
}

// 基准测试
func BenchmarkLogger(b *testing.B) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message", String("key", "value"), Int("count", i))
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	config := DefaultConfig()
	logger, err := NewLogger(config)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("string", "value"),
			Int("int", 42),
			Float64("float", 3.14),
			Bool("bool", true),
		)
	}
}

func BenchmarkGlobalLogger(b *testing.B) {
	ResetGlobal()
	err := Init(DefaultConfig())
	if err != nil {
		b.Fatalf("Init() failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message", String("key", "value"), Int("count", i))
	}
}
