# 日志模块设计

## 1. 概述

日志模块基于 zap 高性能日志库，提供统一的日志接口和结构化日志输出能力，支持多种日志级别、格式和输出目标。

### 1.1 设计目标

- **高性能**: 零内存分配设计，最小化性能开销
- **结构化**: 支持结构化日志输出，便于日志分析
- **灵活性**: 支持多种输出目标和格式
- **易用性**: 简单的 API 接口
- **可观测性**: 完整的日志追踪能力

### 1.2 功能范围

- 多级别日志（Debug, Info, Warn, Error）
- 结构化字段支持
- 多种输出格式（JSON, Console）
- 多种输出目标（Stdout, Stderr, File, Both）
- **日志轮转和归档**：基于大小和时间的自动轮转
- **上下文日志支持**：自动提取上下文字段
- **全局日志器**：便捷的全局日志函数
- **测试专用日志器**：内存捕获和断言工具
- **多输出目标**：同时输出到多个目标
- **高级配置选项**：调用者信息、堆栈跟踪等

## 2. 技术选型

### 2.1 Zap 日志库

选择 zap 的理由：

| 特性     | 说明                       |
| -------- | -------------------------- |
| 高性能   | 零内存分配，性能优异       |
| 结构化   | 原生支持结构化日志         |
| 类型安全 | 强类型字段，避免运行时错误 |
| 灵活配置 | 支持多种配置方式           |
| 生态完善 | 广泛使用，社区活跃         |


## 3. 架构设计

### 3.1 模块结构

```
internal/pkgs/log/
├── logger.go          # 日志接口和实现
├── config.go          # 日志配置
├── encoder.go         # 编码器配置
├── writer.go          # 输出写入器
├── context.go         # 上下文日志
├── logger_test.go     # 单元测试
└── README.md          # 文档
```

### 3.2 组件关系

```
┌─────────────────────────────────────────────────┐
│                 Application                     │
└────────────────────┬────────────────────────────┘
                     │ log.Info()
                     ▼
┌─────────────────────────────────────────────────┐
│              Logger Interface                   │
│  ┌───────────────────────────────────────────┐  │
│  │  - Debug(msg, fields...)                  │  │
│  │  - Info(msg, fields...)                   │  │
│  │  - Warn(msg, fields...)                   │  │
│  │  - Error(msg, fields...)                  │  │
│  │  - With(fields...) Logger                 │  │
│  └───────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────┐
│              Zap Logger Core                    │
│  ┌───────────────┬──────────────┬────────────┐  │
│  │   Encoder     │    Writer    │   Level    │  │
│  │  (JSON/Console) │ (File/Stdout)│ (Info/Debug)│ │
│  └───────────────┴──────────────┴────────────┘  │
└─────────────────────────────────────────────────┘
```

## 4. 数据结构设计

### 4.1 日志配置

日志配置结构体定义在 `internal/pkgs/log/config.go` 中，通过顶层配置结构体引用：

```go
package log

import (
    "fmt"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Config 日志配置
type Config struct {
    // Level 日志级别: debug, info, warn, error
    Level string `json:"level" yaml:"level" mapstructure:"level"`

    // Format 日志格式: json, console
    Format string `json:"format" yaml:"format" mapstructure:"format"`

    // Output 输出目标: stdout, stderr, file, both
    Output string `json:"output" yaml:"output" mapstructure:"output"`

    // FilePath 文件路径（Output 为 file 或 both 时使用）
    FilePath string `json:"file_path" yaml:"file_path" mapstructure:"file_path"`

    // MaxSize 单个日志文件最大大小（MB）
    MaxSize int `json:"max_size" yaml:"max_size" mapstructure:"max_size"`

    // MaxAge 日志文件保留天数
    MaxAge int `json:"max_age" yaml:"max_age" mapstructure:"max_age"`

    // MaxBackups 最大备份文件数
    MaxBackups int `json:"max_backups" yaml:"max_backups" mapstructure:"max_backups"`

    // Compress 是否压缩旧日志文件
    Compress bool `json:"compress" yaml:"compress" mapstructure:"compress"`

    // Development 是否开发模式
    Development bool `json:"development" yaml:"development" mapstructure:"development"`

    // EnableCaller 是否记录调用者信息
    EnableCaller bool `json:"enable_caller" yaml:"enable_caller" mapstructure:"enable_caller"`

    // EnableStacktrace 是否记录堆栈跟踪
    EnableStacktrace bool `json:"enable_stacktrace" yaml:"enable_stacktrace" mapstructure:"enable_stacktrace"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
    return &Config{
        Level:            "info",
        Format:           "json",
        Output:           "stdout",
        FilePath:         "",
        MaxSize:          100,
        MaxAge:           7,
        MaxBackups:       3,
        Compress:         true,
        Development:      false,
        EnableCaller:     false,
        EnableStacktrace: false,
    }
}

// DevelopmentDefaults 返回开发环境默认配置
func DevelopmentDefaults() *Config {
    return &Config{
        Level:            "debug",
        Format:           "console",
        Output:           "stdout",
        FilePath:         "",
        MaxSize:          100,
        MaxAge:           7,
        MaxBackups:       3,
        Compress:         false,
        Development:      true,
        EnableCaller:     true,
        EnableStacktrace: true,
    }
}

// Validate 实现ConfigValidator接口用于配置验证
func (c *Config) Validate() error {
    // 验证日志级别
    validLevels := map[string]bool{
        "debug": true,
        "info":  true,
        "warn":  true,
        "error": true,
    }
    if !validLevels[c.Level] {
        return fmt.Errorf("invalid log level: %s, must be one of debug, info, warn, error", c.Level)
    }

    // 验证日志格式
    validFormats := map[string]bool{
        "json":    true,
        "console": true,
    }
    if !validFormats[c.Format] {
        return fmt.Errorf("invalid log format: %s, must be one of json, console", c.Format)
    }

    // 验证输出目标
    validOutputs := map[string]bool{
        "stdout": true,
        "stderr": true,
        "file":   true,
        "both":   true,
    }
    if !validOutputs[c.Output] {
        return fmt.Errorf("invalid log output: %s, must be one of stdout, stderr, file, both", c.Output)
    }

    // 如果输出到文件，验证文件路径
    if (c.Output == "file" || c.Output == "both") && c.FilePath == "" {
        return fmt.Errorf("file_path is required when output is 'file' or 'both'")
    }

    // 验证文件大小限制
    if c.MaxSize <= 0 {
        return fmt.Errorf("max_size must be positive")
    }

    // 验证保留天数
    if c.MaxAge < 0 {
        return fmt.Errorf("max_age cannot be negative")
    }

    // 验证备份数量
    if c.MaxBackups < 0 {
        return fmt.Errorf("max_backups cannot be negative")
    }

    return nil
}
```

### 4.2 日志接口

```go
// Logger 日志接口
type Logger interface {
    // Debug 调试级别日志
    Debug(msg string, fields ...Field)

    // Info 信息级别日志
    Info(msg string, fields ...Field)

    // Warn 警告级别日志
    Warn(msg string, fields ...Field)

    // Error 错误级别日志
    Error(msg string, fields ...Field)

    // Fatal 致命错误日志（会调用 os.Exit(1)）
    Fatal(msg string, fields ...Field)

    // With 创建带有预设字段的子日志器
    With(fields ...Field) Logger

    // WithContext 创建带有上下文的日志器
    WithContext(ctx context.Context) Logger

    // Sync 刷新缓冲区
    Sync() error
}

// Field 构造器函数
func String(key, value string) Field
func Int(key string, value int) Field
func Int64(key string, value int64) Field
func Float64(key string, value float64) Field
func Bool(key string, value bool) Field
func Error(err error) Field
func Any(key string, value interface{}) Field

// zapLogger zap 日志器实现
type zapLogger struct {
    zap *zap.Logger
}
```

### 4.3 上下文日志支持

```go
// 上下文字段键定义
const (
    RequestIDKey    contextKey = "request_id"
    TraceIDKey      contextKey = "trace_id"
    UserIDKey       contextKey = "user_id"
    CorrelationIDKey contextKey = "correlation_id"
    ComponentKey    contextKey = "component"
    OperationKey    contextKey = "operation"
)

// 上下文工具函数
func WithRequestID(ctx context.Context, requestID string) context.Context
func WithTraceID(ctx context.Context, traceID string) context.Context
func WithUserID(ctx context.Context, userID string) context.Context
func WithComponent(ctx context.Context, component string) context.Context
func WithOperation(ctx context.Context, operation string) context.Context

// 日志器上下文管理
func WithLogger(ctx context.Context, logger Logger) context.Context
func FromLogger(ctx context.Context) Logger
func LoggerFromContext(ctx context.Context, fallbackLogger Logger) Logger
func FromContext(ctx context.Context) []Field
```

## 5. 实现设计

### 5.1 日志器初始化

```go
// NewLogger 创建新的日志器
func NewLogger(cfg *Config) (Logger, error) {
    // 验证配置参数，为空时使用默认配置
    // 调用 buildZapConfig 构建 zap 配置
    // 使用 zapCfg.Build 创建 zap logger 实例
    // 添加调用者跳过和堆栈跟踪选项
    // 返回封装后的日志器实例
}

// buildZapConfig 构建 zap 配置
func buildZapConfig(cfg *Config) zap.Config {
    // 根据是否为开发模式选择生产或开发配置
    // 解析并设置日志级别，默认为 info 级别
    // 根据格式选择编码器（json 或 console）
    // 配置输出路径（文件或标准输出）
    // 设置错误输出路径和调用者信息
    // 返回构建好的 zap 配置
}
```

### 5.2 日志方法实现

```go
// Debug 调试级别日志
func (l *logger) Debug(msg string, fields ...zap.Field) {
    // 调用底层 zap logger 的 Debug 方法输出日志
}

// Info 信息级别日志
func (l *logger) Info(msg string, fields ...zap.Field) {
    // 调用底层 zap logger 的 Info 方法输出日志
}

// Warn 警告级别日志
func (l *logger) Warn(msg string, fields ...zap.Field) {
    // 调用底层 zap logger 的 Warn 方法输出日志
}

// Error 错误级别日志
func (l *logger) Error(msg string, fields ...zap.Field) {
    // 调用底层 zap logger 的 Error 方法输出日志
}

// Fatal 致命错误日志
func (l *logger) Fatal(msg string, fields ...zap.Field) {
    // 调用底层 zap logger 的 Fatal 方法输出日志并退出程序
}

// With 创建带有预设字段的子日志器
func (l *logger) With(fields ...zap.Field) Logger {
    // 基于当前日志器创建带有预设字段的新日志器
    // 返回封装后的 Logger 接口实例
}

// Sync 刷新缓冲区
func (l *logger) Sync() error {
    // 调用底层 zap logger 的 Sync 方法刷新缓冲区
    // 返回刷新操作的结果或错误
}
```

### 5.3 日志轮转

```go
import (
    "gopkg.in/natefinch/lumberjack.v2"
)

// NewRotatingFileWriter 创建日志轮转写入器
func NewRotatingFileWriter(cfg *Config) io.Writer {
    // 基于 lumberjack 创建轮转文件写入器
    // 配置文件路径、最大大小、保留天数、备份数量等参数
    // 返回配置好的 io.Writer 接口
}

// BuildLoggerWithRotation 构建带轮转的日志器
func BuildLoggerWithRotation(cfg *Config) (Logger, error) {
    // 创建生产环境编码器配置，设置时间和级别编码器
    // 根据格式选择 JSON 或 Console 编码器
    // 创建轮转文件写入器并转换为 WriteSyncer
    // 解析日志级别配置
    // 使用编码器、写入器和级别创建 zapcore
    // 创建带有调用者信息和堆栈跟踪的 zap logger
    // 返回封装后的 Logger 接口实例
}
```

### 5.4 上下文日志

```go
package log

import (
    "context"
    "go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "logger"

// WithContext 将日志器放入上下文
func WithContext(ctx context.Context, logger Logger) context.Context {
    // 使用 context.WithValue 将日志器存储在上下文中
    // 返回包含日志器的新上下文
}

// FromContext 从上下文获取日志器
func FromContext(ctx context.Context) Logger {
    // 从上下文中获取日志器，进行类型断言
    // 如果不存在或类型错误，返回默认日志器
}

// WithContext 创建带有上下文的日志器
func (l *logger) WithContext(ctx context.Context) Logger {
    // 从上下文中提取追踪信息字段
    // 如果有字段则创建带有这些字段的子日志器
    // 否则返回当前日志器
}

// extractContextFields 从上下文提取字段
func extractContextFields(ctx context.Context) []zap.Field {
    // 声明字段切片用于存储提取的字段
    // 尝试提取 request_id 并添加到字段列表
    // 尝试提取 user_id 并添加到字段列表
    // 尝试提取 trace_id 并添加到字段列表
    // 返回提取的字段列表
}
```

## 6. 全局日志器

### 6.1 全局实例

```go
var (
    // defaultLogger 默认全局日志器
    defaultLogger Logger

    // once 确保只初始化一次
    once sync.Once
)

// Init 初始化全局日志器
func Init(cfg Config) error {
    // 使用 sync.Once 确保只初始化一次
    // 调用 NewLogger 创建日志器实例
    // 返回初始化过程中的错误
}

// InitDevelopment 初始化开发环境全局日志器
func InitDevelopment(cfg Config) error {
    // 创建开发环境配置
    // 初始化全局日志器
    // 返回初始化过程中的错误
}

// Default 获取默认日志器
func Default() Logger {
    // 检查全局日志器是否已初始化
    // 如果未初始化则使用默认配置创建
    // 返回全局日志器实例
}

// ResetGlobal 重置全局日志器（主要用于测试）
func ResetGlobal()

// 全局日志函数
func Debug(msg string, fields ...Field) {
    // 调用默认日志器的 Debug 方法
}

func Info(msg string, fields ...Field) {
    // 调用默认日志器的 Info 方法
}

func Warn(msg string, fields ...Field) {
    // 调用默认日志器的 Warn 方法
}

func Error(msg string, fields ...Field) {
    // 调用默认日志器的 Error 方法
}

func Fatal(msg string, fields ...Field) {
    // 调用默认日志器的 Fatal 方法
}

// 全局上下文日志函数
func DebugWithContext(ctx context.Context, msg string, fields ...Field) {
    // 使用上下文创建日志器并记录调试日志
}

func InfoWithContext(ctx context.Context, msg string, fields ...Field) {
    // 使用上下文创建日志器并记录信息日志
}

func WarnWithContext(ctx context.Context, msg string, fields ...Field) {
    // 使用上下文创建日志器并记录警告日志
}

func ErrorWithContext(ctx context.Context, msg string, fields ...Field) {
    // 使用上下文创建日志器并记录错误日志
}

// 全局 With 函数
func With(fields ...Field) Logger {
    // 返回带有预设字段的子日志器
}

func WithContext(ctx context.Context) Logger {
    // 返回上下文感知的日志器
}

// Sync 刷新全局日志器缓冲区
func Sync() error {
    // 检查全局日志器是否存在
    // 如果存在则调用其 Sync 方法刷新缓冲区
    // 返回刷新结果或 nil
}
```

## 6.2 测试专用日志器

### 6.2.1 TestLogger

```go
// TestLogger 测试日志器，用于捕获和验证日志
type TestLogger struct {
    mu      sync.Mutex
    entries []LogEntry
}

// LogEntry 日志条目
type LogEntry struct {
    Level   string
    Message string
    Fields  []Field
    Context context.Context
}

// NewTestLogger 创建测试日志器
func NewTestLogger() *TestLogger

// NewTestLoggerWithT 创建测试日志器并自动清理
func NewTestLoggerWithT(t *testing.T) *TestLogger

// NewNoopLogger 创建空操作日志器
func NewNoopLogger() Logger

// 日志条目查询方法
func (t *TestLogger) Entries() []LogEntry
func (t *TestLogger) EntriesByLevel(level string) []LogEntry
func (t *TestLogger) EntriesByMessage(message string) []LogEntry
func (t *TestLogger) EntriesByField(key string, value interface{}) []LogEntry
func (t *TestLogger) Clear()
func (t *TestLogger) Count() int
func (t *TestLogger) HasEntry(level, message string, fields map[string]interface{}) bool
```

### 6.2.2 LogCapture

```go
// LogCapture 日志捕获器，可用于捕获任何日志器的输出
type LogCapture struct {
    mu      sync.Mutex
    entries []LogEntry
}

// NewLogCapture 创建日志捕获器
func NewLogCapture() *LogCapture

// Capture 创建写入到此捕获器的日志器
func (c *LogCapture) Capture() Logger

// WithContext 创建上下文感知的日志器
func (c *LogCapture) WithContext(ctx context.Context) Logger

// Entries 获取所有捕获的日志条目
func (c *LogCapture) Entries() []LogEntry

// Clear 清除所有捕获的日志条目
func (c *LogCapture) Clear()
```

## 7. 字段类型

### 7.1 常用字段

```go
import "go.uber.org/zap"

// 使用 zap 提供的类型安全字段
func Example() {
    log.Info("User logged in",
        zap.String("user_id", "12345"),
        zap.String("username", "admin"),
        zap.Int("attempt", 1),
        zap.Duration("latency", time.Millisecond*100),
        zap.Bool("success", true),
        zap.Time("timestamp", time.Now()),
        zap.Error(err),
    )
}
```

### 7.2 自定义字段

```go
// DeviceField 设备信息字段
func DeviceField(device *Device) zap.Field {
    // 创建 zap.Object 字段用于设备信息
    // 在编码器中添加设备 ID、名称、类型等信息
    // 返回结构化的设备字段
}

// HTTPRequestField HTTP 请求字段
func HTTPRequestField(req *http.Request) zap.Field {
    // 创建 zap.Object 字段用于 HTTP 请求信息
    // 在编码器中添加请求方法、URL、远程地址等信息
    // 返回结构化的请求字段
}
```

## 8. 使用示例

### 8.1 基本使用

```go
package main

import (
    "github.com/your-org/winpower-g2-exporter/internal/pkgs/log"
    "go.uber.org/zap"
)

func main() {
    // 初始化日志
    cfg := log.DefaultConfig()
    cfg.Level = "debug"
    cfg.Format = "json"
    
    if err := log.Init(cfg); err != nil {
        panic(err)
    }
    defer log.Sync()
    
    // 使用日志
    log.Info("Application started",
        zap.String("version", "1.0.0"),
        zap.Int("port", 9090),
    )
    
    log.Debug("Debug information",
        zap.String("module", "collector"),
    )
    
    log.Error("Failed to connect",
        zap.String("host", "192.168.1.100"),
        zap.Error(err),
    )
}
```

### 8.2 子日志器

```go
// 创建模块专用日志器
collectorLogger := log.Default().With(
    zap.String("module", "collector"),
    zap.String("component", "device_fetcher"),
)

collectorLogger.Info("Fetching device data")

// 进一步细化
deviceLogger := collectorLogger.With(
    zap.String("device_id", "device-123"),
)

deviceLogger.Debug("Device data received")
```

### 8.3 上下文日志

```go
func HandleRequest(ctx context.Context) {
    // 从上下文获取日志器
    logger := log.FromContext(ctx)
    
    logger.Info("Processing request")
    
    // 添加更多上下文信息
    logger = logger.With(
        zap.String("operation", "fetch_devices"),
    )
    
    logger.Debug("Calling WinPower API")
}
```

### 8.4 错误日志

```go
func FetchData() error {
    resp, err := http.Get("http://api.example.com/data")
    if err != nil {
        log.Error("HTTP request failed",
            zap.String("url", "http://api.example.com/data"),
            zap.Error(err),
            zap.Stack("stacktrace"),
        )
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        log.Warn("Unexpected status code",
            zap.Int("status", resp.StatusCode),
            zap.String("status_text", resp.Status),
        )
    }
    
    return nil
}
```

## 9. 日志格式

### 9.1 JSON 格式

```json
{
  "level": "info",
  "ts": "2025-10-16T10:30:00.123Z",
  "caller": "collector/fetcher.go:45",
  "msg": "Device data fetched successfully",
  "device_id": "device-123",
  "device_name": "UPS-01",
  "duration_ms": 150,
  "record_count": 10
}
```

### 9.2 Console 格式

```
2025-10-16T10:30:00.123Z    INFO    collector/fetcher.go:45    Device data fetched successfully
    {"device_id": "device-123", "device_name": "UPS-01", "duration_ms": 150, "record_count": 10}
```

## 10. 性能优化

### 10.1 避免字符串拼接

```go
// 不推荐
log.Info("User " + username + " logged in from " + ip)

// 推荐
log.Info("User logged in",
    zap.String("username", username),
    zap.String("ip", ip),
)
```

### 10.2 使用类型安全字段

```go
// 不推荐 - 反射开销
log.Info("Request processed", zap.Any("duration", duration))

// 推荐 - 零分配
log.Info("Request processed", zap.Duration("duration", duration))
```

### 10.3 条件日志

```go
// 避免不必要的字符串格式化
if log.Default().Core().Enabled(zapcore.DebugLevel) {
    expensiveDebugInfo := computeExpensiveDebugInfo()
    log.Debug("Debug info", zap.String("info", expensiveDebugInfo))
}
```

## 11. 测试支持

### 11.1 测试日志器

```go
package log

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zaptest"
    "testing"
)

// NewTestLogger 创建测试日志器
func NewTestLogger(t *testing.T) Logger {
    // 基于 zaptest 创建测试专用的日志器
    // 返回封装后的 Logger 接口实例
}

// NewNoopLogger 创建空日志器（不输出任何内容）
func NewNoopLogger() Logger {
    // 创建 zap 的空操作日志器
    // 返回不输出任何内容的 Logger 接口实例
}
```

### 11.2 单元测试

```go
func TestLogger(t *testing.T) {
    // 创建内存缓冲区
    buf := &bytes.Buffer{}
    
    // 创建测试配置
    cfg := &Config{
        Level:  "debug",
        Format: "json",
        Output: "stdout",
    }
    
    logger, err := NewLogger(cfg)
    assert.NoError(t, err)
    
    // 写入日志
    logger.Info("test message",
        zap.String("key", "value"),
    )
    
    // 验证输出
    assert.Contains(t, buf.String(), "test message")
    assert.Contains(t, buf.String(), "\"key\":\"value\"")
}
```

## 12. 最佳实践

### 12.1 日志级别使用

| 级别  | 使用场景                             |
| ----- | ------------------------------------ |
| Debug | 详细的调试信息，开发和调试时使用     |
| Info  | 一般信息，正常的系统运行状态         |
| Warn  | 警告信息，不影响运行但需要注意       |
| Error | 错误信息，功能异常但程序可以继续运行 |
| Fatal | 致命错误，程序无法继续运行           |

### 12.2 结构化字段

- 使用有意义的字段名
- 保持字段名一致性
- 避免敏感信息
- 使用类型安全字段

### 12.3 日志采样

```go
// 高频日志采样，避免日志过多
// 创建采样器配置时间窗口和采样率
// 使用采样器创建新的 logger 实例
```

### 12.4 敏感信息处理

```go
// 不要直接记录敏感信息
log.Info("User authenticated",
    zap.String("username", username),
    zap.String("password", password), // ❌ 不要这样做
)

// 应该脱敏或省略
log.Info("User authenticated",
    zap.String("username", username),
    zap.Bool("has_password", password != ""), // ✅ 正确做法
)
```

补充规则：
- 不记录 `token` 原文、不记录 `Authorization` 头与 Cookie 内容；如需呈现仅记录布尔或剩余有效期等非敏感信息。
- 认证/采集相关日志仅包含非敏感标识（如 `device_id`、`expires_at`、`status`），避免用户名、密码、token 值进入日志。
- 错误日志应进行脱敏与分类，避免输出原始响应中的敏感字段；必要时使用掩码（例如 `token=abc***xyz`）。

## 13. 配置映射

> **说明**：环境变量和命令行参数的具体解析和绑定实现由 `config` 模块统一处理，本章节仅对日志模块支持的配置参数进行说明。

### 13.1 环境变量支持

通过config模块，日志配置支持通过环境变量进行设置，使用 `WINPOWER_EXPORTER_LOGGING_` 前缀：

```bash
# 基本配置
WINPOWER_EXPORTER_LOGGING_LEVEL=debug
WINPOWER_EXPORTER_LOGGING_FORMAT=json
WINPOWER_EXPORTER_LOGGING_OUTPUT=stdout

# 文件输出配置
WINPOWER_EXPORTER_LOGGING_FILE_PATH=/var/log/winpower-exporter.log
WINPOWER_EXPORTER_LOGGING_MAX_SIZE=100
WINPOWER_EXPORTER_LOGGING_MAX_AGE=7
WINPOWER_EXPORTER_LOGGING_MAX_BACKUPS=3
WINPOWER_EXPORTER_LOGGING_COMPRESS=true

# 高级配置
WINPOWER_EXPORTER_LOGGING_DEVELOPMENT=false
WINPOWER_EXPORTER_LOGGING_ENABLE_CALLER=false
WINPOWER_EXPORTER_LOGGING_ENABLE_STACKTRACE=false
```

### 13.2 命令行参数支持

通过config模块，日志配置支持通过命令行参数进行设置：

```bash
# 基本配置
--logging.level debug
--logging.format json
--logging.output stdout

# 文件输出配置
--logging.file_path /var/log/winpower-exporter.log
--logging.max_size 100
--logging.max_age 7
--logging.max_backups 3
--logging.compress true

# 高级配置
--logging.development false
--logging.enable_caller false
--logging.enable_stacktrace false
```

**实现机制**：config模块会自动将环境变量中的下划线（`_`）转换为配置键中的点号（`.`），例如 `WINPOWER_EXPORTER_LOGGING_LEVEL` 对应配置键 `logging.level`。

### 13.3 配置文件示例

在YAML配置文件中，日志配置应使用 `logging` 作为键名：

```yaml
# config.yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: ""
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true
  development: false
  enable_caller: false
  enable_stacktrace: false
```

## 14. 故障排查

### 14.1 日志查询

```bash
# 查询特定级别日志
jq 'select(.level == "error")' log.json

# 查询特定模块日志
jq 'select(.module == "collector")' log.json

# 查询特定时间范围
jq 'select(.ts >= "2025-10-16T00:00:00")' log.json
```

### 14.2 日志分析

```bash
# 统计错误类型
jq -r 'select(.level == "error") | .error' log.json | sort | uniq -c

# 统计请求耗时
jq -r '.duration_ms' log.json | awk '{sum+=$1; count++} END {print sum/count}'
```
