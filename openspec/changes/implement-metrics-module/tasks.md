# Metrics模块实现任务清单

## 阶段1: 基础结构实现

### 1.1 创建模块目录结构
- [x] 创建`internal/metrics`目录
- [x] 创建`internal/metrics/service.go`文件
- [x] 创建`internal/metrics/metrics.go`文件（指标定义）
- [x] 创建`internal/metrics/types.go`文件（数据结构）

### 1.2 实现核心数据结构
- [x] 实现MetricsService结构体（基于specs/implementation/spec.md）
- [x] 实现DeviceMetrics结构体
- [x] 定义指标命名和标签常量
- [x] 定义Histogram桶配置常量

## 阶段2: 指标初始化和注册

### 2.1 Exporter自监控指标
- [x] 实现`winpower_exporter_up`指标
- [x] 实现`winpower_exporter_requests_total`指标
- [x] 实现`winpower_exporter_request_duration_seconds`指标
- [x] 实现`winpower_exporter_collection_duration_seconds`指标
- [x] 实现`winpower_exporter_scrape_errors_total`指标
- [x] 实现`winpower_exporter_memory_bytes`指标
- [x] 实现`winpower_exporter_device_count`指标

### 2.2 WinPower连接指标
- [x] 实现`winpower_connection_status`指标
- [x] 实现`winpower_auth_status`指标
- [x] 实现`winpower_api_response_time_seconds`指标
- [x] 实现`winpower_token_expiry_seconds`指标
- [x] 实现`winpower_token_valid`指标

### 2.3 设备指标模板
- [x] 实现`winpower_device_connected`指标
- [x] 实现`winpower_device_load_percent`指标
- [x] 实现`winpower_device_load_total_watts`指标
- [x] 实现`winpower_power_watts`指标
- [x] 实现所有电气参数指标（电压、电流、频率等）
- [x] 实现电池状态指标
- [x] 实现UPS状态指标
- [x] 实现故障代码指标

## 阶段3: HTTP Handler实现

### 3.1 核心Handler逻辑
- [x] 实现`NewMetricsService`构造函数
- [x] 实现`HandleMetrics`HTTP handler方法
- [x] 实现请求计数和耗时记录
- [x] 实现错误响应处理

### 3.2 数据协调逻辑
- [x] 实现调用Collector.CollectDeviceData的逻辑
- [x] 实现超时处理和上下文取消
- [x] 实现数据采集成功/失败的处理分支

## 阶段4: 指标更新逻辑

### 4.1 自监控指标更新
- [x] 实现`updateSelfMetrics`方法
- [x] 实现内存使用监控`updateMemoryMetrics`
- [x] 实现请求耗时统计
- [x] 实现错误计数更新

### 4.2 设备指标更新
- [x] 实现`updateMetrics`主更新方法
- [x] 实现`updateDeviceMetrics`设备指标更新方法
- [x] 实现`createDeviceMetrics`设备指标创建方法
- [x] 实现CollectionResult到指标的映射逻辑
- [x] 实现LoadTotalWatt到instantPower的特殊映射

### 4.3 并发安全和内存管理
- [x] 实现读写锁保护机制
- [x] 实现设备指标的动态创建
- [x] 实现不活跃设备指标的清理机制
- [x] 实现线程安全的指标访问

## 阶段5: 错误处理和日志

### 5.1 错误分类和处理
- [x] 实现`handleCollectorError`错误分类处理
- [x] 实现不同错误类型的指标更新
- [x] 实现详细的错误日志记录
- [x] 实现设备级错误处理

### 5.2 日志记录
- [x] 在关键路径添加结构化日志
- [x] 实现性能相关日志记录
- [x] 实现错误详情日志记录
- [x] 实现调试信息日志（可选）

## 阶段6: 单元测试

### 6.1 基础测试
- [x] 创建`service_test.go`测试文件
- [x] 实现NewMetricsService的测试
- [x] 实现指标初始化的测试
- [x] 实现基础数据结构的测试

### 6.2 Mock依赖测试
- [x] 创建MockCollectorInterface
- [x] 实现HandleMetrics的成功场景测试
- [x] 实现HandleMetrics的失败场景测试
- [x] 实现超时场景测试

### 6.3 指标更新测试
- [x] 实现updateMetrics的测试
- [x] 实现updateDeviceMetrics的测试
- [x] 实现createDeviceMetrics的测试
- [x] 使用prometheus/testutil验证指标值

### 6.4 并发安全测试
- [x] 实现并发访问测试
- [x] 实现读写锁正确性测试
- [x] 实现竞态条件检测

### 6.5 错误处理测试
- [x] 实现各种错误场景的测试
- [x] 实现错误分类正确性测试
- [x] 实现错误日志记录验证

## 阶段7: 集成测试和验证

### 7.1 集成测试
- [x] 创建集成测试文件`integration_test.go`
- [x] 实现与真实Collector模块的集成测试
- [x] 实现完整数据流测试
- [x] 验证指标格式正确性

### 7.2 标准合规性验证
- [x] 验证Prometheus格式合规性
- [x] 验证指标命名规范
- [x] 验证标签使用规范
- [x] 验证与设计文档的一致性

## 阶段8: 代码质量保证

### 8.1 代码规范
- [x] 运行`make fmt`格式化代码
- [x] 运行`make lint`静态检查并修复问题
- [x] 确保所有公共方法有文档注释
- [x] 检查错误处理的完整性

### 8.2 测试覆盖率
- [x] 运行`make test-coverage`检查测试覆盖率
- [x] 确保覆盖率达到80%以上
- [x] 补充缺失的测试用例
- [x] 验证边界条件测试

### 8.3 文档更新
- [x] 更新模块README文档（如果需要）
- [x] 验证与设计文档的一致性
- [x] 确保代码示例的正确性
- [x] 更新相关的使用说明

## 验收标准

- [x] 所有单元测试通过：`make test`
- [x] 静态检查通过：`make lint`
- [x] 测试覆盖率达到80%以上：`make test-coverage` (实际达到96.8%)
- [x] 指标格式符合Prometheus规范
- [x] 与Collector模块集成正常工作
- [x] 并发安全性验证通过
- [x] 内存使用合理，无泄漏
- [x] 错误处理完整，日志记录详细