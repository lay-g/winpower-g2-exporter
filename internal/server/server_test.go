package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewHTTPServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid configuration", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if srv == nil {
			t.Fatal("Expected server to be created")
		}
		if srv.cfg != cfg {
			t.Error("Expected config to be set")
		}
		if !mockLog.infoCalled {
			t.Error("Expected Info to be called during initialization")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(nil, mockLog, mockMetrics, mockHealth)
		if err != ErrInvalidConfig {
			t.Errorf("Expected ErrInvalidConfig, got %v", err)
		}
		if srv != nil {
			t.Error("Expected nil server")
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &Config{Port: 0} // Invalid port
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != ErrInvalidConfig {
			t.Errorf("Expected ErrInvalidConfig, got %v", err)
		}
		if srv != nil {
			t.Error("Expected nil server")
		}
	})

	t.Run("nil logger", func(t *testing.T) {
		cfg := DefaultConfig()
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, nil, mockMetrics, mockHealth)
		if err != ErrLoggerNil {
			t.Errorf("Expected ErrLoggerNil, got %v", err)
		}
		if srv != nil {
			t.Error("Expected nil server")
		}
	})

	t.Run("nil metrics service", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, nil, mockHealth)
		if err != ErrMetricsServiceNil {
			t.Errorf("Expected ErrMetricsServiceNil, got %v", err)
		}
		if srv != nil {
			t.Error("Expected nil server")
		}
	})

	t.Run("nil health service", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, nil)
		if err != ErrHealthServiceNil {
			t.Errorf("Expected ErrHealthServiceNil, got %v", err)
		}
		if srv != nil {
			t.Error("Expected nil server")
		}
	})
}

func TestHTTPServer_StartStop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("start and stop server", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Port = 18080 // Use a specific test port
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Start server
		err = srv.Start()
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Give server time to start
		time.Sleep(10 * time.Millisecond)

		// Verify server is running
		if !srv.running {
			t.Error("Expected server to be running")
		}

		// Stop server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = srv.Stop(ctx)
		if err != nil {
			t.Errorf("Failed to stop server: %v", err)
		}

		// Verify server is stopped
		if srv.running {
			t.Error("Expected server to be stopped")
		}
	})

	t.Run("start already running server", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Port = 18081 // Use different port
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Start server
		err = srv.Start()
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			srv.Stop(ctx)
		}()

		// Try to start again
		err = srv.Start()
		if err != ErrServerAlreadyRunning {
			t.Errorf("Expected ErrServerAlreadyRunning, got %v", err)
		}
	})

	t.Run("stop not started server", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Try to stop without starting
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = srv.Stop(ctx)
		if err != ErrServerNotStarted {
			t.Errorf("Expected ErrServerNotStarted, got %v", err)
		}
	})

	t.Run("stop with nil context", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Port = 18082 // Use different port
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Start server
		err = srv.Start()
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Give server time to start
		time.Sleep(10 * time.Millisecond)

		// Stop with nil context (should use default timeout)
		err = srv.Stop(nil)
		if err != nil {
			t.Errorf("Failed to stop server: %v", err)
		}
	})
}

func TestHTTPServer_Routes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("health endpoint", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{
			status:  "ok",
			details: map[string]any{"test": "data"},
		}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify response
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if !mockHealth.checkCalled {
			t.Error("Expected Check to be called")
		}
	})

	t.Run("metrics endpoint", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify response
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if !mockMetrics.handleMetricsCalled {
			t.Error("Expected HandleMetrics to be called")
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request to non-existent endpoint
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify response
		if w.Code != 404 {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("pprof endpoints when enabled", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.EnablePprof = true
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Test pprof index
		req := httptest.NewRequest("GET", "/debug/pprof/", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify response
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("pprof disabled by default", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.EnablePprof = false
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Test pprof index (should return 404)
		req := httptest.NewRequest("GET", "/debug/pprof/", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify response
		if w.Code != 404 {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}

func TestHTTPServer_HealthStatuses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		healthStatus   string
		expectedStatus int
	}{
		{
			name:           "healthy status",
			healthStatus:   "ok",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "healthy alternative",
			healthStatus:   "healthy",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unhealthy status",
			healthStatus:   "unhealthy",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "degraded status",
			healthStatus:   "degraded",
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			mockLog := &mockLogger{}
			mockMetrics := &mockMetricsService{}
			mockHealth := &mockHealthService{
				status: tt.healthStatus,
			}

			srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			// Make request
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			srv.engine.ServeHTTP(w, req)

			// Verify response
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
