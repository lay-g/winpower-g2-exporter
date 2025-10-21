# Metrics 模块配置示例

本目录包含了 WinPower G2 Exporter Metrics 模块的各种配置示例，适用于不同的部署场景和性能需求。

## 配置文件列表

### 基础配置

- **basic-config.yaml** - 标准生产环境配置
  - 适用于大多数生产环境
  - 平衡的性能和可靠性
  - 标准的桶边界设置

### 环境特定配置

- **development-config.yaml** - 开发环境配置
  - 更精细的桶边界便于调试
  - 开发友好的日志格式
  - 较短的调度间隔

- **production-config.yaml** - 生产环境配置
  - 高可用性和安全性配置
  - 生产级别的日志和存储
  - 完整的监控和限流配置

- **high-throughput-config.yaml** - 高吞吐量配置
  - 适用于监控大量设备
  - 性能优化设置
  - 并发处理和缓存配置

### 特殊场景配置

- **custom-buckets-config.yaml** - 自定义桶配置示例
  - 展示不同场景下的桶边界设置
  - 包含详细的配置说明
  - 便于根据实际需求调整

## 配置选择指南

### 根据部署环境选择

| 环境 | 推荐配置 | 特点 |
|------|---------|------|
| 开发测试 | development-config.yaml | 快速反馈，详细日志 |
| 小规模生产 | basic-config.yaml | 简单稳定，易于维护 |
| 大规模生产 | production-config.yaml | 高可用，完整监控 |
| 高吞吐量 | high-throughput-config.yaml | 性能优化，并发处理 |

### 根据设备数量选择

| 设备数量 | 推荐配置 | 调度间隔 | 并发数 |
|---------|---------|---------|-------|
| 1-10 台 | development-config.yaml | 2s | 单线程 |
| 10-50 台 | basic-config.yaml | 5s | 单线程 |
| 50-200 台 | production-config.yaml | 5s | 5 线程 |
| 200+ 台 | high-throughput-config.yaml | 10s | 10+ 线程 |

### 根据网络环境选择

| 网络环境 | 推荐配置 | 超时设置 | 桶边界 |
|---------|---------|---------|--------|
| 本地网络 | development-config.yaml | 10s | 精细桶 |
| 局域网 | basic-config.yaml | 30s | 标准桶 |
| 广域网 | production-config.yaml | 45s | 宽松桶 |
| 不稳定网络 | high-throughput-config.yaml | 60s | 扩展桶 |

## 使用方法

### 1. 直接使用配置文件

```bash
# 使用基础配置
./winpower-exporter --config internal/metrics/examples/basic-config.yaml

# 使用生产配置
./winpower-exporter --config internal/metrics/examples/production-config.yaml
```

### 2. 复制并修改配置

```bash
# 复制基础配置
cp internal/metrics/examples/basic-config.yaml ./config.yaml

# 根据需要修改配置
vim ./config.yaml

# 使用自定义配置
./winpower-exporter --config ./config.yaml
```

### 3. 环境变量覆盖

```bash
# 使用基础配置，但覆盖特定参数
export WINPOWER_EXPORTER_METRICS_NAMESPACE="myapp"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="monitor"

./winpower-exporter --config internal/metrics/examples/basic-config.yaml
```

## 配置验证

### 验证配置文件格式

```bash
# 使用 Go 验证 YAML 格式
go run -c "package main; import (\"gopkg.in/yaml.v2\"; \"os\"; \"fmt\"); func main() { var cfg interface{}; data, _ := os.ReadFile(\"internal/metrics/examples/basic-config.yaml\"); err := yaml.Unmarshal(data, &cfg); if err != nil { fmt.Printf(\"Error: %v\\n\", err); os.Exit(1) } else { fmt.Println(\"Config format is valid\") } }"
```

### 验证配置参数

```bash
# 启动时验证配置（使用 --validate 参数）
./winpower-exporter --config internal/metrics/examples/basic-config.yaml --validate
```

## 桶边界调整指南

### HTTP 请求时延桶

根据实际的 HTTP 响应时间分布调整：

```yaml
# 快速环境（大部分请求 < 100ms）
request_duration_buckets: [0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1]

# 标准环境（大部分请求 < 1s）
request_duration_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2, 5]

# 慢速环境（大部分请求 < 10s）
request_duration_buckets: [0.5, 1, 2.5, 5, 10, 25, 60]
```

### 采集时延桶

根据设备数量和网络状况调整：

```yaml
# 少量设备，快速网络
collection_duration_buckets: [0.1, 0.2, 0.5, 1, 2, 5]

# 中等规模设备
collection_duration_buckets: [0.5, 1, 2.5, 5, 10, 30]

# 大量设备或慢速网络
collection_duration_buckets: [1, 2.5, 5, 15, 30, 60, 120]
```

### API 响应时延桶

根据 WinPower 服务器的响应特性调整：

```yaml
# 快速 API 服务器
api_response_buckets: [0.01, 0.025, 0.05, 0.1, 0.25, 0.5]

# 标准 API 服务器
api_response_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2.5]

# 慢速 API 服务器
api_response_buckets: [0.1, 0.25, 0.5, 1, 2.5, 5, 10]
```

## 性能调优建议

### 高吞吐量场景

1. **增加并发数**
   ```yaml
   scheduler:
     max_workers: 10
     worker_queue_size: 1000
   ```

2. **启用缓存**
   ```yaml
   storage:
     enable_cache: true
     cache_size: 10000
   ```

3. **异步日志**
   ```yaml
   logging:
     async_logging: true
     buffer_size: 10000
   ```

### 低延迟场景

1. **减少调度间隔**
   ```yaml
   scheduler:
     interval: 2s
   ```

2. **优化桶边界**
   ```yaml
   metrics:
     request_duration_buckets: [0.001, 0.005, 0.01, 0.025, 0.05, 0.1]
   ```

3. **禁用同步写入**
   ```yaml
   storage:
     sync_write: false
   ```

## 故障排除

### 常见配置问题

1. **桶边界不递增**
   ```
   错误: Error: buckets must be in increasing order
   解决: 确保桶边界值是递增的
   ```

2. **超时设置过短**
   ```
   现象: 频繁出现采集超时
   解决: 增加 winpower.timeout 和相关的桶边界
   ```

3. **内存使用过高**
   ```
   现象: 内存持续增长
   解决: 检查缓存设置，减少 worker 数量
   ```

### 调试配置

1. **启用调试日志**
   ```yaml
   logging:
     level: "debug"
   ```

2. **启用 pprof**
   ```yaml
   monitoring:
     enable_pprof: true
   ```

3. **验证指标输出**
   ```bash
   curl http://localhost:9090/metrics | head -20
   ```

## 自定义配置模板

### 创建自定义配置模板

1. 从最接近的示例配置开始
2. 根据实际环境调整参数
3. 测试配置的有效性
4. 监控性能指标并根据需要进一步调整

### 配置最佳实践

1. **使用环境变量存储敏感信息**
2. **定期备份配置文件**
3. **使用版本控制管理配置变更**
4. **在生产环境部署前充分测试配置**
5. **监控配置变更对性能的影响**