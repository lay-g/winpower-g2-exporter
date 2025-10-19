# 配置管理模块设计

## 1. 概述

配置管理模块负责处理应用程序的所有配置参数，包括命令行参数、环境变量、配置文件等多种配置源的解析、验证和访问。

### 1.1 设计目标

- **灵活性**: 支持多种配置来源
- **优先级**: 明确的配置优先级规则
- **验证**: 完整的配置参数验证
- **易用性**: 简单的配置访问接口
- **扩展性**: 支持新配置参数的添加

### 1.2 功能范围

- 命令行参数解析
- 环境变量读取
- 配置文件加载（可选，支持 YAML）
- 配置参数验证
- 默认值设置
- 配置访问接口
- **嵌套配置结构**：支持复杂的配置层次结构
- **Viper 集成**：基于 Viper 的配置管理

## 2. 配置参数定义

### 2.1 必需参数

| 参数名            | 环境变量                      | 类型   | 说明                   |
| ----------------- | ----------------------------- | ------ | ---------------------- |
| winpower.url      | WINPOWER_EXPORTER_CONSOLE_URL | string | WinPower HTTP 服务地址 |
| winpower.username | WINPOWER_EXPORTER_USERNAME    | string | WinPower 用户名        |
| winpower.password | WINPOWER_EXPORTER_PASSWORD    | string | WinPower 密码          |

### 2.2 可选参数

| 参数名          | 环境变量                          | 类型     | 默认值      | 说明               |
| --------------- | --------------------------------- | -------- | ----------- | ------------------ |
| port            | WINPOWER_EXPORTER_PORT            | int      | 9090        | Exporter 服务端口  |
| log-level       | WINPOWER_EXPORTER_LOG_LEVEL       | string   | info        | 日志级别           |
| skip-ssl-verify | WINPOWER_EXPORTER_SKIP_SSL_VERIFY | bool     | false       | 跳过 SSL 证书验证  |
| data-dir        | WINPOWER_EXPORTER_DATA_DIR        | string   | ./data      | 数据文件目录        |

## 3. 配置优先级

配置参数的优先级从高到低：

1. **命令行参数**: 最高优先级
2. **环境变量**: 次优先级
3. **配置文件**: 低优先级（可选）
4. **默认值**: 最低优先级

```
┌─────────────────┐
│ Command Line    │  Highest Priority
├─────────────────┤
│ Environment Var │
├─────────────────┤
│ Config File     │  (Optional)
├─────────────────┤
│ Default Value   │  Lowest Priority
└─────────────────┘
```

## 4. 数据结构设计

### 4.1 配置结构体

```go
package config

import (
    "time"
    "os"
)

// Config 应用程序配置
type Config struct {
    Server   ServerConfig   `yaml:"server,omitempty"`
    Log      LogConfig      `yaml:"log,omitempty"`
    WinPower WinPowerConfig `yaml:"winpower"`
    Storage  StorageConfig  `yaml:"storage,omitempty"`

    // 配置文件路径（仅用于加载，不保存到配置文件）
    ConfigFile string `yaml:"-"`
}

// WinPowerConfig WinPower 系统配置
type WinPowerConfig struct {
    // 服务地址 (必需)
    URL string `mapstructure:"url" validate:"required,url"`

    // 用户名 (必需)
    Username string `mapstructure:"username" validate:"required"`

    // 密码 (必需)
    Password string `mapstructure:"password" validate:"required"`

    // 跳过 SSL 证书验证
    SkipSSLVerify bool `mapstructure:"skip_ssl_verify"`

    // 请求超时时间 (供 Collector 和 Auth 模块使用)
    Timeout time.Duration `mapstructure:"timeout"`
}

// ServerConfig Exporter 服务配置
type ServerConfig struct {
    // 监听端口
    Port int `yaml:"port" validate:"min=1,max=65535"`

    // 绑定地址
    Host string `yaml:"host"`

    // 读取超时
    ReadTimeout time.Duration `yaml:"read_timeout"`

    // 写入超时
    WriteTimeout time.Duration `yaml:"write_timeout"`

    // 启用 pprof
    EnablePprof bool `yaml:"enable_pprof"`
}

// LogConfig 日志配置
type LogConfig struct {
    // 日志级别: debug, info, warn, error
    Level string `yaml:"level" validate:"oneof=debug info warn error"`

    // 日志格式: json, console
    Format string `yaml:"format" validate:"oneof=json console"`

    // 日志输出: stdout, stderr, file, both
    Output string `yaml:"output"`

    // 日志文件路径
    FilePath string `yaml:"file_path"`

    // 日志轮转配置
    MaxSize          int  `yaml:"max_size"`           // MB
    MaxAge           int  `yaml:"max_age"`             // days
    MaxBackups       int  `yaml:"max_backups"`     // number of backups
    Compress         bool `yaml:"compress"`           // compress rotated files

    // 高级选项
    EnableCaller     bool `yaml:"enable_caller"`
    EnableStacktrace bool `yaml:"enable_stacktrace"`
}

// StorageConfig 存储配置
type StorageConfig struct {
    // 数据文件目录路径
    DataDir string `yaml:"data_dir" validate:"required"`

    // 文件权限
    FilePermissions os.FileMode `yaml:"file_permissions"`

    // 目录权限
    DirPermissions os.FileMode `yaml:"dir_permissions"`

    // 是否同步写入
    SyncWrite bool `yaml:"sync_write"`

    // 自动创建目录
    CreateDir bool `yaml:"create_dir"`
}
```

### 4.2 配置管理器接口

```go
// Manager 配置管理器接口（基于 Viper）
type Manager interface {
    // Load 加载配置
    Load() error

    // Get 获取配置
    Get() *Config

    // Validate 验证配置
    Validate() error
}

// LoadOption 配置加载选项
type LoadOption func(*manager)

// WithConfigFile 设置配置文件路径
func WithConfigFile(path string) LoadOption

// WithEnvPrefix 设置环境变量前缀
func WithEnvPrefix(prefix string) LoadOption

// 创建管理器的便捷函数
func NewManager(opts ...LoadOption) Manager
func LoadWithOptions(opts ...LoadOption) (Manager, error)
func LoadFromFileWithManager(path string) (Manager, error)
```

### 4.3 简单加载函数

```go
// 便捷加载函数
func Load() (*Config, error)
func LoadFromFile(path string) (*Config, error)
func LoadFromEnv() *Config
```

### 4.4 存储配置详细说明

#### 4.4.1 配置项说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| data_dir | string | "./data" | 数据文件目录路径，支持绝对路径和相对路径 |
| file_permissions | os.FileMode | 0644 | 文件权限设置，默认为读写权限 |
| sync_write | bool | true | 是否同步写入，确保数据持久化 |

#### 4.4.2 配置建议

##### 生产环境配置

```yaml
storage:
  data_dir: "/var/lib/winpower-exporter/data"
  file_permissions: "0644"
  sync_write: true
```

##### 开发环境配置

```yaml
storage:
  data_dir: "./data"
  file_permissions: "0644"
  sync_write: false  # 开发环境可以关闭同步写入以提高性能
```

##### Docker 容器配置

```yaml
storage:
  data_dir: "/app/data"  # 使用容器内的数据目录
  file_permissions: "0644"
  sync_write: true
```

#### 4.4.3 文件命名规则

在配置的数据目录中，每个设备将使用独立的文件存储：

- **文件命名**: `{device_id}.txt` - 使用设备ID作为文件名
- **文件示例**:
  - `a1.txt` - 设备ID为a1的数据文件
  - `b2.txt` - 设备ID为b2的数据文件
  - `ups_main.txt` - 设备ID为ups_main的数据文件

#### 4.4.4 配置验证

```go
// Validate 验证存储配置
func (sc *StorageConfig) Validate() error {
    // 检查数据目录路径
    if sc.DataDir == "" {
        return errors.New("data_dir cannot be empty")
    }

    // 检查目录是否可创建
    if err := os.MkdirAll(sc.DataDir, 0755); err != nil {
        return fmt.Errorf("failed to create data directory: %w", err)
    }

    // 检查目录是否可写
    testFile := filepath.Join(sc.DataDir, ".write_test")
    if err := os.WriteFile(testFile, []byte("test"), sc.FilePermissions); err != nil {
        return fmt.Errorf("data directory is not writable: %w", err)
    }
    os.Remove(testFile)

    return nil
}
```

## 5. 实现设计

### 5.1 初始化流程

```
┌────────────────────────────────────────────────┐
│              Initialize Config                 │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      1. Create Viper Instance                  │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      2. Set Default Values                     │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      3. Bind Environment Variables             │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      4. Parse Command Line Flags               │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      5. Load Config File (if exists)           │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      6. Unmarshal to Config Struct             │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      7. Validate Configuration                 │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      8. Return Config Instance                 │
└────────────────────────────────────────────────┘
```

### 5.2 核心函数

#### 5.2.1 创建配置管理器

```go
// NewManager 创建配置管理器
func NewManager() Manager {
    // 创建空的配置结构体实例
    // 创建新的 Viper 实例用于配置管理
    // 初始化管理器结构体并返回
}
```

#### 5.2.2 加载配置

```go
// Load 加载配置
func (m *manager) Load() error {
    // 调用 setDefaults 方法设置所有配置项的默认值
    // 调用 bindEnvVars 方法绑定环境变量到配置项
    // 调用 loadConfigFile 方法加载配置文件（如果存在）
    // 使用 Viper 将配置解析到 Config 结构体
    // 调用 Validate 方法验证配置的有效性
    // 返回加载结果或错误信息
}
```

#### 5.2.3 设置默认值

```go
// setDefaults 设置默认值
func (m *manager) setDefaults() {
    // 设置服务器相关默认值：端口、主机、读写超时等
    // 设置日志相关默认值：级别、格式、输出方式等
    // 设置 WinPower 相关默认值：SSL验证、超时时间等
    // 设置存储相关默认值：数据目录、文件权限、同步写入等
}
```

#### 5.2.4 绑定环境变量

```go
// bindEnvVars 绑定环境变量
func (m *manager) bindEnvVars() error {
    // 设置环境变量统一前缀为 WINPOWER_EXPORTER
    // 启用自动环境变量绑定功能
    // 设置环境变量键名替换规则（将点替换为下划线）
    // 定义关键配置项与环境变量的映射关系
    // 遍历映射表，显式绑定每个环境变量
    // 处理绑定过程中的错误并返回结果
}
```

#### 5.2.5 配置验证

```go
// Validate 验证配置
func (m *manager) Validate() error {
    // 创建 validator 实例用于结构体标签验证
    // 执行结构体验证，检查所有标签约束
    // 返回验证结果或错误信息
}
```

### 5.3 命令行集成

#### 5.3.1 Cobra 命令定义

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
    cfgFile string
    rootCmd = &cobra.Command{
        Use:   "winpower-g2-exporter",
        Short: "WinPower G2 Prometheus Exporter",
        Long:  "Export WinPower device metrics to Prometheus",
        RunE:  run,
    }
)

func init() {
    cobra.OnInitialize(initConfig)
    
    // 全局配置
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    
    // WinPower 配置
    rootCmd.Flags().String("winpower.url", "", "WinPower HTTP service URL (required)")
    rootCmd.Flags().String("winpower.username", "", "WinPower username (required)")
    rootCmd.Flags().String("winpower.password", "", "WinPower password (required)")
    rootCmd.Flags().Bool("skip-ssl-verify", false, "Skip SSL certificate verification")
    
    // Server 配置
    rootCmd.Flags().Int("port", 9090, "Exporter service port")
    
    // Log 配置
    rootCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
    
        
    // Storage 配置
    rootCmd.Flags().String("data-dir", "./data", "Data files directory path")
    rootCmd.Flags().Bool("sync-write", true, "Enable synchronous write for data persistence")
    
    // 标记必需参数
    rootCmd.MarkFlagRequired("winpower.url")
    rootCmd.MarkFlagRequired("winpower.username")
    rootCmd.MarkFlagRequired("winpower.password")
    
    // 绑定到 Viper
    viper.BindPFlags(rootCmd.Flags())
}
```

## 6. 配置文件支持

### 6.1 配置文件格式

支持 YAML 配置文件格式。

#### YAML 示例

```yaml
# config.yaml
winpower:
  url: "https://192.168.1.100:8081"
  username: "admin"
  password: "password"
  skip_ssl_verify: true
  timeout: 30s

server:
  port: 9090
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s

log:
  level: "info"
  format: "json"
  output: "stdout"

storage:
  data_dir: "./data"
  file_permissions: "0644"
  sync_write: true
```

### 6.2 配置文件加载

```go
// loadConfigFile 加载配置文件
func (m *manager) loadConfigFile() error {
    // 检查是否指定了配置文件路径，有则使用指定路径
    // 未指定时在默认位置搜索配置文件
    // 设置配置文件名和类型（默认为config.yaml）
    // 添加搜索路径：./config、$HOME/.config/winpower-exporter/、/etc/winpower-exporter/
    // 尝试读取配置文件，文件不存在时不报错
    // 处理其他读取错误并返回结果
}
```


## 7. 错误处理

### 7.1 错误类型

```go
var (
    // ErrConfigNotFound 配置文件未找到
    ErrConfigNotFound = errors.New("config file not found")
    
    // ErrInvalidConfig 无效的配置
    ErrInvalidConfig = errors.New("invalid configuration")
    
    // ErrMissingRequired 缺少必需参数
    ErrMissingRequired = errors.New("missing required parameter")
    
    // ErrValidationFailed 验证失败
    ErrValidationFailed = errors.New("configuration validation failed")
)
```

### 7.2 错误处理策略

```go
// Load 加载配置并处理错误
func (m *manager) Load() error {
    if err := m.load(); err != nil {
        // 记录详细错误日志
        log.Error("Failed to load configuration",
            zap.Error(err),
            zap.String("config_file", cfgFile),
        )
        
        // 根据错误类型返回不同的错误信息
        switch {
        case errors.Is(err, ErrMissingRequired):
            return fmt.Errorf("missing required parameters, please check --help")
        case errors.Is(err, ErrValidationFailed):
            return fmt.Errorf("configuration validation failed: %w", err)
        default:
            return fmt.Errorf("failed to load configuration: %w", err)
        }
    }
    
    return nil
}
```

## 8. 测试设计

### 8.1 单元测试

```go
func TestConfigLoad(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        wantErr bool
    }{
        {
            name: "valid config with defaults",
            setup: func() {
                os.Setenv("WINPOWER_EXPORTER_CONSOLE_URL", "http://localhost:8081")
                os.Setenv("WINPOWER_EXPORTER_USERNAME", "admin")
                os.Setenv("WINPOWER_EXPORTER_PASSWORD", "password")
            },
            wantErr: false,
        },
        {
            name: "missing required url",
            setup: func() {
                os.Unsetenv("WINPOWER_EXPORTER_CONSOLE_URL")
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            
            m := NewManager()
            err := m.Load()
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 8.2 集成测试

```go
func TestConfigIntegration(t *testing.T) {
    // 创建临时配置文件
    tmpfile, err := ioutil.TempFile("", "config.*.yaml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())
    
    // 写入配置
    config := `
winpower:
  url: "http://localhost:8081"
  username: "admin"
  password: "password"
server:
  port: 9090
`
    if _, err := tmpfile.Write([]byte(config)); err != nil {
        t.Fatal(err)
    }
    tmpfile.Close()
    
    // 加载配置
    cfgFile = tmpfile.Name()
    m := NewManager()
    if err := m.Load(); err != nil {
        t.Fatalf("Load() error = %v", err)
    }
    
    // 验证配置
    cfg := m.Get()
    if cfg.WinPower.URL != "http://localhost:8081" {
        t.Errorf("WinPower.URL = %v, want %v", cfg.WinPower.URL, "http://localhost:8081")
    }
}
```

## 9. 使用示例

### 9.1 基本使用

```go
package main

import (
    "log"
    "github.com/your-org/winpower-g2-exporter/internal/config"
)

func main() {
    // 创建配置管理器
    cfg := config.NewManager()
    
    // 加载配置
    if err := cfg.Load(); err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // 获取配置
    appConfig := cfg.Get()
    
    // 使用配置
    log.Printf("WinPower URL: %s", appConfig.WinPower.URL)
    log.Printf("Server Port: %d", appConfig.Server.Port)
    log.Printf("Log Level: %s", appConfig.Log.Level)
}
```


## 10. 最佳实践

### 10.1 环境变量命名

- 使用统一前缀: `WINPOWER_EXPORTER_`
- 使用大写字母
- 使用下划线分隔单词
- 避免与系统环境变量冲突

### 10.2 配置验证

- 在启动时验证所有配置
- 提供清晰的错误信息
- 记录配置验证日志
- 验证失败立即退出

### 10.3 敏感信息处理

- 密码不记录到日志
- 支持从环境变量读取敏感信息
- 配置文件设置适当的权限

### 10.4 配置文档

- 为每个配置项提供注释
- 提供YAML配置示例文件
- 文档化配置的默认值
- 说明配置项的影响

## 11. 性能考虑

### 11.1 配置缓存

```go
// 配置加载后缓存，避免重复解析
type manager struct {
    config *Config
    v      *viper.Viper
    mu     sync.RWMutex
}

// Get 获取配置（线程安全）
func (m *manager) Get() *Config {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.config
}
```


## 12. 安全考虑

### 12.1 敏感信息保护

```go
// String 实现，避免密码泄露
func (c *WinPowerConfig) String() string {
    return fmt.Sprintf("WinPowerConfig{URL: %s, Username: %s, Password: ***}",
        c.URL, c.Username)
}
```

### 12.2 配置文件权限

```go
// 检查配置文件权限
func checkConfigFilePermissions(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    
    // 配置文件不应该对其他用户可读
    if info.Mode().Perm()&0044 != 0 {
        return errors.New("config file has insecure permissions")
    }
    
    return nil
}
```

