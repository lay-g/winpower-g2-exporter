## MODIFIED Requirements

### Requirement: 主程序入口点和子命令结构
系统 SHALL 提供主程序入口点，支持server、version、help子命令，以及HTTP服务器启动、配置加载和模块初始化。

#### MODIFIED Scenario: 默认显示帮助信息
- **WHEN** 用户执行 `exporter`
- **THEN** 系统默认显示帮助信息，包含所有可用子命令和用法说明
- **AND** 不会启动HTTP服务器
- **AND** 帮助信息包含如何启动服务器的指导

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