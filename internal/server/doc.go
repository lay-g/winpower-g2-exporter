// Package server implements HTTP server for WinPower G2 Prometheus Exporter.
//
// The server module provides minimal and stable web interfaces:
//   - GET /health: Health check endpoint returning JSON status
//   - GET /metrics: Prometheus metrics endpoint returning text format
//
// Key features:
//   - High performance HTTP server based on Gin framework
//   - Comprehensive middleware chain (logging, recovery, optional CORS/rate limiting)
//   - Optional /debug/pprof endpoints for diagnostics
//   - Graceful startup and shutdown with timeout control
//   - Structured logging and error handling
//
// Architecture:
//   - Single responsibility: Only handles HTTP layer routing and middleware
//   - Dependency injection: Receives Logger, MetricsService, and HealthService
//   - Configuration: Server-specific config with validation
//   - Extensible: Supports custom middleware through registration mechanism
//
// Usage:
//
//	config := server.DefaultConfig()
//	config.Port = 9090
//	config.EnablePprof = false
//
//	srv := server.NewHTTPServer(config, logger, metricsService, healthService)
//
//	// Start server
//	if err := srv.Start(); err != nil {
//		log.Fatal(err)
//	}
//
//	// Graceful shutdown
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	if err := srv.Stop(ctx); err != nil {
//		log.Printf("Server shutdown error: %v", err)
//	}
package server
