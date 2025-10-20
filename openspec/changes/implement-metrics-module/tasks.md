# implement-metrics-module 任务清单

## 实施策略

采用 TDD（测试驱动开发）方法，每个阶段都遵循"先写测试，再写实现"的原则。任务按依赖关系排序，确保每个任务都有明确的验收标准和验证方法。

## 阶段1: 核心结构与接口定义（预计2-3天）

### 1.1 创建模块目录结构
- [ ] 创建 `internal/metrics/` 目录
- [ ] 创建基础文件结构：`metrics.go`, `config.go`, `interfaces.go`
- [ ] 添加 `go.mod` 依赖：`github.com/prometheus/client_golang`
- **验收标准**: 目录结构完整，依赖安装成功

### 1.2 定义核心接口和结构
- [ ] 编写 `interfaces.go`，定义 `MetricManagerInterface`
- [ ] 定义 `MetricManager` 核心结构
- [ ] 定义四个指标分组结构（ExporterMetrics, ConnectionMetrics, DeviceMetrics, EnergyMetrics）
- [ ] 定义 `MetricManagerConfig` 配置结构
- **验收标准**: 接口定义完整，与设计文档一致

### 1.3 编写基础单元测试框架
- [ ] 创建 `metrics_test.go` 文件
- [ ] 编写测试辅助函数（创建测试配置、Mock Logger等）
- [ ] 编写 `TestNewMetricManager_Success` 测试用例
- [ ] 编写 `TestNewMetricManager_InvalidConfig` 测试用例
- **验收标准**: 测试框架搭建完成，基础测试用例通过

### 1.4 实现构造函数
- [ ] 实现 `NewMetricManager(config, logger)` 构造函数
- [ ] 添加配置验证逻辑
- [ ] 集成 prometheus.Registry 初始化
- **验收标准**: 构造函数实现，基础测试用例通过

## 阶段2: Exporter 自监控指标实现（预计2-3天）

### 2.1 编写 Exporter 自监控指标测试
- [ ] 编写 `TestExporterMetrics_Initialization` 测试
- [ ] 编写 `TestSetUp` 方法测试（成功和失败场景）
- [ ] 编写 `TestObserveRequest` 方法测试
- [ ] 编写 `TestIncScrapeError` 方法测试
- [ ] 编写 `TestObserveCollection` 方法测试
- [ ] 编写 `TestIncTokenRefresh` 方法测试
- [ ] 编写 `TestSetDeviceCount` 方法测试
- **验收标准**: 所有 Exporter 自监控指标的测试用例编写完成

### 2.2 实现 Exporter 自监控指标
- [ ] 实现 Exporter 指标初始化逻辑
- [ ] 实现 `SetUp()` 方法
- [ ] 实现 `ObserveRequest()` 方法
- [ ] 实现 `IncScrapeError()` 方法
- [ ] 实现 `ObserveCollection()` 方法
- [ ] 实现 `IncTokenRefresh()` 方法
- [ ] 实现 `SetDeviceCount()` 方法
- **验收标准**: 所有 Exporter 自监控方法实现，测试用例通过

### 2.3 验证指标格式和命名
- [ ] 验证所有指标命名符合 `winpower_exporter_*` 格式
- [ ] 验证指标类型正确（Gauge/Counter/Histogram）
- [ ] 验证标签定义符合设计要求
- **验收标准**: 指标命名和格式验证通过

## 阶段3: WinPower 连接/认证指标实现（预计2天）

### 3.1 编写连接/认证指标测试
- [ ] 编写 `TestConnectionMetrics_Initialization` 测试
- [ ] 编写 `TestSetConnectionStatus` 方法测试
- [ ] 编写 `TestSetAuthStatus` 方法测试
- [ ] 编写 `TestObserveAPI` 方法测试
- [ ] 编写 `TestSetTokenExpiry` 方法测试
- [ ] 编写 `TestSetTokenValid` 方法测试
- **验收标准**: 所有连接/认证指标的测试用例编写完成

### 3.2 实现连接/认证指标
- [ ] 实现连接/认证指标初始化逻辑
- [ ] 实现 `SetConnectionStatus()` 方法
- [ ] 实现 `SetAuthStatus()` 方法
- [ ] 实现 `ObserveAPI()` 方法
- [ ] 实现 `SetTokenExpiry()` 方法
- [ ] 实现 `SetTokenValid()` 方法
- **验收标准**: 所有连接/认证方法实现，测试用例通过

### 3.3 验证连接监控指标
- [ ] 验证指标命名符合 `winpower_*` 格式（不含 exporter）
- [ ] 验证连接状态值使用正确的枚举（1/0）
- [ ] 验证 API 响应时间桶设置合理
- **验收标准**: 连接监控指标验证通过

## 阶段4: 设备/电源指标实现（预计3天）

### 4.1 编写设备/电源指标测试
- [ ] 编写 `TestDeviceMetrics_Initialization` 测试
- [ ] 编写 `TestSetDeviceConnected` 方法测试
- [ ] 编写 `TestSetLoadPercent` 方法测试（包含相线标签）
- [ ] 编写 `TestSetElectricalData` 方法测试（完整参数）
- [ ] 编写标签验证测试（device_id, device_name, device_type, phase）
- **验收标准**: 所有设备/电源指标的测试用例编写完成

### 4.2 实现设备/电源指标
- [ ] 实现设备/电源指标初始化逻辑
- [ ] 实现 `SetDeviceConnected()` 方法
- [ ] 实现 `SetLoadPercent()` 方法
- [ ] 实现 `SetElectricalData()` 方法
- [ ] 实现标签管理和验证逻辑
- **验收标准**: 所有设备/电源方法实现，测试用例通过

### 4.3 验证设备指标标签策略
- [ ] 验证必需标签存在性检查
- [ ] 验证可选标签（phase）处理逻辑
- [ ] 验证标签值格式化和清理
- [ ] 测试高基数控制机制
- **验收标准**: 设备指标标签策略验证通过

## 阶段5: 能耗指标实现（预计2天）

### 5.1 编写能耗指标测试
- [ ] 编写 `TestEnergyMetrics_Initialization` 测试
- [ ] 编写 `TestSetPowerWatts` 方法测试
- [ ] 编写 `TestSetEnergyTotalWh` 方法测试（包含负值支持）
- [ ] 编写能耗精度测试（0.01 Wh 精度）
- **验收标准**: 所有能耗指标的测试用例编写完成

### 5.2 实现能耗指标
- [ ] 实现能耗指标初始化逻辑
- [ ] 实现 `SetPowerWatts()` 方法
- [ ] 实现 `SetEnergyTotalWh()` 方法
- [ ] 实现数值精度控制
- **验收标准**: 所有能耗方法实现，测试用例通过

### 5.3 验证能耗数据完整性
- [ ] 验证累计电能支持负值
- [ ] 验证精度控制正确实现
- [ ] 验证与 Energy 模块的数据契约
- **验收标准**: 能耗数据完整性验证通过

## 阶段6: HTTP 指标暴露实现（预计2-3天）

### 6.1 编写 HTTP 暴露测试
- [ ] 编写 `TestHandler_BasicFunctionality` 测试
- [ ] 编写 `TestHandler_ErrorHandling` 测试
- [ ] 编写 `TestHandler_OpenMetricsFormat` 测试
- **验收标准**: HTTP 暴露功能的测试用例编写完成

### 6.2 实现 HTTP Handler
- [ ] 实现 `Handler()` 方法
- [ ] 集成 promhttp.HandlerFor
- [ ] 配置 OpenMetrics 支持
- **验收标准**: HTTP Handler 实现，测试用例通过

### 6.3 验证指标输出格式
- [ ] 验证 Prometheus 格式正确性
- [ ] 验证所有指标都包含在输出中
- [ ] 验证指标值和标签正确显示
- **验收标准**: 指标输出格式验证通过

## 阶段7: 集成测试（预计2天）

### 7.1 编写集成测试
- [ ] 创建 `integration_test.go` 文件
- [ ] 编写端到端指标流转测试
- [ ] 编写与 WinPower 模块集成测试
- [ ] 编写与 Energy 模块集成测试
- **验收标准**: 集成测试用例编写完成并通过

### 7.2 集成验证
- [ ] 验证指标在不同场景下的正确性
- [ ] 测试模块间的数据流转
- [ ] 验证错误场景的处理
- **验收标准**: 集成测试通过，模块协作正常

## 阶段8: 错误处理和日志记录（预计1-2天）

### 8.1 实现错误处理机制
- [ ] 实现指标更新错误处理
- [ ] 实现配置验证错误处理
- [ ] 实现并发访问异常处理
- [ ] 编写错误处理测试用例
- **验收标准**: 错误处理机制完善，测试通过

### 8.2 集成结构化日志
- [ ] 在关键路径添加结构化日志
- [ ] 实现日志级别控制
- [ ] 验证日志格式和内容
- **验收标准**: 日志记录完整，格式正确

### 8.3 异常恢复测试
- [ ] 测试各种异常场景的恢复能力
- [ ] 验证系统在异常情况下的稳定性
- **验收标准**: 异常恢复测试通过

## 阶段9: 文档和配置示例（预计1天）

### 9.1 编写模块文档
- [ ] 编写 `README.md` 模块说明文档
- [ ] 编写 API 文档和使用示例
- [ ] 编写配置参数说明
- **验收标准**: 文档完整，示例清晰

### 9.2 提供配置示例
- [ ] 提供 YAML 配置示例
- [ ] 提供环境变量配置示例
- [ ] 编写配置验证工具
- **验收标准**: 配置示例可用，验证工具正常

## 阶段10: 代码质量检查（预计1天）

### 10.1 静态代码分析
- [ ] 运行 `make fmt` 格式化代码
- [ ] 运行 `make lint` 静态分析
- [ ] 修复所有静态分析问题
- **验收标准**: 代码格式规范，无静态分析警告

### 10.2 测试覆盖率检查
- [ ] 运行 `make test-coverage` 检查覆盖率
- [ ] 确保测试覆盖率达到 90% 以上
- [ ] 补充遗漏的测试用例
- **验收标准**: 测试覆盖率达标

### 10.3 最终集成验证
- [ ] 运行完整的测试套件
- [ ] 验证与现有模块的集成
- [ ] 确认所有功能正常工作
- **验收标准**: 所有测试通过，集成验证成功

## 依赖关系和并行工作

### 可并行执行的任务
- **阶段1-5** 可以部分并行：不同指标组的测试和实现可以并行进行
- **文档编写** 可以在实现过程中并行进行
- **配置管理** 可以独立实现和测试

### 关键依赖路径
1. **阶段1** → **所有其他阶段**：核心结构必须先完成
2. **阶段2-5** → **阶段6**：指标必须先注册才能暴露
3. **所有实现阶段** → **阶段7**：集成测试需要所有功能完成
4. **阶段7** → **阶段8-10**：验证完成后才能进行最终检查

### 质量门禁
每个阶段都必须满足以下条件才能进入下一阶段：
- [ ] 所有测试用例通过
- [ ] 代码覆盖率达标
- [ ] 静态分析无警告
- [ ] 设计文档要求满足

## 风险缓解措施

### 技术风险
- **Prometheus 版本兼容性**：提前验证依赖版本兼容性
- **并发安全**：充分测试并发场景，使用竞态检测工具

### 进度风险
- **测试复杂性**：为复杂测试预留额外时间
- **集成问题**：安排专门的集成调试时间
- **代码审查**：预留代码审查和修改时间

这个任务清单按照 TDD 原则组织，确保每个功能都有对应的测试用例，并且任务之间的依赖关系清晰，支持增量开发和持续集成。