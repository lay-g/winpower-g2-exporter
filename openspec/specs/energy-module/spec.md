# energy-module Specification

## Purpose
TBD - created by archiving change implement-energy-module. Update Purpose after archive.
## Requirements
### Requirement: 电能计算接口
Energy模块 SHALL提供统一的电能计算接口，支持基于功率数据计算累计电能消耗。

#### Scenario: 成功计算电能
- **WHEN** 调用者提供设备ID和当前功率值
- **THEN** 系统返回计算后的累计电能值（Wh）
- **AND** 计算结果被持久化存储

#### Scenario: 获取电能数据
- **WHEN** 调用者提供设备ID查询电能数据
- **THEN** 系统返回该设备的最新累计电能值
- **AND** 数据来源于持久化存储

### Requirement: 数据一致性保证
Energy模块 SHALL通过全局锁机制确保所有计算操作的串行执行，避免数据竞争。

#### Scenario: 并发计算安全
- **WHEN** 多个goroutine同时调用电能计算接口
- **THEN** 所有计算操作串行执行，无数据竞争
- **AND** 计算结果保持一致性

#### Scenario: 数据完整性
- **WHEN** 计算过程中发生系统故障
- **THEN** 已完成的计算结果不会丢失
- **AND** 系统重启后能正确恢复历史数据

### Requirement: 电能累计计算
Energy模块 SHALL基于功率数据进行积分计算，实现准确的电能累计，使用精确的时间间隔和数值精度。

#### Scenario: 正功率累计
- **WHEN** 输入功率为正值时
- **THEN** 累计电能值单调递增
- **AND** 计算公式为：累计电能(Wh) = 历史电能(Wh) + (功率(W) × 时间间隔(h))
- **AND** 时间戳精度为毫秒级，电能精度为0.01Wh

#### Scenario: 负功率处理
- **WHEN** 输入功率为负值时
- **THEN** 累计电能值递减
- **AND** 表示净能量减少
- **AND** 计算公式同样适用：累计电能 = 历史电能 + (负功率 × 时间间隔)

#### Scenario: 零功率处理
- **WHEN** 输入功率为零时
- **THEN** 累计电能值保持不变
- **AND** 时间线正常推进
- **AND** 时间戳更新以保持数据连续性

#### Scenario: 功率数据来源
- **WHEN** Collector模块调用能量计算时
- **THEN** 功率值为设备实时总负载有功功率 `loadTotalWatt`（单位 `W`）
- **AND** 不使用视在功率 `loadTotalVa` 或单相功率直接参与能耗计算

### Requirement: 存储集成
Energy模块 SHALL完全依赖storage模块进行数据持久化，不直接操作文件系统，使用标准化的数据结构。

#### Scenario: 数据持久化
- **WHEN** 电能计算完成后
- **THEN** 计算结果通过storage接口保存
- **AND** 数据结构包含毫秒时间戳和累计电能值(Wh)
- **AND** 时间戳由Energy模块维护，Collector不直接设置

#### Scenario: 历史数据加载
- **WHEN** 开始新的电能计算时
- **THEN** 从storage加载历史数据
- **AND** 基于历史数据进行累计计算
- **AND** 处理首次访问时文件不存在的情况

#### Scenario: 数据结构标准
- **WHEN** 与storage模块交互时
- **THEN** 使用标准的PowerData结构：`{Timestamp: int64, EnergyWH: float64}`
- **AND** 时间戳为毫秒级，能量单位为Wh

### Requirement: 统计信息
Energy模块 SHALL提供基础统计信息用于监控和调试。

#### Scenario: 计算统计
- **WHEN** 查询统计信息时
- **THEN** 返回总计算次数、错误次数、平均执行时间等
- **AND** 统计信息线程安全

### Requirement: 错误处理
Energy模块 SHALL提供完善的错误处理机制，确保系统稳定性。

#### Scenario: 存储错误处理
- **WHEN** storage操作失败时
- **THEN** 返回明确的错误信息
- **AND** 不影响其他设备的计算

#### Scenario: 输入验证
- **WHEN** 输入参数无效时（如空设备ID）
- **THEN** 返回参数错误
- **AND** 记录错误日志

### Requirement: 性能要求
Energy模块 SHALL满足性能要求，支持实时电能计算。

#### Scenario: 计算性能
- **WHEN** 执行电能计算时
- **THEN** 单次计算时间不超过100ms
- **AND** 支持至少20个设备的并发访问

### Requirement: 日志记录
Energy模块 SHALL提供结构化日志记录，支持问题排查和性能分析。

#### Scenario: 计算日志
- **WHEN** 执行电能计算时
- **THEN** 记录设备ID、功率值、计算结果、执行时间
- **AND** 错误情况详细记录

