# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装依赖
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用和工具
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o winpower-g2-exporter ./cmd/exporter && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o winpower-config-migrate ./cmd/config-migrate

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1001 -S winpower && \
    adduser -u 1001 -S winpower -G winpower

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/winpower-g2-exporter .
COPY --from=builder /app/winpower-config-migrate .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建数据目录
RUN mkdir -p /var/lib/winpower-exporter && \
    chown -R winpower:winpower /app /var/lib/winpower-exporter

# 切换到非 root 用户
USER winpower

# 暴露端口
EXPOSE 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# 启动命令
CMD ["./winpower-g2-exporter", "--config", "configs/config.yaml"]