# Metrics 模块 API 文档

本文档详细描述了 Metrics 模块提供的所有 API 接口和使用示例。

## 核心 API

### MetricManager

指标管理器的主要接口，提供完整的指标管理功能。

#### 构造函数

```go
func NewMetricManager(cfg MetricManagerConfig, logger *zap.Logger) *MetricManager
```

**参数:**
- `cfg`: 指标管理器配置
- `logger`: 结构化日志记录器

**返回:**
- `*MetricManager`: 指标管理器实例

**示例:**
```go
import (
    "go.uber.org/zap"
    "github.com/prometheus/client_golang/prometheus"
    "your-project/internal/metrics"
)

logger := zap.NewExample()
defer logger.Sync()

cfg := metrics.MetricManagerConfig{
    Namespace: "winpower",
    Subsystem: "exporter",
    Registry:  prometheus.NewRegistry(), // 可选，默认会创建新的
}

mm := metrics.NewMetricManager(cfg, logger)
```

#### HTTP Handler

```go
func (m *MetricManager) Handler() http.Handler
```

**返回:**
- `http.Handler`: 用于 `/metrics` 端点的 HTTP 处理器

**示例:**
```go
// 在 HTTP 服务器中注册
http.Handle("/metrics", mm.Handler())

// 或者使用 gin 框架
router.GET("/metrics", gin.WrapH(mm.Handler()))
```

## Exporter 自监控 API

### 运行状态

```go
func (m *MetricManager) SetUp(status bool)
```

设置导出器的运行状态。

**参数:**
- `status`: 运行状态（true=正常，false=异常）

**示例:**
```go
// 启动时设置为正常状态
mm.SetUp(true)

// 检测到异常时设置
if err != nil {
    mm.SetUp(false)
}
```

### HTTP 请求监控

```go
func (m *MetricManager) ObserveRequest(host, method, code string, duration time.Duration)
```

记录 HTTP 请求的指标。

**参数:**
- `host`: 请求主机名
- `method`: HTTP 方法（GET, POST 等）
- `code`: HTTP 状态码
- `duration`: 请求处理时间

**示例:**
```go
// 在 HTTP 中间件中使用
func metricsMiddleware(mm *metrics.MetricManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        duration := time.Since(start)
        mm.ObserveRequest(
            c.Request.Host,
            c.Request.Method,
            strconv.Itoa(c.Writer.Status()),
            duration,
        )
    }
}
```

### 采集错误统计

```go
func (m *MetricManager) IncScrapeError(host, errorType string)
```

增加采集错误计数。

**参数:**
- `host`: WinPower 主机地址
- `errorType`: 错误类型（network, auth, parse 等）

**示例:**
```go
// 在采集错误处理中
if err != nil {
    errorType := "unknown"
    switch {
    case isNetworkError(err):
        errorType = "network"
    case isAuthError(err):
        errorType = "auth"
    case isParseError(err):
        errorType = "parse"
    }
    mm.IncScrapeError(config.WinPower.URL, errorType)
}
```

### 采集耗时监控

```go
func (m *MetricManager) ObserveCollection(status string, duration time.Duration)
```

记录数据采集的耗时。

**参数:**
- `status`: 采集状态（ok=成功，err=失败）
- `duration`: 采集耗时

**示例:**
```go
// 在采集循环中
func (w *WinPowerClient) collectAndMeasure() {
    start := time.Now()

    data, err := w.CollectDeviceData(ctx)

    duration := time.Since(start)
    status := "ok"
    if err != nil {
        status = "err"
    }

    w.metrics.ObserveCollection(status, duration)
}
```

### Token 刷新统计

```go
func (m *MetricManager) IncTokenRefresh(host, result string)
```

记录 Token 刷新次数。

**参数:**
- `host`: WinPower 主机地址
- `result`: 刷新结果（ok=成功，err=失败）

**示例:**
```go
// 在 Token 刷新逻辑中
err := w.refreshToken()
result := "ok"
if err != nil {
    result = "err"
}
mm.IncTokenRefresh(w.config.URL, result)
```

### 设备数量统计

```go
func (m *MetricManager) SetDeviceCount(host, deviceType string, count float64)
```

设置发现的设备数量。

**参数:**
- `host`: WinPower 主机地址
- `deviceType`: 设备类型（UPS, PDU 等）
- `count`: 设备数量

**示例:**
```go
// 在设备发现后
devices := w.DiscoverDevices()
deviceCount := make(map[string]int)
for _, device := range devices {
    deviceCount[device.Type]++
}

for deviceType, count := range deviceCount {
    mm.SetDeviceCount(w.config.URL, deviceType, float64(count))
}
```

## WinPower 连接/认证 API

### 连接状态

```go
func (m *MetricManager) SetConnectionStatus(host, connectionType string, status float64)
```

设置与 WinPower 的连接状态。

**参数:**
- `host`: WinPower 主机地址
- `connectionType`: 连接类型（http, https）
- `status`: 连接状态（1=已连接，0=未连接）

**示例:**
```go
// 在连接状态变化时
connected := 0.0
if w.IsConnected() {
    connected = 1.0
}

connectionType := "http"
if strings.HasPrefix(w.config.URL, "https") {
    connectionType = "https"
}

mm.SetConnectionStatus(w.config.URL, connectionType, connected)
```

### 认证状态

```go
func (m *MetricManager) SetAuthStatus(host, authMethod string, status float64)
```

设置认证状态。

**参数:**
- `host`: WinPower 主机地址
- `authMethod`: 认证方法（token, basic 等）
- `status`: 认证状态（1=已认证，0=未认证）

**示例:**
```go
// 在认证状态变化时
authStatus := 0.0
if w.IsAuthenticated() {
    authStatus = 1.0
}

mm.SetAuthStatus(w.config.URL, "token", authStatus)
```

### API 响应时间

```go
func (m *MetricManager) ObserveAPI(host, endpoint string, duration time.Duration)
```

记录 API 调用的响应时间。

**参数:**
- `host`: WinPower 主机地址
- `endpoint`: API 端点（/api/login, /api/devices 等）
- `duration`: 响应时间

**示例:**
```go
// 在 API 调用中
func (w *WinPowerClient) callAPI(endpoint string) (*Response, error) {
    start := time.Now()

    resp, err := w.httpClient.Get(w.config.URL + endpoint)

    duration := time.Since(start)
    w.metrics.ObserveAPI(w.config.URL, endpoint, duration)

    return resp, err
}
```

### Token 有效期

```go
func (m *MetricManager) SetTokenExpiry(host, userID string, seconds float64)
```

设置 Token 的剩余有效期。

**参数:**
- `host`: WinPower 主机地址
- `userID`: 用户 ID
- `seconds`: 剩余有效时间（秒）

**示例:**
```go
// 在 Token 管理中
if w.token != nil {
    expiry := time.Until(w.token.ExpiresAt).Seconds()
    if expiry > 0 {
        mm.SetTokenExpiry(w.config.URL, w.currentUser, expiry)
    }
}
```

### Token 有效性

```go
func (m *MetricManager) SetTokenValid(host, userID string, valid float64)
```

设置 Token 的有效性。

**参数:**
- `host`: WinPower 主机地址
- `userID`: 用户 ID
- `valid`: Token 有效性（1=有效，0=无效）

**示例:**
```go
// 在 Token 验证后
valid := 0.0
if w.token != nil && !w.token.IsExpired() {
    valid = 1.0
}
mm.SetTokenValid(w.config.URL, w.currentUser, valid)
```

## 设备/电源数据 API

### 设备连接状态

```go
func (m *MetricManager) SetDeviceConnected(deviceID, deviceName, deviceType string, connected float64)
```

设置设备的连接状态。

**参数:**
- `deviceID`: 设备唯一标识符
- `deviceName`: 设备名称
- `deviceType`: 设备类型
- `connected`: 连接状态（1=已连接，0=未连接）

**示例:**
```go
// 在设备状态更新时
for _, device := range devices {
    connected := 0.0
    if device.IsOnline() {
        connected = 1.0
    }

    mm.SetDeviceConnected(device.ID, device.Name, device.Type, connected)
}
```

### 负载百分比

```go
func (m *MetricManager) SetLoadPercent(deviceID, deviceName, deviceType, phase string, pct float64)
```

设置设备的负载百分比。

**参数:**
- `deviceID`: 设备唯一标识符
- `deviceName`: 设备名称
- `deviceType`: 设备类型
- `phase`: 相线标识（可选，单相设备可传空字符串）
- `pct`: 负载百分比（0-100）

**示例:**
```go
// 更新负载百分比
mm.SetLoadPercent("ups-01", "Main UPS", "UPS", "", 75.5)

// 对于三相设备
mm.SetLoadPercent("pdu-01", "Rack PDU", "PDU", "L1", 45.2)
mm.SetLoadPercent("pdu-01", "Rack PDU", "PDU", "L2", 52.8)
mm.SetLoadPercent("pdu-01", "Rack PDU", "PDU", "L3", 48.1)
```

### 电气参数

```go
func (m *MetricManager) SetElectricalData(deviceID, deviceName, deviceType, phase string,
    inV, outV, inA, outA, inHz, outHz, inW, outW, pfo float64)
```

批量设置设备的电气参数。

**参数:**
- `deviceID`: 设备唯一标识符
- `deviceName`: 设备名称
- `deviceType`: 设备类型
- `phase`: 相线标识（可选）
- `inV`: 输入电压（伏特）
- `outV`: 输出电压（伏特）
- `inA`: 输入电流（安培）
- `outA`: 输出电流（安培）
- `inHz`: 输入频率（赫兹）
- `outHz`: 输出频率（赫兹）
- `inW`: 输入有功功率（瓦特）
- `outW`: 输出有功功率（瓦特）
- `pfo`: 输出功率因数

**示例:**
```go
// 批量更新电气参数
mm.SetElectricalData(
    "ups-01", "Main UPS", "UPS", "",
    230.5,  // 输入电压
    230.1,  // 输出电压
    8.2,    // 输入电流
    7.9,    // 输出电流
    50.0,   // 输入频率
    50.0,   // 输出频率
    1891.0, // 输入有功功率
    1817.0, // 输出有功功率
    0.95,   // 输出功率因数
)
```

## 能耗数据 API

### 瞬时功率

```go
func (m *MetricManager) SetPowerWatts(deviceID, deviceName, deviceType string, watts float64)
```

设置设备的瞬时功率。

**参数:**
- `deviceID`: 设备唯一标识符
- `deviceName`: 设备名称
- `deviceType`: 设备类型
- `watts`: 瞬时功率（瓦特）

**示例:**
```go
// 在能耗计算后
mm.SetPowerWatts("ups-01", "Main UPS", "UPS", 1817.5)
```

### 累计电能

```go
func (m *MetricManager) SetEnergyTotalWh(deviceID, deviceName, deviceType string, wh float64)
```

设置设备的累计电能。

**参数:**
- `deviceID`: 设备唯一标识符
- `deviceName`: 设备名称
- `deviceType`: 设备类型
- `wh`: 累计电能（瓦时，允许负值）

**示例:**
```go
// 与 Energy 模块集成
func (e *EnergyService) updateMetrics() {
    for deviceID, data := range e.energyData {
        mm.SetEnergyTotalWh(
            deviceID,
            data.DeviceName,
            data.DeviceType,
            data.TotalWh, // 累计电能，可能为负值
        )
    }
}
```

## 完整使用示例

### 基本集成示例

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "go.uber.org/zap"

    "your-project/internal/metrics"
    "your-project/internal/config"
    "your-project/internal/winpower"
    "your-project/internal/energy"
)

func main() {
    // 初始化日志
    logger := zap.NewProduction()
    defer logger.Sync()

    // 加载配置
    cfg := config.Load()

    // 创建指标管理器
    metricsCfg := metrics.MetricManagerConfig{
        Namespace: "winpower",
        Subsystem: "exporter",
    }
    mm := metrics.NewMetricManager(metricsCfg, logger)

    // 设置导出器状态
    mm.SetUp(true)

    // 创建 WinPower 客户端
    wpClient := winpower.NewClient(cfg.WinPower, logger)

    // 创建 Energy 服务
    energyService := energy.NewService(cfg.Energy, logger)

    // 启动 HTTP 服务器
    router := gin.New()

    // 添加指标中间件
    router.Use(func(c *gin.Context) {
        start := time.Now()
        c.Next()

        duration := time.Since(start)
        mm.ObserveRequest(
            c.Request.Host,
            c.Request.Method,
            string(rune(c.Writer.Status())),
            duration,
        )
    })

    // 注册指标端点
    router.GET("/metrics", gin.WrapH(mm.Handler()))

    // 采集循环
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            collectData(wpClient, energyService, mm, logger)
        }
    }()

    // 启动服务器
    logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
    if err := router.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
        logger.Fatal("Failed to start server", zap.Error(err))
    }
}

func collectData(wpClient *winpower.Client, energyService *energy.Service,
    mm *metrics.MetricManager, logger *zap.Logger) {

    start := time.Now()

    // 采集设备数据
    devices, err := wpClient.CollectDeviceData(context.Background())
    duration := time.Since(start)

    status := "ok"
    if err != nil {
        status = "err"
        mm.IncScrapeError(wpClient.GetURL(), getErrorType(err))
        logger.Error("Failed to collect device data", zap.Error(err))
    }

    mm.ObserveCollection(status, duration.Seconds())

    // 更新设备指标
    for _, device := range devices {
        // 设备连接状态
        connected := 0.0
        if device.IsConnected {
            connected = 1.0
        }
        mm.SetDeviceConnected(device.ID, device.Name, device.Type, connected)

        // 负载百分比
        mm.SetLoadPercent(device.ID, device.Name, device.Type, "", device.LoadPercent)

        // 电气参数
        mm.SetElectricalData(
            device.ID, device.Name, device.Type, "",
            device.InputVoltage, device.OutputVoltage,
            device.InputCurrent, device.OutputCurrent,
            device.InputFrequency, device.OutputFrequency,
            device.InputWatts, device.OutputWatts,
            device.PowerFactor,
        )

        // 能耗数据
        if energy, err := energyService.GetEnergy(device.ID); err == nil {
            mm.SetPowerWatts(device.ID, device.Name, device.Type, energy.PowerWatts)
            mm.SetEnergyTotalWh(device.ID, device.Name, device.Type, energy.TotalWh)
        }
    }

    // 更新连接和认证状态
    updateConnectionMetrics(wpClient, mm)
}

func updateConnectionMetrics(wpClient *winpower.Client, mm *metrics.MetricManager) {
    host := wpClient.GetURL()
    connectionType := "http"
    if strings.HasPrefix(host, "https") {
        connectionType = "https"
    }

    // 连接状态
    connected := 0.0
    if wpClient.IsConnected() {
        connected = 1.0
    }
    mm.SetConnectionStatus(host, connectionType, connected)

    // 认证状态
    authStatus := 0.0
    if wpClient.IsAuthenticated() {
        authStatus = 1.0
    }
    mm.SetAuthStatus(host, "token", authStatus)

    // Token 状态
    if token := wpClient.GetToken(); token != nil {
        expiry := time.Until(token.ExpiresAt).Seconds()
        if expiry > 0 {
            mm.SetTokenExpiry(host, token.UserID, expiry)
            mm.SetTokenValid(host, token.UserID, 1.0)
        } else {
            mm.SetTokenValid(host, token.UserID, 0.0)
        }
    }
}

func getErrorType(err error) string {
    if isNetworkError(err) {
        return "network"
    }
    if isAuthError(err) {
        return "auth"
    }
    if isParseError(err) {
        return "parse"
    }
    return "unknown"
}
```

## 错误处理最佳实践

### 指标更新错误处理

```go
func safeUpdateMetrics(mm *metrics.MetricManager, device Device) {
    defer func() {
        if r := recover(); r != nil {
            logger.Error("Panic in metrics update", zap.Any("panic", r))
        }
    }()

    // 验证设备数据
    if device.ID == "" {
        logger.Warn("Device ID is empty, skipping metrics update")
        return
    }

    // 更新指标，忽略错误以避免影响主流程
    mm.SetDeviceConnected(device.ID, device.Name, device.Type,
        boolToFloat(device.IsConnected))
}

func boolToFloat(b bool) float64 {
    if b {
        return 1.0
    }
    return 0.0
}
```

### 配置验证

```go
func validateMetricsConfig(cfg metrics.MetricManagerConfig) error {
    if cfg.Namespace == "" {
        return fmt.Errorf("metrics namespace cannot be empty")
    }
    if cfg.Subsystem == "" {
        return fmt.Errorf("metrics subsystem cannot be empty")
    }
    return nil
}
```

## 性能优化建议

### 批量更新

```go
func (m *MetricManager) BatchUpdateDevices(devices []Device) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, device := range devices {
        // 批量更新减少锁竞争
        m.updateDeviceMetricsUnsafe(device)
    }
}
```

### 异步更新

```go
func (m *MetricManager) AsyncUpdate(device Device) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                m.logger.Error("Async metrics update panic", zap.Any("panic", r))
            }
        }()

        m.SetDeviceConnected(device.ID, device.Name, device.Type,
            boolToFloat(device.IsConnected))
    }()
}
```

这份 API 文档提供了完整的接口说明和使用示例，帮助开发者正确使用 Metrics 模块的所有功能。