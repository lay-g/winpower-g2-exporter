package server_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
)

// mockMetricsService implements MetricsService interface for testing
type mockMetricsService struct {
	metricsText string
	err         error
}

func (m *mockMetricsService) Render(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.metricsText, nil
}

// mockHealthService implements HealthService interface for testing
type mockHealthService struct {
	status  string
	details map[string]any
}

func (m *mockHealthService) Check(ctx context.Context) (status string, details map[string]any) {
	return m.status, m.details
}

// ExampleNewHTTPServer demonstrates how to create and configure a new HTTP server
func ExampleNewHTTPServer() {
	// Create a logger instance
	logger := log.NewNoopLogger()

	// Create mock services
	metricsService := &mockMetricsService{
		metricsText: `# HELP winpower_exporter_up WinPower Exporter is up
# TYPE winpower_exporter_up gauge
winpower_exporter_up 1

# HELP winpower_devices_connected Number of connected devices
# TYPE winpower_devices_connected gauge
winpower_devices_connected 2`,
	}

	healthService := &mockHealthService{
		status: "ok",
		details: map[string]any{
			"version":    "1.0.0",
			"build_time": "2024-01-01T00:00:00Z",
		},
	}

	// Create server configuration
	config := server.DefaultConfig()
	config.Port = 9090
	config.Host = "0.0.0.0"
	config.Mode = "release"
	config.EnablePprof = false

	// Create HTTP server instance
	_ = server.NewHTTPServer(config, logger, metricsService, healthService)

	// In a real application, you would start the server like this:
	// srv := server.NewHTTPServer(config, logger, metricsService, healthService)
	// if err := srv.Start(); err != nil {
	//     log.Fatal(err)
	// }
	//
	// // Graceful shutdown when application exits
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()
	// if err := srv.Stop(ctx); err != nil {
	//     log.Printf("Server shutdown error: %v", err)
	// }

	fmt.Printf("Server created successfully with config: %s\n", config.String())
	fmt.Printf("Server created successfully\n")

	// Output:
	// Server created successfully with config: ServerConfig{Port: 9090, Host: 0.0.0.0, Mode: release, ReadTimeout: 30s, WriteTimeout: 30s, IdleTimeout: 1m0s, EnablePprof: false, EnableCORS: false, EnableRateLimit: false}
	// Server created successfully
}

// ExampleHTTPServer_customConfiguration demonstrates how to customize server configuration
func ExampleHTTPServer_customConfiguration() {
	logger := log.NewNoopLogger()
	metricsService := &mockMetricsService{
		metricsText: "# HELP test_metric A test metric\ntest_metric 1",
	}
	healthService := &mockHealthService{
		status:  "ok",
		details: map[string]any{},
	}

	// Create custom configuration with non-default values
	config := &server.Config{
		Port:         8080,
		Host:         "127.0.0.1",
		Mode:         "debug",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		EnablePprof:  true,
		EnableCORS:   true,
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
	}

	srv := server.NewHTTPServer(config, logger, metricsService, healthService)
	fmt.Printf("Custom server created with port: %d\n", srv.GetConfig().Port)
	fmt.Printf("Pprof enabled: %t\n", srv.GetConfig().EnablePprof)

	// Output:
	// Custom server created with port: 8080
	// Pprof enabled: true
}

// ExampleHTTPServer_withPprof demonstrates how to enable pprof debugging endpoints
func ExampleHTTPServer_withPprof() {
	logger := log.NewNoopLogger()
	metricsService := &mockMetricsService{
		metricsText: "# HELP example_metric Example\nexample_metric 42",
	}
	healthService := &mockHealthService{
		status:  "ok",
		details: map[string]any{},
	}

	// Enable pprof for debugging
	config := server.DefaultConfig()
	config.EnablePprof = true
	config.Port = 9091

	srv := server.NewHTTPServer(config, logger, metricsService, healthService)

	// Create a test request to verify pprof routes are available
	req := httptest.NewRequest("GET", "/debug/pprof/", nil)
	w := httptest.NewRecorder()
	srv.GetEngine().ServeHTTP(w, req)

	fmt.Printf("Pprof enabled: %t\n", srv.GetConfig().EnablePprof)
	fmt.Printf("Pprof index status: %d\n", w.Code)

	// Output:
	// Pprof enabled: true
	// Pprof index status: 200
}

// ExampleHTTPServer_gracefulShutdown demonstrates how to properly shutdown the server
func ExampleHTTPServer_gracefulShutdown() {
	logger := log.NewNoopLogger()
	metricsService := &mockMetricsService{
		metricsText: "# HELP shutdown_test Testing shutdown\nshutdown_test 1",
	}
	healthService := &mockHealthService{
		status:  "ok",
		details: map[string]any{},
	}

	config := server.DefaultConfig()
	config.Port = 9092 // Use different port to avoid conflicts

	_ = server.NewHTTPServer(config, logger, metricsService, healthService)

	// Simulate graceful shutdown (in real code, this would be done in response to SIGINT/SIGTERM)
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// In a real application, you would start the server first:
	// go func() {
	//     if err := srv.Start(); err != nil && err != http.ErrServerClosed {
	//         log.Printf("Server error: %v", err)
	//     }
	// }()
	//
	// // Wait for shutdown signal
	// sig := <-signals
	// log.Printf("Received signal: %v", sig)
	//
	// // Graceful shutdown
	// if err := srv.Stop(ctx); err != nil {
	//     log.Printf("Shutdown error: %v", err)
	// }

	fmt.Printf("Server configured for graceful shutdown with timeout: %v\n", 5*time.Second)
	fmt.Printf("Default shutdown timeout: %v\n", server.DefaultShutdownTimeout)

	// Output:
	// Server configured for graceful shutdown with timeout: 5s
	// Default shutdown timeout: 30s
}

// ExampleHTTPServer_errorHandling demonstrates how to handle various error scenarios
func ExampleHTTPServer_errorHandling() {
	logger := log.NewNoopLogger()

	// Test with failing metrics service
	metricsService := &mockMetricsService{
		metricsText: "",
		err:         fmt.Errorf("metrics service unavailable"),
	}

	healthService := &mockHealthService{
		status:  "ok",
		details: map[string]any{},
	}

	config := server.DefaultConfig()
	config.Port = 9093

	srv := server.NewHTTPServer(config, logger, metricsService, healthService)

	// Create a test request to metrics endpoint to test error handling
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	srv.GetEngine().ServeHTTP(w, req)

	fmt.Printf("Metrics error handling status: %d\n", w.Code)
	fmt.Printf("Error response includes error field: %t\n",
		w.Body.String() != "" && w.Code == http.StatusServiceUnavailable)

	// Test with failing health service
	failingHealthService := &mockHealthService{
		status: "error",
		details: map[string]any{
			"error": "database connection failed",
		},
	}

	srv2 := server.NewHTTPServer(config, logger, metricsService, failingHealthService)

	// Create a test request to health endpoint
	req2 := httptest.NewRequest("GET", "/health", nil)
	w2 := httptest.NewRecorder()
	srv2.GetEngine().ServeHTTP(w2, req2)

	fmt.Printf("Health error handling status: %d\n", w2.Code)

	// Output:
	// Metrics error handling status: 503
	// Error response includes error field: true
	// Health error handling status: 503
}

// ExampleHTTPServer_integration demonstrates a complete integration example
func ExampleHTTPServer_integration() {
	// This example shows how the server would be used in a real application
	// with proper error handling and graceful shutdown

	logger := log.NewNoopLogger()

	// In a real application, these would be actual service implementations
	metricsService := &mockMetricsService{
		metricsText: `# HELP winpower_devices_total Total number of WinPower devices
# TYPE winpower_devices_total gauge
winpower_devices_total 5

# HELP winpower_uptime_seconds Uptime in seconds
# TYPE winpower_uptime_seconds counter
winpower_uptime_seconds 3600`,
	}

	healthService := &mockHealthService{
		status: "ok",
		details: map[string]any{
			"version":       "1.2.3",
			"build_time":    "2024-06-15T10:30:00Z",
			"git_commit":    "abc123def456",
			"go_version":    "go1.21.0",
			"devices_count": 5,
		},
	}

	// Production-ready configuration
	config := &server.Config{
		Port:         9090,
		Host:         "0.0.0.0",
		Mode:         "release", // Use release mode in production
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		EnablePprof:  false, // Disable pprof in production
		EnableCORS:   false, // Disable CORS unless needed
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Invalid server configuration: %v\n", err)
	}

	// Create server instance
	srv := server.NewHTTPServer(config, logger, metricsService, healthService)

	fmt.Printf("Production server configured:\n")
	fmt.Printf("  - Port: %d\n", srv.GetConfig().Port)
	fmt.Printf("  - Mode: %s\n", srv.GetConfig().Mode)
	fmt.Printf("  - Read Timeout: %v\n", srv.GetConfig().ReadTimeout)
	fmt.Printf("  - Write Timeout: %v\n", srv.GetConfig().WriteTimeout)
	fmt.Printf("  - Pprof Enabled: %t\n", srv.GetConfig().EnablePprof)
	fmt.Printf("  - CORS Enabled: %t\n", srv.GetConfig().EnableCORS)

	// Simulate endpoint testing
	testHealthEndpoint(srv)
	testMetricsEndpoint(srv)
	test404Handling(srv)

	// Output:
	// Production server configured:
	//   - Port: 9090
	//   - Mode: release
	//   - Read Timeout: 30s
	//   - Write Timeout: 30s
	//   - Pprof Enabled: false
	//   - CORS Enabled: false
	// Health endpoint test: 200
	// Metrics endpoint test: 200
	// 404 handling test: 404
}

// testHealthEndpoint tests the health endpoint
func testHealthEndpoint(srv server.Server) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Type assert to access the gin engine
	if httpSrv, ok := srv.(*server.HTTPServer); ok {
		httpSrv.GetEngine().ServeHTTP(w, req)
		fmt.Printf("Health endpoint test: %d\n", w.Code)
	}
}

// testMetricsEndpoint tests the metrics endpoint
func testMetricsEndpoint(srv server.Server) {
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Type assert to access the gin engine
	if httpSrv, ok := srv.(*server.HTTPServer); ok {
		httpSrv.GetEngine().ServeHTTP(w, req)
		fmt.Printf("Metrics endpoint test: %d\n", w.Code)
	}
}

// test404Handling tests 404 error handling
func test404Handling(srv server.Server) {
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	// Type assert to access the gin engine
	if httpSrv, ok := srv.(*server.HTTPServer); ok {
		httpSrv.GetEngine().ServeHTTP(w, req)
		fmt.Printf("404 handling test: %d\n", w.Code)
	}
}
