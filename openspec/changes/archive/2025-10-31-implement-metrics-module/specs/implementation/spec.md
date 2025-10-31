# Metrics模块实现规格

## 概述

本规格定义Metrics模块的具体实现要求，基于设计文档`docs/design/metrics.md`和设计规格。

## ADDED Requirements

### Requirement: 核心数据结构实现

实现MUST包含Metrics模块的核心数据结构，包括MetricsService和DeviceMetrics结构体，确保所有字段和方法都符合设计要求和接口规范。

#### Scenario: MetricsService结构体定义
- **给定**: 需要创建Metrics模块核心服务
- **当**: 实现MetricsService时
- **那么**: 应包含以下字段：
  ```go
  type MetricsService struct {
      registry       *prometheus.Registry
      collector      collector.CollectorInterface
      logger         *zap.Logger

      // Exporter自监控指标
      lastCollectionTime prometheus.Gauge
      collectionDuration *prometheus.HistogramVec
      scrapeErrorsTotal  *prometheus.CounterVec
      memoryUsage        *prometheus.GaugeVec
      deviceCount        prometheus.Gauge

      // WinPower连接指标
      connectionStatus   prometheus.Gauge
      tokenExpiry        prometheus.Gauge

      // 设备指标映射
      deviceMetrics      map[string]*DeviceMetrics
      mu                 sync.RWMutex
  }
  ```

#### Scenario: DeviceMetrics结构体实现
- **给定**: 需要为每个设备创建指标集合
- **当**: 创建DeviceMetrics时
- **那么**: 应包含所有设备相关指标：
  ```go
  type DeviceMetrics struct {
      connected        prometheus.Gauge
      loadPercent      prometheus.Gauge
      inputVoltage     prometheus.Gauge
      outputVoltage    prometheus.Gauge
      inputCurrent     prometheus.Gauge
      outputCurrent    prometheus.Gauge
      inputFrequency   prometheus.Gauge
      outputFrequency  prometheus.Gauge
      activePower      prometheus.Gauge
      instantPower     prometheus.Gauge
      cumulativeEnergy prometheus.Gauge
      batteryCharging  prometheus.Gauge
      batteryCapacity  prometheus.Gauge
      upsTemperature   prometheus.Gauge
      upsMode          prometheus.Gauge
      upsStatus        prometheus.Gauge
      faultCode        *prometheus.GaugeVec
  }
  ```

### Requirement: 指标初始化和注册

实现MUST包含完整的指标初始化和注册逻辑，包括Exporter自监控指标、WinPower连接指标和动态设备指标的创建、配置和Prometheus注册表管理。

#### Scenario: 初始化Exporter自监控指标
- **给定**: 需要创建Exporter自监控指标
- **当**: 调用NewMetricsService时
- **那么**: 应创建并注册以下指标：
  - `winpower_exporter_up` (Gauge)
  - `winpower_exporter_requests_total` (CounterVec)
  - `winpower_exporter_request_duration_seconds` (HistogramVec)
  - `winpower_exporter_collection_duration_seconds` (HistogramVec)
  - `winpower_exporter_scrape_errors_total` (CounterVec)
  - `winpower_exporter_memory_bytes` (GaugeVec)
  - `winpower_exporter_device_count` (Gauge)

#### Scenario: 初始化WinPower连接指标
- **给定**: 需要监控与WinPower的连接状态
- **当**: 创建连接指标时
- **那么**: 应创建并注册：
  - `winpower_connection_status` (Gauge)
  - `winpower_auth_status` (Gauge)
  - `winpower_api_response_time_seconds` (HistogramVec)
  - `winpower_token_expiry_seconds` (Gauge)
  - `winpower_token_valid` (Gauge)

#### Scenario: 动态创建设备指标
- **给定**: 发现新设备或设备信息更新
- **当**: 调用createDeviceMetrics方法时
- **那么**: 应为设备创建完整指标集合
- **并且**: 使用`winpower_host`、`device_id`、`device_name`、`device_type`标签
- **并且**: 将DeviceMetrics实例存储在deviceMetrics映射中

### Requirement: HTTP Handler实现

实现MUST包含标准的HTTP Handler来处理Prometheus的/metrics端点请求，包括请求路由、数据采集触发、指标更新和响应格式化等完整的HTTP处理流程。

#### Scenario: 处理/metrics端点请求
- **给定**: Prometheus发送GET请求到/metrics端点
- **当**: HandleMetrics方法被调用时
- **那么**: 应执行以下流程：
  1. 更新请求计数指标
  2. 调用Collector.CollectDeviceData获取最新数据
  3. 如果采集失败，记录错误并返回500状态码
  4. 如果成功，调用updateMetrics更新所有指标
  5. 使用registry.ServeHTTP返回Prometheus格式数据
  6. 记录请求处理耗时

#### Scenario: 处理Collector调用超时
- **给定**: Collector调用可能超时
- **当**: CollectDeviceData调用超过设定时间时
- **那么**: 应取消上下文并记录超时错误
- **并且**: 返回HTTP 504状态码
- **并且**: 更新超时错误统计指标

### Requirement: 指标更新逻辑实现

实现MUST包含完整的指标更新逻辑，包括自监控指标更新、设备指标更新、内存监控等，确保Collector返回的数据能够正确、及时地反映到Prometheus指标中。

#### Scenario: 更新自监控指标
- **给定**: 采集完成或请求处理完成
- **当**: 调用updateSelfMetrics方法时
- **那么**: 应更新：
  - 最后采集时间戳
  - 采集耗时Histogram
  - 内存使用情况
  - 发现的设备数量

#### Scenario: 更新设备指标
- **给定**: Collector返回设备采集数据
- **当**: 调用updateDeviceMetrics方法时
- **那么**: 应遍历所有设备数据：
  1. 检查设备是否已存在指标，不存在则创建
  2. 更新所有设备相关指标值
  3. 特别处理LoadTotalWatt到instantPower的映射
  4. 更新能耗计算结果指标
  5. 处理故障代码标签

#### Scenario: 内存指标监控
- **给定**: 需要监控exporter内存使用情况
- **当**: 调用updateMemoryMetrics方法时
- **那么**: 应使用runtime.ReadMemStats获取：
  - Alloc (已分配内存)
  - Sys (系统内存)
  - HeapAlloc (堆内存)
- **并且**: 更新相应的Gauge指标

### Requirement: 错误处理和日志记录实现

实现MUST包含完善的错误处理和日志记录机制，包括错误分类、详细日志记录、错误统计和错误恢复策略，确保系统在各种异常情况下都能保持稳定运行。

#### Scenario: Collector错误分类处理
- **给定**: Collector调用返回不同类型的错误
- **当**: handleCollectorError方法被调用时
- **那么**: 应根据错误类型更新相应指标：
  - WinPower采集错误：`error_type="winpower_collection"`
  - 电能计算错误：`error_type="energy_calculation"`
  - 网络超时错误：`error_type="network_timeout"`
- **并且**: 记录包含错误详情的结构化日志

#### Scenario: 设备指标更新失败处理
- **给定**: 某个设备指标更新失败
- **当**: updateDeviceMetrics处理设备时
- **那么**: 应记录设备级错误但不影响其他设备
- **并且**: 在DeviceCollectionInfo中设置ErrorMsg字段
- **并且**: 继续处理其他设备指标

### Requirement: 并发安全和性能优化实现

实现MUST包含并发安全机制和性能优化策略，包括读写锁保护、设备指标动态管理、内存清理和并发访问控制，确保系统在高并发场景下的稳定性和性能。

#### Scenario: 读写锁保护并发访问
- **给定**: 多个goroutine同时访问设备指标
- **当**: updateDeviceMetrics被调用时
- **那么**: 应使用m.mu.Lock()保护写操作
- **并且**: 其他读操作使用m.mu.RLock()保护
- **并且**: 确保锁的释放使用defer语句

#### Scenario: 设备指标清理机制
- **给定**: 长时间未更新的设备指标占用内存
- **当**: 定期清理任务执行时
- **那么**: 应遍历deviceMetrics映射
- **并且**: 删除超过阈值时间（如5分钟）未更新的设备指标
- **并且**: 记录清理操作的日志

### Requirement: 测试实现要求

实现MUST包含完整的测试覆盖，包括单元测试、集成测试、Mock依赖测试和并发安全测试，确保Metrics模块的功能正确性、性能指标和质量标准。

#### Scenario: 单元测试覆盖
- **给定**: 需要确保Metrics模块质量
- **当**: 编写测试时
- **那么**: 应为以下方法编写测试：
  - NewMetricsService - 服务创建和初始化
  - HandleMetrics - HTTP请求处理
  - updateMetrics - 指标更新逻辑
  - updateDeviceMetrics - 设备指标更新
  - createDeviceMetrics - 设备指标创建
  - 错误处理方法

#### Scenario: Mock依赖测试
- **给定**: 需要隔离外部依赖进行测试
- **当**: 创建测试环境时
- **那么**: 应提供以下Mock：
  - MockCollectorInterface - 模拟Collector模块
  - 使用prometheus/testutil进行指标断言
  - 使用gin测试框架进行HTTP handler测试

### Requirement: 配置和常量定义

实现MUST包含完整的配置和常量体系，包括指标命名规范、标签定义、Histogram桶配置等，确保代码的可维护性和配置的统一性。

#### Scenario: 指标命名和标签常量
- **给定**: 需要统一指标命名规范
- **当**: 定义指标常量时
- **那么**: 应定义以下常量：
  ```go
  const (
      Namespace = "winpower"
      ExporterSubsystem = "exporter"
      DeviceSubsystem = "device"
      ConnectionSubsystem = "connection"
  )

  var (
      LabelHost = "winpower_host"
      LabelDeviceID = "device_id"
      LabelDeviceName = "device_name"
      LabelDeviceType = "device_type"
      LabelFaultCode = "fault_code"
      LabelErrorType = "error_type"
  )
  ```

#### Scenario: Histogram桶配置
- **给定**: 需要配置合适的Histogram桶
- **当**: 创建Histogram指标时
- **那么**: 应使用预定义桶配置：
  ```go
  var (
      RequestDurationBuckets = []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5}
      APIDurationBuckets = []float64{0.05, 0.1, 0.2, 0.5, 1}
  )
  ```

## 实现约束

- 必须使用Go 1.25+语法和标准库
- 必须使用prometheus客户端库进行指标管理
- 必须使用zap进行结构化日志记录
- 必须使用gin框架处理HTTP请求
- 所有公共方法必须有完整文档注释
- 所有错误处理必须包含详细的上下文信息
- 必须通过`make lint`静态检查
- 单元测试覆盖率必须达到80%以上