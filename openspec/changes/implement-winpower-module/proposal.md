# 实现WinPower模块提案

## Why

当前项目缺少与WinPower G2系统交互的核心模块，需要实现完整的认证管理、数据采集和解析功能，为整个导出器提供可靠的数据源。

## What Changes

- 实现WinPower客户端，统一管理认证和数据采集
- 实现HTTP客户端，支持SSL/TLS配置和连接复用
- 实现Token管理器，自动处理登录、缓存和刷新
- 实现数据解析器，解析设备数据并标准化输出
- 实现与Energy模块的集成，在采集成功后触发电能计算
- 添加配置管理，支持URL、认证信息和超时等参数配置
- 添加错误处理和日志记录，确保系统稳定运行

## Impact

- **新增能力**: WinPower数据采集和认证管理能力
- **受影响规格**: 需要新增winpower-module规格
- **受影响代码**:
  - 新增`internal/winpower/`模块
  - 与现有`internal/energy/`模块集成
  - 与将来的`scheduler`和`metrics`模块集成
- **配置变更**: 新增winpower配置段，包含认证信息和API配置