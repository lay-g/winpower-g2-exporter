# Docker Compose 部署指南

本目录包含使用 Docker Compose 部署 WinPower G2 Exporter 的配置文件。

## 快速开始

### 1. 准备配置

```bash
# 复制环境变量配置文件
cp .env.example .env

# 编辑 .env 文件，配置 WinPower 连接信息
vim .env
```

### 2. 启动服务

```bash
# 仅启动 exporter（推荐用于生产环境）
docker-compose up -d winpower-exporter

# 或启动完整监控栈（包括 Prometheus 和 Grafana）
docker-compose up -d
```

### 3. 验证服务

```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f winpower-exporter

# 检查健康状态
curl http://localhost:9090/health

# 查看指标
curl http://localhost:9090/metrics
```

### 4. 访问服务

- **Exporter 指标**: http://localhost:9090/metrics
- **Exporter 健康检查**: http://localhost:9090/health
- **Prometheus**: http://localhost:9091 (如果启用)
- **Grafana**: http://localhost:3000 (如果启用，默认账号 admin/admin)

## 配置说明

### 环境变量配置

主要配置项（在 `.env` 文件中设置）：

| 变量                       | 说明              | 默认值                       |
| -------------------------- | ----------------- | ---------------------------- |
| `WINPOWER_BASE_URL`        | WinPower 服务地址 | https://winpower.example.com |
| `WINPOWER_USERNAME`        | 登录用户名        | admin                        |
| `WINPOWER_PASSWORD`        | 登录密码          | password                     |
| `WINPOWER_SKIP_SSL_VERIFY` | 跳过 SSL 验证     | false                        |
| `LOG_LEVEL`                | 日志级别          | info                         |
| `ENABLE_PPROF`             | 启用性能分析      | false                        |

完整配置项请参考 `config/config.example.yaml`。

### 使用配置文件

如果不想使用环境变量，可以挂载配置文件：

1. 取消 `docker-compose.yml` 中的配置文件挂载注释
2. 创建配置文件：

```bash
cp config/config.example.yaml config/config.yaml
vim config/config.yaml
```

### 数据持久化

累计电能数据存储在 `./data` 目录，会自动挂载到容器中，确保数据持久化。

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart winpower-exporter

# 查看日志
docker-compose logs -f winpower-exporter

# 进入容器
docker-compose exec winpower-exporter sh

# 重新构建镜像
docker-compose build --no-cache

# 清理所有资源（包括数据卷）
docker-compose down -v
```

## 生产环境建议

### 安全配置

1. **使用 .env 文件管理敏感信息**，不要将密码硬编码到 docker-compose.yml
2. **设置合适的文件权限**：
   ```bash
   chmod 600 .env
   chmod 755 data/
   ```
3. **在反向代理前终止 TLS**，避免在内网传输中暴露明文

### 性能优化

1. **调整资源限制**：根据实际负载修改 `deploy.resources`
2. **配置日志轮转**：已配置日志大小限制（10MB × 3 个文件）
3. **监控容器健康**：使用 `docker-compose ps` 或监控工具检查健康状态

### 监控配置

1. **Prometheus 抓取间隔**：默认 15 秒，与 exporter 采集周期（5秒）协调
2. **数据保留时间**：默认 30 天，可通过修改 `--storage.tsdb.retention.time` 调整
3. **Grafana 仪表盘**：可在 `deployments/grafana/dashboards/` 添加自定义仪表盘

## 故障排查

### 服务无法启动

```bash
# 查看详细日志
docker-compose logs winpower-exporter

# 检查配置
docker-compose config

# 验证环境变量
docker-compose exec winpower-exporter env | grep WINPOWER
```

### 无法连接 WinPower

1. 检查 `WINPOWER_BASE_URL` 是否正确
2. 验证网络连通性：
   ```bash
   docker-compose exec winpower-exporter wget --spider $WINPOWER_BASE_URL
   ```
3. 如果使用自签证书，设置 `WINPOWER_SKIP_SSL_VERIFY=true`

### 指标数据异常

1. 检查采集日志：
   ```bash
   docker-compose logs -f winpower-exporter | grep -i error
   ```
2. 验证 WinPower API 响应：
   ```bash
   # 在容器内测试 API
   docker-compose exec winpower-exporter sh
   wget -O- "$WINPOWER_BASE_URL/api/data" || true
   ```

### 数据持久化问题

```bash
# 检查数据目录权限
ls -la ./data/

# 查看容器内数据
docker-compose exec winpower-exporter ls -la /app/data/

# 修复权限（如果需要）
sudo chown -R 1001:1001 ./data/
```

## 更新与维护

### 更新镜像

```bash
# 拉取最新代码
git pull

# 重新构建镜像
docker-compose build --no-cache winpower-exporter

# 重启服务
docker-compose up -d winpower-exporter
```

### 备份数据

```bash
# 备份累计电能数据
tar -czf winpower-data-backup-$(date +%Y%m%d).tar.gz data/

# 备份配置
tar -czf winpower-config-backup-$(date +%Y%m%d).tar.gz .env config/
```

### 查看版本信息

```bash
docker-compose exec winpower-exporter ./winpower-g2-exporter --version
```

## 参考资料

- [项目文档](../README.md)
- [配置说明](../CONFIGURATION.md)
- [设计文档](../docs/design/)
- [Prometheus 文档](https://prometheus.io/docs/)
- [Grafana 文档](https://grafana.com/docs/)
