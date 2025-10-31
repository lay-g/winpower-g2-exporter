package config

import "time"

// setDefaults 设置默认配置值
func (l *Loader) setDefaults() {
	// Server 默认配置
	l.viper.SetDefault("server.port", 8080)
	l.viper.SetDefault("server.host", "0.0.0.0")
	l.viper.SetDefault("server.mode", "release")
	l.viper.SetDefault("server.read_timeout", 10*time.Second)
	l.viper.SetDefault("server.write_timeout", 10*time.Second)
	l.viper.SetDefault("server.idle_timeout", 60*time.Second)
	l.viper.SetDefault("server.enable_pprof", false)
	l.viper.SetDefault("server.shutdown_timeout", 30*time.Second)

	// WinPower 默认配置
	l.viper.SetDefault("winpower.timeout", 15*time.Second)
	l.viper.SetDefault("winpower.skip_ssl_verify", false)
	l.viper.SetDefault("winpower.refresh_threshold", 5*time.Minute)
	l.viper.SetDefault("winpower.user_agent", "Mozilla/5.0 (compatible; WinPower-Exporter/1.0)")

	// Storage 默认配置
	l.viper.SetDefault("storage.data_dir", "./data")
	l.viper.SetDefault("storage.file_permissions", 0644)

	// Scheduler 默认配置
	l.viper.SetDefault("scheduler.collection_interval", 5*time.Second)
	l.viper.SetDefault("scheduler.graceful_shutdown_timeout", 5*time.Second)

	// Logging 默认配置
	l.viper.SetDefault("logging.level", "info")
	l.viper.SetDefault("logging.format", "json")
	l.viper.SetDefault("logging.output", "stdout")
	l.viper.SetDefault("logging.file_path", "")
	l.viper.SetDefault("logging.max_size", 100)   // 100 MB
	l.viper.SetDefault("logging.max_age", 30)     // 30 days
	l.viper.SetDefault("logging.max_backups", 10) // 10 files
	l.viper.SetDefault("logging.compress", true)  // 压缩旧日志
	l.viper.SetDefault("logging.development", false)
	l.viper.SetDefault("logging.enable_caller", false)
	l.viper.SetDefault("logging.enable_stacktrace", false)
}
