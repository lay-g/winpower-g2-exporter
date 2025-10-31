package main

import (
	"context"
	"fmt"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
	"github.com/lay-g/winpower-g2-exporter/internal/config"
	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/metrics"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/scheduler"
	"github.com/lay-g/winpower-g2-exporter/internal/server"
	"github.com/lay-g/winpower-g2-exporter/internal/storage"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// App 应用程序结构体，封装所有模块
type App struct {
	Config    *config.Config
	Logger    log.Logger
	Storage   storage.StorageManager
	WinPower  *winpower.Client
	Energy    *energy.EnergyService
	Collector collector.CollectorInterface
	Metrics   *metrics.MetricsService
	Server    server.Server
	Scheduler scheduler.Scheduler
}

// initializeApp 按依赖顺序初始化所有模块
func initializeApp(ctx context.Context, cfg *config.Config, logger log.Logger) (*App, error) {
	// 1. 初始化存储模块
	// 依赖: 配置模块、日志模块
	storageManager, err := storage.NewFileStorageManager(cfg.Storage, logger)
	if err != nil {
		return nil, fmt.Errorf("初始化存储模块失败: %w", err)
	}

	// 2. 初始化 WinPower 模块
	// 依赖: 配置模块、日志模块
	winpowerClient, err := winpower.NewClient(cfg.WinPower, logger)
	if err != nil {
		return nil, fmt.Errorf("初始化 WinPower 模块失败: %w", err)
	}

	// 3. 初始化电能计算模块
	// 依赖: 配置模块、日志模块、存储模块
	energyService := energy.NewEnergyService(storageManager, logger)

	// 4. 初始化采集器模块
	// 依赖: 配置模块、日志模块、WinPower 模块、电能计算模块
	collectorService, err := collector.NewCollectorService(
		winpowerClient,
		energyService,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("初始化采集器模块失败: %w", err)
	}

	// 5. 初始化指标模块
	// 依赖: 配置模块、日志模块、采集器模块
	metricsConfig := &metrics.MetricsConfig{
		Namespace:           "winpower",
		Subsystem:           "exporter",
		WinPowerHost:        cfg.WinPower.BaseURL,
		EnableMemoryMetrics: true,
	}

	metricsService, err := metrics.NewMetricsService(
		collectorService,
		logger,
		metricsConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("初始化指标模块失败: %w", err)
	}

	// 6. 初始化健康检查服务
	healthService := NewHealthService(collectorService, logger)

	// 7. 初始化服务器模块
	// 依赖: 配置模块、日志模块、指标模块、健康检查服务
	loggerAdapter := NewLoggerAdapter(logger)
	httpServer, err := server.NewHTTPServer(
		cfg.Server,
		loggerAdapter,
		metricsService,
		healthService,
	)
	if err != nil {
		return nil, fmt.Errorf("初始化服务器模块失败: %w", err)
	}

	// 8. 初始化调度器模块
	// 依赖: 配置模块、日志模块、采集器模块
	schedulerService, err := scheduler.NewDefaultScheduler(
		cfg.Scheduler,
		&CollectorSchedulerAdapter{collector: collectorService},
		loggerAdapter,
	)
	if err != nil {
		return nil, fmt.Errorf("初始化调度器模块失败: %w", err)
	}

	return &App{
		Config:    cfg,
		Logger:    logger,
		Storage:   storageManager,
		WinPower:  winpowerClient,
		Energy:    energyService,
		Collector: collectorService,
		Metrics:   metricsService,
		Server:    httpServer,
		Scheduler: schedulerService,
	}, nil
}

// Start 启动应用程序
func (app *App) Start(ctx context.Context) error {
	// 1. 启动 HTTP 服务器（非阻塞）
	go func() {
		if err := app.Server.Start(); err != nil {
			app.Logger.Error("HTTP 服务器启动失败", log.Err(err))
		}
	}()

	// 2. 启动调度器（非阻塞）
	if err := app.Scheduler.Start(ctx); err != nil {
		return fmt.Errorf("启动调度器失败: %w", err)
	}

	return nil
}

// Shutdown 优雅关闭应用程序
func (app *App) Shutdown(ctx context.Context) error {
	var errors []error

	// 按相反顺序关闭模块
	// 1. 停止调度器
	if app.Scheduler != nil {
		if err := app.Scheduler.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("关闭调度器失败: %w", err))
			app.Logger.Error("关闭调度器失败", log.Err(err))
		}
	}

	// 2. 停止服务器
	if app.Server != nil {
		if err := app.Server.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("关闭服务器失败: %w", err))
			app.Logger.Error("关闭服务器失败", log.Err(err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("关闭过程中发生 %d 个错误: %v", len(errors), errors)
	}

	return nil
}
