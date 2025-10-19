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
                    │ Trigger Energy   │
                    │ Calculation      │
                    └──────────────────┘
                         │
                         ▼
                    ┌──────────────────┐
                    │ Return Device    │
                    │ Data             │
                    └──────────────────┘
```

## 核心组件设计

### 1. WinPower客户端 (WinPowerClient)

#### 职责
- 统一管理认证和数据采集功能
- 提供 `CollectDeviceData` 接口供调度器调用
- 内部管理Token生命周期和HTTP通信

#### 核心接口

```go
// WinPowerClient WinPower客户端接口
type WinPowerClient interface {
    // CollectDeviceData 采集设备数据并触发电能计算
    CollectDeviceData(ctx context.Context) error

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
func NewHTTPClient(cfg config.WinPowerConfig, logger *zap.Logger) *HTTPClient

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

#### 关键结构

```go
// ParsedDeviceData 解析后的设备数据
type ParsedDeviceData struct {
    DeviceID     string                 `json:"device_id"`
    DeviceName   string                 `json:"device_name"`
    DeviceType   int                    `json:"device_type"`
    DeviceModel  string                 `json:"device_model"`
    Connected    bool                   `json:"connected"`
    RealtimeData map[string]float64     `json:"realtime_data"`
    PowerInfo    PowerInfo              `json:"power_info"`
    EnergyInfo   EnergyInfo             `json:"energy_info"`
    Timestamp    time.Time              `json:"timestamp"`
}

// PowerInfo 功率信息
type PowerInfo struct {
    ActivePower    float64 `json:"active_power"`    // 有功功率(W)
    ReactivePower  float64 `json:"reactive_power"`  // 无功功率(var)
    ApparentPower  float64 `json:"apparent_power"`  // 视在功率(VA)
    PowerFactor    float64 `json:"power_factor"`    // 功率因数
}
```

## 配置设计

### WinPowerConfig配置结构

```go
// WinPowerConfig WinPower模块配置
type WinPowerConfig struct {
    BaseURL         string        `yaml:"base_url" json:"base_url"`
    Username        string        `yaml:"username" json:"username"`
    Password        string        `yaml:"password" json:"password"`
    Timeout         time.Duration `yaml:"timeout" json:"timeout"`
    APITimeout      time.Duration `yaml:"api_timeout" json:"api_timeout"`
    MaxRetries      int           `yaml:"max_retries" json:"max_retries"`
    SkipSSLVerify   bool          `yaml:"skip_ssl_verify" json:"skip_ssl_verify"`
    RefreshThreshold time.Duration `yaml:"refresh_threshold" json:"refresh_threshold"`
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
  max_retries: 2
  skip_ssl_verify: false
  refresh_threshold: 5m
```

## 错误处理与日志

### 错误处理策略

- **网络错误**：重试策略由上层调度控制；采集端返回明确错误
- **认证错误**：优先尝试刷新Token；失败则返回错误等待下次调度
- **解析错误**：记录原始响应片段与字段定位，便于定位问题
- **超时错误**：记录超时信息，调整重试间隔

### 日志记录

- 统一使用结构化日志记录关键指标（耗时、响应码、设备数）
- **敏感信息保护**：不得在日志中记录密码、Token值或Authorization头
- 仅记录非敏感字段：device_id、响应码、耗时、状态等
- 错误信息需脱敏处理

## 与其他模块的集成

### 与Energy模块集成

```go
// 在采集成功后触发电能计算
func (w *WinPowerClient) CollectDeviceData(ctx context.Context) error {
    // 1. 获取有效Token
    token, err := w.tokenManager.GetToken(ctx)
    if err != nil {
        return err
    }

    // 2. 采集设备数据
    data, err := w.collectData(ctx, token)
    if err != nil {
        return err
    }

    // 3. 触发电能计算
    for deviceID, power := range data.Powers {
        if err := w.energyService.Calculate(deviceID, power); err != nil {
            w.logger.Error("电能计算失败",
                zap.String("device_id", deviceID),
                zap.Error(err))
        }
    }

    return nil
}
```

### 与Scheduler模块集成

- 调度器每5秒调用 `winpower.CollectDeviceData(ctx)`
- 统一行为：无论由什么触发，`CollectDeviceData` 都会进行电能计算与持久化

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
}

func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) error {
    return m.collectError
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