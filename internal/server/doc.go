// Package server 提供HTTP服务器实现，负责对外暴露API端点
//
// 职责：
//   - 路由注册和HTTP请求处理
//   - 全局中间件管理（Logger、Recovery）
//   - 启动/停止和优雅关闭
//   - 可选的pprof调试支持
//
// 非职责：
//   - 认证流程（由下层模块处理）
//   - 电能计算（由energy模块处理）
//   - 采集逻辑（由collector模块处理）
//   - 指标转换（由metrics模块处理）
//
// 核心端点：
//   - GET /health  - 健康检查
//   - GET /metrics - Prometheus指标导出
//   - GET /debug/pprof/* - 性能分析（可选）
//
// 使用示例：
//
//	config := server.DefaultConfig()
//	config.Port = 9090
//	config.EnablePprof = true
//
//	srv := server.NewHTTPServer(
//	    config,
//	    logger,
//	    metricsService,
//	    healthService,
//	)
//
//	if err := srv.Start(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// 优雅关闭
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	srv.Stop(ctx)
package server
