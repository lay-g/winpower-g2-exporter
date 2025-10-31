package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("loggerMiddleware logs request", func(t *testing.T) {
		mockLog := &mockLogger{}
		srv := &HTTPServer{
			log: mockLog,
		}

		// Create test router
		router := gin.New()
		router.Use(srv.loggerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		// Make request
		req := httptest.NewRequest("GET", "/test?foo=bar", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify logging
		if !mockLog.infoCalled {
			t.Error("Expected Info to be called")
		}
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("recoveryMiddleware catches panic", func(t *testing.T) {
		mockLog := &mockLogger{}
		srv := &HTTPServer{
			log: mockLog,
		}

		// Create test router
		router := gin.New()
		router.Use(srv.recoveryMiddleware())
		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		// Make request
		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify error logging
		if !mockLog.errorCalled {
			t.Error("Expected Error to be called")
		}
		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}

func TestErrorResponse(t *testing.T) {
	t.Run("NewErrorResponse creates valid response", func(t *testing.T) {
		resp := NewErrorResponse(ErrInvalidConfig, "/test")

		if resp.Error != "invalid server configuration" {
			t.Errorf("Expected error message, got %s", resp.Error)
		}
		if resp.Path != "/test" {
			t.Errorf("Expected path /test, got %s", resp.Path)
		}
		if resp.Time == "" {
			t.Error("Expected timestamp to be set")
		}
	})
}
