package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

// bindFlags 绑定命令行参数
func (l *Loader) bindFlags() error {
	flags := pflag.NewFlagSet("winpower-exporter", pflag.ContinueOnError)

	// Server 配置
	flags.Int("server.port", 8080, "HTTP server port")
	flags.String("server.host", "0.0.0.0", "HTTP server host")
	flags.String("server.mode", "release", "Gin mode (debug|release|test)")
	flags.Duration("server.read-timeout", 10*time.Second, "HTTP read timeout")
	flags.Duration("server.write-timeout", 10*time.Second, "HTTP write timeout")
	flags.Duration("server.idle-timeout", 60*time.Second, "HTTP idle timeout")
	flags.Bool("server.enable-pprof", false, "Enable pprof debug endpoints")
	flags.Duration("server.shutdown-timeout", 30*time.Second, "Graceful shutdown timeout")

	// WinPower 配置
	flags.String("winpower.base-url", "", "WinPower service base URL")
	flags.String("winpower.username", "", "WinPower username")
	flags.String("winpower.password", "", "WinPower password")
	flags.Duration("winpower.timeout", 15*time.Second, "WinPower request timeout")
	flags.Bool("winpower.skip-ssl-verify", false, "Skip SSL certificate verification")
	flags.Duration("winpower.refresh-threshold", 5*time.Minute, "Token refresh threshold")
	flags.String("winpower.user-agent", "Mozilla/5.0 (compatible; WinPower-Exporter/1.0)", "HTTP User-Agent")

	// Storage 配置
	flags.String("storage.data-dir", "./data", "Data directory path")
	flags.Int("storage.file-permissions", 0644, "File permissions (octal)")

	// Scheduler 配置
	flags.Duration("scheduler.collection-interval", 5*time.Second, "Data collection interval")
	flags.Duration("scheduler.graceful-shutdown-timeout", 5*time.Second, "Graceful shutdown timeout")

	// Logging 配置
	flags.String("logging.level", "info", "Log level (debug|info|warn|error|fatal)")
	flags.String("logging.format", "json", "Log format (json|console)")
	flags.String("logging.output", "stdout", "Log output (stdout|stderr|file|both)")
	flags.String("logging.file-path", "", "Log file path")
	flags.Int("logging.max-size", 100, "Max log file size (MB)")
	flags.Int("logging.max-age", 30, "Max log file age (days)")
	flags.Int("logging.max-backups", 10, "Max number of old log files to retain")
	flags.Bool("logging.compress", true, "Compress old log files")
	flags.Bool("logging.development", false, "Enable development mode")
	flags.Bool("logging.enable-caller", false, "Enable caller logging")
	flags.Bool("logging.enable-stacktrace", false, "Enable stacktrace logging")

	// 绑定到 viper（转换短横线为下划线）
	// Parse command line arguments first
	_ = flags.Parse(os.Args[1:])

	// Only bind flags that were actually set on the command line
	// This prevents empty string defaults from overriding environment variables
	changedFlags := make(map[string]bool)
	flags.Visit(func(f *pflag.Flag) {
		changedFlags[f.Name] = true
	})

	// Bind only the changed flags
	flags.VisitAll(func(f *pflag.Flag) {
		if changedFlags[f.Name] {
			if err := l.viper.BindPFlag(f.Name, f); err != nil {
				// 使用标准输出记录错误，因为此模块可能还未初始化logger
				fmt.Fprintf(os.Stderr, "Failed to bind flag %s: %v\n", f.Name, err)
			}
		}
	})

	l.flags = flags
	return nil
}
