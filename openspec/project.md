# Project Context

## Purpose
WinPower G2 Prometheus Exporter 用于从 WinPower 管理系统通过 HTTP/HTTPS 采集设备实时指标，计算累计电能，并以 Prometheus 格式在 `/metrics` 暴露，供监控系统抓取。
目标：简单、边界清晰、易测试、易运维。

## Tech Stack
- Go 1.25+
- Gin（HTTP 服务器）
- zap（结构化日志）
- HTTP/HTTPS 客户端
- Prometheus 指标导出
- Docker/Kubernetes（部署，可选）

## Project Conventions

### Code Style
- 使用 `go fmt` 格式化，遵循 Go 规范与 idiomatic 风格。
- 静态分析使用 `make lint`；统一结构化日志采用 `zap`。所有提案完成后都需要保证通过静态检查。
- 测试文件命名为 `*_test.go`，与实现同包同目录。
- 统一配置前缀 `WINPOWER_EXPORTER_`；CLI 与环境变量可覆盖。
- 错误信息脱敏，不打印凭据和 Token 等敏感内容。

### Architecture Patterns
- 模块化分层：config → logging → storage → auth → energy → collector → metrics → server → scheduler。
- 接口抽象与边界清晰，关键模块可替换（如 Storage）。
- 调度器与 `/metrics` 均通过同一入口 `Collector.CollectDeviceData` 触发采集与能耗累计；`/metrics` 在执行后返回最新数据。
- 仅累计电能持久化到设备级文件，其余指标不存历史数据。
- 可选开启 `/debug/pprof` 进行诊断；统一日志与指标入口便于观测。

### Testing Strategy
- 采用 TDD；在编写提案时需要编写测试用例设计（包含边界场景）。
- 在提案实现过程中，优先编写定义接口和单元测试，再实现功能代码。
- 测试与实现同包，使用 `*_test.go` 组织。
- 使用 Mock 隔离外部依赖（例如 TokenProvider、EnergyStore）。
- 常用命令：`make test`、`make test-coverage`、`make test-integration`、`make test-all`。
- 性能诊断可使用 `/debug/pprof`。

### Git Workflow
- 主分支为 `main`；新特性通过 `feature/<name>` 分支迭代；小步提交。
- 提交信息使用祈使句；所有 PR 需通过测试并完成代码审查。
- 新能力/架构/破坏性变更先在 `openspec/changes` 创建 proposal 并通过审核后实施。
- 部署后归档变更至 `openspec/changes/archive/YYYY-MM-DD-<id>/`。

## Domain Context
- WinPower G2 使用 Token 认证：`POST /api/v1/auth/login` 返回 `deviceId` 与 `token`。
- Token 固定有效期约 1 小时；无需解析 Token 内容；到期前 5–10 分钟刷新。
- 能耗计算：累计电能 `Wh = W × 时间(h)`。
- 典型设备类型：UPS 等；指标标签包含 `winpower_host`、`device_id`、`device_name`、`device_type`、`phase` 等。
- `/metrics` 在处理请求时通过统一入口触发一次采集与能耗累计，然后返回最新注册指标快照。

## Important Constraints
- 调度器固定 5s 周期，与 Prometheus 抓取间隔无关。
- 仅累计电能持久化，其余指标不存历史数据。
- 默认端口 9090；WinPower 默认端口 8081。
- 支持自签证书跳过 SSL 验证（开发/内网）；生产建议由反向代理终止 TLS。
- 考虑 WinPower API 容量与速率限制；必要时启用服务端限流。

## External Dependencies
- WinPower HTTP/HTTPS 服务（认证与设备数据 API）。
- Prometheus（抓取 `/metrics` 指标）。
- 反向代理/Nginx（生产环境 TLS 终止，可选）。
- 操作系统证书信任与 CA（HTTPS 校验）。
- Docker/Kubernetes 运行环境（部署，可选）。
