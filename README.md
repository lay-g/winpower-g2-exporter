# WinPower G2 Exporter

[![CI/CD Pipeline](https://github.com/lay-g/winpower-g2-exporter/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/lay-g/winpower-g2-exporter/actions/workflows/ci.yml)

一个用于采集 WinPower G2 设备数据、计算能耗并以 Prometheus 指标格式导出的 Go 应用程序。

## 概述

WinPower G2 Exporter 通过定时采集 WinPower G2 系统的设备数据，提供实时的设备状态监控和能耗统计功能。该导出器采用模块化设计，支持多设备管理，能够可靠地处理设备连接、数据采集、电能计算和指标暴露。

### 核心特性

- **统一数据采集**：每 5 秒自动采集 WinPower 设备数据
- **能耗计算**：基于功率数据进行精确的电能累计计算
- **Prometheus 兼容**：提供标准的 `/metrics` 端点供 Prometheus 抓取
- **多设备支持**：支持同时监控多个 WinPower 设备
- **自认证管理**：自动处理 Token 刷新和认证流程
- **结构化日志**：基于 zap 的高性能结构化日志记录
- **TDD 开发**：测试驱动开发，确保代码质量

## 架构设计

### 模块架构

```
┌─────────────────────────────────────────────────────────────┐
│                    WinPower G2 Exporter                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │   Config    │  │   Logging   │  │      Storage         │   │
│  │  配置管理    │  │   日志模块   │  │    存储模块          │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │  WinPower   │  │  Collector  │  │       Energy         │   │
│  │ 数据采集模块  │  │ 数据协调模块 │  │    电能计算模块      │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │   Metrics   │  │   Server    │  │     Scheduler       │   │
│  │  指标管理    │  │  HTTP服务   │  │    调度器模块        │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │  WinPower G2    │
                    │     系统        │
                    └─────────────────┘
```

### 数据流程

1. **调度器**：每 5 秒触发数据采集
2. **WinPower 模块**：从 WinPower 系统获取设备数据
3. **Collector 模块**：协调数据采集和电能计算
4. **Energy 模块**：基于功率数据计算累计电能
5. **Storage 模块**：持久化电能累计值
6. **Metrics 模块**：更新 Prometheus 指标
7. **Server 模块**：通过 HTTP 端点暴露指标

## 快速开始

### 环境要求

- Go 1.25+
- WinPower G2 系统
- Prometheus (可选，用于指标收集)

### 安装

#### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/lay-g/winpower-g2-exporter.git
cd winpower-g2-exporter

# 构建
make build

# 运行
./bin/winpower-g2-exporter server
```

#### 使用 Docker

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run
```

### 基本配置

创建配置文件 `config.yaml`：

```yaml
server:
  port: 9090
  host: "0.0.0.0"

winpower:
  base_url: "https://your-winpower-server.com"
  username: "admin"
  password: "your-password"
  timeout: 30s
  skip_ssl_verify: false

storage:
  data_dir: "./data"
  sync_write: true

scheduler:
  collection_interval: 5s

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### 运行服务

```bash
# 使用配置文件
./winpower-g2-exporter server --config config.yaml

# 或使用环境变量
export WINPOWER_EXPORTER_WINPOWER_BASE_URL="https://your-winpower-server.com"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="admin"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="your-password"
./winpower-g2-exporter server
```

### 验证运行

访问健康检查端点：
```bash
curl http://localhost:9090/health
```

访问指标端点：
```bash
curl http://localhost:9090/metrics
```

## 配置说明

### 配置文件

WinPower G2 Exporter 支持多种配置方式：

| 配置方式   | 优先级 | 说明                      |
| ---------- | ------ | ------------------------- |
| 命令行参数 | 最高   | 通过 `--` 前缀的参数      |
| 环境变量   | 中等   | `WINPOWER_EXPORTER_` 前缀 |
| 配置文件   | 最低   | YAML 格式配置文件         |

### 必需配置

| 参数                  | 环境变量                              | 说明              | 示例                           |
| --------------------- | ------------------------------------- | ----------------- | ------------------------------ |
| `--winpower.url`      | `WINPOWER_EXPORTER_WINPOWER_URL`      | WinPower 服务地址 | `https://winpower.example.com` |
| `--winpower.username` | `WINPOWER_EXPORTER_WINPOWER_USERNAME` | 用户名            | `admin`                        |
| `--winpower.password` | `WINPOWER_EXPORTER_WINPOWER_PASSWORD` | 密码              | `secret`                       |

### 可选配置

| 参数                | 环境变量                               | 默认值   | 说明              |
| ------------------- | -------------------------------------- | -------- | ----------------- |
| `--port`            | `WINPOWER_EXPORTER_PORT`               | `9090`   | HTTP 服务端口     |
| `--log-level`       | `WINPOWER_EXPORTER_LOGGING_LEVEL`      | `info`   | 日志级别 (debug   | info | warn | error) |
| `--skip-ssl-verify` | `WINPOWER_EXPORTER_SKIP_SSL_VERIFY`    | `false`  | 跳过 SSL 证书验证 |
| `--data-dir`        | `WINPOWER_EXPORTER_STORAGE_DATA_DIR`   | `./data` | 数据存储目录      |
| `--sync-write`      | `WINPOWER_EXPORTER_STORAGE_SYNC_WRITE` | `true`   | 同步写入模式      |

### 完整配置示例

```yaml
# config.yaml
server:
  port: 9090
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  enable_pprof: false

winpower:
  base_url: "https://winpower.example.com"
  username: "admin"
  password: "secret"
  timeout: 30s
  api_timeout: 10s
  skip_ssl_verify: false
  refresh_threshold: 5m

storage:
  data_dir: "./data"
  file_permissions: 0644

scheduler:
  collection_interval: 5s
  graceful_shutdown_timeout: 5s

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: 100
  max_age: 7
  max_backups: 3
  compress: true
  enable_caller: false
  enable_stacktrace: false
```

## 指标说明

### Exporter 自监控指标

| 指标名称                                        | 类型      | 描述              |
| ----------------------------------------------- | --------- | ----------------- |
| `winpower_exporter_up`                          | Gauge     | Exporter 运行状态 |
| `winpower_exporter_requests_total`              | Counter   | HTTP 请求总数     |
| `winpower_exporter_collection_duration_seconds` | Histogram | 采集+计算整体耗时 |
| `winpower_exporter_scrape_errors_total`         | Counter   | 采集错误总数      |
| `winpower_exporter_device_count`                | Gauge     | 发现的设备数量    |

### WinPower 连接指标

| 指标名称                             | 类型      | 描述              |
| ------------------------------------ | --------- | ----------------- |
| `winpower_connection_status`         | Gauge     | WinPower 连接状态 |
| `winpower_auth_status`               | Gauge     | 认证状态          |
| `winpower_api_response_time_seconds` | Histogram | API 响应时延      |
| `winpower_token_expiry_seconds`      | Gauge     | Token 剩余有效期  |

### 设备状态指标

| 指标名称                            | 类型  | 描述              | 标签                                |
| ----------------------------------- | ----- | ----------------- | ----------------------------------- |
| `winpower_device_connected`         | Gauge | 设备连接状态      | device_id, device_name, device_type |
| `winpower_device_load_percent`      | Gauge | 设备负载百分比    | device_id, device_name, device_type |
| `winpower_device_load_total_watts`  | Gauge | 总负载有功功率(W) | device_id, device_name, device_type |
| `winpower_device_cumulative_energy` | Gauge | 累计电能(Wh)      | device_id, device_name, device_type |

### 电气参数指标

| 指标名称                           | 类型  | 描述         |
| ---------------------------------- | ----- | ------------ |
| `winpower_device_input_voltage`    | Gauge | 输入电压(V)  |
| `winpower_device_output_voltage`   | Gauge | 输出电压(V)  |
| `winpower_device_output_current`   | Gauge | 输出电流(A)  |
| `winpower_device_input_frequency`  | Gauge | 输入频率(Hz) |
| `winpower_device_output_frequency` | Gauge | 输出频率(Hz) |

### 电池参数指标

| 指标名称                                 | 类型  | 描述             |
| ---------------------------------------- | ----- | ---------------- |
| `winpower_device_battery_charging`       | Gauge | 电池充电状态     |
| `winpower_device_battery_capacity`       | Gauge | 电池容量(%)      |
| `winpower_device_battery_remain_seconds` | Gauge | 电池剩余时间(秒) |
| `winpower_device_ups_temperature`        | Gauge | UPS 温度(°C)     |

## 部署指南

### Docker 部署

```dockerfile
# 使用预构建镜像
docker run -d \
  --name winpower-exporter \
  -p 9090:9090 \
  -e WINPOWER_EXPORTER_WINPOWER_BASE_URL="https://winpower.example.com" \
  -e WINPOWER_EXPORTER_WINPOWER_USERNAME="admin" \
  -e WINPOWER_EXPORTER_WINPOWER_PASSWORD="secret" \
  -v $(pwd)/data:/app/data \
  winpower-g2-exporter:latest
```

### Docker Compose 部署

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  winpower-exporter:
    image: winpower-g2-exporter:latest
    container_name: winpower-exporter
    restart: unless-stopped
    ports:
      - "9090:9090"
    environment:
      - WINPOWER_EXPORTER_WINPOWER_BASE_URL=https://winpower.example.com
      - WINPOWER_EXPORTER_WINPOWER_USERNAME=admin
      - WINPOWER_EXPORTER_WINPOWER_PASSWORD=${WINPOWER_PASSWORD}
      - WINPOWER_EXPORTER_LOGGING_LEVEL=info
      - WINPOWER_EXPORTER_LOGGING_FORMAT=json
    volumes:
      - ./data:/app/data
      - ./config.yaml:/app/config.yaml:ro
```

创建 `.env` 文件用于敏感信息：

```bash
# .env
WINPOWER_PASSWORD=your-secret-password
```

启动服务：

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f winpower-exporter

# 停止服务
docker-compose down

# 重新构建并启动
docker-compose up -d --build
```

访问服务：

- **指标端点**: http://localhost:9090/metrics
- **健康检查**: http://localhost:9090/health

### Grafana 仪表板

推荐的 Grafana 查询：

```promql
# 设备总负载功率
winpower_device_load_total_watts

# 累计能耗
winpower_device_cumulative_energy

# 每小时能耗
increase(winpower_device_cumulative_energy[1h]) / 1000

# 设备连接状态
winpower_device_connected == 1

# 5分钟平均功率
avg_over_time(winpower_device_load_total_watts[5m])

# 系统健康状态
winpower_exporter_up == 1
```

## 开发指南

### 项目结构

```
winpower-g2-exporter/
├── cmd/                         # 应用程序入口点
│   └── winpower-g2-exporter/
│       └── main.go               # 主程序入口
├── internal/                    # 内部实现包
│   ├── cmd/                     # 命令行工具实现
│   ├── pkgs/                    # 内部公共包
│   │   └── log/                 # 日志模块
│   ├── config/                  # 配置管理模块
│   ├── storage/                 # 存储模块
│   ├── winpower/                # WinPower模块
│   ├── collector/               # Collector模块
│   ├── energy/                  # 电能计算模块
│   ├── metrics/                 # 指标模块
│   ├── server/                  # HTTP服务模块
│   └── scheduler/               # 调度器模块
├── docs/                        # 项目文档
│   ├── design/                  # 设计文档
│   └── examples/                # 使用示例
├── tests/                       # 测试文件
│   ├── integration/             # 集成测试
│   └── mocks/                   # Mock对象
├── deployments/                 # 部署配置
├── Dockerfile                   # Docker构建文件
├── Makefile                     # 构建和开发脚本
├── go.mod                       # Go模块定义
└── README.md                    # 项目说明文档
```

### 构建命令

```bash
# 构建当前平台
make build

# 构建 Linux AMD64
make build-linux

# 构建所有支持平台
make build-all

# 清理构建产物
make clean

# 格式化代码
make fmt

# 静态分析
make lint

# 安装依赖
make deps

# 更新依赖
make update-deps

# 运行测试
make test

# 测试覆盖率
make test-coverage

# 集成测试
make test-integration

# 运行所有测试
make test-all

# Docker 构建
make docker-build

# Docker 运行
make docker-run
```

### 测试

```bash
# 运行所有测试
make test

# 运行集成测试
make test-integration

# 生成测试覆盖率报告
make test-coverage

# 运行基准测试
go test -bench=. ./...
```

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行静态检查
- 编写单元测试，保持 80% 以上的测试覆盖率
- 遵循测试驱动开发(TDD)原则

## 故障排查

### 常见问题

#### 1. 连接 WinPower 失败

**症状**：日志显示认证失败或连接超时

**解决方案**：
- 检查 WinPower 服务地址是否正确
- 验证用户名和密码是否正确
- 确认网络连接和防火墙设置
- 如果使用自签名证书，设置 `skip_ssl_verify: true`

#### 2. 设备数据为空

**症状**：指标端点返回数据但设备指标为空

**解决方案**：
- 检查 WinPower 系统中是否有设备在线
- 验证用户权限是否足够访问设备数据
- 查看 WinPower API 响应是否正常

#### 3. 电能计算异常

**症状**：累计电能值不更新或计算错误

**解决方案**：
- 检查数据目录权限
- 验证功率数据是否正常获取
- 查看 Energy 模块日志是否有错误

### 日志分析

```bash
# 查看错误日志
jq 'select(.level == "error")' /var/log/winpower-exporter.log

# 查看采集耗时
jq 'select(.msg == "Device data collected") | .duration_ms' /var/log/winpower-exporter.log

# 查看认证状态
jq 'select(.msg | contains("Token"))' /var/log/winpower-exporter.log
```

### 健康检查

```bash
# 检查服务状态
curl -f http://localhost:9090/health || echo "服务异常"

# 检查指标端点
curl -s http://localhost:9090/metrics | head -n 10

# 检查进程状态
ps aux | grep winpower-exporter
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 提交规范

- 使用清晰的提交信息
- 一个提交只做一件事
- 包含相关的测试用例
- 遵循现有的代码风格

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 支持

- 文档：[项目文档](docs/)
- 问题反馈：[GitHub Issues](https://github.com/lay-g/winpower-g2-exporter/issues)
- 讨论：[GitHub Discussions](https://github.com/lay-g/winpower-g2-exporter/discussions)