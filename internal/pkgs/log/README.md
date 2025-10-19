# Log Package

基于 zap 的高性能日志模块，提供统一的日志接口和结构化日志输出能力。

## 特性

- 🚀 高性能：基于 zap 的零内存分配日志库
- 📝 结构化日志：支持 JSON 和 Console 格式
- 🎯 多级别支持：Debug, Info, Warn, Error, Fatal
- 📁 多种输出：标准输出、文件输出、多目标输出
- 🔄 日志轮转：支持文件大小、时间、数量限制的日志轮转
- 🏷️ 上下文感知：支持从 context 自动提取追踪信息
- 🧪 测试友好：提供专用的测试日志器和日志捕获工具
- ⚙️ 灵活配置：支持多种配置方式和默认值

## 快速开始

### 基本使用

```go
package main

import (
    "github.com/your-project/internal/pkgs/log"
)

func main() {
    // 初始化全局日志器
    if err := log.Init(log.DefaultConfig()); err != nil {
        panic(err)
    }
    defer log.Sync()

    // 记录日志
    log.Info("Application started")
    log.Error("Something went wrong", log.Error(err))

    // 带字段的日志
    log.Info("User login",
        log.String("user_id", "12345"),
        log.String("ip", "192.168.1.1"))
}
```

### 自定义配置

```go
config := &log.Config{
    Level:       log.InfoLevel,
    Format:      log.JSONFormat,
    Output:      log.BothOutput,
    Filename:    "/var/log/app.log",
    MaxSize:     100,  // 100MB
    MaxAge:      30,   // 30 days
    MaxBackups:  10,   // 10 files
    Compress:    true,
    EnableCaller: true,
}

logger, err := log.NewLogger(config)
if err != nil {
    panic(err)
}

logger.Info("Custom logger initialized")
```

### 上下文日志

```go
import "context"

func handleRequest(ctx context.Context) {
    // 在上下文中设置追踪信息
    ctx = log.WithRequestID(ctx, "req-123")
    ctx = log.WithUserID(ctx, "user-456")

    // 使用上下文日志
    log.InfoWithContext(ctx, "Processing request")

    // 或者创建上下文日志器
    logger := log.WithContext(ctx)
    logger.Info("Another log entry")
}
```

## 配置选项

### 日志级别

- `debug` - 调试信息
- `info` - 一般信息（默认）
- `warn` - 警告信息
- `error` - 错误信息
- `fatal` - 致命错误

### 输出格式

- `json` - JSON 格式（默认）
- `console` - 控制台友好格式

### 输出目标

- `stdout` - 标准输出
- `stderr` - 标准错误
- `file` - 文件输出
- `both` - 同时输出到标准输出和文件

## 测试支持

### 测试日志器

```go
func TestSomething(t *testing.T) {
    logger := log.NewTestLoggerWithT(t)

    logger.Info("Test log entry")
    logger.Error("Test error")

    // 断言日志内容
    log.AssertLogContains(t, logger, "Test log entry")
    log.AssertLogHasLevel(t, logger, log.ErrorLevel)
    log.AssertLogCount(t, logger, 2)
}
```

### 日志捕获

```go
func TestGlobalLogging(t *testing.T) {
    capture := log.NewLogCapture()
    capture.Start()
    defer capture.Stop()

    // 使用全局日志函数
    log.Info("Global log message")

    entries := capture.Entries()
    if len(entries) == 0 {
        t.Error("Expected log entries, got none")
    }
}
```

## 环境变量配置

支持通过环境变量配置日志：

- `LOG_LEVEL` - 日志级别
- `LOG_FORMAT` - 日志格式
- `LOG_OUTPUT` - 输出目标
- `LOG_FILE` - 日志文件路径

```go
// 从环境变量初始化
err := log.InitializeFromEnv()
```

## 性能考虑

- 使用 zap 的高性能实现，避免不必要的内存分配
- 支持异步日志写入
- 日志级别在编译时确定，避免运行时开销
- 字段构造函数零分配

## 最佳实践

1. **使用结构化字段**：使用 `log.String()`, `log.Int()` 等构造函数而不是格式化字符串
2. **合理设置级别**：生产环境通常使用 `info` 或 `warn` 级别
3. **启用上下文追踪**：在分布式系统中使用 request_id 和 trace_id
4. **日志轮转**：生产环境务必配置日志轮转避免磁盘空间耗尽
5. **测试验证**：使用测试日志器验证关键日志是否正确输出

## 示例配置

### 开发环境

```go
config := log.DevelopmentDefaults()
// 相当于：
// &log.Config{
//     Level:       log.DebugLevel,
//     Format:      log.ConsoleFormat,
//     Output:      log.StdoutOutput,
//     EnableColor: true,
//     EnableCaller: true,
// }
```

### 生产环境

```go
config := &log.Config{
    Level:       log.InfoLevel,
    Format:      log.JSONFormat,
    Output:      log.FileOutput,
    Filename:    "/var/log/app.log",
    MaxSize:     100,
    MaxAge:      30,
    MaxBackups:  10,
    Compress:    true,
    EnableCaller: false,
}
```

### 容器环境

```go
config := &log.Config{
    Level:    log.InfoLevel,
    Format:   log.JSONFormat,
    Output:   log.StdoutOutput, // 输出到 stdout，由容器运行时收集
}
```