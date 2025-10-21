package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ServerLifecycle 测试完整的服务器生命周期 (8.1.1)
func TestIntegration_ServerLifecycle(t *testing.T) {
	t.Run("完整的启动-运行-关闭生命周期", func(t *testing.T) {
		// 设置
		config := DefaultConfig()
		config.Port = 38101 // 使用唯一端口避免冲突
		config.Host = "127.0.0.1"
		config.Mode = "release"
		config.ReadTimeout = 5 * time.Second
		config.WriteTimeout = 5 * time.Second
		config.IdleTimeout = 10 * time.Second

		// 创建mock服务
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// 设置健康检查响应
		mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
			"version": "1.0.0",
			"uptime":  "0s",
		})

		// 设置指标响应
		mockMetrics.On("Render", mock.Anything).Return(`# HELP test_metric value
# TYPE test_metric gauge
test_metric 1`, nil)

		// 设置日志期望
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server starting", mock.Anything).Return()
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()
		mockLogger.On("Info", mock.Anything, mock.Anything).Return() // 用于请求日志

		// 创建服务器
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// 启动服务器
		startErr := make(chan error, 1)
		started := make(chan bool, 1)
		go func() {
			err := server.Start()
			startErr <- err
		}()

		// 等待服务器启动
		go func() {
			// 尝试连接服务器直到成功
			for i := 0; i < 50; i++ {
				resp, err := http.Get(fmt.Sprintf("http://%s:%d/health", config.Host, config.Port))
				if err == nil {
					_ = resp.Body.Close()
					started <- true
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
			started <- false
		}()

		// 验证服务器启动成功
		select {
		case success := <-started:
			require.True(t, success, "服务器应该成功启动并可接受连接")
		case <-time.After(5 * time.Second):
			t.Fatal("服务器启动超时")
		}

		// 测试健康检查端点
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/health", config.Host, config.Port))
		require.NoError(t, err, "健康检查请求应该成功")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "健康检查应该返回200")
		assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

		var healthResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthResp)
		require.NoError(t, err, "健康检查响应应该是有效的JSON")
		assert.Equal(t, "ok", healthResp["status"], "健康状态应该是ok")

		// 测试指标端点
		resp, err = http.Get(fmt.Sprintf("http://%s:%d/metrics", config.Host, config.Port))
		require.NoError(t, err, "指标请求应该成功")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "指标请求应该返回200")
		assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", resp.Header.Get("Content-Type"))

		metricsBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err, "应该能够读取指标响应")
		assert.Contains(t, string(metricsBody), "test_metric 1", "指标响应应该包含预期的指标")

		// 优雅关闭服务器
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		stopErr := server.Stop(ctx)
		assert.NoError(t, stopErr, "服务器应该能够优雅关闭")

		// 等待服务器完全关闭
		select {
		case err := <-startErr:
			// Start()应该在关闭后返回nil或http.ErrServerClosed
			assert.True(t, err == nil || err == http.ErrServerClosed,
				"Start()应该返回nil或ErrServerClosed，得到: %v", err)
		case <-time.After(5 * time.Second):
			t.Error("服务器应该在关闭超时前返回")
		}

		// 验证服务器确实已关闭
		_, err = http.Get(fmt.Sprintf("http://%s:%d/health", config.Host, config.Port))
		assert.Error(t, err, "服务器关闭后不应该接受连接")

		// 验证所有mock期望都被满足
		mockHealth.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
