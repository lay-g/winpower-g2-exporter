<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# AGENTS.md

面向本项目的 AI 助手工作指引（对齐最新设计/实现/协议文档）。

## 项目上下文（最新）
- 目标：采集 WinPower 设备数据、计算能耗，并以 Prometheus 指标导出。
- 技术栈：Go 1.25+、Gin、zap、HTTP、Prometheus，测试驱动开发（TDD）。
- 模块与依赖顺序：config → logging → storage → auth → energy → collector → metrics → server → scheduler。
- 服务端点：`/metrics`（只读）、`/health`；`/debug/pprof` 可选开启用于诊断。

## 关键约束与行为
- 调度器统一周期 5s；与 Prometheus 抓取间隔无关，`/metrics` 不触发采集或计算。
- 仅累计电能持久化到设备级文件；其他指标不存历史。
- 支持自签证书：通过 `--skip-ssl-verify`（或环境变量）跳过 SSL 验证；生产建议由反向代理终止 TLS。
- 不要随意创建任何文档，除非明确要求的时候才需要

## 开发命令（Makefile）
```bash
make build              # 构建当前平台
make build-linux        # 构建 Linux AMD64
make build-all          # 构建所有支持平台
make clean              # 清理构建产物

make fmt                # 格式化代码
make lint               # 静态分析
make dev                # 本地开发环境
make deps               # 安装依赖
make update-deps        # 更新依赖

make test               # 单元测试
make test-coverage      # 测试覆盖率
make test-integration   # 集成测试
make test-all           # 运行所有测试

make docker-build       # 构建镜像
make docker-run         # 运行容器
make docker-push        # 推送镜像
```

## 配置（CLI 与环境变量）
- 必选参数：
  - `--winpower.url`：WinPower 服务地址（含协议与端口）
  - `--winpower.username`：用户名
  - `--winpower.password`：密码
- 可选参数：
  - `--port`（默认 9090）、`--log-level`（debug|info|warn|error，默认 info）
  - `--skip-ssl-verify`（默认 false，自签证书场景）
  - `--data-dir`（默认 `./data`）、`--sync-write`（默认 true）
- 环境变量（前缀 `WINPOWER_EXPORTER_`）：
  - `CONSOLE_URL`、`USERNAME`、`PASSWORD`、`PORT`、`LOG_LEVEL`
  - `SKIP_SSL_VERIFY`、`DATA_DIR`、`SYNC_WRITE`

## 指标结构（摘要）
- 命名：`winpower_exporter_*`（自监控）、`winpower_*`（连接/认证/设备/能耗）。
- 自监控：运行状态、HTTP 请求计数/时延、采集耗时、错误数、设备数、Token 刷新计数。
- 连接/认证：连接状态、认证状态、API 响应时延、Token 剩余有效期/有效性。
- 设备/电源：连接状态、负载%、输入/输出电压/电流/频率、有功功率、功率因数。
- 能耗：瞬时功率（W）、累计电能（Wh，允许负值表示净能量）。
- 标签：`winpower_host`、`version`、`status_code`、`method`、`error_type`、`status`、
  `device_id`、`device_name`、`device_type`、`phase`、`user_id`、`connection_type`、`auth_method`、`api_endpoint`。

## 能耗计算与调度
- 统一采集入口：调度器每 5 秒触发 Collector；解析成功后累计能耗并持久化。
- 只读暴露：`/metrics` 仅返回当前注册的指标，不触发采集或计算。

## 测试策略（TDD）
- 测试命令：`make test`、`make test-coverage`、`make test-integration`、`make test-all`。
- 组织约定：测试文件与实现同包、同目录，命名 `*_test.go`。
- Mock 建议：为 `TokenProvider`、`EnergyStore` 提供 Mock 隔离外部副作用。
- 观察性：按需开启 `/debug/pprof`；关键路径统一结构化日志。

## 参考文档
- 设计：`docs/design/architecture.md`、`docs/design/metrics.md`、`docs/design/auth.md`、`docs/design/server.md`。
- 实现：`docs/implements/overview.md` 及各模块文档。
- 协议：`docs/protocol/authentication.md`。

以上内容与 README、设计/实现/协议文档保持一致，作为日常协作与实现的最新依据。