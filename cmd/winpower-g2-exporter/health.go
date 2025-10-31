package main

import (
	"context"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// HealthService 实现健康检查服务
type HealthService struct {
	collector collector.CollectorInterface
	logger    log.Logger
}

// NewHealthService 创建健康检查服务
func NewHealthService(collector collector.CollectorInterface, logger log.Logger) *HealthService {
	return &HealthService{
		collector: collector,
		logger:    logger,
	}
}

// Check 执行健康检查
func (h *HealthService) Check(ctx context.Context) (status string, details map[string]any) {
	details = make(map[string]any)
	details["timestamp"] = time.Now().Format(time.RFC3339)
	details["version"] = version

	// 简单的健康检查：返回 ok 状态
	status = "ok"
	details["service"] = "running"

	return status, details
}
