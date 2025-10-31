# implement-metrics-module: Metrics模块实现提案

## 概述

本提案旨在实现WinPower G2 Prometheus Exporter的Metrics模块，该模块负责指标的管理、更新和Prometheus格式的暴露。根据设计文档`docs/design/metrics.md`，Metrics模块作为指标管理中心，与Collector模块协调获取最新数据，并提供标准的HTTP Handler。

## 提案目标

- 实现完整的Metrics模块，包括指标定义、注册表管理和HTTP Handler
- 提供与Collector模块的协调接口，触发数据采集并更新指标
- 实现双类指标：Exporter自监控指标和WinPower设备指标
- 确保与现有架构的兼容性，不考虑Server模块集成

## 设计参考

本提案基于以下设计文档：
- `docs/design/metrics.md` - Metrics模块完整设计文档
- `docs/design/collector.md` - Collector模块设计文档（作为依赖模块）
- `openspec/project.md` - 项目约定和架构模式

## 变更内容

### 能力变更

1. **指标管理能力** - 实现完整的Prometheus指标注册表和管理功能
2. **数据协调能力** - 与Collector模块协调，触发数据采集并更新指标
3. **HTTP处理能力** - 提供标准的/metrics端点处理Prometheus抓取请求

### 详细变更

请参见各能力规格文档：
- `specs/design/` - 设计层面的技术规格
- `specs/implementation/` - 实现层面的技术规格

## 实现任务

参见`tasks.md`文件了解详细的实现任务列表。

## 验证标准

- 所有单元测试通过（`make test`）
- 静态检查通过（`make lint`）
- 指标格式符合Prometheus规范
- 与Collector模块集成正常工作