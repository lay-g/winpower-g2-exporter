package log

import (
	"context"
)

// Context key 类型定义
type contextKey string

const (
	// contextKeyRequestID 请求 ID
	contextKeyRequestID contextKey = "request_id"

	// contextKeyTraceID 跟踪 ID
	contextKeyTraceID contextKey = "trace_id"

	// contextKeyUserID 用户 ID
	contextKeyUserID contextKey = "user_id"

	// contextKeyComponent 组件名称
	contextKeyComponent contextKey = "component"

	// contextKeyOperation 操作名称
	contextKeyOperation contextKey = "operation"

	// contextKeyLogger 日志器实例
	contextKeyLogger contextKey = "logger"
)

// WithRequestID 在上下文中设置请求 ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, contextKeyRequestID, requestID)
}

// WithTraceID 在上下文中设置跟踪 ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, contextKeyTraceID, traceID)
}

// WithUserID 在上下文中设置用户 ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}

// WithComponent 在上下文中设置组件名称
func WithComponent(ctx context.Context, component string) context.Context {
	return context.WithValue(ctx, contextKeyComponent, component)
}

// WithOperation 在上下文中设置操作名称
func WithOperation(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, contextKeyOperation, operation)
}

// WithLogger 在上下文中存储日志器
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// FromContext 从上下文中获取日志器，如果不存在则返回全局日志器
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(contextKeyLogger).(Logger); ok {
		return logger
	}
	return Default()
}

// extractContextFields 从上下文中提取日志字段
func extractContextFields(ctx context.Context) []Field {
	var fields []Field

	if requestID, ok := ctx.Value(contextKeyRequestID).(string); ok && requestID != "" {
		fields = append(fields, String("request_id", requestID))
	}

	if traceID, ok := ctx.Value(contextKeyTraceID).(string); ok && traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}

	if userID, ok := ctx.Value(contextKeyUserID).(string); ok && userID != "" {
		fields = append(fields, String("user_id", userID))
	}

	if component, ok := ctx.Value(contextKeyComponent).(string); ok && component != "" {
		fields = append(fields, String("component", component))
	}

	if operation, ok := ctx.Value(contextKeyOperation).(string); ok && operation != "" {
		fields = append(fields, String("operation", operation))
	}

	return fields
}
