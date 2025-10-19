# 认证模块设计

## 1. 概述

认证模块负责管理与 WinPower G2 系统的身份认证,包括用户登录、Token 获取、Token 缓存管理和自动刷新机制。

### 1.1 设计目标

- **安全性**: 安全的认证流程和 Token 管理
- **可靠性**: 自动 Token 刷新和重试机制  
- **性能**: Token 缓存避免频繁登录
- **易用性**: 简单的认证接口
- **可测试**: 支持 Mock 和单元测试

### 1.2 功能范围

- 用户登录认证
- Token 获取和有效期检查
- Token 有效期管理
- Token 自动刷新
- 认证状态监控

## 2. 认证流程

### 2.1 登录流程图

```
┌──────────────┐
│   Start      │
└──────┬───────┘
       │
       ▼
┌────────────────────────┐
│ Check Cached Token     │
└──────┬─────────┬───────┘
       │         │
   Valid│         │Expired/None
       │         ▼
       │    ┌─────────────────┐
       │    │  POST Login API │
       │    │  /api/v1/auth   │
       │    │      /login     │
       │    └──────┬──────────┘
       │           │
       │           ▼
       │    ┌─────────────────┐
       │    │  Parse Response │
       │    │  Extract Token  │
       │    └──────┬──────────┘
       │           │
       │           ▼
       │    ┌─────────────────┐
       │    │  Get Token Info │
       │    │  (No Parse)     │
       │    └──────┬──────────┘
       │           │
       │           ▼
       │    ┌─────────────────┐
       │    │  Cache Token    │
       │    └──────┬──────────┘
       │           │
       └───────────┘
               │
               ▼
       ┌──────────────┐
       │ Return Token │
       └──────────────┘
```

### 2.2 Token 刷新流程

```
┌───────────────────┐
│  API Request      │
└────────┬──────────┘
         │
         ▼
┌────────────────────────┐
│ Check Token Expiry     │
└────────┬───────┬───────┘
         │       │
    Valid│       │Expired (< 5min)
         │       │
         │       ▼
         │  ┌──────────────────┐
         │  │  Refresh Token   │
         │  │  (Re-login)      │
         │  └────────┬─────────┘
         │           │
         └───────────┘
                 │
                 ▼
         ┌──────────────────┐
         │  Use New Token   │
         └──────────────────┘
```

## 3. 数据结构设计

### 3.1 认证配置

认证模块不定义独立的配置结构体，直接使用config模块提供的配置：

```go
// 认证模块通过构造函数接收配置参数
func NewAuthenticator(winPowerConfig *config.WinPowerConfig, refreshThreshold time.Duration, logger log.Logger) Authenticator
```

**配置参数说明**：
- `winPowerConfig`: 使用config模块的WinPowerConfig结构，包含URL、用户名、密码等认证必需参数
- `refreshThreshold`: Token刷新阈值，建议设置为5-10分钟
- `logger`: 日志记录器实例

### 3.2 Token 信息

```go
// TokenInfo Token 信息
type TokenInfo struct {
    // Token 字符串
    Token string

    // 设备 ID
    DeviceID string

    // Token 创建时间
    CreatedAt time.Time

    // Token 过期时间
    ExpiresAt time.Time

    // Token 有效期（小时）
    ExpiredHours int
}

// IsValid 检查 Token 是否有效
func (t *TokenInfo) IsValid() bool {
    return time.Now().Before(t.ExpiresAt)
}

// NeedRefresh 检查是否需要刷新
func (t *TokenInfo) NeedRefresh(threshold time.Duration) bool {
    return time.Until(t.ExpiresAt) < threshold
}

// RemainingTime 剩余有效时间
func (t *TokenInfo) RemainingTime() time.Duration {
    return time.Until(t.ExpiresAt)
}
```

### 3.3 认证接口

```go
// Authenticator 认证器接口
type Authenticator interface {
    // Login 登录获取 Token
    Login(ctx context.Context) (*TokenInfo, error)
    
    // GetToken 获取有效 Token（自动刷新）
    GetToken(ctx context.Context) (string, error)
    
    // RefreshToken 刷新 Token
    RefreshToken(ctx context.Context) error
    
    // IsAuthenticated 检查是否已认证
    IsAuthenticated() bool
    
    // Logout 登出
    Logout() error
}
```

## 4. 实现设计

### 4.1 认证管理器

```go
// authenticator 认证管理器实现
type authenticator struct {
    winPowerConfig *config.WinPowerConfig
    refreshThreshold time.Duration
    client         *http.Client
    token          *TokenInfo
    mu             sync.RWMutex
    logger         log.Logger
}

// NewAuthenticator 创建认证管理器
func NewAuthenticator(winPowerConfig *config.WinPowerConfig, refreshThreshold time.Duration, logger log.Logger) Authenticator {
    // 创建 HTTP 客户端，配置超时和 SSL 设置
    // 根据 winPowerConfig.SkipSSLVerify 配置 TLS 客户端
    // 使用 winPowerConfig.Timeout 设置请求超时
    // 初始化认证管理器结构体并返回
}
```

### 4.2 登录实现

```go
// Login 登录获取 Token
func (a *authenticator) Login(ctx context.Context) (*TokenInfo, error) {
    // 记录调试日志，包含 a.winPowerConfig.URL 和用户名
    // 构建登录请求体，包含 a.winPowerConfig.Username 和 a.winPowerConfig.Password
    // 序列化请求体为 JSON 格式
    // 构建 HTTP 请求，设置正确的 URL 和头部
    // 执行 HTTP POST 请求到登录接口
    // 处理请求错误，记录错误日志
    // 解析响应 JSON 数据
    // 检查响应状态码，处理登录失败情况
    // 调用 extractTokenInfo 提取 Token 信息
    // 使用锁安全地缓存 Token 信息
    // 记录成功日志，包含设备 ID 和过期时间
    // 返回 Token 信息或错误
}
```

### 4.3 Token 信息提取

```go
// extractTokenInfo 提取 Token 信息（不解析 Token 内容）
func (a *authenticator) extractTokenInfo(tokenString, deviceID string, expiredHours int) (*TokenInfo, error) {
    // 构造 TokenInfo 结构体，设置原始 Token 和设备 ID
    // 设置 Token 创建时间为当前时间
    // 根据 expiredHours 计算过期时间（通常为 1 小时）
    // 返回 Token 信息或错误
}
```

### 4.4 自动刷新

```go
// GetToken 获取有效 Token（自动刷新）
func (a *authenticator) GetToken(ctx context.Context) (string, error) {
    // 使用读锁获取缓存的 Token 信息
    // 检查是否为首次获取（无缓存 Token）
    // 如果无缓存，调用 Login 方法进行登录
    // 检查 Token 是否需要刷新（基于阈值）
    // 记录刷新日志，包含剩余时间
    // 尝试刷新 Token，处理刷新失败情况
    // 如果刷新失败但 Token 仍有效，继续使用
    // 重新获取刷新后的 Token 信息
    // 验证 Token 最终有效性
    // 返回 Token 字符串或错误
}

// RefreshToken 刷新 Token
func (a *authenticator) RefreshToken(ctx context.Context) error {
    // 记录刷新调试日志
    // 调用 Login 方法重新登录获取新 Token
    // 返回刷新结果或包装后的错误
}
```

## 5. 错误处理

### 5.1 错误类型定义

```go
var (
    // ErrAuthFailed 认证失败
    ErrAuthFailed = errors.New("authentication failed")
    
    // ErrTokenExpired Token 过期
    ErrTokenExpired = errors.New("token expired")
    
    // ErrTokenInvalid Token 无效
    ErrTokenInvalid = errors.New("token invalid")
    
    // ErrNetworkError 网络错误
    ErrNetworkError = errors.New("network error")
)
```


## 6. 测试设计

### 6.1 Mock 认证器

```go
// MockAuthenticator Mock 认证器
type MockAuthenticator struct {
    token *TokenInfo
    err   error
}

func (m *MockAuthenticator) Login(ctx context.Context) (*TokenInfo, error) {
    // 返回预设的 Token 或错误
}

func (m *MockAuthenticator) GetToken(ctx context.Context) (string, error) {
    // 检查是否有预设错误
    // 返回 Token 字符串或错误
}

func (m *MockAuthenticator) RefreshToken(ctx context.Context) error {
    // 返回预设的错误或 nil
}

func (m *MockAuthenticator) IsAuthenticated() bool {
    // 检查 Token 是否存在且有效
    // 返回认证状态
}

func (m *MockAuthenticator) Logout() error {
    // 返回 nil（表示登出成功）
}
```

### 6.2 单元测试

```go
func TestLogin(t *testing.T) {
    // 创建模拟 HTTP 服务器用于测试
    // 验证请求路径和方法
    // 返回模拟的登录响应数据
    // 创建测试配置，使用模拟服务器地址
    // 创建认证器实例
    // 调用登录方法
    // 验证登录结果和 Token 信息
}
```

## 7. 使用示例

### 7.1 基本使用

```go
// 创建认证器
winPowerConfig := &config.WinPowerConfig{
    URL:           "https://192.168.1.100:8081",
    Username:      "admin",
    Password:      "password",
    SkipSSLVerify: true,
    Timeout:       30 * time.Second,
}

authenticator := auth.NewAuthenticator(winPowerConfig, 5*time.Minute, logger)

// 获取 Token
token, err := authenticator.GetToken(context.Background())
if err != nil {
    log.Fatal(err)
}

// 使用 Token 调用 API
req.Header.Set("Authorization", "Bearer "+token)
```

### 7.2 集成到 HTTP 客户端

```go
// AuthTransport 认证传输层
type AuthTransport struct {
    Base http.RoundTripper
    Auth auth.Authenticator
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // 从认证器获取有效的 Token
    // 处理 Token 获取失败的情况
    // 在请求头中添加 Bearer Token 认证信息
    // 调用底层传输层执行 HTTP 请求
    // 返回响应或错误
}

func (t *AuthTransport) base() http.RoundTripper {
    // 检查是否有自定义的基础传输层
    // 如果有则返回自定义传输层
    // 否则返回默认的 HTTP 传输层
}

// 使用示例
client := &http.Client{
    Transport: &AuthTransport{
        Auth: authenticator,
    },
}
```

## 8. 最佳实践

### 8.1 Token 缓存

- 缓存有效 Token 避免频繁登录
- 使用读写锁保证并发安全
- 设置合理的刷新阈值

### 8.2 错误处理

- 记录详细的错误日志
- 明确错误类型和原因
- 提供清晰的错误信息

### 8.3 安全考虑

- 不在日志中记录密码、Token 值或 `Authorization` 头；记录认证事件时仅包含非敏感字段（如 `device_id`、`expires_at`、`status`）。
- 错误信息脱敏：避免在错误日志中包含用户名/密码/Token 原文；必要时以布尔或掩码形式呈现。
- 使用 HTTPS 传输认证信息；生产环境建议由反向代理进行 TLS 终止。
- Token 安全存储并按到期阈值提前刷新；不在日志或标准输出泄露 Token。

### 8.4 性能优化

- Token 缓存减少网络请求
- 并发安全的 Token 访问
- 合理的超时设置
