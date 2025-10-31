# 更新日志

所有重要的更改都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [0.1.5] - 2025-10-31

### 改进
- 改进 Docker 镜像标签生成，使用语义化版本模式
- 添加 major.minor 标签模式用于更好的版本管理
- 确保 Docker 镜像遵循正确的语义化版本控制

## [0.1.4] - 2025-10-31

### 修复
- 修复 Makefile 中的 release 目标，移除不必要的 clean 依赖
- 改进命令行标志绑定的错误处理，避免静默失败

### 改进
- 优化 build-all 命令的注释说明，明确不适用于 Release 场景
- 增强配置加载的健壮性，提供更好的错误反馈

## [0.1.3] - 2025-10-31

### 修复
- 解决配置加载中环境变量优先级问题，修复命令行标志空值覆盖环境变量的错误

## [0.1.2] - 2025-10-31

### 修复
- 修复 Docker 容器启动命令，使用正确的 server 子命令
- 改进发布流程，确保在打包前清理构建产物

### 改进
- 优化 Makefile 中的 release 目标依赖关系

## [0.1.1] - 2025-10-31

### 修复
- 修复 Docker 容器启动命令，使用明确的 serve 子命令
- 更新 Docker Compose 镜像引用为 GitHub Container Registry
- 清理 WinPower 客户端代码中的空行

### 改进
- 添加 release 和 tag 目标到 .PHONY 声明
- 完善 Docker 相关配置

## [0.1.0] - 2025-10-31

### 新增
- 初始版本发布
- WinPower G2 设备数据采集功能
- 电能计算和累计功能
- Prometheus 指标导出
- HTTP 服务端点（/metrics, /health）
- 配置文件管理
- 调度器功能（5秒周期）
- 存储模块（设备级电能数据持久化）
- 日志模块（结构化日志）
- 支持 SSL 证书验证跳过
- 多平台构建支持（Linux, macOS, Windows）

### 技术栈
- Go 1.25+
- Gin Web 框架
- Zap 日志库
- Prometheus 客户端库

### 架构
- 测试驱动开发（TDD）
- 模块化设计
- 接口抽象
- 依赖注入

[Unreleased]: https://github.com/lay-g/winpower-g2-exporter/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/lay-g/winpower-g2-exporter/releases/tag/v0.1.0