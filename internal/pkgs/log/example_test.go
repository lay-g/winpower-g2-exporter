package log_test

import (
	"context"
	"fmt"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// ExampleInit demonstrates basic logger initialization and usage.
func ExampleInit() {
	// Initialize with default configuration
	log.Init(log.DefaultConfig())

	// Log messages at different levels
	log.Info("application started", log.String("version", "1.0.0"))
	log.Debug("debug info", log.Int("count", 42))

	// Clean up
	log.Default().Sync()
}

// ExampleInitDevelopment demonstrates development environment setup.
func ExampleInitDevelopment() {
	// Initialize with development defaults
	log.InitDevelopment()

	// Logs will include caller information and use console format
	log.Debug("development mode enabled")

	// Clean up
	log.Default().Sync()
}

// ExampleLogger_With demonstrates creating child loggers.
func ExampleLogger_With() {
	log.Init(log.DefaultConfig())

	// Create a module-specific logger
	logger := log.Default().With(
		log.String("module", "collector"),
		log.String("component", "data-fetcher"),
	)

	// All logs from this logger will include module and component fields
	logger.Info("module started")
	logger.Debug("fetching data", log.String("source", "device-001"))

	// Clean up
	log.Default().Sync()
}

// ExampleWithRequestID demonstrates context-aware logging.
func ExampleWithRequestID() {
	log.Init(log.DefaultConfig())

	// Create context with request ID
	ctx := log.WithRequestID(context.Background(), "req-123")

	// Log with context - automatically includes request_id
	log.InfoContext(ctx, "processing request")

	// Add more context fields
	ctx = log.WithTraceID(ctx, "trace-456")
	ctx = log.WithUserID(ctx, "user-789")

	log.InfoContext(ctx, "request completed")

	// Clean up
	log.Default().Sync()
}

// ExampleNewTestLogger demonstrates testing with captured logs.
func ExampleNewTestLogger() {
	// Create test logger
	testLogger := log.NewTestLogger()

	// Use logger in code under test
	testLogger.Info("test message", log.String("key", "value"))
	testLogger.Error("test error", log.String("reason", "example"))

	// Verify logged messages
	entries := testLogger.Entries()
	fmt.Printf("Total entries: %d\n", len(entries))

	// Query specific log levels using zapcore levels
	infoLogs := testLogger.EntriesByLevel(entries[0].Level)  // InfoLevel
	errorLogs := testLogger.EntriesByLevel(entries[1].Level) // ErrorLevel

	fmt.Printf("Info logs: %d\n", len(infoLogs))
	fmt.Printf("Error logs: %d\n", len(errorLogs))

	// Output:
	// Total entries: 2
	// Info logs: 1
	// Error logs: 1
}

// ExampleConfig demonstrates custom configuration.
func ExampleConfig() {
	// Create custom configuration
	config := log.DefaultConfig()
	config.Level = "debug"
	config.Format = "json"
	config.Output = "stdout"
	config.Development = true
	config.EnableCaller = true

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		return
	}

	// Initialize logger
	log.Init(config)

	// Use logger
	log.Debug("custom config loaded")

	// Clean up
	log.Default().Sync()
}

// ExampleLogger_WithContext demonstrates extracting context fields.
func ExampleLogger_WithContext() {
	log.Init(log.DefaultConfig())

	// Create context with multiple fields
	ctx := context.Background()
	ctx = log.WithRequestID(ctx, "req-123")
	ctx = log.WithTraceID(ctx, "trace-456")
	ctx = log.WithComponent(ctx, "api-handler")
	ctx = log.WithOperation(ctx, "process-data")

	// Get logger from context (or use default)
	logger := log.FromContext(ctx)

	// Create context-aware logger
	ctxLogger := logger.WithContext(ctx)

	// All logs will include context fields
	ctxLogger.Info("processing started")
	ctxLogger.Debug("validation complete", log.Int("items", 100))
	ctxLogger.Info("processing finished")

	// Clean up
	log.Default().Sync()
}

// ExampleDevelopmentDefaults demonstrates development configuration.
func ExampleDevelopmentDefaults() {
	// Get development defaults
	config := log.DevelopmentDefaults()

	fmt.Printf("Level: %s\n", config.Level)
	fmt.Printf("Format: %s\n", config.Format)
	fmt.Printf("EnableCaller: %v\n", config.EnableCaller)

	// Output:
	// Level: debug
	// Format: console
	// EnableCaller: true
}

// ExampleField demonstrates using typed field constructors.
func ExampleField() {
	log.Init(log.DefaultConfig())

	// Use various field types
	log.Info("user action",
		log.String("user_id", "user-123"),
		log.Int("action_type", 1),
		log.Bool("success", true),
		log.Duration("elapsed", 150*time.Millisecond),
		log.Float64("score", 95.5),
	)

	// Clean up
	log.Default().Sync()
}
