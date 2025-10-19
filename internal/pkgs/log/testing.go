package log

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// LogEntry 表示一条日志条目
type LogEntry struct {
	Time    time.Time       `json:"time"`
	Level   Level           `json:"level"`
	Message string          `json:"message"`
	Fields  []Field         `json:"fields"`
	Context context.Context `json:"-"`
}

// TestLogger 测试用日志器，捕获日志条目用于测试验证
type TestLogger struct {
	entries []LogEntry
	mu      sync.RWMutex
	level   Level
}

// NewTestLogger 创建新的测试日志器
func NewTestLogger() *TestLogger {
	return &TestLogger{
		entries: make([]LogEntry, 0),
		level:   DebugLevel,
	}
}

// NewTestLoggerWithT 创建与测试对象绑定的测试日志器
func NewTestLoggerWithT(t *testing.T) *TestLogger {
	logger := NewTestLogger()

	// 在测试结束时自动输出所有日志（如果测试失败）
	t.Cleanup(func() {
		if t.Failed() {
			logger.PrintAllLogs(t)
		}
	})

	return logger
}

// Debug 记录调试级别日志
func (tl *TestLogger) Debug(msg string, fields ...Field) {
	tl.log(DebugLevel, msg, fields...)
}

// Info 记录信息级别日志
func (tl *TestLogger) Info(msg string, fields ...Field) {
	tl.log(InfoLevel, msg, fields...)
}

// Warn 记录警告级别日志
func (tl *TestLogger) Warn(msg string, fields ...Field) {
	tl.log(WarnLevel, msg, fields...)
}

// Error 记录错误级别日志
func (tl *TestLogger) Error(msg string, fields ...Field) {
	tl.log(ErrorLevel, msg, fields...)
}

// Fatal 记录致命错误级别日志
func (tl *TestLogger) Fatal(msg string, fields ...Field) {
	tl.log(FatalLevel, msg, fields...)
}

// With 创建带有额外字段的子日志器
func (tl *TestLogger) With(fields ...Field) Logger {
	return &testLoggerWithFields{
		TestLogger: tl,
		fields:     fields,
	}
}

// WithContext 创建带有上下文的日志器
func (tl *TestLogger) WithContext(ctx context.Context) Logger {
	// 提取上下文字段
	contextFields := extractContextFields(ctx)

	// 创建包含上下文字段的测试日志器
	return &contextTestLogger{
		TestLogger: tl,
		ctx:        ctx,
		fields:     contextFields,
	}
}

// Sync 同步缓冲区（测试日志器不需要同步）
func (tl *TestLogger) Sync() error {
	return nil
}

// log 内部日志记录方法
func (tl *TestLogger) log(level Level, msg string, fields ...Field) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	// 只记录当前级别及以上级别的日志
	if !tl.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Message: msg,
		Fields:  fields,
	}

	tl.entries = append(tl.entries, entry)
}

// shouldLog 检查是否应该记录指定级别的日志
func (tl *TestLogger) shouldLog(level Level) bool {
	levels := map[Level]int{
		DebugLevel: 0,
		InfoLevel:  1,
		WarnLevel:  2,
		ErrorLevel: 3,
		FatalLevel: 4,
	}

	return levels[level] >= levels[tl.level]
}

// SetLevel 设置日志级别
func (tl *TestLogger) SetLevel(level Level) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.level = level
}

// Entries 获取所有日志条目
func (tl *TestLogger) Entries() []LogEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	entries := make([]LogEntry, len(tl.entries))
	copy(entries, tl.entries)
	return entries
}

// EntriesByLevel 获取指定级别的日志条目
func (tl *TestLogger) EntriesByLevel(level Level) []LogEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var result []LogEntry
	for _, entry := range tl.entries {
		if entry.Level == level {
			result = append(result, entry)
		}
	}
	return result
}

// EntriesByMessage 获取包含指定消息的日志条目
func (tl *TestLogger) EntriesByMessage(message string) []LogEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	var result []LogEntry
	for _, entry := range tl.entries {
		if strings.Contains(entry.Message, message) {
			result = append(result, entry)
		}
	}
	return result
}

// HasMessage 检查是否包含指定消息的日志
func (tl *TestLogger) HasMessage(message string) bool {
	return len(tl.EntriesByMessage(message)) > 0
}

// HasLevel 检查是否包含指定级别的日志
func (tl *TestLogger) HasLevel(level Level) bool {
	return len(tl.EntriesByLevel(level)) > 0
}

// Count 统计日志条目总数
func (tl *TestLogger) Count() int {
	tl.mu.RLock()
	defer tl.mu.RUnlock()
	return len(tl.entries)
}

// CountByLevel 统计指定级别的日志条目数量
func (tl *TestLogger) CountByLevel(level Level) int {
	return len(tl.EntriesByLevel(level))
}

// Clear 清空所有日志条目
func (tl *TestLogger) Clear() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.entries = tl.entries[:0]
}

// PrintAllLogs 打印所有日志条目（用于测试失败时输出）
func (tl *TestLogger) PrintAllLogs(t *testing.T) {
	entries := tl.Entries()
	if len(entries) == 0 {
		t.Logf("No log entries captured")
		return
	}

	t.Logf("=== Captured Log Entries (%d) ===", len(entries))
	for i, entry := range entries {
		t.Logf("[%d] %s [%s] %s", i+1, entry.Time.Format("15:04:05.000"),
			strings.ToUpper(string(entry.Level)), entry.Message)

		// 打印字段
		for _, field := range entry.Fields {
			t.Logf("    %s", field.Key)
		}
	}
	t.Logf("=== End Log Entries ===")
}

// LastEntry 获取最后一条日志条目
func (tl *TestLogger) LastEntry() *LogEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	if len(tl.entries) == 0 {
		return nil
	}
	return &tl.entries[len(tl.entries)-1]
}

// FirstEntry 获取第一条日志条目
func (tl *TestLogger) FirstEntry() *LogEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	if len(tl.entries) == 0 {
		return nil
	}
	return &tl.entries[0]
}

// LogCapture 日志捕获工具，可以临时替换全局日志器并捕获日志
type LogCapture struct {
	originalLogger Logger
	testLogger     *TestLogger
}

// NewLogCapture 创建新的日志捕获器
func NewLogCapture() *LogCapture {
	testLogger := NewTestLogger()

	return &LogCapture{
		testLogger: testLogger,
	}
}

// Start 开始捕获日志
func (lc *LogCapture) Start() {
	// 保存原始全局日志器
	lc.originalLogger = Default()

	// 创建一个包装器来避免类型冲突
	wrapper := &loggerWrapper{
		testLogger: lc.testLogger,
	}

	// 直接存储包装器，不调用ResetGlobal以避免类型冲突
	globalLogger.Store(wrapper)
	atomic.StoreInt32(&globalInitialized, 1)
}

// Stop 停止捕获日志并恢复原始日志器
func (lc *LogCapture) Stop() {
	if lc.originalLogger != nil {
		globalLogger.Store(lc.originalLogger)
	}
}

// Entries 获取捕获的日志条目
func (lc *LogCapture) Entries() []LogEntry {
	return lc.testLogger.Entries()
}

// TestLogger 获取底层的测试日志器
func (lc *LogCapture) TestLogger() *TestLogger {
	return lc.testLogger
}

// Clear 清空捕获的日志
func (lc *LogCapture) Clear() {
	lc.testLogger.Clear()
}

// NewNoopLogger 创建无操作日志器（用于禁用日志）
func NewNoopLogger() Logger {
	return &noopLogger{}
}

// noopLogger 无操作日志器实现
type noopLogger struct{}

func (nl *noopLogger) Debug(msg string, fields ...Field)      {}
func (nl *noopLogger) Info(msg string, fields ...Field)       {}
func (nl *noopLogger) Warn(msg string, fields ...Field)       {}
func (nl *noopLogger) Error(msg string, fields ...Field)      {}
func (nl *noopLogger) Fatal(msg string, fields ...Field)      {}
func (nl *noopLogger) With(fields ...Field) Logger            { return nl }
func (nl *noopLogger) WithContext(ctx context.Context) Logger { return nl }
func (nl *noopLogger) Sync() error                            { return nil }

// AssertLogContains 断言日志包含指定消息（测试辅助函数）
func AssertLogContains(t *testing.T, logger *TestLogger, message string) {
	if !logger.HasMessage(message) {
		entries := logger.Entries()
		t.Errorf("Expected log to contain message: %s\nActual logs (%d):", message, len(entries))
		for _, entry := range entries {
			t.Errorf("  [%s] %s", entry.Level, entry.Message)
		}
	}
}

// AssertLogHasLevel 断言日志包含指定级别（测试辅助函数）
func AssertLogHasLevel(t *testing.T, logger *TestLogger, level Level) {
	if !logger.HasLevel(level) {
		t.Errorf("Expected log to have level: %s", level)
	}
}

// AssertLogCount 断言日志条目数量（测试辅助函数）
func AssertLogCount(t *testing.T, logger *TestLogger, expectedCount int) {
	actualCount := logger.Count()
	if actualCount != expectedCount {
		t.Errorf("Expected %d log entries, got %d", expectedCount, actualCount)
	}
}

// AssertLogEmpty 断言日志为空（测试辅助函数）
func AssertLogEmpty(t *testing.T, logger *TestLogger) {
	if !logger.Empty() {
		t.Errorf("Expected no log entries, but got %d", logger.Count())
	}
}

// Empty 检查日志是否为空
func (tl *TestLogger) Empty() bool {
	return tl.Count() == 0
}

// contextTestLogger 带上下文的测试日志器
type contextTestLogger struct {
	*TestLogger
	ctx    context.Context
	fields []Field
}

// Debug 记录调试级别日志（带上下文）
func (ctl *contextTestLogger) Debug(msg string, fields ...Field) {
	allFields := append(ctl.fields, fields...)
	ctl.TestLogger.Debug(msg, allFields...)
}

// Info 记录信息级别日志（带上下文）
func (ctl *contextTestLogger) Info(msg string, fields ...Field) {
	allFields := append(ctl.fields, fields...)
	ctl.TestLogger.Info(msg, allFields...)
}

// Warn 记录警告级别日志（带上下文）
func (ctl *contextTestLogger) Warn(msg string, fields ...Field) {
	allFields := append(ctl.fields, fields...)
	ctl.TestLogger.Warn(msg, allFields...)
}

// Error 记录错误级别日志（带上下文）
func (ctl *contextTestLogger) Error(msg string, fields ...Field) {
	allFields := append(ctl.fields, fields...)
	ctl.TestLogger.Error(msg, allFields...)
}

// Fatal 记录致命错误级别日志（带上下文）
func (ctl *contextTestLogger) Fatal(msg string, fields ...Field) {
	allFields := append(ctl.fields, fields...)
	ctl.TestLogger.Fatal(msg, allFields...)
}

// With 创建带有额外字段的子日志器（带上下文）
func (ctl *contextTestLogger) With(fields ...Field) Logger {
	allFields := append(ctl.fields, fields...)
	return &contextTestLogger{
		TestLogger: ctl.TestLogger,
		ctx:        ctl.ctx,
		fields:     allFields,
	}
}

// WithContext 创建带有新上下文的日志器
func (ctl *contextTestLogger) WithContext(ctx context.Context) Logger {
	contextFields := extractContextFields(ctx)
	allFields := append(ctl.fields, contextFields...)
	return &contextTestLogger{
		TestLogger: ctl.TestLogger,
		ctx:        ctx,
		fields:     allFields,
	}
}

// Sync 同步缓冲区（测试日志器不需要同步）
func (ctl *contextTestLogger) Sync() error {
	return ctl.TestLogger.Sync()
}

// testLoggerWithFields 带字段的测试日志器
type testLoggerWithFields struct {
	*TestLogger
	fields []Field
}

// Debug 记录调试级别日志（带字段）
func (tlwf *testLoggerWithFields) Debug(msg string, fields ...Field) {
	allFields := append(tlwf.fields, fields...)
	tlwf.TestLogger.Debug(msg, allFields...)
}

// Info 记录信息级别日志（带字段）
func (tlwf *testLoggerWithFields) Info(msg string, fields ...Field) {
	allFields := append(tlwf.fields, fields...)
	tlwf.TestLogger.Info(msg, allFields...)
}

// Warn 记录警告级别日志（带字段）
func (tlwf *testLoggerWithFields) Warn(msg string, fields ...Field) {
	allFields := append(tlwf.fields, fields...)
	tlwf.TestLogger.Warn(msg, allFields...)
}

// Error 记录错误级别日志（带字段）
func (tlwf *testLoggerWithFields) Error(msg string, fields ...Field) {
	allFields := append(tlwf.fields, fields...)
	tlwf.TestLogger.Error(msg, allFields...)
}

// Fatal 记录致命错误级别日志（带字段）
func (tlwf *testLoggerWithFields) Fatal(msg string, fields ...Field) {
	allFields := append(tlwf.fields, fields...)
	tlwf.TestLogger.Fatal(msg, allFields...)
}

// With 创建带有额外字段的子日志器（带字段）
func (tlwf *testLoggerWithFields) With(fields ...Field) Logger {
	allFields := append(tlwf.fields, fields...)
	return &testLoggerWithFields{
		TestLogger: tlwf.TestLogger,
		fields:     allFields,
	}
}

// WithContext 创建带有上下文的日志器（带字段）
func (tlwf *testLoggerWithFields) WithContext(ctx context.Context) Logger {
	contextFields := extractContextFields(ctx)
	allFields := append(tlwf.fields, contextFields...)
	return &contextTestLogger{
		TestLogger: tlwf.TestLogger,
		ctx:        ctx,
		fields:     allFields,
	}
}

// Sync 同步缓冲区（测试日志器不需要同步）
func (tlwf *testLoggerWithFields) Sync() error {
	return tlwf.TestLogger.Sync()
}

// loggerWrapper 日志器包装器，用于避免atomic.Value类型冲突
type loggerWrapper struct {
	testLogger *TestLogger
}

// Debug 记录调试级别日志
func (lw *loggerWrapper) Debug(msg string, fields ...Field) {
	lw.testLogger.Debug(msg, fields...)
}

// Info 记录信息级别日志
func (lw *loggerWrapper) Info(msg string, fields ...Field) {
	lw.testLogger.Info(msg, fields...)
}

// Warn 记录警告级别日志
func (lw *loggerWrapper) Warn(msg string, fields ...Field) {
	lw.testLogger.Warn(msg, fields...)
}

// Error 记录错误级别日志
func (lw *loggerWrapper) Error(msg string, fields ...Field) {
	lw.testLogger.Error(msg, fields...)
}

// Fatal 记录致命错误级别日志
func (lw *loggerWrapper) Fatal(msg string, fields ...Field) {
	lw.testLogger.Fatal(msg, fields...)
}

// With 创建带有额外字段的子日志器
func (lw *loggerWrapper) With(fields ...Field) Logger {
	return &testLoggerWithFields{
		TestLogger: lw.testLogger,
		fields:     fields,
	}
}

// WithContext 创建带有上下文的日志器
func (lw *loggerWrapper) WithContext(ctx context.Context) Logger {
	contextFields := extractContextFields(ctx)
	return &contextTestLogger{
		TestLogger: lw.testLogger,
		ctx:        ctx,
		fields:     contextFields,
	}
}

// Sync 同步缓冲区
func (lw *loggerWrapper) Sync() error {
	return lw.testLogger.Sync()
}
