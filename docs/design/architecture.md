# 简化版整体架构设计（与模块文档对齐）

## 1. 概述

WinPower G2 Exporter 的目标是以最小依赖、清晰边界的方式采集设备数据、计算电能并以 Prometheus 指标形式导出。本文档在“简单、可维护、易测试”的原则下，重新梳理模块划分与依赖关系。

## 2. 设计原则

- 简单优先：减少不必要的功能与耦合，默认关闭可选特性。
- 明确边界：每个模块只对外暴露少量、稳定的接口。
- 可替换：关键模块通过接口抽象，易于替换实现（如 Storage）。
- 易测试：面向接口编程，模块均可用 Mock 隔离测试。
- 可观察：统一日志与指标采集入口，方便排障与性能分析。

## 3. 项目目录结构

项目采用标准的 Go 项目布局，按照功能模块清晰组织，明确区分公共API和内部实现：

```
winpower-g2-exporter/
├── cmd/                         # 应用程序入口点
│   ├── exporter/                # 主程序（HTTP服务器）
├── internal/                    # 内部实现包（不对外暴露）
│   ├── cmd/                     # 命令行工具实现
│   ├── pkgs/                    # 内部公共包
│   │   └── log/                 # 日志模块
│   ├── config/                  # 配置管理模块实现
│   ├── storage/                 # 存储模块实现
│   ├── auth/                    # 认证模块实现
│   ├── collector/               # 采集模块实现
│   ├── energy/                  # 电能计算模块实现
│   ├── metrics/                 # 指标模块实现
│   ├── server/                  # HTTP服务模块实现(含路由和中间件)
│   └── scheduler/               # 调度器模块实现
├── docs/                        # 项目文档
│   ├── design/                  # 设计文档
│   ├── protocol/                # 协议文档
│   └── examples/                # 使用示例
├── tests/                       # 测试文件
│   ├── integration/             # 集成测试
│   ├── fixtures/                # 测试数据
│   └── mocks/                   # Mock对象
├── openspec/                    # OpenSpec 规范（可选）
├── build/                       # 构建产物
│   ├── bin/                     # 编译后的二进制文件
│   └── dist/                    # 分发包
├── scripts/                     # 构建和部署脚本
├── deployments/                 # 部署配置
├── Dockerfile                   # Docker构建文件
├── Makefile                     # 构建和开发脚本
├── go.mod                       # Go模块定义
├── go.sum                       # Go模块校验和
├── README.md                    # 项目说明文档
├── VERSION                      # 版本号文件
├── CHANGELOG.md                 # 变更日志
└── LICENSE                      # 许可证文件
```

### 目录说明

- **`cmd/`**: 应用程序入口点，包含主程序和各类CLI工具
  - `exporter/`: 主HTTP服务器程序
  - `admin/`: 管理和运维CLI工具
  - `benchmark/`: 性能测试和基准测试工具
  - `config-migrate/`: 配置格式迁移工具

- **`internal/`**: 内部实现包，项目私有逻辑不对外暴露
  - `internal/pkgs/`: 内部公共包，仅供项目内部模块使用
    - `log/`: 结构化日志，提供统一的日志接口
  - `config/`: 配置管理，提供统一的配置加载和验证
  - `storage/`: 存储模块，定义存储接口和具体实现
  - `auth/`: WinPower系统认证实现
  - `collector/`: 设备数据采集实现
  - `energy/`: 电能计算和累加实现
  - `metrics/`: Prometheus指标管理实现
  - `server/`: HTTP服务器和相关处理逻辑
  - `scheduler/`: 定时任务调度实现

- **`tests/`**: 测试相关文件
  - `integration/`: 集成测试和端到端测试
  - `fixtures/`: 测试数据和模拟响应
  - `mocks/`: 测试用的Mock对象

- **`scripts/`**: 构建和部署脚本
- **`deployments/``: 各种环境的部署配置

### 模块组织原则

1. **模块组织原则**:
   - 所有包都在 `internal/` 目录下，仅供项目内部使用
   - `internal/pkgs/` 包含内部公共组件，供项目内其他模块共享

2. **单一职责**: 每个包只负责一个明确的功能领域
3. **接口优先**: 通过 `interface.go` 文件定义公共接口
4. **测试完整**: 每个包都有对应的 `*_test.go` 文件
5. **文档齐全**: 关键模块都有对应的设计文档

### CMD命令设计

#### 1. 主程序 (cmd/exporter/)
```bash
# 启动HTTP服务器
winpower-g2-exporter [flags]

Flags:
  --config string              配置文件路径 (default "config.yaml")
  --port int                   服务端口 (default 9090)
  --log-level string           日志级别 (default "info")
  --winpower.url string        WinPower服务地址
  --winpower.username string   用户名
  --winpower.password string   密码
  --data-dir string            数据目录 (default "./data")
  --skip-ssl-verify            跳过SSL验证
  --sync-write                 同步写入 (default true)
```

#### 2. 管理工具 (cmd/admin/)
```bash
# Token管理
winpower-admin token [command]
  status      检查当前Token状态
  refresh     刷新Token
  validate    验证Token有效性

# 设备管理
winpower-admin device [command]
  list        列出所有设备
  status      检查设备状态
  test        测试设备连接

# 配置管理
winpower-admin config [command]
  validate    验证配置文件
  migrate     迁移配置格式
  generate    生成示例配置
```

#### 3. 性能测试 (cmd/benchmark/)
```bash
# 运行基准测试
winpower-benchmark [command]
  collector   测试采集性能
  storage     测试存储性能
  full        完整系统性能测试
```

#### 4. 配置迁移 (cmd/config-migrate/)
```bash
# 配置格式迁移
winpower-config-migrate [command]
  migrate     迁移配置到新格式
  validate    验证配置文件格式
```

## 4. 模块划分（职责与输入/输出）

1) config（配置）
- 职责：集中管理配置加载（文件/环境变量），提供结构化配置对象。
- 输入：配置文件、环境变量。
- 输出：`Config`（包含 ServerConfig、WinPowerConfig、LogConfig、StorageConfig 等子配置）。
- 设计原则：各模块的config结构体统一在config模块内定义，各模块不直接读取config模块内容，通过构造函数接收配置参数。

2) log（日志）
- 职责：统一结构化日志，提供分级日志 API。
- 输入：`Config.Log`。
- 输出：`Logger` 接口（供各模块使用）。

3) auth（认证）
- 职责：与 WinPower 系统交互，管理 Token 获取与刷新。
- 输入：`Config.WinPower`、`Logger`。
- 输出：`AuthManager` 接口（`GetToken()`、`RefreshToken()`、`IsTokenValid()`）。
- 设计：不定义独立的Config结构，通过构造函数接收WinPowerConfig和刷新阈值参数。

4) collector（采集）
- 职责：封装 WinPower API 调用，拉取设备与功率等实时数据，并在成功采样后触发电能累计。
- 输入：`AuthManager`、`Config.WinPower`、`Logger`。
- 输出：`CollectDeviceData(ctx)` 的执行结果与解析后的 `ParsedDeviceData`（包含功率/设备信息）；同时调用 `Energy.Calculate(deviceID, power)` 触发累计。

5) storage（存储）
- 职责：持久化电能累计值与必要的元数据；可替换实现。
- 输入：`Config.Storage`、`Logger`。
- 输出：`StorageManager` 接口（`Write(deviceID, *PowerData)`、`Read(deviceID) (*PowerData, error)`）。
- 设计：不定义独立的Config结构，通过构造函数接收StorageConfig参数。

6) energy（电能计算）
- 职责：基于功率读数做积分计算并持久化（Wh/kWh）。
- 输入：由 `collector` 在采样到瞬时功率后调用；依赖 `StorageManager` 与 `Logger`。
- 输出：写入累计电能值；对外提供查询接口 `Get(deviceID)`。

7) scheduler（定时调度）
- 职责：按固定周期驱动两类任务：每 1 分钟刷新 Token、每 5 秒触发采集。
- 输入：`AuthManager`、`Collector`、`Logger`。
- 输出：周期性执行、状态日志（`RefreshToken()` 与 `CollectDeviceData()`）。

8) metrics（指标转换）
- 职责：管理并暴露指标注册表；对外提供 `/metrics` 的 HTTP Handler。
- 输入：来自 `collector/energy/auth` 的指标更新（只写）；`Logger`。
- 输出：`MetricManager.Handler()` 返回注册指标快照（不触发采集或计算）。

9) server（HTTP 服务）
- 职责：仅负责 HTTP 层的路由与中间件，暴露 `/metrics` 与 `/health`。
- 输入：`Config.Server`、`Logger`、`metrics` 与 `health` 依赖。
- 输出：HTTP 服务，优雅关闭与基本观察性（日志/pprof 可选）。
- 设计：不定义独立的Config结构，通过构造函数接收ServerConfig参数。

## 5. 依赖关系（简化版）

```
┌───────────────────────────────────────────┐
│                 server                    │
│  - /metrics -> metrics.Handler()          │
│  - /health  -> health.Check()             │
└─────────────┬─────────────────────────────┘
              │
              ▼
┌───────────────────────────────────────────┐
│                 metrics                   │
│  - 暴露注册表最新快照（请求触发统一采集）│
│  - collector/energy/auth 更新指标         │
└─────────────┬──────────────┬──────────────┘
              │              │
              ▼              ▼
┌──────────────────┐   ┌──────────────────┐
│    collector     │   │      energy      │
│ - 使用 auth      │   │ - 使用 storage   │
│ - 配置 winpower  │   │ - 调度 scheduler │
└─────────┬────────┘   └─────────┬────────┘
          │                      │
          ▼                      ▼
┌──────────────────┐   ┌──────────────────┐
│      auth        │   │     storage      │
│ - 配置与日志     │   │ - 配置与日志     │
└─────────┬────────┘   └─────────┬────────┘
          │                      │
          ▼                      ▼
     ┌───────────┐          ┌───────────┐
     │   config   │          │    log     │
     └───────────┘          └───────────┘
```

依赖规则：
- `server` 仅依赖 `metrics`（以及健康检查）与配置/日志。
- `metrics` 不调用 `collector/energy`；仅暴露 Handler 并读取自身注册表；`collector/energy/auth` 在各自流程中更新指标。
- `energy` 由 `collector` 触发计算，依赖 `storage` 写入累计值。
- `collector` 依赖 `auth` 获取 Token；`auth/storage` 通过构造函数接收配置参数，依赖 `log`。
- `scheduler` 触发 `auth.RefreshToken()` 与 `collector.CollectDeviceData()`，不依赖 `server`。
- **config依赖原则**：只有cmd模块依赖config模块进行配置加载和初始化，其他模块通过构造函数接收配置参数。

## 6. 数据流（运行时）

1) 指标请求路径（拉取）：

```
Prometheus GET /metrics
        ↓
server 解析路由与中间件
        ↓
 metrics.Handler()
  - 调用统一采集入口 `collector.CollectDeviceData(ctx)`
  - 采集与累计完成后返回注册表最新快照
```

2) 电能计算路径（定时）：

```
Tick(5s) → collector.CollectDeviceData(ctx)
  - 采样设备数据并解析功率
  - energy.Calculate(deviceID, power)
  - storage.Write(deviceID, *PowerData)
```

## 7. 接口概要（面向实现）

```go
// 日志
type Logger interface {
    Info(msg string, fields ...any)
    Warn(msg string, fields ...any)
    Error(err error, msg string, fields ...any)
}

// 认证
type AuthManager interface {
    GetToken() (string, error)
    RefreshToken() (string, error)
    IsTokenValid() bool
}

// 采集
type Collector interface {
    CollectDeviceData(ctx context.Context) error
    GetConnectionStatus() bool
    GetLastCollectionTime() time.Time
}

// 存储
type StorageManager interface {
    Read(deviceID string) (*PowerData, error)
    Write(deviceID string, data *PowerData) error
}

type PowerData struct {
    Timestamp int64   `json:"timestamp"` // 毫秒时间戳
    EnergyWH  float64 `json:"energy_wh"` // 累计电能(Wh)
}

// 电能
type EnergyInterface interface {
    Calculate(deviceID string, power float64) (float64, error)
    Get(deviceID string) (float64, error)
}

// 指标
type MetricManager interface {
    Handler() gin.HandlerFunc // 返回 /metrics 的处理器
}

// 服务
type HTTPServer interface {
    Start() error
    Stop(ctx context.Context) error
}
```

接口约束：
- 不暴露内部实现细节，依赖均为向下的接口类型。
- 服务之间通过清晰接口交互，避免循环依赖。

## 8. 配置总览（示例）

```yaml
server:
  host: 0.0.0.0
  port: 9090
  mode: release                # debug/release/test
  timeouts:
    read: 30s
    write: 30s
    idle: 60s
  enable_pprof: false
  cors:
    enabled: false
    allow_origins: ["*"]
  rate_limit:
    enabled: false
    requests_per_minute: 1000
    burst: 100

auth:
  base_url: "https://winpower.example.com"
  username: "admin"
  password: "secret"
  timeout: 15s

winpower:
  api_timeout: 10s
  max_retries: 2

scheduler:
  energy_interval: 5s         # 电能计算周期

storage:
  type: file                  # file / memory / custom
  data_dir: ./data

log:
  level: info                 # debug/info/warn/error
```

## 9. 测试与可观察性

- 单元测试：对 `auth/collector/energy/storage/metrics` 分别进行接口级测试。
- 模拟依赖：通过 `TokenProvider`、`EnergyStore` 的 Mock 隔离外部副作用。
- 观察性：服务层统一日志；`/debug/pprof` 可选开启用于性能诊断。

## 10. 演进建议

- 设备类型扩展：在 `collector` 内扩展解析器与领域模型，不影响上层模块。
- 存储替换：通过实现 `EnergyStore` 接口替换文件存储为数据库或 KV。
- 指标扩展：在 `metrics` 新增指标转换器，保持与 `Snapshot/Energy` 解耦。
- 部署优化：生产环境优先使用反向代理终结 TLS；Exporter 保持纯 HTTP。