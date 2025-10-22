# Metrics 模块架构设计

## 概述

Metrics 模块是 WinPower G2 Exporter 的输出层，负责将采集的设备数据转化为标准的 Prometheus 时序指标，并通过 HTTP `/metrics` 端点暴露。本设计遵循模块化、标准化的原则，与已实现的 winpower、energy、storage、logging 模块紧密集成。

## 设计原则

1. **标准化**: 严格遵循 Prometheus 指标命名和标签规范
2. **架构优化**: 采用单 Registry 和合理的锁粒度设计
3. **简洁性**: 暴露必要且稳定的指标集合，避免过度标签
4. **可观测性**: 提供完整的 Exporter 自监控指标
5. **类型安全**: 提供类型安全的指标更新接口

## 架构设计

### 整体架构图

```
┌──────────────────────────────────────────────────────────────┐
│                        Metrics Module                         │
├──────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌─────────────────────────────────────┐  │
│  │ MetricManager │◄►│ Prometheus Registry (single)       │  │
│  │               │  │                                     │  │
│  │ - 指标注册     │  │ - Exporter自监控                   │  │
│  │ - 指标更新     │  │ - WinPower连接/认证                │  │
│  │ - HTTP暴露     │  │ - 设备/电源数据                    │  │
│  │ - 标签管理     │  │ - 能耗数据                        │  │
│  └───────────────┘  └─────────────────────────────────────┘  │
│          ▲                          ▲                         │
│          │                          │                         │
│   数据源模块调用               Prometheus抓取                   │
│   (winpower/energy)           (HTTP /metrics)                 │
├──────────────────────────────────────────────────────────────┤
│                    HTTP /metrics Handler                      │
│  - 统一入口点                                                  │
│  - 返回最新指标快照                                            │
│  - 支持OpenMetrics格式                                        │
└──────────────────────────────────────────────────────────────┘
```

### 模块边界与接口

#### 输入边界
- **WinPower 模块**: 提供设备连接状态、认证状态、设备实时数据
- **Energy 模块**: 提供累计电能数据
- **Config 模块**: 提供指标配置参数（命名空间、子系统等）

#### 输出边界
- **HTTP /metrics**: 对 Prometheus 提供标准格式的指标数据
- **日志输出**: 结构化日志记录指标更新和异常情况

## 核心组件设计

### 1. MetricManager

指标管理器的核心组件，负责所有指标的生命周期管理。

```go
type MetricManager struct {
    registry *prometheus.Registry
    logger   *zap.Logger
    config   MetricManagerConfig

    // Exporter 自监控指标
    exporterMetrics ExporterMetrics

    // WinPower 连接/认证指标
    connectionMetrics ConnectionMetrics

    // 设备/电源指标
    deviceMetrics DeviceMetrics

    // 能耗指标
    energyMetrics EnergyMetrics

    mu sync.RWMutex
}
```

#### 指标分类结构

```go
type ExporterMetrics struct {
    up              prometheus.Gauge
    requestsTotal   *prometheus.CounterVec
    requestDuration *prometheus.HistogramVec
    scrapeErrors    *prometheus.CounterVec
    collectionTime  *prometheus.HistogramVec
    tokenRefresh    *prometheus.CounterVec
    deviceCount     *prometheus.GaugeVec
}

type ConnectionMetrics struct {
    connectionStatus *prometheus.GaugeVec
    authStatus       *prometheus.GaugeVec
    apiResponseTime  *prometheus.HistogramVec
    tokenExpiry      *prometheus.GaugeVec
    tokenValid       *prometheus.GaugeVec
}

type DeviceMetrics struct {
    deviceConnected  *prometheus.GaugeVec
    loadPercent      *prometheus.GaugeVec
    inputVoltage     *prometheus.GaugeVec
    outputVoltage    *prometheus.GaugeVec
    inputCurrent     *prometheus.GaugeVec
    outputCurrent    *prometheus.GaugeVec
    inputFrequency   *prometheus.GaugeVec
    outputFrequency  *prometheus.GaugeVec
    inputWatts       *prometheus.GaugeVec
    outputWatts      *prometheus.GaugeVec
    powerFactorOut   *prometheus.GaugeVec
}

type EnergyMetrics struct {
    energyTotalWh *prometheus.GaugeVec
    powerWatts    *prometheus.GaugeVec
}
```

### 2. 配置管理

```go
type MetricManagerConfig struct {
    Namespace string                `json:"namespace" yaml:"namespace"`     // 默认: "winpower"
    Subsystem string                `json:"subsystem" yaml:"subsystem"`     // 默认: "exporter"
    Registry  *prometheus.Registry  `json:"-" yaml:"-"`                    // 可选注入的Registry

    // 直方图配置
    RequestDurationBuckets []float64 `json:"request_duration_buckets" yaml:"request_duration_buckets"`
    CollectionDurationBuckets []float64 `json:"collection_duration_buckets" yaml:"collection_duration_buckets"`
    APIResponseBuckets     []float64 `json:"api_response_buckets" yaml:"api_response_buckets"`
}
```

### 3. 接口设计

```go
type MetricManagerInterface interface {
    // HTTP暴露
    Handler() http.Handler

    // Exporter自监控
    SetUp(status bool)
    ObserveRequest(host, method, code string, duration time.Duration)
    IncScrapeError(host, errorType string)
    ObserveCollection(status string, duration time.Duration)
    IncTokenRefresh(host, result string)
    SetDeviceCount(host, deviceType string, count float64)

    // WinPower连接/认证
    SetConnectionStatus(host, connectionType string, status float64)
    SetAuthStatus(host, authMethod string, status float64)
    ObserveAPI(host, endpoint string, duration time.Duration)
    SetTokenExpiry(host, userID string, seconds float64)
    SetTokenValid(host, userID string, valid float64)

    // 设备/电源数据
    SetDeviceConnected(deviceID, deviceName, deviceType string, connected float64)
    SetLoadPercent(deviceID, deviceName, deviceType, phase string, pct float64)
    SetElectricalData(deviceID, deviceName, deviceType, phase string,
        inV, outV, inA, outA, inHz, outHz, inW, outW, pfo float64)

    // 能耗数据
    SetPowerWatts(deviceID, deviceName, deviceType string, watts float64)
    SetEnergyTotalWh(deviceID, deviceName, deviceType string, wh float64)
}
```

## 指标设计

### 命名规范

采用 `namespace_subsystem_metric` 格式，默认：
- **命名空间**: `winpower`
- **子系统**: `exporter`
- **示例**: `winpower_exporter_up`

### 指标清单

#### 1. Exporter 自监控指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_exporter_up` | Gauge | Exporter运行状态 | `winpower_host`, `version` |
| `winpower_exporter_requests_total` | Counter | HTTP请求总数 | `winpower_host`, `method`, `status_code` |
| `winpower_exporter_request_duration_seconds` | Histogram | HTTP请求时延 | `winpower_host`, `method`, `status_code` |
| `winpower_exporter_scrape_errors_total` | Counter | 采集错误总数 | `winpower_host`, `error_type` |
| `winpower_exporter_collection_duration_seconds` | Histogram | 采集+累计整体耗时 | `winpower_host`, `status` |
| `winpower_exporter_token_refresh_total` | Counter | Token刷新次数 | `winpower_host`, `result` |
| `winpower_exporter_device_count` | Gauge | 发现的设备数量 | `winpower_host`, `device_type` |

#### 2. WinPower 连接/认证指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_connection_status` | Gauge | 连接状态 | `winpower_host`, `connection_type` |
| `winpower_auth_status` | Gauge | 认证状态 | `winpower_host`, `auth_method` |
| `winpower_api_response_time_seconds` | Histogram | API响应时延 | `winpower_host`, `api_endpoint` |
| `winpower_token_expiry_seconds` | Gauge | Token剩余有效期 | `winpower_host`, `user_id` |
| `winpower_token_valid` | Gauge | Token有效性 | `winpower_host`, `user_id` |

#### 3. 设备/电源指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_device_connected` | Gauge | 设备连接状态 | `device_id`, `device_name`, `device_type` |
| `winpower_load_percent` | Gauge | 负载百分比 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_input_voltage_volts` | Gauge | 输入电压 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_output_voltage_volts` | Gauge | 输出电压 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_input_current_amperes` | Gauge | 输入电流 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_output_current_amperes` | Gauge | 输出电流 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_input_frequency_hertz` | Gauge | 输入频率 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_output_frequency_hertz` | Gauge | 输出频率 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_input_watts` | Gauge | 输入有功功率 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_output_watts` | Gauge | 输出有功功率 | `device_id`, `device_name`, `device_type`, `phase` |
| `winpower_output_power_factor` | Gauge | 输出功率因数 | `device_id`, `device_name`, `device_type`, `phase` |

#### 4. 能耗指标

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_power_watts` | Gauge | 瞬时功率 | `device_id`, `device_name`, `device_type` |
| `winpower_energy_total_wh` | Gauge | 累计电能 | `device_id`, `device_name`, `device_type` |

### 标签策略

#### 设备标签
- `device_id`: 设备唯一标识符（必须）
- `device_name`: 设备名称（必须）
- `device_type`: 设备类型（必须）
- `phase`: 相线标识（可选，仅在多相设备使用）

#### 状态标签
- `winpower_host`: WinPower主机地址（必须）
- `connection_type`: 连接类型（http/https）
- `auth_method`: 认证方法（token等）
- `status`: 状态枚举（ok/err, 1/0）
- `result`: 结果枚举（ok/err, success/failure）

#### 高基数控制
- 严格限制标签枚举值，避免自由文本
- 设备标签采用固定组合，不随意扩展
- 状态类标签使用小集合枚举值

## 架构优化

### 锁策略
- 使用读写锁 (`sync.RWMutex`) 保护指标更新
- 读操作（HTTP暴露）使用读锁，允许并发访问
- 写操作（指标更新）使用写锁，确保数据一致性

### 内存管理
- 单一 Prometheus Registry，避免重复注册
- 预分配指标向量，避免动态扩容
- 控制指标数量，避免内存泄漏

### 实现优化
- 指标更新操作尽可能轻量
- 避免在关键路径进行复杂计算
- 批量更新减少锁竞争

## 集成设计

### 与 WinPower 模块集成

```go
// WinPower 模块采集成功后调用
func (w *WinPowerClient) updateMetrics(data DeviceData) {
    w.metrics.SetDeviceConnected(data.ID, data.Name, data.Type, boolToFloat(data.Connected))
    w.metrics.SetElectricalData(data.ID, data.Name, data.Type, data.Phase,
        data.InputVoltage, data.OutputVoltage,
        data.InputCurrent, data.OutputCurrent,
        data.InputFrequency, data.OutputFrequency,
        data.InputWatts, data.OutputWatts,
        data.PowerFactor)
}
```

### 与 Energy 模块集成

```go
// Energy 模块计算完成后调用
func (e *EnergyService) updateMetrics(deviceID string, power, energy float64) {
    e.metrics.SetPowerWatts(deviceID, e.deviceName, e.deviceType, power)
    e.metrics.SetEnergyTotalWh(deviceID, e.deviceName, e.deviceType, energy)
}
```

### HTTP 集成

```go
func (m *MetricManager) Handler() http.Handler {
    return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
        EnableOpenMetrics: true,
    })
}
```

## 错误处理

### 指标更新错误
- 记录错误日志，但不影响主流程
- 使用默认值或跳过更新，避免程序崩溃
- 提供错误统计指标，便于监控

### HTTP 错误
- 优雅处理网络异常
- 返回标准 HTTP 状态码
- 记录访问日志和错误统计

## 测试策略

### 单元测试
- 使用 `prometheus/testutil` 进行指标断言
- Mock 外部依赖，隔离测试环境
- 覆盖所有指标更新路径

### 集成测试
- 端到端测试完整的指标流转
- 验证 HTTP `/metrics` 端点输出格式
- 测试数据访问的一致性

## 配置示例

```yaml
metrics:
  namespace: "winpower"
  subsystem: "exporter"
  request_duration_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2, 5]
  collection_duration_buckets: [0.1, 0.2, 0.5, 1, 2, 5, 10]
  api_response_buckets: [0.05, 0.1, 0.2, 0.5, 1]
```

## 扩展性考虑

### 指标扩展
- 支持通过配置添加自定义指标
- 预留指标扩展接口
- 支持指标动态注册（如需要）

### 标签扩展
- 支持可选标签配置
- 预留业务特定标签空间
- 支持标签映射和转换

### 格式扩展
- 支持 OpenMetrics 格式
- 支持其他监控系统集成（如需要）
- 预留格式转换接口

这个架构设计确保了 Metrics 模块能够高效、可靠地提供标准化的监控指标，同时保持与现有模块的良好集成和扩展性。