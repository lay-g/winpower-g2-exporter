# Collector Module

## 概述

Collector模块是WinPower G2 Prometheus Exporter的核心协调组件，负责：

- 从WinPower模块采集设备数据
- 触发Energy模块进行电能计算
- 协调数据流程
- 为Scheduler和Metrics模块提供统一的采集接口

## 架构设计

### 核心职责

```
┌─────────────┐
│  Scheduler  │ ──┐
└─────────────┘   │
                  │
┌─────────────┐   │    ┌──────────────┐    ┌──────────────┐
│   Metrics   │ ──┼───→│  Collector   │───→│  WinPower    │
└─────────────┘   │    │   Service    │    │   Client     │
                  │    └──────────────┘    └──────────────┘
                  │            │
                  │            ▼
                  │    ┌──────────────┐
                  └───→│   Energy     │
                       │   Service    │
                       └──────────────┘
```

### 设计原则

1. **单一职责**: 专注于数据采集协调和电能计算触发
2. **依赖倒置**: 通过接口解耦对具体模块的依赖
3. **错误隔离**: 单个设备失败不影响整体采集流程
4. **性能优先**: 最小化数据采集和计算的开销

## 核心接口

### CollectorInterface

```go
type CollectorInterface interface {
    CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}
```

主要的数据采集入口，返回包含所有设备信息的采集结果。

### 依赖接口

```go
// WinPowerClient - WinPower数据采集客户端
type WinPowerClient interface {
    CollectDeviceData(ctx context.Context) ([]winpower.ParsedDeviceData, error)
    GetConnectionStatus() bool
    GetLastCollectionTime() time.Time
}

// EnergyCalculator - 电能计算器
type EnergyCalculator interface {
    Calculate(deviceID string, power float64) (float64, error)
    Get(deviceID string) (float64, error)
}
```

## 数据结构

### CollectionResult

采集结果的顶层结构：

```go
type CollectionResult struct {
    Success        bool                             // 采集是否成功
    DeviceCount    int                              // 设备数量
    Devices        map[string]*DeviceCollectionInfo // 设备详细信息
    CollectionTime time.Time                        // 采集时间戳
    Duration       time.Duration                    // 采集耗时
    ErrorMessage   string                           // 错误信息(如果有)
}
```

### DeviceCollectionInfo

单个设备的完整信息：

```go
type DeviceCollectionInfo struct {
    // 基本信息
    DeviceID       string
    DeviceName     string
    DeviceType     int
    DeviceModel    string
    Connected      bool
    
    // 电气参数
    InputVolt1     float64
    OutputVolt1    float64
    LoadTotalWatt  float64  // 核心字段，用于电能计算
    
    // 电池参数
    BatCapacity    float64
    BatRemainTime  int
    
    // 电能计算结果
    EnergyCalculated bool
    EnergyValue      float64  // 累计电能(Wh)
    
    // 错误信息
    ErrorMsg       string
}
```

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/lay-g/winpower-g2-exporter/internal/collector"
    "github.com/lay-g/winpower-g2-exporter/internal/energy"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
    "github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

func main() {
    // 创建依赖
    logger := log.NewLogger(log.DefaultConfig())
    winpowerClient := winpower.NewClient(/* ... */)
    energyService := energy.NewEnergyService(/* ... */)
    
    // 创建Collector服务
    collectorService, err := collector.NewCollectorService(
        winpowerClient,
        energyService,
        logger,
    )
    if err != nil {
        log.Fatal("Failed to create collector:", err)
    }
    
    // 采集设备数据
    ctx := context.Background()
    result, err := collectorService.CollectDeviceData(ctx)
    if err != nil {
        log.Fatal("Collection failed:", err)
    }
    
    // 处理结果
    for deviceID, info := range result.Devices {
        if info.EnergyCalculated {
            log.Printf("Device %s: Power=%.2fW, Energy=%.2fWh\n",
                deviceID, info.LoadTotalWatt, info.EnergyValue)
        } else {
            log.Printf("Device %s: Error - %s\n", deviceID, info.ErrorMsg)
        }
    }
}
```

### 在Scheduler中使用

```go
// Scheduler定期调用Collector
func (s *Scheduler) collectData() {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := s.collector.CollectDeviceData(ctx)
    if err != nil {
        s.logger.Error("Collection failed", log.Err(err))
        return
    }
    
    s.logger.Info("Collection completed",
        log.Int("device_count", result.DeviceCount),
        log.Duration("duration", result.Duration))
}
```

### 在Metrics中使用

```go
// Metrics按需调用Collector获取最新数据
func (m *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
    ctx := context.Background()
    result, err := m.collector.CollectDeviceData(ctx)
    if err != nil {
        m.logger.Error("Failed to collect data", log.Err(err))
        return
    }
    
    for _, device := range result.Devices {
        if device.EnergyCalculated {
            ch <- prometheus.MustNewConstMetric(
                m.energyDesc,
                prometheus.CounterValue,
                device.EnergyValue,
                device.DeviceID,
            )
        }
    }
}
```

## 错误处理

### 错误类型

```go
const (
    ErrorTypeWinPowerCollection  // WinPower采集错误
    ErrorTypeEnergyCalculation   // 电能计算错误
    ErrorTypeDataConversion      // 数据转换错误
)
```

### 错误隔离策略

1. **WinPower采集错误**: 直接返回，不进行设备处理
2. **单设备能耗计算错误**: 记录错误但继续处理其他设备
3. **数据转换错误**: 使用默认值或跳过字段

### 错误处理示例

```go
result, err := collector.CollectDeviceData(ctx)
if err != nil {
    // WinPower级别的错误
    log.Error("Collection failed", log.Err(err))
    return
}

// 检查设备级别的错误
for deviceID, info := range result.Devices {
    if !info.EnergyCalculated {
        log.Warn("Device energy calculation failed",
            log.String("device_id", deviceID),
            log.String("error", info.ErrorMsg))
    }
}
```

## 性能考虑

### 优化策略

1. **并发处理**: 对多个设备的电能计算进行并发处理(未来优化)
2. **内存管理**: 及时释放不需要的数据结构
3. **日志优化**: 使用结构化日志，避免字符串拼接

### 性能指标

- **采集延迟**: 从开始到返回结果的总时间
- **设备处理速率**: 单位时间内能处理的设备数量
- **错误率**: 采集失败的比例

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./internal/collector/... -v

# 查看测试覆盖率
go test ./internal/collector/... -coverprofile=coverage.txt
go tool cover -func=coverage.txt

# HTML格式查看覆盖率
go tool cover -html=coverage.txt
```

### Mock对象

模块提供了完整的Mock实现用于测试：

- `MockCollector`: CollectorInterface的Mock
- `MockWinPowerClient`: WinPowerClient的Mock
- `MockEnergyCalculator`: EnergyCalculator的Mock

使用示例：

```go
func TestMyComponent(t *testing.T) {
    mockCollector := &collector.MockCollector{
        CollectDeviceDataFunc: func(ctx context.Context) (*collector.CollectionResult, error) {
            return &collector.CollectionResult{
                Success: true,
                Devices: map[string]*collector.DeviceCollectionInfo{
                    "device1": {
                        DeviceID: "device1",
                        LoadTotalWatt: 1000.0,
                        EnergyValue: 500.0,
                    },
                },
            }, nil
        },
    }
    
    // 使用mockCollector进行测试
    result, err := mockCollector.CollectDeviceData(context.Background())
    // ...
}
```

## 日志记录

Collector模块使用结构化日志记录关键操作：

```go
// 采集开始
logger.Debug("Starting device data collection")

// 采集成功
logger.Info("Device data collection completed",
    log.Int("device_count", result.DeviceCount),
    log.Bool("success", result.Success),
    log.Duration("duration", result.Duration))

// WinPower采集失败
logger.Error("Failed to collect data from WinPower", log.Err(err))

// 设备级错误
logger.Warn("Energy calculation failed for device",
    log.String("device_id", device.DeviceID),
    log.Err(err))
```

## 依赖关系

### 上游依赖

- `internal/winpower`: WinPower客户端，提供设备数据采集
- `internal/energy`: 电能服务，提供电能计算功能
- `internal/pkgs/log`: 日志模块，提供结构化日志

### 接口实现检查

Collector模块定义了自己需要的接口（`WinPowerClient` 和 `EnergyCalculator`），遵循依赖倒置原则。

**编译时检查** (`interface_check.go`):
```go
// 如果实现不匹配，编译会失败
var (
    _ WinPowerClient   = (*winpower.Client)(nil)
    _ EnergyCalculator = (*energy.EnergyService)(nil)
)
```

**运行时测试** (`interface_check_test.go`):
- `TestInterfaceCompliance`: 验证各个具体实现满足接口定义
- `TestInterfaceCompatibility`: 验证真实实现可以与Collector协同工作

运行接口检查测试：
```bash
# 验证接口实现
go test ./internal/collector/... -run TestInterface -v

# 编译检查（如果接口不匹配会失败）
go build ./internal/collector/...
```

### 下游使用者

- `internal/scheduler`: 调度器，定期触发数据采集
- `internal/metrics`: 指标收集器，按需获取设备数据

## 设计文档

详细的设计文档参见：

- [Collector模块设计](../../../docs/design/collector.md)
- [系统架构设计](../../../docs/design/architecture.md)
- [提案设计文档](../../../openspec/changes/implement-collector-module/design.md)

## 待办事项

- [ ] 实现并发处理优化(可选)
- [ ] 添加数据缓存机制(可选)
- [ ] 支持数据过滤和转换规则(可选)

## 版本历史

- v1.0.0 (2025-10-30): 初始实现
  - 基础数据采集功能
  - 电能计算触发
  - 完整的错误处理
  - 100%测试覆盖率
