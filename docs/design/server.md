# HTTP 服务器模块设计（简化版）

## 1. 概述

HTTP 服务器模块负责暴露最小且稳定的 Web 接口：`/metrics` 与 `/health`。模块仅处理 HTTP 层的路由、中间件与优雅关闭，不参与业务计算，最大化可维护性与可替换性。

## 2. 设计目标

- 高性能：基于 Gin，路由与中间件链最小化。
- 高可靠：恢复中间件、防止崩溃；优雅关闭。
- 易维护：职责单一、接口稳定、依赖清晰。
- 可观察：结构化日志、可选 pprof。

## 3. 职责边界

- 负责：路由注册、全局中间件、启动/停止、pprof（可选）。
- 不负责：认证流程、电能计算、采集逻辑、指标转换，这些由下层模块提供接口。

## 4. 依赖与输入/输出

- 输入：`server.Config`、`Logger`、`MetricsService`、`HealthService`。
- 输出：HTTP 服务（监听端口）；`/metrics` 返回 Prometheus 文本；`/health` 返回 JSON。

依赖关系（简化）：

```
server ──> metrics  ──> collector, energy
   │
   └──> health
```

## 5. 配置结构

Server模块定义自己的配置结构体，并通过注册机制与config模块集成：

```go
// internal/server/config.go
package server

import "time"

type Config struct {
    Port         int           `yaml:"port" validate:"min=1,max=65535"`
    Host         string        `yaml:"host" validate:"required"`
    Mode         string        `yaml:"mode" validate:"oneof=debug release test"`
    ReadTimeout  time.Duration `yaml:"read_timeout" validate:"min=1s"`
    WriteTimeout time.Duration `yaml:"write_timeout" validate:"min=1s"`
    IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"min=1s"`
    EnablePprof  bool          `yaml:"enable_pprof"`
}

func DefaultConfig() *Config {
    return &Config{
        Port:         9090,
        Host:         "0.0.0.0",
        Mode:         "release",
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
        EnablePprof:  false,
    }
}

func (c *Config) Validate() error {
    // 配置验证逻辑
    return nil
}

// Server模块通过构造函数接收配置参数
func NewHTTPServer(config *Config, logger Logger, metrics MetricsService, health HealthService) Server
```

**配置结构位置**：
- Config定义在`internal/server/config.go`中
- 包含Host、Port、Mode、ReadTimeout、WriteTimeout、IdleTimeout、EnablePprof等字段
- 可扩展CORS和RateLimit子配置结构

**配置参数说明**：
- `config`: 使用server模块的Config结构，包含HTTP服务器的所有配置参数
- `logger`: 日志记录器实例
- `metrics`: 指标服务接口
- `health`: 健康检查服务接口

**配置注册**：
- Server模块通过init()函数自动注册配置到config系统
- Config模块发现并加载server配置到统一配置结构中

说明：
- 为简化维护，不在服务器内直接提供 TLS/证书配置，生产环境建议由反向代理终结 TLS。

## 6. 关键接口与伪代码

```go
// 依赖接口（由业务模块提供）
type MetricsService interface {
    Render(ctx context.Context) (string, error) // Prometheus 文本
}

type HealthService interface {
    Check(ctx context.Context) (status string, details map[string]any)
}

// Server 对外接口
type Server interface {
    Start() error
    Stop(ctx context.Context) error
}

// 实现结构
type HTTPServer struct {
    cfg     *Config
    log     Logger
    engine  *gin.Engine
    srv     *http.Server
    metrics MetricsService
    health  HealthService
}

func NewHTTPServer(config *Config, log Logger, metrics MetricsService, health HealthService) *HTTPServer {
    // 1) 设置 Gin 模式
    // 2) 创建引擎与中间件
    // 3) 注册路由与 pprof（可选）
    // 4) 构建 http.Server（超时/头部限制）
}

// 中间件链（按顺序）
func (s *HTTPServer) setupGlobalMiddleware() {
    // Logger
    // Recovery
    // CORS（按配置）
    // RateLimit（按配置）
    // Metrics（简单统计，如请求耗时）
    // Timeout（如需要）
}

func (s *HTTPServer) setupRoutes() {
    // GET /health
    // GET /metrics
    // 404 处理器
    // /debug/pprof（可选）
}

func (s *HTTPServer) Start() error {
    // ListenAndServe（不启用 TLS）
}

func (s *HTTPServer) Stop(ctx context.Context) error {
    // 优雅关闭（Shutdown + 超时控制）
}
```

## 7. 中间件与路由

中间件（默认最小可用集）：
- Logger：记录方法、路径、耗时、状态码；避免输出敏感信息。
- Recovery：捕获 panic，返回 500 并记录错误。
- CORS（可选）：按配置允许跨域；默认关闭。
- RateLimit（可选）：简单令牌桶限流；默认关闭。
- Metrics（可选）：记录基本请求耗时与数量；可用于内部监控。

路由：
- GET `/health`：返回 `{status: "ok", timestamp: <RFC3339>, version: <semver>}`。
- GET `/metrics`：调用 `MetricsService.Render()`，返回 `text/plain; version=0.0.4`。
- 404：统一 JSON：`{"error":"not_found","path":"/xxx","ts":"..."}`。
- `/debug/pprof`：`EnablePprof=true` 时启用。

## 8. 请求流程（简化）

```
Client ──HTTP──> Gin Engine
   │            [Logger] -> [Recovery] -> [CORS?] -> [RateLimit?]
   │                    -> Route Match
   ├── GET /health  ───> health.Check()  ───> JSON
   └── GET /metrics ───> metrics.Render() ───> Prometheus 文本
```

## 9. 启停与优雅关闭

- 启动：记录 `host/port/mode`；不使用 TLS；`http.ErrServerClosed` 不视为错误。
- 停止：接受 `ctx` 控制超时（默认 30s）；确保连接优雅关闭；记录关闭结果。

## 10. 错误处理与安全

- 错误处理：统一使用结构化日志记录；对外返回最小必要信息，避免泄露内部细节。
- 安全建议：
  - 通过反向代理统一终结 TLS 与认证；Exporter 保持纯净职责。
  - 启用 CORS 时，尽量限定来源域并关闭凭证。
  - 限流在高并发场景下开启，防止抓取风暴影响稳定性。

## 11. 扩展点

- 自定义中间件：通过 `Use()` 接口插入额外中间件（例如审计）。
- 管理端点：如需运维接口，建议单独分组 `/admin`，并置于反向代理认证之后。
- 指标：`MetricsService` 可自由扩展指标类型，Server 不感知具体格式以保持解耦。