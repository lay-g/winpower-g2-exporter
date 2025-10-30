# 实现电能计算模块

## Why

根据项目设计文档 `docs/design/energy.md`，需要实现一个极简单线程架构的电能计算模块，用于从功率数据计算累计电能消耗。该模块采用全局锁确保串行执行，专注于为UPS设备提供精确的电能累计计算功能。

## What Changes

- 实现EnergyService核心服务类，提供电能计算和查询接口（Calculate、Get、GetStats）
- 实现全局锁机制（sync.RWMutex）确保所有计算操作串行执行，避免数据竞争
- 集成storage模块进行电能数据的持久化存储，使用标准PowerData结构
- 实现精确的电能累计计算（Wh = W × 时间间隔），支持正负功率和零功率处理
- 提供简单统计信息（总计算次数、错误次数、平均执行时间等）用于监控和调试
- 确保时间戳精度为毫秒级，电能精度为0.01Wh
- 实现单元测试和集成测试确保功能正确性
- 明确功率数据来源约定（使用loadTotalWatt总负载有功功率）

## Impact

- **Affected specs**: 新增 `energy-module` 规范
- **Affected code**:
  - `internal/energy/` - 新增电能计算模块
- **Dependencies**: 依赖storage模块接口
- **Testing**: 需要完整的单元测试和集成测试覆盖