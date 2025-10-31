# Scheduler Module Design Document

## 概述

本文档详细描述了Scheduler模块的实现设计，该模块基于简化设计原则，负责定时触发数据采集功能。

## 架构决策

### 设计原则

1. **简化优先**: 避免复杂的任务队列、优先级管理等功能
2. **职责单一**: 仅负责定时触发，不处理数据采集逻辑
3. **可靠性优先**: 错误不应影响后续周期的执行
4. **测试友好**: 接口设计便于单元测试和集成测试

### 关键权衡

- **简单性 vs 灵活性**: 选择固定间隔触发，不支持动态调整间隔
- **性能 vs 可靠性**: 选择单线程顺序执行，避免并发复杂性
- **功能范围**: 明确排除任务管理、依赖关系等复杂功能

## 模块架构

### 组件关系图

```
┌─────────────────────────────────────────────────┐
│                 Scheduler Module                │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────┐│
│  │           DefaultScheduler                  ││
│  │                                             ││
│  │  • config: *Config                          ││
│  │  • collector: CollectorInterface            ││
│  │  • logger: logger.Logger                    ││
│  │  • ticker: *time.Ticker                     ││
│  │  • ctx, cancel: context                     ││
│  │  • running: bool (with mutex)               ││
│  │  • wg: sync.WaitGroup                       ││
│  │                                             ││
│  │  Start() ──────────────────────────────────┐│
│  │  Stop()  ──────────────────────────────────┐│
│  └─────────────────────┬───────────────────────┘│
│                        │                       │
│         every 5 seconds │                       │
│                        ▼                       │
│  ┌─────────────────────────────────────────────┐│
│  │        Collector.CollectDeviceData()        ││
│  └─────────────────────────────────────────────┘│
└─────────────────────────────────────────────────┘
```

### 接口设计

#### Scheduler Interface

```go
type Scheduler interface {
    // Start 启动调度器，开始定时触发数据采集
    Start(ctx context.Context) error

    // Stop 停止调度器，优雅关闭所有goroutine
    Stop(ctx context.Context) error
}
```

#### 依赖接口

```go
// CollectorInterface 来自collector模块
type CollectorInterface interface {
    CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}

// Logger 来自项目的统一日志接口
type Logger interface {
    Info(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
}
```

## 实现细节

### 配置管理

配置结构基于设计文档，实现验证逻辑：

```go
type Config struct {
    CollectionInterval      time.Duration `yaml:"collection_interval"`
    GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
}

func DefaultConfig() *Config {
    return &Config{
        CollectionInterval:      5 * time.Second,
        GracefulShutdownTimeout: 5 * time.Second,
    }
}
```

### 核心实现

#### DefaultScheduler结构

```go
type DefaultScheduler struct {
    config    *Config
    collector CollectorInterface
    logger    Logger

    // 运行时状态
    ticker    *time.Ticker
    ctx       context.Context
    cancel    context.CancelFunc
    wg        sync.WaitGroup
    running   bool
    mu        sync.RWMutex
}
```

#### 启动流程

1. 检查当前状态，防止重复启动
2. 创建带取消的context
3. 创建5秒间隔的Ticker
4. 启动goroutine监听ticker事件
5. 在每个tick时调用collector.CollectDeviceData
6. 记录执行结果或错误
7. 更新运行状态

#### 停止流程

1. 检查当前状态，防止重复停止
2. 取消context通知goroutine退出
3. 停止ticker
4. 等待goroutine完成（带超时）
5. 清理资源
6. 更新运行状态

### 错误处理策略

1. **采集失败**: 记录错误日志，继续下一周期
2. **超时错误**: 在停止流程中处理，返回超时错误
3. **并发错误**: 使用mutex保护状态变更
4. **依赖错误**: 通过构造函数依赖注入，启动时验证

### 测试策略

#### 单元测试覆盖

1. **配置验证测试**: 边界值、错误输入
2. **接口测试**: Start/Stop方法的正常和异常流程
3. **并发测试**: 多线程访问的安全性
4. **Mock测试**: 使用mock collector和logger

#### 集成测试场景

1. **完整生命周期**: 启动→运行多个周期→停止
2. **错误恢复**: Collector错误后的继续运行
3. **快速启停**: 快速启动和停止的边界情况
4. **资源清理**: 验证goroutine和资源正确释放

## 与其他模块的集成

### Collector模块集成

- 依赖注入CollectorInterface
- 调用CollectDeviceData方法
- 处理CollectionResult（主要是记录日志）

### Config模块集成

- 配置项注册到主配置结构
- 支持YAML配置文件
- 环境变量覆盖支持

### 主程序集成

```go
// 在main.go中的集成示例
func main() {
    // ... 其他初始化 ...

    schedulerConfig := config.Scheduler
    scheduler := scheduler.NewDefaultScheduler(
        schedulerConfig,
        collector,
        logger,
    )

    // 启动scheduler
    go func() {
        if err := scheduler.Start(context.Background()); err != nil {
            logger.Error("scheduler start failed", "error", err)
        }
    }()

    // 应用关闭时停止scheduler
    defer func() {
        if err := scheduler.Stop(context.Background()); err != nil {
            logger.Error("scheduler stop failed", "error", err)
        }
    }()

    // ... 启动HTTP服务器等 ...
}
```

## 性能考虑

1. **内存使用**: 最小化状态存储，主要是配置和运行时标志
2. **CPU使用**: 大部分时间处于睡眠状态，仅在tick时活动
3. **goroutine管理**: 仅使用一个goroutine进行定时触发
4. **锁竞争**: 读写锁优化状态查询操作

## 监控和可观测性

1. **日志记录**: 启动、停止、每次采集尝试的结果
2. **状态暴露**: 可选择性暴露运行状态（通过HTTP接口）
3. **错误统计**: 记录连续失败次数（可选功能）

## 未来扩展可能性

虽然当前设计保持简化，但为未来扩展预留了空间：

1. **动态间隔调整**: 可通过配置热更新支持
2. **健康检查接口**: 可添加HTTP接口检查状态
3. **指标集成**: 可添加Prometheus指标监控调度器状态
4. **多任务支持**: 架构上可扩展支持多个定时任务

这些扩展都需要在保持简化核心的前提下谨慎添加。