# cli-interface Specification

## Purpose
TBD - created by archiving change implement-cmd-module. Update Purpose after archive.
## Requirements
### Requirement: CLI-001 - Root Command Structure
应用程序 MUST 实现根命令结构，提供统一的命令行入口点和默认行为。
**Description**: 实现根命令结构，提供统一的命令行入口点和默认行为。

#### Scenario: 根命令默认行为
**Given** 用户执行 `./winpower-g2-exporter` 命令
**When** 没有指定任何子命令
**Then** 应用程序应该显示帮助信息而不是错误
**And** 根命令必须使用 winpower-g2-exporter 作为程序名
**And** 必须支持持久化参数（--config, --verbose）
**And** 必须包含程序简短描述和详细说明

### Requirement: CLI-002 - Server Subcommand
应用程序 MUST 实现 server 子命令，用于启动 HTTP 服务器。
**Description**: 实现 server 子命令，用于启动 HTTP 服务器。

#### Scenario: 服务器启动
**Given** 用户执行 `./winpower-g2-exporter server --config config.yaml --server.port 8080` 命令
**When** server 子命令执行时
**Then** 应用程序必须加载指定配置文件
**And** 必须在命令行指定的端口（8080）启动 HTTP 服务器，覆盖配置文件中的端口设置
**And** 必须提供优雅关闭机制
**And** 必须正确初始化所有模块（按依赖顺序）
**And** 必须处理 SIGINT 和 SIGTERM 信号
**And** 启动失败时必须提供清晰的错误信息

#### Scenario: 默认端口启动
**Given** 用户执行 `./winpower-g2-exporter server --config config.yaml` 命令
**When** server 子命令执行时，没有指定 --server.port 参数
**Then** 应用程序必须使用默认端口9090启动 HTTP 服务器
**And** 命令行参数的默认值必须为9090

### Requirement: CLI-003 - Version Subcommand
应用程序 MUST 实现 version 子命令，显示应用程序版本信息。
**Description**: 实现 version 子命令，显示应用程序版本信息。

#### Scenario: 版本信息显示 - JSON格式
**Given** 用户执行 `./winpower-g2-exporter version --format json` 命令
**When** version 子命令执行时
**Then** 应用程序必须以 JSON 格式输出版本信息
**And** 必须支持 --format 参数（text/json）
**And** 版本号必须从编译时注入的变量读取
**And** JSON 输出格式必须正确且可解析

#### Scenario: 版本信息显示 - 默认文本格式
**Given** 用户执行 `./winpower-g2-exporter version` 命令，没有指定 --format 参数
**When** version 子命令执行时
**Then** 应用程序必须默认使用 text 格式输出版本信息
**And** 必须显示版本号、Go 版本、编译时间、Commit ID、平台信息
**And** 版本号必须从编译时注入的变量读取

### Requirement: CLI-004 - Help Subcommand
应用程序 MUST 实现 help 子命令，显示命令帮助信息。
**Description**: 实现 help 子命令，显示命令帮助信息。

#### Scenario: 子命令帮助显示
**Given** 用户执行 `./winpower-g2-exporter help server` 命令
**When** help 子命令执行时
**Then** 应用程序必须显示 server 子命令的详细帮助信息
**And** 必须支持显示根命令帮助
**And** 必须支持显示子命令帮助
**And** 帮助信息必须包含用法示例
**And** 帮助信息格式必须清晰易读

### Requirement: CLI-005 - Configuration Binding
应用程序 MUST 实现配置文件和环境变量的绑定机制。
**Description**: 实现配置文件和环境变量的绑定机制。

#### Scenario: 环境变量配置覆盖
**Given** 用户设置了环境变量 `WINPOWER_EXPORTER_SERVER_PORT=8080`
**When** 用户执行 `./winpower-g2-exporter server` 命令
**Then** 服务器必须在 8080 端口启动，覆盖配置文件中的默认端口设置
**And** 命令行 --server.port 参数的优先级必须高于环境变量
**And** 应用程序必须支持从指定路径加载配置文件
**And** 必须支持在多个路径搜索配置文件
**And** 环境变量必须可以覆盖配置文件中的值
**And** 环境变量必须使用 WINPOWER_EXPORTER_ 前缀
**And** 配置绑定失败时必须提供清晰的错误信息

### Requirement: CLI-006 - Module Initialization Order
应用程序 MUST 确保所有模块按正确的依赖顺序初始化。
**Description**: 确保所有模块按正确的依赖顺序初始化。

#### Scenario: 模块启动顺序
**Given** 应用程序启动时
**When** 模块初始化过程开始
**Then** 应用程序必须按以下顺序初始化模块：config → logging → storage → winpower → energy → collector → metrics → server → scheduler
**And** 必须严格按照依赖顺序初始化模块
**And** 每个模块初始化失败时必须提供清晰错误信息
**And** 必须支持模块初始化过程的日志记录
**And** 初始化失败时必须正确清理已初始化的模块

### Requirement: CLI-007 - Graceful Shutdown
应用程序 MUST 实现优雅关闭机制，确保所有模块正确清理资源。
**Description**: 实现优雅关闭机制，确保所有模块正确清理资源。

#### Scenario: 信号处理和优雅关闭
**Given** 应用程序正在运行
**When** 用户发送 SIGINT 信号（Ctrl+C）
**Then** 应用程序必须停止接收新的请求
**And** 必须完成正在处理的请求
**And** 必须按相反顺序关闭所有模块
**And** 必须保存必要的状态信息
**And** 必须正确处理 SIGINT 和 SIGTERM 信号
**And** 必须按与初始化相反的顺序关闭模块
**And** 必须给予正在进行的操作足够时间完成
**And** 关闭过程中必须记录适当日志

### Requirement: CLI-008 - Build-time Information Injection
应用程序 MUST 支持编译时注入版本信息。
**Description**: 支持编译时注入版本信息。

#### Scenario: 构建时版本信息注入
**Given** 构建脚本从 VERSION 文件读取版本号 `v1.0.0`
**When** 执行构建过程时
**Then** 应用程序必须将版本信息注入到二进制文件中
**And** 必须支持从 VERSION 文件读取版本号
**And** 必须支持注入构建时间（RFC3339 格式）
**And** 必须支持注入 Git Commit ID
**And** 必须提供默认值用于开发环境
**And** Makefile 必须集成构建脚本

