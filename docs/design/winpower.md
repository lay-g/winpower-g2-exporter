# WinPower模块设计文档

## 概述

WinPower模块负责与WinPower G2系统的完整交互，包括身份认证、Token管理、数据采集、解析与错误处理。该模块通过 HTTP/HTTPS 与 WinPower API 通信，是整个导出器的核心数据源。

### 设计目标

- **可靠性**：稳定获取数据，具备完善的错误处理与日志
- **性能**：复用单个 `http.Client` 降低开销，合理设置超时
- **容错性**：对网络错误、认证错误、格式错误进行分层处理
- **可扩展性**：支持不同设备类型与响应格式
- **简洁性**：统一管理认证和采集逻辑，简化系统架构

## 架构设计

### 模块结构图

```
┌───────────────────────────────────────────────────────────────┐
│                    WinPower Module                             │
├───────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │   HTTP       │  │   Token     │  │    Data Parser      │   │
│  │   Client     │  │   Manager   │  │                     │   │
│  │              │  │             │  │ ┌─────────────────┐ │   │
│  │ ┌──────────┐ │  │ ┌─────────┐ │  │ │ Device Data     │ │   │
│  │ │ Error    │ │  │ │ Token   │ │  │ │ Parser          │ │   │
│  │ │ Handler  │ │  │ │ Cache   │ │  │ └─────────────────┘ │   │
│  │ └──────────┘ │  │ └─────────┘ │  │ ┌─────────────────┐ │   │
│  │              │  │             │  │ │ Metrics Parser  │ │   │
│  │              │  │             │  │ └─────────────────┘ │   │
│  └──────────────┘  └─────────────┘  │ ┌─────────────────┐ │   │
│                                     │ │ Validation      │ │   │
│                                     │ └─────────────────┘ │   │
├───────────────────────────────────────────────────────────────┤
│                        Interfaces                             │
│  ┌────────────────────────────────────────────────────────┐   │
│  │            WinPowerClient Interface                    │   │
│  │  - CollectDeviceData(ctx) // 始终包含电能计算          │   │
│  │  - GetConnectionStatus()                               │   │
│  │  - GetLastCollectionTime()                             │   │
│  └────────────────────────────────────────────────────────┘   │
└───────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────┐
│                   WinPower System                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │ Auth API    │  │ Device API  │  │   SSL/TLS Layer     │   │
│  │ /api/v1/    │  │ /api/v1/    │  │                     │   │
│  │ auth/login  │  │ deviceData/ │  │ ┌─────────────────┐ │   │
│  │             │  │ detail/list │  │ │ Self-signed     │ │   │
│  └─────────────┘  └─────────────┘  │ │ Certificates    │ │   │
│                                    │ └─────────────────┘ │   │
│                                    └─────────────────────┘   │
└──────────────────────────────────────────────────────────────┘
```

### 数据流程图

```
采集请求
    │
    ▼
┌─────────────────┐     Token验证      ┌──────────────────┐
│ Start Collection│ ──────────────────►│ Check Token      │
└─────────────────┘                   └──────────────────┘
    │                                    │
    │                                    ▼ Token过期?
    │                                   是 │
    │                                    ▼
    │                           ┌──────────────────┐
    │                           │ Refresh Token    │
    │                           └──────────────────┘
    │                                    │
    ▼                                   否
┌─────────────────┐                     │
│ HTTP Request    │ ◄───────────────────┘
│ to WinPower     │
└─────────────────┘
    │
    ▼ 响应成功?
┌─────────────────┐ 是 ┌──────────────────┐
│ Error Handler   │───►│ Parse Response   │
└─────────────────┘    └──────────────────┘
    │ 否                     │
    ▼                        ▼
┌─────────────────┐   ┌──────────────────┐
│ Return Error    │   │ Validate Data    │
└─────────────────┘   └──────────────────┘
                         │
                         ▼
                    ┌──────────────────┐
                    │ Return Device    │
                    │ Data (Raw Only)  │
                    └──────────────────┘
```

## 核心组件设计

### 1. WinPower客户端 (WinPowerClient)

#### 职责
- 统一管理认证和数据采集功能
- 提供 `CollectDeviceData` 接口供调度器调用
- **仅负责数据采集，不触发电能计算**
- 内部管理Token生命周期和HTTP通信

#### 核心接口

```go
// WinPowerClient WinPower客户端接口
type WinPowerClient interface {
    // CollectDeviceData 采集设备数据（仅原始数据，不触发电能计算）
    CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error)

    // GetConnectionStatus 获取连接状态
    GetConnectionStatus() bool

    // GetLastCollectionTime 获取最后采集时间
    GetLastCollectionTime() time.Time
}
```

### 2. HTTP客户端 (HTTPClient)

#### 职责
- 管理与 WinPower 的 HTTP 请求与响应
- 处理 SSL/TLS 配置与证书验证（可跳过）
- 设置统一的请求超时与 User-Agent
- 复用单个 `http.Client` 实例

#### 数据结构

```go
// HTTPClient 封装与 WinPower 的 HTTP 通信
type HTTPClient struct {
    client    *http.Client      // Go 标准库 HTTP 客户端（单实例复用）
    baseURL   string            // WinPower 基础 URL
    timeout   time.Duration     // 请求超时时间
    skipSSL   bool              // 是否跳过 SSL 验证
    userAgent string            // User-Agent 字符串
    logger    *zap.Logger       // 日志记录器
}
```

#### 核心方法

```go
// NewHTTPClient 创建新的 HTTP 客户端
func NewHTTPClient(cfg *Config, logger *zap.Logger) *HTTPClient

// Do 执行 HTTP 请求
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error)

// Get 执行 GET 请求
func (c *HTTPClient) Get(url string, headers map[string]string) (*http.Response, error)

// Post 执行 POST 请求
func (c *HTTPClient) Post(url string, body interface{}, headers map[string]string) (*http.Response, error)
```

### 3. Token管理器 (TokenManager)

#### 职责
- 管理 JWT Token 的获取、缓存和刷新
- 处理登录认证流程
- 提供Token有效性检查

#### 数据结构

```go
// TokenInfo Token 信息
type TokenInfo struct {
    Token       string    // Token 字符串
    DeviceID    string    // 设备 ID
    CreatedAt   time.Time // Token 创建时间
    ExpiresAt   time.Time // Token 过期时间
    ExpiresHours int      // Token 有效期（小时）
}

// TokenManager Token管理器
type TokenManager struct {
    client           *HTTPClient
    username         string
    password         string
    tokenCache       *TokenInfo
    refreshThreshold time.Duration
    logger           *zap.Logger
    mutex            sync.RWMutex
}
```

#### 核心方法

```go
// GetToken 获取有效Token（自动刷新）
func (tm *TokenManager) GetToken(ctx context.Context) (string, error)

// RefreshToken 刷新Token
func (tm *TokenManager) RefreshToken(ctx context.Context) error

// IsTokenValid 检查Token是否有效
func (tm *TokenManager) IsTokenValid() bool

// Login 登录获取Token
func (tm *TokenManager) Login(ctx context.Context) error
```

### 4. 数据解析器 (DataParser)

#### 职责
- 解析设备数据响应
- 验证数据格式与完整性
- 产出标准化的设备数据结构
- 将WinPower API响应转换为内部数据结构

#### 数据转换流程

数据解析器负责将WinPower API的响应数据转换为 `ParsedDeviceData` 结构体。转换流程如下：

```
WinPower API响应
    │
    ▼ JSON反序列化
┌───────────────────────────────────────────────────────────────┐
│                   APIResponse结构体                           │
│  {                                                            │
│      "total": 1,                                             │
│      "pageSize": 20,                                         │
│      "currentPage": 1,                                       │
│      "data": [                                               │
│          {                                                   │
│              "assetDevice": { /* 设备基本信息 */ },           │
│              "realtime": { /* 实时数据 */ },                 │
│              "config": { /* 配置参数 */ },                   │
│              "setting": { /* 设置信息 */ },                  │
│              "activeAlarms": [ /* 活动告警 */ ],             │
│              "controlSupported": { /* 控制功能 */ },          │
│              "connected": true                               │
│          }                                                   │
│      ],                                                       │
│      "code": "000000",                                        │
│      "msg": "OK"                                              │
│  }                                                            │
└───────────────────────────────────────────────────────────────┘
    │
    ▼ 数据提取与类型转换
┌───────────────────────────────────────────────────────────────┐
│                 ParsedDeviceData结构体                        │
│  {                                                            │
│      DeviceID:     string,       // 来自assetDevice.id       │
│      DeviceName:   string,       // 来自assetDevice.alias     │
│      DeviceType:   int,          // 来自assetDevice.deviceType│
│      DeviceModel:  string,       // 来自assetDevice.model     │
│      Connected:    bool,         // 来自connected字段         │
│      RealtimeData: RealtimeData, // 来自realtime字段转换      │
│      Timestamp:    time.Time     // 当前采集时间              │
│  }                                                            │
└───────────────────────────────────────────────────────────────┘
```

#### 详细转换映射

**1. 基本信息映射 (assetDevice → ParsedDeviceData)**

| API字段              | ParsedDeviceData字段 | 类型转换 | 说明 |
| -------------------- | -------------------- | -------- | ---- |
| assetDevice.id       | DeviceID             | string   | 设备唯一标识符 |
| assetDevice.alias    | DeviceName           | string   | 设备别名 |
| assetDevice.deviceType | DeviceType        | int      | 设备类型 |
| assetDevice.model    | DeviceModel          | string   | 设备型号 |
| connected            | Connected            | bool     | 连接状态 |

**2. 实时数据转换 (realtime → RealtimeData)**

实时数据的转换涉及字符串到数值类型的转换，以及字段名的重新映射：

```go
// 转换示例
func parseRealtimeData(realtime map[string]interface{}) RealtimeData {
    data := RealtimeData{}

    // 字符串转float64
    if volt, ok := realtime["inputVolt1"].(string); ok {
        if val, err := strconv.ParseFloat(volt, 64); err == nil {
            data.InputVolt1 = val
        }
    }

    // 字符串转int
    if mode, ok := realtime["mode"].(string); ok {
        if val, err := strconv.Atoi(mode); err == nil {
            data.Mode = val
        }
    }

    // 字符串转bool (isCharging)
    if charging, ok := realtime["isCharging"].(string); ok {
        data.IsCharging = charging == "1"
    }

    // 能耗关键字段：总负载有功功率
    if watt, ok := realtime["loadTotalWatt"].(string); ok {
        if val, err := strconv.ParseFloat(watt, 64); err == nil {
            data.LoadTotalWatt = val  // 用于能耗计算的核心字段
        }
    }

    return data
}
```

**3. 实时数据字段映射表**

| API字段 | ParsedDeviceData字段 | 目标类型 | 能耗计算用途 |
| ------- | ------------------- | -------- | ------------ |
| inputVolt1 | InputVolt1 | float64 | - |
| outputVoltageType | OutputVoltageType | int | - |
| loadPercent | LoadPercent | float64 | - |
| isCharging | IsCharging | bool | - |
| batVoltP | BatVoltP | float64 | - |
| outputCurrent1 | OutputCurrent1 | float64 | - |
| upsTemperature | UpsTemperature | float64 | - |
| mode | Mode | int | - |
| batRemainTime | BatRemainTime | int | - |
| loadVa1 | LoadVa1 | float64 | - |
| faultCode | FaultCode | string | - |
| outputVolt1 | OutputVolt1 | float64 | - |
| outputFreq | OutputFreq | float64 | - |
| inputFreq | InputFreq | float64 | - |
| loadTotalVa | LoadTotalVa | float64 | - |
| batteryStatus | BatteryStatus | int | - |
| testStatus | TestStatus | int | - |
| **loadTotalWatt** | **LoadTotalWatt** | **float64** | **✅ 能耗计算输入** |
| loadWatt1 | LoadWatt1 | float64 | - |
| batCapacity | BatCapacity | float64 | - |
| status | Status | int | - |

#### 转换实现示例

```go
// API响应结构体定义
type DeviceDataResponse struct {
    Total       int                    `json:"total"`
    PageSize    int                    `json:"pageSize"`
    CurrentPage int                    `json:"currentPage"`
    Data        []RawDeviceInfo        `json:"data"`
    Code        string                 `json:"code"`
    Msg         string                 `json:"msg"`
}

type RawDeviceInfo struct {
    AssetDevice     RawAssetDevice     `json:"assetDevice"`
    Realtime        map[string]interface{} `json:"realtime"`
    Config          map[string]interface{} `json:"config"`
    Setting         map[string]interface{} `json:"setting"`
    ActiveAlarms    []interface{}      `json:"activeAlarms"`
    ControlSupported map[string]interface{} `json:"controlSupported"`
    Connected       bool               `json:"connected"`
}

type RawAssetDevice struct {
    ID              string `json:"id"`
    DeviceType      int    `json:"deviceType"`
    Model           string `json:"model"`
    Alias           string `json:"alias"`
    // ... 其他字段
}

// 数据转换函数
func (p *DataParser) ParseDeviceData(response *DeviceDataResponse) ([]ParsedDeviceData, error) {
    var devices []ParsedDeviceData

    for _, rawDevice := range response.Data {
        device := ParsedDeviceData{
            DeviceID:     rawDevice.AssetDevice.ID,
            DeviceName:   rawDevice.AssetDevice.Alias,
            DeviceType:   rawDevice.AssetDevice.DeviceType,
            DeviceModel:  rawDevice.AssetDevice.Model,
            Connected:    rawDevice.Connected,
            RealtimeData: p.parseRealtimeData(rawDevice.Realtime),
            Timestamp:    time.Now(),
        }

        devices = append(devices, device)
    }

    return devices, nil
}
```

#### 数据验证与错误处理

**1. 必填字段验证**
- DeviceID不能为空
- 能耗计算字段LoadTotalWatt必须有效
- 设备连接状态必须明确

**2. 数值范围验证**
- 电压值在合理范围内 (0-500V)
- 功率值为非负数
- 百分比字段在0-100范围内

**3. 转换错误处理**
- 字符串转数值失败时记录警告并使用默认值
- 缺失字段时记录调试信息
- 批量转换错误时提供详细的错误定位

#### 关键结构

```go
// ParsedDeviceData 解析后的设备数据
type ParsedDeviceData struct {
    DeviceID     string       `json:"device_id"`
    DeviceName   string       `json:"device_name"`
    DeviceType   int          `json:"device_type"`
    DeviceModel  string       `json:"device_model"`
    Connected    bool         `json:"connected"`
    RealtimeData RealtimeData `json:"realtime_data"`
    Timestamp    time.Time    `json:"timestamp"`
}

// RealtimeData 实时数据结构体（基于WinPower G2 API协议）
type RealtimeData struct {
    InputVolt1              float64 `json:"input_volt_1"`               // 输入电压 (相1) (V)
    OutputVoltageType       int     `json:"output_voltage_type"`        // 输出电压类型
    LoadPercent             float64 `json:"load_percent"`                // 负载百分比 (%)
    IsCharging              bool    `json:"is_charging"`                 // 是否正在充电 (1=是, 0=否)
    InputTransformerType    int     `json:"input_transformer_type"`      // 输入变压器类型
    BatVoltP                float64 `json:"bat_volt_p"`                  // 电池电压百分比 (V)
    OutputCurrent1          float64 `json:"output_current_1"`            // 输出电流 (相1) (A)
    UpsTemperature          float64 `json:"ups_temperature"`             // UPS 温度 (°C)
    Mode                    int     `json:"mode"`                        // UPS 工作模式
    BatRemainTime           int     `json:"bat_remain_time"`             // 电池剩余时间 (秒)
    LoadVa1                 float64 `json:"load_va_1"`                   // 负载视在功率 (相1) (VA)
    FaultCode               string  `json:"fault_code"`                  // 故障代码
    OutputVolt1             float64 `json:"output_volt_1"`               // 输出电压 (相1) (V)
    OutputFreq              float64 `json:"output_freq"`                 // 输出频率 (Hz)
    InputFreq               float64 `json:"input_freq"`                  // 输入频率 (Hz)
    LoadTotalVa             float64 `json:"load_total_va"`               // 总负载视在功率 (VA)
    BatteryStatus           int     `json:"battery_status"`              // 电池状态
    TestStatus              int     `json:"test_status"`                 // 测试状态
    LoadTotalWatt           float64 `json:"load_total_watt"`             // 总负载有功功率 (W) - 能耗计算使用此字段
    LoadWatt1               float64 `json:"load_watt_1"`                 // 负载有功功率 (相1) (W)
    BatCapacity             float64 `json:"bat_capacity"`                // 电池容量 (%)
    Status                  int     `json:"status"`                      // 设备状态
}
```

## 配置设计

### WinPower配置结构

WinPower模块定义自己的配置结构体，由config模块主动引用：

```go
// internal/winpower/config.go
package winpower

import "time"

type Config struct {
    BaseURL         string        `yaml:"base_url" validate:"required,url"`
    Username        string        `yaml:"username" validate:"required"`
    Password        string        `yaml:"password" validate:"required"`
    Timeout         time.Duration `yaml:"timeout" validate:"min=1s"`
    APITimeout      time.Duration `yaml:"api_timeout" validate:"min=1s"`
    SkipSSLVerify   bool          `yaml:"skip_ssl_verify"`
    RefreshThreshold time.Duration `yaml:"refresh_threshold" validate:"min=1m"`
}

func DefaultConfig() *Config {
    return &Config{
        Timeout:          15 * time.Second,
        APITimeout:       10 * time.Second,
        SkipSSLVerify:    false,
        RefreshThreshold: 5 * time.Minute,
    }
}

func (c *Config) Validate() error {
    // 配置验证逻辑
    return nil
}
```

### 配置示例

```yaml
winpower:
  base_url: "https://winpower.example.com"
  username: "admin"
  password: "secret"
  timeout: 15s
  api_timeout: 10s
  skip_ssl_verify: false
  refresh_threshold: 5m
```

## 错误处理与日志

### 错误处理策略

- **网络错误**：采集端返回明确错误，由上层调度控制处理策略
- **认证错误**：优先尝试刷新Token；失败则返回错误等待下次调度
- **解析错误**：记录原始响应片段与字段定位，便于定位问题
- **超时错误**：记录超时信息，由上层调度控制处理策略

### 日志记录

- 统一使用结构化日志记录关键指标（耗时、响应码、设备数）
- **敏感信息保护**：不得在日志中记录密码、Token值或Authorization头
- 仅记录非敏感字段：device_id、响应码、耗时、状态等
- 错误信息需脱敏处理

## 与其他模块的集成

### 与Energy模块的职责分离

**重要**：WinPower模块不直接调用Energy模块，电能计算由Collector模块负责。

```go
// WinPower模块仅负责数据采集，不触发电能计算
func (w *WinPowerClient) CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error) {
    // 1. 获取有效Token
    token, err := w.tokenManager.GetToken(ctx)
    if err != nil {
        return nil, err
    }

    // 2. 采集设备数据
    data, err := w.collectData(ctx, token)
    if err != nil {
        return nil, err
    }

    // 3. 返回解析后的设备数据，不进行电能计算
    // 电能计算由Collector模块在获取数据后独立触发
    w.logger.Debug("Device data collected successfully",
        zap.Int("device_count", len(data)))

    return data, nil
}
```

### 与Collector模块的协作

- WinPower模块提供原始设备数据（包括功率信息）
- Collector模块调用WinPower获取数据，然后独立触发电能计算
- 明确的职责边界：WinPower负责采集，Collector负责触发计算

### 与Scheduler模块集成

- 调度器每5秒调用 `winpower.CollectDeviceData(ctx)`
- **职责明确**：WinPower仅负责数据采集，不进行电能计算与持久化
- 完整的数据流：Scheduler → WinPower(采集) → Collector(触发计算) → Energy(计算) → Storage(持久化)

### 与Metrics模块集成

- 在采集和认证过程中更新相关指标
- 暴露连接状态、采集次数、错误计数等指标

## 测试设计

### 单元测试

```go
// MockWinPowerClient Mock客户端
type MockWinPowerClient struct {
    connectionStatus bool
    lastCollection   time.Time
    collectError     error
    mockData         []ParsedDeviceData
}

func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error) {
    if m.collectError != nil {
        return nil, m.collectError
    }
    return m.mockData, nil
}

func (m *MockWinPowerClient) GetConnectionStatus() bool {
    return m.connectionStatus
}

func (m *MockWinPowerClient) GetLastCollectionTime() time.Time {
    return m.lastCollection
}
```

### 集成测试

- 使用模拟HTTP服务器测试认证流程
- 使用模拟设备数据测试采集和解析流程
- 测试Token自动刷新机制

## 性能优化

### 连接复用

- 复用单个 `http.Client` 实例
- 合理配置连接池参数
- 避免频繁创建和销毁连接

### 缓存策略

- Token缓存避免频繁登录
- 使用读写锁保证并发安全
- 设置合理的刷新阈值

### 并发控制

- Token管理使用读写锁
- 避免并发采集冲突
- 合理设置超时时间

## 安全考虑

### 认证安全

- 密码和Token安全存储
- 使用HTTPS传输认证信息
- Token按到期阈值提前刷新

### 信息安全

- 日志中不记录敏感信息
- 错误信息脱敏处理
- 避免在URL中传递敏感参数

### 网络安全

- 支持SSL/TLS证书验证
- 可配置跳过自签名证书验证
- 生产环境建议使用反向代理终止TLS

## 监控指标

### 认证指标

- Token刷新次数
- 认证成功/失败次数
- Token剩余有效时间

### 采集指标

- 采集次数
- 采集成功率
- 平均采集耗时
- 设备连接状态

### 错误指标

- 网络错误次数
- 解析错误次数
- 超时错误次数
- 总错误率

## 最佳实践

### 配置管理

- 使用配置文件管理认证信息
- 敏感配置可通过环境变量注入
- 定期轮换认证凭据

### 错误处理

- 分层处理不同类型错误
- 记录详细的错误上下文
- 提供清晰的错误信息

### 日志记录

- 使用结构化日志格式
- 记录关键操作和指标
- 避免记录敏感信息

### 测试覆盖

- 单元测试覆盖核心逻辑
- 集成测试验证端到端流程
- 使用Mock隔离外部依赖