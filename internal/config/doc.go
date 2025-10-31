// Package config 提供统一的配置管理功能。
//
// Config 模块基于 viper 和 pflag 实现，支持多层级配置文件搜索、环境变量、
// 命令行参数的配置加载机制，并提供可扩展的配置验证接口。
//
// # 配置源优先级
//
// 配置按以下优先级进行合并（高优先级覆盖低优先级）：
//  1. 默认值：代码中定义的默认配置
//  2. 配置文件：从搜索路径找到的配置文件
//  3. 环境变量：WINPOWER_EXPORTER_ 前缀的环境变量
//  4. 命令行参数：通过 pflag 定义的命令行参数
//
// # 配置文件搜索路径
//
// 配置文件按以下优先级顺序搜索（找到第一个存在的配置文件即停止）：
//  1. ./config.yaml
//  2. ./config/config.yaml
//  3. $HOME/config/winpower-exporter/config.yaml
//  4. /etc/winpower-exporter/config.yaml
//
// # 环境变量命名规则
//
// 环境变量使用 WINPOWER_EXPORTER_ 前缀，配置键名中的 . 替换为 _：
//
//	WINPOWER_EXPORTER_SERVER_PORT=9090
//	WINPOWER_EXPORTER_WINPOWER_BASE_URL=https://winpower.example.com
//	WINPOWER_EXPORTER_LOGGING_LEVEL=debug
//
// # 基本使用
//
//	// 创建配置加载器
//	loader := config.NewLoader()
//
//	// 加载配置
//	cfg, err := loader.Load()
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//
//	// 验证配置
//	if err := cfg.Validate(); err != nil {
//	    log.Fatalf("Invalid config: %v", err)
//	}
//
//	// 使用配置
//	server := server.New(cfg.Server)
package config
