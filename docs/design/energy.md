# 电能计算模块设计文档（极简方案）

## 概述

电能计算模块（Energy）是 WinPower G2 Prometheus Exporter 的核心组件，负责从设备指标数据中计算和累计电能消耗。该模块采用极简的设计理念，使用单一全局锁确保数据一致性，专注于为UPS设备提供精确的电能累计计算功能。

**重要说明**:
- 本模块仅由collector模块触发计算，不提供定时触发机制
- 只保留累计电能数据，间隔电能计算交给Prometheus处理
- 专注支持UPS设备类型，简化设备兼容性处理
- 通过全局锁机制确保计算的串行执行和数据一致性

### 设计目标

- **极简性**: 最简化的代码结构，单一职责，易于理解和维护
- **数据一致性**: 通过全局锁确保所有计算操作串行执行，完全避免数据竞争
- **可靠性**: 完全依赖storage模块进行数据持久化，确保数据不丢失
- **专注性**: 专门针对UPS设备优化，提供精确的电能计算

## 架构设计

### 极简设计概述

电能计算模块采用**极简单线程架构**，通过全局锁确保所有计算操作的串行执行，彻底避免并发问题：

- **全局串行**：所有设备的电能计算都串行执行，完全避免数据竞争
- **极简结构**：只有单一服务类，消除复杂的队列、路由器、池管理
- **直接计算**：没有任务缓冲和队列调度，直接执行计算逻辑
- **最小依赖**：只依赖storage模块进行数据持久化

### 模块结构图

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
├─────────────────────────────────────────────────────────────┤
│                        Public APIs                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Calculate(deviceID, power) -> float64             │   │
│  │  Get(deviceID) -> float64                          │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Storage Module                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         StorageManager Interface                    │   │
│  │  - Write(deviceID, *PowerData)                     │   │
│  │  - Read(deviceID) -> *PowerData                    │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 数据流程图

```
Collector模块调用
    │
    ▼
┌─────────────────┐
│ energy.Calculate│
│ (deviceID,      │
│  power)         │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 获取全局锁       │
│ (确保串行执行)   │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 加载历史数据     │
│ (从storage)     │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 计算累计电能     │
│ (功率×时间间隔)  │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 保存数据到storage│
│ (原子写入)       │
└─────────────────┘
    │
    ▼
┌─────────────────┐
│ 释放锁并返回     │
│ 电能值          │
└─────────────────┘
```

### 极简架构优势

#### 可靠性保证

1. **绝对数据一致性**：所有操作串行执行，彻底消除并发问题
2. **无锁竞争**：单一全局锁，不存在死锁和锁竞争问题
3. **简化错误处理**：线性执行路径，错误处理更简单直观

#### 维护性提升

1. **代码简洁**：核心逻辑只需几百行代码
2. **易于理解**：没有复杂的并发控制，逻辑清晰
3. **调试简单**：串行执行，调试和问题定位更容易

#### 性能特点

1. **低延迟**：对于少量设备（<20个），串行执行延迟可忽略
2. **低开销**：没有队列管理和任务调度的额外开销
3. **预测性强**：执行时间完全可预测，没有并发抖动

## 电能计算原理

### 计算公式

#### 1. 电能累计计算

```
时间间隔 (h) = (当前时间 - 上次更新时间) / 3600
间隔电能 (Wh) = 输入功率 (W) × 时间间隔 (h)
累计电能 (Wh) = 历史累计电能 + 间隔电能
```

### 数据精度控制

#### 1. 时间精度
- 时间戳精度：毫秒级
- 累计计算：无时间间隔限制，不管中间间隔多久都需要计算累计

#### 2. 数值精度
- 功率精度：使用输入的原始精度
- 电能精度：0.01Wh
- 数据类型：float64

## 核心组件设计

### 1. EnergyService（单一服务类）

#### 职责
- 提供模块的统一对外接口
- 通过全局锁确保所有计算的串行执行
- 协调与storage模块的交互
- 维护简单直观的执行状态

#### 数据结构

```go
// EnergyService 电能服务（极简架构）
type EnergyService struct {
    storage storage.StorageManager    // 存储接口
    logger  *zap.Logger              // 日志器
    mutex   sync.RWMutex             // 全局读写锁，确保串行执行

    // 统计信息（可选）
    stats   *SimpleStats             // 简单统计信息
}

// SimpleStats 简单统计信息
type SimpleStats struct {
    TotalCalculations int64           `json:"total_calculations"`
    TotalErrors       int64           `json:"total_errors"`
    LastUpdateTime    time.Time       `json:"last_update_time"`
    AvgCalculationTime time.Duration  `json:"avg_calculation_time"`
    mutex             sync.RWMutex    // 保护统计信息
}
```

#### 核心方法

```go
// NewEnergyService 创建电能服务
func NewEnergyService(
    storage storage.StorageManager,
    logger *zap.Logger,
) *EnergyService

// Calculate 计算电能（对外接口，串行执行）
func (es *EnergyService) Calculate(deviceID string, power float64) (float64, error)

// Get 获取最新电能数据（对外接口）
func (es *EnergyService) Get(deviceID string) (float64, error)

// GetStats 获取简单统计信息
func (es *EnergyService) GetStats() *SimpleStats
```

#### 实现要点

```go
func NewEnergyService(storage storage.StorageManager, logger *zap.Logger) *EnergyService
// 实现逻辑：
// 1. 保存存储接口引用
// 2. 保存日志器引用
// 3. 初始化全局读写锁
// 4. 初始化简单统计信息
// 5. 返回服务实例

func (es *EnergyService) Calculate(deviceID string, power float64) (float64, error)
// 实现逻辑：
// 1. 获取全局写锁（确保串行执行）
// 2. 记录开始时间和统计信息
// 3. 加载历史数据
// 4. 计算时间间隔和间隔电能
// 5. 计算新的累计电能
// 6. 保存数据到storage
// 7. 更新统计信息
// 8. 释放锁并返回结果

func (es *EnergyService) Get(deviceID string) (float64, error)
// 实现逻辑：
// 1. 获取读锁（允许并发读取）
// 2. 从storage读取设备数据
// 3. 释放锁并返回结果

func (es *EnergyService) GetStats() *SimpleStats
// 实现逻辑：
// 1. 获取读锁保护统计信息
// 2. 返回统计信息副本

// calculateTotalEnergy 计算累计电能（内部方法）
func (es *EnergyService) calculateTotalEnergy(historyData *storage.PowerData, currentPower float64, currentTime time.Time) (float64, error)
// 实现逻辑：
// 1. 计算时间间隔（小时）
// 2. 计算间隔电能 = 功率 × 时间间隔
// 3. 计算新的累计电能 = 历史电能 + 间隔电能
// 4. 返回累计电能值

// loadHistoryData 加载历史数据（内部方法）
func (es *EnergyService) loadHistoryData(deviceID string) (*storage.PowerData, error)
// 实现逻辑：
// 1. 调用storage.Read读取历史数据
// 2. 处理文件不存在等错误情况
// 3. 返回历史数据或nil

// saveData 保存数据（内部方法）
func (es *EnergyService) saveData(deviceID string, energy float64) error
// 实现逻辑：
// 1. 创建新的PowerData结构
// 2. 调用storage.Write保存数据
// 3. 返回保存结果

// updateStats 更新统计信息（内部方法）
func (es *EnergyService) updateStats(success bool, duration time.Duration)
// 实现逻辑：
// 1. 更新总计算次数和错误次数
// 2. 计算平均执行时间
// 3. 更新最后更新时间
```

## 接口设计

### Energy Interface

```go
// EnergyInterface 电能模块接口
type EnergyInterface interface {
    // Calculate 计算电能
    Calculate(deviceID string, power float64) (float64, error)

    // Get 获取最新电能数据
    Get(deviceID string) (float64, error)
}

// EnergyService 电能服务（极简实现）
type EnergyService struct {
    storage storage.StorageManager   // 存储接口
    logger  *zap.Logger             // 日志器
    mutex   sync.RWMutex            // 全局读写锁，确保串行执行
    stats   *SimpleStats            // 简单统计信息
}
```

### Storage Interface 依赖

```go
// 电能模块依赖的存储接口（由storage模块提供）
type StorageManager interface {
    // Write 写入设备电能数据
    Write(deviceID string, data *PowerData) error

    // Read 读取设备电能数据
    Read(deviceID string) (*PowerData, error)
}

// PowerData 电能数据结构（storage模块定义）
type PowerData struct {
    Timestamp int64   `json:"timestamp"` // 毫秒时间戳
    EnergyWH  float64 `json:"energy_wh"` // 累计电能(Wh)
}
```

## 使用示例

### 功率来源约定

统一约定：能耗累计的输入功率为设备实时总负载有功功率 `loadTotalWatt`（单位 `W`）。

- **Collector模块**负责从WinPower协议响应中提取该字段
- Collector调用 `energy.Calculate(deviceID, power)` 时传入功率值
- **WinPower模块**仅提供原始数据，不直接调用energy模块

说明：当设备仅暴露分相功率时，Collector应先汇总为总负载有功功率后再参与累计；不使用视在功率 `loadTotalVa` 或单相 `loadWatt1` 直接参与能耗计算。

### 基本使用

```go
package main

import (
    "go.uber.org/zap"
    "your-project/internal/energy"
    "your-project/internal/storage"
)

func main() {
    // 创建日志器
    logger := zap.NewProduction()
    defer logger.Sync()

    // 创建存储管理器
    storageConfig := &storage.StorageConfig{
        DataDir:         "./data",
        SyncWrite:       true,
        FilePermissions: 0644,
    }
    storageManager := storage.NewFileStorageManager(storageConfig, logger)

    // 创建电能服务（极简架构，无需额外启动）
    energyService := energy.NewEnergyService(storageManager, logger)

    // 模拟collector调用
    deviceID := "ups-001"
    power := 500.0

    // 计算电能
    totalEnergy, err := energyService.Calculate(deviceID, power)
    if err != nil {
        logger.Error("Failed to calculate energy",
            zap.String("device", deviceID),
            zap.Error(err))
        return
    }

    logger.Info("Energy calculation completed",
        zap.String("device", deviceID),
        zap.Float64("total_energy", totalEnergy),
        zap.Float64("current_power", power))

    // 获取最新电能数据
    currentEnergy, err := energyService.Get(deviceID)
    if err != nil {
        logger.Error("Failed to get energy data", zap.Error(err))
        return
    }

    logger.Info("Current energy data",
        zap.String("device", deviceID),
        zap.Float64("total_energy", currentEnergy))

    // 获取统计信息
    stats := energyService.GetStats()
    logger.Info("Energy service statistics",
        zap.Int64("total_calculations", stats.TotalCalculations),
        zap.Int64("total_errors", stats.TotalErrors),
        zap.Duration("avg_calculation_time", stats.AvgCalculationTime))
}
```

### 在Collector模块中的集成

```go
// 在 collector 模块中使用 energy 模块（极简架构）
func (c *Collector) processDeviceData(device *ParsedDeviceData) {
    // 获取功率值（由collector模块负责功率提取）
    power := device.PowerInfo.LoadTotalWatt // 来自协议字段 loadTotalWatt

    // 调用energy模块计算电能
    totalEnergy, err := c.energyService.Calculate(device.DeviceID, power)
    if err != nil {
        c.logger.Error("Failed to calculate energy",
            zap.String("device", device.DeviceID),
            zap.Error(err))
        return
    }

    // 更新设备数据中的电能信息
    device.EnergyInfo.TotalEnergy = totalEnergy
    device.EnergyInfo.CurrentPower = power
    // LastUpdate 由 energy 模块内部维护并写入存储，collector 不覆盖

    c.logger.Debug("Energy calculation completed",
        zap.String("device", device.DeviceID),
        zap.Float64("total_energy", totalEnergy),
        zap.Float64("current_power", power))
}
```

## 监控和日志

### 简化监控指标

```go
// 能量模块监控指标（极简版）
var (
    // 计算总数
    energyCalculationsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "winpower_energy_calculations_total",
            Help: "Total number of energy calculations",
        },
        []string{"device_id", "result"},
    )

    // 计算持续时间
    energyCalculationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "winpower_energy_calculation_duration_seconds",
            Help: "Duration of energy calculations",
        },
        []string{"device_id"},
    )

    // 总电能值
    energyTotalWh = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "winpower_energy_total_wh",
            Help: "Total energy consumption in Wh",
        },
        []string{"device_id"},
    )

    // 计算错误率
    energyCalculationErrors = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "winpower_energy_calculation_error_rate",
            Help: "Energy calculation error rate (0-1)",
        },
    )
)
```

### 结构化日志记录

```go
// 能量计算日志记录示例
func (es *EnergyService) Calculate(deviceID string, power float64) (float64, error) {
    es.mutex.Lock()
    defer es.mutex.Unlock()

    start := time.Now()
    logger := es.logger.With(
        zap.String("device_id", deviceID),
        zap.Float64("power", power),
    )

    logger.Debug("Starting energy calculation")

    // ... 计算逻辑 ...

    duration := time.Since(start)

    if err != nil {
        logger.Error("Energy calculation failed",
            zap.Error(err),
            zap.Duration("duration", duration))
        return 0, err
    }

    logger.Info("Energy calculation completed",
        zap.Float64("total_energy", totalEnergy),
        zap.Duration("duration", duration))

    return totalEnergy, nil
}
```

## 测试设计

### 单元测试场景

#### 基础功能测试
- **TestEnergyService_Calculate**: 验证首次电能计算从0开始
- **TestEnergyService_Get**: 验证获取设备电能数据功能
- **TestEnergyService_SequentialCalculations**: 验证连续计算时电能累计正确性

#### 并发安全测试
- **TestEnergyService_ConcurrentAccess**: 验证多goroutine并发访问的安全性
  - 测试场景：10个goroutine，每个执行20次计算
  - 验证点：无数据竞争、统计信息正确、计算结果一致

#### Mock测试配置
- 使用MockStorage模拟存储接口行为
- 测试首次访问（文件不存在）和后续访问（历史数据存在）
- 验证storage接口调用次数和参数正确性

### 集成测试场景

#### 端到端测试
- **TestEnergyService_Integration**: 验证完整的电能计算流程
  - 测试场景：使用真实文件存储，创建临时目录
  - 验证点：数据持久化、文件创建、服务重启后数据恢复
  - 测试步骤：两次计算间隔100ms，验证电能累计增长

#### 数据一致性测试
- **TestEnergyService_DataConsistency**: 验证多次计算的数据一致性
  - 测试场景：不同功率和延迟的组合计算
  - 验证点：电能单调递增（正功率情况下）、最终数据一致性
  - 测试数据：4组不同功率值（1000W、800W、1200W、900W）和延迟

#### 集成测试配置
- 使用真实文件存储（非Mock）
- 临时目录管理（测试后自动清理）
- 支持short模式跳过（testing.Short()）
- 验证文件系统操作和数据恢复

## 最佳实践

### 使用建议

1. **适用场景**：
   - 设备数量较少（<20个）的场景
   - 对可靠性要求高于性能要求的场景
   - 开发和维护资源有限的项目
   - 需要简单易懂代码的团队

2. **性能特点**：
   - 对于WinPower G2系统（5秒采集一次），串行执行延迟可忽略
   - 内存占用极低，无额外队列开销
   - 执行时间完全可预测，无并发抖动

3. **监控建议**：
   - 重点关注计算成功率
   - 监控平均计算时间
   - 设置电能数据未更新告警

### 监控告警

```yaml
# Prometheus告警规则示例
groups:
- name: winpower_energy
  rules:
  # 计算失败告警
  - alert: WinPowerEnergyCalculationFailed
    expr: increase(winpower_energy_calculations_total{result="error"}[5m]) > 5
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "WinPower energy calculation failures detected"
      description: "More than 5 energy calculation failures in the last 5 minutes for device {{ $labels.device_id }}."

  # 电能数据未更新告警
  - alert: WinPowerEnergyNotUpdating
    expr: abs(delta(winpower_energy_total_wh[10m])) < 0.001
    for: 15m
    labels:
      severity: warning
    annotations:
      summary: "WinPower energy data not updating"
      description: "Energy consumption hasn't updated for device {{ $labels.device_id }} in 15 minutes."

  # 计算延迟过高告警
  - alert: WinPowerEnergyCalculationSlow
    expr: histogram_quantile(0.95, rate(winpower_energy_calculation_duration_seconds_bucket[5m])) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "WinPower energy calculation slow"
      description: "95th percentile of energy calculation duration is {{ $value }}s for device {{ $labels.device_id }}."
```

## 总结

## Collector 集成与职责边界

- **唯一触发机制**：仅由 Collector 模块在采样到瞬时功率时调用 `Calculate(deviceID, power)` 触发能量计算
- **时间戳维护**：`LastUpdate` 由 Energy 模块在持久化时维护与写入存储，Collector 不直接设置此字段
- **负功率语义**：当功率为负时，累计能量以负值累加，表示净能量减少；当功率为 0 时，时间线推进但累计值不变
- **指标归属**：`winpower_energy_total_wh` 由 Energy 模块更新；`winpower_power_watts` 由 Collector 更新
- **模块职责**：各模块按照职责分工协同工作

电能计算模块采用**极简单线程架构**，在保持原有简洁设计的基础上彻底简化了实现，具有以下特点：

### 设计优势

1. **极简性**: 代码结构最简化，单一服务类，易于理解和维护
2. **绝对可靠性**: 全局串行执行，彻底消除并发问题和数据竞争
3. **零配置**: 无需复杂的队列配置和参数调优
4. **存储解耦**: 完全依赖storage模块，不直接操作文件系统
5. **专注UPS**: 针对UPS设备优化
6. **调试友好**: 线性执行路径，问题定位简单

### 技术特点

- **全局锁**: 单一读写锁确保所有操作串行执行
- **直接计算**: 无任务缓冲和队列调度，直接执行计算逻辑
- **简单统计**: 基础的计算统计信息，满足监控需求
- **错误处理**: 线性执行路径，错误处理简单直观

### 性能特点

- **低延迟**: 对于WinPower G2系统（5秒采集间隔），串行延迟可忽略
- **低开销**: 无队列管理和任务调度的额外开销
- **预测性强**: 执行时间完全可预测，无并发抖动
- **内存友好**: 极低的内存占用

### 适用场景

- WinPower G2系统的中小规模UPS设备电能监控
- 对可靠性要求高于性能要求的场景
- 开发和维护资源有限的项目
- 需要简单易懂代码的团队
- 首次实施电能监控功能的项目

该极简设计为WinPower G2 Prometheus Exporter提供了最简单可靠的电能计算功能，通过消除并发复杂性，在确保数据安全性的前提下大幅提升了代码的可维护性和可理解性。