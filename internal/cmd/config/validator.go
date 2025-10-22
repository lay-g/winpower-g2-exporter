package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Validator 配置验证器接口
type Validator interface {
	// Validate 验证配置的有效性
	Validate(config *Config) error

	// ValidateServer 验证服务器配置
	ValidateServer(config *ServerConfig) error

	// ValidateWinPower 验证 WinPower 配置
	ValidateWinPower(config *WinPowerConfig) error

	// ValidateLogging 验证日志配置
	ValidateLogging(config *LoggingConfig) error

	// ValidateStorage 验证存储配置
	ValidateStorage(config *StorageConfig) error

	// ValidateCollector 验证采集器配置
	ValidateCollector(config *CollectorConfig) error

	// ValidateMetrics 验证指标配置
	ValidateMetrics(config *MetricsConfig) error

	// ValidateEnergy 验证能耗配置
	ValidateEnergy(config *EnergyConfig) error

	// ValidateScheduler 验证调度器配置
	ValidateScheduler(config *SchedulerConfig) error

	// ValidateAuth 验证认证配置
	ValidateAuth(config *AuthConfig) error
}

// ConfigValidator 配置验证器实现
type ConfigValidator struct {
	logger *zap.Logger
}

// NewValidator 创建新的配置验证器
func NewValidator(logger *zap.Logger) *ConfigValidator {
	return &ConfigValidator{
		logger: logger,
	}
}

// Validate 验证配置的有效性
func (v *ConfigValidator) Validate(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	v.logger.Debug("Starting configuration validation")

	// 验证各个模块配置
	validators := []struct {
		name string
		fn   func() error
	}{
		{"server", func() error { return v.ValidateServer(&config.Server) }},
		{"winpower", func() error { return v.ValidateWinPower(&config.WinPower) }},
		{"logging", func() error { return v.ValidateLogging(&config.Logging) }},
		{"storage", func() error { return v.ValidateStorage(&config.Storage) }},
		{"collector", func() error { return v.ValidateCollector(&config.Collector) }},
		{"metrics", func() error { return v.ValidateMetrics(&config.Metrics) }},
		{"energy", func() error { return v.ValidateEnergy(&config.Energy) }},
		{"scheduler", func() error { return v.ValidateScheduler(&config.Scheduler) }},
		{"auth", func() error { return v.ValidateAuth(&config.Auth) }},
	}

	for _, validator := range validators {
		if err := validator.fn(); err != nil {
			return fmt.Errorf("%s configuration validation failed: %w", validator.name, err)
		}
		v.logger.Debug("Module configuration validated successfully", zap.String("module", validator.name))
	}

	v.logger.Info("Configuration validation completed successfully")
	return nil
}

// ValidateServer 验证服务器配置
func (v *ConfigValidator) ValidateServer(config *ServerConfig) error {
	// 验证主机地址
	if config.Host == "" {
		return fmt.Errorf("server host cannot be empty")
	}

	// 验证主机地址格式
	if config.Host != "0.0.0.0" && config.Host != "localhost" {
		if net.ParseIP(config.Host) == nil {
			// 尝试解析为域名
			if _, err := net.LookupHost(config.Host); err != nil {
				return fmt.Errorf("invalid server host address: %s", config.Host)
			}
		}
	}

	// 验证端口号
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got: %d", config.Port)
	}

	// 验证超时时间
	if config.ReadTimeout <= 0 {
		return fmt.Errorf("server read timeout must be positive, got: %v", config.ReadTimeout)
	}
	if config.WriteTimeout <= 0 {
		return fmt.Errorf("server write timeout must be positive, got: %v", config.WriteTimeout)
	}
	if config.IdleTimeout <= 0 {
		return fmt.Errorf("server idle timeout must be positive, got: %v", config.IdleTimeout)
	}

	return nil
}

// ValidateWinPower 验证 WinPower 配置
func (v *ConfigValidator) ValidateWinPower(config *WinPowerConfig) error {
	// 验证 URL
	if config.URL == "" {
		return fmt.Errorf("winpower URL cannot be empty")
	}

	if err := v.validateURL(config.URL, "winpower"); err != nil {
		return err
	}

	// 验证用户名
	if config.Username == "" {
		return fmt.Errorf("winpower username cannot be empty")
	}

	if len(config.Username) > 255 {
		return fmt.Errorf("winpower username too long (max 255 characters)")
	}

	// 验证密码
	if config.Password == "" {
		return fmt.Errorf("winpower password cannot be empty")
	}

	// 验证超时时间
	if config.Timeout <= 0 {
		return fmt.Errorf("winpower timeout must be positive, got: %v", config.Timeout)
	}

	// 验证重试间隔
	if config.RetryInterval <= 0 {
		return fmt.Errorf("winpower retry interval must be positive, got: %v", config.RetryInterval)
	}

	// 验证最大重试次数
	if config.MaxRetries < 0 {
		return fmt.Errorf("winpower max retries cannot be negative, got: %d", config.MaxRetries)
	}

	if config.MaxRetries > 10 {
		v.logger.Warn("WinPower max retries is quite high", zap.Int("max_retries", config.MaxRetries))
	}

	return nil
}

// ValidateLogging 验证日志配置
func (v *ConfigValidator) ValidateLogging(config *LoggingConfig) error {
	// 验证日志级别
	validLevels := []string{"debug", "info", "warn", "error"}
	if !v.contains(validLevels, config.Level) {
		return fmt.Errorf("invalid log level: %s, valid levels: %v", config.Level, validLevels)
	}

	// 验证日志格式
	validFormats := []string{"json", "console"}
	if !v.contains(validFormats, config.Format) {
		return fmt.Errorf("invalid log format: %s, valid formats: %v", config.Format, validFormats)
	}

	// 验证输出方式
	validOutputs := []string{"stdout", "stderr", "file"}
	if !v.contains(validOutputs, config.Output) {
		return fmt.Errorf("invalid log output: %s, valid outputs: %v", config.Output, validOutputs)
	}

	// 如果输出到文件，验证文件路径
	if config.Output == "file" {
		if config.Filename == "" {
			return fmt.Errorf("log filename cannot be empty when output is file")
		}

		// 验证文件目录是否可创建
		dir := filepath.Dir(config.Filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create log directory: %w", err)
		}
	}

	// 验证日志轮转参数
	if config.MaxSize <= 0 {
		return fmt.Errorf("log max size must be positive, got: %d", config.MaxSize)
	}
	if config.MaxAge < 0 {
		return fmt.Errorf("log max age cannot be negative, got: %d", config.MaxAge)
	}
	if config.MaxBackups < 0 {
		return fmt.Errorf("log max backups cannot be negative, got: %d", config.MaxBackups)
	}

	return nil
}

// ValidateStorage 验证存储配置
func (v *ConfigValidator) ValidateStorage(config *StorageConfig) error {
	// 验证数据目录
	if config.DataDir == "" {
		return fmt.Errorf("storage data directory cannot be empty")
	}

	// 验证数据目录路径
	if !filepath.IsAbs(config.DataDir) {
		// 转换为绝对路径
		absPath, err := filepath.Abs(config.DataDir)
		if err != nil {
			return fmt.Errorf("cannot get absolute path for data directory: %w", err)
		}
		config.DataDir = absPath
	}

	// 检查目录是否可创建且有写权限
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return fmt.Errorf("cannot create data directory: %w", err)
	}

	// 测试写权限
	testFile := filepath.Join(config.DataDir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("data directory is not writable: %w", err)
	}
	_ = os.Remove(testFile) // 清理测试文件，忽略错误

	return nil
}

// ValidateCollector 验证采集器配置
func (v *ConfigValidator) ValidateCollector(config *CollectorConfig) error {
	// 验证采集间隔
	if config.Interval <= 0 {
		return fmt.Errorf("collector interval must be positive, got: %v", config.Interval)
	}

	// 采集间隔不应太短（避免过于频繁的请求）
	if config.Interval < time.Second {
		v.logger.Warn("Collector interval is very short, may cause high load",
			zap.Duration("interval", config.Interval))
	}

	// 验证超时时间
	if config.Timeout <= 0 {
		return fmt.Errorf("collector timeout must be positive, got: %v", config.Timeout)
	}

	// 超时时间应小于采集间隔
	if config.Timeout >= config.Interval {
		return fmt.Errorf("collector timeout (%v) must be less than interval (%v)", config.Timeout, config.Interval)
	}

	// 验证最大并发数
	if config.MaxConcurrent <= 0 {
		return fmt.Errorf("collector max concurrent must be positive, got: %d", config.MaxConcurrent)
	}

	if config.MaxConcurrent > 50 {
		v.logger.Warn("Collector max concurrent is quite high", zap.Int("max_concurrent", config.MaxConcurrent))
	}

	return nil
}

// ValidateMetrics 验证指标配置
func (v *ConfigValidator) ValidateMetrics(config *MetricsConfig) error {
	// 验证指标路径
	if config.Path == "" {
		return fmt.Errorf("metrics path cannot be empty")
	}

	if !strings.HasPrefix(config.Path, "/") {
		return fmt.Errorf("metrics path must start with '/', got: %s", config.Path)
	}

	// 验证命名空间
	if config.Namespace == "" {
		return fmt.Errorf("metrics namespace cannot be empty")
	}

	// 验证命名空间格式（只能包含字母、数字和下划线）
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(config.Namespace) {
		return fmt.Errorf("invalid metrics namespace format: %s", config.Namespace)
	}

	// 验证子系统格式
	if config.Subsystem != "" {
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(config.Subsystem) {
			return fmt.Errorf("invalid metrics subsystem format: %s", config.Subsystem)
		}
	}

	// 验证帮助文本
	if config.HelpText == "" {
		return fmt.Errorf("metrics help text cannot be empty")
	}

	return nil
}

// ValidateEnergy 验证能耗配置
func (v *ConfigValidator) ValidateEnergy(config *EnergyConfig) error {
	// 验证计算间隔
	if config.Interval <= 0 {
		return fmt.Errorf("energy interval must be positive, got: %v", config.Interval)
	}

	// 验证精度
	if config.Precision < 0 || config.Precision > 10 {
		return fmt.Errorf("energy precision must be between 0 and 10, got: %d", config.Precision)
	}

	// 验证存储周期
	if config.StoragePeriod <= 0 {
		return fmt.Errorf("energy storage period must be positive, got: %v", config.StoragePeriod)
	}

	// 存储周期应该是计算间隔的整数倍
	if config.StoragePeriod.Seconds() < config.Interval.Seconds() {
		v.logger.Warn("Energy storage period is shorter than calculation interval",
			zap.Duration("storage_period", config.StoragePeriod),
			zap.Duration("interval", config.Interval))
	}

	return nil
}

// ValidateScheduler 验证调度器配置
func (v *ConfigValidator) ValidateScheduler(config *SchedulerConfig) error {
	// 验证调度间隔
	if config.Interval <= 0 {
		return fmt.Errorf("scheduler interval must be positive, got: %v", config.Interval)
	}

	// 调度间隔不应太短（避免过于频繁的调度）
	if config.Interval < time.Second {
		v.logger.Warn("Scheduler interval is very short, may cause high CPU usage",
			zap.Duration("interval", config.Interval))
	}

	return nil
}

// ValidateAuth 验证认证配置
func (v *ConfigValidator) ValidateAuth(config *AuthConfig) error {
	// 如果认证未启用，跳过验证
	if !config.Enabled {
		v.logger.Info("Authentication is disabled")
		return nil
	}

	// 验证认证方法
	validMethods := []string{"token", "basic", "oauth2"}
	if !v.contains(validMethods, config.Method) {
		return fmt.Errorf("invalid auth method: %s, valid methods: %v", config.Method, validMethods)
	}

	// 根据认证方法验证必需的配置
	switch config.Method {
	case "token":
		if config.TokenURL == "" {
			return fmt.Errorf("token URL cannot be empty for token authentication")
		}
		if err := v.validateURL(config.TokenURL, "token URL"); err != nil {
			return err
		}
		if config.Username == "" {
			return fmt.Errorf("username cannot be empty for token authentication")
		}
		if config.Password == "" {
			return fmt.Errorf("password cannot be empty for token authentication")
		}

	case "basic":
		if config.Username == "" {
			return fmt.Errorf("username cannot be empty for basic authentication")
		}
		if config.Password == "" {
			return fmt.Errorf("password cannot be empty for basic authentication")
		}

	case "oauth2":
		if config.TokenURL == "" {
			return fmt.Errorf("token URL cannot be empty for OAuth2 authentication")
		}
		if err := v.validateURL(config.TokenURL, "OAuth2 token URL"); err != nil {
			return err
		}
	}

	// 验证超时时间
	if config.Timeout <= 0 {
		return fmt.Errorf("auth timeout must be positive, got: %v", config.Timeout)
	}

	// 验证缓存TTL
	if config.CacheTTL <= 0 {
		return fmt.Errorf("auth cache TTL must be positive, got: %v", config.CacheTTL)
	}

	return nil
}

// validateURL 验证URL格式
func (v *ConfigValidator) validateURL(rawURL, context string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid %s format: %w", context, err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("%s must include scheme (http/https)", context)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("%s must include host", context)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("%s scheme must be http or https, got: %s", context, parsedURL.Scheme)
	}

	return nil
}

// contains 检查字符串是否在切片中
func (v *ConfigValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}