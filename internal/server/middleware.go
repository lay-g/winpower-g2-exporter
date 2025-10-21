package server

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// Rate limiter implementation
type rateLimiter struct {
	clients map[string]*clientBucket
	mutex   sync.RWMutex
}

type clientBucket struct {
	requests  int
	lastReset time.Time
}

// Global rate limiter instance
var globalRateLimiter = &rateLimiter{
	clients: make(map[string]*clientBucket),
}

// Constants for rate limiting
const (
	// RateLimitRequests defines the maximum number of requests per window
	RateLimitRequests = 100
	// RateLimitWindow defines the time window for rate limiting
	RateLimitWindow = 1 * time.Minute
	// RateLimitCleanupInterval defines how often to clean up old client entries
	RateLimitCleanupInterval = 5 * time.Minute
)

// loggingMiddleware creates a middleware that logs HTTP requests
func loggingMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Prepare log fields
		fields := []log.Field{
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path),
			log.String("query", c.Request.URL.RawQuery),
			log.Int("status", c.Writer.Status()),
			log.Int("status_code", c.Writer.Status()),
			log.String("duration", duration.String()),
			log.String("remote_addr", c.Request.RemoteAddr),
			log.String("user_agent", c.Request.Header.Get("User-Agent")),
			log.String("referer", c.Request.Header.Get("Referer")),
		}

		// Add query parameters as structured data
		if len(c.Request.URL.Query()) > 0 {
			queryParams := make(map[string]interface{})
			for key, values := range c.Request.URL.Query() {
				if len(values) == 1 {
					queryParams[key] = values[0]
				} else {
					queryParams[key] = values
				}
			}
			fields = append(fields, log.Any("query_params", queryParams))
		}

		// Add request ID if available
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			fields = append(fields, log.String("request_id", requestID))
		}

		// Add client IP (extracted from headers if available)
		if clientIP := c.ClientIP(); clientIP != "" {
			fields = append(fields, log.String("client_ip", clientIP))
		}

		// Log based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("HTTP request completed with server error", fields...)
		case status >= 400:
			logger.Warn("HTTP request completed with client error", fields...)
		case status >= 300:
			logger.Info("HTTP request completed with redirect", fields...)
		default:
			logger.Info("HTTP request completed", fields...)
		}
	}
}

// recoveryMiddleware creates a middleware that recovers from panics
func recoveryMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				logger.Error("Panic recovered",
					log.String("panic", formatPanicValue(err)),
					log.String("path", c.Request.URL.Path),
					log.String("method", c.Request.Method),
					log.String("remote_addr", c.Request.RemoteAddr),
					log.String("user_agent", c.Request.Header.Get("User-Agent")),
					log.String("client_ip", c.ClientIP()),
					log.String("stack_trace", string(debug.Stack())),
				)

				// Prepare error response
				errorResponse := ErrorResponse{
					Error:     "internal_server_error",
					Message:   "Internal server error",
					Path:      c.Request.URL.Path,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
				}

				// Set headers and send error response
				c.Header("Content-Type", "application/json; charset=utf-8")
				c.JSON(500, errorResponse)

				// Abort the request processing
				c.Abort()
			}
		}()

		c.Next()
	}
}

// corsMiddleware creates a middleware that handles CORS
func corsMiddleware(logger log.Logger, config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If CORS is disabled, skip processing
		if !config.EnableCORS {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")

		// If no Origin header, this is not a cross-origin request
		if origin == "" {
			c.Next()
			return
		}

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With,X-Request-ID,X-Custom-Header")
		c.Header("Access-Control-Expose-Headers", "X-Total-Count,X-Request-ID")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			// Get requested method and headers
			requestedMethod := c.Request.Header.Get("Access-Control-Request-Method")
			requestedHeaders := c.Request.Header.Get("Access-Control-Request-Headers")

			// Set specific headers for preflight
			if requestedMethod != "" {
				c.Header("Access-Control-Allow-Methods", requestedMethod+",GET,POST,PUT,DELETE,OPTIONS,PATCH")
			}
			if requestedHeaders != "" {
				c.Header("Access-Control-Allow-Headers", requestedHeaders+",Content-Type,Authorization,X-Requested-With,X-Request-ID")
			}

			// Set cache duration for preflight requests
			c.Header("Access-Control-Max-Age", "86400") // 24 hours

			// Log CORS preflight
			logger.Debug("CORS preflight request processed",
				log.String("origin", origin),
				log.String("method", c.Request.Method),
				log.String("requested_method", requestedMethod),
				log.String("path", c.Request.URL.Path),
			)

			// Return 204 No Content for preflight
			c.AbortWithStatus(204)
			return
		}

		// Log CORS request
		logger.Debug("CORS request processed",
			log.String("origin", origin),
			log.String("method", c.Request.Method),
			log.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

// rateLimitMiddleware creates a middleware that implements rate limiting
func rateLimitMiddleware(logger log.Logger, config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If rate limiting is disabled, skip processing
		if !config.EnableRateLimit {
			c.Next()
			return
		}

		// Get client identifier (IP address)
		clientIP := c.ClientIP()
		if clientIP == "" {
			// Fallback to remote address if ClientIP fails
			clientIP = strings.Split(c.Request.RemoteAddr, ":")[0]
		}

		// Check rate limit
		allowed, remaining, resetTime := globalRateLimiter.checkLimit(clientIP)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", RateLimitRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		if !allowed {
			// Log rate limit exceeded
			logger.Warn("Rate limit exceeded",
				log.String("client_ip", clientIP),
				log.String("method", c.Request.Method),
				log.String("path", c.Request.URL.Path),
				log.String("user_agent", c.Request.Header.Get("User-Agent")),
				log.Int("limit", RateLimitRequests),
				log.Int("remaining", remaining),
			)

			// Prepare rate limit response
			errorResponse := ErrorResponse{
				Error:     "rate_limit_exceeded",
				Message:   "Rate limit exceeded",
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}

			// Set headers and send rate limit response
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("Retry-After", "60") // Suggest retry after 60 seconds
			c.JSON(429, errorResponse)

			// Abort the request processing
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkLimit checks if the client has exceeded the rate limit
func (rl *rateLimiter) checkLimit(clientID string) (bool, int, time.Time) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Get or create client bucket
	bucket, exists := rl.clients[clientID]
	if !exists {
		bucket = &clientBucket{
			requests:  0,
			lastReset: now,
		}
		rl.clients[clientID] = bucket
	}

	// Reset if window has expired
	if now.Sub(bucket.lastReset) >= RateLimitWindow {
		bucket.requests = 0
		bucket.lastReset = now
	}

	// Check if under limit
	if bucket.requests < RateLimitRequests {
		bucket.requests++
		remaining := RateLimitRequests - bucket.requests
		resetTime := bucket.lastReset.Add(RateLimitWindow)
		return true, remaining, resetTime
	}

	// Rate limit exceeded
	resetTime := bucket.lastReset.Add(RateLimitWindow)
	return false, 0, resetTime
}

// formatPanicValue formats a panic value for logging
func formatPanicValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprintf("%v", value)
	}
}
