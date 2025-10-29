# WinPower G2 API 认证协议

## 概述

WinPower G2 系统使用基于 Token 的认证机制。客户端需要通过用户名和密码进行登录认证，获取访问令牌，后续的 API 请求都需要在 Header 中携带此令牌。Token 的具体类型和格式由系统决定，客户端无需解析 Token 内容。

## 认证流程

### 1. 用户登录

#### 请求信息

- **URL**: `/api/v1/auth/login`
- **方法**: `POST`
- **Content-Type**: `application/json`
- **Content-Language**: `en`

#### 请求头

```
POST /api/v1/auth/login HTTP/1.1
Host: {host}:{port}
Connection: keep-alive
Content-Length: {length}
User-Agent: Mozilla/5.0 (compatible; WinPower-Exporter/1.0)
Accept: application/json, text/plain, */*
Content-Type: application/json
Content-language: en
```

#### 请求体

```json
{
    "username": "{username}",
    "password": "{password}"
}
```

**参数说明:**

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| username | string | 是 | 登录用户名 |
| password | string | 是 | 登录密码 |

#### 响应信息

**成功响应 (HTTP 200)**

```json
{
    "code": "000000",
    "message": "OK",
    "data": {
        "deviceId": "{device_id}",
        "token": "{access_token}"
    }
}
```

**响应字段说明:**

| 字段名 | 类型 | 说明 |
|--------|------|------|
| code | string | 响应状态码，"000000" 表示成功 |
| message | string | 响应消息 |
| data.deviceId | string | 设备唯一标识符 |
| data.token | string | 访问令牌 |

### 2. Token 管理

#### Token 有效期

- Token 有效期为固定 1 小时（3600 秒）
- 客户端在获取 Token 时记录创建时间，计算过期时间
- 客户端无需解析 Token 内容来确定过期时间
- Token 过期后需要重新登录获取新 Token

#### Token 使用

- Token 通过 Authorization header 传输
- 使用 Bearer 认证方案
- 客户端应缓存 Token 避免频繁登录
- 在 Token 过期前自动刷新

### 3. 使用 Token 访问 API

在后续的 API 请求中，需要在请求头中添加 Authorization 字段：

```
Authorization: Bearer {access_token}
```

#### 示例请求头

```
GET /api/v1/deviceData/detail/list HTTP/1.1
Host: {host}:{port}
Connection: keep-alive
Authorization: Bearer {access_token}
User-Agent: Mozilla/5.0 (compatible; WinPower-Exporter/1.0)
Accept: application/json, text/plain, */*
Content-language: zh-CN
```

## 错误处理

### 常见错误响应

#### 认证失败 (HTTP 401)

```json
{
    "code": "error_code",
    "message": "Authentication failed",
    "data": null
}
```

#### Token 过期 (HTTP 401)

当 Token 过期时，需要重新进行登录获取新的 Token。

#### 无效 Token (HTTP 401)

```json
{
    "code": "INVALID_TOKEN",
    "message": "Invalid or expired token",
    "data": null
}
```

## 安全注意事项

1. **HTTPS 传输**: 建议在生产环境中使用 HTTPS 协议传输认证信息
2. **Token 存储**: Token 应该安全存储，避免泄露
3. **Token 刷新**: Token 有效期为固定 1 小时，建议在过期前 5-10 分钟自动刷新
4. **SSL 验证**: 对于使用自签名证书的 WinPower 系统，可以配置跳过 SSL 验证
5. **日志脱敏**: 日志中不应记录用户名、密码、访问令牌或 `Authorization` 头等敏感信息；认证事件仅记录非敏感字段（如设备ID、剩余有效期），错误信息需脱敏或掩码处理。

## 实现建议

1. **Token 缓存**: 客户端应该缓存 Token 并在过期前自动刷新
2. **错误处理**: 实现适当的错误处理机制
3. **连接池**: 使用 HTTP 连接池提高性能
4. **超时设置**: 设置合理的请求超时时间

## 示例代码

### cURL 示例

```bash
# 登录获取 Token
curl -X POST "http://192.168.1.100:8081/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "Content-language: en" \
  -d '{"username":"admin","password":"password"}'

# 使用 Token 访问 API
curl -X GET "http://192.168.1.100:8081/api/v1/deviceData/detail/list" \
  -H "Authorization: Bearer {access_token}" \
  -H "Accept: application/json"
```

### Go 示例

```go
type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Data    struct {
        DeviceID string `json:"deviceId"`
        Token    string `json:"token"`
    } `json:"data"`
}

// 登录获取 Token
func login(client *http.Client, host, username, password string) (string, error) {
    loginReq := LoginRequest{
        Username: username,
        Password: password,
    }

    jsonData, _ := json.Marshal(loginReq)

    resp, err := client.Post(
        fmt.Sprintf("http://%s/api/v1/auth/login", host),
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var loginResp LoginResponse
    if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
        return "", err
    }

    return loginResp.Data.Token, nil
}
```