# 环境变量配置示例

本文档提供了使用环境变量配置 WinPower G2 Exporter Metrics 模块的详细示例和最佳实践。

## 环境变量命名规范

WinPower G2 Exporter 使用 `WINPOWER_EXPORTER_` 作为前缀，Metrics 相关的配置使用 `WINPOWER_EXPORTER_METRICS_` 前缀。

### 命名转换规则

1. **配置文件路径** → **环境变量名**
   - `metrics.namespace` → `WINPOWER_EXPORTER_METRICS_NAMESPACE`
   - `metrics.subsystem` → `WINPOWER_EXPORTER_METRICS_SUBSYSTEM`

2. **数组配置** → **JSON 格式字符串**
   - `metrics.request_duration_buckets` → `WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS`

## 基础环境变量配置

### 最小配置

```bash
# 仅设置必需的环境变量
export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter"

# 启动应用
./winpower-exporter
```

### 开发环境配置

```bash
# 基础配置
export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter-dev"

# WinPower 连接配置
export WINPOWER_EXPORTER_WINPOWER_URL="http://localhost:8080"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="dev"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="dev123"
export WINPOWER_EXPORTER_WINPOWER_SKIP_SSL_VERIFY="true"

# 服务器配置
export WINPOWER_EXPORTER_SERVER_PORT="9091"
export WINPOWER_EXPORTER_SERVER_HOST="127.0.0.1"

# 日志配置
export WINPOWER_EXPORTER_LOGGING_LEVEL="debug"
export WINPOWER_EXPORTER_LOGGING_FORMAT="console"

# 启动应用
./winpower-exporter
```

### 生产环境配置

```bash
# 基础配置
export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter"

# WinPower 连接配置
export WINPOWER_EXPORTER_WINPOWER_URL="https://winpower-prod.company.com:8443"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="exporter"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="${WINPOWER_PASSWORD}"  # 从现有环境变量读取
export WINPOWER_EXPORTER_WINPOWER_TIMEOUT="45s"
export WINPOWER_EXPORTER_WINPOWER_SKIP_SSL_VERIFY="false"

# 服务器配置
export WINPOWER_EXPORTER_SERVER_PORT="9090"
export WINPOWER_EXPORTER_SERVER_HOST="0.0.0.0"
export WINPOWER_EXPORTER_SERVER_READ_TIMEOUT="60s"
export WINPOWER_EXPORTER_SERVER_WRITE_TIMEOUT="60s"

# 日志配置
export WINPOWER_EXPORTER_LOGGING_LEVEL="info"
export WINPOWER_EXPORTER_LOGGING_FORMAT="json"
export WINPOWER_EXPORTER_LOGGING_OUTPUT="/var/log/winpower-exporter/app.log"

# 存储配置
export WINPOWER_EXPORTER_STORAGE_DATA_DIR="/var/lib/winpower-exporter"
export WINPOWER_EXPORTER_STORAGE_SYNC_WRITE="true"

# 调度配置
export WINPOWER_EXPORTER_SCHEDULER_INTERVAL="5s"

# 启动应用
./winpower-exporter
```

## 直方图桶配置

### 设置桶边界（JSON 格式）

```bash
# HTTP 请求时延桶（单位：秒）
export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS='[0.05, 0.1, 0.2, 0.5, 1, 2, 5]'

# 采集时延桶（单位：秒）
export WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS='[0.1, 0.2, 0.5, 1, 2, 5, 10]'

# API 响应时延桶（单位：秒）
export WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS='[0.05, 0.1, 0.2, 0.5, 1]'

# 启动应用
./winpower-exporter
```

### 不同场景的桶配置

#### 快速响应环境
```bash
export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS='[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1]'
export WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS='[0.05, 0.1, 0.25, 0.5, 1, 2.5, 5]'
export WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS='[0.01, 0.025, 0.05, 0.1, 0.25, 0.5]'
```

#### 慢速网络环境
```bash
export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS='[0.5, 1, 2.5, 5, 10, 25, 60]'
export WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS='[1, 2.5, 5, 15, 30, 60, 120]'
export WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS='[0.5, 1, 2.5, 5, 10, 25]'
```

#### 高吞吐量环境
```bash
export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS='[0.25, 0.5, 1, 2.5, 5, 10, 25, 60]'
export WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS='[1, 2.5, 5, 15, 30, 60, 120, 300]'
export WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS='[0.5, 1, 2.5, 5, 15, 30]'
```

## Docker 环境配置

### Docker Compose 示例

```yaml
# docker-compose.yml
version: '3.8'

services:
  winpower-exporter:
    image: winpower-exporter:latest
    container_name: winpower-exporter
    ports:
      - "9090:9090"
    environment:
      # 基础配置
      - WINPOWER_EXPORTER_METRICS_NAMESPACE=winpower
      - WINPOWER_EXPORTER_METRICS_SUBSYSTEM=exporter

      # WinPower 连接配置
      - WINPOWER_EXPORTER_WINPOWER_URL=http://winpower-server:8080
      - WINPOWER_EXPORTER_WINPOWER_USERNAME=admin
      - WINPOWER_EXPORTER_WINPOWER_PASSWORD=${WINPOWER_PASSWORD}
      - WINPOWER_EXPORTER_WINPOWER_TIMEOUT=30s

      # 服务器配置
      - WINPOWER_EXPORTER_SERVER_PORT=9090
      - WINPOWER_EXPORTER_SERVER_HOST=0.0.0.0

      # 日志配置
      - WINPOWER_EXPORTER_LOGGING_LEVEL=info
      - WINPOWER_EXPORTER_LOGGING_FORMAT=json

      # 存储配置
      - WINPOWER_EXPORTER_STORAGE_DATA_DIR=/app/data
      - WINPOWER_EXPORTER_STORAGE_SYNC_WRITE=true

      # 调度配置
      - WINPOWER_EXPORTER_SCHEDULER_INTERVAL=5s

      # 直方图桶配置
      - WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS=[0.05, 0.1, 0.2, 0.5, 1, 2, 5]
      - WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS=[0.1, 0.2, 0.5, 1, 2, 5, 10]
      - WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS=[0.05, 0.1, 0.2, 0.5, 1]

    volumes:
      - ./data:/app/data
      - ./logs:/app/logs

    restart: unless-stopped
```

### Dockerfile 示例

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o winpower-exporter ./cmd/exporter

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# 创建非 root 用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

COPY --from=builder /app/winpower-exporter .
COPY --from=builder /app/internal/metrics/examples/basic-config.yaml ./config.yaml

# 设置默认环境变量
ENV WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
ENV WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter"
ENV WINPOWER_EXPORTER_SERVER_PORT="9090"
ENV WINPOWER_EXPORTER_LOGGING_LEVEL="info"

# 创建必要的目录
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app

USER appuser

EXPOSE 9090

CMD ["./winpower-exporter"]
```

### 运行 Docker 容器

```bash
# 构建镜像
docker build -t winpower-exporter:latest .

# 运行容器（使用环境变量）
docker run -d \
  --name winpower-exporter \
  -p 9090:9090 \
  -e WINPOWER_EXPORTER_METRICS_NAMESPACE="myapp" \
  -e WINPOWER_EXPORTER_METRICS_SUBSYSTEM="monitor" \
  -e WINPOWER_EXPORTER_WINPOWER_URL="http://winpower-server:8080" \
  -e WINPOWER_EXPORTER_WINPOWER_USERNAME="admin" \
  -e WINPOWER_EXPORTER_WINPOWER_PASSWORD="password" \
  -v $(pwd)/data:/app/data \
  winpower-exporter:latest

# 查看容器日志
docker logs -f winpower-exporter
```

## Kubernetes 环境配置

### ConfigMap 示例

```yaml
# k8s-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: winpower-exporter-config
  namespace: monitoring
data:
  # 基础配置
  WINPOWER_EXPORTER_METRICS_NAMESPACE: "winpower"
  WINPOWER_EXPORTER_METRICS_SUBSYSTEM: "exporter"

  # WinPower 连接配置
  WINPOWER_EXPORTER_WINPOWER_URL: "http://winpower-service.monitoring.svc.cluster.local:8080"
  WINPOWER_EXPORTER_WINPOWER_USERNAME: "exporter"

  # 服务器配置
  WINPOWER_EXPORTER_SERVER_PORT: "9090"
  WINPOWER_EXPORTER_SERVER_HOST: "0.0.0.0"

  # 日志配置
  WINPOWER_EXPORTER_LOGGING_LEVEL: "info"
  WINPOWER_EXPORTER_LOGGING_FORMAT: "json"

  # 存储配置
  WINPOWER_EXPORTER_STORAGE_DATA_DIR: "/app/data"
  WINPOWER_EXPORTER_STORAGE_SYNC_WRITE: "true"

  # 调度配置
  WINPOWER_EXPORTER_SCHEDULER_INTERVAL: "5s"

  # 直方图桶配置
  WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS: "[0.05, 0.1, 0.2, 0.5, 1, 2, 5]"
  WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS: "[0.1, 0.2, 0.5, 1, 2, 5, 10]"
  WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS: "[0.05, 0.1, 0.2, 0.5, 1]"
```

### Secret 示例

```yaml
# k8s-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: winpower-exporter-secrets
  namespace: monitoring
type: Opaque
data:
  # Base64 编码的密码
  WINPOWER_EXPORTER_WINPOWER_PASSWORD: "cGFzc3dvcmQxMjM="  # "password123"
```

### Deployment 示例

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: winpower-exporter
  namespace: monitoring
  labels:
    app: winpower-exporter
spec:
  replicas: 2
  selector:
    matchLabels:
      app: winpower-exporter
  template:
    metadata:
      labels:
        app: winpower-exporter
    spec:
      containers:
      - name: winpower-exporter
        image: winpower-exporter:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 9090
          name: metrics
        envFrom:
        - configMapRef:
            name: winpower-exporter-config
        - secretRef:
            name: winpower-exporter-secrets
        volumeMounts:
        - name: data-volume
          mountPath: /app/data
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
        livenessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 9090
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: winpower-exporter-data
---
apiVersion: v1
kind: Service
metadata:
  name: winpower-exporter-service
  namespace: monitoring
  labels:
    app: winpower-exporter
spec:
  selector:
    app: winpower-exporter
  ports:
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

## 环境配置脚本

### 开发环境启动脚本

```bash
#!/bin/bash
# dev-start.sh - 开发环境启动脚本

echo "Starting WinPower Exporter in development mode..."

# 设置开发环境变量
export WINPOWER_EXPORTER_METRICS_NAMESPACE="winpower"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="exporter-dev"

export WINPOWER_EXPORTER_WINPOWER_URL="http://localhost:8080"
export WINPOWER_EXPORTER_WINPOWER_USERNAME="dev"
export WINPOWER_EXPORTER_WINPOWER_PASSWORD="dev123"
export WINPOWER_EXPORTER_WINPOWER_SKIP_SSL_VERIFY="true"

export WINPOWER_EXPORTER_SERVER_PORT="9091"
export WINPOWER_EXPORTER_SERVER_HOST="127.0.0.1"

export WINPOWER_EXPORTER_LOGGING_LEVEL="debug"
export WINPOWER_EXPORTER_LOGGING_FORMAT="console"

export WINPOWER_EXPORTER_STORAGE_DATA_DIR="./dev-data"
export WINPOWER_EXPORTER_STORAGE_SYNC_WRITE="false"

export WINPOWER_EXPORTER_SCHEDULER_INTERVAL="2s"

# 开发环境的桶配置
export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS='[0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1]'
export WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS='[0.05, 0.1, 0.25, 0.5, 1, 2.5, 5]'
export WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS='[0.01, 0.025, 0.05, 0.1, 0.25, 0.5]'

# 创建必要的目录
mkdir -p ./dev-data ./logs

# 启动应用
echo "Configuration loaded. Starting exporter..."
./winpower-exporter
```

### 生产环境启动脚本

```bash
#!/bin/bash
# prod-start.sh - 生产环境启动脚本

echo "Starting WinPower Exporter in production mode..."

# 检查必需的环境变量
required_vars=(
    "WINPOWER_EXPORTER_WINPOWER_URL"
    "WINPOWER_EXPORTER_WINPOWER_USERNAME"
    "WINPOWER_EXPORTER_WINPOWER_PASSWORD"
)

for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "Error: Required environment variable $var is not set"
        exit 1
    fi
done

# 设置生产环境默认值
export WINPOWER_EXPORTER_METRICS_NAMESPACE="${WINPOWER_EXPORTER_METRICS_NAMESPACE:-winpower}"
export WINPOWER_EXPORTER_METRICS_SUBSYSTEM="${WINPOWER_EXPORTER_METRICS_SUBSYSTEM:-exporter}"

export WINPOWER_EXPORTER_SERVER_PORT="${WINPOWER_EXPORTER_SERVER_PORT:-9090}"
export WINPOWER_EXPORTER_SERVER_HOST="${WINPOWER_EXPORTER_SERVER_HOST:-0.0.0.0}"

export WINPOWER_EXPORTER_LOGGING_LEVEL="${WINPOWER_EXPORTER_LOGGING_LEVEL:-info}"
export WINPOWER_EXPORTER_LOGGING_FORMAT="${WINPOWER_EXPORTER_LOGGING_FORMAT:-json}"

export WINPOWER_EXPORTER_STORAGE_DATA_DIR="${WINPOWER_EXPORTER_STORAGE_DATA_DIR:-/var/lib/winpower-exporter}"
export WINPOWER_EXPORTER_STORAGE_SYNC_WRITE="${WINPOWER_EXPORTER_STORAGE_SYNC_WRITE:-true}"

export WINPOWER_EXPORTER_SCHEDULER_INTERVAL="${WINPOWER_EXPORTER_SCHEDULER_INTERVAL:-5s}"

# 生产环境的桶配置（如果未设置）
if [ -z "$WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS" ]; then
    export WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS='[0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30]'
fi

if [ -z "$WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS" ]; then
    export WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS='[0.5, 1, 2.5, 5, 10, 30, 60, 120]'
fi

if [ -z "$WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS" ]; then
    export WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS='[0.1, 0.25, 0.5, 1, 2.5, 5, 10]'
fi

# 创建必要的目录
mkdir -p "$WINPOWER_EXPORTER_STORAGE_DATA_DIR"

# 显示配置摘要
echo "=== Production Configuration ==="
echo "Namespace: $WINPOWER_EXPORTER_METRICS_NAMESPACE"
echo "Subsystem: $WINPOWER_EXPORTER_METRICS_SUBSYSTEM"
echo "WinPower URL: $WINPOWER_EXPORTER_WINPOWER_URL"
echo "Server Port: $WINPOWER_EXPORTER_SERVER_PORT"
echo "Data Directory: $WINPOWER_EXPORTER_STORAGE_DATA_DIR"
echo "================================"

# 启动应用
echo "Starting exporter..."
./winpower-exporter
```

## 环境变量验证

### 验证脚本

```bash
#!/bin/bash
# validate-env.sh - 环境变量验证脚本

echo "Validating WinPower Exporter environment variables..."

# 验证必需的基础配置
validate_required() {
    local var_name="$1"
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo "❌ Required environment variable $var_name is not set"
        return 1
    else
        echo "✅ $var_name: $var_value"
        return 0
    fi
}

# 验证 JSON 数组格式
validate_json_array() {
    local var_name="$1"
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo "⚠️  Optional environment variable $var_name is not set"
        return 0
    fi

    # 简单的 JSON 数组格式验证
    if [[ "$var_value" =~ ^\[.*\]$ ]]; then
        echo "✅ $var_name: $var_value"
        return 0
    else
        echo "❌ $var_name is not a valid JSON array format: $var_value"
        return 1
    fi
}

# 验证端口号
validate_port() {
    local var_name="$1"
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo "⚠️  Optional environment variable $var_name is not set"
        return 0
    fi

    if [[ "$var_value" =~ ^[0-9]+$ ]] && [ "$var_value" -ge 1 ] && [ "$var_value" -le 65535 ]; then
        echo "✅ $var_name: $var_value"
        return 0
    else
        echo "❌ $var_name is not a valid port number: $var_value"
        return 1
    fi
}

# 验证 URL 格式
validate_url() {
    local var_name="$1"
    local var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        echo "⚠️  Optional environment variable $var_name is not set"
        return 0
    fi

    if [[ "$var_value" =~ ^https?:// ]]; then
        echo "✅ $var_name: $var_value"
        return 0
    else
        echo "❌ $var_name is not a valid URL: $var_value"
        return 1
    fi
}

echo "=== Required Variables ==="
validate_required "WINPOWER_EXPORTER_METRICS_NAMESPACE"
validate_required "WINPOWER_EXPORTER_METRICS_SUBSYSTEM"

echo ""
echo "=== WinPower Connection ==="
validate_required "WINPOWER_EXPORTER_WINPOWER_URL"
validate_required "WINPOWER_EXPORTER_WINPOWER_USERNAME"
validate_required "WINPOWER_EXPORTER_WINPOWER_PASSWORD"
validate_url "WINPOWER_EXPORTER_WINPOWER_URL"

echo ""
echo "=== Server Configuration ==="
validate_port "WINPOWER_EXPORTER_SERVER_PORT"
validate_port "WINPOWER_EXPORTER_SERVER_HOST"

echo ""
echo "=== Metrics Buckets ==="
validate_json_array "WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS"
validate_json_array "WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS"
validate_json_array "WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS"

echo ""
echo "=== Storage Configuration ==="
validate_required "WINPOWER_EXPORTER_STORAGE_DATA_DIR"

echo ""
echo "Validation completed."
```

### 使用验证脚本

```bash
# 给脚本执行权限
chmod +x validate-env.sh

# 运行验证
./validate-env.sh

# 预期输出示例：
# Validating WinPower Exporter environment variables...
# === Required Variables ===
# ✅ WINPOWER_EXPORTER_METRICS_NAMESPACE: winpower
# ✅ WINPOWER_EXPORTER_METRICS_SUBSYSTEM: exporter
#
# === WinPower Connection ===
# ✅ WINPOWER_EXPORTER_WINPOWER_URL: http://localhost:8080
# ✅ WINPOWER_EXPORTER_WINPOWER_USERNAME: admin
# ✅ WINPOWER_EXPORTER_WINPOWER_PASSWORD: password
# ✅ WINPOWER_EXPORTER_WINPOWER_URL: http://localhost:8080
#
# === Server Configuration ===
# ✅ WINPOWER_EXPORTER_SERVER_PORT: 9090
# ⚠️  Optional environment variable WINPOWER_EXPORTER_SERVER_HOST is not set
#
# === Metrics Buckets ===
# ✅ WINPOWER_EXPORTER_METRICS_REQUEST_DURATION_BUCKETS: [0.05, 0.1, 0.2, 0.5, 1, 2, 5]
# ✅ WINPOWER_EXPORTER_METRICS_COLLECTION_DURATION_BUCKETS: [0.1, 0.2, 0.5, 1, 2, 5, 10]
# ✅ WINPOWER_EXPORTER_METRICS_API_RESPONSE_BUCKETS: [0.05, 0.1, 0.2, 0.5, 1]
#
# === Storage Configuration ===
# ✅ WINPOWER_EXPORTER_STORAGE_DATA_DIR: ./data
#
# Validation completed.
```

## 最佳实践

### 1. 安全性

- 使用 Secret 管理敏感信息（密码、API 密钥等）
- 避免在命令行中传递密码
- 定期轮换访问凭据

### 2. 可维护性

- 使用环境配置文件管理不同环境的配置
- 为每个环境创建独立的配置脚本
- 使用版本控制管理配置变更

### 3. 监控和调试

- 在生产环境中记录配置变更
- 使用验证脚本确保配置正确性
- 监控环境变量对性能的影响

### 4. 部署一致性

- 使用相同的配置模板部署到不同环境
- 通过环境变量覆盖特定参数
- 定期同步配置模板