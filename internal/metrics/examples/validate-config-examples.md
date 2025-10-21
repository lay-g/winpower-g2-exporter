# 配置验证工具使用指南

本文档介绍了如何使用 Metrics 模块的配置验证工具来验证配置文件和环境变量。

## 工具概述

配置验证工具是一个独立的命令行程序，用于验证 WinPower G2 Exporter Metrics 模块的配置文件和环境变量设置。

### 功能特性

- ✅ 验证 YAML/JSON 配置文件格式
- ✅ 验证环境变量设置
- ✅ 提供详细的错误和警告信息
- ✅ 生成配置优化建议
- ✅ 支持 JSON 格式输出
- ✅ 批量验证多个配置

## 安装和构建

### 构建验证工具

```bash
# 进入项目根目录
cd /path/to/winpower-g2-exporter

# 构建验证工具
go build -o bin/validate-config ./internal/metrics/cmd/validate-config

# 或者使用 make 命令（如果配置了）
make validate-config
```

### 验证工具可执行文件

构建完成后，验证工具将位于 `bin/validate-config`（或根据你的构建配置）。

## 基本使用方法

### 验证配置文件

```bash
# 验证 YAML 配置文件
./bin/validate-config -config config.yaml

# 验证 JSON 配置文件
./bin/validate-config -config config.json

# 显示详细输出
./bin/validate-config -config config.yaml -verbose

# 输出 JSON 格式结果
./bin/validate-config -config config.yaml -json
```

### 验证环境变量

```bash
# 只验证环境变量
./bin/validate-config -env-only

# 验证环境变量并显示详细信息
./bin/validate-config -env-only -verbose
```

### 同时验证配置文件和环境变量

```bash
# 验证配置文件和环境变量
./bin/validate-config -config config.yaml

# 显示完整验证结果
./bin/validate-config -config config.yaml -verbose

# 输出 JSON 格式
./bin/validate-config -config config.yaml -json
```

## 命令行选项

| 选项 | 描述 | 示例 |
|------|------|------|
| `-config` | 指定要验证的配置文件路径 | `-config config.yaml` |
| `-env-only` | 只验证环境变量，不验证配置文件 | `-env-only` |
| `-json` | 以 JSON 格式输出验证结果 | `-json` |
| `-verbose` | 显示详细的验证结果和建议 | `-verbose` |
| `-help` | 显示帮助信息 | `-help` |

## 使用示例

### 示例 1：验证基础配置文件

```bash
# 创建测试配置文件
cat > test-config.yaml << EOF
metrics:
  namespace: "winpower"
  subsystem: "exporter"
  request_duration_buckets: [0.05, 0.1, 0.2, 0.5, 1, 2, 5]
  collection_duration_buckets: [0.1, 0.2, 0.5, 1, 2, 5, 10]
  api_response_buckets: [0.05, 0.1, 0.2, 0.5, 1]
EOF

# 验证配置文件
./bin/validate-config -config test-config.yaml
```

**预期输出：**
```
🔍 Configuration File Validation:
✅ Configuration is valid
```

### 示例 2：验证有问题的配置

```bash
# 创建有问题的配置文件
cat > invalid-config.yaml << EOF
metrics:
  namespace: ""  # 空命名空间
  subsystem: "exporter"
  request_duration_buckets: [0.5, 0.1, 1, 2, 5]  # 非递增桶
  collection_duration_buckets: []  # 空桶数组
EOF

# 验证配置文件
./bin/validate-config -config invalid-config.yaml -verbose
```

**预期输出：**
```
🔍 Configuration File Validation:
❌ Configuration is invalid (3 errors, 0 warnings)

❌ Errors:
  • metrics namespace cannot be empty
  • request_duration buckets must be in increasing order: bucket[0] (0.500000) <= bucket[1] (0.100000)
  • collection_duration buckets cannot be empty
```

### 示例 3：验证环境变量

```bash
# 设置环境变量
export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter"
export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS="[0.05, 0.1, 0.2, 0.5, 1, 2, 5]"

# 验证环境变量
./bin/validate-config -env-only -verbose
```

**预期输出：**
```
🔍 Environment Variables Validation:
✅ Configuration is valid
```

### 示例 4：JSON 输出格式

```bash
# 验证配置并输出 JSON 格式
./bin/validate-config -config config.yaml -json
```

**预期输出：**
```json
{
  "overall_valid": true,
  "results": [
    {
      "type": "config_file",
      "file": "config.yaml",
      "valid": true,
      "errors": [],
      "warnings": []
    },
    {
      "type": "environment",
      "valid": true,
      "errors": [],
      "warnings": []
    }
  ]
}
```

## 配置建议

验证工具会根据配置内容生成优化建议：

### 示例：获取配置建议

```bash
# 创建需要优化的配置
cat > optimize-me.yaml << EOF
metrics:
  namespace: "winpower"
  subsystem: "exporter"
  request_duration_buckets: [1, 5, 10]  # 桶太少，第一个桶太大
  collection_duration_buckets: [0.5, 2, 8]  # 桶太少，最后一个桶太小
EOF

# 验证并获取建议
./bin/validate-config -config optimize-me.yaml -verbose
```

**预期输出：**
```
🔍 Configuration File Validation:
✅ Configuration is valid (with 2 warnings)

⚠️  Warnings:
  • request_duration has too few buckets (3), consider using at least 5 buckets for better granularity
  • collection_duration has too few buckets (3), consider using at least 5 buckets for better granularity

📋 Suggestions:
  • Consider adding more request_duration buckets for better granularity
  • Consider adding more collection_duration buckets for better granularity
  • Consider adding smaller request_duration buckets for fast requests
  • Consider adding larger collection_duration buckets for slow collections
```

## 验证规则

### 配置文件验证规则

1. **命名空间验证**
   - 不能为空
   - 只能包含字母、数字、下划线
   - 不能以数字开头
   - 建议长度不超过 50 字符

2. **子系统验证**
   - 不能为空
   - 只能包含字母、数字、下划线
   - 不能以数字开头
   - 建议长度不超过 50 字符

3. **直方图桶验证**
   - 桶数组不能为空
   - 桶边界必须递增
   - 所有桶值必须为正数
   - 建议桶数量在 5-20 之间
   - 根据桶类型检查边界的合理性

### 环境变量验证规则

1. **必需变量**
   - `WINPOWER_EXPORTER_METRICS_NAMESPACE`
   - `WINPOWER_EXPORTER_METRICS_SUBSYSTEM`

2. **可选变量**
   - `WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS`（JSON 数组格式）
   - `WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS`（JSON 数组格式）
   - `WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS`（JSON 数组格式）

## 集成到 CI/CD 流程

### GitHub Actions 示例

```yaml
# .github/workflows/validate-config.yml
name: Validate Configuration

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  validate-config:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21

    - name: Build validator
      run: |
        go build -o bin/validate-config ./internal/metrics/cmd/validate-config

    - name: Validate configuration files
      run: |
        # 验证所有配置文件
        for config in internal/metrics/examples/*.yaml; do
          echo "Validating $config"
          ./bin/validate-config -config "$config" -json
        done

    - name: Validate environment variables
      run: |
        # 设置测试环境变量
        export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
        export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter"

        # 验证环境变量
        ./bin/validate-config -env-only
```

### Makefile 集成

```makefile
# Makefile
.PHONY: validate-config validate-all-configs

# 构建验证工具
validate-config:
	go build -o bin/validate-config ./internal/metrics/cmd/validate-config

# 验证单个配置文件
validate-config-file:
	@if [ -z "$(CONFIG)" ]; then \
		echo "Usage: make validate-config-file CONFIG=<config-file>"; \
		exit 1; \
	fi
	./bin/validate-config -config $(CONFIG)

# 验证所有示例配置
validate-all-configs: validate-config
	@echo "Validating all example configurations..."
	@for config in internal/metrics/examples/*.yaml; do \
		echo "Validating $$config"; \
		./bin/validate-config -config "$$config" || exit 1; \
	done
	@echo "All configurations are valid!"

# 验证环境变量
validate-env: validate-config
	@echo "Validating environment variables..."
	./bin/validate-config -env-only
```

### Docker 集成

```dockerfile
# Dockerfile.validator
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o validate-config ./internal/metrics/cmd/validate-config

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/validate-config /usr/local/bin/
COPY internal/metrics/examples/ /configs/

ENTRYPOINT ["/usr/local/bin/validate-config"]
```

```bash
# 使用 Docker 验证配置
docker build -f Dockerfile.validator -t config-validator .
docker run --rm config-validator -config /configs/basic-config.yaml
```

## 故障排除

### 常见问题

1. **配置文件找不到**
   ```
   Error: reading file: open config.yaml: no such file or directory
   ```
   **解决方案**: 检查文件路径是否正确，确保文件存在

2. **无效的文件格式**
   ```
   Error: parsing YAML: yaml: line 5: mapping values are not allowed in this context
   ```
   **解决方案**: 检查 YAML 文件语法，确保缩进和格式正确

3. **JSON 桶格式错误**
   ```
   Error: invalid JSON format for request_duration buckets: invalid character ']' looking for beginning of value
   ```
   **解决方案**: 确保环境变量中的 JSON 数组格式正确

4. **缺少必需的环境变量**
   ```
   Error: required environment variable WINPOWER_EXPORTER_METRICS_NAMESPACE is not set
   ```
   **解决方案**: 设置所有必需的环境变量

### 调试技巧

1. **使用详细输出**
   ```bash
   ./bin/validate-config -config config.yaml -verbose
   ```

2. **使用 JSON 输出进行程序化处理**
   ```bash
   ./bin/validate-config -config config.yaml -json | jq '.overall_valid'
   ```

3. **验证特定部分**
   ```bash
   # 只验证环境变量
   ./bin/validate-config -env-only

   # 验证配置文件并跳过环境变量
   WINPOWER_EXPORTER_METRICS_NAMESPACE="" ./bin/validate-config -config config.yaml
   ```

## 扩展和定制

### 添加自定义验证规则

可以通过修改 `validator.go` 文件来添加自定义验证规则：

```go
// 添加自定义验证函数
func (v *Validator) validateCustomRule(cfg MetricManagerConfig, result *ValidationResult) {
    // 实现自定义验证逻辑
    if cfg.Namespace == "forbidden" {
        result.Errors = append(result.Errors, "namespace 'forbidden' is not allowed")
        result.Valid = false
    }
}
```

### 添加新的配置项支持

要支持新的配置项验证，需要：

1. 更新 `MetricManagerConfig` 结构
2. 在 `ValidateConfig` 方法中添加验证逻辑
3. 更新配置文件示例
4. 添加相应的测试用例

这份配置验证工具使用指南提供了完整的工具使用说明和最佳实践，帮助用户确保配置的正确性和优化性。