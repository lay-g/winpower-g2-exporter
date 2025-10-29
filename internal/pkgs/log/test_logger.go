package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogEntry 日志条目
type LogEntry struct {
	Level   zapcore.Level
	Message string
	Fields  []Field
	Context context.Context
}

// testLoggerData 测试日志器共享数据
type testLoggerData struct {
	mu      sync.Mutex
	entries []LogEntry
}

// TestLogger 测试日志器，用于捕获和验证日志
type TestLogger struct {
	data *testLoggerData
	ctx  context.Context
}

// NewTestLogger 创建测试日志器
func NewTestLogger() *TestLogger {
	return &TestLogger{
		data: &testLoggerData{
			entries: make([]LogEntry, 0),
		},
		ctx: context.Background(),
	}
}

// NewNoopLogger 创建空操作日志器
func NewNoopLogger() Logger {
	return &zapLogger{
		logger: zap.NewNop(),
	}
}

// Debug 记录调试日志
func (t *TestLogger) Debug(msg string, fields ...Field) {
	t.log(zapcore.DebugLevel, msg, fields...)
}

// Info 记录信息日志
func (t *TestLogger) Info(msg string, fields ...Field) {
	t.log(zapcore.InfoLevel, msg, fields...)
}

// Warn 记录警告日志
func (t *TestLogger) Warn(msg string, fields ...Field) {
	t.log(zapcore.WarnLevel, msg, fields...)
}

// Error 记录错误日志
func (t *TestLogger) Error(msg string, fields ...Field) {
	t.log(zapcore.ErrorLevel, msg, fields...)
}

// Fatal 记录致命错误日志
func (t *TestLogger) Fatal(msg string, fields ...Field) {
	t.log(zapcore.FatalLevel, msg, fields...)
}

// With 创建带有预设字段的子日志器
func (t *TestLogger) With(fields ...Field) Logger {
	// 共享同一个 data 结构的引用
	return &TestLogger{
		data: t.data,
		ctx:  t.ctx,
	}
}

// WithContext 创建带有上下文的日志器
func (t *TestLogger) WithContext(ctx context.Context) Logger {
	// 共享同一个 data 结构的引用
	return &TestLogger{
		data: t.data,
		ctx:  ctx,
	}
}

// Sync 刷新缓冲区（空实现）
func (t *TestLogger) Sync() error {
	return nil
}

// Core 返回空的 zapcore.Core（测试用）
func (t *TestLogger) Core() zapcore.Core {
	return zapcore.NewNopCore()
}

// log 内部日志记录方法
func (t *TestLogger) log(level zapcore.Level, msg string, fields ...Field) {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	entry := LogEntry{
		Level:   level,
		Message: msg,
		Fields:  fields,
		Context: t.ctx,
	}
	t.data.entries = append(t.data.entries, entry)
}

// Entries 获取所有日志条目
func (t *TestLogger) Entries() []LogEntry {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	// 返回副本以避免并发问题
	entries := make([]LogEntry, len(t.data.entries))
	copy(entries, t.data.entries)
	return entries
}

// EntriesByLevel 按日志级别过滤条目
func (t *TestLogger) EntriesByLevel(level zapcore.Level) []LogEntry {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	var filtered []LogEntry
	for _, entry := range t.data.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// EntriesByMessage 按消息过滤条目
func (t *TestLogger) EntriesByMessage(message string) []LogEntry {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	var filtered []LogEntry
	for _, entry := range t.data.entries {
		if entry.Message == message {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// EntriesByField 按字段过滤条目
func (t *TestLogger) EntriesByField(key string, value interface{}) []LogEntry {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	var filtered []LogEntry
	for _, entry := range t.data.entries {
		if hasField(entry.Fields, key, value) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// Clear 清除所有日志条目
func (t *TestLogger) Clear() {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	t.data.entries = make([]LogEntry, 0)
}

// Count 返回日志条目数量
func (t *TestLogger) Count() int {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	return len(t.data.entries)
}

// HasEntry 检查是否存在指定的日志条目
func (t *TestLogger) HasEntry(level zapcore.Level, message string, fields map[string]interface{}) bool {
	t.data.mu.Lock()
	defer t.data.mu.Unlock()

	for _, entry := range t.data.entries {
		if entry.Level != level {
			continue
		}
		if entry.Message != message {
			continue
		}
		if !matchFields(entry.Fields, fields) {
			continue
		}
		return true
	}
	return false
}

// LogCapture 日志捕获器，可用于捕获任何日志器的输出
type LogCapture struct {
	mu      sync.Mutex
	entries []LogEntry
	ctx     context.Context
}

// NewLogCapture 创建日志捕获器
func NewLogCapture() *LogCapture {
	return &LogCapture{
		entries: make([]LogEntry, 0),
		ctx:     context.Background(),
	}
}

// Capture 创建写入到此捕获器的日志器
func (c *LogCapture) Capture() Logger {
	return &captureLogger{
		capture: c,
	}
}

// WithContext 创建上下文感知的日志器
func (c *LogCapture) WithContext(ctx context.Context) Logger {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ctx = ctx
	return c.Capture()
}

// Entries 获取所有捕获的日志条目
func (c *LogCapture) Entries() []LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries := make([]LogEntry, len(c.entries))
	copy(entries, c.entries)
	return entries
}

// Clear 清除所有捕获的日志条目
func (c *LogCapture) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make([]LogEntry, 0)
}

// captureLogger 捕获日志器实现
type captureLogger struct {
	capture *LogCapture
	fields  []Field
	ctx     context.Context
}

// Debug 记录调试日志
func (l *captureLogger) Debug(msg string, fields ...Field) {
	l.log(zapcore.DebugLevel, msg, fields...)
}

// Info 记录信息日志
func (l *captureLogger) Info(msg string, fields ...Field) {
	l.log(zapcore.InfoLevel, msg, fields...)
}

// Warn 记录警告日志
func (l *captureLogger) Warn(msg string, fields ...Field) {
	l.log(zapcore.WarnLevel, msg, fields...)
}

// Error 记录错误日志
func (l *captureLogger) Error(msg string, fields ...Field) {
	l.log(zapcore.ErrorLevel, msg, fields...)
}

// Fatal 记录致命错误日志
func (l *captureLogger) Fatal(msg string, fields ...Field) {
	l.log(zapcore.FatalLevel, msg, fields...)
}

// With 创建带有预设字段的子日志器
func (l *captureLogger) With(fields ...Field) Logger {
	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	return &captureLogger{
		capture: l.capture,
		fields:  allFields,
		ctx:     l.ctx,
	}
}

// WithContext 创建带有上下文的日志器
func (l *captureLogger) WithContext(ctx context.Context) Logger {
	return &captureLogger{
		capture: l.capture,
		fields:  l.fields,
		ctx:     ctx,
	}
}

// Sync 刷新缓冲区（空实现）
func (l *captureLogger) Sync() error {
	return nil
}

// Core 返回空的 zapcore.Core（测试用）
func (l *captureLogger) Core() zapcore.Core {
	return zapcore.NewNopCore()
}

// log 内部日志记录方法
func (l *captureLogger) log(level zapcore.Level, msg string, fields ...Field) {
	l.capture.mu.Lock()
	defer l.capture.mu.Unlock()

	allFields := make([]Field, 0, len(l.fields)+len(fields))
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	ctx := l.ctx
	if ctx == nil {
		ctx = l.capture.ctx
	}

	entry := LogEntry{
		Level:   level,
		Message: msg,
		Fields:  allFields,
		Context: ctx,
	}
	l.capture.entries = append(l.capture.entries, entry)
}

// hasField 检查字段列表中是否包含指定的键值对
func hasField(fields []Field, key string, value interface{}) bool {
	for _, field := range fields {
		if field.Key == key {
			// 简单比较，实际使用中可能需要更复杂的比较逻辑
			switch v := value.(type) {
			case string:
				if field.String == v {
					return true
				}
			case int:
				if field.Integer == int64(v) {
					return true
				}
			case int64:
				if field.Integer == v {
					return true
				}
			case bool:
				if field.Integer == 0 && !v {
					return true
				}
				if field.Integer == 1 && v {
					return true
				}
			}
		}
	}
	return false
}

// matchFields 检查字段列表是否匹配指定的字段映射
func matchFields(fields []Field, expected map[string]interface{}) bool {
	if len(expected) == 0 {
		return true
	}

	for key, value := range expected {
		if !hasField(fields, key, value) {
			return false
		}
	}
	return true
}
