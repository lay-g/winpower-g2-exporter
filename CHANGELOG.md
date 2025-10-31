# 更新日志

所有重要的更改都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

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