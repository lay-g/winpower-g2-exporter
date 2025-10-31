# Metrics模块设计文档

## 架构设计

Metrics模块作为WinPower G2 Prometheus Exporter的指标管理中心，负责指标的创建、更新和暴露。基于设计文档`docs/design/metrics.md`，本模块与Collector模块协调获取最新数据，并提供标准的Prometheus指标暴露能力。

### 模块边界和职责

- **指标管理**: 创建和维护所有Prometheus指标
- **数据协调**: 与Collector模块交互，触发数据采集
- **HTTP处理**: 提供/metrics端点处理Prometheus抓取请求
- **错误处理**: 处理采集失败和指标更新错误
- **并发控制**: 确保多线程环境下的数据一致性

### 不包含的内容

- Server模块集成（后续工作）
- 调度器集成（由Collector模块处理）
- WinPower客户端实现（由WinPower模块处理）
- 电能计算逻辑（由Energy模块处理）

## 数据流设计

```text
Prometheus抓取请求
        │
        ▼
HTTP请求 → HandleMetrics() → 调用Collector.CollectDeviceData()
        │                      │
        │                      ▼
        │              返回CollectionResult
        │                      │
        ▼                      ▼
  更新所有指标指标 ←─────┘
        │
        ▼
  返回Prometheus格式数据
```

## 关键设计决策

1. **统一数据入口**: 所有指标更新都通过Collector.CollectDeviceData()触发
2. **动态设备指标**: 根据发现的设备动态创建指标实例
3. **内存管理**: 定期清理不活跃设备的指标，避免内存泄漏
4. **错误分类**: 对不同类型的错误进行分类统计
5. **并发安全**: 使用读写锁保护指标更新操作

## 与现有架构的集成

Metrics模块依赖于Collector模块的接口，确保数据流的统一性。模块设计遵循项目的分层架构原则，与其他模块保持清晰的边界。

## 性能考虑

- 使用读写锁优化并发访问性能
- 批量更新指标减少锁竞争
- 延迟清理策略避免频繁的内存分配
- 合理的Histogram桶配置平衡精度和性能