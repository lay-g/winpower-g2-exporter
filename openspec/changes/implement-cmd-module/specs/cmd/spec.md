## ADDED Requirements

### Requirement: 主程序入口点和子命令结构
系统 SHALL 提供主程序入口点，支持server、version、help子命令，以及HTTP服务器启动、配置加载和模块初始化。

#### Scenario: 默认启动HTTP服务器
- **WHEN** 用户执行 `exporter`
- **THEN** 系统默认执行server子命令，加载配置文件并启动HTTP服务器
- **AND** 返回启动成功的日志信息

#### Scenario: 显式启动HTTP服务器
- **WHEN** 用户执行 `exporter server --config config.yaml --port 9090`
- **THEN** 系统加载配置文件并启动HTTP服务器监听指定端口
- **AND** 返回启动成功的日志信息

#### Scenario: 显示版本信息
- **WHEN** 用户执行 `exporter version`
- **THEN** 系统显示应用程序版本、构建信息和Git提交哈希
- **AND** 退出程序

#### Scenario: 显示帮助信息
- **WHEN** 用户执行 `exporter help` 或 `exporter --help`
- **THEN** 系统显示完整的子命令列表和用法说明
- **AND** 包含每个子命令的选项和配置文件格式示例

#### Scenario: 配置文件验证失败
- **WHEN** 用户提供无效的配置文件
- **THEN** 系统显示详细的配置错误信息
- **AND** 退出程序并返回非零状态码

#### Scenario: 优雅关闭处理
- **WHEN** 系统接收到SIGTERM信号
- **THEN** 系统优雅关闭所有服务和连接
- **AND** 完成资源清理后退出

### Requirement: 命令行参数解析
系统 SHALL 支持完整的命令行参数解析，包括配置文件、端口、日志级别等选项。

#### Scenario: 使用命令行参数覆盖配置
- **WHEN** 用户指定 `exporter server --log-level debug`
- **THEN** 系统使用debug级别覆盖配置文件中的日志级别
- **AND** 其他配置项保持不变

#### Scenario: 子命令参数验证
- **WHEN** 用户执行 `exporter server --invalid-option`
- **THEN** 系统显示无效选项的错误信息
- **AND** 显示server子命令的正确用法

### Requirement: 模块依赖初始化
系统 SHALL 按照正确的依赖顺序初始化各个模块，确保模块间的协作。

#### Scenario: 模块初始化失败处理
- **WHEN** 某个模块初始化失败（如无法连接WinPower服务）
- **THEN** 系统记录详细的错误信息
- **AND** 优雅关闭已初始化的模块
- **AND** 退出程序并返回错误状态码

#### Scenario: 模块健康检查
- **WHEN** 系统完成所有模块初始化
- **THEN** 系统执行健康检查确保所有模块正常工作
- **AND** 在日志中报告初始化状态

### Requirement: 多平台构建支持
系统 SHALL 支持多平台构建，包括Linux、Windows和macOS。

#### Scenario: 跨平台构建
- **WHEN** 执行 `make build-all`
- **THEN** 系统生成名为exporter的多个平台可执行文件
- **AND** 每个可执行文件都包含正确的平台信息

#### Scenario: Docker容器化构建
- **WHEN** 执行 `make docker-build`
- **THEN** 系统构建包含exporter主程序的Docker镜像
- **AND** 镜像大小优化并包含必要的基础依赖