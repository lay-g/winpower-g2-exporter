# 模块化配置管理设计

## 1. 概述

配置管理采用分布式模块化设计，每个模块负责管理和维护自己的配置结构。配置加载工具包提供统一的配置加载接口，支持YAML配置文件、环境变量等多种配置源。

### 1.1 设计目标

- **模块化**: 每个模块拥有独立的配置结构和管理逻辑
- **解耦**: 模块间通过配置接口传递，避免直接依赖
- **一致性**: 统一的配置加载模式和验证机制
- **向后兼容**: 保持现有YAML配置文件格式不变
- **易测试**: 每个模块的配置可独立测试

### 1.2 核心原则

- **配置所有权**: 配置属于使用它的模块，不在中央配置包中定义
- **接口统一**: 所有配置都实现相同的接口模式
- **职责分离**: 配置加载工具只负责加载逻辑，不包含业务配置
- **渐进迁移**: 支持逐步从集中式配置迁移到模块化配置

## 2. 架构设计

### 2.1 配置架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Configuration System                    │
├─────────────────────────────────────────────────────────────┤
│  ┌────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Storage      │  │      Auth        │  │   Energy     │ │
│  │   Config       │  │     Config       │  │    Config    │ │
│  └────────┬───────┘  └────────┬────────┘  └──────┬───────┘ │
│           │                   │                  │         │
│  ┌────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Collector    │  │     Server      │  │   (More)     │ │
│  │   Config       │  │     Config       │  │   Modules    │ │
│  └────────┬───────┘  └────────┬────────┘  └──────────────┘ │
│           │                   │                            │
│           └─────────┬─────────┴────────────────────────────┘ │
│                     ▼                                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         Configuration Loader (pkgs/config)          │   │
│  │  - YAML parsing                                     │   │
│  │  - Environment variable binding                     │   │
│  │  - Validation utilities                            │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Configuration Sources                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐   │
│  │ YAML File   │  │ Environment │  │ Default Values      │   │
│  │ config.yaml │  │ Variables   │  │ (per module)        │   │
│  └─────────────┘  └─────────────┘  └─────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 配置包结构

```
internal/
├── pkgs/
│   └── config/                 # 配置加载工具包
│       ├── loader.go          # 配置加载器
│       ├── interface.go       # 配置接口定义
│       └── utils.go           # 配置工具函数
├── storage/
│   ├── config.go              # Storage模块配置
│   └── config_test.go         # Storage配置测试
├── auth/
│   ├── config.go              # Auth模块配置
│   └── config_test.go         # Auth配置测试
├── energy/
│   ├── config.go              # Energy模块配置
│   └── config_test.go         # Energy配置测试
├── collector/
│   ├── config.go              # Collector模块配置
│   └── config_test.go         # Collector配置测试
└── server/
    ├── config.go              # Server模块配置
    └── config_test.go         # Server配置测试
```

## 3. 配置接口设计

### 3.1 统一配置接口

```go
// Config 统一配置接口
type Config interface {
    // Validate 验证配置的有效性
    Validate() error

    // SetDefaults 设置默认值
    SetDefaults()

    // String 返回配置的字符串表示（敏感信息脱敏）
    String() string
}

// Provider 配置提供者接口
type Provider interface {
    // Load 从配置源加载配置
    Load() (Config, error)

    // LoadFromEnv 仅从环境变量加载配置
    LoadFromEnv() Config

    // GetConfigPath 获取配置文件路径
    GetConfigPath() string
}
```

### 3.2 配置加载器

```go
// Loader 配置加载器
type Loader struct {
    prefix    string           // 环境变量前缀
    v         *viper.Viper    // Viper实例
    configMap map[string]interface{} // 配置映射
}

// NewLoader 创建新的配置加载器
func NewLoader(prefix string) *Loader

// LoadModule 加载指定模块的配置
func (l *Loader) LoadModule(moduleName string, configStruct interface{}) error

// BindEnv 绑定环境变量
func (l *Loader) BindEnv(key string, envKeys ...string) error

// Validate 验证配置
func (l *Loader) Validate(config Config) error
```

## 4. 模块配置示例

### 4.1 Storage模块配置

```go
// internal/storage/config.go
package storage

import (
    "os"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

// Config Storage模块配置
type Config struct {
    DataDir         string      `yaml:"data_dir" env:"STORAGE_DATA_DIR" validate:"required"`
    CreateDir       bool        `yaml:"create_dir" env:"STORAGE_CREATE_DIR"`
    SyncWrite       bool        `yaml:"sync_write" env:"STORAGE_SYNC_WRITE"`
    FilePermissions os.FileMode `yaml:"file_permissions" env:"STORAGE_FILE_PERMISSIONS"`
    DirPermissions  os.FileMode `yaml:"dir_permissions" env:"STORAGE_DIR_PERMISSIONS"`
}

// NewConfig 创建新的Storage配置
func NewConfig(loader *config.Loader) (*Config, error) {
    cfg := &Config{}
    cfg.SetDefaults()

    if err := loader.LoadModule("storage", cfg); err != nil {
        return nil, err
    }

    return cfg, cfg.Validate()
}

// Validate 验证配置
func (c *Config) Validate() error {
    if c.DataDir == "" {
        return errors.New("storage: data_dir is required")
    }
    // 更多验证逻辑...
    return nil
}

// SetDefaults 设置默认值
func (c *Config) SetDefaults() {
    if c.DataDir == "" {
        c.DataDir = "./data"
    }
    if c.FilePermissions == 0 {
        c.FilePermissions = 0644
    }
    if c.DirPermissions == 0 {
        c.DirPermissions = 0755
    }
    if !c.CreateDir {
        c.CreateDir = true
    }
    if !c.SyncWrite {
        c.SyncWrite = true
    }
}

// String 返回配置的字符串表示
func (c *Config) String() string {
    return fmt.Sprintf("StorageConfig{DataDir: %s, CreateDir: %t, SyncWrite: %t}",
        c.DataDir, c.CreateDir, c.SyncWrite)
}
```

### 4.2 Auth模块配置

```go
// internal/auth/config.go
package auth

import (
    "time"
    "github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

// Config Auth模块配置
type Config struct {
    URL             string        `yaml:"url" env:"AUTH_URL" validate:"required,url"`
    Username        string        `yaml:"username" env:"AUTH_USERNAME" validate:"required"`
    Password        string        `yaml:"password" env:"AUTH_PASSWORD" validate:"required"`
    Timeout         time.Duration `yaml:"timeout" env:"AUTH_TIMEOUT"`
    RefreshThreshold time.Duration `yaml:"refresh_threshold" env:"AUTH_REFRESH_THRESHOLD"`
    SkipSSLVerify   bool          `yaml:"skip_ssl_verify" env:"AUTH_SKIP_SSL_VERIFY"`
    MaxRetries      int           `yaml:"max_retries" env:"AUTH_MAX_RETRIES"`
}

// NewConfig 创建新的Auth配置
func NewConfig(loader *config.Loader) (*Config, error) {
    cfg := &Config{}
    cfg.SetDefaults()

    if err := loader.LoadModule("auth", cfg); err != nil {
        return nil, err
    }

    return cfg, cfg.Validate()
}

// Validate 验证配置
func (c *Config) Validate() error {
    if c.URL == "" {
        return errors.New("auth: url is required")
    }
    if c.Username == "" {
        return errors.New("auth: username is required")
    }
    if c.Password == "" {
        return errors.New("auth: password is required")
    }
    if c.Timeout <= 0 {
        return errors.New("auth: timeout must be positive")
    }
    return nil
}

// SetDefaults 设置默认值
func (c *Config) SetDefaults() {
    if c.Timeout == 0 {
        c.Timeout = 30 * time.Second
    }
    if c.RefreshThreshold == 0 {
        c.RefreshThreshold = 5 * time.Minute
    }
    if c.MaxRetries == 0 {
        c.MaxRetries = 3
    }
}

// String 返回配置的字符串表示（敏感信息脱敏）
func (c *Config) String() string {
    return fmt.Sprintf("AuthConfig{URL: %s, Username: %s, Password: ***}",
        c.URL, c.Username)
}
```

## 5. 配置文件格式

### 5.1 YAML配置文件

配置文件保持现有格式，确保向后兼容：

```yaml
# config.yaml
storage:
  data_dir: "./data"
  create_dir: true
  sync_write: true
  file_permissions: 0644
  dir_permissions: 0755

auth:
  url: "https://192.168.1.100:8081"
  username: "admin"
  password: "password"
  timeout: 30s
  refresh_threshold: 5m
  skip_ssl_verify: true
  max_retries: 3

collector:
  interval: 5s
  timeout: 10s
  max_retries: 2
  batch_size: 10

energy:
  calculation_interval: 5s
  precision: 6
  unit: "wh"
  aggregation_window: 1m

server:
  host: "0.0.0.0"
  port: 9090
  read_timeout: 30s
  write_timeout: 30s
  enable_pprof: false
  cors_enabled: false
```

### 5.2 环境变量映射

环境变量命名保持现有规则：

```bash
# Storage配置
export WINPOWER_EXPORTER_STORAGE_DATA_DIR="/var/lib/winpower-exporter"
export WINPOWER_EXPORTER_STORAGE_SYNC_WRITE="true"

# Auth配置
export WINPOWER_EXPORTER_AUTH_URL="https://winpower.example.com"
export WINPOWER_EXPORTER_AUTH_USERNAME="admin"
export WINPOWER_EXPORTER_AUTH_PASSWORD="secret"

# Server配置
export WINPOWER_EXPORTER_SERVER_PORT="9090"
export WINPOWER_EXPORTER_SERVER_HOST="0.0.0.0"
```

## 6. 配置加载流程

### 6.1 初始化流程

```
┌────────────────────────────────────────────────┐
│              Application Startup               │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      1. Create Configuration Loader           │
│      - Set environment variable prefix         │
│      - Initialize Viper instance               │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      2. Load Configuration File               │
│      - Parse YAML file                         │
│      - Handle file not found gracefully       │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      3. Bind Environment Variables            │
│      - Apply module-specific prefixes          │
│      - Override file settings                  │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      4. Load Module Configurations            │
│      - storage → auth → energy → collector     │
│      - Each module validates its config        │
└─────────────────┬──────────────────────────────┘
                  │
                  ▼
┌────────────────────────────────────────────────┐
│      5. Initialize Services                   │
│      - Pass configs to service constructors    │
│      - Start the application                   │
└────────────────────────────────────────────────┘
```

### 6.2 依赖加载顺序

```go
func main() {
    // 1. 创建配置加载器
    loader := config.NewLoader("WINPOWER_EXPORTER")

    // 2. 按依赖顺序加载模块配置
    storageConfig, err := storage.NewConfig(loader)
    if err != nil {
        log.Fatalf("Failed to load storage config: %v", err)
    }

    authConfig, err := auth.NewConfig(loader)
    if err != nil {
        log.Fatalf("Failed to load auth config: %v", err)
    }

    energyConfig, err := energy.NewConfig(loader)
    if err != nil {
        log.Fatalf("Failed to load energy config: %v", err)
    }

    collectorConfig, err := collector.NewConfig(loader)
    if err != nil {
        log.Fatalf("Failed to load collector config: %v", err)
    }

    serverConfig, err := server.NewConfig(loader)
    if err != nil {
        log.Fatalf("Failed to load server config: %v", err)
    }

    // 3. 初始化服务（按依赖顺序）
    logger := log.New(logConfig)

    storageManager, err := storage.NewManager(storageConfig, logger)
    if err != nil {
        log.Fatalf("Failed to create storage manager: %v", err)
    }

    authService, err := auth.NewService(authConfig, logger)
    if err != nil {
        log.Fatalf("Failed to create auth service: %v", err)
    }

    // ... 继续初始化其他服务

    // 4. 启动应用
    app := &App{
        Storage:  storageManager,
        Auth:     authService,
        // ...
    }

    if err := app.Run(); err != nil {
        log.Fatalf("Failed to start application: %v", err)
    }
}
```

## 7. 配置验证和错误处理

### 7.1 验证策略

- **必需字段验证**: 检查所有必需参数是否提供
- **类型验证**: 确保配置值类型正确
- **范围验证**: 验证数值在合理范围内
- **格式验证**: 检查URL、文件路径等格式
- **依赖验证**: 验证相关配置项的一致性

### 7.2 错误处理

```go
// 配置加载错误的统一处理
type ConfigError struct {
    Module  string
    Field   string
    Value   interface{}
    Message string
    Cause   error
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config error in %s.%s: %s (value: %v)",
        e.Module, e.Field, e.Message, e.Value)
}

// 使用示例
if err := config.Validate(); err != nil {
    var configErr *ConfigError
    if errors.As(err, &configErr) {
        log.Fatalf("Configuration validation failed: %s", configErr.Error())
    }
    log.Fatalf("Unexpected configuration error: %v", err)
}
```

## 8. 测试策略

### 8.1 单元测试

每个模块的配置都需要完整的单元测试：

```go
func TestConfig(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() *Config
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            setup: func() *Config {
                cfg := &Config{DataDir: "/tmp", CreateDir: true}
                cfg.SetDefaults()
                return cfg
            },
            wantErr: false,
        },
        {
            name: "missing required field",
            setup: func() *Config {
                cfg := &Config{}
                cfg.SetDefaults()
                cfg.DataDir = "" // 清空必需字段
                return cfg
            },
            wantErr: true,
            errMsg:  "data_dir is required",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := tt.setup()
            err := cfg.Validate()

            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("Validate() error = %v, expected to contain %s", err, tt.errMsg)
            }
        })
    }
}
```

### 8.2 集成测试

测试配置加载的完整流程：

```go
func TestConfigLoading(t *testing.T) {
    // 创建临时配置文件
    tmpFile, err := ioutil.TempFile("", "test-config.*.yaml")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpFile.Name())

    // 写入测试配置
    configContent := `
storage:
  data_dir: "/tmp/test"
  sync_write: true
auth:
  url: "http://localhost:8080"
  username: "test"
  password: "test"
`

    if _, err := tmpFile.Write([]byte(configContent)); err != nil {
        t.Fatal(err)
    }
    tmpFile.Close()

    // 测试配置加载
    loader := config.NewLoader("TEST")
    loader.SetConfigFile(tmpFile.Name())

    storageConfig, err := storage.NewConfig(loader)
    if err != nil {
        t.Fatalf("Failed to load storage config: %v", err)
    }

    if storageConfig.DataDir != "/tmp/test" {
        t.Errorf("Expected DataDir to be '/tmp/test', got '%s'", storageConfig.DataDir)
    }

    if !storageConfig.SyncWrite {
        t.Errorf("Expected SyncWrite to be true, got false")
    }
}
```

## 9. 最佳实践

### 9.1 配置设计原则

- **最小权限**: 每个模块只能访问自己的配置
- **敏感信息保护**: 密码等敏感信息在日志中脱敏
- **默认值安全**: 默认值应该是最安全和最常用的选项
- **验证严格**: 在启动时验证所有配置，运行时不做假设

### 9.2 性能考虑

- **延迟加载**: 只在需要时加载配置
- **配置缓存**: 加载后的配置在内存中缓存
- **避免重复解析**: 配置文件只解析一次
- **轻量级验证**: 验证逻辑应该高效

### 9.3 运维友好

- **清晰的错误信息**: 配置错误时提供明确的错误位置和修复建议
- **环境变量文档**: 为每个环境变量提供清晰的文档说明
- **配置示例**: 提供完整的配置文件示例
- **迁移工具**: 如需迁移配置格式，提供自动化工具

## 10. 迁移指南

### 10.1 从集中式配置迁移

1. **分析现有配置**: 识别每个模块使用的配置项
2. **创建模块配置**: 在对应模块中创建配置结构
3. **更新构造函数**: 修改模块构造函数接收配置参数
4. **测试验证**: 确保迁移后功能正常
5. **清理旧配置**: 删除集中的配置结构

### 10.2 向后兼容性

- 保持YAML配置文件格式不变
- 保持环境变量命名规则不变
- 提供配置迁移工具（如需要）
- 在文档中清晰标注变更

这种模块化配置设计提供了更好的模块边界、更清晰的职责分离，同时保持了系统的可维护性和可测试性。