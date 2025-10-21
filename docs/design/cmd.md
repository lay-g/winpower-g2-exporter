# CMD 模块设计文档

## 概述

CMD 模块提供 WinPower G2 Exporter 的命令行入口点，实现基于子命令的CLI结构，支持 server、version、help 三个子命令，遵循TDD开发原则。

## 设计目标

1. **简洁的子命令结构**：提供清晰直观的命令行接口
2. **模块化设计**：支持各模块的正确初始化和协作
3. **配置管理**：统一的配置加载和验证机制
4. **优雅关闭**：支持信号处理和资源清理
5. **跨平台支持**：支持多平台构建和部署

## 架构设计

### 子命令结构

```
exporter [command] [flags]

Commands:
  server    启动HTTP服务器（默认命令）
  version   显示版本信息
  help      显示帮助信息

Flags:
  --config string        配置文件路径 (default "./config.yaml")
  --port int             HTTP服务端口 (default 9090)
  --log-level string     日志级别 (debug|info|warn|error) (default "info")
  --data-dir string      数据目录 (default "./data")
  --skip-ssl-verify      跳过SSL验证 (default false)
  --sync-write           同步写入 (default true)
  -h, --help             显示帮助信息
```

### 模块架构

```
cmd/
├── exporter/
│   └── main.go              # 主程序入口点
└── internal/
    ├── cmd/
    │   ├── root.go          # 根命令定义
    │   ├── server.go        # server子命令实现
    │   ├── version.go       # version子命令实现
    │   └── help.go          # help子命令实现
    ├── config/
    │   ├── loader.go        # 配置加载器
    │   └── validator.go     # 配置验证器
    └── lifecycle/
        ├── starter.go       # 应用启动器
        └── shutdown.go      # 优雅关闭处理
```

## 接口设计

### Commander 接口

```go
type Commander interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args []string) error
    Validate(args []string) error
}
```

### ConfigLoader 接口

```go
type ConfigLoader interface {
    Load(path string) (*Config, error)
    Validate(config *Config) error
    MergeWithCLI(config *Config, cliArgs *CLIArgs) *Config
}
```

### LifecycleManager 接口

```go
type LifecycleManager interface {
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
    RegisterShutdownHook(hook func() error)
}
```

## 实现细节

### 1. 主程序入口点 (main.go)

```go
func main() {
    ctx := context.Background()

    // 创建根命令
    rootCmd := cmd.NewRootCommand()

    // 执行命令
    if err := rootCmd.Execute(ctx, os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### 2. 子命令实现

#### Server 子命令
- 解析配置文件和命令行参数
- 按依赖顺序初始化模块：config → logging → storage → auth → energy → collector → metrics → server → scheduler
- 启动HTTP服务器
- 处理信号和优雅关闭

#### Version 子命令
- 从项目根目录的 `VERSION` 文件读取版本号
- 显示应用程序版本
- 显示构建时间、Git提交哈希
- 显示Go版本信息

#### Help 子命令
- 显示所有可用子命令
- 显示每个子命令的用法和选项
- 提供配置文件格式示例

### 3. 配置加载机制

1. **默认配置**：内置默认配置值
2. **配置文件**：YAML格式配置文件
3. **环境变量**：WINPOWER_EXPORTER_ 前缀的环境变量
4. **命令行参数**：最高优先级，覆盖其他配置

配置优先级：命令行参数 > 环境变量 > 配置文件 > 默认值

### 4. 模块初始化顺序

```
config (配置加载)
   ↓
logging (日志初始化)
   ↓
storage (存储初始化)
   ↓
auth (认证模块)
   ↓
energy (能耗计算)
   ↓
collector (数据采集)
   ↓
metrics (指标注册)
   ↓
server (HTTP服务器)
   ↓
scheduler (调度器)
```

### 5. 优雅关闭机制

1. **信号处理**：监听 SIGTERM、SIGINT 信号
2. **关闭顺序**：按初始化的逆序关闭模块
3. **超时控制**：设置关闭超时时间，强制退出超时模块
4. **资源清理**：确保所有资源（连接、文件、goroutine）正确清理

## 错误处理

### 启动阶段错误
- 配置文件不存在或格式错误
- 模块初始化失败
- 端口被占用
- 权限不足

### 运行时错误
- 模块运行异常
- 外部服务连接失败
- 资源耗尽

### 关闭阶段错误
- 模块关闭超时
- 资源清理失败

所有错误都会：
1. 记录详细的错误日志
2. 返回适当的错误码
3. 提供用户友好的错误信息

## 测试策略

### 单元测试
- 每个子命令的独立测试
- 配置加载和验证测试
- 错误处理场景测试
- 生命周期管理测试

### 集成测试
- 完整的启动流程测试
- 模块间协作测试
- 优雅关闭流程测试

### 端到端测试
- 不同配置组合的启动测试
- 信号处理测试
- 多平台构建测试

## 构建和部署

### 构建目标
- Linux AMD64 (主要生产环境)
- Windows AMD64 (Windows环境)
- macOS AMD64 (开发和测试)
- Linux ARM64 (ARM架构支持)

### Docker 支持
- 多阶段构建优化镜像大小
- 非root用户运行
- 健康检查支持
- 配置文件挂载支持

## 安全考虑

1. **配置文件安全**：敏感信息通过环境变量传递
2. **权限控制**：最小权限原则运行
3. **输入验证**：严格的配置参数验证
4. **日志安全**：避免记录敏感信息

## 性能优化

1. **启动时间**：优化模块初始化顺序
2. **内存使用**：及时释放不需要的资源
3. **并发安全**：正确的并发控制和锁机制
4. **缓存机制**：合理使用缓存减少重复计算

## 监控和可观测性

1. **启动监控**：记录启动时间和各模块初始化时间
2. **健康检查**：提供健康检查端点
3. **指标导出**：导出启动相关的Prometheus指标
4. **结构化日志**：使用结构化日志便于分析

## 版本管理

### 版本文件结构
- 版本号存储在项目根目录的 `VERSION` 文件中
- 文件内容格式：`0.1.0` (遵循语义化版本控制，不包含v前缀)
- 构建时将版本信息编译到二进制文件中，显示时可选择性添加v前缀

### 版本信息结构
```go
type VersionInfo struct {
    Version   string    // 从 VERSION 文件读取
    BuildTime time.Time // 构建时间
    GitCommit string    // Git 提交哈希
    GoVersion string    // Go 版本信息
    Platform  string    // 运行平台
}
```

### 实现示例
```go
// 在构建时通过 -ldflags 注入版本信息
var (
    Version   = "dev"     // 默认开发版本
    BuildTime = "unknown" // 构建时间
    GitCommit = "unknown" // Git 提交哈希
)

func GetVersionInfo() *VersionInfo {
    return &VersionInfo{
        Version:   Version,
        BuildTime: BuildTime,
        GitCommit: GitCommit,
        GoVersion: runtime.Version(),
        Platform:  runtime.GOOS + "/" + runtime.GOARCH,
    }
}
```

### 构建脚本
```bash
# 读取版本文件并添加v前缀用于显示
VERSION=$(cat VERSION)
DISPLAY_VERSION="v${VERSION}"

# 构建时注入版本信息
go build -ldflags "-X main.Version=${VERSION} -X main.DisplayVersion=${DISPLAY_VERSION} -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') -X main.GitCommit=$(git rev-parse HEAD)"
```

使用语义化版本控制 (Semantic Versioning)：
- MAJOR.MINOR.PATCH
- 构建信息包含Git提交哈希和构建时间
- 支持版本比较和兼容性检查