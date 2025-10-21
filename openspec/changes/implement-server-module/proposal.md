## Why
实现 HTTP 服务器模块以提供 `/metrics` 和 `/health` 端点，完成 WinPower G2 Prometheus Exporter 的核心网络服务层。该模块将基于设计文档中的简化版架构，提供高性能、高可靠性的 Web 接口。

## What Changes
- 实现基于 Gin 的 HTTP 服务器模块
- 提供 `/metrics` 端点用于 Prometheus 指标暴露
- 提供 `/health` 端点用于健康检查
- 实现完整的中间件链（日志、恢复、CORS、限流、指标）
- 支持可选的 `/debug/pprof` 诊断端点
- 实现优雅启动和关闭机制
- 定义模块配置结构并与 config 模块集成
- 提供完整的 TDD 测试覆盖

## Impact
- **Affected specs**: 无（新增模块）
- **Affected code**:
  - 新增 `internal/server/` 模块
  - 依赖 `logging` 模块提供 Logger
  - 依赖 `metrics` 模块的 MetricManagerInterface
  - 依赖后续可能实现的 health 检查服务
- **Breaking changes**: 无（纯新增功能）
- **Dependencies**: 需要安装 Gin 依赖到 go.mod