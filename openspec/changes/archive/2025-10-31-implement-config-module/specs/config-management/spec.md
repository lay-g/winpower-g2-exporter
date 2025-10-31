# config-management Specification

## Purpose
Implement unified configuration management for WinPower G2 Exporter with multi-source configuration loading, validation, and type-safe access.

## ADDED Requirements

### Requirement: 提供类型安全的配置访问接口
The ConfigManager interface SHALL provide type-safe configuration access methods that support basic data types including string, integer, boolean, and string slice for getting and setting configuration values.

#### Scenario: 应用程序启动时配置加载
- **WHEN** 应用程序启动时
- **THEN** 系统需要加载完整配置
- **AND** 应该能够通过 ConfigManager 接口获取所有必要的配置项
- **AND** 配置值应该是正确的类型

#### Scenario: 运行时配置访问
- **WHEN** 应用程序运行时需要访问配置
- **THEN** 系统应该提供类型安全的配置访问方法
- **AND** 应该支持配置键的层级结构访问（如 server.port）

### Requirement: 支持配置文件、环境变量、命令行参数的优先级合并
The configuration loader SHALL support loading configuration from multiple sources and merge them according to predefined priority, where higher priority sources override lower priority sources.

#### Scenario: 配置优先级验证
- **WHEN** 存在多个配置源时
- **THEN** 系统应该按照正确的优先级合并配置
- **AND** 命令行参数 > 环境变量 > 配置文件 > 默认值
- **AND** 最终配置应该反映最高优先级的设置

#### Scenario: 部分配置缺失
- **WHEN** 某些配置项在部分源中缺失时
- **THEN** 系统应该能够从其他源获取配置
- **OR** 使用合理的默认值
- **AND** 不应该导致配置加载失败

### Requirement: 按优先级自动搜索配置文件位置
The configuration loader SHALL automatically search for configuration files in predefined path order and stop searching when the first existing configuration file is found.

#### Scenario: 配置文件位置搜索
- **WHEN** 应用程序启动时
- **THEN** 系统应该按照预定义顺序搜索配置文件
- **AND** 工作目录 -> 工作目录/config -> 用户配置目录 -> 系统配置目录
- **AND** 应该使用找到的第一个配置文件

#### Scenario: 配置文件不存在
- **WHEN** 所有搜索路径都不存在配置文件时
- **THEN** 系统应该记录警告信息
- **AND** 继续使用默认配置和环境变量
- **AND** 不应该导致程序启动失败

### Requirement: 提供可扩展的配置验证接口
The system SHALL provide a ConfigValidator interface that allows each module to implement custom configuration validation logic.

#### Scenario: 模块配置验证
- **WHEN** 配置加载完成后
- **THEN** 各模块应该能够验证自己的配置
- **AND** 验证失败时应该返回具体的错误信息
- **AND** 所有模块验证都通过才能继续启动

#### Scenario: 端口范围验证
- **WHEN** 验证服务器配置时
- **THEN** 系统应该检查端口号是否在有效范围（1-65535）
- **AND** 端口号无效时应该返回明确的错误信息
- **AND** 应该包含期望的范围信息

### Requirement: 为所有配置项提供合理的默认值
The system SHALL provide reasonable default values for all configuration items to ensure the application can run normally without configuration files.

#### Scenario: 默认配置启动
- **WHEN** 没有任何配置文件或环境变量时
- **THEN** 应用程序应该能够使用默认配置启动
- **AND** 所有必需的配置项都应该有默认值
- **AND** 默认值应该是生产可用的

#### Scenario: 默认值合理性
- **WHEN** 使用默认配置时
- **THEN** 服务器端口应该是9090
- **AND** 日志级别应该是info
- **AND** 调度器间隔应该是5秒
- **AND** 所有默认值都应该符合项目约定

### Requirement: 支持WINPOWER_EXPORTER_前缀的环境变量
The system SHALL support using WINPOWER_EXPORTER_ prefixed environment variables to override configuration, where dots in configuration keys are converted to underscores.

#### Scenario: 环境变量配置覆盖
- **WHEN** 设置了环境变量 WINPOWER_EXPORTER_SERVER_PORT=8080 时
- **THEN** 系统应该能够正确解析并覆盖服务器端口配置
- **AND** 配置键 server.port 应该映射到环境变量 SERVER_PORT
- **AND** 应该支持所有配置项的环境变量覆盖

#### Scenario: 环境变量格式转换
- **WHEN** 配置键包含点号时（如 scheduler.collection_interval）
- **THEN** 系统应该将其转换为下划线格式（SCHEDULER_COLLECTION_INTERVAL）
- **AND** 应该正确处理嵌套结构的映射

### Requirement: 支持YAML格式的配置文件
The configuration loader SHALL support YAML format configuration files and be able to correctly parse nested structures and complex data types.

#### Scenario: YAML配置文件解析
- **WHEN** 提供YAML格式的配置文件时
- **THEN** 系统应该能够正确解析所有配置项
- **AND** 包括嵌套结构和数组类型
- **AND** 应该处理格式错误的情况

#### Scenario: 配置文件格式验证
- **WHEN** 配置文件格式不正确时
- **THEN** 系统应该检测到格式错误
- **AND** 返回具体的错误信息和位置
- **AND** 不应该导致程序崩溃

### Requirement: 支持通过命令行参数覆盖配置
The system SHALL support overriding configuration through command line arguments, where the argument format uses module.option pattern.

#### Scenario: 命令行参数配置
- **WHEN** 使用 --server.port 8080 参数启动时
- **THEN** 系统应该能够正确解析并覆盖服务器端口
- **AND** 命令行参数应该具有最高优先级
- **AND** 应该支持所有主要配置项

#### Scenario: 帮助信息显示
- **WHEN** 使用 --help 参数时
- **THEN** 系统应该显示所有可用的命令行参数
- **AND** 包括参数说明和默认值
- **AND** 格式应该清晰易读

### Requirement: 提供完善的错误处理和日志记录
The system SHALL provide detailed error information and logging during configuration loading to help diagnose configuration issues.

#### Scenario: 配置加载失败处理
- **WHEN** 配置加载失败时
- **THEN** 系统应该记录详细的错误信息
- **AND** 包括失败的原因和位置
- **AND** 应该提供解决建议

#### Scenario: 配置加载日志
- **WHEN** 配置加载成功时
- **THEN** 系统应该记录配置加载的过程
- **AND** 包括使用的配置文件路径
- **AND** 以及环境变量和命令行参数的应用情况

### Requirement: 确保敏感信息安全处理
The system SHALL securely handle sensitive configuration information and avoid outputting passwords and other sensitive data in logs.

#### Scenario: 敏感信息脱敏
- **WHEN** 记录配置相关的日志时
- **THEN** 系统应该检测并脱敏敏感信息
- **AND** 如密码、token等字段
- **AND** 应该用占位符替代实际值

#### Scenario: 配置文件权限检查
- **WHEN** 读取配置文件时
- **THEN** 系统应该检查文件权限
- **AND** 对于权限过宽的文件应该发出警告
- **AND** 建议使用更安全的权限设置