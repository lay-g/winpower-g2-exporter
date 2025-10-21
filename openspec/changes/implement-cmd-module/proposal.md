## Why

为了实现完整的WinPower G2 Exporter应用程序，需要建立命令行入口点，实现主HTTP服务器程序，确保所有模块能够正确初始化和协作。

详细设计请参考 [docs/design/cmd.md](../../../docs/design/cmd.md)

## What Changes

- 添加主程序入口点 (cmd/exporter/)，实现exporter二进制程序
- 实现统一的命令行参数解析和配置加载机制，支持以下子命令：
  - `server`: 启动HTTP服务器（默认子命令）
  - `version`: 显示版本信息
  - `help`: 显示帮助信息
- 按照TDD原则，先编写测试再实现功能代码

## Impact

- Affected specs: cmd (新增)
- Affected code:
  - cmd/exporter/ - 主程序入口
  - internal/cmd/ - 命令行工具共享实现
  - Makefile - 构建脚本更新
  - 根目录go.mod - 依赖管理