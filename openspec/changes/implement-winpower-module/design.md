# implement-winpower-module 设计文档

## 架构设计

### 整体架构

WinPower 模块采用分层架构设计，确保模块间的职责分离和松耦合：

```
┌─────────────────────────────────────────────────────────────┐
│                     WinPower Module                        │
├─────────────────────────────────────────────────────────────┤
│  WinPowerClient Interface (统一入口)                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ CollectDeviceData()                                 │   │
│  │ GetConnectionStatus()                               │   │
│  │ GetLastCollectionTime()                             │   │
│  └─────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  Component Layer (组件层)                                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐   │
│  │HTTPClient   │ │TokenManager │ │DataParser           │   │
│  │             │ │             │ │                     │   │
│  │- Request    │ │- Login      │ │- Parse Response     │   │
│  │- SSL Config │ │- Cache      │ │- Validate Data      │   │
│  │- Timeout    │ │- Refresh    │ │- Type Conversion    │   │
│  └─────────────┘ └─────────────┘ └─────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│  Configuration Layer (配置层)                               │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Config                                             │   │
│  │ - BaseURL, Username, Password                      │   │
│  │ - Timeout, SSL Settings                            │   │
│  │ - Validation & Defaults                            │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────────┐
                    │  WinPower System    │
                    │  HTTP/HTTPS API     │
                    └─────────────────────┘
```

### 核心设计原则

1. **单一职责**：每个组件只负责特定的功能领域
2. **依赖注入**：通过接口实现依赖倒置，便于测试和替换
3. **错误隔离**：分层错误处理，避免错误传播
4. **配置驱动**：所有参数可通过配置调整
5. **并发安全**：支持多 goroutine 安全访问

## 组件设计

### 1. WinPowerClient (主客户端)

**职责**：作为模块的统一入口，协调各组件完成数据采集任务

**设计要点**：
- 实现 `WinPowerClient` 接口
- 组合 HTTPClient、TokenManager、DataParser
- 管理组件生命周期和错误处理
- 提供连接状态跟踪

**核心流程**：
```go
func (w *WinPowerClient) CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error) {
    // 1. 检查连接状态
    if !w.IsHealthy() {
        return nil, ErrConnectionNotReady
    }

    // 2. 获取有效 Token
    token, err := w.tokenManager.GetToken(ctx)
    if err != nil {
        w.logger.Error("Failed to get token", zap.Error(err))
        return nil, err
    }

    // 3. 发起数据请求
    response, err := w.httpClient.GetDeviceData(ctx, token)
    if err != nil {
        w.logger.Error("Failed to fetch device data", zap.Error(err))
        return nil, err
    }

    // 4. 解析响应数据
    data, err := w.dataParser.ParseResponse(response)
    if err != nil {
        w.logger.Error("Failed to parse response", zap.Error(err))
        return nil, err
    }

    // 5. 更新状态信息
    w.updateCollectionStatus(len(data))

    return data, nil
}
```

### 2. HTTPClient (HTTP 通信层)

**职责**：管理 HTTP 通信细节，封装底层网络操作

**设计要点**：
- 封装 `http.Client`，支持连接复用
- 支持 SSL/TLS 配置和验证跳过
- 统一的超时策略
- 结构化日志记录

**关键实现**：
```go
type HTTPClient struct {
    client    *http.Client
    baseURL   string
    timeout   time.Duration
    skipSSL   bool
    userAgent string
    logger    *zap.Logger
}

func (c *HTTPClient) configureTLS() *tls.Config {
    if c.skipSSL {
        return &tls.Config{
            InsecureSkipVerify: true,
        }
    }
    return nil
}
```

### 3. TokenManager (Token 管理)

**职责**：管理认证 Token 的完整生命周期

**设计要点**：
- Token 获取、缓存和自动刷新
- 并发安全的 Token 访问
- 智能刷新策略（过期前刷新）
- 认证错误的分类处理

**刷新策略**：
```go
func (tm *TokenManager) shouldRefresh() bool {
    if tm.tokenCache == nil {
        return true
    }

    // 在过期前阈值时间进行刷新
    timeUntilExpiry := time.Until(tm.tokenCache.ExpiresAt)
    return timeUntilExpiry <= tm.refreshThreshold
}
```

### 4. DataParser (数据解析)

**职责**：将 WinPower API 响应转换为标准化数据结构

**设计要点**：
- JSON 响应解析和验证
- 数据类型转换和清洗
- 错误数据的优雅处理
- 详细的解析日志

**数据转换流程**：
```go
func (p *DataParser) parseRealtimeData(raw map[string]interface{}) RealtimeData {
    data := RealtimeData{}

    // 能耗关键字段转换
    if watt, ok := raw["loadTotalWatt"].(string); ok {
        if val, err := strconv.ParseFloat(watt, 64); err == nil {
            data.LoadTotalWatt = val
            p.logger.Debug("Parsed loadTotalWatt", zap.Float64("value", val))
        } else {
            p.logger.Warn("Failed to parse loadTotalWatt",
                zap.String("raw", watt), zap.Error(err))
        }
    }

    // 其他字段转换...
    return data
}
```

## 数据流设计

### 采集流程

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Scheduler   │───►│WinPowerClient│───►│TokenManager │
│   每 5s     │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘
                           │                   │
                           ▼                   ▼
                    ┌─────────────┐    ┌─────────────┐
                    │ HTTPClient  │    │    API      │
                    │             │───►│  /api/v1/    │
                    └─────────────┘    │ auth/login  │
                           │           └─────────────┘
                           ▼                   │
                    ┌─────────────┐           ▼
                    │DataParser   │    ┌─────────────┐
                    │             │◄───│ HTTP Response│
                    └─────────────┘    └─────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ParsedDevice │
                    │    Data     │
                    └─────────────┘
```

### 错误处理流程

根据项目整体设计原则，WinPower 模块**不负责重试决策**，重试由上层 Scheduler 模块负责：

```
Error Detection
       │
       ▼
┌─────────────────┐    Error Type?     ┌─────────────────┐
│  Log Error      │────────────────────►│ Handle Error    │
│  (Sanitized)    │                    │  (Specific)     │
└─────────────────┘                    └─────────────────┘
       │                                       │
       ▼                                       ▼
┌─────────────────┐                    ┌─────────────────┐
│ Update Status   │                    │ Token Refresh?  │
│ Metrics         │                    │                 │
└─────────────────┘                    └─────────────────┘
       │                                       │
       ▼                                       ▼
┌─────────────────┐                    ┌─────────────────┐
│ Return Error    │                    │ Return Error    │
│ to Caller       │                    │ to Caller       │
└─────────────────┘                    └─────────────────┘
                                               │
                                         (Scheduler)
                                               │
                                               ▼
                                        ┌──────────────┐
                                        │Retry Decision│
                                        └──────────────┘
```

**重试责任划分**：
- **WinPower 模块**：只负责识别错误类型、记录详细错误信息、执行 Token 刷新
- **Scheduler 模块**：根据错误类型和业务策略决定是否重试
- **Collector 模块**：传递错误信息，不进行自动重试

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

// Validate 实现ConfigValidator接口用于配置验证
func (c *Config) Validate() error {
    // 验证BaseURL
    if c.BaseURL == "" {
        return fmt.Errorf("winpower base_url cannot be empty")
    }

    // 验证BaseURL格式
    if _, err := url.Parse(c.BaseURL); err != nil {
        return fmt.Errorf("invalid winpower base_url format: %w", err)
    }

    // 验证用户名
    if c.Username == "" {
        return fmt.Errorf("winpower username cannot be empty")
    }

    // 验证密码
    if c.Password == "" {
        return fmt.Errorf("winpower password cannot be empty")
    }

    // 验证超时配置
    if c.Timeout <= 0 {
        return fmt.Errorf("winpower timeout must be positive, got %v", c.Timeout)
    }
    if c.APITimeout <= 0 {
        return fmt.Errorf("winpower api_timeout must be positive, got %v", c.APITimeout)
    }

    // 验证刷新阈值
    if c.RefreshThreshold < time.Minute {
        return fmt.Errorf("winpower refresh_threshold must be at least 1 minute, got %v", c.RefreshThreshold)
    }

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

## 测试策略

### 单元测试结构

```
internal/winpower/
├── winpower_client_test.go      # 主客户端测试
├── http_client_test.go          # HTTP 客户端测试
├── token_manager_test.go        # Token 管理测试
├── data_parser_test.go          # 数据解析测试
├── config_test.go               # 配置测试
├── mocks/
│   ├── mock_http_client.go      # HTTP 客户端 Mock
│   ├── mock_token_manager.go    # Token 管理 Mock
│   └── mock_data_parser.go      # 数据解析 Mock
└── fixtures/
    ├── auth_response.json       # 认证响应示例
    ├── device_data.json         # 设备数据示例
    └── error_responses.json     # 错误响应示例
```

### 测试覆盖策略

1. **正常流程测试**：验证核心功能的正确性
2. **边界条件测试**：测试极端情况下的行为
3. **错误场景测试**：验证错误处理的正确性
4. **并发安全测试**：验证多 goroutine 环境下的安全性
5. **性能基准测试**：确保性能指标达标

### Mock 策略

```go
// MockWinPowerClient 提供可控的测试环境
type MockWinPowerClient struct {
    mockData     []ParsedDeviceData
    mockError    error
    callCount    int
    lastCallTime time.Time
}

func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error) {
    m.callCount++
    m.lastCallTime = time.Now()

    if m.mockError != nil {
        return nil, m.mockError
    }

    return m.mockData, nil
}
```

## 性能优化

### 连接管理优化

1. **连接复用**：使用单个 `http.Client` 实例
2. **连接池配置**：合理的 `MaxIdleConns` 和 `MaxIdleConnsPerHost`
3. **Keep-Alive**：启用 HTTP Keep-Alive 减少连接开销

### 内存优化

1. **对象复用**：复用解析过程中的临时对象
2. **内存池**：对于频繁分配的对象使用 sync.Pool
3. **及时释放**：确保大对象的及时 GC

### 并发优化

1. **读写锁**：Token 管理使用读写锁提高并发性能
2. **无锁设计**：在可能的情况下避免使用锁
3. **并发控制**：合理控制并发采集的数量

## 安全考虑

### 数据安全

1. **敏感信息保护**：
   - 不记录密码和 Token
   - 日志中脱敏处理
   - 内存中及时清理敏感数据

2. **传输安全**：
   - 优先使用 HTTPS
   - 支持 SSL 证书验证
   - 可配置跳过自签名证书验证

### 访问控制

1. **最小权限原则**：只请求必要的 API 权限
2. **会话管理**：Token 的安全存储和更新
3. **错误信息脱敏**：避免泄露系统内部信息

## 监控和可观测性

### 关键指标

1. **业务指标**：
   - 采集成功率
   - 平均采集耗时
   - 设备连接状态

2. **技术指标**：
   - HTTP 请求耗时
   - Token 刷新频率
   - 错误率统计

3. **资源指标**：
   - 内存使用量
   - Goroutine 数量
   - 连接池状态

### 日志策略

1. **结构化日志**：使用 zap 记录结构化日志
2. **分级记录**：根据重要性记录不同级别的日志
3. **上下文追踪**：使用 trace ID 关联相关操作
4. **敏感信息过滤**：确保不记录敏感信息

## 扩展性设计

### 接口扩展

1. **版本兼容**：API 接口支持版本演进
2. **设备适配**：支持不同设备类型的适配
3. **协议扩展**：为新的通信协议预留扩展点

### 配置扩展

1. **动态配置**：支持运行时配置更新
2. **环境适配**：不同环境的配置隔离
3. **功能开关**：支持功能的动态开启/关闭

### 功能扩展

1. **多数据源**：支持从多个 WinPower 实例采集数据
2. **数据过滤**：支持基于规则的数据过滤
3. **缓存机制**：支持可配置的数据缓存策略