# implement-metrics-module 变更提案

## 概述

本提案旨在实现 WinPower G2 Exporter 的指标管理模块（metrics），该模块负责将采集的设备数据转化为 Prometheus 兼容的时序指标，并通过 HTTP `/metrics` 端点统一暴露。

**设计原则**：遵循 Prometheus 指标规范，提供简洁、标准化、高性能的指标管理，与已实现的 winpower、energy、storage、logging 模块紧密集成，支持完整的可观测性需求。

## Why

### 当前痛点
- **监控缺失**：缺少标准化的 Prometheus 指标暴露机制，无法与监控系统集成
- **架构不完整**：作为数据导出的关键环节，metrics模块的缺失导致整体架构无法闭环
- **可观测性不足**：无法提供 Exporter 自监控和设备状态的可视化指标

### 业务价值
- **监控集成**：提供标准的 Prometheus 指标，无缝接入现有监控体系
- **运维可视化**：支持设备状态、能耗趋势、连接状态等多维度监控
- **告警支持**：为基础监控和告警系统提供数据支撑
- **性能分析**：提供采集耗时、API响应时间等性能指标

### 技术必要性
- **架构完整性**：metrics模块是数据采集到监控展示的最终环节
- **标准符合**：遵循 Prometheus 指标命名和标签规范
- **性能优化**：采用单 Registry 和最小锁粒度设计，确保高性能

## What Changes

### 新增模块
- **metrics-manager**: 指标管理服务，提供 Prometheus 指标的注册、更新和暴露
- **metrics-config**: metrics模块配置管理，支持命名空间和自定义配置

### 新增接口
- **MetricManagerInterface**: 定义指标管理的标准接口
- **Handler()**: 返回 /metrics 的 HTTP Handler
- **UpdateMetrics()**: 批量更新设备和连接指标

### 新增功能
- **指标注册**: 管理 Exporter 自监控、连接状态、设备数据、能耗四类指标
- **HTTP暴露**: 提供 `/metrics` 端点，支持 Prometheus 抓取
- **标签管理**: 统一标签策略，控制基数避免高基数风险
- **性能观测**: 集成请求时延、采集耗时等性能指标

## 能力范围

### 新增能力
- **metrics-management**: 实现完整的指标管理服务
- **metrics-config**: 定义metrics模块的配置结构
- **prometheus-integration**: 与 Prometheus 生态的完整集成

### 边界约束
- 不触发数据采集，仅提供指标更新和暴露功能
- 完全依赖 winpower 和 energy 模块提供的数据
- 不存储历史数据，仅维护当前指标状态
- 支持单个 Registry，避免多 Registry 冲突

## 架构影响

### 模块依赖
- **依赖**: logging模块（已实现）、winpower模块（已实现）、energy模块（已实现）
- **被依赖**: server模块（待实现）
- **配置**: 在metrics模块内定义Config结构，由config模块统一加载

### 接口设计
```go
type MetricManagerInterface interface {
    Handler() http.Handler
    UpdateDeviceMetrics(data DeviceData) error
    UpdateConnectionMetrics(status ConnectionStatus) error
    UpdateEnergyMetrics(deviceID string, power, energy float64) error
}
```

### 核心组件
- **MetricManager**: 基于 Prometheus Registry 的指标管理器
- **指标分类**: Exporter自监控、WinPower连接/认证、设备/电源、能耗四类
- **标签策略**: 设备标签采用 device_id/device_name/device_type 组合

## 实现策略

### 阶段1: 核心结构与接口定义
- 定义MetricManagerInterface接口和MetricManager结构
- 实现基础的Prometheus指标注册
- 定义metrics模块配置结构

### 阶段2: 指标管理核心功能
- 实现四类指标的创建和管理
- 实现类型安全的指标更新方法
- 集成logging模块提供结构化日志

### 阶段3: HTTP暴露与集成
- 实现 /metrics 端点的 HTTP Handler
- 提供指标快照功能
- 与winpower和energy模块集成测试

### 阶段4: 测试与验证
- 编写全面的单元测试，使用 prometheus/testutil
- 进行端到端测试验证指标格式
- 性能测试验证指标更新和暴露性能

## 验收标准

1. **指标完整性**: 正确实现四类指标的注册和更新
2. **标准符合**: 遵循 Prometheus 指标命名和标签规范
3. **性能要求**: 指标更新延迟<5ms，/metrics响应时间<50ms
4. **测试覆盖**: 单元测试覆盖率达到90%以上
5. **集成测试**: 与已实现模块正确集成，数据流转正常
6. **错误处理**: 优雅处理各种异常情况，不影响其他模块

## 风险与缓解

### 主要风险
- **指标基数**: 设备标签可能导致高基数问题
- **性能瓶颈**: 频繁指标更新可能影响整体性能
- **内存使用**: 大量时序指标可能导致内存占用过高

### 缓解措施
- 采用严格标签策略，控制标签枚举值
- 使用最小锁粒度，优化热路径性能
- 限制指标数量，避免不必要的时序数据
- 提供详细的监控指标和日志

## 后续演进

- 支持自定义指标和标签扩展
- 支持指标过滤和采样
- 集成更多可观测性特性
- 支持多 Registry 场景（如需要）