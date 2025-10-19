# 指标暴露模块设计文档（重构版）

## 概述

指标暴露模块（Metrics）是导出器的输出层，负责将 WinPower 模块采集并经 Energy 模块累计后的数据，转化为 Prometheus 兼容的时序指标，并通过 HTTP `/metrics` 统一暴露。该重构版严格对齐最新的 WinPower、Energy、Storage、Logging、Config 设计，强调简洁、标准化、性能与可观测性。

## 设计原则

- 标准化：遵循 Prometheus 指标与标签命名规范，采用 Histogram 观测时延。
- 简洁性：暴露必要且稳定的指标集合，避免过度标签与高基数风险。
- 一致性：与 Collector 的采集语义、Energy 的累计语义保持一致；对认证与连接状态提供观测。
- 性能优先：单 `prometheus.Registry`，最小锁粒度；避免动态创建海量 time-series。
- 可观测性：提供 Exporter 自监控指标与推荐 PromQL 查询，便于运维与告警。

## 与上下游模块的契约

- WinPower
  - 始终通过 `CollectDeviceData(ctx)` 完成数据拉取与解析，并触发 Energy 累计与 Storage 持久化。
  - Metrics 模块提供类型安全的更新接口，WinPower 在成功解析与累计后调用，用于更新设备、电源与连接相关指标。

 

- Energy
  - 仅维护“累计电能（Wh）”且允许负值表示净能量；不计算“间隔电能”。
  - Metrics 只暴露 `energy_total_wh` 与 `power_watts`。间隔能耗通过 PromQL 在 Prometheus 侧计算。

- Storage
  - 提供设备级文件持久化，不被 Metrics 直接调用；WinPower/能源累计完成后负责写入。

- Logging（zap）
  - Metrics 更新与 HTTP 处理过程记录核心日志（级别可控），避免高频指标写入的冗长日志。

- Config
  - 支持 `metrics.namespace` 与 `metrics.subsystem` 前缀配置；允许注入自定义 `Registry`（默认新建）。

## 架构与组件

### 理解上下文：数据流与职责边界

- 数据路径：WinPower 解析设备数据 → Energy 维护累计电能 → Storage 持久化。
- 统一入口与暴露：Scheduler 的 `Tick(5s)` 和 HTTP `/metrics` 请求都调用 `WinPower.CollectDeviceData(ctx)` 完成采样与能耗累计 → MetricManager 更新指标 → 返回最新快照。
- 能耗语义：仅暴露瞬时功率 `power_watts` 与累计电能 `energy_total_wh`；任意时间窗口的间隔能耗通过 PromQL `increase()` 计算。
- 标签策略：设备标签使用 `device_id/device_name/device_type`，必要时使用 `phase`；状态类标签严格控枚举，避免高基数。
- 认证观测：WinPower 模块内部管理Token状态并更新相关指标；Metrics 不参与认证流程，仅记录观测结果。
- Server 集成：将 `MetricManager.Handler()` 挂载到 `/metrics`。该端点会调用采集方法并返回最新注册指标快照。

### 模块结构

```
┌──────────────────────────────────────────────────────────────┐
│                        Metrics Module                         │
├──────────────────────────────────────────────────────────────┤
│  ┌───────────────┐  ┌─────────────────────────────────────┐  │
│  │ MetricManager │◄►│ Prometheus Registry (single)       │  │
│  └───────────────┘  └─────────────────────────────────────┘  │
│          ▲                          ▲                         │
│          │                          │                         │
│   Collector / Energy (data)  Auth (status observe)           │
├──────────────────────────────────────────────────────────────┤
│                    HTTP /metrics Handler                      │
└──────────────────────────────────────────────────────────────┘
```

### 核心数据结构（示意）

```go
type MetricManager struct {
    registry *prometheus.Registry
    logger   *zap.Logger

    // Exporter 自监控
    requestDuration *prometheus.HistogramVec
    requestsTotal   *prometheus.CounterVec
    up              prometheus.Gauge
    scrapeErrors    *prometheus.CounterVec
    collectionTime  *prometheus.HistogramVec
    tokenRefresh    *prometheus.CounterVec
    deviceCount     *prometheus.GaugeVec

    // WinPower 连接/认证
    connectionStatus *prometheus.GaugeVec
    authStatus       *prometheus.GaugeVec
    apiResponseTime  *prometheus.HistogramVec
    tokenExpiry      *prometheus.GaugeVec
    tokenValid       *prometheus.GaugeVec

    // 设备/电源
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

    // 能耗（仅累计 + 瞬时功率）
    energyTotalWh    *prometheus.GaugeVec
    powerWatts       *prometheus.GaugeVec

    mu sync.RWMutex
}

// MetricManagerConfig 指标管理器配置（内部配置）
type MetricManagerConfig struct {
    Namespace string                `json:"namespace" yaml:"namespace"`
    Subsystem string                `json:"subsystem" yaml:"subsystem"`
    Registry  *prometheus.Registry  `json:"-" yaml:"-"`
}
```

### 关键接口

```go
func NewMetricManager(cfg MetricManagerConfig, logger *zap.Logger) *MetricManager
func (m *MetricManager) Handler() http.Handler // 返回 /metrics 的 Handler

// Exporter 观测
func (m *MetricManager) SetUp(status bool)
func (m *MetricManager) ObserveRequest(host, method, code string, sec float64)
func (m *MetricManager) ObserveCollection(status string, sec float64)
func (m *MetricManager) IncTokenRefresh(host, result string)
func (m *MetricManager) SetDeviceCount(host, deviceType string, n float64)

// WinPower 连接/认证
func (m *MetricManager) SetConnectionStatus(host, connectionType string, status float64)
func (m *MetricManager) ObserveAPI(host, endpoint string, sec float64)
func (m *MetricManager) SetTokenValid(host, userID string, valid float64)
func (m *MetricManager) SetTokenExpiry(host, userID string, seconds float64)

// 设备/电源
func (m *MetricManager) SetDeviceConnected(deviceID, deviceName, deviceType string, connected float64)
func (m *MetricManager) SetLoadPercent(deviceID, deviceName, deviceType, phase string, pct float64)
func (m *MetricManager) SetElectrical(deviceID, deviceName, deviceType, phase string,
    inV, outV, inA, outA, inHz, outHz, inW, outW, pfo float64)

// 能耗
func (m *MetricManager) SetPowerWatts(deviceID, deviceName, deviceType string, watts float64)
func (m *MetricManager) SetEnergyTotalWh(deviceID, deviceName, deviceType string, wh float64)
```

## 指标清单与命名

默认命名采用 `namespace_subsystem_metric`，建议：`namespace=winpower`、`subsystem=exporter`。示例：`winpower_exporter_requests_total`。

### Exporter 自监控

- `winpower_exporter_up` Gauge：Exporter 运行状态（1/0）。标签：`winpower_host`、`version`。
- `winpower_exporter_requests_total` Counter：HTTP 请求总数。标签：`winpower_host`、`status_code`、`method`。
- `winpower_exporter_request_duration_seconds` Histogram：请求时延。标签同上。
- `winpower_exporter_scrape_errors_total` Counter：采集错误总数。标签：`winpower_host`、`error_type`。
- `winpower_exporter_collection_duration_seconds` Histogram：采集+累计整体耗时。标签：`winpower_host`、`status`（ok/err）。
- `winpower_exporter_token_refresh_total` Counter：Token 刷新次数。标签：`winpower_host`、`result`（ok/err）。
- `winpower_exporter_device_count` Gauge：发现的设备数量。标签：`winpower_host`、`device_type`。

### WinPower 连接/认证

- `winpower_connection_status` Gauge：连接状态（1/0）。标签：`winpower_host`、`connection_type`。
- `winpower_auth_status` Gauge：认证状态（1/0）。标签：`winpower_host`、`auth_method`。
- `winpower_api_response_time_seconds` Histogram：API 响应时延。标签：`winpower_host`、`api_endpoint`。
- `winpower_token_expiry_seconds` Gauge：Token 剩余有效期。标签：`winpower_host`、`user_id`。
- `winpower_token_valid` Gauge：Token 有效性（1/0）。标签：`winpower_host`、`user_id`。

### 设备/电源（相线标签可选）

- `winpower_device_connected` Gauge：设备连接状态（1/0）。标签：`device_id`、`device_name`、`device_type`。
- `winpower_load_percent` Gauge：负载百分比。标签：`device_id`、`device_name`、`device_type`、`phase`。
- `winpower_input_voltage` / `winpower_output_voltage` Gauge：输入/输出电压。标签：同上。
- `winpower_input_current` / `winpower_output_current` Gauge：输入/输出电流。标签：同上。
- `winpower_input_frequency` / `winpower_output_frequency` Gauge：输入/输出频率。标签：同上。
- `winpower_input_watts` / `winpower_output_watts` Gauge：输入/输出有功功率。标签：同上。
- `winpower_output_power_factor` Gauge：输出功率因数。标签：同上。

### 能耗（与 Energy 契合）

- `winpower_power_watts` Gauge：瞬时功率（由 Collector 提供，来源协议字段 `loadTotalWatt`）。标签：`device_id`、`device_name`、`device_type`。
- `winpower_energy_total_wh` Gauge：累计电能（允许负值；由 Energy 维护）。标签：同上。

> 重要：不再暴露 `energy_interval_wh`。间隔能耗由 Prometheus 侧使用 PromQL 计算（见下文）。

## 标签策略与高基数控制

- 必选设备标签：`device_id`、`device_name`、`device_type`；相线 `phase` 仅在确有意义时使用。
- 控制标签枚举值：`status`、`result` 等仅使用小集合（`ok/err`、`1/0`）。
- 避免动态属性类标签（如自由文本型号、地区等）进入高频指标，必要时仅在低频指标或日志中呈现。

## HTTP 暴露

```go
func (m *MetricManager) Handler() http.Handler {
    return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
        EnableOpenMetrics: true,
        // 可选：错误处理、汇总禁用等
    })
}
```

Server 模块将此 Handler 绑定到 `/metrics` 路由即可。

重要：为保证统一入口与数据一致性，`/metrics` 处理过程中会触发 `Collector.CollectDeviceData(ctx)` 并等待完成；随后返回最新注册指标数据。

## 与 Collector / Auth 的集成示例

```go
// 采集成功后更新设备与能耗
mm.SetDeviceConnected(dev.ID, dev.Name, dev.Type, boolToFloat(dev.Connected))
mm.SetPowerWatts(dev.ID, dev.Name, dev.Type, parsed.Power.Watts)
mm.SetEnergyTotalWh(dev.ID, dev.Name, dev.Type, energy.TotalWh)

// 采集耗时观测
mm.ObserveCollection("ok", duration.Seconds())

// Token 刷新结果
mm.IncTokenRefresh(cfg.WinPower.URL, resultString(err))
```

## 推荐 PromQL（间隔能耗与常用观测）

- 间隔电能（5 分钟）：
  - `increase(winpower_energy_total_wh{device_id="ups-01"}[5m])`
- 每小时能耗（kWh）：
  - `increase(winpower_energy_total_wh[1h]) / 1000`
- 瞬时功率均值（5 分钟）：
  - `avg_over_time(winpower_power_watts[5m])`
- Exporter 健康：
  - `winpower_exporter_up == 1`
- 采集错误率：
  - `rate(winpower_exporter_scrape_errors_total[5m]) / rate(winpower_exporter_requests_total[5m])`

## 性能与采样

- Histogram 桶建议：请求/采集耗时设置 `[0.05, 0.1, 0.2, 0.5, 1, 2, 5]` 秒；API 响应 `[0.05, 0.1, 0.2, 0.5, 1]` 秒。
- 单 Registry，避免跨包多 Registry 带来的重复与冲突。
- 避免在热路径打印冗长日志，必要时以 Debug 级别并采样。

## 配置

```yaml
metrics:
  namespace: "winpower"     # 指标前缀（默认）
  subsystem: "exporter"     # 子系统（默认）
```

## 测试设计

- 单元测试：使用 `prometheus/testutil` 断言指标值与标签；聚焦 `SetEnergyTotalWh`、`ObserveCollection` 等关键路径。
- 端到端：启动 `/metrics`，用 `curl`/Prometheus 抓取并校验格式与内容。

## 迁移说明（相较旧版）

- 删除 `winpower_energy_interval_wh`，统一改为 PromQL 侧计算间隔能耗。
- 收敛标签集合，降低时序基数；设备标签以 `device_id/device_name/device_type` 为主。
- 接口简化且类型安全，避免散乱的字符串拼接更新。

## 总结

重构后的 Metrics 模块以“稳定的指标集合、清晰的接口、与 Energy 的严格契合”为核心，减少不必要复杂度并提升可运维性。结合推荐的 PromQL 与 Grafana 可快速搭建能耗与健康可视化，满足导出器的生产可观测需求。