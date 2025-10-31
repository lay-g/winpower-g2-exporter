# Server模块设计文档

## 架构决策

### 1. 框架选择：Gin
**决策**：选择Gin作为HTTP框架
**理由**：
- 高性能的路由引擎
- 丰富的中间件生态
- 社区活跃，文档完善
- 符合项目的技术栈要求

### 2. 配置管理策略
**决策**：模块内定义配置，外部注册集成
**理由**：
- 保持模块边界清晰
- 配置验证逻辑内聚
- 支持不同配置源的统一集成
- 便于测试和模块替换

### 3. 依赖注入模式
**决策**：通过构造函数注入依赖接口
**理由**：
- 明确依赖关系，提高可测试性
- 支持接口隔离和模块替换
- 符合Go的依赖管理最佳实践

## 接口设计

### Server接口
```go
type Server interface {
    Start() error
    Stop(ctx context.Context) error
}
```

**设计考量**：
- 接口最小化，只包含核心生命周期方法
- Stop方法接受context，支持超时控制
- 避免暴露内部实现细节

### MetricsService接口
```go
type MetricsService interface {
    HandleMetrics(c *gin.Context)
}
```

**设计考量**：
- 直接使用metrics模块提供的Gin handler
- metrics模块负责数据采集和Prometheus格式返回
- server模块只负责路由注册，不处理业务逻辑

### HealthService接口
```go
type HealthService interface {
    Check(ctx context.Context) (status string, details map[string]any)
}
```

**设计考量**：
- 支持详细健康检查信息
- 状态字符串标准化（ok/unhealthy等）
- 便于扩展健康检查逻辑

## 中间件设计

### Logger中间件
```go
func (s *HTTPServer) loggerMiddleware() gin.HandlerFunc {
    return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
        // 结构化日志输出
        return ""
    })
}
```

**设计考量**：
- 记录请求方法、路径、状态码、耗时
- 避免记录敏感信息
- 统一日志格式

### Recovery中间件
```go
func (s *HTTPServer) recoveryMiddleware() gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
        // 统一错误处理和日志记录
    })
}
```

**设计考量**：
- 捕获panic，防止服务崩溃
- 返回标准化的错误响应
- 记录详细的错误日志

## 错误处理策略

### HTTP错误响应
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Path    string `json:"path"`
    Time    string `json:"ts"`
}
```

**设计考量**：
- 统一错误响应格式
- 提供必要的调试信息
- 避免泄露内部实现细节

### 错误日志级别
- **ERROR**：服务启动失败、端口冲突等
- **WARN**：请求超时、无效路径等
- **INFO**：服务启停、请求处理等
- **DEBUG**：详细请求信息（仅开发模式）

## 路由设计

### 核心路由
- `GET /health` - 健康检查
- `GET /metrics` - Prometheus指标导出（直接使用metrics模块的handler）

### 调试路由（可选）
- `GET /debug/pprof/*` - 性能分析

### 错误处理路由
- `*/*` - 404错误处理

**设计考量**：
- 路由命名标准化
- 支持未来功能扩展
- 统一错误响应格式
- metrics端点直接使用业务模块提供的handler，避免重复实现

### 路由实现
```go
func (s *HTTPServer) setupRoutes() {
    // 健康检查路由
    s.engine.GET("/health", s.health.Check)

    // 指标导出路由 - 直接使用metrics模块的handler
    s.engine.GET("/metrics", s.metrics.HandleMetrics)

    // 404 处理器
    s.engine.NoRoute(s.handleNotFound)

    // /debug/pprof（可选）
    if s.cfg.EnablePprof {
        s.setupPprofRoutes()
    }
}
```

## 实现结构

### HTTPServer结构体
```go
type HTTPServer struct {
    cfg     *Config
    log     Logger
    engine  *gin.Engine
    srv     *http.Server
    metrics interface{ HandleMetrics(c *gin.Context) }
    health  HealthService
}
```

### 构造函数
```go
func NewHTTPServer(
    config *Config,
    log Logger,
    metrics interface{ HandleMetrics(c *gin.Context) },
    health HealthService,
) *HTTPServer {
    // 设置Gin模式
    gin.SetMode(config.Mode)

    engine := gin.New()

    server := &HTTPServer{
        cfg:     config,
        log:     log,
        engine:  engine,
        metrics: metrics,
        health:  health,
    }

    // 设置中间件
    server.setupGlobalMiddleware()

    // 注册路由
    server.setupRoutes()

    // 创建http.Server
    server.srv = &http.Server{
        Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
        Handler:      engine,
        ReadTimeout:  config.ReadTimeout,
        WriteTimeout: config.WriteTimeout,
        IdleTimeout:  config.IdleTimeout,
    }

    return server
}
```

## 优雅关闭机制

### 关闭流程
1. 接收关闭信号
2. 停止接受新连接
3. 等待现有连接处理完成
4. 释放资源
5. 记录关闭日志

### 超时控制
- 默认关闭超时：30秒
- 支持通过context控制
- 强制关闭超时：5秒

## 安全考虑

### 输入验证
- 路径参数验证
- 查询参数验证
- 请求头验证

### 输出过滤
- 敏感信息过滤
- 错误信息脱敏
- 日志输出控制

### 生产建议
- 通过反向代理终结TLS
- 添加认证中间件
- 限制请求频率

## 性能优化

### 中间件顺序
1. Logger（最外层）
2. Recovery（次外层）
3. 业务逻辑（内层）

### 连接管理
- 合理设置超时参数
- 启用Keep-Alive
- 连接池管理

### 内存优化
- 避免内存泄漏
- 及时释放资源
- 监控内存使用

## 可扩展性

### 中间件扩展
```go
func (s *HTTPServer) Use(middleware ...gin.HandlerFunc) {
    s.engine.Use(middleware...)
}
```

### 路由扩展
```go
func (s *HTTPServer) RegisterRoutes(group *gin.RouterGroup) {
    // 自定义路由注册
}
```

### 服务扩展
- 支持插件化架构
- 配置驱动的功能开关
- 模块化的服务组合

## 测试策略

### 单元测试
- 接口实现测试
- 中间件测试
- 错误处理测试

### 集成测试
- 端到端请求测试
- 优雅关闭测试
- 并发请求测试

### 性能测试
- 负载测试
- 压力测试
- 内存泄漏检测