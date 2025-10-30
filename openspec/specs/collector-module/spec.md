# collector-module Specification

## Purpose
TBD - created by archiving change implement-collector-module. Update Purpose after archive.
## Requirements
### Requirement: 数据采集协调功能
Collector模块 SHALL 提供统一的数据采集入口，作为Scheduler和Metrics模块调用的协调中心，确保数据采集的一致性和可靠性。
**Description**: Collector模块需要实现CollectDeviceData方法，作为Scheduler和Metrics模块调用的统一入口点。

#### Scenario: 调度器定时触发数据采集
**Given** Collector服务已初始化并连接到WinPower和Energy模块
**When** 调度器每5秒调用CollectDeviceData方法
**Then** Collector应该从WinPower获取最新设备数据并为每个设备触发电能计算
**And** 返回包含所有设备信息的CollectionResult结构体

#### Scenario: Metrics模块按需触发数据采集
**Given** Prometheus通过/metrics端点请求最新指标
**When** Metrics模块调用CollectDeviceData方法
**Then** Collector应该立即执行数据采集确保返回最新状态
**And** 采集完成后Metrics模块能够获取到最新的设备数据

### Requirement: WinPower数据集成
Collector SHALL 成功从WinPower模块获取设备数据，并将原始数据转换为标准化的内部数据结构。
**Description**: Collector需要调用WinPower客户端的CollectDeviceData方法，获取完整的设备实时数据。

#### Scenario: 成功获取WinPower设备数据
**Given** WinPower服务正常运行且有设备连接
**When** Collector调用WinPower客户端的CollectDeviceData方法
**Then** 应该返回包含所有设备信息的ParsedDeviceData结构体
**And** 数据应该包含设备的所有电气参数、状态信息和功率数据

#### Scenario: WinPower数据获取失败
**Given** WinPower服务不可用或网络连接异常
**When** Collector调用WinPower客户端的CollectDeviceData方法
**Then** 应该返回详细的错误信息
**And** 错误信息应该包含失败原因和时间戳

### Requirement: 电能计算触发
Collector SHALL 作为唯一触发电能计算的模块，遍历所有设备并为每个设备触发Energy模块的电能计算功能。
**Description**: Collector需要遍历所有获取到的设备，使用每个设备的LoadTotalWatt值触发Energy模块进行电能计算。

#### Scenario: 单个设备电能计算成功
**Given** 设备数据包含有效的LoadTotalWatt值
**When** Collector为该设备调用Energy.Calculate方法
**Then** Energy模块应该成功计算累计电能值
**And** 计算结果应该包含在DeviceCollectionInfo的EnergyValue字段中

#### Scenario: 单个设备电能计算失败
**Given** 某个设备的电能计算过程中出现错误
**When** Collector处理该设备的电能计算
**Then** 应该记录详细的错误信息但不影响其他设备的计算
**And** 该设备的EnergyCalculated字段应该设置为false
**And** ErrorMsg字段应该包含具体的错误描述

### Requirement: 数据转换和映射
Collector SHALL 将WinPower原始数据转换为标准化的CollectionResult格式，确保数据类型的一致性和完整性。
**Description**: Collector需要实现完整的数据转换逻辑，将WinPower API返回的字符串格式数据转换为对应的数据类型。

#### Scenario: 电气参数数据转换
**Given** WinPower返回的realtime数据包含字符串格式的电气参数
**When** Collector进行数据转换
**Then** 电压、电流、频率等参数应该正确转换为float64类型
**And** 转换后的数值应该保持原始精度

#### Scenario: 状态和枚举值转换
**Given** WinPower返回的状态信息为字符串格式的数字
**When** Collector进行数据转换
**Then** 状态码应该正确转换为对应的int类型
**And** 布尔值（如isCharging）应该正确转换为bool类型

#### Scenario: 功率数据转换（关键字段）
**Given** WinPower返回的LoadTotalWatt字段为字符串格式（如"500.5"）
**When** Collector进行数据转换
**Then** LoadTotalWatt应该准确转换为float64类型（如500.5）
**And** 该值将作为电能计算的核心输入参数

### Requirement: 错误处理机制
Collector SHALL 实现分层的错误处理机制，区分不同类型的错误并采用相应的处理策略，确保单点故障不影响整体系统。
**Description**: Collector需要区分不同类型的错误并采用相应的处理策略，确保单点故障不影响整体系统。

#### Scenario: WinPower采集完全失败
**Given** WinPower服务完全不可用
**When** Collector尝试采集数据
**Then** 应该立即返回错误给调用方
**And** 错误信息应该包含失败的具体原因
**And** 不应该进行任何自动重试

#### Scenario: 部分设备处理失败
**Given** 多个设备中有一个或多个设备的数据处理失败
**When** Collector处理设备数据
**Then** 成功的设备应该正常处理并包含在结果中
**And** 失败的设备应该记录错误信息但不影响其他设备
**And** CollectionResult的Success字段应该根据整体情况设置

#### Scenario: 数据转换错误
**Given** 某个设备的数据格式异常或无法转换
**When** Collector进行数据转换
**Then** 应该记录转换错误并使用合理的默认值
**And** 该设备的ErrorMsg字段应该包含转换错误信息
**And** 继续处理其他设备和字段

### Requirement: 性能和可靠性
Collector SHALL 在合理的时间内完成数据采集和计算，确保系统的实时性和响应性要求。
**Description**: Collector需要优化性能，确保单次采集的总时间在可接受范围内。

#### Scenario: 正常情况下的性能表现
**Given** 系统正常运行且网络延迟在预期范围内
**When** Collector执行完整的数据采集流程
**Then** 总耗时应该在5秒以内
**And** CollectionResult的Duration字段应该准确记录执行时间

#### Scenario: 大量设备的处理性能
**Given** 系统连接了多个设备（如10个以上）
**When** Collector并发处理所有设备的电能计算
**Then** 应该在合理时间内完成所有设备的处理
**And** 内存使用应该保持在合理范围内

### Requirement: 接口兼容性
Collector SHALL 通过接口与依赖模块解耦，支持依赖注入和Mock测试，确保模块间的松耦合和高可测试性。
**Description**: Collector需要定义清晰的接口，支持依赖注入和Mock测试。

#### Scenario: 与WinPower模块的接口集成
**Given** WinPower模块提供了标准的客户端接口
**When** Collector初始化时注入WinPower客户端
**Then** Collector应该能够通过接口调用所有必需的方法
**And** 接口调用应该处理所有的返回值和错误

#### Scenario: 与Energy模块的接口集成
**Given** Energy模块提供了标准的计算接口
**When** Collector为设备触发电能计算
**Then** 应该通过EnergyInterface接口调用Calculate方法
**And** 正确处理计算结果和可能的错误

### Requirement: 日志和监控
Collector SHALL 提供充分的日志记录用于问题诊断，在关键操作点记录结构化日志，包含足够的上下文信息。
**Description**: Collector需要在关键操作点记录结构化日志，包含足够的上下文信息。

#### Scenario: 成功采集的日志记录
**Given** 数据采集和电能计算都成功完成
**When** Collector完成一次完整的采集流程
**Then** 应该记录包含设备数量、总耗时等关键信息的成功日志
**And** 日志级别应该为INFO或DEBUG

#### Scenario: 错误情况的日志记录
**Given** 数据采集过程中发生任何错误
**When** 错误发生时
**Then** 应该记录详细的错误日志，包含错误类型、设备ID、堆栈信息等
**And** 日志级别应该为ERROR或WARN
**And** 不应该记录敏感的设备配置信息

#### Scenario: 性能相关的日志记录
**Given** 需要监控Collector的性能表现
**When** 每次采集完成后
**Then** 应该记录性能指标如耗时、处理的设备数量等
**And** 在性能异常时应该记录警告日志

