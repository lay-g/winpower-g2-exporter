# WinPower G2 Prometheus Exporter

## 项目说明

WinPower G2 Prometheus Exporter 用于从 WinPower 管理系统实时采集设备指标，并以 Prometheus 兼容格式暴露给监控系统。

**说明**: 本项目仅负责数据采集与指标导出，不内置告警。累计电能按设备持久化到文件，其余指标为实时数据。

## 架构与特性

```
┌─────────────────┐    HTTP     ┌──────────────────┐    HTTP     ┌─────────────────┐
│   Prometheus    │ ◄────────── │   WinPower G2    │ ◄────────── │   WinPower 设备  │
│    服务器       │             │   Exporter       │             │   管理系统       │
└─────────────────┘             └──────────────────┘             └─────────────────┘
```

- 组件：HTTP 服务(`/metrics`/`/health`)、采集器、指标转换、配置管理、日志、能耗计算、文件持久化、定时任务。
- 采集：固定周期触发（采集每 5 秒、Token 刷新每 1 分钟）。
- 能耗：累计电能仅按设备持久化；瞬时功率按 `V × A × PF` 计算。
- 指标：遵循 Prometheus 命名与标签规范；只读暴露不触发采集。

## 技术栈

- `Go 1.23+`
- `Gin`（Web 框架）
- `zap`（结构化日志）
- `Viper`（配置管理）
- `Lumberjack`（日志轮转）
- `HTTP`（与 WinPower 通信）
- `Prometheus`（指标格式）

## 模块概览

- 配置管理（`internal/config`）：嵌套配置结构、Viper集成、热重载、多格式支持。
- 日志模块（`pkgs/log`）：上下文日志、轮转、多输出、结构化日志。
- 数据采集与指标转换：从 WinPower 拉取并转换为 Prometheus 指标。
- 电能计算与持久化：累计电能（设备级文件，原子写入）。
- 调度器：统一入口定时触发采集与 Token 刷新。

### ✅ 配置管理模块（`internal/config`）

基于 Viper 的企业级配置管理系统：

**功能特性：**
- 嵌套配置结构（Server、Log、Storage、WinPower）
- 多源配置加载（CLI参数、环境变量、配置文件）
- 配置热重载（运行时配置更新）
- 多格式支持（YAML、JSON、TOML）
- 自动配置迁移（从扁平格式到嵌套格式）
- 完整的配置验证和错误报告

**配置加载优先级：**
1. 命令行参数（最高优先级）
2. 环境变量（WINPOWER_EXPORTER_* 前缀）
3. 配置文件
4. 默认值（最低优先级）

### ✅ 日志模块（`pkgs/log`）

基于 Zap 的高性能结构化日志系统：

**功能特性：**
- 上下文感知日志（自动上下文提取和传播）
- 日志轮转（基于大小和时间的自动轮转）
- 多输出目标（stdout、stderr、文件、同时输出）
- 结构化日志（JSON 和控制台格式）
- 全局日志函数和测试专用日志器
- 开发和生产模式优化

## 配置与启动

### 加载优先级
1. 命令行参数（最高）
2. 环境变量（前缀 `WINPOWER_EXPORTER_*`）
3. 配置文件（YAML）
4. 内置默认值（最低）

### 配置文件示例

将 `config.example.yaml` 复制为 `config.yaml` 并按需修改：

```yaml
# 新的嵌套配置格式（推荐）
server:
  port: 9090
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  enable_pprof: false

log:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true
  enable_caller: false
  enable_stacktrace: false

winpower:
  url: "https://winpower.example.com:8443"
  username: "admin"
  password: "secret123"
  timeout: 30s
  skip_ssl_verify: false

storage:
  data_dir: "./data"
  file_permissions: "0644"
  dir_permissions: "0755"
  sync_write: true
  create_dir: true
```

**向后兼容的扁平格式**：
```yaml
port: 9090
log_level: info
data_dir: ./data
skip_ssl_verify: false
sync_write: true

winpower:
  url: https://winpower.example.com:8443
  username: admin
  password: secret123
```

启动：
```bash
./winpower-g2-exporter --config=config.yaml
```

### 命令行参数

- 必需：`--winpower.url`、`--winpower.username`、`--winpower.password`
- 可选：`--port`、`--log-level`、`--skip-ssl-verify`、`--data-dir`、`--sync-write`

说明：采集固定为每 5 秒，与 Prometheus 抓取间隔无关。

### 环境变量启动

**传统环境变量（向后兼容）**：
```bash
export WINPOWER_EXPORTER_CONSOLE_URL="https://192.168.1.100:8081"
export WINPOWER_EXPORTER_USERNAME="admin"
export WINPOWER_EXPORTER_PASSWORD="password"
export WINPOWER_EXPORTER_PORT="9090"
export WINPOWER_EXPORTER_LOG_LEVEL="info"
export WINPOWER_EXPORTER_SKIP_SSL_VERIFY="true"
export WINPOWER_EXPORTER_DATA_DIR="./data"
export WINPOWER_EXPORTER_SYNC_WRITE="true"

./winpower-g2-exporter
```

**嵌套环境变量（新功能）**：
```bash
export WINPOWER_EXPORTER_SERVER_PORT="9090"
export WINPOWER_EXPORTER_SERVER_HOST="0.0.0.0"
export WINPOWER_EXPORTER_LOG_LEVEL="info"
export WINPOWER_EXPORTER_LOG_FORMAT="json"
export WINPOWER_EXPORTER_LOG_OUTPUT="stdout"
export WINPOWER_EXPORTER_LOG_MAX_SIZE="100"
export WINPOWER_EXPORTER_STORAGE_DATA_DIR="./data"
export WINPOWER_EXPORTER_STORAGE_SYNC_WRITE="true"
export WINPOWER_EXPORTER_WINPOWER_URL="https://192.168.1.100:8081"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="admin"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="password"
export WINPOWER_EXPORTER_WINPOWER_TIMEOUT="30s"

./winpower-g2-exporter
```

**环境变量映射**：
- `WINPOWER_EXPORTER_CONSOLE_URL` → `winpower.url`
- `WINPOWER_EXPORTER_USERNAME` → `winpower.username`
- `WINPOWER_EXPORTER_PASSWORD` → `winpower.password`
- 其余扁平字段直接映射（如 `WINPOWER_EXPORTER_PORT`、`WINPOWER_EXPORTER_LOG_LEVEL`）

## 指标与能耗

### 指标分类

- Exporter 自监控：运行状态、HTTP 请求与时延、采集耗时/错误、设备数量、Token 刷新计数。
- WinPower 连接/认证：连接状态、认证状态、API 时延、Token 剩余有效期/有效性。
- 设备/电源：连接状态、负载%、电压/电流/频率、有功功率、功率因数。
- 能耗：瞬时功率（W）、累计电能（Wh；允许负值表示净能量）。

### 能耗计算与持久化

- 触发：调度器每 5 秒调用统一采集；成功解析后累计并持久化。
- 只读：`/metrics` 仅返回当前注册指标，不触发采集或计算。
- 文件：每设备独立文件，命名 `{device_id}.txt`，存放于 `data_dir`。
- 文件内容：
  - 行1：最后更新时间（毫秒时间戳）
  - 行2：累计电能（Wh）
- 计算公式：
  - 功率：`W = V × A × PF`
  - 电能：`Wh = W × 间隔(h)`
  - 示例：`1000W × (5s / 3600s/h) ≈ 1.39Wh`

### 指标命名示例

```
# Exporter 自监控
winpower_exporter_up{winpower_host="xxx", version="1.0.0"} 1
winpower_exporter_requests_total{winpower_host="xxx", status_code="200", method="GET"} 1000
winpower_exporter_request_duration_seconds{winpower_host="xxx", status_code="200", method="GET"} 0.005
winpower_exporter_scrape_errors_total{winpower_host="xxx", error_type="parse"} 5
winpower_exporter_collection_duration_seconds{winpower_host="xxx", status="ok"} 0.100
winpower_exporter_device_count{winpower_host="xxx", device_type="ups"} 12
winpower_exporter_token_refresh_total{winpower_host="xxx", result="ok"} 3

# WinPower 连接/认证
winpower_connection_status{winpower_host="xxx", connection_type="api"} 1
winpower_auth_status{winpower_host="xxx", auth_method="password"} 1
winpower_api_response_time_seconds{winpower_host="xxx", api_endpoint="/api/v1/device/detail/list"} 0.100
winpower_token_expiry_seconds{winpower_host="xxx", user_id="admin"} 3600
winpower_token_valid{winpower_host="xxx", user_id="admin"} 1

# 设备/电源（phase 可选）
winpower_device_connected{device_id="ups-01", device_name="UPS 主机", device_type="ups"} 1
winpower_load_percent{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 45.0
winpower_input_voltage{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 220.5
winpower_output_voltage{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 230.1
winpower_input_current{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 10.2
winpower_output_current{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 9.8
winpower_input_frequency{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 50.0
winpower_output_frequency{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 50.0
winpower_input_watts{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 2000.0
winpower_output_watts{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 1900.0
winpower_output_power_factor{device_id="ups-01", device_name="UPS 主机", device_type="ups", phase="A"} 0.95

# 能耗（仅累计 + 瞬时功率）
winpower_power_watts{device_id="ups-01", device_name="UPS 主机", device_type="ups"} 2240.0
winpower_energy_total_wh{device_id="ups-01", device_name="UPS 主机", device_type="ups"} 15000.0
```

标签说明：
- `winpower_host`：WinPower 服务器地址（Exporter/连接类指标）
- `version`：Exporter 版本（自监控）
- `status_code`、`method`：HTTP 请求状态与方法（自监控）
- `error_type`：采集错误类型（自监控）
- `status`：采集状态（ok/err，自监控）
- `device_id`、`device_name`、`device_type`：设备标识（设备/电源/能耗）
- `phase`：相线（A/B/C，按需使用）
- `user_id`：认证用户标识（Token 指标）
- `connection_type`：连接类型（如 api）
- `auth_method`：认证方式（如 password）
- `api_endpoint`：API 端点（连接/响应时延观测）

## 部署

### Docker 部署

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o winpower-g2-exporter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/winpower-g2-exporter .
EXPOSE 9090
CMD ["./winpower-g2-exporter"]
```

```bash
# 构建镜像
docker build -t winpower-g2-exporter .

# 运行容器
docker run -d \
  -p 9090:9090 \
  -e WINPOWER_EXPORTER_CONSOLE_URL="https://winpower-server:8081" \
  -e WINPOWER_EXPORTER_USERNAME="admin" \
  -e WINPOWER_EXPORTER_PASSWORD="password" \
  -e WINPOWER_EXPORTER_SKIP_SSL_VERIFY="true" \
  winpower-g2-exporter
```

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: winpower-g2-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: winpower-g2-exporter
  template:
    metadata:
      labels:
        app: winpower-g2-exporter
    spec:
      containers:
      - name: exporter
        image: winpower-g2-exporter:latest
        ports:
        - containerPort: 9090
        env:
        - name: WINPOWER_EXPORTER_CONSOLE_URL
          value: "https://winpower-server:8081"
        - name: WINPOWER_EXPORTER_USERNAME
          value: "admin"
        - name: WINPOWER_EXPORTER_PASSWORD
          valueFrom:
            secretKeyRef:
              name: winpower-secrets
              key: password
        - name: WINPOWER_EXPORTER_SKIP_SSL_VERIFY
          value: "true"
        - name: WINPOWER_EXPORTER_LOG_LEVEL
          value: "info"
        - name: WINPOWER_EXPORTER_DATA_DIR
          value: "/var/lib/winpower-exporter/data"
---
apiVersion: v1
kind: Service
metadata:
  name: winpower-g2-exporter
spec:
  selector:
    app: winpower-g2-exporter
  ports:
  - port: 9090
    targetPort: 9090
```

## Prometheus 配置

在 Prometheus 配置文件中添加以下抓取任务：

```yaml
scrape_configs:
  - job_name: 'winpower-g2-exporter'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    scrape_timeout: 10s
```

## 监控与可视化

### Grafana 仪表板

建议创建以下 Grafana 仪表板来监控 WinPower 设备：

- 设备状态概览
- 电源使用情况
- 环境监控
- 系统性能

**注意**: 本项目不提供告警功能。告警规则应在 Prometheus 中配置，或者使用 Grafana 的告警功能。

## 开发与构建

### 环境要求
- Go 1.25+
- Make（构建与测试管理）
- Git（版本控制）

### 依赖与常用命令

```bash
# 构建
make build            # 当前平台主程序
make build-linux      # Linux AMD64 主程序
make build-all        # 所有支持平台主程序
make build-tools      # 构建工具（配置迁移工具）
make build-all-tools  # 构建主程序和所有工具
make clean            # 清理

# 配置迁移
./build/winpower-config-migrate migrate config.yaml      # 迁移配置
./build/winpower-config-migrate validate config.yaml     # 验证配置
./build/winpower-config-migrate help                       # 迁移工具帮助

# 测试
make test             # 单元测试
make test-coverage    # 覆盖率报告
make test-integration # 集成测试
make test-all         # 全部测试

# 开发工具
make fmt              # 格式化
make lint             # 静态分析
make dev              # 开发环境
make deps             # 安装依赖
make update-deps      # 更新依赖

# Docker
make docker-build     # 构建镜像
make docker-run       # 运行容器
make docker-push      # 推送镜像

# 发布
make release          # 创建发布版本
make tag              # 创建 Git 标签
```

### 配置迁移工具

v0.1.0 包含配置迁移工具，帮助从旧的扁平配置格式迁移到新的嵌套格式：

```bash
# 构建迁移工具
make build-tools

# 自动迁移配置（自动创建备份）
./build/winpower-config-migrate migrate config.yaml

# 验证迁移后的配置
./build/winpower-config-migrate validate config.yaml

# 查看迁移工具帮助
./build/winpower-config-migrate help
```

**迁移特性**：
- 自动检测配置格式
- 安全备份原始配置
- 零停机迁移
- 完整的配置验证
- 回滚支持

## 故障排除

### 常见问题

1. **连接失败**: 检查 WinPower 服务地址和网络连通性
2. **认证错误**: 验证用户名和密码是否正确
3. **指标为空**: 确认 WinPower API 返回数据格式正确

### 日志查看

```bash
# 启用调试日志
./winpower-g2-exporter --log-level=debug

# 查看特定错误
grep "ERROR" /var/log/winpower-exporter.log
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

如有问题或建议，请通过以下方式联系：

- 提交 Issue: [GitHub Issues](https://github.com/your-username/winpower-g2-exporter/issues)
- 邮箱: your-email@example.com

---

注：文档已统一去重并按“概览 → 配置 → 指标 → 部署 → 开发 → 维护”组织，保持简洁一致。
- 支持多配置源（CLI、环境变量、YAML 文件）
- 配置优先级控制
- 完整的配置验证
- 类型安全的配置访问
- 默认值支持

**核心文件：**
- `config.go`: 配置结构定义和加载逻辑
- `file.go`: YAML 文件加载和保存
- `validation.go`: 配置验证规则
- `errors.go`: 错误类型定义

**测试覆盖：** 67.9% (单元测试)

### ✅ 日志模块 (`pkgs/log`)

基于 Zap 的高性能结构化日志系统：

**功能特性：**
- 统一的日志接口
- 结构化字段支持
- 可配置的日志级别 (debug, info, warn, error)
- 生产模式（JSON 格式）和开发模式（人类可读）
- 子 Logger 支持（With 方法）

**核心文件：**
- `interface.go`: Logger 接口和字段构造函数
- `logger.go`: Zap 实现
- `init.go`: 从配置初始化

**测试覆盖：** 93.7% (单元测试)

### ✅ 集成测试 (`tests/integration`)

端到端集成测试验证配置和日志模块协同工作：

**测试场景：**
- 配置加载和日志初始化
- 配置文件加载和日志集成
- 环境变量配置和日志集成
- 开发模式日志
- 子 Logger 上下文传递
- 错误场景处理

**测试覆盖：** 所有集成场景均已验证

## 使用示例

### 基本用法

```go
package main

import (
    "github.com/lay-g/winpower-g2-exporter/internal/config"
    "github.com/lay-g/winpower-g2-exporter/pkgs/log"
)

func main() {
    // 加载配置
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }

    // 验证配置
    if err := cfg.Validate(); err != nil {
        panic(err)
    }

    // 初始化日志
    logger, err := log.Init(cfg)
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // 使用日志
    logger.Info("Application started",
        log.Int("port", cfg.Port),
        log.String("log_level", cfg.LogLevel),
    )

    // 创建带上下文的子 logger
    httpLogger := logger.With(
        log.String("component", "http_server"),
    )
    httpLogger.Info("Server listening")
}
```

### 配置文件使用

```bash
# 创建配置文件
cp config.example.yaml config.yaml

# 编辑配置
vim config.yaml

# 使用配置文件启动
./winpower-g2-exporter --config=config.yaml
```

### 环境变量使用

```bash
# 设置环境变量
export WINPOWER_EXPORTER_PORT=9090
export WINPOWER_EXPORTER_LOG_LEVEL=debug
export WINPOWER_EXPORTER_CONSOLE_URL=https://example.com
export WINPOWER_EXPORTER_USERNAME=admin
export WINPOWER_EXPORTER_PASSWORD=secret

# 启动（自动读取环境变量）
./winpower-g2-exporter
```

## 目录结构

## TDD 测试驱动开发

项目严格遵循 TDD（测试驱动开发）方法论：

### 开发流程
1. **红阶段**: 编写失败的测试用例
2. **绿阶段**: 编写最少代码使测试通过
3. **重构阶段**: 改进代码质量，保持测试通过

### 测试策略
- **单元测试**: 每个模块和函数都有对应的单元测试
- **集成测试**: 模块间交互的集成测试
- **端到端测试**: 完整功能流程的测试
- **测试覆盖率**: 不低于 80% 的代码覆盖率

### 测试组织
- **同包测试**: 测试代码与生产代码在同一包内
- **测试文件**: 使用 `*_test.go` 命名规范
- **Mock 和 Stub**: 使用测试替身隔离外部依赖
- **测试数据**: 使用表驱动测试提供多组测试数据

### 质量保证
- **持续集成**: 每次提交都运行完整测试套件
- **代码覆盖率**: 使用 `make test-coverage` 生成覆盖率报告
- **静态分析**: 使用 `make lint` 进行代码质量检查
- **性能测试**: 关键路径的性能基准测试

## 重要说明

### 功能范围
- ✅ **数据抓取**: 从 WinPower 系统收集设备指标
- ✅ **指标导出**: 以 Prometheus 格式提供指标数据
- ✅ **电能计算**: 定时计算并持久化电能消耗
- ✅ **实时采集**: 直接调用 WinPower API，无缓存延迟
- ❌ **告警功能**: 不提供告警功能，告警由 Prometheus/Grafana 处理

### 数据处理
- **实时性**: 所有指标数据都通过直接调用 WinPower API 实时获取
- **电能持久化**: 仅电能计算结果保存到文件，确保累计值不丢失
- **无历史存储**: 其他指标不存储历史数据，减少内存占用

### 日志配置
项目使用 zap 作为高性能日志库，支持结构化日志输出。通过 `--log-level` 参数控制日志详细程度。

---

**注意**: 请确保在生产环境中使用强密码和 HTTPS 连接以保护 WinPower 系统安全。