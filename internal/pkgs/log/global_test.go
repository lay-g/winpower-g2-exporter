package log

import (
	"context"
	"testing"
)

func TestInit(t *testing.T) {
	// 重置全局状态
	ResetGlobal()

	// 测试用nil配置初始化
	err := Init(nil)
	if err != nil {
		t.Errorf("Expected no error with nil config, got %v", err)
	}

	if !IsInitialized() {
		t.Error("Expected global logger to be initialized")
	}

	ResetGlobal()
}

func TestInitDevelopment(t *testing.T) {
	ResetGlobal()

	err := InitDevelopment()
	if err != nil {
		t.Errorf("Expected no error initializing development logger, got %v", err)
	}

	if !IsInitialized() {
		t.Error("Expected global logger to be initialized")
	}

	ResetGlobal()
}

func TestResetGlobal(t *testing.T) {
	// 先初始化
	_ = InitDevelopment()

	if !IsInitialized() {
		t.Error("Expected global logger to be initialized before reset")
	}

	ResetGlobal()

	if IsInitialized() {
		t.Error("Expected global logger to be not initialized after reset")
	}
}

func TestDefault(t *testing.T) {
	ResetGlobal()

	// 未初始化时应该自动初始化
	logger := Default()
	if logger == nil {
		t.Error("Expected default logger to be created")
	}

	if !IsInitialized() {
		t.Error("Expected global logger to be initialized after Default() call")
	}

	ResetGlobal()
}

func TestGlobalLogging(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	// 测试全局日志函数的基本功能（不会崩溃）
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	GlobalError("error message")

	// 验证日志器已初始化
	if !IsInitialized() {
		t.Error("Global logger should be initialized")
	}

	ResetGlobal()
}

func TestGlobalWith(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	childLogger := With(String("component", "test"))
	if childLogger == nil {
		t.Error("Expected child logger to be created")
	}

	// 测试不会崩溃
	childLogger.Info("test message")

	ResetGlobal()
}

func TestGlobalWithContext(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	// 使用全局日志器测试上下文功能
	contextLogger := WithContext(ctx)
	if contextLogger == nil {
		t.Error("Expected context logger to be created")
	}

	// 验证不会崩溃
	contextLogger.Info("test message")

	ResetGlobal()
}

func TestGlobalContextLogging(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")

	// 测试全局上下文日志函数不会崩溃
	DebugWithContext(ctx, "debug message")
	InfoWithContext(ctx, "info message")
	WarnWithContext(ctx, "warn message")
	ErrorWithContext(ctx, "error message")

	// 验证全局日志器已初始化
	if !IsInitialized() {
		t.Error("Global logger should be initialized")
	}

	ResetGlobal()
}

func TestGetLevel(t *testing.T) {
	ResetGlobal()

	// 未初始化时应该返回默认级别
	level := GetLevel()
	if level != InfoLevel {
		t.Errorf("Expected default level to be %s, got %s", InfoLevel, level)
	}

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	// 获取初始化后的级别
	level = GetLevel()
	if level != DebugLevel {
		t.Errorf("Expected development level to be %s, got %s", DebugLevel, level)
	}

	ResetGlobal()
}

func TestSetLevel(t *testing.T) {
	ResetGlobal()

	// 先初始化
	_ = InitDevelopment()

	err := SetLevel(ErrorLevel)
	if err != nil {
		t.Errorf("Expected no error setting level, got %v", err)
	}

	ResetGlobal()
}

func TestSetOutput(t *testing.T) {
	ResetGlobal()

	_ = InitDevelopment()

	err := SetOutput(StderrOutput)
	if err != nil {
		t.Errorf("Expected no error setting output, got %v", err)
	}

	ResetGlobal()
}

func TestSetFormat(t *testing.T) {
	ResetGlobal()

	_ = InitDevelopment()

	err := SetFormat(JSONFormat)
	if err != nil {
		t.Errorf("Expected no error setting format, got %v", err)
	}

	ResetGlobal()
}

func TestSetFile(t *testing.T) {
	ResetGlobal()

	_ = InitDevelopment()

	tempFile := t.TempDir() + "/test.log"
	err := SetFile(tempFile, 100, 30, 10, true)
	if err != nil {
		t.Errorf("Expected no error setting file, got %v", err)
	}

	ResetGlobal()
}

func TestReconfigure(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	_ = InitDevelopment()

	// 重新配置为生产日志器
	config := &Config{
		Level:    InfoLevel,
		Format:   JSONFormat,
		Output:   StdoutOutput,
		Compress: true,
	}

	err := Reconfigure(config)
	if err != nil {
		t.Errorf("Expected no error reconfiguring, got %v", err)
	}

	if !IsInitialized() {
		t.Error("Expected global logger to remain initialized")
	}

	ResetGlobal()
}

func TestReconfigureWithNil(t *testing.T) {
	ResetGlobal()

	_ = InitDevelopment()

	err := Reconfigure(nil)
	if err == nil {
		t.Error("Expected error with nil config")
	}

	ResetGlobal()
}

func TestReconfigureWithInvalid(t *testing.T) {
	ResetGlobal()

	_ = InitDevelopment()

	config := &Config{
		Level:  Level("invalid"),
		Format: JSONFormat,
		Output: StdoutOutput,
	}

	err := Reconfigure(config)
	if err == nil {
		t.Error("Expected error with invalid config")
	}

	ResetGlobal()
}

func TestSync(t *testing.T) {
	ResetGlobal()

	// 未初始化时Sync应该不返回错误
	err := Sync()
	if err != nil {
		t.Errorf("Expected no error syncing uninitialized logger, got %v", err)
	}

	// 初始化后使用文件输出，避免stdout/stderr同步问题
	tempFile := t.TempDir() + "/test.log"
	config := DevelopmentDefaults()
	config.Output = FileOutput
	config.Filename = tempFile

	err = Init(config)
	if err != nil {
		t.Fatalf("Failed to initialize logger with file output: %v", err)
	}

	err = Sync()
	if err != nil {
		t.Errorf("Expected no error syncing initialized logger with file output, got %v", err)
	}

	ResetGlobal()
}

func TestGlobalConfig(t *testing.T) {
	ResetGlobal()

	// 未初始化时应该返回默认配置
	config := GetGlobalConfig()
	if config == nil {
		t.Error("Expected global config to not be nil")
	}

	// 初始化后应该返回当前配置
	_ = InitDevelopment()

	config = GetGlobalConfig()
	if config == nil {
		t.Error("Expected config to not be nil after initialization")
		return
	}

	if config.Level != DebugLevel {
		t.Errorf("Expected development config level to be %s, got %s", DebugLevel, config.Level)
	}

	ResetGlobal()
}

func TestGlobalWithCustomLogger(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	// 测试全局函数不会崩溃
	Info("global message")

	// 验证全局日志器已初始化
	if !IsInitialized() {
		t.Error("Global logger should be initialized")
	}

	ResetGlobal()
}

func TestGlobalFatal(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	// Fatal会调用os.Exit，所以我们不能直接测试它
	// 但我们可以验证日志器设置正确
	if !IsInitialized() {
		t.Error("Expected global logger to be initialized")
	}

	ResetGlobal()
}

func TestGlobalConcurrency(t *testing.T) {
	ResetGlobal()

	// 初始化开发日志器
	err := InitDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialize development logger: %v", err)
	}

	// 并发测试
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			Info("message from goroutine", Int("id", id))
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证全局日志器已初始化
	if !IsInitialized() {
		t.Error("Global logger should be initialized")
	}

	ResetGlobal()
}
