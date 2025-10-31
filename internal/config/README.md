# Config Package

统一配置管理模块，提供多源配置加载、验证和管理功能。

## 功能特性

- **多源配置支持**: 支持配置文件、环境变量、命令行参数
- **配置优先级**: 默认值 < 配置文件 < 环境变量 < 命令行参数
- **自动搜索**: 多层级配置文件自动搜索
- **类型安全**: 提供类型安全的配置访问方法
- **可扩展验证**: 各模块实现自定义验证逻辑
- **错误友好**: 详细的错误信息和上下文

## 配置文件搜索路径

配置文件按以下优先级顺序搜索（找到第一个存在的配置文件即停止）：

1. `./config.yaml` - 工作目录
2. `./config/config.yaml` - 工作目录/config
3. `$HOME/config/winpower-exporter/config.yaml` - 用户配置目录
4. `/etc/winpower-exporter/config.yaml` - 系统配置目录

## 环境变量

环境变量使用 `WINPOWER_EXPORTER_` 前缀，配置键名中的 `.` 替换为 `_`：

```bash
export WINPOWER_EXPORTER_SERVER_PORT=9090
export WINPOWER_EXPORTER_WINPOWER_BASE_URL=https://winpower.example.com
export WINPOWER_EXPORTER_LOGGING_LEVEL=debug
```

## 命令行参数

命令行参数使用 `[module].[option]` 格式：

```bash
./winpower-g2-exporter \
  --server.port 9090 \
  --winpower.base-url https://winpower.example.com \
  --logging.level debug
```

## 使用示例

### 基本使用

```go
package main

import (
    "log"
    "github.com/lay-g/winpower-g2-exporter/internal/config"
)

func main() {
    // 创建配置加载器
    loader := config.NewLoader()

    // 加载配置
    cfg, err := loader.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 验证配置
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }

    // 使用配置
    log.Printf("Server will listen on %s:%d", cfg.Server.Host, cfg.Server.Port)
}
```

### 动态配置访问

```go
loader := config.NewLoader()

// 获取配置值
port := loader.GetInt("server.port")
host := loader.GetString("server.host")
enabled := loader.GetBool("server.enable_pprof")

// 设置配置值
loader.Set("server.port", 9090)

// 检查配置是否已设置
if loader.IsSet("server.port") {
    log.Println("Port is configured")
}
```

## 配置文件示例

```yaml
server:
  port: 9090
  host: "0.0.0.0"
  mode: "release"
  read_timeout: 30s
  write_timeout: 30s
  enable_pprof: false

winpower:
  base_url: "https://winpower.example.com"
  username: "admin"
  password: "password"
  timeout: 30s
  skip_ssl_verify: false

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
```

## 接口

### ConfigValidator

各模块的配置结构体需要实现此接口：

```go
type ConfigValidator interface {
    Validate() error
}
```

### ConfigManager

配置管理器提供的接口：

```go
type ConfigManager interface {
    Load() (*Config, error)
    Get(key string) interface{}
    GetString(key string) string
    GetInt(key string) int
    GetBool(key string) bool
    GetStringSlice(key string) []string
    Set(key string, value interface{})
    IsSet(key string) bool
    Validate() error
}
```

## 错误处理

配置模块提供详细的错误类型和上下文：

```go
// 标准错误
var (
    ErrConfigNotFound   = errors.New("config file not found")
    ErrInvalidConfig    = errors.New("invalid configuration")
    ErrConfigValidation = errors.New("configuration validation failed")
    ErrConfigLoad       = errors.New("failed to load configuration")
    ErrConfigParse      = errors.New("failed to parse configuration")
    ErrFlagBinding      = errors.New("failed to bind command line flags")
)

// 详细错误类型
type ConfigError struct {
    Field   string  // 错误字段
    Message string  // 错误消息
    Err     error   // 底层错误
}
```

## 测试

运行测试：

```bash
# 单元测试
go test ./internal/config/...

# 测试覆盖率
go test -cover ./internal/config/...

# 详细测试输出
go test -v ./internal/config/...

# 集成测试
go test -v ./internal/config/ -run Integration
```

## 最佳实践

1. **优先使用配置文件**: 将大部分配置写入配置文件
2. **敏感信息用环境变量**: 密码等敏感信息通过环境变量传递
3. **命令行参数用于临时覆盖**: 仅在需要临时修改时使用
4. **始终验证配置**: 加载配置后立即调用 Validate()
5. **错误处理**: 使用 errors.Is 和 errors.As 检查具体错误类型

## 相关文档

- [配置模块设计文档](../../docs/design/config.md)
- [Server配置](../server/config.go)
- [WinPower配置](../winpower/config.go)
- [Storage配置](../storage/config.go)
- [Scheduler配置](../scheduler/config.go)
- [Logging配置](../pkgs/log/config.go)
