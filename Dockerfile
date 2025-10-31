# 构建阶段
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git ca-certificates tzdata make

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 使用 Makefile 构建（包含版本信息注入）
RUN make build-linux && \
    mv build/linux-amd64/winpower-g2-exporter winpower-g2-exporter

# 运行阶段
FROM alpine:latest

# 安装必要工具
RUN apk --no-cache add ca-certificates tzdata wget

# 创建非 root 用户
RUN addgroup -g 1001 -S winpower && \
    adduser -u 1001 -S winpower -G winpower

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/winpower-g2-exporter .

# 创建配置和数据目录
RUN mkdir -p /app/config /app/data

# 复制配置示例文件
COPY --from=builder /app/config/config.example.yaml /app/config/config.example.yaml

# 设置目录权限
RUN chown -R winpower:winpower /app

# 切换到非 root 用户
USER winpower

# 暴露端口
EXPOSE 9090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/health || exit 1

# 启动命令
CMD ["./winpower-g2-exporter", "serve"]