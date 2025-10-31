package server

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
)

// Verify that interfaces are correctly defined
func TestInterfaces(t *testing.T) {
	t.Run("Server interface", func(t *testing.T) {
		var _ Server = (*HTTPServer)(nil)
	})

	t.Run("MetricsService interface", func(t *testing.T) {
		var _ MetricsService = (*mockMetricsService)(nil)
	})

	t.Run("HealthService interface", func(t *testing.T) {
		var _ HealthService = (*mockHealthService)(nil)
	})

	t.Run("Logger interface", func(t *testing.T) {
		var _ Logger = (*mockLogger)(nil)
	})
}

// Mock implementations for testing

type mockMetricsService struct {
	handleMetricsCalled bool
	handleMetricsFunc   func(c *gin.Context)
}

func (m *mockMetricsService) HandleMetrics(c *gin.Context) {
	m.handleMetricsCalled = true
	if m.handleMetricsFunc != nil {
		m.handleMetricsFunc(c)
		return
	}
	c.String(200, "# HELP test_metric Test metric\n# TYPE test_metric gauge\ntest_metric 1\n")
}

type mockHealthService struct {
	checkCalled bool
	checkFunc   func(ctx context.Context) (string, map[string]any)
	status      string
	details     map[string]any
}

func (m *mockHealthService) Check(ctx context.Context) (string, map[string]any) {
	m.checkCalled = true
	if m.checkFunc != nil {
		return m.checkFunc(ctx)
	}
	if m.status == "" {
		m.status = "ok"
	}
	if m.details == nil {
		m.details = map[string]any{"service": "healthy"}
	}
	return m.status, m.details
}

type mockLogger struct {
	infoCalled  bool
	errorCalled bool
	warnCalled  bool
	debugCalled bool
	messages    []string
}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.infoCalled = true
	m.messages = append(m.messages, msg)
}

func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.errorCalled = true
	m.messages = append(m.messages, msg)
}

func (m *mockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.warnCalled = true
	m.messages = append(m.messages, msg)
}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.debugCalled = true
	m.messages = append(m.messages, msg)
}
