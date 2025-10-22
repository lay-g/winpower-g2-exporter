# 模块配置集成规格

## MODIFIED Requirements

### Requirement: Storage模块配置集成
The system SHALL modify the Storage module configuration to implement the unified Config interface.

#### Scenario: Storage配置验证
- **GIVEN** 一个Storage配置结构体
- **WHEN** 调用Validate方法
- **THEN** 应该验证data_dir不为空
- **AND** 应该验证文件权限和目录权限的有效性
- **AND** 应该验证同步写入和创建目录选项

#### Scenario: Storage配置环境变量支持
- **GIVEN** 设置了WINPOWER_EXPORTER_STORAGE_*环境变量
- **WHEN** 创建Storage配置时
- **THEN** 环境变量应该正确映射到配置字段
- **AND** 应该覆盖YAML文件中的默认值

#### Scenario: Storage配置默认值设置
- **GIVEN** 一个空的Storage配置
- **WHEN** 调用SetDefaults方法
- **THEN** data_dir应该设置为"./data"
- **AND** file_permissions应该设置为0644
- **AND** dir_permissions应该设置为0755
- **AND** sync_write应该设置为true
- **AND** create_dir应该设置为true

### Requirement: WinPower模块配置集成
The system SHALL modify the WinPower module configuration to maintain existing validation logic while implementing the unified interface.

#### Scenario: WinPower配置URL验证
- **GIVEN** 一个WinPower配置
- **WHEN** 调用Validate方法
- **THEN** 应该验证URL格式正确
- **AND** 应该验证URL使用http或https协议
- **AND** 应该验证主机名有效性

#### Scenario: WinPower配置敏感信息处理
- **GIVEN** WinPower配置包含用户名和密码
- **WHEN** 调用String方法
- **THEN** 密码应该被脱敏显示
- **AND** 用户名可以正常显示
- **AND** URL应该正常显示

#### Scenario: WinPower配置环境变量支持
- **GIVEN** 设置了WINPOWER_EXPORTER_URL、USERNAME、PASSWORD环境变量
- **WHEN** 创建WinPower配置时
- **THEN** 这些环境变量应该正确映射到配置字段
- **AND** 应该覆盖默认值和YAML值

### Requirement: Energy模块配置集成
The system SHALL modify the Energy module configuration to support numerical precision validation.

#### Scenario: Energy配置精度验证
- **GIVEN** 一个Energy配置
- **WHEN** 调用Validate方法
- **THEN** 应该验证precision为正数
- **AND** 应该验证max_calculation_time为合理范围
- **AND** 应该验证negative_power_allowed为布尔值

#### Scenario: Energy配置默认值保持
- **GIVEN** 一个Energy配置
- **WHEN** 调用DefaultConfig函数
- **THEN** precision应该设置为0.01
- **AND** enable_stats应该设置为true
- **AND** max_calculation_time应该设置为1秒
- **AND** negative_power_allowed应该设置为true

### Requirement: Server模块配置集成
The system SHALL modify the Server module configuration to support network parameter validation.

#### Scenario: Server配置端口验证
- **GIVEN** 一个Server配置
- **WHEN** 调用Validate方法
- **THEN** 应该验证port在有效范围内(1-65535)
- **AND** 应该验证host不为空
- **AND** 应该验证超时时间为正值

#### Scenario: Server配置功能开关验证
- **GIVEN** Server配置的各种功能开关
- **WHEN** 调用Validate方法
- **THEN** 应该验证enable_pprof为布尔值
- **AND** 应该验证enable_cors为布尔值
- **AND** 应该验证enable_rate_limit为布尔值

### Requirement: Scheduler模块配置集成
The system SHALL modify the Scheduler module configuration to support time interval validation.

#### Scenario: Scheduler配置间隔验证
- **GIVEN** 一个Scheduler配置
- **WHEN** 调用Validate方法
- **THEN** 应该验证collection_interval为正值
- **AND** 应该验证graceful_shutdown_timeout为合理范围
- **AND** 应该验证时间间隔不冲突

## ADDED Requirements

### Requirement: 配置加载顺序依赖
The system SHALL load module configurations in dependency order.

#### Scenario: 配置加载顺序执行
- **GIVEN** 多个模块需要配置
- **WHEN** 启动应用程序时
- **THEN** 应该按以下顺序加载配置：
  - storage (基础存储配置)
  - winpower (外部连接配置)
  - energy (计算配置)
  - server (服务配置)
  - scheduler (调度配置)
- **AND** 如果前置配置加载失败应该停止后续加载

### Requirement: 配置依赖验证
The system SHALL validate configuration dependencies between modules.

#### Scenario: 验证模块间配置依赖
- **GIVEN** 存在模块间配置依赖关系
- **WHEN** 加载完所有配置后
- **THEN** 应该验证storage.data_dir目录可访问
- **AND** 应该验证winpower配置的连接参数有效
- **AND** 应该验证server端口未被占用
- **AND** 如果依赖不满足应该返回详细错误信息


### Requirement: 配置向后兼容性
The system SHALL maintain backward compatibility with existing configuration formats.

#### Scenario: 兼容旧配置格式
- **GIVEN** 使用旧版本配置文件
- **WHEN** 使用新配置系统加载时
- **THEN** 应该成功加载配置
- **AND** 应该应用适当的默认值
- **AND** 应该记录兼容性警告日志

### Requirement: 配置迁移支持
The system SHALL provide configuration migration tools.

#### Scenario: 自动配置迁移
- **GIVEN** 检测到旧版本配置格式
- **WHEN** 启动应用程序时
- **THEN** 应该提供配置迁移建议
- **AND** 可以选择自动迁移配置文件
- **AND** 应该备份原始配置文件