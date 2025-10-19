# 数据采集模块设计文档

## 概述

数据采集模块（Collector）负责从 WinPower 系统获取设备数据，是整个导出器的数据源。该模块通过 HTTP/HTTPS 与 WinPower API 通信，涵盖认证、连接管理、数据获取、解析与错误处理。根据最新设计，模块不使用任何形式的连接池，统一复用一个 `http.Client` 实例以降低复杂性并确保稳定性。

### 设计目标

- 可靠性：稳定获取数据，具备完善的错误处理与日志
- 性能：复用单个 `http.Client` 降低开销，合理设置超时
- 容错性：对网络错误、认证错误、格式错误进行分层处理
- 可扩展性：支持不同设备类型与响应格式
- 简洁性：不使用连接池，不维护复杂的队列或 worker

## 架构设计

### 模块结构图

```
┌───────────────────────────────────────────────────────────────┐
│                       Collector Module                        │
├───────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │   HTTP       │  │   Auth      │  │    Data Parser      │   │
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
│  │              Collector Interface                       │   │
│  │  - CollectDeviceData(ctx)  // 始终包含电能计算          │   │
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
                    │ Data             │
                    └──────────────────┘
```

### 与调度器协作

- 调度器每 5 秒调用一次 `CollectDeviceData(ctx)`，从设备接口拉取最新数据并触发电能计算。
- 统一行为：无论由 Prometheus 拉取还是由调度器触发，`CollectDeviceData` 都会进行电能计算与持久化。

```go
// 统一采集入口：始终包含电能计算
func (c *Collector) CollectDeviceData(ctx context.Context) error {
    // 1) 获取/刷新 Token
    // 2) 调用设备数据接口
    // 3) 解析并校验数据
    // 4) 计算当前功率与增量电能（调用 Energy 模块累计）
    // 5) 调用存储持久化
    return nil
}
```

## 核心组件设计

### 1. HTTP 客户端 (HTTPClient)

#### 职责
- 管理与 WinPower 的 HTTP 请求与响应
- 处理 SSL/TLS 配置与证书验证（可跳过）
- 设置统一的请求超时与 User-Agent
- 简化：不使用连接池；复用单个 `http.Client` 实例

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

// SetTimeout 设置请求超时
func (c *HTTPClient) SetTimeout(timeout time.Duration)

// Close 关闭客户端并清理资源（预留）
func (c *HTTPClient) Close() error
```

#### 实现要点

```go
// NewHTTPClient 创建 HTTP 客户端实例（不配置连接池）
func NewHTTPClient(cfg config.WinPowerConfig, logger *zap.Logger) *HTTPClient {
    // 配置 TLS（可选跳过证书验证）
    tlsConfig := &tls.Config{InsecureSkipVerify: cfg.SkipSSL}

    transport := &http.Transport{
        TLSClientConfig:     tlsConfig,
        TLSHandshakeTimeout: 10 * time.Second,
        // 不设置任何连接池相关参数（如 MaxIdleConns 等）
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   cfg.Timeout,
    }

    return &HTTPClient{
        client:    client,
        baseURL:   cfg.BaseURL,
        timeout:   cfg.Timeout,
        skipSSL:   cfg.SkipSSL,
        userAgent: "WinPower-G2-Exporter/1.0",
        logger:    logger,
    }
}

// Do 执行请求并设置统一头
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
    req.Header.Set("User-Agent", c.userAgent)
    return c.client.Do(req)
}
```

### 2. 认证管理器 (AuthManager)

#### 职责
- 管理 JWT Token 的获取、缓存和刷新
- 处理登录认证流程
- 暴露 `GetToken` 与 `RefreshToken` 供调度器与采集流程使用

#### 关键数据结构（摘要）

```go
type AuthManager struct {
    client     *HTTPClient
    username   string
    password   string
    tokenCache *TokenCache
    logger     *zap.Logger
    mutex      sync.RWMutex
}

type TokenCache struct {
    token       string
    expiresAt   time.Time
    issuedAt    time.Time
    lastRefresh time.Time
    mutex       sync.RWMutex
}
```

#### 核心方法（摘要）

```go
func (am *AuthManager) GetToken() (string, error)
func (am *AuthManager) RefreshToken() (string, error)
func (am *AuthManager) IsTokenValid() bool
```

### 3. 数据解析器 (DataParser)

#### 职责
- 解析设备数据响应
- 验证数据格式与完整性
- 产出标准化的设备数据结构，供电能计算使用

#### 关键结构（摘要）

```go
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
```

## 错误处理与日志

- 网络错误：重试策略由上层调度控制；采集端返回明确错误
- 认证错误：优先尝试刷新 Token；失败则返回错误等待下次调度
- 解析错误：记录原始响应片段与字段定位，便于定位问题
- 统一使用结构化日志记录关键指标（耗时、响应码、设备数）
- 敏感信息：采集与请求日志不得包含密码、Token 值或 `Authorization` 头；仅记录 `device_id`、响应码与耗时等非敏感信息，错误内容需脱敏。

## 与存储的集成

- 电能计算完成后将增量与累计结果写入存储层（具体实现见 storage 设计文档）
- 采集时间戳与设备标识作为主键参与幂等写入，避免重复累计

## 配置建议

- `timeout` 建议不超过 10s；网络抖动场景可适当提高
- `skipSSL` 仅在自签证书且安全可控环境下启用
- 所有请求统一复用一个 `http.Client`，保证简单与稳定