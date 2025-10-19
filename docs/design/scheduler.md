# 定时任务模块设计（简化版）

## 1. 概述

定时任务模块（Scheduler）按照简化设计仅负责一项固定任务：
- 每 5 秒触发 WinPower 模块的统一采集方法（CollectDeviceData），并间接完成电能计算

该模块不再维护任务队列、并发池、优先队列或复杂生命周期，仅提供启动与停止的基本控制，并在固定时间间隔驱动数据采集任务。Token刷新已集成到WinPower模块内部，由该模块自动管理。

### 1.1 设计目标

- 精确性：基于 `time.Ticker` 提供稳定的定时触发
- 可靠性：错误就地记录，不影响后续周期触发
- 简洁性：不使用队列、worker、依赖管理或分布式调度
- 可监控：记录基础执行日志（开始、结束、耗时、错误）

### 1.2 功能范围

- ✅ 固定定时触发：数据采集（每 5 秒）
- ✅ 启停控制：模块启动、优雅停止
- ✅ 基础日志：任务触发与结果状态
- ❌ 任务队列与并发池：不支持
- ❌ 自动重试、任务依赖、分布式调度：不支持
- ❌ Token刷新：已集成到WinPower模块内部

## 2. 架构设计

### 2.1 模块架构图

```
┌─────────────────────────────────────────────┐
│                 Scheduler                   │
├─────────────────────────────────────────────┤
│  ┌────────────────────────────────────────┐│
│  │        Data Collect Ticker (5s)        ││
│  └────────────────────┬───────────────────┘│
│                       │                   │
│               CollectDeviceData()         │
│                       ▼                   │
│  ┌────────────────────────────────────────┐│
│  │           WinPower Module              ││
│  │        (认证 + 数据采集)               ││
│  └────────────────────────────────────────┘│
└─────────────────────────────────────────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │  Energy Module   │  ← 由 WinPower 触发电能累计
                 └──────────────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │   Storage Layer  │
                 └──────────────────┘
```

### 2.2 调度流程

数据采集与电能计算（每 5 秒）

```
Tick(5s) → WinPower.CollectDeviceData() → Energy 累计 → 存储 → Log(result)
```

说明：WinPower模块内部自动管理Token刷新，Scheduler只需要定时触发数据采集即可。

## 3. 接口设计

### 3.1 核心接口

```go
// Scheduler 简化接口
type Scheduler interface {
    // Start 启动调度器（创建两个 Ticker 并监听事件）
    Start(ctx context.Context) error

    // Stop 停止调度器（优雅关闭、释放资源）
    Stop(ctx context.Context) error
}
```

### 3.2 配置结构

```go
// SchedulerConfig 调度器配置（内部配置）
type SchedulerConfig struct {
    // 数据采集间隔（默认 5s）
    CollectionInterval     time.Duration `yaml:"collection_interval" json:"collection_interval"`

    // 优雅关闭超时（默认 30s）
    GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout" json:"graceful_shutdown_timeout"`
}

func DefaultSchedulerConfig() *SchedulerConfig {
    return &SchedulerConfig{
        CollectionInterval:      5 * time.Second,
        GracefulShutdownTimeout: 30 * time.Second,
    }
}
```

## 4. 详细实现（示意）

### 4.1 调度器实现

```go
// DefaultScheduler 简化实现
type DefaultScheduler struct {
    config        *SchedulerConfig
    winpowerClient WinPowerClient // 依赖：WinPower模块
    logger        logger.Logger

    collectTicker *time.Ticker

    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
    running       bool
    mu            sync.RWMutex
}

func NewDefaultScheduler(cfg *SchedulerConfig, wc WinPowerClient, l logger.Logger) *DefaultScheduler {
    if cfg == nil { cfg = DefaultSchedulerConfig() }
    return &DefaultScheduler{ config: cfg, winpowerClient: wc, logger: l }
}

// Start 启动数据采集 Ticker 并监听触发
func (ds *DefaultScheduler) Start(ctx context.Context) error {
    ds.mu.Lock()
    defer ds.mu.Unlock()
    if ds.running { return nil }

    ds.ctx, ds.cancel = context.WithCancel(ctx)
    ds.collectTicker = time.NewTicker(ds.config.CollectionInterval)

    // 数据采集与电能计算循环
    ds.wg.Add(1)
    go func() {
        defer ds.wg.Done()
        for {
            select {
            case <-ds.ctx.Done():
                return
            case <-ds.collectTicker.C:
                if err := ds.winpowerClient.CollectDeviceData(ds.ctx); err != nil {
                    ds.logger.Error("collect+energy failed", "error", err)
                } else {
                    ds.logger.Info("collect+energy ok")
                }
            }
        }
    }()

    ds.running = true
    ds.logger.Info("scheduler started")
    return nil
}

// Stop 优雅停止
func (ds *DefaultScheduler) Stop(ctx context.Context) error {
    ds.mu.Lock()
    if !ds.running {
        ds.mu.Unlock()
        return nil
    }
    ds.running = false
    ds.mu.Unlock()

    ds.logger.Info("scheduler stopping")
    if ds.cancel != nil { ds.cancel() }
    if ds.collectTicker != nil { ds.collectTicker.Stop() }

    done := make(chan struct{})
    go func(){ ds.wg.Wait(); close(done) }()

    select {
    case <-done:
        ds.logger.Info("scheduler stopped")
        return nil
    case <-time.After(ds.config.GracefulShutdownTimeout):
        ds.logger.Warn("scheduler stop timeout")
        return context.DeadlineExceeded
    }
}
```

### 4.2 错误处理策略

- 采集或计算失败：记录错误并等待下一周期，不做额外重试
- Token刷新失败：WinPower模块内部处理，Scheduler不关心
- 停止流程：强制停止 Ticker，等待 goroutine 结束，超过超时则返回错误

## 5. 政策与限制

- 不支持任务注册/注销，任务列表固定为一项
- 不支持队列、并发执行与优先级管理
- 不提供自动重试或任务依赖控制
- Token刷新由WinPower模块内部管理，Scheduler不直接处理
- 通过日志实现轻量监控；如需指标可在后续版本单独引入