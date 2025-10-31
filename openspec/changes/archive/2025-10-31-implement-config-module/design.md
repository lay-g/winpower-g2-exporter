# Config 模块设计

## 架构概述

Config 模块采用分层架构设计，主要包含以下组件：

1. **ConfigManager 接口层**：提供统一的配置访问接口
2. **Loader 实现层**：负责多源配置加载和合并
3. **Validator 验证层**：提供可扩展的配置验证机制
4. **Defaults 默认值层**：管理所有配置项的默认值

## 接口设计

### ConfigValidator 接口

```go
type ConfigValidator interface {
    Validate() error
}
```

所有模块的配置结构体都需要实现此接口，提供模块特定的验证逻辑。

### ConfigManager 接口

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

## 配置加载机制

### 配置源优先级（从低到高）

1. **默认值**：代码中定义的默认配置
2. **配置文件**：从搜索路径找到的配置文件
3. **环境变量**：`WINPOWER_EXPORTER_` 前缀的环境变量
4. **命令行参数**：通过 pflag 定义的命令行参数

### 配置文件搜索路径

1. `./config.yaml`
2. `./config/config.yaml`
3. `$HOME/config/winpower-exporter/config.yaml`
4. `/etc/winpower-exporter/config.yaml`

## 配置结构设计

### 顶层 Config 结构体

```go
type Config struct {
    Server    *server.Config    `yaml:"server" mapstructure:"server"`
    WinPower  *winpower.Config  `yaml:"winpower" mapstructure:"winpower"`
    Storage   *storage.Config   `yaml:"storage" mapstructure:"storage"`
    Scheduler *scheduler.Config `yaml:"scheduler" mapstructure:"scheduler"`
    Logging   *log.Config       `yaml:"logging" mapstructure:"logging"`
}
```

## 错误处理策略

- **配置文件不存在**：记录警告，使用默认配置
- **配置文件格式错误**：记录错误，程序退出
- **配置验证失败**：记录错误，程序退出
- **环境变量格式错误**：记录警告，使用默认值

## 测试策略

- **单元测试**：测试各组件的独立功能
- **集成测试**：测试完整配置加载流程
- **错误场景测试**：测试各种错误情况的处理
- **Mock 测试**：隔离外部依赖进行测试

## 安全考虑

- 敏感信息优先从环境变量读取
- 避免在日志中输出敏感配置
- 配置文件权限检查
- 输入验证防止注入攻击