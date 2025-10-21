# Metrics 模块

WinPower G2 Exporter 的指标管理模块，负责将采集的设备数据转化为标准的 Prometheus 时序指标，并通过 HTTP `/metrics` 端点暴露。

## 概述

Metrics 模块是导出器的输出层，提供标准化的 Prometheus 指标管理功能。该模块遵循模块化、标准化的设计原则，与 WinPower、Energy、Storage、Logging 模块紧密集成，为监控和告警系统提供可靠的数据支撑。

## 功能特性

- **标准化指标**: 严格遵循 Prometheus 指标命名和标签规范
- **四类指标覆盖**: Exporter自监控、WinPower连接/认证、设备/电源数据、能耗数据
- **HTTP暴露**: 提供 `/metrics` 端点，支持 Prometheus 抓取
- **标签管理**: 统一标签策略，控制基数避免高基数风险
- **类型安全**: 提供类型安全的指标更新接口
- **性能优化**: 单 Registry 设计，合理的锁粒度

## 架构设计

### 核心组件

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
└──────────────────────────────────────────────────────────────┘
```

### 模块依赖

- **依赖**: logging模块、winpower模块、energy模块
- **被依赖**: server模块
- **配置**: config模块统一加载metrics配置

## 指标分类

### 1. Exporter 自监控指标

监控导出器自身的运行状态和性能指标。

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_exporter_up` | Gauge | Exporter运行状态 | `winpower_host`, `version` |
| `winpower_exporter_requests_total` | Counter | HTTP请求总数 | `winpower_host`, `method`, `status_code` |
| `winpower_exporter_request_duration_seconds` | Histogram | HTTP请求时延 | `winpower_host`, `method`, `status_code` |
| `winpower_exporter_scrape_errors_total` | Counter | 采集错误总数 | `winpower_host`, `error_type` |
| `winpower_exporter_collection_duration_seconds` | Histogram | 采集+累计整体耗时 | `winpower_host`, `status` |
| `winpower_exporter_token_refresh_total` | Counter | Token刷新次数 | `winpower_host`, `result` |
| `winpower_exporter_device_count` | Gauge | 发现的设备数量 | `winpower_host`, `device_type` |

### 2. WinPower 连接/认证指标

监控与 WinPower 服务的连接状态和认证情况。

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_connection_status` | Gauge | 连接状态 | `winpower_host`, `connection_type` |
| `winpower_auth_status` | Gauge | 认证状态 | `winpower_host`, `auth_method` |
| `winpower_api_response_time_seconds` | Histogram | API响应时延 | `winpower_host`, `api_endpoint` |
| `winpower_token_expiry_seconds` | Gauge | Token剩余有效期 | `winpower_host`, `user_id` |
| `winpower_token_valid` | Gauge | Token有效性 | `winpower_host`, `user_id` |

### 3. 设备/电源指标

监控设备的电气参数和运行状态。

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

### 4. 能耗指标

监控设备的能耗数据，与 Energy 模块紧密集成。

| 指标名称 | 类型 | 描述 | 标签 |
|---------|------|------|------|
| `winpower_power_watts` | Gauge | 瞬时功率 | `device_id`, `device_name`, `device_type` |
| `winpower_energy_total_wh` | Gauge | 累计电能 | `device_id`, `device_name`, `device_type` |

## 标签策略

### 设备标签
- `device_id`: 设备唯一标识符（必须）
- `device_name`: 设备名称（必须）
- `device_type`: 设备类型（必须）
- `phase`: 相线标识（可选，仅在多相设备使用）

### 状态标签
- `winpower_host`: WinPower主机地址（必须）
- `connection_type`: 连接类型（http/https）
- `auth_method`: 认证方法（token等）
- `status`: 状态枚举（ok/err, 1/0）
- `result`: 结果枚举（ok/err, success/failure）

### 高基数控制
- 严格限制标签枚举值，避免自由文本
- 设备标签采用固定组合，不随意扩展
- 状态类标签使用小集合枚举值

## 使用示例

### 基本使用

```go
package main

import (
    "go.uber.org/zap"
    "github.com/prometheus/client_golang/prometheus"

    "your-project/internal/metrics"
    "your-project/internal/config"
)

func main() {
    logger := zap.NewNop()

    // 创建指标管理器
    cfg := metrics.MetricManagerConfig{
        Namespace: "winpower",
        Subsystem: "exporter",
    }

    mm := metrics.NewMetricManager(cfg, logger)

    // 获取 HTTP Handler
    handler := mm.Handler()

    // 更新指标
    mm.SetUp(true) // 设置导出器状态为正常
    mm.SetDeviceConnected("ups-01", "Main UPS", "UPS", 1)
    mm.SetPowerWatts("ups-01", "Main UPS", "UPS", 1500.5)
}
```

### 与 WinPower 模块集成

```go
// WinPower 模块采集成功后调用
func (w *WinPowerClient) updateMetrics(data DeviceData) {
    // 更新设备连接状态
    connected := 0.0
    if data.Connected {
        connected = 1.0
    }
    w.metrics.SetDeviceConnected(data.ID, data.Name, data.Type, connected)

    // 更新电气参数
    w.metrics.SetElectricalData(data.ID, data.Name, data.Type, data.Phase,
        data.InputVoltage, data.OutputVoltage,
        data.InputCurrent, data.OutputCurrent,
        data.InputFrequency, data.OutputFrequency,
        data.InputWatts, data.OutputWatts,
        data.PowerFactor)

    // 记录采集耗时
    w.metrics.ObserveCollection("ok", data.CollectionDuration.Seconds())
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

## 配置说明

### 基本配置

```yaml
metrics:
  namespace: "winpower"      # 指标前缀
  subsystem: "exporter"      # 子系统
```

### 高级配置

```yaml
metrics:
  namespace: "winpower"
  subsystem: "exporter"
  request_duration_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2, 5]        # HTTP请求时延桶
  collection_duration_buckets: [0.1, 0.2, 0.5, 1, 2, 5, 10]      # 采集时延桶
  api_response_buckets: [0.05, 0.1, 0.2, 0.5, 1]                 # API响应时延桶
```

## 推荐 PromQL 查询

### 间隔能耗计算
```promql
# 5分钟间隔电能
increase(winpower_energy_total_wh{device_id="ups-01"}[5m])

# 每小时能耗（kWh）
increase(winpower_energy_total_wh[1h]) / 1000
```

### 设备状态监控
```promql
# 瞬时功率均值（5分钟）
avg_over_time(winpower_power_watts[5m])

# 设备连接状态
winpower_device_connected == 1

# 负载百分比
winpower_load_percent{device_id="ups-01"}
```

### 导出器健康监控
```promql
# 导出器运行状态
winpower_exporter_up == 1

# 采集错误率
rate(winpower_exporter_scrape_errors_total[5m]) / rate(winpower_exporter_requests_total[5m])

# 采集耗时
histogram_quantile(0.95, rate(winpower_exporter_collection_duration_seconds_bucket[5m]))
```

## 性能优化

### 内存管理
- 单一 Prometheus Registry，避免重复注册
- 预分配指标向量，避免动态扩容
- 控制指标数量，避免内存泄漏

### 并发控制
- 使用读写锁 (`sync.RWMutex`) 保护指标更新
- 读操作（HTTP暴露）使用读锁，允许并发访问
- 写操作（指标更新）使用写锁，确保数据一致性

### 实现优化
- 指标更新操作尽可能轻量
- 避免在关键路径进行复杂计算
- 批量更新减少锁竞争

## 错误处理

### 指标更新错误
- 记录错误日志，但不影响主流程
- 使用默认值或跳过更新，避免程序崩溃
- 提供错误统计指标，便于监控

### HTTP 错误
- 优雅处理网络异常
- 返回标准 HTTP 状态码
- 记录访问日志和错误统计

## 扩展性

### 指标扩展
- 支持通过配置添加自定义指标
- 预留指标扩展接口
- 支持指标动态注册（如需要）

### 标签扩展
- 支持可选标签配置
- 预留业务特定标签空间
- 支持标签映射和转换

## 故障排除

### 常见问题

1. **指标不显示**
   - 检查指标是否正确注册
   - 确认标签值是否符合规范
   - 验证数据是否正确更新

2. **高基数警告**
   - 检查标签枚举值是否过多
   - 减少不必要的标签
   - 使用标签聚合

3. **性能问题**
   - 检查指标更新频率
   - 优化锁粒度
   - 减少不必要的指标计算

### 调试工具

```bash
# 检查指标输出
curl http://localhost:9090/metrics

# 检查特定指标
curl http://localhost:9090/metrics | grep winpower_exporter_up
```

## 版本历史

- **v1.0.0**: 初始版本，实现基本指标管理功能
- 支持四类指标的完整实现
- 集成 HTTP 暴露和配置管理
- 完整的测试覆盖和文档

## 贡献指南

1. 遵循现有代码风格和架构设计
2. 添加新指标前请评估高基数风险
3. 确保新功能有完整的测试覆盖
4. 更新相关文档和示例

## 许可证

本模块遵循项目整体许可证。