## ADDED Requirements

### Requirement: WinPower客户端接口
系统 SHALL 提供WinPowerClient接口，用于与WinPower G2系统进行统一的数据采集和认证管理。

#### Scenario: 成功采集设备数据
- **WHEN** 调用CollectDeviceData方法且认证有效
- **THEN** 系统应从WinPower API获取设备数据并触发电能计算
- **AND** 返回采集成功状态

#### Scenario: 连接状态查询
- **WHEN** 调用GetConnectionStatus方法
- **THEN** 系统应返回与WinPower系统的当前连接状态

#### Scenario: 最后采集时间查询
- **WHEN** 调用GetLastCollectionTime方法
- **THEN** 系统应返回最后一次成功采集的时间戳

### Requirement: HTTP客户端管理
系统 SHALL 提供HTTPClient，用于管理与WinPower系统的HTTP通信，支持SSL/TLS配置和连接复用。

#### Scenario: HTTPS请求配置
- **WHEN** 配置了WinPower服务器地址和SSL选项
- **THEN** HTTPClient应支持HTTPS请求并可配置跳过SSL验证

#### Scenario: 连接复用和超时
- **WHEN** 发起多个API请求
- **THEN** HTTPClient应复用TCP连接并遵守配置的超时时间

#### Scenario: 请求头和User-Agent设置
- **WHEN** 发起HTTP请求
- **THEN** 系统应设置适当的请求头和User-Agent标识

### Requirement: Token认证管理
系统 SHALL 提供TokenManager，用于管理WinPower的JWT Token生命周期，包括登录、缓存和自动刷新。

#### Scenario: 初始登录获取Token
- **WHEN** 使用有效的用户名和密码调用Login方法
- **THEN** 系统应向WinPower认证API发起请求并缓存返回的Token

#### Scenario: Token有效性检查
- **WHEN** 检查缓存的Token有效性
- **THEN** 系统应验证Token是否过期并在需要时自动刷新

#### Scenario: 自动Token刷新
- **WHEN** Token即将过期且需要访问API
- **THEN** 系统应自动使用保存的凭据刷新Token并更新缓存

#### Scenario: 并发安全的Token访问
- **WHEN** 多个goroutine同时访问Token
- **THEN** 系统应使用读写锁确保并发安全

### Requirement: 设备数据解析
系统 SHALL 提供数据解析功能，用于解析WinPower API返回的设备数据并标准化输出。

#### Scenario: 设备基础信息解析
- **WHEN** 接收到设备数据响应
- **THEN** 系统应解析设备ID、名称、类型、连接状态等基础信息

#### Scenario: 功率和电能数据解析
- **WHEN** 接收到实时数据
- **THEN** 系统应解析有功功率、无功功率、功率因数等电能相关指标

#### Scenario: 多相数据支持
- **WHEN** 设备提供多相（A/B/C）数据
- **THEN** 系统应分别解析各相的电压、电流、频率等参数

#### Scenario: 数据验证和格式化
- **WHEN** 解析设备数据
- **THEN** 系统应验证数据完整性并将数值转换为标准单位

### Requirement: 与Energy模块集成
WinPower模块 SHALL 在成功采集设备数据后，自动触发电能计算和持久化。

#### Scenario: 自动触发电能计算
- **WHEN** 成功采集到设备的瞬时功率数据
- **THEN** 系统应调用Energy模块的Calculate方法进行电能累计

#### Scenario: 能耗计算错误处理
- **WHEN** Energy模块返回计算错误
- **THEN** 系统应记录错误日志但不影响其他设备的能耗计算

### Requirement: 配置管理
系统 SHALL 提供WinPower模块的配置结构，支持通过配置文件和环境变量进行配置。

#### Scenario: 必需配置验证
- **WHEN** 加载WinPower配置
- **THEN** 系统应验证BaseURL、Username、Password等必需字段

#### Scenario: 可选配置默认值
- **WHEN** 未提供Timeout、MaxRetries等可选配置
- **THEN** 系统应使用合理的默认值

#### Scenario: SSL证书配置
- **WHEN** 配置了SkipSSLVerify选项
- **THEN** 系统应在HTTPS请求中相应地跳过或执行证书验证

### Requirement: 错误处理和日志记录
系统 SHALL 提供完善的错误处理机制和结构化日志记录。

#### Scenario: 网络错误处理
- **WHEN** 发生网络连接错误
- **THEN** 系统应返回明确的错误类型并记录详细日志

#### Scenario: 认证错误处理
- **WHEN** 登录失败或Token无效
- **THEN** 系统应尝试刷新Token或返回认证错误

#### Scenario: 敏感信息保护
- **WHEN** 记录日志
- **THEN** 系统不得记录密码、Token等敏感信息

#### Scenario: 结构化日志
- **WHEN** 记录操作日志
- **THEN** 系统应使用结构化格式包含设备ID、耗时、状态等字段

### Requirement: 监控指标暴露
系统 SHALL 暴露与WinPower连接和采集相关的监控指标。

#### Scenario: 连接状态指标
- **WHEN** 查询连接状态
- **THEN** 系统应暴露winpower_connection_status指标

#### Scenario: 采集性能指标
- **WHEN** 执行数据采集
- **THEN** 系统应暴露采集耗时、成功率、错误计数等指标

#### Scenario: Token管理指标
- **WHEN** 执行Token操作
- **THEN** 系统应暴露Token刷新次数、剩余有效期等指标