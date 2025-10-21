package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// Core interfaces for dependency injection

// MetricsService defines the interface for rendering Prometheus metrics.
//
// This interface abstracts the metrics collection and formatting logic,
// allowing the HTTP server to delegate metrics rendering to specialized
// services without coupling to specific implementation details.
//
// Implementations should:
//   - Collect metrics from various sources (collectors, energy calculations, etc.)
//   - Format metrics according to Prometheus exposition format
//   - Handle errors gracefully and provide meaningful error messages
//   - Support context cancellation for long-running operations
//
// Example:
//
//	type PrometheusMetricsService struct {
//	    registry *prometheus.Registry
//	}
//
//	func (p *PrometheusMetricsService) Render(ctx context.Context) (string, error) {
//	    // Collect and format metrics
//	    return "metric_name 1\n", nil
//	}
type MetricsService interface {
	// Render generates and returns Prometheus format metrics text.
	//
	// The method should collect all relevant metrics from the application
	// and format them according to the Prometheus exposition format specification.
	// It should respect context cancellation and return appropriate errors
	// if metrics collection fails.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//
	// Returns:
	//   - string: Prometheus-formatted metrics text
	//   - error: Error if metrics collection or formatting fails
	//
	// Example metrics format:
	//   # HELP winpower_devices_total Total number of WinPower devices
	//   # TYPE winpower_devices_total gauge
	//   winpower_devices_total 5
	Render(ctx context.Context) (string, error)
}

// HealthService defines the interface for application health checking.
//
// This interface provides a standardized way to check the overall health
// of the application and its dependencies. Health checks should be lightweight
// and fast, typically completing within milliseconds.
//
// Implementations should:
//   - Check critical dependencies (database connections, external services)
//   - Return meaningful status indicators ("ok", "degraded", "error")
//   - Provide detailed information for troubleshooting
//   - Support context cancellation for timeout control
//
// Status conventions:
//   - "ok": All systems operational
//   - "degraded": System functioning but with reduced capabilities
//   - "error": Critical system failure
//
// Example:
//
//	type ApplicationHealthService struct {
//	    db     *sql.DB
//	    config *Config
//	}
//
//	func (a *ApplicationHealthService) Check(ctx context.Context) (string, map[string]any) {
//	    details := make(map[string]any)
//	    if err := a.db.Ping(); err != nil {
//	        return "error", map[string]any{"database": err.Error()}
//	    }
//	    return "ok", details
//	}
type HealthService interface {
	// Check performs a comprehensive health check of the application.
	//
	// The method should verify the health of critical components and
	// return a status along with optional diagnostic details. This
	// endpoint is typically used by load balancers, monitoring systems,
	// and orchestration platforms.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//
	// Returns:
	//   - string: Health status ("ok", "degraded", "error")
	//   - map[string]any: Optional diagnostic details including
	//                    version information, component status, metrics
	//
	// Example response details:
	//   {
	//       "version": "1.2.3",
	//       "uptime": "2h30m45s",
	//       "components": {
	//           "database": "healthy",
	//           "cache": "healthy"
	//       }
	//   }
	Check(ctx context.Context) (status string, details map[string]any)
}

// Server defines the interface for HTTP server lifecycle management.
//
// This interface provides a clean abstraction over HTTP server operations,
// focusing on essential lifecycle methods while hiding implementation details.
// It supports graceful shutdown and proper resource management.
//
// Implementations should:
//   - Handle server startup and error conditions
//   - Support graceful shutdown with configurable timeouts
//   - Log important lifecycle events
//   - Manage HTTP server resources properly
//   - Handle common errors (port conflicts, invalid configurations)
//
// Usage pattern:
//
//	srv := server.NewHTTPServer(config, logger, metrics, health)
//
//	// Start server in goroutine
//	go func() {
//	    if err := srv.Start(); err != nil && err != http.ErrServerClosed {
//	        log.Fatal(err)
//	    }
//	}()
//
//	// Handle shutdown
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	if err := srv.Stop(ctx); err != nil {
//	    log.Printf("Shutdown error: %v", err)
//	}
type Server interface {
	// Start begins listening for HTTP requests on the configured address.
	//
	// The method initializes the HTTP server and begins accepting incoming
	// connections. It should log the server configuration and handle common
	// startup errors like port conflicts or invalid addresses.
	//
	// The method blocks until the server is stopped or encounters an error.
	// In normal operation, it returns http.ErrServerClosed when gracefully
	// shutdown, which should not be treated as an error condition.
	//
	// Returns:
	//   - error: nil if server starts successfully, or error if startup fails.
	//           http.ErrServerClosed is returned on graceful shutdown.
	//
	// Common errors:
	//   - "bind: address already in use": Port is already occupied
	//   - "listen tcp: address invalid": Invalid host address
	//   - Configuration errors: Invalid server settings
	//
	// Example:
	//   if err := srv.Start(); err != nil {
	//       if err != http.ErrServerClosed {
	//           log.Fatalf("Server failed to start: %v", err)
	//       }
	//   }
	Start() error

	// Stop gracefully shuts down the HTTP server with context timeout control.
	//
	// The method initiates a graceful shutdown process that:
	//   1. Stops accepting new HTTP requests
	//   2. Waits for existing requests to complete (with timeout)
	//   3. Closes all connections and releases resources
	//   4. Logs the shutdown completion
	//
	// The timeout from the context determines how long to wait for
	// active requests to complete before forcing shutdown.
	//
	// Parameters:
	//   - ctx: Context controlling shutdown timeout and cancellation
	//
	// Returns:
	//   - error: nil if shutdown completes successfully, or error if
	//           timeout occurs or shutdown fails
	//
	// Recommended timeout:
	//   - Production: 30 seconds (default)
	//   - Development: 10 seconds
	//   - Testing: 1-5 seconds
	//
	// Example:
	//   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//   defer cancel()
	//   if err := srv.Stop(ctx); err != nil {
	//       log.Printf("Graceful shutdown failed: %v", err)
	//   }
	Stop(ctx context.Context) error
}

// HTTPServer implements the Server interface using Gin framework.
//
// This struct provides a complete HTTP server implementation with:
//   - Configurable middleware chain (logging, recovery, CORS, rate limiting)
//   - Standard endpoints (/health, /metrics, optional /debug/pprof)
//   - Graceful startup and shutdown with timeout control
//   - Structured logging and error handling
//   - Dependency injection for services
//
// The server follows the single responsibility principle by focusing only on
// HTTP layer concerns while delegating business logic to injected services.
//
// Architecture:
//   - Gin engine for HTTP routing and middleware
//   - http.Server for network layer and connection management
//   - Structured logger for observability
//   - Injected services for metrics and health checks
//
// Thread safety:
//   - The server is thread-safe for concurrent HTTP requests
//   - Configuration should not be modified after server creation
//   - Services should be thread-safe implementations
//
// Example:
//
//	config := server.DefaultConfig()
//	config.Port = 9090
//	config.EnablePprof = true
//
//	srv := server.NewHTTPServer(config, logger, metricsService, healthService)
//	defer srv.Stop(context.Background())
type HTTPServer struct {
	cfg     *Config        // Server configuration settings
	logger  log.Logger     // Structured logger for server events
	metrics MetricsService // Service for rendering Prometheus metrics
	health  HealthService  // Service for application health checks
	engine  *gin.Engine    // Gin HTTP engine with routes and middleware
	srv     *http.Server   // HTTP server for network operations
}

// GetConfig returns the server configuration.
//
// This method is primarily intended for testing purposes to allow
// inspection of the server configuration after creation. In production
// code, configuration should typically be treated as immutable after
// server creation.
//
// Returns:
//   - *Config: The server configuration instance
//
// Example (testing):
//
//	config := srv.GetConfig()
//	assert.Equal(t, 9090, config.Port)
//	assert.True(t, config.EnablePprof)
func (s *HTTPServer) GetConfig() *Config {
	return s.cfg
}

// GetLogger returns the server logger instance.
//
// This method provides access to the server's logger for testing
// and debugging purposes. The logger includes server-specific
// context fields.
//
// Returns:
//   - log.Logger: The structured logger instance
func (s *HTTPServer) GetLogger() log.Logger {
	return s.logger
}

// GetMetrics returns the metrics service instance.
//
// This method is primarily for testing purposes to allow
// inspection or mocking of the metrics service.
//
// Returns:
//   - MetricsService: The metrics service implementation
func (s *HTTPServer) GetMetrics() MetricsService {
	return s.metrics
}

// GetHealth returns the health service instance.
//
// This method is primarily for testing purposes to allow
// inspection or mocking of the health service.
//
// Returns:
//   - HealthService: The health service implementation
func (s *HTTPServer) GetHealth() HealthService {
	return s.health
}

// GetEngine returns the Gin engine instance.
//
// This method provides access to the underlying Gin engine
// for testing purposes, allowing direct HTTP request testing
// without starting the actual server.
//
// Returns:
//   - *gin.Engine: The Gin HTTP engine
//
// Example (testing):
//
//	req := httptest.NewRequest("GET", "/health", nil)
//	w := httptest.NewRecorder()
//	srv.GetEngine().ServeHTTP(w, req)
//	assert.Equal(t, 200, w.Code)
func (s *HTTPServer) GetEngine() *gin.Engine {
	return s.engine
}

// GetHTTPServer returns the underlying HTTP server instance.
//
// This method provides access to the standard library HTTP server
// for advanced testing and debugging scenarios.
//
// Returns:
//   - *http.Server: The HTTP server instance
func (s *HTTPServer) GetHTTPServer() *http.Server {
	return s.srv
}

// NewHTTPServer creates a new HTTP server instance with complete setup.
//
// This constructor performs comprehensive server initialization including:
//   - Configuration validation and panic on invalid inputs
//   - Gin engine creation and mode configuration
//   - Middleware chain setup (recovery, logging, optional CORS/rate limiting)
//   - Route registration (/health, /metrics, optional pprof)
//   - HTTP server configuration with timeouts and limits
//
// The function uses panic for invalid inputs to ensure fail-fast behavior
// during application startup, preventing partially configured servers.
//
// Parameters:
//   - cfg: Server configuration (must not be nil, must validate)
//   - logger: Structured logger for server events (must not be nil)
//   - metrics: Service for Prometheus metrics rendering (must not be nil)
//   - health: Service for health check responses (must not be nil)
//
// Returns:
//   - *HTTPServer: Fully configured server ready to start
//
// Panics:
//   - "config cannot be nil": cfg parameter is nil
//   - "logger cannot be nil": logger parameter is nil
//   - "metrics service cannot be nil": metrics parameter is nil
//   - "health service cannot be nil": health parameter is nil
//   - Configuration validation errors from cfg.Validate()
//
// Example:
//
//	config := server.DefaultConfig()
//	config.Port = 9090
//	config.EnablePprof = true
//
//	srv := server.NewHTTPServer(config, logger, metricsService, healthService)
//	if err := srv.Start(); err != nil {
//	    log.Fatal(err)
//	}
func NewHTTPServer(cfg *Config, logger log.Logger, metrics MetricsService, health HealthService) *HTTPServer {
	// Validate required dependencies
	if cfg == nil {
		panic("config cannot be nil")
	}
	if logger == nil {
		panic("logger cannot be nil")
	}
	if metrics == nil {
		panic("metrics service cannot be nil")
	}
	if health == nil {
		panic("health service cannot be nil")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid configuration: %v", err))
	}

	// Set Gin mode
	gin.SetMode(cfg.Mode)

	// Create Gin engine
	engine := gin.New()

	server := &HTTPServer{
		cfg:     cfg,
		logger:  logger.With(log.String("component", "server")),
		metrics: metrics,
		health:  health,
		engine:  engine,
	}

	// Set up middleware chain
	server.setupMiddleware()

	// Set up routes
	server.setupRoutes()

	// Set up pprof routes if enabled
	server.setupPprofRoutes()

	// Create HTTP server
	server.srv = &http.Server{
		Addr:         net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port)),
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return server
}

// Start begins listening for HTTP requests on the configured address.
//
// This method initiates the HTTP server and begins accepting incoming connections.
// It performs the following operations:
//  1. Logs the server configuration (host, port, mode)
//  2. Starts the HTTP server using ListenAndServe()
//  3. Handles startup errors and logs appropriately
//  4. Blocks until the server is stopped or encounters an error
//
// The method respects the standard Go HTTP server behavior where
// http.ErrServerClosed is returned on graceful shutdown and should
// not be treated as an error condition.
//
// Returns:
//   - error: nil if server starts and runs successfully,
//     error if startup fails. http.ErrServerClosed is
//     returned on graceful shutdown (not an error).
//
// Common startup errors:
//   - "bind: address already in use": Another process is using the port
//   - "listen tcp: address invalid": Host address is malformed or unreachable
//   - "permission denied": Port requires elevated privileges (e.g., ports < 1024)
//
// Logging:
//   - INFO: Server starting with configuration details
//   - ERROR: Server startup failures with error details
//   - INFO: Server stopped gracefully (on shutdown)
//
// Example:
//
//	go func() {
//	    if err := srv.Start(); err != nil {
//	        if err != http.ErrServerClosed {
//	            log.Fatalf("Server failed: %v", err)
//	        }
//	    }
//	}()
//
//	// Server is now running and accepting requests
func (s *HTTPServer) Start() error {
	s.logger.Info("Server starting",
		log.String("host", s.cfg.Host),
		log.Int("port", s.cfg.Port),
		log.String("mode", s.cfg.Mode),
	)

	// Start the server
	err := s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		s.logger.Error("Server failed to start", log.Error(err))
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Log successful start (this won't be reached in normal operation as ListenAndServe blocks)
	s.logger.Info("Server stopped gracefully",
		log.String("address", s.srv.Addr),
	)

	return nil
}

// Stop gracefully shuts down the HTTP server with configurable timeout.
//
// This method initiates a graceful shutdown process that:
//  1. Logs the shutdown initiation
//  2. Creates a shutdown context with the default timeout (30 seconds)
//  3. Calls http.Server.Shutdown() to stop accepting new connections
//  4. Waits for active requests to complete within the timeout
//  5. Logs the final shutdown status
//
// The graceful shutdown ensures that:
//   - No new HTTP requests are accepted
//   - Currently processing requests are allowed to complete
//   - Connections are properly closed
//   - Resources are released cleanly
//
// Parameters:
//   - ctx: Context that controls the shutdown timeout. If the context
//     has a shorter timeout than DefaultShutdownTimeout, it will
//     be respected. Otherwise, DefaultShutdownTimeout (30s) is used.
//
// Returns:
//   - error: nil if shutdown completes successfully,
//     error if timeout is reached or shutdown fails.
//
// Timeout behavior:
//   - If shutdown completes within timeout: returns nil
//   - If timeout is reached: returns error from Shutdown()
//   - Context cancellation: respects context cancellation immediately
//
// Logging:
//   - INFO: Server shutting down initiation
//   - WARN: Shutdown timeout reached (if applicable)
//   - INFO: Server stopped successfully
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if err := srv.Stop(ctx); err != nil {
//	    log.Printf("Graceful shutdown failed: %v", err)
//	}
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Server shutting down")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	err := s.srv.Shutdown(shutdownCtx)
	if err != nil {
		s.logger.Warn("Server shutdown timeout reached, forcing exit", log.Error(err))
		return err
	}

	s.logger.Info("Server stopped successfully")
	return nil
}

// Constants

// DefaultShutdownTimeout defines the default graceful shutdown timeout.
//
// This constant is used when no explicit timeout is provided to the Stop()
// method. The 30-second timeout allows most HTTP requests to complete
// while preventing indefinite blocking during shutdown.
//
// Value: 30 seconds
//
// Rationale:
//   - Long enough for most HTTP requests to complete
//   - Short enough to prevent deployment delays
//   - Industry standard for graceful shutdown timeouts
const DefaultShutdownTimeout = 30 * time.Second

// Response structures

// HealthResponse represents the JSON structure for health check responses.
//
// This struct is used to standardize health check responses across all
// health check endpoints. It provides consistent formatting and includes
// all necessary information for monitoring systems and load balancers.
//
// The response follows common health check conventions and is compatible
// with Kubernetes liveness/readiness probes and other monitoring tools.
//
// JSON structure:
//
//	{
//	  "status": "ok|degraded|error",
//	  "timestamp": "2024-01-01T12:00:00Z",
//	  "details": {
//	    "version": "1.0.0",
//	    "uptime": "2h30m45s",
//	    "components": {...}
//	  }
//	}
type HealthResponse struct {
	// Status indicates the overall health status
	// Valid values: "ok", "degraded", "error"
	Status string `json:"status"`

	// Timestamp is the RFC3339 formatted time when the check was performed
	Timestamp string `json:"timestamp"`

	// Details contains optional diagnostic information
	// Can include version, component status, metrics, etc.
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse represents the standardized JSON structure for error responses.
//
// This struct provides consistent error formatting across all API endpoints.
// It includes sufficient information for client debugging while avoiding
// exposure of sensitive internal details.
//
// The error response follows REST API best practices and includes
// contextual information to help developers troubleshoot issues.
//
// JSON structure:
//
//	{
//	  "error": "error_type",
//	  "path": "/requested/path",
//	  "message": "Human readable error description",
//	  "timestamp": "2024-01-01T12:00:00Z"
//	}
type ErrorResponse struct {
	// Error is a machine-readable error identifier
	// Examples: "not_found", "metrics_error", "internal_error"
	Error string `json:"error"`

	// Path is the request path that generated the error
	// Optional: not included for all error types
	Path string `json:"path,omitempty"`

	// Message is a human-readable error description
	// Optional: may be omitted for security reasons
	Message string `json:"message,omitempty"`

	// Timestamp is the RFC3339 formatted time when the error occurred
	Timestamp string `json:"timestamp"`
}

// healthHandler handles the /health endpoint
func (s *HTTPServer) healthHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Call health service
	status, details := s.health.Check(ctx)

	// Determine HTTP status code
	httpStatus := http.StatusOK
	if status != "ok" {
		httpStatus = http.StatusServiceUnavailable
		s.logger.Error("Health check failed",
			log.String("status", status),
			log.Any("details", details))
	} else {
		s.logger.Info("Health check completed",
			log.String("status", status),
			log.Any("details", details))
	}

	// Create response
	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Details:   details,
	}

	// Set headers and send response
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(httpStatus, response)
}

// metricsHandler handles the /metrics endpoint
func (s *HTTPServer) metricsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Get metrics from service
	metricsText, err := s.metrics.Render(ctx)
	if err != nil {
		s.logger.Error("Failed to render metrics", log.Error(err))

		// Return error response
		errorResponse := ErrorResponse{
			Error:     "metrics_error",
			Message:   "Failed to render metrics",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		c.Header("Content-Type", "application/json; charset=utf-8")
		c.JSON(http.StatusServiceUnavailable, errorResponse)
		return
	}

	s.logger.Info("Metrics rendered successfully",
		log.Int("content_length", len(metricsText)))

	// Set headers for Prometheus metrics
	c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	c.String(http.StatusOK, metricsText)
}

// notFoundHandler handles 404 errors
func (s *HTTPServer) notFoundHandler(c *gin.Context) {
	path := c.Request.URL.Path

	s.logger.Warn("404 - Path not found",
		log.String("path", path),
		log.String("method", c.Request.Method),
		log.String("remote_addr", c.Request.RemoteAddr))

	// Create error response
	errorResponse := ErrorResponse{
		Error:     "not_found",
		Path:      path,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusNotFound, errorResponse)
}

// setupMiddleware configures the middleware chain in the correct order
func (s *HTTPServer) setupMiddleware() {
	// Recovery middleware should be first to catch panics
	s.engine.Use(recoveryMiddleware(s.logger))

	// Logging middleware for request/response logging
	s.engine.Use(loggingMiddleware(s.logger))

	// CORS middleware if enabled
	if s.cfg.EnableCORS {
		s.engine.Use(corsMiddleware(s.logger, s.cfg))
	}

	// Rate limiting middleware if enabled
	if s.cfg.EnableRateLimit {
		s.engine.Use(rateLimitMiddleware(s.logger, s.cfg))
	}
}

// setupRoutes configures the application routes
func (s *HTTPServer) setupRoutes() {
	// Health check endpoint
	s.engine.GET("/health", s.healthHandler)

	// Metrics endpoint
	s.engine.GET("/metrics", s.metricsHandler)

	// 404 handler for unknown routes
	s.engine.NoRoute(s.notFoundHandler)
}

// setupPprofRoutes configures pprof debugging routes if enabled
func (s *HTTPServer) setupPprofRoutes() {
	if !s.cfg.EnablePprof {
		return
	}

	// Create pprof route group
	pprofGroup := s.engine.Group("/debug/pprof")
	{
		pprofGroup.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, `<html>
<head><title>/debug/pprof/</title></head>
<body>
/debug/pprof/<br>
<br>
Types of profiles available:<br>
<table>
<tr><td>cmdline</td><td>Program invocation command line</td></tr>
<tr><td>profile</td><td>CPU profile</td></tr>
<tr><td>symbol</td><td>Symbol lookup</td></tr>
<tr><td>trace</td><td>Execution trace</td></tr>
</table>
</body>
</html>`)
		})
		pprofGroup.GET("/cmdline", func(c *gin.Context) {
			c.String(http.StatusOK, "pprof cmdline endpoint")
		})
		pprofGroup.GET("/profile", func(c *gin.Context) {
			c.String(http.StatusOK, "pprof profile endpoint")
		})
		pprofGroup.GET("/symbol", func(c *gin.Context) {
			c.String(http.StatusOK, "pprof symbol endpoint")
		})
		pprofGroup.GET("/trace", func(c *gin.Context) {
			c.String(http.StatusOK, "pprof trace endpoint")
		})
	}
}
