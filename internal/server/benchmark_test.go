package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

// BenchmarkHTTPServer_HealthCheckLatency 测试健康检查请求处理延迟 (8.2.1)
func BenchmarkHTTPServer_HealthCheckLatency(b *testing.B) {
	// 设置
	config := DefaultConfig()
	config.Port = 38120
	config.Host = "127.0.0.1"
	config.Mode = "release" // 使用release模式获得最佳性能

	// 创建mock服务
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// 设置快速响应以最小化mock延迟
	mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
		"status":       "healthy",
		"latency_test": "active",
	})

	// 设置日志期望（静默以避免影响基准测试）
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Server starting", mock.Anything).Return()
	mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// 创建并启动服务器
	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 准备基准测试
	url := fmt.Sprintf("http://%s:%d/health", config.Host, config.Port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 预热
	for i := 0; i < 10; i++ {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}

	// 重置基准测试计时器
	b.ResetTimer()

	// 执行基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Errorf("请求失败: %v", err)
				continue
			}
			_ = resp.Body.Close()
		}
	})

	// 停止计时器
	b.StopTimer()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Stop(ctx)
	<-startErr

	mockHealth.AssertExpectations(b)
	mockLogger.AssertExpectations(b)
}

// BenchmarkHTTPServer_MetricsLatency 测试指标请求处理延迟 (8.2.1)
func BenchmarkHTTPServer_MetricsLatency(b *testing.B) {
	// 设置
	config := DefaultConfig()
	config.Port = 38121
	config.Host = "127.0.0.1"
	config.Mode = "release"

	// 创建mock服务
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// 设置快速响应的指标
	mockMetrics.On("Render", mock.Anything).Return(`# HELP benchmark_test_metric Test metric for benchmarking
# TYPE benchmark_test_metric gauge
benchmark_test_metric 1
# HELP benchmark_requests_total Total benchmark requests
# TYPE benchmark_requests_total counter
benchmark_requests_total 1000`, nil)

	// 设置日志期望
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Server starting", mock.Anything).Return()
	mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// 创建并启动服务器
	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 准备基准测试
	url := fmt.Sprintf("http://%s:%d/metrics", config.Host, config.Port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 预热
	for i := 0; i < 10; i++ {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}

	// 重置基准测试计时器
	b.ResetTimer()

	// 执行基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Errorf("请求失败: %v", err)
				continue
			}
			_ = resp.Body.Close()
		}
	})

	// 停止计时器
	b.StopTimer()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Stop(ctx)
	<-startErr

	mockMetrics.AssertExpectations(b)
	mockLogger.AssertExpectations(b)
}

// BenchmarkHTTPServer_MixedEndpointsLatency 测试混合端点请求处理延迟 (8.2.1)
func BenchmarkHTTPServer_MixedEndpointsLatency(b *testing.B) {
	// 设置
	config := DefaultConfig()
	config.Port = 38122
	config.Host = "127.0.0.1"
	config.Mode = "release"

	// 创建mock服务
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// 设置响应
	mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
		"mixed_benchmark": "active",
	})

	mockMetrics.On("Render", mock.Anything).Return(`# HELP mixed_benchmark_metric Mixed endpoint benchmark test
# TYPE mixed_benchmark_metric gauge
mixed_benchmark_metric 1`, nil)

	// 设置日志期望
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Server starting", mock.Anything).Return()
	mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// 创建并启动服务器
	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 准备基准测试
	healthURL := fmt.Sprintf("http://%s:%d/health", config.Host, config.Port)
	metricsURL := fmt.Sprintf("http://%s:%d/metrics", config.Host, config.Port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 预热
	for i := 0; i < 10; i++ {
		if resp, err := client.Get(healthURL); err == nil {
			_ = resp.Body.Close()
		}
		if resp, err := client.Get(metricsURL); err == nil {
			_ = resp.Body.Close()
		}
	}

	// 重置基准测试计时器
	b.ResetTimer()

	// 执行基准测试 - 70%健康检查，30%指标
	healthCounter := 0
	metricsCounter := 0

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var url string
			// 70%健康检查，30%指标
			if (healthCounter+metricsCounter)%10 < 7 {
				url = healthURL
				healthCounter++
			} else {
				url = metricsURL
				metricsCounter++
			}

			resp, err := client.Get(url)
			if err != nil {
				b.Errorf("请求失败: %v", err)
				continue
			}
			_ = resp.Body.Close()
		}
	})

	// 停止计时器
	b.StopTimer()

	b.Logf("基准测试统计: 健康检查=%d, 指标=%d", healthCounter, metricsCounter)

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Stop(ctx)
	<-startErr

	mockHealth.AssertExpectations(b)
	mockMetrics.AssertExpectations(b)
	mockLogger.AssertExpectations(b)
}

// BenchmarkHTTPServer_ConcurrentHandling 测试并发处理能力 (8.2.2)
func BenchmarkHTTPServer_ConcurrentHandling(b *testing.B) {
	b.Run("低并发-1个goroutine", func(b *testing.B) {
		benchmarkConcurrentHandling(b, 1)
	})

	b.Run("中等并发-10个goroutine", func(b *testing.B) {
		benchmarkConcurrentHandling(b, 10)
	})

	b.Run("高并发-50个goroutine", func(b *testing.B) {
		benchmarkConcurrentHandling(b, 50)
	})

	b.Run("极高并发-100个goroutine", func(b *testing.B) {
		benchmarkConcurrentHandling(b, 100)
	})
}

// benchmarkConcurrentHandling 并发处理基准测试的辅助函数
func benchmarkConcurrentHandling(b *testing.B, goroutines int) {
	// 设置
	config := DefaultConfig()
	config.Port = 38123 + goroutines // 使用不同端口避免冲突
	config.Host = "127.0.0.1"
	config.Mode = "release"
	config.ReadTimeout = 30 * time.Second
	config.WriteTimeout = 30 * time.Second
	config.IdleTimeout = 60 * time.Second

	// 创建mock服务
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// 设置响应（快速响应以测试服务器并发能力）
	mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
		"concurrent_test": "active",
		"goroutines":      goroutines,
	})

	// 设置日志期望
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Server starting", mock.Anything).Return()
	mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// 创建并启动服务器
	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 准备基准测试
	url := fmt.Sprintf("http://%s:%d/health", config.Host, config.Port)
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 预热
	for i := 0; i < goroutines*2; i++ {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}

	// 重置基准测试计时器
	b.ResetTimer()

	// 执行并发基准测试
	var wg sync.WaitGroup
	requestsPerGoroutine := b.N / goroutines

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				resp, err := client.Get(url)
				if err != nil {
					b.Errorf("请求失败: %v", err)
					continue
				}
				_ = resp.Body.Close()
			}
		}()
	}

	// 等待所有请求完成
	wg.Wait()

	// 停止计时器
	b.StopTimer()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Stop(ctx)
	<-startErr

	mockHealth.AssertExpectations(b)
	mockLogger.AssertExpectations(b)
}

// BenchmarkHTTPServer_MixedWorkloadConcurrent 测试混合工作负载并发处理 (8.2.2)
func BenchmarkHTTPServer_MixedWorkloadConcurrent(b *testing.B) {
	// 设置
	config := DefaultConfig()
	config.Port = 38128
	config.Host = "127.0.0.1"
	config.Mode = "release"
	config.EnableCORS = true
	config.EnableRateLimit = true

	// 创建mock服务
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// 设置响应
	mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
		"mixed_workload": "active",
		"test_type":      "concurrent_mixed",
	})

	mockMetrics.On("Render", mock.Anything).Return(`# HELP mixed_workload_requests Mixed workload concurrent test
# TYPE mixed_workload_requests counter
mixed_workload_requests 1`, nil)

	// 设置日志期望
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Server starting", mock.Anything).Return()
	mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", "Rate limit exceeded", mock.Anything).Return().Maybe() // 限流警告可能触发

	// 创建并启动服务器
	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 准备基准测试
	healthURL := fmt.Sprintf("http://%s:%d/health", config.Host, config.Port)
	metricsURL := fmt.Sprintf("http://%s:%d/metrics", config.Host, config.Port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 预热
	for i := 0; i < 20; i++ {
		if resp, err := client.Get(healthURL); err == nil {
			_ = resp.Body.Close()
		}
		if resp, err := client.Get(metricsURL); err == nil {
			_ = resp.Body.Close()
		}
	}

	// 重置基准测试计时器
	b.ResetTimer()

	// 并发参数
	const (
		healthGoroutines  = 15
		metricsGoroutines = 10
	)

	var wg sync.WaitGroup
	requestsPerGoroutine := b.N / (healthGoroutines + metricsGoroutines)

	// 健康检查goroutines
	for i := 0; i < healthGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				resp, err := client.Get(healthURL)
				if err != nil {
					b.Errorf("健康检查请求失败 (goroutine %d): %v", goroutineID, err)
					continue
				}
				_ = resp.Body.Close()
			}
		}(i)
	}

	// 指标goroutines
	for i := 0; i < metricsGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				resp, err := client.Get(metricsURL)
				if err != nil {
					b.Errorf("指标请求失败 (goroutine %d): %v", goroutineID, err)
					continue
				}
				_ = resp.Body.Close()
			}
		}(i)
	}

	// 等待所有请求完成
	wg.Wait()

	// 停止计时器
	b.StopTimer()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_ = server.Stop(ctx)
	<-startErr

	mockHealth.AssertExpectations(b)
	mockMetrics.AssertExpectations(b)
	mockLogger.AssertExpectations(b)
}

// BenchmarkHTTPServer_MiddlewareOverhead 测试中间件开销 (8.2.1)
func BenchmarkHTTPServer_MiddlewareOverhead(b *testing.B) {
	b.Run("最小中间件", func(b *testing.B) {
		benchmarkMiddlewareOverhead(b, false, false, false)
	})

	b.Run("启用CORS", func(b *testing.B) {
		benchmarkMiddlewareOverhead(b, true, false, false)
	})

	b.Run("启用限流", func(b *testing.B) {
		benchmarkMiddlewareOverhead(b, false, true, false)
	})

	b.Run("启用全部中间件", func(b *testing.B) {
		benchmarkMiddlewareOverhead(b, true, true, false)
	})
}

// benchmarkMiddlewareOverhead 中间件开销基准测试的辅助函数
func benchmarkMiddlewareOverhead(b *testing.B, enableCORS, enableRateLimit, enablePprof bool) {
	// 设置
	config := DefaultConfig()
	config.Port = 38129
	config.Host = "127.0.0.1"
	config.Mode = "release"
	config.EnableCORS = enableCORS
	config.EnableRateLimit = enableRateLimit
	config.EnablePprof = enablePprof

	// 创建mock服务
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// 设置响应
	mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
		"middleware_test": "active",
		"cors":            enableCORS,
		"rate_limit":      enableRateLimit,
		"pprof":           enablePprof,
	})

	// 设置日志期望
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Server starting", mock.Anything).Return()
	mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
	mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", "Rate limit exceeded", mock.Anything).Return().Maybe() // 限流警告可能触发

	// 创建并启动服务器
	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	startErr := make(chan error, 1)
	go func() {
		startErr <- server.Start()
	}()

	// 等待服务器启动
	time.Sleep(200 * time.Millisecond)

	// 准备基准测试
	url := fmt.Sprintf("http://%s:%d/health", config.Host, config.Port)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 预热
	for i := 0; i < 10; i++ {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}

	// 重置基准测试计时器
	b.ResetTimer()

	// 执行基准测试
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(url)
			if err != nil {
				b.Errorf("请求失败: %v", err)
				continue
			}
			_ = resp.Body.Close()
		}
	})

	// 停止计时器
	b.StopTimer()

	// 关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Stop(ctx)
	<-startErr

	mockHealth.AssertExpectations(b)
	mockLogger.AssertExpectations(b)
}

// BenchmarkHTTPServer_DifferentModes 测试不同运行模式的性能 (8.2.1)
func BenchmarkHTTPServer_DifferentModes(b *testing.B) {
	modes := []string{"debug", "release", "test"}

	for _, mode := range modes {
		b.Run("模式_"+mode, func(b *testing.B) {
			// 设置
			config := DefaultConfig()
			config.Port = 38130
			config.Host = "127.0.0.1"
			config.Mode = mode

			// 创建mock服务
			mockMetrics := new(MockMetricsService)
			mockHealth := new(MockHealthService)
			mockLogger := new(MockLogger)

			// 设置响应
			mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
				"mode":             mode,
				"performance_test": "active",
			})

			// 设置日志期望
			mockLogger.On("With", mock.Anything).Return(mockLogger)
			mockLogger.On("Info", "Server starting", mock.Anything).Return()
			mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
			mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
			mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
			mockLogger.On("Info", mock.Anything, mock.Anything).Return()

			// 创建并启动服务器
			server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

			startErr := make(chan error, 1)
			go func() {
				startErr <- server.Start()
			}()

			// 等待服务器启动
			time.Sleep(200 * time.Millisecond)

			// 准备基准测试
			url := fmt.Sprintf("http://%s:%d/health", config.Host, config.Port)
			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			// 预热
			for i := 0; i < 10; i++ {
				resp, err := client.Get(url)
				if err == nil {
					_ = resp.Body.Close()
				}
			}

			// 重置基准测试计时器
			b.ResetTimer()

			// 执行基准测试
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					resp, err := client.Get(url)
					if err != nil {
						b.Errorf("请求失败: %v", err)
						continue
					}
					_ = resp.Body.Close()
				}
			})

			// 停止计时器
			b.StopTimer()

			// 关闭服务器
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = server.Stop(ctx)
			<-startErr

			mockHealth.AssertExpectations(b)
			mockLogger.AssertExpectations(b)
		})
	}
}
