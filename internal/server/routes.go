package server

import (
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// setupRoutes configures all HTTP routes
func (s *HTTPServer) setupRoutes() {
	// Health check endpoint
	s.engine.GET("/health", s.handleHealth)

	// Metrics endpoint - delegate to metrics service
	s.engine.GET("/metrics", s.metrics.HandleMetrics)

	// 404 handler
	s.engine.NoRoute(s.handleNotFound)

	// Optional pprof endpoints
	if s.cfg.EnablePprof {
		s.setupPprofRoutes()
	}
}

// handleHealth handles health check requests
func (s *HTTPServer) handleHealth(c *gin.Context) {
	status, details := s.health.Check(c.Request.Context())

	// Determine HTTP status code based on health status
	httpStatus := 200
	if status != "ok" && status != "healthy" {
		httpStatus = 503
	}

	// Build response
	response := map[string]any{
		"status":  status,
		"details": details,
	}

	c.JSON(httpStatus, response)
}

// handleNotFound handles 404 errors
func (s *HTTPServer) handleNotFound(c *gin.Context) {
	c.JSON(404, NewErrorResponse(
		ErrServerNotStarted, // placeholder error
		c.Request.URL.Path,
	))
}

// setupPprofRoutes sets up pprof profiling routes
func (s *HTTPServer) setupPprofRoutes() {
	pprofGroup := s.engine.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	s.log.Info("Pprof endpoints enabled", "prefix", "/debug/pprof")
}
