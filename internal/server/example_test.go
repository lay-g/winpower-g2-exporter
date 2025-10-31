package server_test

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"go.uber.org/zap"
)

// ExampleNewHTTPServer demonstrates how to create and start an HTTP server
func ExampleNewHTTPServer() {
	// Create configuration
	cfg := server.DefaultConfig()
	cfg.Port = 8080
	cfg.Mode = "release"

	// Create logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Create mock services for demonstration
	metricsService := &mockMetrics{}
	healthService := &mockHealth{}

	// Create server
	srv, err := server.NewHTTPServer(cfg, &zapLoggerAdapter{logger}, metricsService, healthService)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// Start server
	if err := srv.Start(); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	// Stop server gracefully
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	fmt.Println("Server lifecycle completed")
	// Output: Server lifecycle completed
}

// ExampleConfig demonstrates configuration usage
func ExampleConfig() {
	// Create default configuration
	cfg := server.DefaultConfig()
	fmt.Printf("Default port: %d\n", cfg.Port)
	fmt.Printf("Default mode: %s\n", cfg.Mode)

	// Customize configuration
	cfg.Port = 9090
	cfg.EnablePprof = true

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
	} else {
		fmt.Println("Configuration is valid")
	}

	// Output:
	// Default port: 8080
	// Default mode: release
	// Configuration is valid
}

// Mock implementations for examples

type mockMetrics struct{}

func (m *mockMetrics) HandleMetrics(c *gin.Context) {
	c.String(200, "# Example metrics\n")
}

type mockHealth struct{}

func (m *mockHealth) Check(ctx context.Context) (string, map[string]any) {
	return "ok", map[string]any{"service": "healthy"}
}

// Adapter to convert zap.Logger to server.Logger interface
type zapLoggerAdapter struct {
	logger *zap.Logger
}

func (z *zapLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Info(msg, toZapFields(keysAndValues)...)
}

func (z *zapLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Error(msg, toZapFields(keysAndValues)...)
}

func (z *zapLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Warn(msg, toZapFields(keysAndValues)...)
}

func (z *zapLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Debug(msg, toZapFields(keysAndValues)...)
}

func toZapFields(keysAndValues []interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, keysAndValues[i+1]))
	}
	return fields
}
