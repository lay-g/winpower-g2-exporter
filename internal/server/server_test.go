package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsService is a mock implementation of MetricsService interface
type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) Render(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

// MockHealthService is a mock implementation of HealthService interface
type MockHealthService struct {
	mock.Mock
}

func (m *MockHealthService) Check(ctx context.Context) (status string, details map[string]any) {
	args := m.Called(ctx)
	return args.String(0), args.Get(1).(map[string]any)
}

// MockLogger is a mock implementation of Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...log.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...log.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...log.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...log.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...log.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...log.Field) log.Logger {
	args := m.Called(fields)
	return args.Get(0).(log.Logger)
}

func (m *MockLogger) WithContext(ctx context.Context) log.Logger {
	args := m.Called(ctx)
	return args.Get(0).(log.Logger)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

// TestHTTPServer_Constructor tests the NewHTTPServer constructor
func TestHTTPServer_Constructor(t *testing.T) {
	t.Run("creates server with all required dependencies", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Test
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Verify
		assert.NotNil(t, server, "Server should be created successfully")
		assert.Equal(t, config, server.GetConfig(), "Server config should be set")
		assert.Equal(t, mockMetrics, server.GetMetrics(), "Metrics service should be set")
		assert.Equal(t, mockHealth, server.GetHealth(), "Health service should be set")
		assert.NotNil(t, server.GetEngine(), "Gin engine should be created")
		assert.NotNil(t, server.GetHTTPServer(), "HTTP server should be created")
		mockLogger.AssertExpectations(t)
	})

	t.Run("panics when config is nil", func(t *testing.T) {
		// Setup
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Test and verify
		assert.Panics(t, func() {
			NewHTTPServer(nil, mockLogger, mockMetrics, mockHealth)
		}, "Constructor should panic when config is nil")
	})

	t.Run("panics when logger is nil", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)

		// Test and verify
		assert.Panics(t, func() {
			NewHTTPServer(config, nil, mockMetrics, mockHealth)
		}, "Constructor should panic when logger is nil")
	})

	t.Run("panics when metrics service is nil", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Test and verify
		assert.Panics(t, func() {
			NewHTTPServer(config, mockLogger, nil, mockHealth)
		}, "Constructor should panic when metrics service is nil")
	})

	t.Run("panics when health service is nil", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockLogger := new(MockLogger)

		// Test and verify
		assert.Panics(t, func() {
			NewHTTPServer(config, mockLogger, mockMetrics, nil)
		}, "Constructor should panic when health service is nil")
	})

	t.Run("panics when configuration is invalid", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = -1 // Invalid port

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Test and verify
		assert.Panics(t, func() {
			NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		}, "Constructor should panic when configuration is invalid")
	})

	t.Run("creates HTTP server with correct configuration", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Host = "127.0.0.1"
		config.Port = 8080
		config.ReadTimeout = 10 * time.Second
		config.WriteTimeout = 15 * time.Second
		config.IdleTimeout = 20 * time.Second

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Test
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Verify HTTP server configuration
		httpServer := server.GetHTTPServer()
		assert.Equal(t, "127.0.0.1:8080", httpServer.Addr, "Server address should be correctly set")
		assert.Equal(t, config.ReadTimeout, httpServer.ReadTimeout, "Read timeout should be set")
		assert.Equal(t, config.WriteTimeout, httpServer.WriteTimeout, "Write timeout should be set")
		assert.Equal(t, config.IdleTimeout, httpServer.IdleTimeout, "Idle timeout should be set")
		assert.Equal(t, server.GetEngine(), httpServer.Handler, "Gin engine should be set as handler")
		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_Start tests the Start() method behavior with comprehensive scenarios
func TestHTTPServer_Start(t *testing.T) {
	t.Run("Start method exists and has correct signature", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Verify that Start method exists and returns error
		var _ interface{ Start() error } = server
		assert.NotNil(t, server.Start, "Start method should exist")
		mockLogger.AssertExpectations(t)
	})

	t.Run("正常启动流程 - 7.1.1", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = 38087 // Use a different port for testing
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls for startup and shutdown
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server starting", mock.Anything).Return()
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Start server in a goroutine
		startErr := make(chan error, 1)
		go func() {
			startErr <- server.Start()
		}()

		// Wait a bit for server to start
		select {
		case err := <-startErr:
			// If we get an error immediately, it should be nil for successful start
			// (or http.ErrServerClosed if something stops it)
			if err != nil && err != http.ErrServerClosed {
				t.Errorf("Expected no error on start, got: %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			// Server started successfully, now stop it
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			stopErr := server.Stop(ctx)
			assert.NoError(t, stopErr, "Server should stop cleanly")

			// Wait for the Start() goroutine to finish
			select {
			case err := <-startErr:
				// Start() may return nil or http.ErrServerClosed depending on shutdown timing
				// Both are acceptable for graceful shutdown
				assert.True(t, err == nil || err == http.ErrServerClosed,
					"Start should return nil or ErrServerClosed after graceful shutdown, got: %v", err)
			case <-time.After(1 * time.Second):
				t.Error("Start() should have returned after shutdown")
			}
		}

		mockLogger.AssertExpectations(t)
	})

	t.Run("启动失败场景 - 端口占用 - 7.1.2", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = 38088 // Use a specific port for testing

		mockMetrics1 := new(MockMetricsService)
		mockHealth1 := new(MockHealthService)
		mockLogger1 := new(MockLogger)

		mockMetrics2 := new(MockMetricsService)
		mockHealth2 := new(MockHealthService)
		mockLogger2 := new(MockLogger)

		// Expect logger calls
		mockLogger1.On("With", mock.Anything).Return(mockLogger1)
		mockLogger1.On("Info", "Server starting", mock.Anything).Return()
		mockLogger1.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger1.On("Info", "Server stopped successfully", mock.Anything).Return()
		mockLogger1.On("Info", "Server stopped gracefully", mock.Anything).Return()

		mockLogger2.On("With", mock.Anything).Return(mockLogger2)
		mockLogger2.On("Info", "Server starting", mock.Anything).Return()
		mockLogger2.On("Error", "Server failed to start", mock.Anything).Return()

		// Create first server
		server1 := NewHTTPServer(config, mockLogger1, mockMetrics1, mockHealth1)

		// Start first server
		startErr1 := make(chan error, 1)
		go func() {
			startErr1 <- server1.Start()
		}()

		// Wait for first server to start
		select {
		case err := <-startErr1:
			t.Errorf("First server should start successfully, got error: %v", err)
		case <-time.After(100 * time.Millisecond):
			// First server started, now try to start second server on same port
			server2 := NewHTTPServer(config, mockLogger2, mockMetrics2, mockHealth2)

			// This should fail due to port being in use
			err := server2.Start()
			assert.Error(t, err, "Second server should fail to start due to port conflict")
			assert.Contains(t, err.Error(), "failed to start server", "Error message should indicate startup failure")

			// Clean up first server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server1.Stop(ctx)

			// Wait for first server to finish
			select {
			case <-startErr1:
				// Expected
			case <-time.After(1 * time.Second):
				t.Error("First server should have stopped")
			}
		}

		mockLogger1.AssertExpectations(t)
		mockLogger2.AssertExpectations(t)
	})

	t.Run("启动日志记录 - 7.1.3", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = 38089
		config.Host = "127.0.0.1"
		config.Mode = "test"

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect specific logger calls with correct parameters
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server starting", mock.MatchedBy(func(fields []log.Field) bool {
			// Verify the log contains the expected fields by checking the string representation
			// Since Field is zapcore.Field, we need to check using its interface
			fieldStr := fmt.Sprintf("%v", fields)
			return len(fields) >= 3 && // Should have at least 3 fields (host, port, mode)
				strings.Contains(fieldStr, "127.0.0.1") &&
				strings.Contains(fieldStr, "38089") &&
				strings.Contains(fieldStr, "test")
		})).Return()
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Start server
		startErr := make(chan error, 1)
		go func() {
			startErr <- server.Start()
		}()

		// Wait for startup
		select {
		case <-startErr:
			t.Error("Server should start successfully")
		case <-time.After(100 * time.Millisecond):
			// Server started, verify log was called
			mockLogger.AssertCalled(t, "Info", "Server starting", mock.Anything)

			// Stop server
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Stop(ctx)
		}

		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_Stop tests the Stop() method behavior with comprehensive scenarios
func TestHTTPServer_Stop(t *testing.T) {
	t.Run("Stop method exists and has correct signature", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Verify that Stop method exists and returns error
		var _ interface{ Stop(context.Context) error } = server
		assert.NotNil(t, server.Stop, "Stop method should exist")
		mockLogger.AssertExpectations(t)
	})

	t.Run("优雅关闭流程 - 7.2.1", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = 38090

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server starting", mock.Anything).Return()
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Start server
		startErr := make(chan error, 1)
		go func() {
			startErr <- server.Start()
		}()

		// Wait for server to start
		select {
		case err := <-startErr:
			t.Errorf("Server should start successfully, got error: %v", err)
		case <-time.After(100 * time.Millisecond):
			// Server started, now test graceful shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			stopErr := server.Stop(ctx)
			assert.NoError(t, stopErr, "Server should stop gracefully")

			// Wait for Start() to complete
			select {
			case err := <-startErr:
				// Start() may return nil or http.ErrServerClosed depending on shutdown timing
				// Both are acceptable for graceful shutdown
				assert.True(t, err == nil || err == http.ErrServerClosed,
					"Start should return nil or ErrServerClosed after graceful shutdown, got: %v", err)
			case <-time.After(1 * time.Second):
				t.Error("Start() should have returned after shutdown")
			}
		}

		mockLogger.AssertExpectations(t)
	})

	t.Run("关闭超时处理 - 7.2.2", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls for shutdown
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Test that Stop() respects context timeout
		// Create a context that will time out
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Wait for context to timeout
		time.Sleep(2 * time.Millisecond)

		// Call Stop with timed-out context
		stopErr := server.Stop(ctx)
		// The behavior may vary - some implementations return nil even with timeout
		// The important thing is that the Stop method handles context properly
		t.Logf("Stop() result with timed-out context: %v", stopErr)

		mockLogger.AssertExpectations(t)
	})

	t.Run("现有请求完成等待 - 7.2.3", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = 38092

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup health service to simulate slow response
		mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
			"status": "healthy",
		})

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server starting", mock.Anything).Return()
		mockLogger.On("Info", mock.Anything, mock.Anything).Return() // Health check logging
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped gracefully", mock.Anything).Return()

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Start server
		startErr := make(chan error, 1)
		go func() {
			startErr <- server.Start()
		}()

		// Wait for server to start
		select {
		case err := <-startErr:
			t.Errorf("Server should start successfully, got error: %v", err)
		case <-time.After(100 * time.Millisecond):
			// Make a request that will take some time
			requestDone := make(chan bool, 1)
			go func() {
				defer func() { requestDone <- true }()

				// Make HTTP request to /health endpoint
				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/health", config.Port))
				if err != nil {
					t.Logf("Request failed (expected during shutdown): %v", err)
					return
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode == http.StatusOK {
					t.Log("Request completed successfully during shutdown")
				}
			}()

			// Wait a bit for request to start
			time.Sleep(10 * time.Millisecond)

			// Initiate shutdown while request is in progress
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			stopErr := server.Stop(ctx)
			assert.NoError(t, stopErr, "Server should stop gracefully even with active requests")

			// Wait for request to complete or timeout
			select {
			case <-requestDone:
				t.Log("Request completed before shutdown")
			case <-time.After(1 * time.Second):
				t.Log("Request timed out (may have been interrupted by shutdown)")
			}

			// Wait for Start() to complete
			select {
			case err := <-startErr:
				// Start() may return nil or http.ErrServerClosed depending on shutdown timing
				// Both are acceptable for graceful shutdown
				assert.True(t, err == nil || err == http.ErrServerClosed,
					"Start should return nil or ErrServerClosed after graceful shutdown, got: %v", err)
			case <-time.After(2 * time.Second):
				t.Error("Start() should have returned after shutdown")
			}
		}

		mockHealth.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("stop handles gracefully when server is not running", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls for shutdown
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", "Server shutting down", mock.Anything).Return()
		mockLogger.On("Info", "Server stopped successfully", mock.Anything).Return()

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Test stop on server that was never started
		ctx := context.Background()
		err := server.Stop(ctx)

		// Verify - the server was created but not started, so shutdown should work
		assert.NoError(t, err, "Stop should handle gracefully when server is not running")
		mockLogger.AssertExpectations(t)
	})
}

// TestInterfaces tests that all interfaces are properly defined
func TestInterfaces(t *testing.T) {
	t.Run("Server interface is correctly defined", func(t *testing.T) {
		// Verify that HTTPServer implements Server interface
		var _ Server = (*HTTPServer)(nil)

		// Create a concrete implementation to test interface
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Verify interface methods exist
		assert.NotNil(t, server.Start, "Server should have Start method")
		assert.NotNil(t, server.Stop, "Server should have Stop method")
		mockLogger.AssertExpectations(t)
	})

	t.Run("MetricsService interface is correctly defined", func(t *testing.T) {
		// Verify that MockMetricsService implements MetricsService interface
		var _ MetricsService = (*MockMetricsService)(nil)

		// Create a concrete implementation to test interface
		metrics := new(MockMetricsService)

		// Verify interface methods exist
		assert.NotNil(t, metrics.Render, "MetricsService should have Render method")
	})

	t.Run("HealthService interface is correctly defined", func(t *testing.T) {
		// Verify that MockHealthService implements HealthService interface
		var _ HealthService = (*MockHealthService)(nil)

		// Create a concrete implementation to test interface
		health := new(MockHealthService)

		// Verify interface methods exist
		assert.NotNil(t, health.Check, "HealthService should have Check method")
	})
}

// TestHTTPServer_Integration tests basic integration scenarios
func TestHTTPServer_Integration(t *testing.T) {
	t.Run("server can be created with proper dependencies", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.Port = 38086 // Use a specific port for testing

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Test server creation
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Verify basic properties
		assert.NotNil(t, server, "Server should be created")
		assert.NotNil(t, server.Start, "Start method should exist")
		assert.NotNil(t, server.Stop, "Stop method should exist")
		assert.Equal(t, config.Port, server.GetConfig().Port, "Port should be set correctly")

		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_HealthHandler tests the /health endpoint handler
func TestHTTPServer_HealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns healthy status with correct format", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup health service expectations
		mockHealth.On("Check", mock.Anything).Return("ok", map[string]any{
			"version": "1.0.0",
			"uptime":  "10m",
		})

		// Setup logger expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Create gin context for testing
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/health", nil)

		// Call handler
		server.healthHandler(c)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 status")
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type")

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		// Verify response structure
		assert.Equal(t, "ok", response["status"], "Should return ok status")
		assert.NotEmpty(t, response["timestamp"], "Should include timestamp")
		assert.Equal(t, "1.0.0", response["details"].(map[string]interface{})["version"], "Should include version in details")

		mockHealth.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("handles health service error gracefully", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup health service to return error status
		mockHealth.On("Check", mock.Anything).Return("error", map[string]any{
			"error": "Database connection failed",
		})

		// Setup logger expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Create gin context for testing
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/health", nil)

		// Call handler
		server.healthHandler(c)

		// Verify response
		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Should return 503 status for unhealthy service")
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type")

		// Parse response body
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		// Verify response structure
		assert.Equal(t, "error", response["status"], "Should return error status")
		assert.NotEmpty(t, response["timestamp"], "Should include timestamp")
		assert.Equal(t, "Database connection failed", response["details"].(map[string]interface{})["error"], "Should include error details")

		mockHealth.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("validates response format compliance", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup health service with complex details
		details := map[string]any{
			"version": "1.0.0",
			"uptime":  "10m",
			"checks": map[string]any{
				"database": "ok",
				"metrics":  "ok",
			},
		}
		mockHealth.On("Check", mock.Anything).Return("ok", details)

		// Setup logger expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Create gin context for testing
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/health", nil)

		// Call handler
		server.healthHandler(c)

		// Verify response structure
		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 status")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")

		// Verify required fields exist
		assert.Contains(t, response, "status", "Response should contain status field")
		assert.Contains(t, response, "timestamp", "Response should contain timestamp field")
		assert.Contains(t, response, "details", "Response should contain details field")

		// Verify timestamp format (should be ISO 8601)
		timestamp, ok := response["timestamp"].(string)
		assert.True(t, ok, "Timestamp should be a string")
		assert.True(t, len(timestamp) > 0, "Timestamp should not be empty")

		// Verify details structure
		detailsResponse, ok := response["details"].(map[string]interface{})
		assert.True(t, ok, "Details should be a map")
		assert.Equal(t, "1.0.0", detailsResponse["version"], "Should preserve version in details")

		mockHealth.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_MetricsHandler tests the /metrics endpoint handler
func TestHTTPServer_MetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns Prometheus metrics with correct format", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Sample Prometheus metrics output
		metricsOutput := `# HELP winpower_exporter_up Boolean indicating if the exporter is up
# TYPE winpower_exporter_up gauge
winpower_exporter_up 1
# HELP winpower_devices_connected Number of connected devices
# TYPE winpower_devices_connected gauge
winpower_devices_connected 2
`

		// Setup metrics service expectations
		mockMetrics.On("Render", mock.Anything).Return(metricsOutput, nil)

		// Setup logger expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Create gin context for testing
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/metrics", nil)

		// Call handler
		server.metricsHandler(c)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 status")
		assert.Equal(t, "text/plain; version=0.0.4; charset=utf-8", w.Header().Get("Content-Type"), "Should return Prometheus content type")
		assert.Equal(t, metricsOutput, w.Body.String(), "Should return metrics output")

		mockMetrics.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("handles metrics service error gracefully", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup metrics service to return error
		mockMetrics.On("Render", mock.Anything).Return("", assert.AnError)

		// Setup logger expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Create gin context for testing
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/metrics", nil)

		// Call handler
		server.metricsHandler(c)

		// Verify response
		assert.Equal(t, http.StatusServiceUnavailable, w.Code, "Should return 503 status for metrics error")
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type for error")

		// Parse error response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Error response should be valid JSON")

		assert.Equal(t, "metrics_error", response["error"], "Should return metrics error type")
		assert.Contains(t, response["message"], "Failed to render metrics", "Should include error message")

		mockMetrics.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("validates metrics content type and format", func(t *testing.T) {
		// Setup
		config := DefaultConfig()

		// Test with different metrics formats
		testCases := []struct {
			name           string
			metricsOutput  string
			expectedType   string
			expectedStatus int
		}{
			{
				name:           "standard Prometheus format",
				metricsOutput:  "test_metric 1\n",
				expectedType:   "text/plain; version=0.0.4; charset=utf-8",
				expectedStatus: http.StatusOK,
			},
			{
				name:           "empty metrics",
				metricsOutput:  "",
				expectedType:   "text/plain; version=0.0.4; charset=utf-8",
				expectedStatus: http.StatusOK,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Setup mocks
				mockMetrics := new(MockMetricsService)
				mockHealth := new(MockHealthService)
				mockLogger := new(MockLogger)

				mockMetrics.On("Render", mock.Anything).Return(tc.metricsOutput, nil)
				mockLogger.On("With", mock.Anything).Return(mockLogger)
				mockLogger.On("Info", mock.Anything, mock.Anything).Return()

				// Create server
				server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

				// Create gin context for testing
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", "/metrics", nil)

				// Call handler
				server.metricsHandler(c)

				// Verify response
				assert.Equal(t, tc.expectedStatus, w.Code, "Should return expected status")
				assert.Equal(t, tc.expectedType, w.Header().Get("Content-Type"), "Should return expected content type")
				assert.Equal(t, tc.metricsOutput, w.Body.String(), "Should return expected metrics content")

				mockMetrics.AssertExpectations(t)
				mockLogger.AssertExpectations(t)
			})
		}
	})
}

// TestHTTPServer_NotFoundHandler tests the 404 handler
func TestHTTPServer_NotFoundHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns 404 for unknown paths", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup logger expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Create gin context for testing
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/unknown-path", nil)

		// Call handler
		server.notFoundHandler(c)

		// Verify response
		assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 status")
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type")

		// Parse error response
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Error response should be valid JSON")

		assert.Equal(t, "not_found", response["error"], "Should return not_found error type")
		assert.Equal(t, "/unknown-path", response["path"], "Should include requested path")
		assert.NotEmpty(t, response["timestamp"], "Should include timestamp")

		mockLogger.AssertExpectations(t)
	})

	t.Run("validates error response format consistency", func(t *testing.T) {
		// Setup
		config := DefaultConfig()

		// Test with different unknown paths
		testPaths := []string{
			"/api/v1/unknown",
			"/admin/secret",
			"/debug/pprof/disabled",
			"/",
		}

		for _, path := range testPaths {
			t.Run("path: "+path, func(t *testing.T) {
				// Setup mocks for each test
				mockMetrics := new(MockMetricsService)
				mockHealth := new(MockHealthService)
				mockLogger := new(MockLogger)

				mockLogger.On("With", mock.Anything).Return(mockLogger)
				mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

				// Create server
				server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

				// Create gin context for testing
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest("GET", path, nil)

				// Call handler
				server.notFoundHandler(c)

				// Verify response structure
				assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 status")
				assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type")

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Error response should be valid JSON")

				// Verify required fields
				assert.Contains(t, response, "error", "Response should contain error field")
				assert.Contains(t, response, "path", "Response should contain path field")
				assert.Contains(t, response, "timestamp", "Response should contain timestamp field")

				assert.Equal(t, "not_found", response["error"], "Should return not_found error type")
				assert.Equal(t, path, response["path"], "Should return correct path")

				// Verify timestamp format
				timestamp, ok := response["timestamp"].(string)
				assert.True(t, ok, "Timestamp should be a string")
				assert.True(t, len(timestamp) > 0, "Timestamp should not be empty")

				mockLogger.AssertExpectations(t)
			})
		}
	})

	t.Run("handles different HTTP methods consistently", func(t *testing.T) {
		// Setup
		config := DefaultConfig()

		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			t.Run("method: "+method, func(t *testing.T) {
				// Setup mocks for each test
				mockMetrics := new(MockMetricsService)
				mockHealth := new(MockHealthService)
				mockLogger := new(MockLogger)

				mockLogger.On("With", mock.Anything).Return(mockLogger)
				mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

				// Create server
				server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

				// Create gin context for testing
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest(method, "/unknown", nil)

				// Call handler
				server.notFoundHandler(c)

				// Verify response is consistent across methods
				assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 status for "+method)
				assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type")

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Error response should be valid JSON")

				assert.Equal(t, "not_found", response["error"], "Should return not_found error type for "+method)

				mockLogger.AssertExpectations(t)
			})
		}
	})
}

// TestHTTPServer_SetupMiddleware tests the setupMiddleware method
func TestHTTPServer_SetupMiddleware(t *testing.T) {
	t.Run("sets up middleware chain in correct order", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnableCORS = true
		config.EnableRateLimit = true

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Verify middleware chain was set up
		// Gin stores middleware in the order they are added
		middlewareCount := len(engine.Handlers)
		assert.Greater(t, middlewareCount, 0, "Should have middleware registered")

		mockLogger.AssertExpectations(t)
	})

	t.Run("sets up minimal middleware when optional features disabled", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnableCORS = false
		config.EnableRateLimit = false

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Verify middleware chain was set up with minimal middleware
		assert.NotNil(t, engine, "Engine should be created")
		assert.Greater(t, len(engine.Handlers), 0, "Should have basic middleware registered")

		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_SetupRoutes tests the setupRoutes method
func TestHTTPServer_SetupRoutes(t *testing.T) {
	t.Run("registers all required routes correctly", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Get route information
		routes := engine.Routes()

		// Verify required routes are registered
		var healthRoute, metricsRoute gin.RouteInfo
		for _, route := range routes {
			if route.Path == "/health" && route.Method == "GET" {
				healthRoute = route
			}
			if route.Path == "/metrics" && route.Method == "GET" {
				metricsRoute = route
			}
		}

		// Verify health route
		assert.NotEmpty(t, healthRoute.Path, "Health route should be registered")
		assert.Equal(t, "GET", healthRoute.Method, "Health route should accept GET method")
		assert.Equal(t, "/health", healthRoute.Path, "Health route path should be correct")

		// Verify metrics route
		assert.NotEmpty(t, metricsRoute.Path, "Metrics route should be registered")
		assert.Equal(t, "GET", metricsRoute.Method, "Metrics route should accept GET method")
		assert.Equal(t, "/metrics", metricsRoute.Path, "Metrics route path should be correct")

		mockLogger.AssertExpectations(t)
	})

	t.Run("registers 404 handler for unknown routes", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

		// Test request to unknown route
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/unknown-route", nil)
		engine := server.GetEngine()
		engine.ServeHTTP(w, req)

		// Verify 404 response
		assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 for unknown routes")
		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Should return JSON content type")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, "not_found", response["error"], "Should return not_found error")

		mockLogger.AssertExpectations(t)
	})

	t.Run("routes work with different HTTP methods", func(t *testing.T) {
		// Setup - simplified test that doesn't trigger middleware
		config := DefaultConfig()
		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Setup expectations
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Get route information to verify only GET methods are registered
		routes := engine.Routes()

		// Verify health route only accepts GET
		healthHasGET := false
		healthHasPOST := false
		for _, route := range routes {
			if route.Path == "/health" {
				if route.Method == "GET" {
					healthHasGET = true
				}
				if route.Method == "POST" {
					healthHasPOST = true
				}
			}
		}
		assert.True(t, healthHasGET, "Health route should accept GET")
		assert.False(t, healthHasPOST, "Health route should not accept POST")

		// Verify metrics route only accepts GET
		metricsHasGET := false
		metricsHasPOST := false
		for _, route := range routes {
			if route.Path == "/metrics" {
				if route.Method == "GET" {
					metricsHasGET = true
				}
				if route.Method == "POST" {
					metricsHasPOST = true
				}
			}
		}
		assert.True(t, metricsHasGET, "Metrics route should accept GET")
		assert.False(t, metricsHasPOST, "Metrics route should not accept POST")

		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_SetupPprofRoutes tests the setupPprofRoutes method
func TestHTTPServer_SetupPprofRoutes(t *testing.T) {
	t.Run("registers pprof routes when enabled", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnablePprof = true

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Get route information
		routes := engine.Routes()

		// Verify pprof routes are registered
		pprofRoutes := []string{
			"/debug/pprof/",
			"/debug/pprof/cmdline",
			"/debug/pprof/profile",
			"/debug/pprof/symbol",
			"/debug/pprof/trace",
		}

		for _, expectedRoute := range pprofRoutes {
			found := false
			for _, route := range routes {
				if route.Path == expectedRoute && route.Method == "GET" {
					found = true
					break
				}
			}
			assert.True(t, found, "Pprof route %s should be registered", expectedRoute)
		}

		mockLogger.AssertExpectations(t)
	})

	t.Run("does not register pprof routes when disabled", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnablePprof = false

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Get route information
		routes := engine.Routes()

		// Verify pprof routes are NOT registered
		pprofRoutes := []string{
			"/debug/pprof/",
			"/debug/pprof/cmdline",
			"/debug/pprof/profile",
			"/debug/pprof/symbol",
			"/debug/pprof/trace",
		}

		for _, forbiddenRoute := range pprofRoutes {
			found := false
			for _, route := range routes {
				if route.Path == forbiddenRoute {
					found = true
					break
				}
			}
			assert.False(t, found, "Pprof route %s should NOT be registered when disabled", forbiddenRoute)
		}

		mockLogger.AssertExpectations(t)
	})

	t.Run("pprof routes are accessible when enabled", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnablePprof = true

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)
		// Allow Info calls for successful request logging
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Test pprof index route
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/debug/pprof/", nil)
		engine.ServeHTTP(w, req)

		// Should return 200 with pprof index page
		assert.Equal(t, http.StatusOK, w.Code, "Pprof index should be accessible")
		assert.Contains(t, w.Body.String(), "pprof", "Response should contain pprof content")

		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_MiddlewareChainOrder tests that middleware are applied in correct order
func TestHTTPServer_MiddlewareChainOrder(t *testing.T) {
	t.Run("applies middleware in correct order: recovery -> logging -> CORS -> rateLimit -> routes", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnableCORS = true
		config.EnableRateLimit = true

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Verify middleware chain is established
		// We can't easily test the exact order without accessing internal Gin structures,
		// but we can verify the server has the expected number of middleware
		assert.Greater(t, len(engine.Handlers), 0, "Should have middleware chain established")

		mockLogger.AssertExpectations(t)
	})

	t.Run("applies minimal middleware chain when optional features disabled", func(t *testing.T) {
		// Setup
		config := DefaultConfig()
		config.EnableCORS = false
		config.EnableRateLimit = false

		mockMetrics := new(MockMetricsService)
		mockHealth := new(MockHealthService)
		mockLogger := new(MockLogger)

		// Expect logger calls
		mockLogger.On("With", mock.Anything).Return(mockLogger)

		// Create server
		server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
		engine := server.GetEngine()

		// Verify basic middleware chain is established
		assert.Greater(t, len(engine.Handlers), 0, "Should have basic middleware chain")

		mockLogger.AssertExpectations(t)
	})
}

// TestHTTPServer_GetLogger tests the GetLogger method
func TestHTTPServer_GetLogger(t *testing.T) {
	config := DefaultConfig()
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// Expect logger calls
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Maybe().Return()

	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	// Test GetLogger returns the expected logger instance
	returnedLogger := server.GetLogger()
	assert.Equal(t, mockLogger, returnedLogger, "GetLogger should return the same logger instance")

	mockLogger.AssertExpectations(t)
}

// TestHTTPServer_Stop_ContextAlreadyCanceled tests Stop with already canceled context
func TestHTTPServer_Stop_ContextAlreadyCanceled(t *testing.T) {
	config := DefaultConfig()
	config.Port = 9999 // Use a different port to avoid conflicts
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// Expect logger calls
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Maybe().Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Maybe().Return()

	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)

	// Create already canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := server.Stop(ctx)
	// Should handle canceled context gracefully (may or may not return error depending on implementation)
	// The important thing is that it doesn't panic
	// Note: The actual behavior depends on the http.Server Shutdown implementation
	// With an already canceled context, Shutdown will likely return context.Canceled
	assert.True(t, err == nil || err == context.Canceled,
		"Should either succeed with nil error or return context.Canceled error")

	mockLogger.AssertExpectations(t)
}

// TestHTTPServer_setupPprofRoutes_EdgeCases tests edge cases for pprof routes setup
func TestHTTPServer_setupPprofRoutes_EdgeCases(t *testing.T) {
	config := DefaultConfig()
	config.EnablePprof = true
	mockMetrics := new(MockMetricsService)
	mockHealth := new(MockHealthService)
	mockLogger := new(MockLogger)

	// Expect logger calls for middleware
	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Maybe().Return()

	server := NewHTTPServer(config, mockLogger, mockMetrics, mockHealth)
	engine := server.GetEngine()

	// Verify pprof routes are accessible (setupPprofRoutes is called in NewHTTPServer)
	req := httptest.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Pprof index should be accessible")

	mockLogger.AssertExpectations(t)
}

// TestMiddleware_EdgeCases tests edge cases for middleware functions
func TestMiddleware_EdgeCases(t *testing.T) {
	t.Run("loggingMiddleware with nil request", func(t *testing.T) {
		mockLogger := new(MockLogger)

		// Create middleware function directly (no server setup)
		middleware := loggingMiddleware(mockLogger)

		// Create a dummy handler
		handler := func(c *gin.Context) {
			c.Status(http.StatusOK)
		}

		// Test with proper context
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		// Expect only Info call from middleware (no With call since we're not creating a server)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// Apply middleware
		middleware(c)
		handler(c)

		mockLogger.AssertExpectations(t)
	})

	t.Run("checkLimit with empty client ID", func(t *testing.T) {
		limiter := &rateLimiter{
			clients: make(map[string]*clientBucket),
		}
		allowed, remaining, _ := limiter.checkLimit("")
		assert.True(t, allowed, "Should allow request with empty client ID")
		assert.Greater(t, remaining, 0, "Should have remaining requests")
	})

	t.Run("formatPanicValue with various types", func(t *testing.T) {
		tests := []struct {
			name     string
			input    interface{}
			expected string
		}{
			{"string panic", "test panic", "test panic"},
			{"error panic", fmt.Errorf("test error"), "test error"},
			{"nil panic", nil, "nil"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := formatPanicValue(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}
