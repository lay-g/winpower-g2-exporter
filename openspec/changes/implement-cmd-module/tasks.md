# 任务列表

## 1. 项目结构设置
- [x] 创建 `cmd/winpower-g2-exporter/` 目录结构
- [x] 创建 `internal/cmd/` 目录用于共享的命令行工具实现
- [x] 添加必要的依赖到 go.mod（cobra, viper）
- [x] 更新 VERSION 文件（如果不存在）

## 2. 根命令实现
- [x] 实现 `cmd/winpower-g2-exporter/main.go` 主程序入口
- [x] 实现 `cmd/winpower-g2-exporter/root.go` 根命令结构
- [x] 添加持久化参数（--config, --verbose）
- [x] 实现配置绑定逻辑
- [x] 实现根命令的 Execute 方法

## 3. Server 子命令实现
- [x] 实现 `cmd/winpower-g2-exporter/server.go` server 子命令
- [x] 添加命令行参数（--port, --config）
- [x] 实现服务器启动逻辑
- [x] 实现模块初始化顺序管理
- [x] 实现信号处理和优雅关闭

## 4. Version 子命令实现
- [x] 实现 `cmd/winpower-g2-exporter/version.go` version 子命令
- [x] 添加输出格式参数（--format）
- [x] 实现版本信息结构
- [x] 实现 JSON 和文本格式输出
- [x] 添加编译时变量注入

## 5. Help 子命令实现
- [x] 实现 `cmd/winpower-g2-exporter/help.go` help 子命令
- [x] 集成 Cobra 帮助系统
- [x] 自定义帮助信息格式
- [x] 添加使用示例

## 6. 配置集成
- [x] 实现配置文件搜索逻辑
- [x] 实现环境变量绑定
- [x] 实现配置验证
- [x] 实现配置错误处理

## 7. 应用程序生命周期管理
- [x] 实现 App 结构体封装所有模块
- [x] 实现模块初始化方法
- [x] 实现模块关闭方法
- [x] 实现启动流程错误处理

## 8. 测试实现
- [x] 编写根命令单元测试
- [x] 编写 server 子命令单元测试
- [x] 编写 version 子命令单元测试
- [x] 编写配置绑定测试
- [x] 编写模块初始化测试
- [x] 编写集成测试

## 9. Makefile 更新
- [x] 添加构建时版本信息注入
- [x] 更新构建命令支持 main.go 入口
- [x] 添加多平台构建支持
- [x] 验证构建脚本正确性

## 10. 文档和验证
- [x] 更新 README.md 添加命令行使用说明
- [x] 验证所有测试通过（make test）
- [x] 验证静态检查通过（make lint）
- [x] 验证测试覆盖率达标（make test-coverage）
- [x] 测试完整的命令行功能

## 依赖关系
- 任务 1-5 可以并行进行
- 任务 6-7 依赖于任务 2-5 的完成
- 任务 8 依赖于任务 1-7 的完成
- 任务 9 可以与任务 8 并行进行
- 任务 10 依赖于所有前面的任务

## 验收标准
- 所有命令行功能正常工作
- 版本信息正确显示和注入
- 模块按正确顺序初始化
- 优雅关闭机制正常
- 单元测试覆盖率 80%以上
- 所有静态检查通过