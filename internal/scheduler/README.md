# Scheduler Module

调度器模块负责按固定间隔（默认5秒）定时触发数据采集操作。该模块遵循简化设计原则，专注于可靠性和易测试性。

## 功能特性

- **固定间隔触发**：每5秒触发一次数据采集（可配置）
- **优雅启停**：支持优雅启动和关闭操作
- **错误恢复**：单次采集错误不影响后续周期
- **结构化日志**：集成项目统一的日志系统
- **线程安全**：使用互斥锁保护状态管理
- **测试友好**：基于接口设计，便于模拟测试

## 设计原则

- **简单性优于灵活性**：固定间隔触发，不支持动态调整
- **可靠性优于性能**：单线程顺序执行，避免并发复杂性
- **职责清晰**：仅负责定时触发，不处理采集逻辑
- **接口驱动**：便于单元测试和集成测试

## 使用示例

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/lay-g/winpower-g2-exporter/internal/scheduler"
)

func main() {
    // 创建配置
    config := scheduler.DefaultConfig()
    // 或自定义配置
    config = &scheduler.Config{
        CollectionInterval:      5 * time.Second,
        GracefulShutdownTimeout: 5 * time.Second,
    }

    // 创建调度器（需要collector和logger的实现）
    sched, err := scheduler.NewDefaultScheduler(config, collector, logger)
    if err != nil {
        log.Fatal("failed to create scheduler:", err)
    }

    // 启动调度器
    ctx := context.Background()
    if err := sched.Start(ctx); err != nil {
        log.Fatal("failed to start scheduler:", err)
    }

    // 运行一段时间后优雅停止
    time.Sleep(30 * time.Second)
    if err := sched.Stop(ctx); err != nil {
        log.Error("failed to stop scheduler:", err)
    }
}
```

## 接口定义

### Scheduler Interface

```go
type Scheduler interface {
    // Start 启动调度器，开始定时触发数据采集
    Start(ctx context.Context) error

    // Stop 停止调度器，优雅关闭所有goroutine
    Stop(ctx context.Context) error
}
```

### 依赖接口

```go
// CollectorInterface 数据采集器接口
type CollectorInterface interface {
    CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}

// Logger 日志接口
type Logger interface {
    Info(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Debug(msg string, fields ...interface{})
}
```

## 配置说明

### Config 结构

```go
type Config struct {
    // CollectionInterval 数据采集间隔
    // 默认值：5秒
    // 最小值：1秒
    // 最大值：1小时
    CollectionInterval time.Duration

    // GracefulShutdownTimeout 优雅关闭超时时间
    // 默认值：5秒
    // 必须为正值
    GracefulShutdownTimeout time.Duration
}
```

### 默认配置

```go
config := scheduler.DefaultConfig()
// config.CollectionInterval = 5 * time.Second
// config.GracefulShutdownTimeout = 5 * time.Second
```

### 配置验证

配置会自动验证以下约束：
- `CollectionInterval` 必须在 1秒 到 1小时 之间
- `GracefulShutdownTimeout` 必须为正值

## 错误处理

模块定义了以下错误常量：

- `ErrAlreadyRunning`：调度器已经在运行
- `ErrNotRunning`：调度器未运行
- `ErrShutdownTimeout`：优雅关闭超时
- `ErrNilCollector`：collector 为 nil
- `ErrNilLogger`：logger 为 nil
- `ErrNilConfig`：config 为 nil

## 测试

模块提供完整的单元测试覆盖（97.4%）：

```bash
# 运行测试
make test

# 查看覆盖率
go test ./internal/scheduler/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 测试场景

- 配置验证测试
- 启动/停止流程测试
- 数据采集触发测试
- 错误恢复测试
- 并发安全测试
- 日志记录测试

## 架构关系

```
┌─────────────────────────────────────────────────┐
│                 Scheduler Module                │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────┐│
│  │           DefaultScheduler                  ││
│  │                                             ││
│  │  • config: *Config                          ││
│  │  • collector: CollectorInterface            ││
│  │  • logger: Logger                           ││
│  │  • ticker: *time.Ticker                     ││
│  │  • ctx, cancel: context                     ││
│  │  • running: bool (with mutex)               ││
│  │                                             ││
│  │  Start() ──────────────────────────────────┐││
│  │  Stop()  ──────────────────────────────────┐││
│  └─────────────────────┬───────────────────────┘│
│                        │                       │
│         every 5 seconds│                       │
│                        ▼                       │
│  ┌─────────────────────────────────────────────┐│
│  │        Collector.CollectDeviceData()        ││
│  └─────────────────────────────────────────────┘│
└─────────────────────────────────────────────────┘
```

## 性能考虑

- **内存使用**：最小化状态存储，主要是配置和运行时标志
- **CPU使用**：大部分时间处于睡眠状态，仅在tick时活动
- **Goroutine管理**：仅使用一个goroutine进行定时触发
- **锁竞争**：读写锁优化状态查询操作

## 未来扩展

虽然当前设计保持简化，但为未来扩展预留了空间：

1. **动态间隔调整**：可通过配置热更新支持
2. **健康检查接口**：可添加HTTP接口检查状态
3. **指标集成**：可添加Prometheus指标监控调度器状态
4. **多任务支持**：架构上可扩展支持多个定时任务

## 参考文档

- [设计文档](../../docs/design/scheduler.md)
- [架构设计](../../docs/design/architecture.md)
- [Collector模块](../collector/README.md)
