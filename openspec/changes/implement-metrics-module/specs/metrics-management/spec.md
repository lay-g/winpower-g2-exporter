# Metrics Management 能力规范

## ADDED Requirements

### Requirement: Metrics Manager 初始化
系统必须提供 Metrics Manager 组件，负责 Prometheus 指标的注册、管理和暴露。系统 SHALL 确保 Metrics Manager 能够正确初始化并管理所有指标类型。

#### Scenario: 成功初始化 Metrics Manager
**Given** 有效的 Metrics Manager 配置
**When** 调用 `NewMetricManager(config, logger)` 构造函数
**Then** 系统应该返回一个 MetricManager 实例
**And** 该实例应该初始化所有四类指标（Exporter自监控、WinPower连接/认证、设备/电源、能耗）
**And** 所有指标应该使用指定的命名空间和子系统前缀

#### Scenario: 初始化失败处理
**Given** 无效的 Metrics Manager 配置
**When** 调用 `NewMetricManager(config, logger)` 构造函数
**Then** 系统应该返回错误
**And** 错误信息应该包含具体的配置问题

### Requirement: 指标注册和管理
Metrics Manager 必须能够注册和管理所有 Prometheus 指标，确保指标的完整性和一致性。系统 SHALL 正确注册所有四类指标并维护其状态一致性。

#### Scenario: 注册 Exporter 自监控指标
**Given** Metrics Manager 实例
**When** 初始化完成
**Then** 系统应该注册 7 个 Exporter 自监控指标
**And** 指标命名应该遵循 `winpower_exporter_*` 格式
**And** 每个指标应该具有正确的类型（Gauge/Counter/Histogram）

#### Scenario: 注册 WinPower 连接/认证指标
**Given** Metrics Manager 实例
**When** 初始化完成
**Then** 系统应该注册 5 个 WinPower 连接/认证指标
**And** 指标命名应该遵循 `winpower_*` 格式（不含 exporter 前缀）
**And** 所有连接/认证相关标签应该正确定义

#### Scenario: 注册设备/电源指标
**Given** Metrics Manager 实例
**When** 初始化完成
**Then** 系统应该注册 11 个设备/电源指标
**And** 每个指标应该包含 `device_id`, `device_name`, `device_type` 标签
**And** 多相设备指标应该支持可选的 `phase` 标签

#### Scenario: 注册能耗指标
**Given** Metrics Manager 实例
**When** 初始化完成
**Then** 系统应该注册 2 个能耗指标（瞬时功率和累计电能）
**And** 累计电能指标应该支持负值（净能量）
**And** 指标单位应该正确标注（watts, wh）

### Requirement: 指标更新接口
系统必须提供类型安全的指标更新接口，支持批量更新。系统 SHALL 提供线程安全的指标更新方法并确保数据一致性。

#### Scenario: 更新设备连接状态
**Given** Metrics Manager 实例和设备数据
**When** 调用 `SetDeviceConnected(deviceID, deviceName, deviceType, connected)`
**Then** 系统应该更新 `winpower_device_connected` 指标
**And** 指标应该包含正确的设备标签和连接状态值
**And** 更新操作应该是线程安全的

#### Scenario: 更新电气参数
**Given** Metrics Manager 实例和电气测量数据
**When** 调用 `SetElectricalData()` 方法
**Then** 系统应该更新所有相关的电气指标
**And** 电压、电流、频率、功率因数等指标应该正确设置
**And** 所有指标应该使用相同的设备标签组合

#### Scenario: 更新能耗数据
**Given** Metrics Manager 实例和能耗计算结果
**When** 调用 `SetPowerWatts()` 和 `SetEnergyTotalWh()` 方法
**Then** 系统应该更新瞬时功率和累计电能指标
**And** 累计电能应该支持负值输入
**And** 指标应该保持 0.01 精度

#### Scenario: 批量更新设备指标
**Given** Metrics Manager 实例和多个设备数据
**When** 调用批量更新接口
**Then** 系统应该一次性更新所有相关指标
**And** 更新过程中应该保持数据一致性
**And** 批量操作应该比单独更新更高效

### Requirement: HTTP 指标暴露
系统必须通过 HTTP `/metrics` 端点暴露 Prometheus 格式的指标数据。系统 SHALL 提供标准的 HTTP 端点来暴露所有注册的指标并支持 Prometheus 格式。

#### Scenario: 处理指标请求
**Given** Metrics Manager 实例和 HTTP 请求
**When** 调用 `Handler()` 方法处理 `/metrics` 请求
**Then** 系统应该返回 Prometheus 格式的指标数据
**And** 响应该包含所有已注册的指标及其当前值
**And** 响应该支持 OpenMetrics 格式

#### Scenario: 错误处理和状态码
**Given** Metrics Manager 实例和异常情况
**When** 发生内部错误或指标不可用
**Then** 系统应该返回适当的 HTTP 状态码
**And** 错误信息应该记录在日志中
**And** 系统应该继续正常处理其他请求

### Requirement: Exporter 自监控
系统必须提供完整的 Exporter 自监控能力，包括运行状态、请求统计和运行时指标。系统 SHALL 监控 Exporter 的运行状态并记录关键运行时指标。

#### Scenario: 运行状态监控
**Given** Metrics Manager 实例
**When** 调用 `SetUp(true)` 方法
**Then** `winpower_exporter_up` 指标应该设置为 1
**And** 指标应该包含 `winpower_host` 和 `version` 标签
**When** 调用 `SetUp(false)` 方法
**Then** `winpower_exporter_up` 指标应该设置为 0

#### Scenario: HTTP 请求统计
**Given** Metrics Manager 实例和 HTTP 请求信息
**When** 调用 `ObserveRequest(host, method, code, duration)`
**Then** 系统应该增加 `winpower_exporter_requests_total` 计数器
**And** 系统应该记录请求时延到 `winpower_exporter_request_duration_seconds`
**And** 所有相关标签应该正确设置

#### Scenario: 采集性能监控
**Given** Metrics Manager 实例和采集结果
**When** 调用 `ObserveCollection(status, duration)`
**Then** 系统应该记录采集耗时到 `winpower_exporter_collection_duration_seconds`
**And** 指标应该包含 `winpower_host` 和 `status` 标签
**And** 直方图桶应该适合采集耗时分布

### Requirement: 连接和认证状态监控
系统必须监控 WinPower 连接状态和认证状态，提供详细的连接和认证可观测性。系统 SHALL 监控 WinPower 的连接状态并提供认证相关的可观测性指标。

#### Scenario: 连接状态监控
**Given** Metrics Manager 实例和连接状态信息
**When** 调用 `SetConnectionStatus(host, connectionType, status)`
**Then** `winpower_connection_status` 指标应该更新
**And** 指标应该包含 `winpower_host` 和 `connection_type` 标签
**And** 状态值应该使用 1（连接）或 0（断开）

#### Scenario: API 响应时间监控
**Given** Metrics Manager 实例和 API 调用结果
**When** 调用 `ObserveAPI(host, endpoint, duration)`
**Then** 系统应该记录 API 响应时延
**And** `winpower_api_response_time_seconds` 应该包含正确的标签
**And** 直方图桶应该适合 API 响应时间分布

#### Scenario: Token 状态监控
**Given** Metrics Manager 实例和 Token 信息
**When** 调用 `SetTokenValid()` 和 `SetTokenExpiry()` 方法
**Then** 系统应该更新 Token 有效性和剩余有效期指标
**And** 指标应该包含 `winpower_host` 和 `user_id` 标签
**And** 有效期应该以秒为单位

### Requirement: 配置管理
系统必须支持灵活的配置管理，允许自定义指标命名空间和子系统。系统 SHALL 支持配置化的指标参数并提供合理的默认值。

#### Scenario: 自定义命名空间和子系统
**Given** 自定义的命名空间和子系统配置
**When** 初始化 Metrics Manager
**Then** 所有指标名称应该使用指定的命名空间和子系统前缀
**And** 指标名称格式应该为 `namespace_subsystem_metric`

#### Scenario: 默认配置处理
**Given** 部分配置缺失
**When** 初始化 Metrics Manager
**Then** 系统应该使用合理的默认值
**And** 默认值应该确保良好的可观测性

### Requirement: 错误处理和日志记录
系统必须提供完善的错误处理和结构化日志记录，确保系统的可观测性和可维护性。系统 SHALL 提供结构化的错误处理并记录关键操作日志。

#### Scenario: 指标更新错误处理
**Given** 指标更新过程中的错误
**When** 发生更新失败
**Then** 系统应该记录错误日志
**And** 错误不应该影响其他指标的更新
**And** 系统应该继续正常运行

#### Scenario: 数据安全
**Given** Metrics Manager 运行时状态
**When** 执行指标更新操作
**Then** 系统应该保证数据一致性
**And** 不应该发生数据竞争或死锁
**And** 所有更新操作都应该正确应用

#### Scenario: 结构化日志记录
**Given** 指标操作的各个阶段
**When** 执行指标更新、HTTP 请求等操作
**Then** 系统应该记录结构化日志
**And** 日志应该包含关键上下文信息
**And** 日志级别应该根据操作重要性适当设置


## MODIFIED Requirements

### Requirement: 指标命名规范一致性
所有指标命名必须与现有设计文档保持一致，确保与其他模块的集成兼容性。系统 SHALL 确保所有指标命名严格遵循设计文档中定义的规范。

#### Scenario: 指标命名验证
**Given** 所有已注册的指标
**When** 检查指标名称
**Then** 指标名称应该严格遵循 `namespace_subsystem_metric` 格式
**And** Exporter 自监控指标应该使用 `winpower_exporter_*` 前缀
**And** 业务指标应该使用 `winpower_*` 前缀（不含 exporter）

#### Scenario: 标签命名验证
**Given** 所有指标的标签定义
**When** 检查标签名称
**Then** 标签名称应该使用下划线分隔的小写格式
**And** 标签值应该符合枚举约束
**And** 设备标签应该使用标准组合（device_id, device_name, device_type）

## REMOVED Requirements

### 无

此规范文档定义了 Metrics Management 能力的完整需求，涵盖了初始化、指标管理、HTTP暴露、自监控、连接监控、配置管理、错误处理等各个方面。所有需求都按照 TDD 原则设计，支持编写对应的测试用例。