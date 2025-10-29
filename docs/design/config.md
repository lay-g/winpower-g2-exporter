# Config 模块设计文档

## 概述

Config 模块提供 WinPower G2 Exporter 的统一配置管理功能，基于 viper 和 pflag 实现，支持多层级配置文件搜索、环境变量、命令行参数的配置加载机制，并提供可扩展的配置验证接口。

## 设计目标

1. **统一配置管理**：引用各模块的配置结构，提供统一的配置加载机制
2. **多源配置支持**：支持文件、环境变量、命令行参数等多种配置源
3. **多层级文件搜索**：按优先级自动搜索配置文件位置
4. **可扩展验证**：提供统一的配置验证接口，各模块可实现自己的验证逻辑
5. **类型安全**：提供类型安全的配置访问方法
6. **TDD 开发**：遵循测试驱动开发原则

## 配置加载机制

### 配置文件搜索路径

配置文件按以下优先级顺序搜索（找到第一个存在的配置文件即停止）：

1. **工作目录**：`./config.yaml`
2. **工作目录/config**：`./config/config.yaml`
3. **用户配置目录**：`$HOME/config/winpower-exporter/config.yaml`
4. **系统配置目录**：`/etc/winpower-exporter/config.yaml`

### 配置优先级

配置按以下优先级进行合并（高优先级覆盖低优先级）：

1. **默认值**：代码中定义的默认配置
2. **配置文件**：从上述搜索路径找到的配置文件
3. **环境变量**：`WINPOWER_EXPORTER_` 前缀的环境变量
4. **命令行参数**：通过 pflag 定义的命令行参数

## 接口设计

### ConfigValidator 接口

```go
// ConfigValidator 配置验证接口
// 各功能模块的配置结构体需要实现此接口以提供自定义验证逻辑
type ConfigValidator interface {
    Validate() error
}
```

### ConfigManager 接口

```go
// ConfigManager 配置管理器接口
type ConfigManager interface {
    // Load 加载配置
    Load() (*Config, error)

    // Get 获取配置值
    Get(key string) interface{}

    // GetString 获取字符串配置值
    GetString(key string) string

    // GetInt 获取整数配置值
    GetInt(key string) int

    // GetBool 获取布尔配置值
    GetBool(key string) bool

    // GetStringSlice 获取字符串切片配置值
    GetStringSlice(key string) []string

    // Set 设置配置值
    Set(key string, value interface{})

    // IsSet 检查配置是否已设置
    IsSet(key string) bool

    // Validate 验证完整配置
    Validate() error
}
```

## 配置结构设计

### 顶层 Config 结构体

```go
// Config 顶层配置结构体
// 引用各模块的配置结构体
type Config struct {
    // Server 服务器配置
    Server *server.Config `yaml:"server" mapstructure:"server"`

    // WinPower WinPower连接配置
    WinPower *winpower.Config `yaml:"winpower" mapstructure:"winpower"`

    // Storage 存储配置
    Storage *storage.Config `yaml:"storage" mapstructure:"storage"`

    // Scheduler 调度器配置
    Scheduler *scheduler.Config `yaml:"scheduler" mapstructure:"scheduler"`

    // Logging 日志配置
    Logging *log.Config `yaml:"logging" mapstructure:"logging"`
}
```

### 各模块配置结构体

各模块的配置结构体定义在各自的模块中，config模块通过引用的方式使用：

- **server.Config**: 定义在 `internal/server/config.go`，包含HTTP服务器相关配置
- **winpower.Config**: 定义在 `internal/winpower/config.go`，包含WinPower连接配置
- **storage.Config**: 定义在 `internal/storage/config.go`，包含文件存储配置
- **scheduler.Config**: 定义在 `internal/scheduler/config.go`，包含调度器配置
- **log.Config**: 定义在 `internal/pkgs/log/config.go`，包含日志配置

### 配置验证接口实现

各模块的配置结构体都实现了 `ConfigValidator` 接口：

```go
// 示例：server.Config 的验证实现
func (c *Config) Validate() error {
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("server port must be between 1 and 65535")
    }
    if c.Host == "" {
        return fmt.Errorf("server host cannot be empty")
    }
    return nil
}
```

## 实现细节

### 1. 配置加载器实现

```go
// Loader 配置加载器
type Loader struct {
    viper       *viper.Viper
    flags       *pflag.FlagSet
    searchPaths []string
}

// NewLoader 创建新的配置加载器
func NewLoader() *Loader {
    v := viper.New()
    v.SetConfigType("yaml")
    v.SetConfigName("config")

    // 设置环境变量前缀
    v.SetEnvPrefix("WINPOWER_EXPORTER")
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.AutomaticEnv()

    // 定义搜索路径
    searchPaths := []string{
        ".",                     // 工作目录
        "./config",              // 工作目录/config
        path.Join(os.Getenv("HOME"), "config/winpower-exporter"), // 用户配置目录
        "/etc/winpower-exporter", // 系统配置目录
    }

    for _, path := range searchPaths {
        v.AddConfigPath(path)
    }

    return &Loader{
        viper:       v,
        searchPaths: searchPaths,
    }
}

// Load 加载配置
func (l *Loader) Load() (*Config, error) {
    // 设置默认值
    l.setDefaults()

    // 绑定命令行参数
    if err := l.bindFlags(); err != nil {
        return nil, fmt.Errorf("failed to bind flags: %w", err)
    }

    // 读取配置文件
    if err := l.viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // 配置文件不存在是允许的，使用默认配置
            log.Info("No config file found, using defaults and environment variables")
        } else {
            return nil, fmt.Errorf("failed to read config file: %w", err)
        }
    }

    // 解析到配置结构体
    var config Config
    if err := l.viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &config, nil
}
```

### 2. 命令行参数绑定

```go
// bindFlags 绑定命令行参数
func (l *Loader) bindFlags() error {
    flags := pflag.NewFlagSet("winpower-exporter", pflag.ExitOnError)

    // Server 配置
    flags.Int("server.port", 9090, "HTTP server port")
    flags.String("server.host", "0.0.0.0", "HTTP server host")
    flags.Bool("server.enable-pprof", false, "Enable pprof debug endpoints")

    // WinPower 配置
    flags.String("winpower.url", "", "WinPower service URL")
    flags.String("winpower.username", "", "WinPower username")
    flags.String("winpower.password", "", "WinPower password")
    flags.Bool("winpower.skip-ssl-verify", false, "Skip SSL verification")

    // Storage 配置
    flags.String("storage.data_dir", "./data", "Data directory path")
    flags.Bool("storage.sync_write", true, "Enable synchronous write")

    // Logging 配置
    flags.String("logging.level", "info", "Log level (debug|info|warn|error)")
    flags.String("logging.format", "json", "Log format (json|console)")

    // 绑定到 viper
    if err := l.viper.BindPFlags(flags); err != nil {
        return err
    }

    l.flags = flags
    return nil
}
```

### 3. 默认值设置

```go
// setDefaults 设置默认配置值
func (l *Loader) setDefaults() {
    // Server 默认配置
    l.viper.SetDefault("server.port", 9090)
    l.viper.SetDefault("server.host", "0.0.0.0")
    l.viper.SetDefault("server.read_timeout", "30s")
    l.viper.SetDefault("server.write_timeout", "30s")
    l.viper.SetDefault("server.enable_pprof", false)

    // WinPower 默认配置
    l.viper.SetDefault("winpower.timeout", "30s")
    l.viper.SetDefault("winpower.skip_ssl_verify", false)
    l.viper.SetDefault("winpower.refresh_threshold", "5m")

    // Storage 默认配置
    l.viper.SetDefault("storage.data_dir", "./data")
    l.viper.SetDefault("storage.file_permissions", 0644)

    // Scheduler 默认配置
    l.viper.SetDefault("scheduler.collection_interval", "5s")
    l.viper.SetDefault("scheduler.graceful_shutdown_timeout", "5s")

    // Logging 默认配置
    l.viper.SetDefault("logging.level", "info")
    l.viper.SetDefault("logging.format", "json")
    l.viper.SetDefault("logging.output", "stdout")
}
```

### 4. 配置验证实现

```go
// Validate 验证完整配置
func (c *Config) Validate() error {
    // 验证各个模块的配置
    validators := []ConfigValidator{
        c.Server,
        c.WinPower,
        c.Storage,
        c.Scheduler,
        c.Logging,
    }

    for _, validator := range validators {
        if validator != nil {
            if err := validator.Validate(); err != nil {
                return fmt.Errorf("validation failed: %w", err)
            }
        }
    }

    return nil
}
```

## 环境变量映射

### 环境变量命名规则

环境变量使用 `WINPOWER_EXPORTER_` 前缀，配置键名中的 `.` 替换为 `_`：

```
WINPOWER_EXPORTER_SERVER_PORT=9090
WINPOWER_EXPORTER_SERVER_HOST=0.0.0.0
WINPOWER_EXPORTER_WINPOWER_BASE_URL=https://winpower.example.com
WINPOWER_EXPORTER_WINPOWER_USERNAME=admin
WINPOWER_EXPORTER_WINPOWER_PASSWORD=password
WINPOWER_EXPORTER_STORAGE_DATA_DIR=/var/lib/winpower-exporter
WINPOWER_EXPORTER_SCHEDULER_COLLECTION_INTERVAL=5s
WINPOWER_EXPORTER_LOGGING_LEVEL=debug
```

### 命令行参数命名规则

命令行参数使用 `[module].[option]` 格式，单词之间用 `-` 分割：

```
--server.port 9090
--server.host 0.0.0.0
--winpower.base_url https://winpower.example.com
--winpower.username admin
--winpower.password password
--storage.data_dir /var/lib/winpower-exporter
--scheduler.collection_interval 5s
--logging.level debug
```

## 使用示例

### 基本使用

```go
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
    server := server.New(cfg.Server)
    // ...
}
```

### 配置文件示例

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
  password: "password"
  timeout: 30s
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
```

## 错误处理

### 配置文件错误

- **文件不存在**：记录警告日志，使用默认配置
- **格式错误**：记录错误日志，退出程序
- **权限问题**：记录错误日志，退出程序

### 配置验证错误

- **必填字段缺失**：记录错误日志，退出程序
- **字段值无效**：记录错误日志，退出程序
- **字段类型错误**：记录错误日志，退出程序

### 运行时配置错误

- **环境变量格式错误**：记录警告日志，使用默认值
- **命令行参数错误**：显示帮助信息，退出程序

## 测试策略

### 单元测试

- 配置加载器功能测试
- 默认值设置测试
- 环境变量解析测试
- 命令行参数绑定测试
- 各模块配置验证测试

### 集成测试

- 完整配置加载流程测试
- 多源配置合并测试
- 配置文件搜索测试
- 错误场景处理测试

### 测试文件组织

```
internal/config/
├── config.go              # 主要配置结构体
├── loader.go              # 配置加载器
├── validator.go           # 配置验证器
├── defaults.go            # 默认值设置
├── config_test.go         # 配置功能测试
├── loader_test.go         # 加载器测试
├── validator_test.go      # 验证器测试
├── integration_test.go    # 集成测试
└── fixtures/              # 测试配置文件
    ├── valid_config.yaml
    ├── invalid_config.yaml
    └── partial_config.yaml
```

## 安全考虑

1. **敏感信息**：密码等敏感信息优先从环境变量读取
2. **权限检查**：配置文件权限检查，避免安全风险
3. **输入验证**：严格验证所有配置输入，防止注入攻击
4. **日志安全**：避免在日志中输出敏感配置信息