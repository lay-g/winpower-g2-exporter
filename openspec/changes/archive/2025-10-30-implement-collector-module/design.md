# Collector Module Design Document

## 设计基础

本文档基于原始设计文档 [`docs/design/collector.md`](../../../docs/design/collector.md) 进行详细的技术设计实现，补充了具体的接口定义、数据结构和实现细节。

## 架构设计

### 模块定位

Collector模块在WinPower G2 Prometheus Exporter中承担核心协调职责，是数据流的中枢组件。它连接上游的WinPower数据源和下游的Energy计算模块，同时为Scheduler和Metrics模块提供统一的数据采集接口。

### 设计原则

1. **单一职责**: 专注于数据采集协调和电能计算触发
2. **依赖倒置**: 通过接口解耦对具体模块的依赖
3. **错误隔离**: 单个设备失败不影响整体采集流程
4. **性能优先**: 最小化数据采集和计算的开销

## 组件设计

### 核心服务架构

```
┌─────────────────────────────────────────────────────────────┐
│                  Collector Module                          │
│                (数据采集与触发协调)                          │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              CollectorService                        │   │
│  │                                                     │   │
│  │  ┌─────────────────┐    ┌─────────────────────┐     │   │
│  │  │                 │    │                     │     │   │
│  │  │   数据获取协调   │    │    电能计算触发      │     │   │
│  │  │                 │    │                     │     │   │
│  │  │ • 调用WinPower  │    │ • 触发Energy模块    │     │   │
│  │  │ • 解析响应数据   │    │ • 处理计算错误      │     │   │
│  │  │ • 处理采集错误   │    │ • 记录执行日志      │     │   │
│  │  │                 │    │                     │     │   │
│  │  └─────────────────┘    └─────────────────────┘     │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 数据流设计

```
调度器触发 (每5秒)               Metrics模块触发 (按需)
        │                              │
        ▼                              ▼
┌─────────────────────────────────────────────────────┐
│         Collector.CollectDeviceData()              │
│            (统一的数据采集入口)                      │
└─────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────┐
│ 调用WinPower        │
│ CollectDeviceData   │
└─────────────────────┘
           │
           ▼
┌─────────────────────┐
│ 获取设备数据         │
│ • LoadTotalWatt     │
│ • 设备状态信息       │
└─────────────────────┘
           │
           ▼
┌─────────────────────┐
│ 遍历每个设备         │
│ 触发电能计算         │
└─────────────────────┘
           │
           ▼
┌─────────────────────┐
│ Energy.Calculate    │
│ (deviceID, power)   │
└─────────────────────┘
           │
           ▼
┌─────────────────────┐
│ 记录执行日志         │
│ 返回采集结果         │
└─────────────────────┘
```

## 接口设计

### 核心接口

```go
type CollectorInterface interface {
    CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}
```

### 依赖接口

```go
// WinPowerClient 由winpower模块提供
type WinPowerClient interface {
    CollectDeviceData(ctx context.Context) (*winpower.ParsedDeviceData, error)
    GetConnectionStatus() bool
    GetLastCollectionTime() time.Time
}

// EnergyInterface 由energy模块提供
type EnergyInterface interface {
    Calculate(deviceID string, power float64) (float64, error)
    Get(deviceID string) (float64, error)
}
```

## 数据结构设计

### 采集结果结构

```go
type CollectionResult struct {
    Success        bool                              `json:"success"`
    DeviceCount    int                               `json:"device_count"`
    Devices        map[string]*DeviceCollectionInfo  `json:"devices"`
    CollectionTime time.Time                         `json:"collection_time"`
    Duration       time.Duration                     `json:"duration"`
    ErrorMessage   string                            `json:"error_message"`
}

type DeviceCollectionInfo struct {
    // 基本信息
    DeviceID       string    `json:"device_id"`
    DeviceName     string    `json:"device_name"`
    DeviceType     int       `json:"device_type"`
    DeviceModel    string    `json:"device_model"`
    Connected      bool      `json:"connected"`
    LastUpdateTime time.Time `json:"last_update_time"`

    // 电气参数
    InputVolt1         float64 `json:"input_volt_1"`
    InputFreq          float64 `json:"input_freq"`
    OutputVolt1        float64 `json:"output_volt_1"`
    OutputCurrent1     float64 `json:"output_current_1"`
    OutputFreq         float64 `json:"output_freq"`
    OutputVoltageType  int     `json:"output_voltage_type"`

    // 负载和功率参数
    LoadPercent    float64 `json:"load_percent"`
    LoadTotalWatt  float64 `json:"load_total_watt"`   // 核心字段，用于能耗计算
    LoadTotalVa    float64 `json:"load_total_va"`
    LoadWatt1      float64 `json:"load_watt_1"`
    LoadVa1        float64 `json:"load_va_1"`

    // 电池参数
    IsCharging     bool    `json:"is_charging"`
    BatVoltP       float64 `json:"bat_volt_p"`
    BatCapacity    float64 `json:"bat_capacity"`
    BatRemainTime  int     `json:"bat_remain_time"`
    BatteryStatus  int     `json:"battery_status"`

    // UPS状态参数
    UpsTemperature float64 `json:"ups_temperature"`
    Mode           int     `json:"mode"`
    Status         int     `json:"status"`
    TestStatus     int     `json:"test_status"`
    FaultCode      string  `json:"fault_code"`

    // 其他参数
    InputTransformerType int `json:"input_transformer_type"`

    // 能耗计算结果
    EnergyCalculated bool    `json:"energy_calculated"`
    EnergyValue      float64 `json:"energy_value"`

    // 错误信息
    ErrorMsg string `json:"error_msg"`
}
```

## 错误处理策略

### 分层错误处理

1. **WinPower采集错误**:
   - 记录详细错误信息
   - 直接返回给上层处理
   - 不进行自动重试

2. **电能计算错误**:
   - 单设备失败不影响其他设备
   - 记录设备级错误信息
   - 继续处理其他设备

3. **数据转换错误**:
   - 记录转换失败的详细信息
   - 使用默认值或跳过该字段
   - 不中断整体采集流程

### 错误分类

```go
type ErrorType int

const (
    ErrorTypeWinPowerCollection ErrorType = iota
    ErrorTypeEnergyCalculation
    ErrorTypeDataConversion
)

type CollectionError struct {
    Type      ErrorType `json:"type"`
    DeviceID  string    `json:"device_id,omitempty"`
    Message   string    `json:"message"`
    Cause     error     `json:"cause,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

## 性能考虑

### 优化策略

1. **并发处理**: 对多个设备的电能计算进行并发处理
2. **内存管理**: 及时释放不需要的数据结构
3. **日志优化**: 使用结构化日志，避免字符串拼接
4. **错误缓存**: 避免重复记录相同的错误信息

### 性能指标

- **采集延迟**: 从开始到返回结果的总时间
- **内存使用**: 采集过程中的内存峰值
- **错误率**: 采集失败的比例
- **吞吐量**: 单位时间内能处理的设备数量

## 测试策略

### 单元测试

- **核心逻辑测试**: 测试数据采集和转换逻辑
- **错误处理测试**: 验证各种错误场景的处理
- **Mock测试**: 使用Mock对象隔离外部依赖

### 集成测试

- **端到端测试**: 验证完整的数据流
- **性能测试**: 验证性能指标达标
- **错误恢复测试**: 验证错误恢复机制

## 部署考虑

### 配置参数

- **超时设置**: 数据采集和计算的超时时间
- **并发控制**: 并发处理的设备数量限制
- **重试策略**: 是否启用重试机制（初期不建议）

### 监控指标

Collector模块的监控指标由Metrics模块统一管理：
- 采集成功率
- 平均采集耗时
- 错误分类统计
- 设备连接状态

## 安全考虑

### 数据安全

- **敏感信息**: 避免在日志中记录敏感的设备信息
- **数据验证**: 验证从WinPower获取的数据格式和范围

### 访问控制

- **接口权限**: 通过依赖注入控制访问权限
- **资源限制**: 限制资源使用，防止资源耗尽

## 扩展性设计

### 接口扩展

- **插件化**: 支持不同类型的数据源
- **版本兼容**: 支持WinPower API的版本升级

### 功能扩展

- **数据过滤**: 支持按条件过滤设备数据
- **数据转换**: 支持自定义数据转换规则
- **缓存机制**: 可选的数据缓存功能