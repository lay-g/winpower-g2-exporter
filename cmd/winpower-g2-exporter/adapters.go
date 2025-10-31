package main

import (
	"context"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"go.uber.org/zap"
)

// LoggerAdapter 适配器，将 log.Logger 转换为其他模块需要的接口
type LoggerAdapter struct {
	logger log.Logger
}

// NewLoggerAdapter 创建日志适配器
func NewLoggerAdapter(logger log.Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: logger}
}

// Info 实现简单的 Info 接口
func (l *LoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, convertFields(keysAndValues...)...)
}

// Error 实现简单的 Error 接口
func (l *LoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	l.logger.Error(msg, convertFields(keysAndValues...)...)
}

// Warn 实现简单的 Warn 接口
func (l *LoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	l.logger.Warn(msg, convertFields(keysAndValues...)...)
}

// Debug 实现简单的 Debug 接口
func (l *LoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	l.logger.Debug(msg, convertFields(keysAndValues...)...)
}

// convertFields 将 key-value 对转换为 zap.Field
func convertFields(keysAndValues ...interface{}) []log.Field {
	if len(keysAndValues)%2 != 0 {
		// 如果参数个数是奇数，忽略最后一个
		keysAndValues = keysAndValues[:len(keysAndValues)-1]
	}

	fields := make([]log.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, keysAndValues[i+1]))
	}
	return fields
}

// CollectorSchedulerAdapter 适配器，适配 collector 到 scheduler 需要的接口
type CollectorSchedulerAdapter struct {
	collector *collector.CollectorService
}

// CollectDeviceData 实现 scheduler.CollectorInterface
func (c *CollectorSchedulerAdapter) CollectDeviceData(ctx context.Context) (*scheduler.CollectionResult, error) {
	result, err := c.collector.CollectDeviceData(ctx)
	if err != nil {
		return &scheduler.CollectionResult{
			Success:      false,
			DeviceCount:  0,
			ErrorMessage: err.Error(),
		}, err
	}

	return &scheduler.CollectionResult{
		Success:      result.Success,
		DeviceCount:  result.DeviceCount,
		ErrorMessage: result.ErrorMessage,
	}, nil
}
