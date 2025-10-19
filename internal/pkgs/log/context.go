package log

import (
	"context"
	"time"
)

// 上下文字段键常量
const (
	// 请求追踪相关
	RequestIDKey = "request_id"
	TraceIDKey   = "trace_id"
	SpanIDKey    = "span_id"

	// 用户相关
	UserIDKey    = "user_id"
	UsernameKey  = "username"
	UserAgentKey = "user_agent"

	// 系统相关
	HostnameKey = "hostname"
	ServiceKey  = "service"
	VersionKey  = "version"
	InstanceKey = "instance"

	// 业务相关
	DeviceIDKey = "device_id"
	ActionKey   = "action"
	ResourceKey = "resource"
	ResultKey   = "result"

	// 性能相关
	DurationKey = "duration"
	LatencyKey  = "latency"
	SizeKey     = "size"

	// 错误相关
	ErrorCodeKey = "error_code"
	ErrorTypeKey = "error_type"
	RetryKey     = "retry"
)

// 自定义上下文键类型，避免字符串冲突
type contextKey string

// 内部上下文键
var (
	loggerKey    = contextKey("logger")
	fieldsKey    = contextKey("fields")
	requestIDKey = contextKey("request_id")
	traceIDKey   = contextKey("trace_id")
	userIDKey    = contextKey("user_id")
)

// WithRequestID 在上下文中设置请求ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithTraceID 在上下文中设置追踪ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithSpanID 在上下文中设置Span ID
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, contextKey("span_id"), spanID)
}

// WithUserID 在上下文中设置用户ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithUsername 在上下文中设置用户名
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, contextKey("username"), username)
}

// WithUserAgent 在上下文中设置用户代理
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, contextKey("user_agent"), userAgent)
}

// WithService 在上下文中设置服务名
func WithService(ctx context.Context, service string) context.Context {
	return context.WithValue(ctx, contextKey("service"), service)
}

// WithVersion 在上下文中设置版本
func WithVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, contextKey("version"), version)
}

// WithDeviceID 在上下文中设置设备ID
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, contextKey("device_id"), deviceID)
}

// WithAction 在上下文中设置操作
func WithAction(ctx context.Context, action string) context.Context {
	return context.WithValue(ctx, contextKey("action"), action)
}

// WithDuration 在上下文中设置持续时间
func WithDuration(ctx context.Context, duration time.Duration) context.Context {
	return context.WithValue(ctx, contextKey("duration"), duration)
}

// WithLogger 在上下文中设置日志器
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// WithFields 在上下文中设置字段
func WithFields(ctx context.Context, fields ...Field) context.Context {
	existingFields, _ := ctx.Value(fieldsKey).([]Field)
	allFields := append(existingFields, fields...)
	return context.WithValue(ctx, fieldsKey, allFields)
}

// FromContext 从上下文中获取日志器
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return nil
}

// RequestIDFromContext 从上下文中获取请求ID
func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// TraceIDFromContext 从上下文中获取追踪ID
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// UserIDFromContext 从上下文中获取用户ID
func UserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDKey).(string); ok {
		return userID
	}
	return ""
}

// FieldsFromContext 从上下文中获取字段
func FieldsFromContext(ctx context.Context) []Field {
	if fields, ok := ctx.Value(fieldsKey).([]Field); ok {
		return fields
	}
	return nil
}

// extractContextFields 从上下文中提取常用字段
func extractContextFields(ctx context.Context) []Field {
	var fields []Field

	// 提取请求追踪相关字段
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		fields = append(fields, String(RequestIDKey, requestID))
	}

	if traceID := TraceIDFromContext(ctx); traceID != "" {
		fields = append(fields, String(TraceIDKey, traceID))
	}

	if spanID, ok := ctx.Value(contextKey("span_id")).(string); ok && spanID != "" {
		fields = append(fields, String(SpanIDKey, spanID))
	}

	// 提取用户相关字段
	if userID := UserIDFromContext(ctx); userID != "" {
		fields = append(fields, String(UserIDKey, userID))
	}

	if username, ok := ctx.Value(contextKey("username")).(string); ok && username != "" {
		fields = append(fields, String(UsernameKey, username))
	}

	if userAgent, ok := ctx.Value(contextKey("user_agent")).(string); ok && userAgent != "" {
		fields = append(fields, String(UserAgentKey, userAgent))
	}

	// 提取系统相关字段
	if service, ok := ctx.Value(contextKey("service")).(string); ok && service != "" {
		fields = append(fields, String(ServiceKey, service))
	}

	if version, ok := ctx.Value(contextKey("version")).(string); ok && version != "" {
		fields = append(fields, String(VersionKey, version))
	}

	if deviceID, ok := ctx.Value(contextKey("device_id")).(string); ok && deviceID != "" {
		fields = append(fields, String(DeviceIDKey, deviceID))
	}

	// 提取业务相关字段
	if action, ok := ctx.Value(contextKey("action")).(string); ok && action != "" {
		fields = append(fields, String(ActionKey, action))
	}

	if duration, ok := ctx.Value(contextKey("duration")).(time.Duration); ok {
		fields = append(fields, Duration(DurationKey, duration))
	}

	// 提取自定义字段
	if customFields := FieldsFromContext(ctx); len(customFields) > 0 {
		fields = append(fields, customFields...)
	}

	return fields
}

// LoggerFromContext 从上下文中获取日志器，如果没有则创建带有上下文字段的日志器
func LoggerFromContext(ctx context.Context, defaultLogger Logger) Logger {
	// 首先尝试从上下文中获取日志器
	if logger := FromContext(ctx); logger != nil {
		return logger
	}

	// 如果没有，使用默认日志器并添加上下文字段
	if defaultLogger != nil {
		return defaultLogger.WithContext(ctx)
	}

	return nil
}

// DebugWithContext 在上下文中记录调试日志
func DebugWithContext(ctx context.Context, msg string, fields ...Field) {
	if logger := FromContext(ctx); logger != nil {
		logger.WithContext(ctx).Debug(msg, fields...)
	}
}

// InfoWithContext 在上下文中记录信息日志
func InfoWithContext(ctx context.Context, msg string, fields ...Field) {
	if logger := FromContext(ctx); logger != nil {
		logger.WithContext(ctx).Info(msg, fields...)
	}
}

// WarnWithContext 在上下文中记录警告日志
func WarnWithContext(ctx context.Context, msg string, fields ...Field) {
	if logger := FromContext(ctx); logger != nil {
		logger.WithContext(ctx).Warn(msg, fields...)
	}
}

// ErrorWithContext 在上下文中记录错误日志
func ErrorWithContext(ctx context.Context, msg string, fields ...Field) {
	if logger := FromContext(ctx); logger != nil {
		logger.WithContext(ctx).Error(msg, fields...)
	}
}

// FatalWithContext 在上下文中记录致命错误日志
func FatalWithContext(ctx context.Context, msg string, fields ...Field) {
	if logger := FromContext(ctx); logger != nil {
		logger.WithContext(ctx).Fatal(msg, fields...)
	}
}
