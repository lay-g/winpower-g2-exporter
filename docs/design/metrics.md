# Metrics模块设计文档

## 概述

Metrics模块是WinPower G2 Prometheus Exporter的指标管理与暴露层，负责收集、管理和暴露各类监控指标。该模块作为Server模块的依赖，提供标准的Gin Handler作为唯一入口，专门处理Prometheus格式的指标数据。

### 设计目标

- **统一接口**: 提供标准的Gin Handler，与Server模块无缝集成
- **双类指标**: 管理exporter自监控指标和winpower设备指标
- **数据协调**: 与Collector模块紧密协作，获取最新的采集数据
- **标准暴露**: 遵循Prometheus指标规范，提供标准的/metrics端点

## 架构设计

### 模块职责

Metrics模块在系统架构中作为指标管理中心，负责指标的定义、更新和暴露：

```text
                    ┌─────────────────────────────────────┐
                    │           Metrics Module            │
                    │         (指标管理与暴露)              │
                    │                                     │
                    │  ┌─────────────────────────────────┐ │
                    │  │        MetricsService           │ │
                    │  │                                 │ │
                    │  │  ┌─────────────┐ ┌─────────────┐ │ │
                    │  │  │             │ │             │ │ │
                    │  │  │ 指标管理     │ │ HTTP Handler │ │ │
                    │  │  │             │ │             │ │ │
                    │  │  │ • 自监控指标 │ │ • /metrics  │ │ │
                    │  │  │ • 设备指标   │ │ • 数据协调   │ │ │
                    │  │  │ • 注册表管理 │ │ • 格式暴露   │ │ │
                    │  │  │             │ │             │ │ │
                    │  │  └─────────────┘ └─────────────┘ │ │
                    │  └─────────────────────────────────┘ │
                    └─────────────────────────────────────┘
                                   ▲
                                   │ 调用采集
                                   │
                    ┌─────────────────────────────────────┐
                    │         Collector Module            │
                    │        (数据采集与协调)              │
                    └─────────────────────────────────────┘
```

### 数据流程图

```text
                Prometheus抓取请求
                           │
                           ▼
                ┌─────────────────────────────────────┐
                │   Metrics.HandleMetrics()          │
                │         (Gin Handler)               │
                └─────────────────────────────────────┘
                           │
                           ▼
                ┌─────────────────────────────────────┐
                │   调用Collector采集最新数据          │
                │  Collector.CollectDeviceData()      │
                └─────────────────────────────────────┘
                           │
                           ▼
                ┌─────────────────────────────────────┐
                │     更新指标并返回Prometheus格式     │
                │     格式化输出                      │
                └─────────────────────────────────────┘
```

**关键流程说明：**

1. **请求处理**: Prometheus通过/metrics端点请求指标数据
2. **数据触发**: Metrics模块调用Collector触发即时数据采集
3. **指标更新**: 根据采集结果更新所有相关指标
4. **格式输出**: 将指标数据格式化为Prometheus标准格式返回

## 核心组件设计

### 1. MetricsService

#### 职责

- 管理Prometheus指标注册表
- 提供Gin Handler作为/metrics端点
- 与Collector模块协调获取最新数据
- 维护exporter自监控指标

#### 数据结构

```go
// MetricsService 指标管理服务
type MetricsService struct {
    registry       *prometheus.Registry    // Prometheus注册表
    collector      collector.CollectorInterface // Collector接口
    logger         *zap.Logger             // 日志记录器

    // Exporter自监控指标
    lastCollectionTime prometheus.Gauge    // 最后采集时间
    memoryUsage       *prometheus.GaugeVec // 内存使用指标

    // WinPower连接指标
    connectionStatus   prometheus.Gauge    // 连接状态
    tokenExpiry        prometheus.Gauge    // Token剩余有效期

    // 设备指标
    deviceMetrics      map[string]*DeviceMetrics // 设备指标映射
    mu                 sync.RWMutex              // 读写锁
}

// DeviceMetrics 设备指标集合
type DeviceMetrics struct {
    // 基本状态
    connected        prometheus.Gauge    // 连接状态
    loadPercent      prometheus.Gauge    // 负载百分比

    // 电气参数
    inputVoltage     prometheus.Gauge    // 输入电压
    outputVoltage    prometheus.Gauge    // 输出电压
    inputCurrent     prometheus.Gauge    // 输入电流
    outputCurrent    prometheus.Gauge    // 输出电流
    inputFrequency   prometheus.Gauge    // 输入频率
    outputFrequency  prometheus.Gauge    // 输出频率
    activePower      prometheus.Gauge    // 有功功率
    powerFactor      prometheus.Gauge    // 功率因数

    // 能耗指标
    instantPower     prometheus.Gauge    // 瞬时功率
    cumulativeEnergy prometheus.Gauge    // 累计电能
}
```

#### 核心方法

```go
// NewMetricsService 创建指标服务
func NewMetricsService(
    collector collector.CollectorInterface,
    logger *zap.Logger,
) *MetricsService

// HandleMetrics Gin Handler - /metrics端点
func (m *MetricsService) HandleMetrics(c *gin.Context)

// updateMetrics 更新所有指标（私有方法）
func (m *MetricsService) updateMetrics(ctx context.Context) error

// updateDeviceMetrics 更新设备指标（私有方法）
func (m *MetricsService) updateDeviceMetrics(deviceID string, data *winpower.DeviceData)
```

## 指标体系设计

### 指标分类

Metrics模块管理两大类指标：Exporter自监控指标和WinPower设备指标。所有指标直接映射自Collector模块的`DeviceCollectionInfo`结构体，确保数据完整性。

### 完整指标清单

采用 `namespace_subsystem_metric` 命名规范：`namespace=winpower`、`subsystem=exporter`。

#### 1. Exporter自监控指标

| 指标名称                                        | 类型      | 描述              | 标签            |
| ----------------------------------------------- | --------- | ----------------- | --------------- |
| `winpower_exporter_up`                          | Gauge     | Exporter运行状态  | `winpower_host` |
| `winpower_exporter_requests_total`              | Counter   | HTTP请求总数      | `winpower_host` |
| `winpower_exporter_request_duration_seconds`    | Histogram | 请求时延          | `winpower_host` |
| `winpower_exporter_collection_duration_seconds` | Histogram | 采集+计算整体耗时 | `winpower_host` |
| `winpower_exporter_scrape_errors_total`         | Counter   | 采集错误总数      | `winpower_host` |
| `winpower_exporter_token_refresh_total`         | Counter   | Token刷新次数     | `winpower_host` |
| `winpower_exporter_device_count`                | Gauge     | 发现的设备数量    | `winpower_host` |
| `winpower_exporter_memory_bytes`                | Gauge     | 内存使用量        | `winpower_host` |

#### 2. WinPower连接/认证指标

| 指标名称                             | 类型      | 描述             | 标签            |
| ------------------------------------ | --------- | ---------------- | --------------- |
| `winpower_connection_status`         | Gauge     | WinPower连接状态 | `winpower_host` |
| `winpower_auth_status`               | Gauge     | 认证状态         | `winpower_host` |
| `winpower_api_response_time_seconds` | Histogram | API响应时延      | `winpower_host` |
| `winpower_token_expiry_seconds`      | Gauge     | Token剩余有效期  | `winpower_host` |
| `winpower_token_valid`               | Gauge     | Token有效性      | `winpower_host` |

#### 3. 设备状态指标

| 指标名称                                | 类型  | 描述               | 标签                                                  |
| --------------------------------------- | ----- | ------------------ | ----------------------------------------------------- |
| `winpower_device_connected`             | Gauge | 设备连接状态       | `winpower_host`,`device_id`,`device_name`,`device_type` |
| `winpower_device_last_update_timestamp` | Gauge | 设备最后更新时间戳 | 同上                                                  |

#### 4. 电气参数指标

| 分类         | 指标名称                                  | 类型  | 标签（winpower_host, device_id, device_name, device_type） |
| ------------ | ----------------------------------------- | ----- | ----------------------------------------------- |
| **输入参数** | `winpower_device_input_voltage`           | Gauge | 输入电压(伏特)                                  |
|              | `winpower_device_input_frequency`         | Gauge | 输入频率(赫兹)                                  |
| **输出参数** | `winpower_device_output_voltage`          | Gauge | 输出电压(伏特)                                  |
|              | `winpower_device_output_current`          | Gauge | 输出电流(安培)                                  |
|              | `winpower_device_output_frequency`        | Gauge | 输出频率(赫兹)                                  |
|              | `winpower_device_output_voltage_type`     | Gauge | 输出电压类型                                    |
| **负载功率** | `winpower_device_load_percent`            | Gauge | 设备负载百分比（核心指标）                      |
|              | `winpower_device_load_total_watts`        | Gauge | 总负载有功功率(W)（核心指标）                   |
|              | `winpower_device_load_total_va`           | Gauge | 总负载视在功率(VA)                              |
|              | `winpower_device_load_watts_phase1`       | Gauge | 相1有功功率(W)                                  |
|              | `winpower_device_load_va_phase1`          | Gauge | 相1视在功率(VA)                                 |
| **电池参数** | `winpower_device_battery_charging`        | Gauge | 电池充电状态(1=充电)                            |
|              | `winpower_device_battery_voltage_percent` | Gauge | 电池电压百分比(%)                               |
|              | `winpower_device_battery_capacity`        | Gauge | 电池容量(%)                                     |
|              | `winpower_device_battery_remain_seconds`  | Gauge | 电池剩余时间(秒)                                |
|              | `winpower_device_battery_status`          | Gauge | 电池状态码                                      |
| **UPS状态**  | `winpower_device_ups_temperature`         | Gauge | UPS温度(°C)                                     |
|              | `winpower_device_ups_mode`                | Gauge | UPS工作模式                                     |
|              | `winpower_device_ups_status`              | Gauge | 设备状态码                                      |
|              | `winpower_device_ups_test_status`         | Gauge | 测试状态码                                      |
|              | `winpower_device_ups_fault_code`          | Gauge | UPS故障代码（额外标签：fault_code）             |
| **其他参数** | `winpower_device_input_transformer_type`  | Gauge | 输入变压器类型                                  |
| **能耗指标** | `winpower_device_cumulative_energy`       | Gauge | 累计电能(Wh，与Energy模块集成)                  |
|              | `winpower_power_watts`                    | Gauge | 瞬时功率(由Collector提供)                       |

### 标签策略

**统一标签配置**：
- **Exporter自监控指标**: 仅使用 `winpower_host` 标签
- **WinPower连接/认证指标**: 仅使用 `winpower_host` 标签
- **设备相关指标**: 使用 `winpower_host`,`device_id`,`device_name`,`device_type` 标签
- **故障代码**: 仅用于UPS故障指标，作为额外标签 `fault_code`

**高基数控制**：避免使用自由文本作为标签值，保持标签枚举值的有限性

## 接口设计

### 主要接口

```go
// MetricsService 指标服务接口
type MetricsService interface {
    // HandleMetrics 处理/metrics请求的Gin Handler
    HandleMetrics(c *gin.Context)
}

// CollectorInterface Collector模块接口（由collector模块提供）
type CollectorInterface interface {
    CollectDeviceData(ctx context.Context) (*winpower.ParsedDeviceData, error)
}
```

### HTTP Handler接口

```go
// HandleMetrics 提供标准的Prometheus /metrics端点
// GET /metrics
// 返回: text/plain; version=0.0.4; charset=utf-8
//
// 功能:
// 1. 调用Collector采集最新数据
// 2. 更新所有指标
// 3. 返回Prometheus格式的指标数据
func (m *MetricsService) HandleMetrics(c *gin.Context)
```

## 与Collector模块的集成

### 核心流程

Metrics模块的核心职责是协调数据采集和指标暴露：

```go
func (m *MetricsService) HandleMetrics(c *gin.Context) {
    // 1. 触发数据采集
    collectionResult, err := m.collector.CollectDeviceData(c.Request.Context())
    if err != nil {
        m.logger.Error("Failed to collect device data", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Collection failed"})
        return
    }

    // 2. 更新指标（基于CollectionResult结构体）
    m.updateMetrics(collectionResult)

    // 3. 返回Prometheus格式数据
    m.registry.ServeHTTP(c.Writer, c.Request)
}
```

### 指标更新逻辑

```go
func (m *MetricsService) updateMetrics(result *collector.CollectionResult) {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 更新自监控指标
    m.lastCollectionTime.SetToCurrentTime()
    m.collectionDuration.Observe(result.Duration.Seconds())
    m.deviceCount.Set(float64(result.DeviceCount))

    // 更新设备指标
    for deviceID, deviceInfo := range result.Devices {
        labels := prometheus.Labels{
            "device_id":   info.DeviceID,
            "device_name": info.DeviceName,
            "device_type": strconv.Itoa(info.DeviceType),
        }

        // 核心功率指标（LoadTotalWatt作为主要指标）
        m.loadTotalWatts.With(labels).Set(info.LoadTotalWatt)
        m.instantPower.With(labels).Set(info.LoadTotalWatt) // 瞬时功率与总负载相同

        // 其他电气参数...
        m.loadPercent.With(labels).Set(info.LoadPercent)
        m.outputVoltage.With(labels).Set(info.OutputVolt1)
        // ... 省略其他指标更新
    }
}
```

### 内存监控

```go
func (m *MetricsService) updateMemoryMetrics() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    m.memoryUsage.WithLabelValues("alloc").Set(float64(m.Alloc))
    m.memoryUsage.WithLabelValues("sys").Set(float64(m.Sys))
    m.memoryUsage.WithLabelValues("heap").Set(float64(m.HeapAlloc))
}
```

## 错误处理与测试

### 错误处理策略

1. **Collector调用失败**: 记录错误日志，返回HTTP 500状态码
2. **指标更新失败**: 记录错误日志，但不影响HTTP响应
3. **格式化输出失败**: 记录错误日志，返回HTTP 500状态码

### 测试设计

- **单元测试**: 使用 `prometheus/testutil` 断言指标值与标签；Mock Collector接口
- **集成测试**: 验证完整的`/metrics`端点响应
- **端到端测试**: 使用Prometheus抓取并校验格式与内容

## 部署与监控

### Prometheus配置

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'winpower-exporter'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: /metrics
```

### 常用PromQL查询

```promql
# 核心监控
winpower_device_load_total_watts           # 总负载有功功率（核心）
winpower_device_cumulative_energy          # 累计电能
winpower_device_connected == 1             # 在线设备

# 能耗分析
increase(winpower_device_cumulative_energy[1h]) / 1000    # 每小时能耗(kWh)
avg_over_time(winpower_device_load_total_watts[5m])        # 5分钟平均功率

# 系统健康
winpower_exporter_up == 1                           # Exporter状态
winpower_connection_status == 1                      # 连接状态
rate(winpower_exporter_scrape_errors_total[5m])      # 错误率
```

### 性能优化

- **指标更新**: 使用读写锁保护并发访问，批量更新减少锁竞争
- **内存管理**: 动态创建设备指标，定期清理不活跃设备
- **HTTP响应**: 使用Prometheus官方库高效格式化

### Histogram桶配置

- 请求/采集耗时: `[0.05, 0.1, 0.2, 0.5, 1, 2, 5]` 秒
- API响应时间: `[0.05, 0.1, 0.2, 0.5, 1]` 秒

## 总结

Metrics模块作为WinPower G2 Exporter的指标管理中心，具有以下核心特点：

1. **数据一致性**: 所有指标直接映射自Collector的`DeviceCollectionInfo`结构体
2. **核心突出**: `LoadTotalWatt`作为核心功率指标，支撑实时监控和能耗计算
3. **标签优化**: 统一标签策略，控制高基数，`FaultCode`作为标签便于聚合
4. **统一入口**: 提供标准Gin Handler，与Server模块无缝集成
5. **实时协调**: `/metrics`端点触发即时数据采集，确保数据新鲜度
6. **标准兼容**: 严格遵循Prometheus指标规范，便于集成和监控

重构后的Metrics模块减少了不必要的复杂度，提升了可运维性，结合推荐的PromQL与Grafana可快速搭建完整的能耗与健康可视化方案。