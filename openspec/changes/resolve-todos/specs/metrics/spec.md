## ADDED Requirements

### Requirement: 指标模块完善
系统 SHALL 完善指标模块实现，包括布尔指标转换、设备指标完整性和标签标准化。

#### Scenario: 布尔指标转换
- **WHEN** 需要将设备状态转换为Prometheus指标
- **THEN** 系统接受各种布尔输入格式（bool, string, int）
- **AND** 标准转换为0/1数值
- **AND** 设置适当的指标标签和描述
- **AND** 确保转换的一致性和可预测性

#### Scenario: 设备指标完整性
- **WHEN** 系统采集设备数据时
- **THEN** 提供设备状态指标（连接状态、健康状态、运行模式）
- **AND** 提供设备性能指标（电压、电流、功率、效率）
- **AND** 提供电能计算指标（瞬时功率、累计电能）
- **AND** 每个指标都有清晰的定义和单位

#### Scenario: 指标标签标准化
- **WHEN** 注册设备指标时
- **THEN** 包含标准标签：device_id, device_name, device_type, winpower_host
- **AND** 状态标签值标准化（online, offline, maintenance）
- **AND** 健康状态标签标准化（normal, warning, critical）
- **AND** 确保标签值的一致性

#### Scenario: 指标质量保证
- **WHEN** 系统处理指标数据时
- **THEN** 验证指标值的有效性和合理性
- **AND** 检测和处理异常的指标值
- **AND** 提供指标质量监控和告警
- **AND** 保持指标历史连续性