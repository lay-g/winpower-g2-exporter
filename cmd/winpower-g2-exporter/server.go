package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lay-g/winpower-g2-exporter/internal/config"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/spf13/cobra"
)

// NewServerCmd 创建 server 子命令
func NewServerCmd() *cobra.Command {
	var cfgFile string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "启动 HTTP 服务器",
		Long: `启动 WinPower G2 Exporter HTTP 服务器

使用 Ctrl+C 或发送 SIGTERM 信号可以优雅地关闭服务器。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(cfgFile)
		},
	}

	// 添加命令行参数
	cmd.Flags().StringVarP(&cfgFile, "config", "c", "",
		"配置文件路径")

	return cmd
}

// runServer 执行服务器启动逻辑
func runServer(cfgFile string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. 加载配置
	loader := config.NewLoader()
	if cfgFile != "" {
		// 如果指定了配置文件，优先使用
		if err := initConfig(cfgFile); err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}
	}

	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 2. 初始化日志
	logger, err := log.NewLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("初始化日志失败: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("开始启动 WinPower G2 Exporter",
		log.String("version", version),
		log.String("build_time", buildTime),
		log.String("commit_id", commitID))

	// 3. 初始化应用程序
	app, err := initializeApp(ctx, cfg, logger)
	if err != nil {
		logger.Error("初始化应用失败", log.Err(err))
		return fmt.Errorf("初始化应用失败: %w", err)
	}

	// 4. 设置信号处理
	setupSignalHandler(cancel, logger)

	// 5. 启动应用
	logger.Info("WinPower G2 Exporter 启动完成")
	if err := app.Start(ctx); err != nil {
		logger.Error("应用启动失败", log.Err(err))
		return fmt.Errorf("应用启动失败: %w", err)
	}

	// 6. 等待退出
	<-ctx.Done()
	logger.Info("收到退出信号，开始优雅关闭")

	// 7. 优雅关闭
	if err := app.Shutdown(ctx); err != nil {
		logger.Error("应用关闭失败", log.Err(err))
		return fmt.Errorf("应用关闭失败: %w", err)
	}

	logger.Info("WinPower G2 Exporter 已停止")
	return nil
}

// setupSignalHandler 设置信号处理
func setupSignalHandler(cancel context.CancelFunc, logger log.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("收到信号", log.String("signal", sig.String()))
		cancel()
	}()
}
