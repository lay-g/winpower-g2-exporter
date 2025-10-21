# Metrics 模块配置说明

本文档详细描述了 Metrics 模块的配置参数和使用方法。

## 配置结构

### MetricManagerConfig

```go
type MetricManagerConfig struct {
    Namespace string                `json:"namespace" yaml:"namespace"`
    Subsystem string                `json:"subsystem" yaml:"subsystem"`
    Registry  *prometheus.Registry  `json:"-" yaml:"-"`

    // 直方图桶配置
    RequestDurationBuckets     []float64 `json:"request_duration_buckets" yaml:"request_duration_buckets"`
    CollectionDurationBuckets  []float64 `json:"collection_duration_buckets" yaml:"collection_duration_buckets"`
    APIResponseBuckets         []float64 `json:"api_response_buckets" yaml:"api_response_buckets"`
}
```

## 配置参数详解

### 基础配置

#### Namespace

**描述**: Prometheus 指标的命名空间前缀

**类型**: `string`

**默认值**: `"winpower"`

**示例**:
```yaml
metrics:
  namespace: "winpower"
```

**说明**:
- 用于构建完整的指标名称：`{namespace}_{subsystem}_{metric_name}`
- 建议使用项目或服务名称
- 必须符合 Prometheus 命名规范（仅包含字母、数字、下划线）

#### Subsystem

**描述**: Prometheus 指标的子系统标识

**类型**: `string`

**默认值**: `"exporter"`

**示例**:
```yaml
metrics:
  subsystem: "exporter"
```

**说明**:
- 用于标识指标的子系统或模块
- 常用值：`exporter`, `collector`, `monitor`
- 与 namespace 组合形成完整的指标前缀

#### Registry

**描述**: 自定义 Prometheus 注册表

**类型**: `*prometheus.Registry`

**默认值**: `nil`（自动创建新的注册表）

**示例**:
```go
// 使用自定义注册表
customRegistry := prometheus.NewRegistry()
cfg := metrics.MetricManagerConfig{
    Namespace: "winpower",
    Subsystem: "exporter",
    Registry:  customRegistry,
}

mm := metrics.NewMetricManager(cfg, logger)
```

**说明**:
- 高级配置，通常不需要设置
- 用于与现有的 Prometheus 注册表集成
- 如果不设置，会自动创建新的注册表

### 直方图配置

#### RequestDurationBuckets

**描述**: HTTP 请求持续时间的直方图桶边界

**类型**: `[]float64`

**默认值**: `[0.05, 0.1, 0.2, 0.5, 1, 2, 5]`

**单位**: 秒

**示例**:
```yaml
metrics:
  request_duration_buckets: [0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
```

**说明**:
- 用于 `winpower_exporter_request_duration_seconds` 指标
- 桶边界应该根据实际的请求延迟分布来调整
- 第一个桶应该小于最小的预期延迟
- 最后一个桶应该大于最大的预期延迟

#### CollectionDurationBuckets

**描述**: 数据采集持续时间的直方图桶边界

**类型**: `[]float64`

**默认值**: `[0.1, 0.2, 0.5, 1, 2, 5, 10]`

**单位**: 秒

**示例**:
```yaml
metrics:
  collection_duration_buckets: [0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 15, 30]
```

**说明**:
- 用于 `winpower_exporter_collection_duration_seconds` 指标
- 桶边界应该根据数据采集的实际耗时来调整
- 考虑网络延迟、设备响应时间等因素

#### APIResponseBuckets

**描述**: WinPower API 响应时间的直方图桶边界

**类型**: `[]float64`

**默认值**: `[0.05, 0.1, 0.2, 0.5, 1]`

**单位**: 秒

**示例**:
```yaml
metrics:
  api_response_buckets: [0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5]
```

**说明**:
- 用于 `winpower_api_response_time_seconds` 指标
- 桶边界应该根据 WinPower API 的响应特性来调整
- 考虑本地网络和远程服务器的响应时间

## 配置文件格式

### YAML 配置

```yaml
# config.yaml
metrics:
  # 基础配置
  namespace: "winpower"          # 指标命名空间
  subsystem: "exporter"          # 子系统标识

  # 直方图桶配置（可选）
  request_duration_buckets:       # HTTP 请求时延桶（秒）
    - 0.05
    - 0.1
    - 0.2
    - 0.5
    - 1
    - 2
    - 5

  collection_duration_buckets:   # 采集时延桶（秒）
    - 0.1
    - 0.2
    - 0.5
    - 1
    - 2
    - 5
    - 10

  api_response_buckets:          # API 响应时延桶（秒）
    - 0.05
    - 0.1
    - 0.2
    - 0.5
    - 1
```

### JSON 配置

```json
{
  "metrics": {
    "namespace": "winpower",
    "subsystem": "exporter",
    "request_duration_buckets": [0.05, 0.1, 0.2, 0.5, 1, 2, 5],
    "collection_duration_buckets": [0.1, 0.2, 0.5, 1, 2, 5, 10],
    "api_response_buckets": [0.05, 0.1, 0.2, 0.5, 1]
  }
}
```

## 环境变量配置

### 命名规范

环境变量使用 `WINPOWER_EXPORTER_METRICS_` 前缀，遵循以下转换规则：

1. 配置字段名转换为大写
2. 下划线替换嵌套结构
3. 添加前缀

### 环境变量列表

| 环境变量 | 对应配置字段 | 类型 | 默认值 | 示例 |
|---------|-------------|------|-------|------|
| `WINPOWER_EXPORTER_METRICS_NAMESPACE` | `metrics.namespace` | string | `winpower` | `WINPOWER_EXPORTER_METRICS_NAMESPACE=myapp` |
| `WINPOWER_EXPORTER_METRICS_SUBSYSTEM` | `metrics.subsystem` | string | `exporter` | `WINPOWER_EXPORTER_METRICS_SUBSYSTEM=collector` |

### 环境变量配置示例

```bash
# 基础配置
export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter"

# 启动应用
./winpower-exporter --config config.yaml
```

## 配置加载示例

### 从文件加载

```go
package main

import (
    "os"
    "gopkg.in/yaml.v2"
    "go.uber.org/zap"
    "your-project/internal/metrics"
)

type Config struct {
    Metrics metrics.MetricManagerConfig `yaml:"metrics"`
}

func loadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

func main() {
    // 加载配置文件
    cfg, err := loadConfig("config.yaml")
    if err != nil {
        panic(err)
    }

    // 创建指标管理器
    logger := zap.NewProduction()
    mm := metrics.NewMetricManager(cfg.Metrics, logger)

    // 使用指标管理器...
}
```

### 命令行参数覆盖

```go
package main

import (
    "flag"
    "go.uber.org/zap"
    "your-project/internal/metrics"
)

func main() {
    // 命令行参数
    var (
        metricsNamespace = flag.String("metrics-namespace", "winpower", "Metrics namespace")
        metricsSubsystem = flag.String("metrics-subsystem", "exporter", "Metrics subsystem")
    )
    flag.Parse()

    // 构建配置
    cfg := metrics.MetricManagerConfig{
        Namespace: *metricsNamespace,
        Subsystem: *metricsSubsystem,
    }

    // 创建指标管理器
    logger := zap.NewProduction()
    mm := metrics.NewMetricManager(cfg, logger)

    // 使用指标管理器...
}
```

## 配置验证

### 基本验证规则

```go
func (cfg *MetricManagerConfig) Validate() error {
    // 验证命名空间
    if cfg.Namespace == "" {
        return fmt.Errorf("metrics namespace cannot be empty")
    }
    if !isValidPrometheusName(cfg.Namespace) {
        return fmt.Errorf("invalid metrics namespace: %s", cfg.Namespace)
    }

    // 验证子系统
    if cfg.Subsystem == "" {
        return fmt.Errorf("metrics subsystem cannot be empty")
    }
    if !isValidPrometheusName(cfg.Subsystem) {
        return fmt.Errorf("invalid metrics subsystem: %s", cfg.Subsystem)
    }

    // 验证直方图桶
    if err := validateBuckets(cfg.RequestDurationBuckets, "request_duration"); err != nil {
        return err
    }
    if err := validateBuckets(cfg.CollectionDurationBuckets, "collection_duration"); err != nil {
        return err
    }
    if err := validateBuckets(cfg.APIResponseBuckets, "api_response"); err != nil {
        return err
    }

    return nil
}

func isValidPrometheusName(name string) bool {
    // Prometheus 指标名称规范：只能包含字母、数字、下划线，且不能以数字开头
    matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
    return matched
}

func validateBuckets(buckets []float64, name string) error {
    if len(buckets) == 0 {
        return fmt.Errorf("%s buckets cannot be empty", name)
    }

    // 验证桶边界递增
    for i := 1; i < len(buckets); i++ {
        if buckets[i] <= buckets[i-1] {
            return fmt.Errorf("%s buckets must be in increasing order", name)
        }
    }

    // 验证所有桶都为正数
    for i, bucket := range buckets {
        if bucket <= 0 {
            return fmt.Errorf("%s bucket[%d] must be positive", name, i)
        }
    }

    return nil
}
```

### 配置最佳实践

#### 1. 桶边界选择

```go
// 基于实际延迟数据选择桶边界
func calculateOptimalBuckets(durations []time.Duration) []float64 {
    if len(durations) == 0 {
        return []float64{0.1, 0.5, 1, 2, 5}
    }

    // 转换为秒数并排序
    seconds := make([]float64, len(durations))
    for i, d := range durations {
        seconds[i] = d.Seconds()
    }
    sort.Float64s(seconds)

    // 计算百分位数
    p50 := percentile(seconds, 0.5)
    p90 := percentile(seconds, 0.9)
    p95 := percentile(seconds, 0.95)
    p99 := percentile(seconds, 0.99)

    // 生成桶边界
    return []float64{
        p50 / 10,    // 小于 50% 的最小延迟
        p50,
        p90,
        p95,
        p99,
        p99 * 2,     // 覆盖异常情况
    }
}
```

#### 2. 环境特定配置

```yaml
# config.development.yaml
metrics:
  namespace: "winpower"
  subsystem: "exporter-dev"
  request_duration_buckets: [0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5]

---

# config.production.yaml
metrics:
  namespace: "winpower"
  subsystem: "exporter"
  request_duration_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30]
```

## 配置迁移

### 从旧版本迁移

如果需要从旧版本的配置格式迁移，可以提供迁移工具：

```go
func migrateConfig(oldConfig OldConfig) metrics.MetricManagerConfig {
    return metrics.MetricManagerConfig{
        Namespace: oldConfig.PrometheusNamespace,
        Subsystem: oldConfig.PrometheusSubsystem,
        RequestDurationBuckets: oldConfig.HTTPBuckets,
        CollectionDurationBuckets: oldConfig.ScrapeBuckets,
        APIResponseBuckets: oldConfig.APICallBuckets,
    }
}
```

## 故障排除

### 常见配置错误

1. **命名空间包含无效字符**
   ```
   错误: Error: invalid metrics namespace: my-app
   解决: 使用 "my_app" 或 "myapp" 替代
   ```

2. **直方图桶未按递增顺序排列**
   ```
   错误: Error: request_duration buckets must be in increasing order
   解决: 确保桶边界是递增的，如 [0.1, 0.5, 1, 2, 5]
   ```

3. **桶边界包含负值或零**
   ```
   错误: Error: collection_duration bucket[0] must be positive
   解决: 确保所有桶边界都为正数
   ```

### 调试配置

```go
func debugMetricsConfig(cfg metrics.MetricManagerConfig) {
    fmt.Printf("Metrics Configuration:\n")
    fmt.Printf("  Namespace: %s\n", cfg.Namespace)
    fmt.Printf("  Subsystem: %s\n", cfg.Subsystem)
    fmt.Printf("  Request Buckets: %v\n", cfg.RequestDurationBuckets)
    fmt.Printf("  Collection Buckets: %v\n", cfg.CollectionDurationBuckets)
    fmt.Printf("  API Response Buckets: %v\n", cfg.APIResponseBuckets)

    // 预览指标名称
    fmt.Printf("  Example Metric Names:\n")
    fmt.Printf("    winpower_exporter_up\n")
    fmt.Printf("    %s_%s_requests_total\n", cfg.Namespace, cfg.Subsystem)
}
```

这份配置说明文档提供了完整的配置参数说明、示例和最佳实践，帮助用户正确配置 Metrics 模块。