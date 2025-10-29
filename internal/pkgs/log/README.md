# Log Module

高性能、结构化日志模块，基于 zap 实现。

## 特性

- **高性能**: 基于 zap 的零内存分配设计
- **结构化日志**: 类型安全的字段构造
- **多种输出**: 支持 stdout、stderr、文件输出
- **日志轮转**: 集成 lumberjack 实现自动轮转
- **上下文感知**: 自动提取和传播上下文字段
- **测试支持**: 提供测试专用日志器和日志捕获工具

## 快速开始

### 基本使用

```go
import "winpower-g2-exporter/internal/pkgs/log"

// 使用默认配置初始化
log.Init(log.DefaultConfig())

// 记录日志
log.Info("application started", log.String("version", "1.0.0"))
log.Debug("debug info", log.Int("count", 42))
log.Error("error occurred", log.Error(err))
```

### 开发环境配置

```go
// 使用开发环境默认配置
log.InitDevelopment()

// 或者自定义配置
config := log.DefaultConfig()
config.Development = true
config.Level = "debug"
log.Init(config)
```

### 上下文感知日志

```go
import "context"

// 创建带有请求ID的上下文
ctx := log.WithRequestID(context.Background(), "req-123")

// 使用上下文记录日志（自动包含 request_id）
log.InfoContext(ctx, "processing request")
```

### 创建模块日志器

```go
// 为特定模块创建日志器
logger := log.Default().With(
    log.String("module", "collector"),
    log.String("component", "data-fetcher"),
)

logger.Info("module started")
```

## 配置

### 配置结构

```go
type Config struct {
    Level            string  // 日志级别: debug, info, warn, error, fatal
    Format           string  // 输出格式: json, console
    Output           string  // 输出目标: stdout, stderr, file, both
    FilePath         string  // 文件输出路径
    MaxSize          int     // 文件最大大小 (MB)
    MaxAge           int     // 最大保留天数
    MaxBackups       int     // 最大备份文件数
    Compress         bool    // 是否压缩备份文件
    Development      bool    // 开发模式
    EnableCaller     bool    // 是否记录调用位置
    EnableStacktrace bool    // 是否记录堆栈跟踪
}
```

### 默认配置

生产环境：
- Level: info
- Format: json
- Output: stdout
- EnableCaller: false
- EnableStacktrace: false (仅 error 级别)

开发环境：
- Level: debug
- Format: console
- Output: stdout
- EnableCaller: true
- EnableStacktrace: true (error 和 fatal 级别)

## 字段类型

日志模块提供类型安全的字段构造器：

```go
log.String("key", "value")      // 字符串
log.Int("key", 123)              // 整数
log.Int64("key", 123)            // 64位整数
log.Float64("key", 1.23)         // 浮点数
log.Bool("key", true)            // 布尔值
log.Duration("key", time.Second) // 时间间隔
log.Time("key", time.Now())      // 时间戳
log.Error(err)                   // 错误（键名为 "error"）
log.Any("key", value)            // 任意类型（使用反射）
```

## 上下文工具

```go
// 设置上下文字段
ctx = log.WithRequestID(ctx, "req-123")
ctx = log.WithTraceID(ctx, "trace-456")
ctx = log.WithUserID(ctx, "user-789")
ctx = log.WithComponent(ctx, "collector")
ctx = log.WithOperation(ctx, "fetch-data")

// 在上下文中存储日志器
ctx = log.WithLogger(ctx, logger)

// 从上下文获取日志器
logger = log.FromContext(ctx)
```

## 测试支持

### 测试日志器

```go
// 创建测试日志器
testLogger := log.NewTestLogger()

// 执行被测代码
myFunc(testLogger)

// 验证日志输出
entries := testLogger.Entries()
assert.Equal(t, 1, len(entries))
assert.Equal(t, "expected message", entries[0].Message)
assert.Equal(t, log.InfoLevel, entries[0].Level)

// 查询特定日志
infoLogs := testLogger.EntriesWithLevel(log.InfoLevel)
errorLogs := testLogger.EntriesWithLevel(log.ErrorLevel)
```

### 日志捕获

```go
// 捕获现有日志器的输出
capture := log.CaptureLogger(existingLogger)

// 执行被测代码
myFunc(existingLogger)

// 验证捕获的日志
entries := capture.Entries()
assert.Contains(t, entries[0].Message, "expected")
```

### 空日志器

```go
// 创建不输出任何日志的日志器（用于基准测试）
noopLogger := log.NewNoopLogger()
```

## 性能优化

### 条件日志

```go
// 仅在 debug 级别启用时才执行昂贵的计算
if logger.Core().Enabled(zapcore.DebugLevel) {
    expensiveData := computeExpensiveData()
    logger.Debug("expensive debug info", log.String("data", expensiveData))
}
```

### 避免内存分配

```go
// 好的实践：使用类型化字段
logger.Info("user action", log.String("user", userID), log.Int("action", actionType))

// 不好的实践：字符串格式化
logger.Info(fmt.Sprintf("user %s performed action %d", userID, actionType))
```

## 安全注意事项

### 敏感信息保护

不要记录敏感信息：
- ❌ 密码、令牌、密钥
- ❌ 完整的 Authorization 头
- ❌ Cookie 内容
- ❌ 个人身份信息（PII）

可以记录的信息：
- ✅ 用户ID（哈希后）
- ✅ 请求ID、跟踪ID
- ✅ 认证状态（成功/失败）
- ✅ 操作类型和结果

### 信息脱敏

```go
// 使用布尔值替代敏感值
log.Info("authentication", log.Bool("authenticated", true))

// 部分掩码
log.Info("token refresh", log.String("token_prefix", token[:10]+"..."))
```

## 故障排查

### 日志文件权限问题

确保应用有权限写入日志目录：

```bash
chmod 755 /var/log/winpower-g2-exporter
chown app-user:app-group /var/log/winpower-g2-exporter
```

### 日志级别不生效

检查配置加载是否正确：

```go
config := log.DefaultConfig()
config.Level = "debug"
if err := config.Validate(); err != nil {
    log.Fatal("invalid config", log.Error(err))
}
log.Init(config)
```

### 日志未输出

确保在程序退出前调用 Sync：

```go
defer log.Default().Sync()
```

## 最佳实践

1. **使用结构化字段**: 始终使用类型化字段而非字符串格式化
2. **合适的日志级别**: Debug < Info < Warn < Error < Fatal
3. **上下文传播**: 使用 context 传递请求相关信息
4. **模块化日志器**: 为每个模块创建独立的日志器实例
5. **错误处理**: Error 级别用于可恢复错误，Fatal 用于不可恢复错误
6. **性能考虑**: 在性能关键路径使用条件日志
7. **测试验证**: 在测试中验证关键日志输出
8. **安全意识**: 避免记录敏感信息

## 参考

- [设计文档](../../../docs/design/logging.md)
- [zap 文档](https://pkg.go.dev/go.uber.org/zap)
- [lumberjack 文档](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2)
