# Config 模块实现设计

## 架构概述

本设计文档详细说明了如何实现统一的配置管理系统，该系统基于 `docs/design/config.md` 的设计原则，采用模块化、解耦的架构。

## 系统架构

### 核心组件

```
┌─────────────────────────────────────────────────────────────┐
│                    Configuration System                    │
├─────────────────────────────────────────────────────────────┤
│  ┌────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Storage      │  │     WinPower    │  │   Energy     │ │
│  │   Config       │  │     Config       │  │    Config    │ │
│  └────────┬───────┘  └────────┬────────┘  └──────┬───────┘ │
│           │                   │                  │         │
│  ┌────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Server       │  │    Scheduler    │  │   Metrics    │ │
│  │   Config       │  │     Config       │  │    Config    │ │
│  └────────┬───────┘  └────────┬────────┘  └──────┬───────┘ │
│           │                   │                  │         │
│           └─────────┬─────────┴────────────────────┘ │
│                     ▼                                      │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         Configuration Loader (pkgs/config)          │   │
│  │  - YAML parsing                                     │   │
│  │  - Environment variable binding                     │   │
│  │  - Validation utilities                            │   │
│  │  - Caching layer                                   │   │
│  │  - Error handling                                  │   │
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

## 接口设计

### 1. 核心接口

```go
// Config 统一配置接口
type Config interface {
    // Validate 验证配置的有效性
    Validate() error

    // SetDefaults 设置默认值
    SetDefaults()

    // String 返回配置的字符串表示（敏感信息脱敏）
    String() string

    // Clone 创建配置的深拷贝
    Clone() Config
}

// Provider 配置提供者接口
type Provider interface {
    // Load 从配置源加载配置
    Load() (Config, error)

    // LoadFromEnv 仅从环境变量加载配置
    LoadFromEnv() Config

    // GetConfigPath 获取配置文件路径
    GetConfigPath() string

    // SetConfigPath 设置配置文件路径
    SetConfigPath(path string)
}

// Loader 配置加载器接口
type Loader interface {
    // LoadModule 加载指定模块的配置
    LoadModule(moduleName string, configStruct interface{}) error

    // BindEnv 绑定环境变量
    BindEnv(key string, envKeys ...string) error

    // Validate 验证配置
    Validate(config Config) error

    // Watch 监控配置文件变更
    Watch(callback func(Config)) error

    // StopWatching 停止监控
    StopWatching() error
}
```

### 2. 错误处理

```go
// ConfigError 配置错误类型
type ConfigError struct {
    Module  string
    Field   string
    Value   interface{}
    Message string
    Code    string
    Cause   error
}

func (e *ConfigError) Error() string
func (e *ConfigError) Unwrap() error

// ValidationError 验证错误类型
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
    Code    string
}

// 错误代码常量
const (
    ErrCodeRequiredField   = "REQUIRED_FIELD"
    ErrCodeInvalidType     = "INVALID_TYPE"
    ErrCodeInvalidRange    = "INVALID_RANGE"
    ErrCodeInvalidFormat   = "INVALID_FORMAT"
    ErrCodeFileNotFound    = "FILE_NOT_FOUND"
    ErrCodeParseError      = "PARSE_ERROR"
    ErrCodePermissionError = "PERMISSION_ERROR"
)
```

## 实现细节

### 1. 配置加载器实现

```go
type loader struct {
    prefix       string
    viper        *viper.Viper
    configPath   string
    watchers     map[string]*fsnotify.Watcher
    mu           sync.RWMutex
    cache        map[string]interface{}
    cacheEnabled bool
}

func NewLoader(prefix string) Loader {
    v := viper.New()
    v.SetEnvPrefix(prefix)
    v.AutomaticEnv(envReplacer)

    return &loader{
        prefix:       prefix,
        viper:        v,
        watchers:     make(map[string]*fsnotify.Watcher),
        cache:        make(map[string]interface{}),
        cacheEnabled: true,
    }
}

func (l *loader) LoadModule(moduleName string, configStruct interface{}) error {
    l.mu.RLock()
    if cached, exists := l.cache[moduleName]; exists && l.cacheEnabled {
        l.mu.RUnlock()
        return l.copyConfig(cached, configStruct)
    }
    l.mu.RUnlock()

    // 从YAML文件加载
    if l.configPath != "" {
        if err := l.loadFromFile(moduleName, configStruct); err != nil {
            return err
        }
    }

    // 绑定环境变量
    if err := l.bindEnvironmentVariables(moduleName, configStruct); err != nil {
        return err
    }

    // 设置默认值
    l.setDefaults(moduleName, configStruct)

    // 验证配置
    if config, ok := configStruct.(Config); ok {
        if err := config.Validate(); err != nil {
            return err
        }
    }

    // 缓存配置
    l.mu.Lock()
    l.cache[moduleName] = configStruct
    l.mu.Unlock()

    return nil
}
```

### 2. 环境变量绑定

```go
func (l *loader) bindEnvironmentVariables(moduleName string, configStruct interface{}) error {
    t := reflect.TypeOf(configStruct).Elem()
    v := reflect.ValueOf(configStruct).Elem()

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fieldValue := v.Field(i)

        if !fieldValue.CanSet() {
            continue
        }

        envKey := l.getEnvKey(moduleName, field)
        if envValue := os.Getenv(envKey); envValue != "" {
            if err := l.setFieldValue(fieldValue, envValue); err != nil {
                return fmt.Errorf("failed to set field %s: %w", field.Name, err)
            }
        }
    }

    return nil
}

func (l *loader) getEnvKey(moduleName string, field reflect.StructField) string {
    if envTag := field.Tag.Get("env"); envTag != "" {
        return fmt.Sprintf("%s_%s", l.prefix, envTag)
    }
    return fmt.Sprintf("%s_%s_%s", l.prefix, strings.ToUpper(moduleName), strings.ToUpper(field.Name))
}
```

### 3. 配置文件监控

```go
func (l *loader) Watch(callback func(Config)) error {
    if l.configPath == "" {
        return fmt.Errorf("no config file path set")
    }

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return err
    }

    if err := watcher.Add(l.configPath); err != nil {
        watcher.Close()
        return err
    }

    l.watchers[l.configPath] = watcher

    go func() {
        defer watcher.Close()
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                if event.Op&fsnotify.Write == fsnotify.Write {
                    l.handleConfigChange(callback)
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Printf("config watcher error: %v", err)
            }
        }
    }()

    return nil
}

func (l *loader) handleConfigChange(callback func(Config)) {
    // 清空缓存
    l.mu.Lock()
    l.cache = make(map[string]interface{})
    l.mu.Unlock()

    // 重新加载配置
    // 这里可以触发回调通知相关模块
    log.Println("Configuration file changed, reloading...")
}
```

## 模块配置适配

### 1. Storage 模块配置适配

```go
type Config struct {
    DataDir         string      `yaml:"data_dir" json:"data_dir" env:"STORAGE_DATA_DIR"`
    FilePermissions os.FileMode `yaml:"file_permissions" json:"file_permissions" env:"STORAGE_FILE_PERMISSIONS"`
    DirPermissions  os.FileMode `yaml:"dir_permissions" json:"dir_permissions" env:"STORAGE_DIR_PERMISSIONS"`
    SyncWrite       bool        `yaml:"sync_write" json:"sync_write" env:"STORAGE_SYNC_WRITE"`
    CreateDir       bool        `yaml:"create_dir" json:"create_dir" env:"STORAGE_CREATE_DIR"`
}

func (c *Config) Validate() error {
    if c.DataDir == "" {
        return &ConfigError{
            Module:  "storage",
            Field:   "data_dir",
            Value:   c.DataDir,
            Message: "data directory is required",
            Code:    ErrCodeRequiredField,
        }
    }

    // 验证目录可访问性
    if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
        if c.CreateDir {
            if err := os.MkdirAll(c.DataDir, c.DirPermissions); err != nil {
                return &ConfigError{
                    Module:  "storage",
                    Field:   "data_dir",
                    Value:   c.DataDir,
                    Message: fmt.Sprintf("failed to create data directory: %v", err),
                    Code:    ErrCodePermissionError,
                    Cause:   err,
                }
            }
        } else {
            return &ConfigError{
                Module:  "storage",
                Field:   "data_dir",
                Value:   c.DataDir,
                Message: "data directory does not exist and create_dir is false",
                Code:    ErrCodeFileNotFound,
            }
        }
    }

    return nil
}

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
}

func (c *Config) String() string {
    return fmt.Sprintf("StorageConfig{DataDir: %s, FilePermissions: %o, DirPermissions: %o, SyncWrite: %t, CreateDir: %t}",
        c.DataDir, c.FilePermissions, c.DirPermissions, c.SyncWrite, c.CreateDir)
}

func (c *Config) Clone() Config {
    return &Config{
        DataDir:         c.DataDir,
        FilePermissions: c.FilePermissions,
        DirPermissions:  c.DirPermissions,
        SyncWrite:       c.SyncWrite,
        CreateDir:       c.CreateDir,
    }
}
```

## 性能优化

### 1. 配置缓存

- 使用内存缓存避免重复文件解析
- 支持缓存失效和更新
- 使用读写锁保证并发安全

### 2. 延迟加载

- 只在需要时加载特定模块配置
- 支持配置预加载以提高响应速度
- 使用懒加载减少启动时间

### 3. 批量操作

- 支持批量加载多个模块配置
- 减少文件系统访问次数
- 提高配置初始化效率

## 安全考虑

### 1. 敏感信息保护

- 密码等敏感字段在日志中自动脱敏
- 支持敏感字段标记和处理
- 防止敏感信息意外泄露

### 2. 配置验证

- 严格的类型和格式验证
- 防止恶意配置注入
- 支持配置白名单机制

### 3. 权限控制

- 验证配置文件访问权限
- 检查目录创建权限
- 防止权限提升攻击

## 错误处理策略

### 1. 分层错误处理

- 底层错误包装为结构化错误
- 提供详细的错误上下文信息
- 支持错误链追踪

### 2. 错误恢复

- 支持配置加载失败时的降级策略
- 提供默认配置作为后备方案
- 记录详细的错误日志

### 3. 用户友好错误

- 提供清晰的错误修复建议
- 包含配置示例和文档链接
- 支持配置诊断工具

## 测试策略

### 1. 单元测试

- 每个配置模块独立测试
- 使用Mock隔离外部依赖
- 覆盖正常和异常场景

### 2. 集成测试

- 测试完整配置加载流程
- 验证模块间配置交互
- 测试配置文件变更响应

### 3. 性能测试

- 配置加载性能基准测试
- 并发访问安全性测试
- 内存使用效率测试

这个设计确保了配置系统的模块化、可扩展性和安全性，同时保持了良好的性能和用户体验。