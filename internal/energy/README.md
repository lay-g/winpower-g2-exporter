# Energy Module

电能计算模块（Energy）是 WinPower G2 Prometheus Exporter 的核心组件，负责从设备指标数据中计算和累计电能消耗。

## 概述

该模块采用**极简单线程架构**，通过全局锁确保所有计算操作的串行执行，彻底避免并发问题：

- **全局串行**：所有设备的电能计算都串行执行，完全避免数据竞争
- **极简结构**：只有单一服务类，消除复杂的队列、路由器、池管理
- **直接计算**：没有任务缓冲和队列调度，直接执行计算逻辑
- **最小依赖**：只依赖storage模块进行数据持久化

## 特性

- ✅ 精确的电能累计计算（Wh = W × 时间间隔）
- ✅ 支持正负功率和零功率处理
- ✅ 完全依赖storage模块进行数据持久化
- ✅ 提供统计信息用于监控和调试
- ✅ 线程安全的并发访问
- ✅ 高测试覆盖率（96.2%）

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    Energy Module                             │
│                  (极简单线程架构)                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              EnergyService                          │   │
│  │                                                     │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │Global Mutex │  │ Calculator  │  │   Storage   │  │   │
│  │  │             │  │             │  │  Interface  │  │   │
│  │  │ - Serial    │  │ - Energy    │  │ - Read      │  │   │
│  │  │   Execution │  │   Accumulate│  │ - Write     │  │   │
│  │  │ - Data      │  │ - Time      │  │ - Exists    │  │   │
│  │  │   Consistency│ │   Calculate │  │             │  │   │
│  │  │ - No        │  │ - Data      │  │             │  │   │
│  │  │   Race      │  │   Update    │  │             │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 使用示例

### 基本使用

```go
package main

import (
    "github.com/lay-g/winpower-g2-exporter/internal/energy"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
    "github.com/lay-g/winpower-g2-exporter/internal/storage"
)

func main() {
    // 创建日志器
    logger := log.NewTestLogger()

    // 创建存储管理器
    storageConfig := &storage.Config{
        DataDir:         "./data",
        FilePermissions: 0644,
    }
    storageManager, err := storage.NewFileStorageManager(storageConfig, logger)
    if err != nil {
        panic(err)
    }

    // 创建电能服务（极简架构，无需额外启动）
    energyService := energy.NewEnergyService(storageManager, logger)

    // 模拟collector调用
    deviceID := "ups-001"
    power := 500.0

    // 计算电能
    totalEnergy, err := energyService.Calculate(deviceID, power)
    if err != nil {
        logger.Error("Failed to calculate energy", log.Err(err))
        return
    }

    logger.Info("Energy calculation completed",
        log.String("device", deviceID),
        log.Float64("total_energy", totalEnergy),
        log.Float64("current_power", power))

    // 获取最新电能数据
    currentEnergy, err := energyService.Get(deviceID)
    if err != nil {
        logger.Error("Failed to get energy data", log.Err(err))
        return
    }

    logger.Info("Current energy data",
        log.String("device", deviceID),
        log.Float64("total_energy", currentEnergy))

    // 获取统计信息
    stats := energyService.GetStats()
    logger.Info("Energy service statistics",
        log.Int64("total_calculations", stats.GetTotalCalculations()),
        log.Int64("total_errors", stats.GetTotalErrors()))
}
```

### 功率来源约定

统一约定：能耗累计的输入功率为设备实时总负载有功功率 `loadTotalWatt`（单位 `W`）。

- **Collector模块**负责从WinPower协议响应中提取该字段
- Collector调用 `energy.Calculate(deviceID, power)` 时传入功率值
- **WinPower模块**仅提供原始数据，不直接调用energy模块

说明：当设备仅暴露分相功率时，Collector应先汇总为总负载有功功率后再参与累计；不使用视在功率 `loadTotalVa` 或单相 `loadWatt1` 直接参与能耗计算。

## 接口定义

### EnergyInterface

```go
type EnergyInterface interface {
    // Calculate 计算电能
    Calculate(deviceID string, power float64) (float64, error)

    // Get 获取最新电能数据
    Get(deviceID string) (float64, error)

    // GetStats 获取统计信息
    GetStats() *Stats
}
```

### Stats

```go
type Stats struct {
    TotalCalculations  int64         `json:"total_calculations"`
    TotalErrors        int64         `json:"total_errors"`
    LastUpdateTime     time.Time     `json:"last_update_time"`
    AvgCalculationTime time.Duration `json:"avg_calculation_time"`
}
```

## 电能计算原理

### 计算公式

```
时间间隔 (h) = (当前时间 - 上次更新时间) / 3600
间隔电能 (Wh) = 输入功率 (W) × 时间间隔 (h)
累计电能 (Wh) = 历史累计电能 + 间隔电能
```

### 数据精度

- 时间精度：毫秒级
- 电能精度：0.01Wh
- 数据类型：float64

### 功率处理

- **正功率**：累计电能单调递增
- **负功率**：累计电能递减（表示净能量减少）
- **零功率**：累计电能保持不变，时间线正常推进

## 错误处理

模块定义了以下错误类型：

- `ErrInvalidDeviceID`: 设备ID无效
- `ErrInvalidPower`: 功率值无效
- `ErrStorageRead`: 存储读取失败
- `ErrStorageWrite`: 存储写入失败
- `ErrCalculation`: 电能计算失败

## 测试

### 运行测试

```bash
# 运行单元测试
make test

# 运行集成测试
go test -v ./internal/energy/... -count=1

# 运行测试覆盖率
go test -coverprofile=coverage.txt -covermode=atomic ./internal/energy/...
```

### 测试覆盖率

当前测试覆盖率：**96.2%**

### 测试场景

#### 单元测试
- 基础功能测试（Calculate、Get、GetStats）
- 连续计算测试（电能累计正确性）
- 并发安全测试（多goroutine并发访问）
- 边界条件测试（首次访问、负功率、零功率等）
- 错误处理测试（存储错误、参数错误等）

#### 集成测试
- 端到端集成测试（完整的电能计算流程）
- 数据一致性测试（多次计算的数据一致性）
- 文件存储集成测试
- 服务重启后数据恢复

## 性能特点

- **低延迟**：对于少量设备（<20个），串行执行延迟可忽略
- **低开销**：没有队列管理和任务调度的额外开销
- **预测性强**：执行时间完全可预测，没有并发抖动
- **内存友好**：极低的内存占用

## 适用场景

- WinPower G2系统的中小规模UPS设备电能监控
- 对可靠性要求高于性能要求的场景
- 开发和维护资源有限的项目
- 需要简单易懂代码的团队

## 依赖

- `internal/storage`: 数据持久化
- `internal/pkgs/log`: 日志记录

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](../../LICENSE) 文件。
