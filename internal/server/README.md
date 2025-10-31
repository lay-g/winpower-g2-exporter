# Server Module

HTTP服务器模块，提供WinPower G2 Exporter的Web接口。

## 概述

Server模块基于Gin框架实现，负责对外暴露HTTP API端点，包括：

- `/health` - 健康检查端点
- `/metrics` - Prometheus指标导出端点
- `/debug/pprof/*` - 性能分析端点（可选）

## 特性

- **高性能**：基于Gin框架的高性能路由
- **高可靠**：内置Recovery中间件防止崩溃，支持优雅关闭
- **易维护**：职责单一，接口清晰，依赖解耦
- **可观察**：结构化日志记录所有请求
- **灵活配置**：支持多种运行模式和超时配置

## 架构设计

### 核心组件

```
┌─────────────────────────────────────────┐
│          HTTPServer (Gin)               │
├─────────────────────────────────────────┤
│  Middleware:                            │
│  - Recovery (panic处理)                 │
│  - Logger (请求日志)                     │
├─────────────────────────────────────────┤
│  Routes:                                │
│  - GET /health    → HealthService       │
│  - GET /metrics   → MetricsService      │
│  - NoRoute        → 404 Handler         │
│  - /debug/pprof/* → pprof (可选)        │
└─────────────────────────────────────────┘
```

### 依赖关系

```
Server → MetricsService (提供 /metrics 处理)
      → HealthService  (提供 /health 检查)
      → Logger         (日志记录)
```

## 使用方法

### 基本用法

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/lay-g/winpower-g2-exporter/internal/server"
    "go.uber.org/zap"
)

func main() {
    // 创建配置
    cfg := server.DefaultConfig()
    cfg.Port = 8080
    cfg.EnablePprof = true

    // 创建日志
    logger, _ := zap.NewProduction()

    // 创建服务
    metricsService := NewMetricsService() // 实现 server.MetricsService
    healthService := NewHealthService()   // 实现 server.HealthService

    // 创建HTTP服务器
    srv, err := server.NewHTTPServer(cfg, logger, metricsService, healthService)
    if err != nil {
        logger.Fatal("Failed to create server", zap.Error(err))
    }

    // 启动服务器
    if err := srv.Start(); err != nil {
        logger.Fatal("Failed to start server", zap.Error(err))
    }

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Stop(ctx); err != nil {
        logger.Error("Server shutdown error", zap.Error(err))
    }
}
```

### 自定义配置

```go
cfg := &server.Config{
    Port:            9090,
    Host:            "127.0.0.1",
    Mode:            "debug",           // debug, release, test
    ReadTimeout:     15 * time.Second,
    WriteTimeout:    15 * time.Second,
    IdleTimeout:     90 * time.Second,
    EnablePprof:     true,
    ShutdownTimeout: 45 * time.Second,
}

// 验证配置
if err := cfg.Validate(); err != nil {
    log.Fatal(err)
}
```

## 配置说明

| 字段            | 类型     | 默认值    | 说明                        |
| --------------- | -------- | --------- | --------------------------- |
| Port            | int      | 8080      | 监听端口 (1-65535)          |
| Host            | string   | "0.0.0.0" | 绑定地址                    |
| Mode            | string   | "release" | Gin模式: debug/release/test |
| ReadTimeout     | duration | 10s       | 读取超时                    |
| WriteTimeout    | duration | 10s       | 写入超时                    |
| IdleTimeout     | duration | 60s       | 空闲超时                    |
| EnablePprof     | bool     | false     | 启用pprof端点               |
| ShutdownTimeout | duration | 30s       | 优雅关闭超时                |

## 接口定义

### Server接口

```go
type Server interface {
    Start() error
    Stop(ctx context.Context) error
}
```

### MetricsService接口

实现此接口以提供Prometheus指标：

```go
type MetricsService interface {
    HandleMetrics(c *gin.Context)
}
```

### HealthService接口

实现此接口以提供健康检查：

```go
type HealthService interface {
    Check(ctx context.Context) (status string, details map[string]any)
}
```

健康状态值：
- `"ok"` 或 `"healthy"` → HTTP 200
- 其他值 → HTTP 503

### Logger接口

实现此接口以提供日志记录：

```go
type Logger interface {
    Info(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Debug(msg string, keysAndValues ...interface{})
}
```

## API端点

### GET /health

健康检查端点。

**响应示例**：
```json
{
  "status": "ok",
  "details": {
    "service": "healthy",
    "uptime": "1h30m"
  }
}
```

**状态码**：
- `200` - 服务健康
- `503` - 服务不健康

### GET /metrics

Prometheus指标导出端点，返回Prometheus文本格式的指标数据。

**响应示例**：
```
# HELP winpower_device_connected Device connection status
# TYPE winpower_device_connected gauge
winpower_device_connected{device_id="1",device_name="UPS1"} 1
```

### GET /debug/pprof/*

性能分析端点（需要配置 `EnablePprof: true`）。

可用端点：
- `/debug/pprof/` - 索引页面
- `/debug/pprof/cmdline` - 命令行参数
- `/debug/pprof/profile` - CPU分析
- `/debug/pprof/heap` - 堆内存分析
- `/debug/pprof/goroutine` - Goroutine分析
- 更多...

## 中间件

### Logger中间件

记录每个HTTP请求的详细信息：
- 请求方法和路径
- 状态码
- 响应时间
- 客户端IP
- User-Agent

### Recovery中间件

捕获panic，防止服务崩溃：
- 自动恢复panic
- 记录错误日志
- 返回标准化的500错误响应

## 错误处理

所有错误响应都使用统一的JSON格式：

```json
{
  "error": "error message",
  "path": "/requested/path",
  "ts": "2025-10-31T10:00:00Z"
}
```

## 优雅关闭

服务器支持优雅关闭机制：

1. 停止接受新连接
2. 等待现有请求完成（最长 `ShutdownTimeout`）
3. 关闭服务器
4. 释放资源

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
if err := srv.Stop(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./internal/server/... -v

# 运行测试并生成覆盖率报告
go test ./internal/server/... -cover -coverprofile=coverage.out

# 查看覆盖率详情
go tool cover -html=coverage.out
```

### 使用Mock

项目提供了Mock实现用于测试：

```go
import "github.com/lay-g/winpower-g2-exporter/internal/server/mocks"

mockMetrics := &mocks.MetricsService{
    HandleMetricsFunc: func(c *gin.Context) {
        c.String(200, "test metrics")
    },
}

mockHealth := &mocks.HealthService{}
mockLogger := &mocks.Logger{}

srv, _ := server.NewHTTPServer(cfg, mockLogger, mockMetrics, mockHealth)
```

## 性能优化

### 中间件顺序

中间件按以下顺序执行，确保最优性能和错误处理：

1. Recovery（最外层，捕获所有panic）
2. Logger（记录请求信息）
3. 业务路由处理

### 连接管理

合理配置超时参数以优化资源使用：

```go
cfg := server.DefaultConfig()
cfg.ReadTimeout = 10 * time.Second  // 防止慢客户端
cfg.WriteTimeout = 10 * time.Second // 防止慢写入
cfg.IdleTimeout = 60 * time.Second  // 保持连接重用
```

## 安全考虑

### 生产环境建议

1. **使用反向代理终结TLS**
   ```
   [Client] --HTTPS--> [Nginx/Caddy] --HTTP--> [Server]
   ```

2. **禁用pprof端点**
   ```go
   cfg.EnablePprof = false  // 生产环境
   ```

3. **限制监听地址**
   ```go
   cfg.Host = "127.0.0.1"  // 仅本地访问
   ```

4. **添加认证中间件**（如需要）
   ```go
   // 自定义认证逻辑
   ```

## 故障排查

### 服务启动失败

1. **端口被占用**
   ```
   Error: listen tcp :8080: bind: address already in use
   ```
   解决：更改端口或停止占用端口的进程

2. **权限不足**
   ```
   Error: listen tcp :80: bind: permission denied
   ```
   解决：使用1024以上端口或以root运行（不推荐）

### 请求超时

检查配置的超时参数是否合理：
```go
cfg.ReadTimeout = 30 * time.Second  // 增加读取超时
cfg.WriteTimeout = 30 * time.Second // 增加写入超时
```

### 内存泄漏

启用pprof进行分析：
```bash
# 查看堆内存
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# 查看goroutine
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

## 参考资料

- [Gin框架文档](https://gin-gonic.com/docs/)
- [Go net/http包](https://pkg.go.dev/net/http)
- [Prometheus指标格式](https://prometheus.io/docs/instrumenting/exposition_formats/)
- [pprof性能分析](https://pkg.go.dev/net/http/pprof)
