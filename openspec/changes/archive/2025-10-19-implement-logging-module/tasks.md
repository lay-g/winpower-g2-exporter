# 实现日志模块任务清单

## 任务概述

按照设计文档 `docs/design/logging.md` 实现完整的日志模块，包括核心接口、配置管理、多种输出支持、上下文日志和测试工具。

## 任务列表

### 1. 基础结构搭建
- [x] 创建 `internal/pkgs/log` 包目录结构
- [x] 创建基础文件：`logger.go`, `config.go`, `encoder.go`, `writer.go`, `context.go`
- [x] 添加 Go 依赖：`go.uber.org/zap`, `gopkg.in/natefinch/lumberjack.v2`
- [x] 创建包文档 `README.md`

### 2. 配置管理实现
- [x] 实现 `Config` 结构体（参考设计文档 4.1）
- [x] 实现 `DefaultConfig()` 函数
- [x] 实现 `DevelopmentDefaults()` 函数
- [x] 添加配置验证逻辑
- [x] 编写配置相关的单元测试

### 3. 核心日志接口实现
- [x] 定义 `Logger` 接口（参考设计文档 4.2）
- [x] 实现 `zapLoggerImpl` 结构体
- [x] 实现基础日志方法：`Debug`, `Info`, `Warn`, `Error`, `Fatal`
- [x] 实现 `With()` 方法创建子日志器
- [x] 实现字段构造函数：`String`, `Int`, `Float64`, `Bool`, `Error`, `Any`
- [x] 实现 `Sync()` 方法
- [x] 编写核心接口的单元测试

### 4. 编码器配置实现
- [x] 实现 `buildZapConfig()` 函数（参考设计文档 5.1）
- [x] 实现 JSON 编码器配置
- [x] 实现 Console 编码器配置
- [x] 添加时间格式化配置
- [x] 添加级别格式化配置
- [x] 编写编码器相关的单元测试

### 5. 输出管理实现
- [x] 实现 `NewRotatingFileWriter()` 函数（参考设计文档 5.3）
- [x] 实现标准输出/错误输出支持
- [x] 实现文件输出支持
- [x] 实现日志轮转功能（基于 lumberjack）
- [x] 实现多输出目标支持（stdout, stderr, file, both）
- [x] 编写输出管理相关的单元测试

### 6. 上下文日志支持实现
- [x] 定义上下文字段键常量（参考设计文档 4.3）
- [x] 实现上下文工具函数：`WithRequestID`, `WithTraceID`, `WithUserID` 等
- [x] 实现 `WithContext()` 方法
- [x] 实现 `extractContextFields()` 函数
- [x] 实现上下文日志器管理：`WithLogger`, `FromLogger`, `LoggerFromContext`
- [x] 编写上下文日志相关的单元测试

### 7. 全局日志器实现
- [x] 实现全局日志器变量和初始化控制（参考设计文档 6.1）
- [x] 实现 `Init()` 函数
- [x] 实现 `InitDevelopment()` 函数
- [x] 实现 `Default()` 函数
- [x] 实现 `ResetGlobal()` 函数
- [x] 实现全局日志函数：`Debug`, `Info`, `Warn`, `Error`, `Fatal`
- [x] 实现全局上下文日志函数：`DebugWithContext`, `InfoWithContext` 等
- [x] 实现全局 `With()` 和 `WithContext()` 函数
- [x] 编写全局日志器相关的单元测试

### 8. 测试专用工具实现
- [x] 实现 `TestLogger` 结构体（参考设计文档 6.2.1）
- [x] 实现 `LogEntry` 结构体
- [x] 实现 `NewTestLogger()` 和 `NewTestLoggerWithT()` 函数
- [x] 实现日志条目查询方法：`Entries()`, `EntriesByLevel()`, `EntriesByMessage()` 等
- [x] 实现 `LogCapture` 结构体（参考设计文档 6.2.2）
- [x] 实现 `NewNoopLogger()` 函数
- [x] 编写测试工具相关的单元测试

### 9. 日志器初始化实现
- [x] 实现 `NewLogger()` 函数（参考设计文档 5.1）
- [x] 实现 `BuildLoggerWithRotation()` 函数（参考设计文档 5.3）
- [x] 添加调用者信息配置
- [x] 添加堆栈跟踪配置
- [x] 处理初始化错误
- [x] 编写初始化相关的单元测试

### 10. 性能优化和最佳实践
- [x] 实现日志采样功能（可选）
- [x] 添加敏感信息脱敏逻辑
- [x] 实现条件日志优化
- [x] 添加性能基准测试
- [x] 验证内存分配情况

### 11. 集成测试
- [x] 编写端到端集成测试
- [x] 测试不同配置组合
- [x] 测试文件输出和轮转
- [x] 测试上下文日志功能
- [x] 测试全局日志器功能
- [x] 测试错误处理和恢复

### 12. 文档和示例
- [x] 完善 `README.md` 文档
- [x] 添加基本使用示例
- [x] 添加高级配置示例
- [x] 添加测试使用示例
- [x] 添加故障排查指南
- [x] 更新项目整体文档中的日志模块说明

## 验证标准

每个任务完成后需要满足以下标准：
1. **功能正确性**：代码实现符合设计文档要求
2. **测试覆盖**：单元测试覆盖率达到 90% 以上
3. **代码质量**：通过 `go fmt` 和 `make lint` 检查
4. **性能要求**：满足高性能要求，避免不必要的内存分配
5. **文档完整**：关键函数和结构体有完整的注释

## 依赖关系

任务执行顺序：
1. 任务 1-3：基础结构和核心接口
2. 任务 4-5：编码器和输出管理
3. 任务 6-7：上下文支持和全局日志器
4. 任务 8-9：测试工具和初始化
5. 任务 10-12：优化、集成测试和文档

每个阶段的任务完成后需要进行集成测试，确保与之前已完成的功能正常协作。