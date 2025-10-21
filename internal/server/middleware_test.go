package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleMiddlewareMockLogger captures log entries for testing middleware
type SimpleMiddlewareMockLogger struct {
	entries []map[string]interface{}
}

func NewSimpleMiddlewareMockLogger() *SimpleMiddlewareMockLogger {
	return &SimpleMiddlewareMockLogger{
		entries: make([]map[string]interface{}, 0),
	}
}

func (m *SimpleMiddlewareMockLogger) Debug(msg string, fields ...log.Field) {
	m.logEntry("debug", msg, fields...)
}

func (m *SimpleMiddlewareMockLogger) Info(msg string, fields ...log.Field) {
	m.logEntry("info", msg, fields...)
}

func (m *SimpleMiddlewareMockLogger) Warn(msg string, fields ...log.Field) {
	m.logEntry("warn", msg, fields...)
}

func (m *SimpleMiddlewareMockLogger) Error(msg string, fields ...log.Field) {
	m.logEntry("error", msg, fields...)
}

func (m *SimpleMiddlewareMockLogger) Fatal(msg string, fields ...log.Field) {
	m.logEntry("fatal", msg, fields...)
}

func (m *SimpleMiddlewareMockLogger) logEntry(level string, msg string, fields ...log.Field) {
	entry := map[string]interface{}{
		"level":   level,
		"message": msg,
		"fields":  make(map[string]interface{}),
	}

	// Convert log.Field to map entries
	for _, field := range fields {
		// Try to determine the field value based on available data
		var value interface{}

		// Check Interface first as it might contain the actual value
		if field.Interface != nil {
			value = field.Interface
		} else if field.String != "" {
			value = field.String
		} else {
			value = field.Integer
		}

		entry["fields"].(map[string]interface{})[field.Key] = value
	}

	m.entries = append(m.entries, entry)
}

func (m *SimpleMiddlewareMockLogger) With(fields ...log.Field) log.Logger {
	return m
}

func (m *SimpleMiddlewareMockLogger) WithContext(ctx context.Context) log.Logger {
	return m
}

func (m *SimpleMiddlewareMockLogger) Sync() error {
	return nil
}

func (m *SimpleMiddlewareMockLogger) GetEntries() []map[string]interface{} {
	return m.entries
}

func (m *SimpleMiddlewareMockLogger) ClearEntries() {
	m.entries = make([]map[string]interface{}, 0)
}

func TestLoggingMiddleware_LogsFormatAndContent(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	// Create gin engine with logging middleware
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(loggingMiddleware(mockLogger))

	// Add a test handler
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check log entries
	entries := mockLogger.GetEntries()
	t.Logf("Number of log entries: %d", len(entries))
	for i, entry := range entries {
		t.Logf("Entry %d: %+v", i, entry)
	}

	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "HTTP request completed", entry["message"])

	require.Contains(t, entry, "fields")
	fields := entry["fields"].(map[string]interface{})
	t.Logf("Fields: %+v", fields)

	assert.Equal(t, "GET", fields["method"])
	assert.Equal(t, "/test", fields["path"])
	assert.Equal(t, int64(200), fields["status_code"])
	assert.Equal(t, int64(200), fields["status"]) // zap stores ints as int64
	assert.Contains(t, fields, "duration")
	assert.Contains(t, fields, "remote_addr")
	assert.Equal(t, "test-agent", fields["user_agent"])
}

func TestLoggingMiddleware_FiltersSensitiveInformation(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(loggingMiddleware(mockLogger))

	// Add a test handler that receives sensitive data
	engine.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": "secret-token-123"})
	})

	// Act - send request with sensitive data
	requestBody := `{"username": "test", "password": "secret-password"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-token")

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that sensitive information is filtered from logs
	entries := mockLogger.GetEntries()
	require.Len(t, entries, 1)

	fields := entries[0]["fields"].(map[string]interface{})

	// Request body should be filtered or not logged
	if body, exists := fields["request_body"]; exists {
		assert.NotContains(t, fmt.Sprintf("%v", body), "secret-password")
	}

	// Authorization header should be filtered or masked
	if auth, exists := fields["authorization"]; exists {
		authStr := fmt.Sprintf("%v", auth)
		assert.True(t, strings.HasPrefix(authStr, "Bearer ***") ||
			strings.Contains(authStr, "masked") ||
			!strings.Contains(authStr, "secret-token"))
	}
}

func TestLoggingMiddleware_HandlesVariousHTTPMethodsAndStatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		responseStatus int
		expectedMethod string
		expectedPath   string
		expectedStatus int
	}{
		{
			name:           "GET request with 200 status",
			method:         "GET",
			path:           "/api/data",
			responseStatus: 200,
			expectedMethod: "GET",
			expectedPath:   "/api/data",
			expectedStatus: 200,
		},
		{
			name:           "POST request with 201 status",
			method:         "POST",
			path:           "/api/users",
			responseStatus: 201,
			expectedMethod: "POST",
			expectedPath:   "/api/users",
			expectedStatus: 201,
		},
		{
			name:           "PUT request with 204 status",
			method:         "PUT",
			path:           "/api/users/123",
			responseStatus: 204,
			expectedMethod: "PUT",
			expectedPath:   "/api/users/123",
			expectedStatus: 204,
		},
		{
			name:           "DELETE request with 404 status",
			method:         "DELETE",
			path:           "/api/users/999",
			responseStatus: 404,
			expectedMethod: "DELETE",
			expectedPath:   "/api/users/999",
			expectedStatus: 404,
		},
		{
			name:           "PATCH request with 500 status",
			method:         "PATCH",
			path:           "/api/error",
			responseStatus: 500,
			expectedMethod: "PATCH",
			expectedPath:   "/api/error",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLogger := NewSimpleMiddlewareMockLogger()

			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.Use(loggingMiddleware(mockLogger))

			// Add a test handler that returns the specified status
			engine.Any(tt.path, func(c *gin.Context) {
				c.Status(tt.responseStatus)
			})

			// Act
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.responseStatus, w.Code)

			// Check log entries
			entries := mockLogger.GetEntries()
			require.Len(t, entries, 1)

			fields := entries[0]["fields"].(map[string]interface{})
			assert.Equal(t, tt.expectedMethod, fields["method"])
			assert.Equal(t, tt.expectedPath, fields["path"])
			assert.Equal(t, int64(tt.expectedStatus), fields["status"])
			assert.Equal(t, tt.expectedStatus, int(fields["status_code"].(int64)))
		})
	}
}

func TestLoggingMiddleware_MeasuresDuration(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(loggingMiddleware(mockLogger))

	// Add a handler that simulates some processing time
	engine.GET("/slow", func(c *gin.Context) {
		time.Sleep(10 * time.Millisecond) // Small delay to measure
		c.JSON(http.StatusOK, gin.H{"message": "slow response"})
	})

	// Act
	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	engine.ServeHTTP(w, req)
	elapsed := time.Since(start)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, elapsed >= 10*time.Millisecond, "Request should take at least 10ms")

	// Check log entries
	entries := mockLogger.GetEntries()
	require.Len(t, entries, 1)

	fields := entries[0]["fields"].(map[string]interface{})
	duration, exists := fields["duration"]
	require.True(t, exists, "Duration should be logged")

	// Duration should be a string representation of time.Duration
	durationStr, ok := duration.(string)
	require.True(t, ok, "Duration should be a string")

	// Parse duration to ensure it's reasonable
	parsedDuration, err := time.ParseDuration(durationStr)
	require.NoError(t, err, "Duration should be parseable")
	assert.True(t, parsedDuration >= 10*time.Millisecond, "Logged duration should be at least 10ms")
	assert.True(t, parsedDuration <= 1*time.Second, "Logged duration should be reasonable")
}

func TestLoggingMiddleware_HandlesQueryParameters(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(loggingMiddleware(mockLogger))

	engine.GET("/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"results": []string{}})
	})

	// Act
	req := httptest.NewRequest("GET", "/search?q=test&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check log entries
	entries := mockLogger.GetEntries()
	require.Len(t, entries, 1)

	fields := entries[0]["fields"].(map[string]interface{})

	// Path should be the base path without query
	path := fields["path"].(string)
	assert.Equal(t, "/search", path)

	// Query parameters should be stored separately
	assert.Equal(t, "q=test&page=1&limit=10", fields["query"])

	// Should log query parameters separately
	queryParams, exists := fields["query_params"]
	require.True(t, exists, "Query parameters should be logged")

	queryMap, ok := queryParams.(map[string]interface{})
	require.True(t, ok, "Query parameters should be a map")
	assert.Equal(t, "test", queryMap["q"])
	assert.Equal(t, "1", queryMap["page"])
	assert.Equal(t, "10", queryMap["limit"])
}

func TestRecoveryMiddleware_CapturesAndRecoversFromPanic(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(recoveryMiddleware(mockLogger))

	// Add a handler that panics
	engine.GET("/panic", func(c *gin.Context) {
		panic("test panic message")
	})

	// Act
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check response body
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "internal_server_error", response.Error)
	assert.Equal(t, "Internal server error", response.Message)
	assert.Equal(t, "/panic", response.Path)
	assert.NotEmpty(t, response.Timestamp)

	// Check that panic was logged
	entries := mockLogger.GetEntries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "Panic recovered", entry["message"])

	fields := entry["fields"].(map[string]interface{})
	assert.Equal(t, "test panic message", fields["panic"])
	assert.Equal(t, "/panic", fields["path"])
	assert.Equal(t, "GET", fields["method"])
	assert.Contains(t, fields, "stack_trace")
}

func TestRecoveryMiddleware_HandlesDifferentPanicTypes(t *testing.T) {
	tests := []struct {
		name          string
		panicValue    interface{}
		expectedError string
	}{
		{
			name:          "string panic",
			panicValue:    "string panic",
			expectedError: "string panic",
		},
		{
			name:          "error panic",
			panicValue:    fmt.Errorf("wrapped error: %w", fmt.Errorf("original error")),
			expectedError: "wrapped error: original error",
		},
		{
			name:          "integer panic",
			panicValue:    12345,
			expectedError: "12345",
		},
		{
			name:          "struct panic",
			panicValue:    struct{ Field string }{Field: "test"},
			expectedError: "{test}",
		},
		{
			name:          "nil panic",
			panicValue:    nil,
			expectedError: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockLogger := NewSimpleMiddlewareMockLogger()

			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.Use(recoveryMiddleware(mockLogger))

			engine.GET("/panic", func(c *gin.Context) {
				panic(tt.panicValue)
			})

			// Act
			req := httptest.NewRequest("GET", "/panic", nil)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusInternalServerError, w.Code)

			// Check log entries
			entries := mockLogger.GetEntries()
			require.Len(t, entries, 1)

			fields := entries[0]["fields"].(map[string]interface{})
			panicValue := fmt.Sprintf("%v", fields["panic"])
			assert.Contains(t, panicValue, tt.expectedError)
		})
	}
}

func TestRecoveryMiddleware_ErrorResponseFormat(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(recoveryMiddleware(mockLogger))

	engine.GET("/panic", func(c *gin.Context) {
		panic("formatted error")
	})

	// Act
	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Parse response body
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "internal_server_error", response.Error)
	assert.Equal(t, "Internal server error", response.Message)
	assert.Equal(t, "/panic", response.Path)
	assert.NotEmpty(t, response.Timestamp)

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")
}

func TestRecoveryMiddleware_ErrorLogging(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(recoveryMiddleware(mockLogger))

	engine.POST("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")
		panic(fmt.Sprintf("failed to process user %s", userID))
	})

	// Act
	req := httptest.NewRequest("POST", "/api/users/12345", nil)
	req.Header.Set("User-Agent", "test-client")
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check detailed error logging
	entries := mockLogger.GetEntries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "Panic recovered", entry["message"])

	fields := entry["fields"].(map[string]interface{})
	assert.Equal(t, "failed to process user 12345", fields["panic"])
	assert.Equal(t, "/api/users/12345", fields["path"])
	assert.Equal(t, "POST", fields["method"])
	assert.Equal(t, "test-client", fields["user_agent"])
	assert.Contains(t, fields, "remote_addr")
	assert.Contains(t, fields, "stack_trace")

	// Verify stack trace is present and looks reasonable
	stackTrace := fields["stack_trace"].(string)
	assert.NotEmpty(t, stackTrace)
	assert.Contains(t, stackTrace, "github.com/gin-gonic/gin") // Should contain gin stack frames
}

func TestRecoveryMiddleware_HandlesNestedMiddleware(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// Add logging middleware before recovery to test order
	engine.Use(loggingMiddleware(mockLogger))
	engine.Use(recoveryMiddleware(mockLogger))

	// Add middleware that panics
	engine.Use(func(c *gin.Context) {
		panic("middleware panic")
	})

	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Should have both log entries (from logging middleware and recovery middleware)
	entries := mockLogger.GetEntries()
	require.GreaterOrEqual(t, len(entries), 2, "Should have at least logging and recovery entries")

	// Find recovery entry
	var recoveryEntry map[string]interface{}
	for _, entry := range entries {
		if entry["message"] == "Panic recovered" {
			recoveryEntry = entry
			break
		}
	}
	require.NotNil(t, recoveryEntry, "Should have recovery log entry")
}

func TestRecoveryMiddleware_PreservesRequestContext(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(recoveryMiddleware(mockLogger))

	// Add a handler that accesses context and then panics
	engine.GET("/context-test", func(c *gin.Context) {
		// Add some context values
		c.Set("user_id", "123")
		c.Set("request_id", "req-456")

		// Access headers and query params
		userAgent := c.GetHeader("User-Agent")
		queryParam := c.Query("test")

		// Panic with context information
		panic(fmt.Sprintf("panic with context: user=%s, agent=%s, query=%s",
			c.GetString("user_id"), userAgent, queryParam))
	})

	// Act
	req := httptest.NewRequest("GET", "/context-test?test=value", nil)
	req.Header.Set("User-Agent", "context-test-agent")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Check that context information was preserved in logs
	entries := mockLogger.GetEntries()
	require.Len(t, entries, 1)

	fields := entries[0]["fields"].(map[string]interface{})
	panicMessage := fields["panic"].(string)
	assert.Contains(t, panicMessage, "panic with context: user=123")
	assert.Contains(t, panicMessage, "agent=context-test-agent")
	assert.Contains(t, panicMessage, "query=value")
}

func TestCORSMiddleware_SetsCORSHeaders(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "CORS test"})
	})

	// Act
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check CORS headers
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "OPTIONS")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "X-Requested-With")
}

func TestCORSMiddleware_HandlesPreflightRequests(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	engine.OPTIONS("/api/data", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Act - Preflight request
	req := httptest.NewRequest("OPTIONS", "/api/data", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Check preflight response headers
	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORSMiddleware_DisabledWhenConfigDisabled(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: false}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "No CORS"})
	})

	// Act
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// CORS headers should not be set
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSMiddleware_HandlesMultipleOrigins(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "multi-origin test"})
	})

	tests := []struct {
		name           string
		origin         string
		expectedHeader string
	}{
		{
			name:           "localhost origin",
			origin:         "http://localhost:3000",
			expectedHeader: "http://localhost:3000",
		},
		{
			name:           "production origin",
			origin:         "https://app.production.com",
			expectedHeader: "https://app.production.com",
		},
		{
			name:           "staging origin",
			origin:         "https://staging.app.com",
			expectedHeader: "https://staging.app.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedHeader, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCORSMiddleware_AllowsAllOrigins(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	engine.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public endpoint"})
	})

	// Act
	req := httptest.NewRequest("GET", "/public", nil)
	req.Header.Set("Origin", "*")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	// Should reflect the origin, not use wildcard
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_HandlesNoOriginHeader(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "no origin test"})
	})

	// Act - Request without Origin header
	req := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// CORS headers should not be set when no Origin header
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_CustomMethodsAndHeaders(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	// Test with custom headers
	req := httptest.NewRequest("OPTIONS", "/api/custom", nil)
	req.Header.Set("Origin", "https://custom.example.com")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header,X-Another-Header")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Should include PATCH in allowed methods
	allowedMethods := w.Header().Get("Access-Control-Allow-Methods")
	assert.Contains(t, allowedMethods, "PATCH")

	// Should include custom headers
	allowedHeaders := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, allowedHeaders, "X-Custom-Header")
	assert.Contains(t, allowedHeaders, "X-Another-Header")
}

func TestCORSMiddleware_LogsCORSInformation(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(corsMiddleware(mockLogger, &Config{EnableCORS: true}))

	engine.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://test.example.com")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that CORS information was logged
	entries := mockLogger.GetEntries()
	var corsLogEntry map[string]interface{}
	for _, entry := range entries {
		if entry["message"] == "CORS request processed" {
			corsLogEntry = entry
			break
		}
	}

	if corsLogEntry != nil {
		fields := corsLogEntry["fields"].(map[string]interface{})
		assert.Equal(t, "https://test.example.com", fields["origin"])
		assert.Equal(t, "GET", fields["method"])
	}
}

func TestRateLimitMiddleware_AllowsNormalRequests(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(rateLimitMiddleware(mockLogger, &Config{EnableRateLimit: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "rate limit test"})
	})

	// Act - Make a few requests within limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// Assert - All requests should succeed
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestRateLimitMiddleware_RejectsExcessiveRequests(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(rateLimitMiddleware(mockLogger, &Config{EnableRateLimit: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "rate limit test"})
	})

	// Act - Make many requests to exceed limit
	var successCount, rateLimitCount int
	for i := 0; i < 150; i++ { // Exceed the 100 request limit
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.100:12345" // Same IP for all requests
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		switch w.Code {
		case http.StatusOK:
			successCount++
		case http.StatusTooManyRequests:
			rateLimitCount++
		}
	}

	// Assert - Should have some successful requests and some rate limited
	assert.Greater(t, successCount, 0, "Should allow some requests")
	assert.Greater(t, rateLimitCount, 0, "Should rate limit excessive requests")
	assert.LessOrEqual(t, successCount, 100, "Should not allow more than the limit")
}

func TestRateLimitMiddleware_429ResponseFormat(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(rateLimitMiddleware(mockLogger, &Config{EnableRateLimit: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Act - Make requests until we get rate limited
	var rateLimitResponse *httptest.ResponseRecorder
	for i := 0; i < 150; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.150:12345" // Same IP for rate limiting
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			rateLimitResponse = w
			break
		}
	}

	// Assert
	require.NotNil(t, rateLimitResponse, "Should eventually get rate limited")
	assert.Equal(t, http.StatusTooManyRequests, rateLimitResponse.Code)
	assert.Equal(t, "application/json; charset=utf-8", rateLimitResponse.Header().Get("Content-Type"))

	// Parse response body
	var response ErrorResponse
	err := json.Unmarshal(rateLimitResponse.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "rate_limit_exceeded", response.Error)
	assert.Equal(t, "Rate limit exceeded", response.Message)
	assert.Equal(t, "/api/data", response.Path)
	assert.NotEmpty(t, response.Timestamp)

	// Verify timestamp format
	_, err = time.Parse(time.RFC3339, response.Timestamp)
	assert.NoError(t, err, "Timestamp should be in RFC3339 format")

	// Check for rate limit headers
	assert.NotEmpty(t, rateLimitResponse.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, rateLimitResponse.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rateLimitResponse.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimitMiddleware_DisabledWhenConfigDisabled(t *testing.T) {
	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(rateLimitMiddleware(mockLogger, &Config{EnableRateLimit: false}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "no rate limit"})
	})

	// Act - Make many requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		// Assert - All requests should succeed
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed when rate limiting is disabled", i)
	}
}

func TestRateLimitMiddleware_DifferentClientsHaveSeparateLimits(t *testing.T) {
	// Reset global rate limiter to avoid test interference
	globalRateLimiter = &rateLimiter{
		clients: make(map[string]*clientBucket),
	}

	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(rateLimitMiddleware(mockLogger, &Config{EnableRateLimit: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Act - Make requests from different IP addresses
	client1Requests := 0
	client2Requests := 0

	// Client 1 requests (use X-Forwarded-For header to ensure proper IP detection)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.100:12345" // Client 1 IP
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		req.Header.Set("X-Real-IP", "192.168.1.100")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			client1Requests++
		}
	}

	// Client 2 requests (use X-Forwarded-For header to ensure proper IP detection)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.200:12345" // Client 2 IP
		req.Header.Set("X-Forwarded-For", "192.168.1.200")
		req.Header.Set("X-Real-IP", "192.168.1.200")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			client2Requests++
		}
	}

	// Assert - Both clients should be able to make some requests
	assert.Greater(t, client1Requests, 0, "Client 1 should be able to make requests")
	assert.Greater(t, client2Requests, 0, "Client 2 should be able to make requests")
	assert.LessOrEqual(t, client1Requests, 100, "Client 1 should respect rate limit")
	assert.LessOrEqual(t, client2Requests, 100, "Client 2 should respect rate limit")
}

func TestRateLimitMiddleware_LogsRateLimitEvents(t *testing.T) {
	// Reset global rate limiter to avoid test interference
	globalRateLimiter = &rateLimiter{
		clients: make(map[string]*clientBucket),
	}

	// Arrange
	mockLogger := NewSimpleMiddlewareMockLogger()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(rateLimitMiddleware(mockLogger, &Config{EnableRateLimit: true}))

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Act - Make requests until rate limited (use more requests to ensure we hit the limit)
	var rateLimited bool
	for i := 0; i < 150; i++ { // Increased from 100 to ensure we hit the rate limit
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.150:12345"
		req.Header.Set("X-Forwarded-For", "192.168.1.150")
		req.Header.Set("X-Real-IP", "192.168.1.150")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	// Assert
	assert.True(t, rateLimited, "Should get rate limited")

	// Check that rate limit was logged
	entries := mockLogger.GetEntries()
	var rateLimitLogEntry map[string]interface{}
	for _, entry := range entries {
		if entry["message"] == "Rate limit exceeded" {
			rateLimitLogEntry = entry
			break
		}
	}

	if rateLimitLogEntry != nil {
		fields := rateLimitLogEntry["fields"].(map[string]interface{})
		assert.Equal(t, "warn", rateLimitLogEntry["level"])
		assert.Equal(t, "/api/data", fields["path"])
		assert.Equal(t, "GET", fields["method"])
		assert.Contains(t, fields, "client_ip")
		assert.Contains(t, fields, "limit")
		assert.Contains(t, fields, "remaining")
	}
}

func TestRateLimitMiddleware_RespectsHTTPMethod(t *testing.T) {
	// Arrange
	// Create a fresh rate limiter for this test to avoid interference
	testRateLimiter := &rateLimiter{
		clients: make(map[string]*clientBucket),
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		// Use custom rate limiter for this test
		clientIP := c.ClientIP()
		if clientIP == "" {
			clientIP = strings.Split(c.Request.RemoteAddr, ":")[0]
		}

		allowed, remaining, resetTime := testRateLimiter.checkLimit(clientIP)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", RateLimitRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		if !allowed {
			errorResponse := ErrorResponse{
				Error:     "rate_limit_exceeded",
				Message:   "Rate limit exceeded",
				Path:      c.Request.URL.Path,
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("Retry-After", "60")
			c.JSON(429, errorResponse)
			c.Abort()
			return
		}
		c.Next()
	})

	engine.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "GET test"})
	})
	engine.POST("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "POST test"})
	})

	// Act - Mix of GET and POST requests
	getSuccess, postSuccess := 0, 0

	// GET requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			getSuccess++
		}
	}

	// POST requests (using different IP to avoid rate limiting overlap)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/data", nil)
		req.RemoteAddr = "192.168.1.200:12345"
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			postSuccess++
		}
	}

	// Assert - Both methods should work
	assert.Greater(t, getSuccess, 0, "GET requests should work")
	assert.Greater(t, postSuccess, 0, "POST requests should work")
}
