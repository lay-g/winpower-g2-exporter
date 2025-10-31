// Package mocks provides mock implementations for testing the server module
package mocks

import (
	"context"

	"github.com/gin-gonic/gin"
)

// MetricsService is a mock implementation of server.MetricsService
type MetricsService struct {
	HandleMetricsCalled int
	HandleMetricsFunc   func(c *gin.Context)
}

// HandleMetrics implements server.MetricsService
func (m *MetricsService) HandleMetrics(c *gin.Context) {
	m.HandleMetricsCalled++
	if m.HandleMetricsFunc != nil {
		m.HandleMetricsFunc(c)
		return
	}
	// Default behavior
	c.String(200, "# HELP test_metric Test metric\n# TYPE test_metric gauge\ntest_metric 1\n")
}

// HealthService is a mock implementation of server.HealthService
type HealthService struct {
	CheckCalled int
	CheckFunc   func(ctx context.Context) (string, map[string]any)
	Status      string
	Details     map[string]any
}

// Check implements server.HealthService
func (m *HealthService) Check(ctx context.Context) (string, map[string]any) {
	m.CheckCalled++
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx)
	}
	// Default behavior
	status := m.Status
	if status == "" {
		status = "ok"
	}
	details := m.Details
	if details == nil {
		details = map[string]any{"service": "healthy"}
	}
	return status, details
}

// Logger is a mock implementation of server.Logger
type Logger struct {
	InfoCalled  int
	ErrorCalled int
	WarnCalled  int
	DebugCalled int
	Messages    []string
}

// Info implements server.Logger
func (m *Logger) Info(msg string, keysAndValues ...interface{}) {
	m.InfoCalled++
	m.Messages = append(m.Messages, msg)
}

// Error implements server.Logger
func (m *Logger) Error(msg string, keysAndValues ...interface{}) {
	m.ErrorCalled++
	m.Messages = append(m.Messages, msg)
}

// Warn implements server.Logger
func (m *Logger) Warn(msg string, keysAndValues ...interface{}) {
	m.WarnCalled++
	m.Messages = append(m.Messages, msg)
}

// Debug implements server.Logger
func (m *Logger) Debug(msg string, keysAndValues ...interface{}) {
	m.DebugCalled++
	m.Messages = append(m.Messages, msg)
}
