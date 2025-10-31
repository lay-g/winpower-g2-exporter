package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// HTTPServer implements the Server interface using Gin framework
type HTTPServer struct {
	cfg     *Config
	log     Logger
	engine  *gin.Engine
	srv     *http.Server
	metrics MetricsService
	health  HealthService

	// Server state management
	mu      sync.Mutex
	running bool
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(
	config *Config,
	log Logger,
	metrics MetricsService,
	health HealthService,
) (*HTTPServer, error) {
	// Validate inputs
	if config == nil {
		return nil, ErrInvalidConfig
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrLoggerNil
	}
	if metrics == nil {
		return nil, ErrMetricsServiceNil
	}
	if health == nil {
		return nil, ErrHealthServiceNil
	}

	// Set Gin mode
	gin.SetMode(config.Mode)

	// Create Gin engine without default middleware
	engine := gin.New()

	// Create server instance
	server := &HTTPServer{
		cfg:     config,
		log:     log,
		engine:  engine,
		metrics: metrics,
		health:  health,
		running: false,
	}

	// Setup middleware
	server.setupGlobalMiddleware()

	// Setup routes
	server.setupRoutes()

	// Create http.Server
	server.srv = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      engine,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	log.Info("HTTP server initialized",
		"host", config.Host,
		"port", config.Port,
		"mode", config.Mode,
		"pprof_enabled", config.EnablePprof,
	)

	return server, nil
}

// Start starts the HTTP server
func (s *HTTPServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return ErrServerAlreadyRunning
	}
	s.running = true
	s.mu.Unlock()

	s.log.Info("Starting HTTP server",
		"addr", s.srv.Addr,
	)

	// Start server in a goroutine
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error("HTTP server error",
				"error", err,
			)
		}
	}()

	s.log.Info("HTTP server started successfully",
		"addr", s.srv.Addr,
	)

	return nil
}

// Stop gracefully shuts down the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return ErrServerNotStarted
	}
	s.mu.Unlock()

	s.log.Info("Shutting down HTTP server gracefully",
		"timeout", s.cfg.ShutdownTimeout.String(),
	)

	// Use provided context or create one with shutdown timeout
	shutdownCtx := ctx
	if ctx == nil {
		var cancel context.CancelFunc
		shutdownCtx, cancel = context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
	}

	// Shutdown server
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		s.log.Error("HTTP server shutdown error",
			"error", err,
		)
		return err
	}

	s.mu.Lock()
	s.running = false
	s.mu.Unlock()

	s.log.Info("HTTP server stopped successfully")
	return nil
}

// setupGlobalMiddleware sets up global middleware
func (s *HTTPServer) setupGlobalMiddleware() {
	// Recovery middleware (must be first to catch panics from other middleware)
	s.engine.Use(s.recoveryMiddleware())

	// Logger middleware
	s.engine.Use(s.loggerMiddleware())
}
