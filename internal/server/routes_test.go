package server

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handleHealth returns correct response", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{
			status:  "ok",
			details: map[string]any{"uptime": "1h"},
		}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		if !mockHealth.checkCalled {
			t.Error("Expected Check to be called")
		}
	})

	t.Run("handleHealth with custom check function", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}

		called := false
		mockHealth := &mockHealthService{
			checkFunc: func(ctx context.Context) (string, map[string]any) {
				called = true
				return "custom", map[string]any{"custom": "data"}
			},
		}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify custom function was called
		if !called {
			t.Error("Expected custom check function to be called")
		}
	})

	t.Run("handleNotFound returns 404", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request to non-existent path
		req := httptest.NewRequest("GET", "/invalid", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify
		if w.Code != 404 {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("metrics endpoint delegates to service", func(t *testing.T) {
		cfg := DefaultConfig()
		mockLog := &mockLogger{}

		customHandlerCalled := false
		mockMetrics := &mockMetricsService{
			handleMetricsFunc: func(c *gin.Context) {
				customHandlerCalled = true
				c.String(200, "custom metrics")
			},
		}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Make request
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		srv.engine.ServeHTTP(w, req)

		// Verify
		if !customHandlerCalled {
			t.Error("Expected custom metrics handler to be called")
		}
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("setupPprofRoutes creates pprof endpoints", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.EnablePprof = true
		mockLog := &mockLogger{}
		mockMetrics := &mockMetricsService{}
		mockHealth := &mockHealthService{}

		srv, err := NewHTTPServer(cfg, mockLog, mockMetrics, mockHealth)
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		pprofEndpoints := []string{
			"/debug/pprof/",
			"/debug/pprof/cmdline",
			"/debug/pprof/profile",
			"/debug/pprof/symbol",
			"/debug/pprof/trace",
			"/debug/pprof/allocs",
			"/debug/pprof/block",
			"/debug/pprof/goroutine",
			"/debug/pprof/heap",
			"/debug/pprof/mutex",
			"/debug/pprof/threadcreate",
		}

		for _, endpoint := range pprofEndpoints {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()
			srv.engine.ServeHTTP(w, req)

			// Should not return 404
			if w.Code == 404 {
				t.Errorf("Endpoint %s returned 404", endpoint)
			}
		}
	})
}
