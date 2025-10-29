# CMD 模块设计文档

## 概述

CMD 模块是 WinPower G2 Exporter 的命令行接口层，基于 Cobra 库实现，提供统一的命令行入口点和子命令管理。模块负责处理命令行参数解析、子命令路由以及应用程序的生命周期管理，为用户提供友好的交互界面。

## 设计目标

1. **统一命令行接口**：提供一致的用户交互体验
2. **模块化子命令**：清晰的子命令划分，职责明确
3. **默认行为友好**：默认显示帮助信息，降低用户学习成本
4. **可扩展架构**：便于添加新的子命令和功能
5. **编译时信息注入**：支持构建时注入版本信息

## 命令结构

### 根命令

```
winpower-g2-exporter [flags] [command]
```

### 子命令

1. **server** - 启动 HTTP 服务器
2. **help** - 显示帮助信息（默认命令）
3. **version** - 显示版本信息

## 接口设计

### CLI 应用接口

```go
// CLI 应用程序接口
type CLI interface {
    // Execute 执行命令行程序
    Execute() error

    // SetArgs 设置命令行参数
    SetArgs(args []string)

    // AddCommand 添加子命令
    AddCommand(cmd *cobra.Command)
}
```

### 根命令结构

```go
// RootCmd 根命令
type RootCmd struct {
    cfgFile string
    verbose bool
    cmd     *cobra.Command
}

// NewRootCmd 创建根命令
func NewRootCmd() *RootCmd {
    root := &RootCmd{
        cmd: cobra.Command{
            Use:   "winpower-g2-exporter",
            Short: "WinPower G2 设备数据采集和 Prometheus 指标导出器",
            Long: `WinPower G2 Exporter 是一个用于采集 WinPower 设备数据、
计算能耗并以 Prometheus 指标形式导出的工具。

支持采集设备状态、电能数据，并提供 HTTP 接口供 Prometheus 抓取。`,
        },
    }

    // 设置默认子命令为 help
    root.cmd.SetHelpCommand(helpCmd)

    return root
}
```

### Server 子命令

```go
// ServerCmd server 子命令
type ServerCmd struct {
    port    int
    cfgFile string

    cmd *cobra.Command
}

// NewServerCmd 创建 server 子命令
func NewServerCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "server",
        Short: "启动 HTTP 服务器",
        Long: `启动 WinPower G2 Exporter HTTP 服务器，
提供 /metrics 和 /health 端点。`,
        RunE: runServer,
    }

    // 添加命令行参数
    cmd.Flags().IntVarP(&port, "port", "p", 9090, "HTTP 服务器端口")
    cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")

    return cmd
}

// runServer 执行服务器启动逻辑
func runServer(cmd *cobra.Command, args []string) error {
    // 1. 加载配置
    // 2. 初始化日志
    // 3. 启动服务器
    // 4. 处理优雅关闭
}
```

### Help 子命令

```go
// HelpCmd help 子命令
type HelpCmd struct {
    cmd *cobra.Command
}

// NewHelpCmd 创建 help 子命令
func NewHelpCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "help [command]",
        Short: "显示帮助信息",
        Long: `显示任何命令的帮助信息。如果没有指定命令，
则显示根命令的帮助信息。`,
        RunE: runHelp,
    }

    return cmd
}

// runHelp 执行帮助显示逻辑
func runHelp(cmd *cobra.Command, args []string) error {
    // 使用 Cobra 的帮助系统
}
```

### Version 子命令

```go
// VersionInfo 版本信息结构
type VersionInfo struct {
    Version     string `json:"version"`     // 版本号
    GoVersion   string `json:"go_version"`  // Go 运行时版本
    BuildTime   string `json:"build_time"`  // 编译时间
    CommitID    string `json:"commit_id"`   // Commit ID
    Platform    string `json:"platform"`    // 运行平台
    Compiler    string `json:"compiler"`    // 编译器信息
}

// VersionCmd version 子命令
type VersionCmd struct {
    format string // 输出格式: json, text

    cmd *cobra.Command
}

// NewVersionCmd 创建 version 子命令
func NewVersionCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "version",
        Short: "显示版本信息",
        Long: `显示应用程序的版本信息，包括：
- 版本号
- Go 运行时信息
- 编译时间
- Git Commit ID
- 平台信息`,
        RunE: runVersion,
    }

    // 添加输出格式参数
    cmd.Flags().StringVarP(&format, "format", "f", "text",
        "输出格式 (text|json)")

    return cmd
}

// runVersion 执行版本信息显示逻辑
func runVersion(cmd *cobra.Command, args []string) error {
    info := getVersionInfo()

    switch format {
    case "json":
        return outputJSON(info)
    default:
        return outputText(info)
    }
}

// getVersionInfo 获取版本信息
func getVersionInfo() *VersionInfo {
    return &VersionInfo{
        Version:     version,     // 编译时注入
        GoVersion:   runtime.Version(),
        BuildTime:   buildTime,   // 编译时注入
        CommitID:    commitID,    // 编译时注入
        Platform:    runtime.GOOS + "/" + runtime.GOARCH,
        Compiler:    runtime.Compiler,
    }
}
```

## 编译时信息注入

### VERSION 文件

项目使用根目录下的 `VERSION` 文件作为版本号的唯一来源：

```
# VERSION 文件内容示例
v1.0.0
```

**文件位置**：项目根目录 `/VERSION`

**作用**：
- 提供统一的版本号管理
- 支持构建脚本自动读取版本信息
- 便于版本发布和管理

**版本号格式**：遵循语义化版本规范 (SemVer) `v{major}.{minor}.{patch}`

### 构建标签

使用 Go 的构建标签在编译时注入版本信息：

```bash
# 构建时注入版本信息（从 VERSION 文件读取）
VERSION=$(cat VERSION) go build -ldflags="-X main.version=${VERSION} \
             -X main.buildTime=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
             -X main.commitID=$(git rev-parse HEAD)" \
             -o winpower-g2-exporter cmd/winpower-g2-exporter/main.go
```

### 变量定义

```go
// main.go
package main

import (
    "github.com/spf13/cobra"
)

// 编译时注入的变量，默认值
var (
    version   = "dev"
    buildTime = ""
    commitID  = ""
)

func main() {
    root := cmd.NewRootCmd()
    root.Execute()
}
```

## 配置集成

### 配置文件参数

```go
// 根命令持久化参数
func (r *RootCmd) addPersistentFlags() {
    r.cmd.PersistentFlags().StringVarP(&r.cfgFile, "config", "c", "",
        "配置文件路径")
    r.cmd.PersistentFlags().BoolVarP(&r.verbose, "verbose", "v", false,
        "详细输出模式")
}

// 配置绑定逻辑
func (r *RootCmd) bindConfig() error {
    if r.cfgFile != "" {
        viper.SetConfigFile(r.cfgFile)
    } else {
        // 使用默认配置文件搜索路径
        viper.SetConfigName("config")
        viper.AddConfigPath(".")
        viper.AddConfigPath("./config")
        viper.AddConfigPath("$HOME/.config/winpower-exporter")
        viper.AddConfigPath("/etc/winpower-exporter")
    }

    // 绑定环境变量
    viper.SetEnvPrefix("WINPOWER_EXPORTER")
    viper.AutomaticEnv()

    return viper.ReadInConfig()
}
```

## 生命周期管理

### 启动流程

```go
// 应用启动流程
func (r *RootCmd) Execute() error {
    // 1. 解析命令行参数
    // 2. 绑定配置
    // 3. 初始化日志
    // 4. 执行子命令
    return r.cmd.Execute()
}

// 服务器启动流程
func runServer(cmd *cobra.Command, args []string) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 1. 加载配置
    if err := loadConfig(); err != nil {
        return fmt.Errorf("加载配置失败: %w", err)
    }

    // 2. 初始化日志
    if err := initLogger(); err != nil {
        return fmt.Errorf("初始化日志失败: %w", err)
    }

    // 3. 初始化模块
    app, err := initializeApp(ctx)
    if err != nil {
        return fmt.Errorf("初始化应用失败: %w", err)
    }

    // 4. 设置信号处理
    setupSignalHandler(cancel)

    // 5. 启动服务器
    log.Info("启动 WinPower G2 Exporter")
    return app.Start(ctx)
}
```

## 模块启动顺序

### 启动依赖关系

WinPower G2 Exporter 的模块启动遵循严格的依赖顺序，确保每个模块在依赖的模块完全初始化后才开始工作：

```
启动顺序: config → logging → storage → winpower → energy → collector → metrics → server → scheduler
```

### 详细启动流程

```go
// initializeApp 按依赖顺序初始化所有模块
func initializeApp(ctx context.Context) (*App, error) {
    cfg := config.Get()

    // 1. 初始化日志模块 (logging)
    // 依赖: 配置模块
    logger, err := log.New(cfg.Logging)
    if err != nil {
        return nil, fmt.Errorf("初始化日志模块失败: %w", err)
    }
    logger.Info("日志模块初始化完成")

    // 2. 初始化存储模块 (storage)
    // 依赖: 配置模块、日志模块
    storage, err := storage.New(cfg.Storage, logger)
    if err != nil {
        return nil, fmt.Errorf("初始化存储模块失败: %w", err)
    }
    logger.Info("存储模块初始化完成")

    // 3. 初始化 WinPower 模块 (winpower)
    // 依赖: 配置模块、日志模块
    winpower, err := winpower.NewClient(cfg.WinPower, logger)
    if err != nil {
        return nil, fmt.Errorf("初始化 WinPower 模块失败: %w", err)
    }
    logger.Info("WinPower 模块初始化完成")

    // 4. 初始化电能计算模块 (energy)
    // 依赖: 配置模块、日志模块、存储模块
    energy, err := energy.NewCalculator(cfg.Energy, logger, storage)
    if err != nil {
        return nil, fmt.Errorf("初始化电能计算模块失败: %w", err)
    }
    logger.Info("电能计算模块初始化完成")

    // 5. 初始化采集器模块 (collector)
    // 依赖: 配置模块、日志模块、WinPower 模块、电能计算模块
    collector, err := collector.New(cfg.Collector, logger, winpower, energy)
    if err != nil {
        return nil, fmt.Errorf("初始化采集器模块失败: %w", err)
    }
    logger.Info("采集器模块初始化完成")

    // 6. 初始化指标模块 (metrics)
    // 依赖: 配置模块、日志模块、采集器模块
    metrics, err := metrics.NewRegistry(cfg.Metrics, logger, collector)
    if err != nil {
        return nil, fmt.Errorf("初始化指标模块失败: %w", err)
    }
    logger.Info("指标模块初始化完成")

    // 7. 初始化服务器模块 (server)
    // 依赖: 配置模块、日志模块、指标模块
    server, err := server.New(cfg.Server, logger, metrics)
    if err != nil {
        return nil, fmt.Errorf("初始化服务器模块失败: %w", err)
    }
    logger.Info("服务器模块初始化完成")

    // 8. 初始化调度器模块 (scheduler)
    // 依赖: 配置模块、日志模块、采集器模块
    scheduler, err := scheduler.New(cfg.Scheduler, logger, collector)
    if err != nil {
        return nil, fmt.Errorf("初始化调度器模块失败: %w", err)
    }
    logger.Info("调度器模块初始化完成")

    return &App{
        Config:    cfg,
        Logger:    logger,
        Storage:   storage,
        WinPower:  winpower,
        Energy:    energy,
        Collector: collector,
        Metrics:   metrics,
        Server:    server,
        Scheduler: scheduler,
    }, nil
}
```

### 模块启动时序图

```
时间 →
┌─────────────────────────────────────────────────────────────────┐
│ 启动流程                                                         │
├─────────────────────────────────────────────────────────────────┤
│ loadConfig()                                                    │
│    ↓                                                           │
│ initLogger()                                                    │
│    ↓                                                           │
│ initializeApp()                                                 │
│    ├─ storage.New()      ←─ 依赖配置和日志                      │
│    ├─ winpower.New()     ←─ 依赖配置和日志                      │
│    ├─ energy.New()       ←─ 依赖配置、日志、存储                 │
│    ├─ collector.New()    ←─ 依赖WinPower和Energy                 │
│    ├─ metrics.New()      ←─ 依赖Collector                       │
│    ├─ server.New()       ←─ 依赖Metrics                          │
│    └─ scheduler.New()    ←─ 依赖Collector                       │
│    ↓                                                           │
│ setupSignalHandler()                                             │
│    ↓                                                           │
│ server.Start()                                                  │
│    ↓                                                           │
│ scheduler.Start()  ←─ 在独立 goroutine 中运行                   │
└─────────────────────────────────────────────────────────────────┘
```

### 模块依赖关系表

| 模块 | 启动顺序 | 依赖模块 | 作用 |
|------|----------|----------|------|
| **config** | 1 | 无 | 配置管理，为其他模块提供配置 |
| **logging** | 2 | config | 日志记录，为其他模块提供日志接口 |
| **storage** | 3 | config, logging | 数据持久化，存储电能累计值 |
| **winpower** | 4 | config, logging | WinPower 系统连接和数据采集 |
| **energy** | 5 | config, logging, storage | 电能计算和累计 |
| **collector** | 6 | config, logging, winpower, energy | 数据采集协调器 |
| **metrics** | 7 | config, logging, collector | Prometheus 指标管理 |
| **server** | 8 | config, logging, metrics | HTTP 服务器，提供端点 |
| **scheduler** | 9 | config, logging, collector | 定时调度器，触发生成 |

### 启动失败处理

```go
// 启动失败时的回滚处理
func (app *App) Shutdown(ctx context.Context) error {
    var errors []error

    // 按相反顺序关闭模块
    if app.Scheduler != nil {
        if err := app.Scheduler.Stop(ctx); err != nil {
            errors = append(errors, fmt.Errorf("关闭调度器失败: %w", err))
        }
    }

    if app.Server != nil {
        if err := app.Server.Stop(ctx); err != nil {
            errors = append(errors, fmt.Errorf("关闭服务器失败: %w", err))
        }
    }

    if app.WinPower != nil {
        if err := app.WinPower.Close(); err != nil {
            errors = append(errors, fmt.Errorf("关闭 WinPower 连接失败: %w", err))
        }
    }

    if app.Storage != nil {
        if err := app.Storage.Close(); err != nil {
            errors = append(errors, fmt.Errorf("关闭存储失败: %w", err))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("关闭过程中发生错误: %v", errors)
    }

    return nil
}
```

### 启动日志示例

```
2024-01-15T10:00:00Z  INFO  cmd/server.go:45  开始启动 WinPower G2 Exporter
2024-01-15T10:00:00Z  INFO  cmd/server.go:52  配置加载完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:58  日志模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:65  存储模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:72  WinPower 模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:79  电能计算模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:86  采集器模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:93  指标模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:100  服务器模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:107  调度器模块初始化完成
2024-01-15T10:00:01Z  INFO  cmd/server.go:114  HTTP 服务器启动在端口 9090
2024-01-15T10:00:01Z  INFO  cmd/server.go:120  调度器启动，采集间隔 5s
2024-01-15T10:00:01Z  INFO  cmd/server.go:125  WinPower G2 Exporter 启动完成
```

### 优雅关闭

```go
// 信号处理
func setupSignalHandler(cancel context.CancelFunc) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        sig := <-sigChan
        log.Info("收到信号 %v，开始优雅关闭", sig)
        cancel()
    }()
}
```

## 错误处理

### 统一错误处理

```go
// 错误处理包装器
func handleError(err error) {
    if err == nil {
        return
    }

    // 根据错误类型显示不同的信息
    switch {
    case errors.Is(err, context.Canceled):
        log.Info("程序被取消")
    case errors.Is(err, context.DeadlineExceeded):
        log.Error("操作超时")
    default:
        log.Error("程序执行失败: %v", err)
    }

    os.Exit(1)
}
```

### 配置错误处理

```go
// 配置验证错误处理
func validateConfig(cfg *config.Config) error {
    if cfg.WinPower.URL == "" {
        return fmt.Errorf("WinPower URL 不能为空")
    }
    if cfg.WinPower.Username == "" {
        return fmt.Errorf("WinPower 用户名不能为空")
    }
    if cfg.WinPower.Password == "" {
        return fmt.Errorf("WinPower 密码不能为空")
    }

    return nil
}
```

## 使用示例

### 基本使用

```bash
# 显示帮助信息（默认行为）
./winpower-g2-exporter

# 或显式调用 help
./winpower-g2-exporter help

# 显示 server 子命令帮助
./winpower-g2-exporter help server

# 启动服务器
./winpower-g2-exporter server

# 使用自定义配置文件
./winpower-g2-exporter server --config /path/to/config.yaml

# 指定端口
./winpower-g2-exporter server --port 8080
```

### 版本信息

```bash
# 显示版本信息
./winpower-g2-exporter version

# JSON 格式输出
./winpower-g2-exporter version --format json
```

### 环境变量

```bash
# 设置环境变量
export WINPOWER_EXPORTER_SERVER_PORT=9090
export WINPOWER_EXPORTER_WINPOWER_URL=https://winpower.example.com
export WINPOWER_EXPORTER_LOGGING_LEVEL=debug

# 启动服务器
./winpower-g2-exporter server
```

## 项目结构

```
cmd/
└── winpower-g2-exporter/
    ├── main.go                    # 主程序入口
    ├── root.go                   # 根命令实现
    ├── server.go                 # server 子命令
    ├── help.go                   # help 子命令
    ├── version.go                # version 子命令
    └── root_test.go              # 测试文件
```

## 依赖管理

### 主要依赖

```bash
# 安装依赖
go get github.com/spf13/cobra      # CLI 框架
go get github.com/spf13/viper    # 配置管理
go get github.com/spf13/pflag    # 命令行参数
```

### 版本信息依赖

版本信息显示不需要额外依赖，使用标准库即可：
- `runtime` - Go 运行时信息
- `encoding/json` - JSON 格式输出

## 构建和部署

### Makefile 集成

```makefile
# 版本号从 VERSION 文件读取
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")

# 构建命令
build:
	@echo "构建 winpower-g2-exporter (版本: $(VERSION))..."
	go build -ldflags="-X main.version=$(VERSION) \
		-X main.buildTime=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ') \
		-X main.commitID=$(shell git rev-parse HEAD 2>/dev/null || echo "")" \
		-o bin/winpower-g2-exporter cmd/winpower-g2-exporter/main.go

# 构建 Linux 版本
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

# 构建所有平台
build-all:
	@echo "构建所有平台..."
	# 多平台构建逻辑

# 显示版本信息
version:
	@echo "版本: $(VERSION)"
	@echo "编译时间: $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')"
	@echo "Commit ID: $(shell git rev-parse HEAD 2>/dev/null || echo "未知")"
```

### Docker 集成

```dockerfile
# Dockerfile
# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG VERSION=dev
ARG COMMIT

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 设置构建参数
ARG TARGETOS
ARG TARGETARCH

# 复制 VERSION 文件（如果存在）
COPY VERSION* ./

# 使用 make 构建二进制文件
RUN VERSION=${VERSION} COMMIT=${COMMIT} make build

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/winpower-g2-exporter .
COPY --from=builder /app/config.yaml.example .

# 设置文件权限
RUN chown -R appuser:appgroup /app
USER appuser

# 暴露端口
EXPOSE 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./winpower-g2-exporter version || exit 1

CMD ["./winpower-g2-exporter", "server"]
```

### Docker 构建命令

```bash
# 单平台构建
docker build -t winpower-g2-exporter:latest .

# 多平台构建
docker buildx build \
    --platform linux/amd64,linux/arm64,linux/arm/v7 \
    -t winpower-g2-exporter:latest \
    --push .

# 带版本信息构建（从 VERSION 文件读取）
docker build \
    --build-arg VERSION=$(cat VERSION) \
    --build-arg COMMIT=$(git rev-parse HEAD) \
    -t winpower-g2-exporter:$(cat VERSION) .

# 或者指定特定版本
docker build \
    --build-arg VERSION=v1.0.0 \
    --build-arg COMMIT=$(git rev-parse HEAD) \
    -t winpower-g2-exporter:v1.0.0 .
```

## 测试策略

### 单元测试

- 根命令功能测试
- 子命令参数解析测试
- 版本信息格式化测试
- 配置绑定测试

### 集成测试

- 完整命令行流程测试
- 配置文件加载测试
- 环境变量解析测试

### 测试文件组织

```go
// root_test.go
package cmd

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRootCmd_Execute(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantErr  bool
    }{
        {
            name:    "默认显示帮助",
            args:    []string{},
            wantErr: false,
        },
        {
            name:    "help 命令",
            args:    []string{"help"},
            wantErr: false,
        },
        {
            name:    "version 命令",
            args:    []string{"version"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            root := NewRootCmd()
            root.SetArgs(tt.args)
            err := root.Execute()

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 安全考虑

1. **参数验证**：严格验证所有命令行参数
2. **配置文件权限**：检查配置文件读取权限
3. **敏感信息**：避免在帮助信息中暴露敏感配置
4. **错误信息**：错误信息不包含敏感细节
5. **路径安全**：验证配置文件路径的安全性

## 性能考虑

CMD 模块本身不涉及性能关键路径，主要注意事项：

1. **快速启动**：最小化初始化开销
2. **内存使用**：避免不必要的内存分配
3. **错误处理**：高效处理配置错误

此设计确保 CMD 模块提供清晰、易用的命令行界面，同时保持代码的可维护性和可扩展性。