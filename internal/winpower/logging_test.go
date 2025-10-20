package winpower

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestLogLevel_String tests the LogLevel String method
func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		logLevel LogLevel
		expected string
	}{
		{"debug", LogLevelDebug, "debug"},
		{"info", LogLevelInfo, "info"},
		{"warn", LogLevelWarn, "warn"},
		{"error", LogLevelError, "error"},
		{"unknown", LogLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.logLevel.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLogLevel_ToZapLevel tests the LogLevel ToZapLevel method
func TestLogLevel_ToZapLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel LogLevel
		expected string
	}{
		{"debug_to_debug", LogLevelDebug, "debug"},
		{"info_to_info", LogLevelInfo, "info"},
		{"warn_to_warn", LogLevelWarn, "warn"},
		{"error_to_error", LogLevelError, "error"},
		{"unknown_to_info", LogLevel(999), "info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zapLevel := tt.logLevel.ToZapLevel()
			if got := zapLevel.String(); got != tt.expected {
				t.Errorf("LogLevel.ToZapLevel() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLoggerConfig tests the logger configuration
func TestLoggerConfig(t *testing.T) {
	t.Run("DefaultLoggerConfig", func(t *testing.T) {
		config := DefaultLoggerConfig()
		if config == nil {
			t.Fatal("DefaultLoggerConfig() should not return nil")
		}

		if config.Level != LogLevelInfo {
			t.Errorf("DefaultLoggerConfig() Level = %v, want %v", config.Level, LogLevelInfo)
		}

		if config.Development != false {
			t.Errorf("DefaultLoggerConfig() Development = %v, want false", config.Development)
		}

		if config.EnableConsole != true {
			t.Errorf("DefaultLoggerConfig() EnableConsole = %v, want true", config.EnableConsole)
		}

		if len(config.SensitiveFields) == 0 {
			t.Error("DefaultLoggerConfig() should have sensitive fields configured")
		}
	})
}

// TestNewLogger tests the logger creation
func TestNewLogger(t *testing.T) {
	t.Run("with_nil_config", func(t *testing.T) {
		logger, err := NewLogger(nil)
		if err != nil {
			t.Fatalf("NewLogger(nil) should not return error: %v", err)
		}
		defer func() {
					if err := logger.Close(); err != nil {
						t.Logf("Warning: failed to close logger: %v", err)
					}
				}()

		if logger == nil {
			t.Fatal("NewLogger() should not return nil logger")
		}
	})

	t.Run("with_custom_config", func(t *testing.T) {
		config := &LoggerConfig{
			Level:            LogLevelDebug,
			Development:      true,
			EnableConsole:    true,
			EnableCaller:     false,
			EnableStacktrace: true,
		}

		logger, err := NewLogger(config)
		if err != nil {
			t.Fatalf("NewLogger() should not return error: %v", err)
		}
		defer func() {
					if err := logger.Close(); err != nil {
						t.Logf("Warning: failed to close logger: %v", err)
					}
				}()

		if logger == nil {
			t.Fatal("NewLogger() should not return nil logger")
		}
	})

	t.Run("NewLoggerWithDefaults", func(t *testing.T) {
		logger, err := NewLoggerWithDefaults()
		if err != nil {
			t.Fatalf("NewLoggerWithDefaults() should not return error: %v", err)
		}
		defer func() {
					if err := logger.Close(); err != nil {
						t.Logf("Warning: failed to close logger: %v", err)
					}
				}()

		if logger == nil {
			t.Fatal("NewLoggerWithDefaults() should not return nil logger")
		}
	})
}

// TestLogger_WithField tests field addition
func TestLogger_WithField(t *testing.T) {
	logger := &Logger{
		config: DefaultLoggerConfig(),
		fields: make(map[string]interface{}),
	}

	// Test adding a field
	newLogger := logger.WithField("key1", "value1")
	if newLogger.fields["key1"] != "value1" {
		t.Errorf("WithField() = %v, want %v", newLogger.fields["key1"], "value1")
	}

	// Test that original logger is not modified
	if _, exists := logger.fields["key1"]; exists {
		t.Error("WithField() should not modify original logger")
	}

	// Test chaining
	chainedLogger := logger.WithField("key2", "value2").WithField("key3", "value3")
	if chainedLogger.fields["key2"] != "value2" || chainedLogger.fields["key3"] != "value3" {
		t.Error("WithField() chaining not working correctly")
	}
}

// TestLogger_WithFields tests multiple field addition
func TestLogger_WithFields(t *testing.T) {
	logger := &Logger{
		config: DefaultLoggerConfig(),
		fields: make(map[string]interface{}),
	}

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	newLogger := logger.WithFields(fields)

	if newLogger.fields["key1"] != "value1" {
		t.Errorf("WithFields() key1 = %v, want %v", newLogger.fields["key1"], "value1")
	}

	if newLogger.fields["key2"] != 42 {
		t.Errorf("WithFields() key2 = %v, want %v", newLogger.fields["key2"], 42)
	}

	if newLogger.fields["key3"] != true {
		t.Errorf("WithFields() key3 = %v, want %v", newLogger.fields["key3"], true)
	}
}

// TestLogger_WithDeviceID tests device ID field addition
func TestLogger_WithDeviceID(t *testing.T) {
	logger := &Logger{
		config: DefaultLoggerConfig(),
		fields: make(map[string]interface{}),
	}

	deviceID := "device-123"
	newLogger := logger.WithDeviceID(deviceID)

	if newLogger.fields["device_id"] != deviceID {
		t.Errorf("WithDeviceID() = %v, want %v", newLogger.fields["device_id"], deviceID)
	}
}

// TestLogger_WithOperation tests operation field addition
func TestLogger_WithOperation(t *testing.T) {
	logger := &Logger{
		config: DefaultLoggerConfig(),
		fields: make(map[string]interface{}),
	}

	operation := "collect_data"
	newLogger := logger.WithOperation(operation)

	if newLogger.fields["operation"] != operation {
		t.Errorf("WithOperation() = %v, want %v", newLogger.fields["operation"], operation)
	}
}

// TestLogger_SensitiveDataMasking tests sensitive data masking
func TestLogger_SensitiveDataMasking(t *testing.T) {
	config := &LoggerConfig{
		SensitiveFields: []string{"password", "token", "secret"},
	}

	logger := &Logger{
		config: config,
		fields: make(map[string]interface{}),
	}

	// Test masking of known sensitive fields
	newLogger := logger.WithField("password", "secret123")
	if masked, ok := newLogger.fields["password"].(string); ok {
		if masked == "secret123" {
			t.Error("Sensitive field 'password' should be masked")
		}
		if len(masked) == 0 {
			t.Error("Masked password should not be empty")
		}
	}

	// Test non-sensitive field is not masked
	newLogger2 := logger.WithField("username", "testuser")
	if newLogger2.fields["username"] != "testuser" {
		t.Errorf("Non-sensitive field should not be masked, got %v", newLogger2.fields["username"])
	}
}

// TestLogger_SpecializedLogging tests specialized logging methods
func TestLogger_SpecializedLogging(t *testing.T) {
	// Use zaptest logger for testing
	zapLogger := zaptest.NewLogger(t)
	defer func() {
					if err := zapLogger.Sync(); err != nil {
						t.Logf("Warning: failed to sync zap logger: %v", err)
					}
				}()

	logger := &Logger{
		zapLogger: zapLogger,
		config:    DefaultLoggerConfig(),
		fields:    make(map[string]interface{}),
	}

	t.Run("LogCollectionStart", func(t *testing.T) {
		// This should not panic
		logger.LogCollectionStart(5)
	})

	t.Run("LogCollectionEnd", func(t *testing.T) {
		// Success case
		logger.LogCollectionEnd(5, 100*time.Millisecond, nil)

		// Error case
		err := NewRequestError("collect", "collection failed")
		logger.LogCollectionEnd(0, 50*time.Millisecond, err)
	})

	t.Run("LogConnectionStatus", func(t *testing.T) {
		logger.LogConnectionStatus("connected", "")
		logger.LogConnectionStatus("disconnected", "device-123")
	})

	t.Run("LogAuthentication", func(t *testing.T) {
		logger.LogAuthentication(true, "testuser", nil)
		logger.LogAuthentication(false, "testuser", NewAuthError("login", "invalid credentials"))
	})
}

// TestLogger_LogError tests WinPower error logging
func TestLogger_LogError(t *testing.T) {
	zapLogger := zaptest.NewLogger(t)
	defer func() {
					if err := zapLogger.Sync(); err != nil {
						t.Logf("Warning: failed to sync zap logger: %v", err)
					}
				}()

	logger := &Logger{
		zapLogger: zapLogger,
		config:    DefaultLoggerConfig(),
		fields:    make(map[string]interface{}),
	}

	t.Run("log_winpower_error", func(t *testing.T) {
		err := NewConnectionError("connect", "connection failed").
			WithDeviceID("device-456").
			WithContext("attempt", 3).
			WithCause(NewInternalError("system", "network down"))

		// This should not panic
		logger.LogError(err)
	})

	t.Run("log_generic_error", func(t *testing.T) {
		genericErr := &testError{message: "generic error"}

		// This should not panic
		logger.LogError(genericErr)
	})
}

// TestLogger_WithContext tests context integration
func TestLogger_WithContext(t *testing.T) {
	zapLogger := zaptest.NewLogger(t)
	defer func() {
					if err := zapLogger.Sync(); err != nil {
						t.Logf("Warning: failed to sync zap logger: %v", err)
					}
				}()

	logger := &Logger{
		zapLogger: zapLogger,
		config:    DefaultLoggerConfig(),
		fields:    make(map[string]interface{}),
	}

	// Use typed keys to avoid lint warnings
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, TraceIDKey, "trace-456")
	ctx = context.WithValue(ctx, UserIDKey, "user-789")

	newLogger := logger.WithContext(ctx)

	if newLogger.fields["request_id"] != "req-123" {
		t.Errorf("WithContext() request_id = %v, want %v", newLogger.fields["request_id"], "req-123")
	}

	if newLogger.fields["trace_id"] != "trace-456" {
		t.Errorf("WithContext() trace_id = %v, want %v", newLogger.fields["trace_id"], "trace-456")
	}

	if newLogger.fields["user_id"] != "user-789" {
		t.Errorf("WithContext() user_id = %v, want %v", newLogger.fields["user_id"], "user-789")
	}
}

// testError is a simple error implementation for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// TestLogger_Production tests logger in production mode
func TestLogger_Production(t *testing.T) {
	config := &LoggerConfig{
		Level:            LogLevelInfo,
		Development:      false,
		EnableConsole:    false,
		EnableFile:       true,
		FilePath:         "/tmp/test-winpower.log",
		EnableCaller:     true,
		EnableStacktrace: false,
	}

	logger, err := NewLogger(config)
	if err != nil {
		t.Fatalf("NewLogger() with production config should not return error: %v", err)
	}
	defer func() {
					if err := logger.Close(); err != nil {
						t.Logf("Warning: failed to close logger: %v", err)
					}
				}()
	defer func() {
					if err := os.Remove("/tmp/test-winpower.log"); err != nil {
						t.Logf("Warning: failed to remove test log file: %v", err)
					}
				}()

	// Test basic logging
	logger.Info("Test production logging")
	logger.Error("Test production error logging")

	// Test closing
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() should not return error: %v", err)
	}
}

// TestLogger_ConcurrentLogging tests concurrent logging safety
func TestLogger_ConcurrentLogging(t *testing.T) {
	logger := &Logger{
		zapLogger: zaptest.NewLogger(t),
		config:    DefaultLoggerConfig(),
		fields:    make(map[string]interface{}),
	}
	defer func() {
					if err := logger.Close(); err != nil {
						t.Logf("Warning: failed to close logger: %v", err)
					}
				}()

	// Test concurrent field additions and logging
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			l := logger.WithField("goroutine", id)
			for j := 0; j < 10; j++ {
				l.Info("Concurrent logging test",
					zap.Int("iteration", j))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
