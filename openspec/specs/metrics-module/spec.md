# metrics-module Specification

## Purpose
定义Metrics模块的完整功能规格，包括指标管理体系结构、双类指标管理、数据映射转换、错误处理和性能优化等核心要求。
## Requirements
### Requirement: 指标管理体系结构

Metrics模块SHALL作为系统的指标管理中心，提供统一的指标创建、更新和暴露能力。

#### Scenario: Metrics模块作为指标管理中心
- **给定**: 系统需要统一管理Prometheus指标
- **当**: Metrics模块初始化时
- **那么**: 应创建MetricsService实例，包含Prometheus注册表和指标定义
- **并且**: 提供Gin Handler作为/metrics端点入口

#### Scenario: 与Collector模块协调数据采集
- **给定**: Prometheus请求/metrics端点
- **当**: HandleMetrics方法被调用时
- **那么**: 应调用Collector.CollectDeviceData获取最新数据
- **并且**: 基于采集结果更新所有相关指标
- **并且**: 返回Prometheus格式的指标数据

### Requirement: 双类指标管理

Metrics模块SHALL管理两类不同的指标：Exporter自监控指标用于监控exporter自身的运行状态，WinPower设备指标用于暴露UPS等设备的监控数据。

#### Scenario: Exporter自监控指标
- **给定**: 需要监控exporter自身运行状态
- **当**: Metrics模块初始化时
- **那么**: 应创建Exporter自监控指标，包括运行状态、请求计数、内存使用等
- **并且**: 使用`winpower_host`标签进行标识

#### Scenario: WinPower设备指标
- **给定**: 需要暴露WinPower设备监控数据
- **当**: 接收到Collector采集结果时
- **那么**: 应创建或更新设备指标，包括电气参数、负载功率、电池状态等
- **并且**: 使用`winpower_host`、`device_id`、`device_name`、`device_type`标签

### Requirement: 数据映射和转换

Metrics模块SHALL能够将Collector模块返回的CollectionResult结构体数据正确映射到对应的Prometheus指标，确保数据类型转换的正确性和完整性。

#### Scenario: Collector结果到指标映射
- **给定**: Collector返回CollectionResult结构体
- **当**: updateMetrics方法处理采集结果时
- **那么**: 应将CollectionResult中的所有字段映射到对应的Prometheus指标
- **并且**: 确保数据类型转换的正确性（如string到float64）

#### Scenario: 核心功率指标处理
- **给定**: 设备的LoadTotalWatt数据
- **当**: 更新设备指标时
- **那么**: 应将LoadTotalWatt同时更新到`winpower_device_load_total_watts`和`winpower_power_watts`指标
- **并且**: 确保瞬时功率指标与总负载功率保持一致

### Requirement: 错误处理和日志记录

Metrics模块SHALL具备完善的错误处理机制，能够区分不同类型的错误并进行相应的处理，同时提供详细的日志记录以便问题排查和系统监控。

#### Scenario: Collector调用失败
- **给定**: Collector.CollectDeviceData返回错误
- **当**: HandleMetrics方法处理请求时
- **那么**: 应记录详细错误日志
- **并且**: 返回HTTP 500状态码和错误信息
- **并且**: 更新错误统计指标

#### Scenario: 指标更新失败
- **给定**: 某些指标更新过程中出现错误
- **当**: updateMetrics方法执行时
- **那么**: 应记录错误日志但不影响HTTP响应
- **并且**: 继续更新其他指标

### Requirement: 内存管理和性能优化

Metrics模块SHALL具备良好的内存管理机制和性能优化策略，包括动态设备指标管理、并发访问控制和资源清理机制，确保系统在高负载下的稳定运行。

#### Scenario: 动态设备指标管理
- **给定**: 新设备被发现或旧设备下线
- **当**: 处理Collector采集结果时
- **那么**: 应为新设备创建指标实例
- **并且**: 定期清理长时间未更新的设备指标
- **并且**: 使用读写锁保护并发访问

#### Scenario: 并发请求处理
- **给定**: 多个Prometheus抓取请求同时到达
- **当**: HandleMetrics方法被并发调用时
- **那么**: 应使用读写锁保护指标更新操作
- **并且**: 确保数据一致性和线程安全

### Requirement: 指标格式和标准兼容性

Metrics模块SHALL严格遵循Prometheus指标规范，提供标准的指标格式和命名规范，确保与Prometheus生态系统的完全兼容性和良好的监控体验。

#### Scenario: Prometheus格式输出
- **给定**: Prometheus抓取/metrics端点
- **当**: 生成响应时
- **那么**: 应输出标准的Prometheus文本格式
- **并且**: 包含所有指标的定义和当前值
- **并且**: 遵循Prometheus指标命名规范

#### Scenario: Histogram桶配置
- **给定**: 需要测量请求和采集耗时分布
- **当**: 创建Histogram指标时
- **那么**: 应使用预定义的桶配置：请求/采集耗时`[0.05, 0.1, 0.2, 0.5, 1, 2, 5]`秒
- **并且**: API响应时间`[0.05, 0.1, 0.2, 0.5, 1]`秒

