package winpower

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Context keys for type-safe context values
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
	UserIDKey    contextKey = "user_id"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	// LogLevelDebug represents debug level logging
	LogLevelDebug LogLevel = iota
	// LogLevelInfo represents info level logging
	LogLevelInfo
	// LogLevelWarn represents warning level logging
	LogLevelWarn
	// LogLevelError represents error level logging
	LogLevelError
)

// String returns the string representation of the log level
func (ll LogLevel) String() string {
	switch ll {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
	}
}

// ToZapLevel converts LogLevel to zapcore.Level
func (ll LogLevel) ToZapLevel() zapcore.Level {
	switch ll {
	case LogLevelDebug:
		return zapcore.DebugLevel
	case LogLevelInfo:
		return zapcore.InfoLevel
	case LogLevelWarn:
		return zapcore.WarnLevel
	case LogLevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// LoggerConfig represents configuration for the structured logger
type LoggerConfig struct {
	// Level is the minimum log level to output
	Level LogLevel `yaml:"level" json:"level"`

	// Development indicates whether to use development mode logging
	Development bool `yaml:"development" json:"development"`

	// EnableConsole enables console output
	EnableConsole bool `yaml:"enable_console" json:"enable_console"`

	// EnableFile enables file output
	EnableFile bool `yaml:"enable_file" json:"enable_file"`

	// FilePath is the path to the log file
	FilePath string `yaml:"file_path" json:"file_path"`

	// MaxSize is the maximum size of a log file before rotation
	MaxSize int `yaml:"max_size" json:"max_size"`

	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `yaml:"max_backups" json:"max_backups"`

	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `yaml:"max_age" json:"max_age"`

	// EnableColors enables colored console output
	EnableColors bool `yaml:"enable_colors" json:"enable_colors"`

	// EnableCaller enables caller information in logs
	EnableCaller bool `yaml:"enable_caller" json:"enable_caller"`

	// EnableStacktrace enables stack traces for error logs
	EnableStacktrace bool `yaml:"enable_stacktrace" json:"enable_stacktrace"`

	// TimeEncoding specifies the time encoding format
	TimeEncoding string `yaml:"time_encoding" json:"time_encoding"`

	// SensitiveFields are field names that should be masked in logs
	SensitiveFields []string `yaml:"sensitive_fields" json:"sensitive_fields"`
}

// DefaultLoggerConfig returns a default logger configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:            LogLevelInfo,
		Development:      false,
		EnableConsole:    true,
		EnableFile:       false,
		MaxSize:          100, // 100MB
		MaxBackups:       3,
		MaxAge:           7, // 7 days
		EnableColors:     true,
		EnableCaller:     true,
		EnableStacktrace: false,
		TimeEncoding:     "iso8601",
		SensitiveFields:  []string{"password", "token", "secret", "key"},
	}
}

// Logger wraps zap logger with WinPower-specific functionality
type Logger struct {
	zapLogger *zap.Logger
	config    *LoggerConfig
	fields    map[string]interface{}
}

// NewLogger creates a new structured logger with the given configuration
func NewLogger(config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLoggerConfig()
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(config.Level.ToZapLevel())
	zapConfig.Development = config.Development
	zapConfig.DisableCaller = !config.EnableCaller
	zapConfig.DisableStacktrace = !config.EnableStacktrace

	// Configure output
	if config.EnableConsole && config.EnableFile {
		zapConfig.OutputPaths = []string{"stdout", config.FilePath}
		zapConfig.ErrorOutputPaths = []string{"stderr", config.FilePath}
	} else if config.EnableConsole {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	} else if config.EnableFile && config.FilePath != "" {
		zapConfig.OutputPaths = []string{config.FilePath}
		zapConfig.ErrorOutputPaths = []string{config.FilePath}
	}

	// Configure encoder
	if config.Development {
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		zapConfig.EncoderConfig = zap.NewProductionEncoderConfig()
	}

	// Configure time format
	switch config.TimeEncoding {
	case "iso8601":
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "epoch":
		zapConfig.EncoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	case "epochMillis":
		zapConfig.EncoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	}

	// Configure colors for development mode
	if config.Development && config.EnableColors {
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	return &Logger{
		zapLogger: zapLogger,
		config:    config,
		fields:    make(map[string]interface{}),
	}, nil
}

// NewLoggerWithDefaults creates a new logger with default configuration
func NewLoggerWithDefaults() (*Logger, error) {
	return NewLogger(nil)
}

// WithField adds a field to all subsequent log messages
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		zapLogger: l.zapLogger,
		config:    l.config,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = l.maskSensitiveValue(key, value)
	return newLogger
}

// WithFields adds multiple fields to all subsequent log messages
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		zapLogger: l.zapLogger,
		config:    l.config,
		fields:    make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = l.maskSensitiveValue(k, v)
	}

	return newLogger
}

// WithDeviceID adds device ID to all subsequent log messages
func (l *Logger) WithDeviceID(deviceID string) *Logger {
	return l.WithField("device_id", deviceID)
}

// WithOperation adds operation name to all subsequent log messages
func (l *Logger) WithOperation(operation string) *Logger {
	return l.WithField("operation", operation)
}

// maskSensitiveValue masks sensitive information in log output
func (l *Logger) maskSensitiveValue(key string, value interface{}) interface{} {
	keyLower := fmt.Sprintf("%v", key)
	for _, sensitiveField := range l.config.SensitiveFields {
		if keyLower == sensitiveField {
			return maskSensitive(fmt.Sprintf("%v", value))
		}
	}
	return value
}

// buildZapFields converts the internal fields to zap fields
func (l *Logger) buildZapFields() []zap.Field {
	fields := make([]zap.Field, 0, len(l.fields))
	for k, v := range l.fields {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...zap.Field) {
	allFields := append(l.buildZapFields(), fields...)
	l.zapLogger.Debug(message, allFields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...zap.Field) {
	allFields := append(l.buildZapFields(), fields...)
	l.zapLogger.Info(message, allFields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...zap.Field) {
	allFields := append(l.buildZapFields(), fields...)
	l.zapLogger.Warn(message, allFields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...zap.Field) {
	allFields := append(l.buildZapFields(), fields...)
	l.zapLogger.Error(message, allFields...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, fields ...zap.Field) {
	allFields := append(l.buildZapFields(), fields...)
	l.zapLogger.Fatal(message, allFields...)
}

// LogError logs a WinPower error with structured information
func (l *Logger) LogError(err error) {
	if wErr, ok := err.(*Error); ok {
		fields := []zap.Field{
			zap.String("error_type", wErr.Type.String()),
			zap.String("error_severity", wErr.Severity.String()),
			zap.String("error_code", wErr.Code),
			zap.String("operation", wErr.Operation),
			zap.Time("timestamp", wErr.Timestamp),
			zap.Bool("retryable", wErr.Retryable),
		}

		if wErr.DeviceID != "" {
			fields = append(fields, zap.String("device_id", wErr.DeviceID))
		}

		if wErr.Cause != nil {
			fields = append(fields, zap.NamedError("cause", wErr.Cause))
		}

		if len(wErr.Context) > 0 {
			fields = append(fields, zap.Any("context", wErr.Context))
		}

		allFields := append(l.buildZapFields(), fields...)

		switch wErr.Severity {
		case ErrorSeverityLow, ErrorSeverityMedium:
			l.zapLogger.Warn(wErr.Message, allFields...)
		case ErrorSeverityHigh:
			l.zapLogger.Error(wErr.Message, allFields...)
		case ErrorSeverityCritical:
			l.zapLogger.Fatal(wErr.Message, allFields...)
		default:
			l.zapLogger.Error(wErr.Message, allFields...)
		}
	} else {
		fields := append(l.buildZapFields(), zap.NamedError("error", err))
		l.zapLogger.Error("Unexpected error", fields...)
	}
}

// LogCollectionStart logs the start of data collection
func (l *Logger) LogCollectionStart(deviceCount int) {
	l.Info("Starting device data collection",
		append(l.buildZapFields(),
			zap.Int("device_count", deviceCount),
			zap.Time("start_time", time.Now()),
		)...)
}

// LogCollectionEnd logs the end of data collection
func (l *Logger) LogCollectionEnd(deviceCount int, duration time.Duration, err error) {
	fields := append(l.buildZapFields(),
		zap.Int("device_count", deviceCount),
		zap.Duration("duration", duration),
		zap.Time("end_time", time.Now()),
	)

	if err != nil {
		fields = append(fields, zap.NamedError("error", err))
		l.Warn("Device data collection completed with errors", fields...)
	} else {
		l.Info("Device data collection completed successfully", fields...)
	}
}

// LogConnectionStatus logs connection status changes
func (l *Logger) LogConnectionStatus(status string, deviceID string) {
	fields := append(l.buildZapFields(),
		zap.String("status", status),
		zap.Time("timestamp", time.Now()),
	)

	if deviceID != "" {
		fields = append(fields, zap.String("device_id", deviceID))
	}

	l.Info("Connection status changed", fields...)
}

// LogAuthentication logs authentication attempts
func (l *Logger) LogAuthentication(success bool, username string, err error) {
	fields := append(l.buildZapFields(),
		zap.Bool("success", success),
		zap.String("username", l.maskSensitiveValue("username", username).(string)),
		zap.Time("timestamp", time.Now()),
	)

	if err != nil {
		fields = append(fields, zap.NamedError("error", err))
		l.Warn("Authentication attempt", fields...)
	} else {
		l.Info("Authentication successful", fields...)
	}
}

// GetZapLogger returns the underlying zap logger for advanced usage
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Close closes the logger and flushes any pending entries
func (l *Logger) Close() error {
	return l.zapLogger.Sync()
}

// WithContext adds context information from the context to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract common context values like request ID, trace ID, etc.
	// Support both typed keys and string keys for backward compatibility
	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		l = l.WithField("request_id", requestID)
	} else if requestID := ctx.Value("request_id"); requestID != nil {
		l = l.WithField("request_id", requestID)
	}

	if traceID := ctx.Value(TraceIDKey); traceID != nil {
		l = l.WithField("trace_id", traceID)
	} else if traceID := ctx.Value("trace_id"); traceID != nil {
		l = l.WithField("trace_id", traceID)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		l = l.WithField("user_id", userID)
	} else if userID := ctx.Value("user_id"); userID != nil {
		l = l.WithField("user_id", userID)
	}

	return l
}
