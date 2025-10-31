# Server模块实现提案

## 变更概述

本提案旨在实现WinPower G2 Prometheus Exporter的HTTP服务器模块，提供最小且稳定的Web接口（`/metrics`和`/health`），基于Gin框架实现高性能、高可靠、易维护的HTTP服务。

## Why

### 业务需求
- **指标导出需求**：Prometheus监控系统需要通过HTTP协议抓取设备指标数据，要求提供标准的/metrics端点
- **健康检查需求**：容器化部署和负载均衡器需要通过/health端点检查服务可用性
- **调试支持需求**：开发和运维过程中需要pprof功能进行性能分析和问题诊断

### 技术驱动因素
- **架构完整性**：Server模块是整个系统架构的最后一层，负责对外暴露HTTP接口
- **模块化设计**：需要实现职责单一的HTTP服务层，与业务逻辑完全解耦
- **可维护性**：基于Gin框架简化HTTP服务开发，提高代码可维护性

### 约束条件
- **不与config模块集成**：本提案仅定义配置结构，config模块集成将在后续独立工作中完成
- **TDD开发模式**：严格遵循测试驱动开发，先编写测试用例再实现功能
- **性能要求**：支持并发访问，响应时间在可接受范围内，无明显性能瓶颈

## 设计目标

根据 `docs/design/server.md` 的设计文档，实现一个职责单一的HTTP服务器模块：

- **高性能**：基于Gin，路由与中间件链最小化
- **高可靠**：恢复中间件、防止崩溃；优雅关闭
- **易维护**：职责单一、接口稳定、依赖清晰
- **可观察**：结构化日志、可选pprof

## 模块边界

### 职责
- 路由注册和HTTP请求处理
- 全局中间件管理（Logger、Recovery）
- 启动/停止和优雅关闭
- 可选的pprof调试支持

### 非职责
- 认证流程（由下层模块处理）
- 电能计算（由energy模块处理）
- 采集逻辑（由collector模块处理）
- 指标转换（由metrics模块处理）

## 依赖关系

```
server ──> metrics  ──> collector, energy
   │
   └──> health
```

## 配置定义

参考设计文档，server模块将定义自己的配置结构：

```go
type Config struct {
    Port         int           `yaml:"port" validate:"min=1,max=65535"`
    Host         string        `yaml:"host" validate:"required"`
    Mode         string        `yaml:"mode" validate:"oneof=debug release test"`
    ReadTimeout  time.Duration `yaml:"read_timeout" validate:"min=1s"`
    WriteTimeout time.Duration `yaml:"write_timeout" validate:"min=1s"`
    IdleTimeout  time.Duration `yaml:"idle_timeout" validate:"min=1s"`
    EnablePprof  bool          `yaml:"enable_pprof"`
}
```

**注意**：本提案仅定义配置结构，不涉及与config模块的集成，那将在后续独立工作中处理。

## 关键接口

```go
// Server对外接口
type Server interface {
    Start() error
    Stop(ctx context.Context) error
}

// 依赖接口（由业务模块提供）
type MetricsService interface {
    Render(ctx context.Context) (string, error)
}

type HealthService interface {
    Check(ctx context.Context) (status string, details map[string]any)
}
```

## What Changes

### 新增文件
- `internal/server/server.go` - HTTP服务器实现，包含Server接口和HTTPServer结构体
- `internal/server/config.go` - 服务器配置结构体定义和验证
- `internal/server/middleware.go` - Logger和Recovery中间件实现
- `internal/server/routes.go` - 路由注册和端点处理
- `internal/server/server_test.go` - 服务器单元测试
- `internal/server/integration_test.go` - 集成测试

### 新增接口
- `Server` 接口 - 定义Start/Stop方法
- `MetricsService` 接口 - 指标渲染服务接口
- `HealthService` 接口 - 健康检查服务接口

### 新增配置结构
- `Config` 结构体 - 包含端口、主机、超时等服务器配置参数

### 功能特性
- 基于Gin框架的HTTP服务器
- 支持/metrics端点用于Prometheus指标导出
- 支持/health端点用于健康检查
- 可选的pprof调试支持
- 优雅关闭机制
- 结构化日志记录

## 实现计划

1. **配置定义**：实现Config结构体和验证逻辑
2. **接口定义**：定义Server、MetricsService、HealthService接口
3. **HTTP服务器实现**：基于Gin的HTTPServer结构体
4. **中间件实现**：Logger和Recovery中间件
5. **路由注册**：/health和/metrics端点
6. **优雅关闭**：实现Start/Stop方法
7. **测试覆盖**：单元测试和集成测试

## 参考资料

- `docs/design/server.md` - HTTP服务器模块设计文档
- `docs/design/architecture.md` - 系统架构设计
- `openspec/project.md` - 项目上下文和约定