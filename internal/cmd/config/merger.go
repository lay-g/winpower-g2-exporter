package config

import (
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
)

// Merger 配置合并器接口
type Merger interface {
	// Merge 合并多个配置源，按优先级从低到高
	Merge(configs ...*Config) (*Config, error)

	// MergeWithDefaults 将默认配置与其他配置合并
	MergeWithDefaults(defaults *Config, configs ...*Config) (*Config, error)

	// MergeServer 合并服务器配置
	MergeServer(base, overlay *ServerConfig) *ServerConfig

	// MergeWinPower 合并 WinPower 配置
	MergeWinPower(base, overlay *WinPowerConfig) *WinPowerConfig

	// MergeLogging 合并日志配置
	MergeLogging(base, overlay *LoggingConfig) *LoggingConfig

	// MergeStorage 合并存储配置
	MergeStorage(base, overlay *StorageConfig) *StorageConfig

	// MergeCollector 合并采集器配置
	MergeCollector(base, overlay *CollectorConfig) *CollectorConfig

	// MergeMetrics 合并指标配置
	MergeMetrics(base, overlay *MetricsConfig) *MetricsConfig

	// MergeEnergy 合并能耗配置
	MergeEnergy(base, overlay *EnergyConfig) *EnergyConfig

	// MergeScheduler 合并调度器配置
	MergeScheduler(base, overlay *SchedulerConfig) *SchedulerConfig

	// MergeAuth 合并认证配置
	MergeAuth(base, overlay *AuthConfig) *AuthConfig
}

// ConfigMerger 配置合并器实现
type ConfigMerger struct {
	logger *zap.Logger
}

// NewMerger 创建新的配置合并器
func NewMerger(logger *zap.Logger) *ConfigMerger {
	return &ConfigMerger{
		logger: logger,
	}
}

// Merge 合并多个配置源，按优先级从低到高
func (m *ConfigMerger) Merge(configs ...*Config) (*Config, error) {
	if len(configs) == 0 {
		return nil, fmt.Errorf("at least one configuration is required")
	}

	// 检查是否所有配置都是 nil
	allNil := true
	for _, config := range configs {
		if config != nil {
			allNil = false
			break
		}
	}
	if allNil {
		return nil, fmt.Errorf("at least one non-nil configuration is required")
	}

	m.logger.Debug("Starting configuration merge", zap.Int("config_count", len(configs)))

	// 找到第一个非 nil 配置作为基础
	var result *Config
	for _, config := range configs {
		if config != nil {
			result = &Config{}
			*result = *config
			break
		}
	}

	// 依次合并后续配置
	mergedFirst := false
	for i, config := range configs {
		if config == nil || !mergedFirst {
			if config != nil && !mergedFirst {
				mergedFirst = true
			}
			continue
		}

		if err := m.mergeConfig(result, config); err != nil {
			return nil, fmt.Errorf("failed to merge config at index %d: %w", i, err)
		}

		m.logger.Debug("Configuration merged successfully", zap.Int("step", i))
	}

	m.logger.Info("Configuration merge completed successfully")
	return result, nil
}

// MergeWithDefaults 将默认配置与其他配置合并
func (m *ConfigMerger) MergeWithDefaults(defaults *Config, configs ...*Config) (*Config, error) {
	if defaults == nil {
		return nil, fmt.Errorf("default configuration cannot be nil")
	}

	// 将默认配置作为第一个配置源
	allConfigs := make([]*Config, 1, len(configs)+1)
	allConfigs[0] = defaults
	allConfigs = append(allConfigs, configs...)

	return m.Merge(allConfigs...)
}

// mergeConfig 合并两个配置对象
func (m *ConfigMerger) mergeConfig(base, overlay *Config) error {
	// 合并各个模块配置
	base.Server = *m.MergeServer(&base.Server, &overlay.Server)
	base.WinPower = *m.MergeWinPower(&base.WinPower, &overlay.WinPower)
	base.Logging = *m.MergeLogging(&base.Logging, &overlay.Logging)
	base.Storage = *m.MergeStorage(&base.Storage, &overlay.Storage)
	base.Collector = *m.MergeCollector(&base.Collector, &overlay.Collector)
	base.Metrics = *m.MergeMetrics(&base.Metrics, &overlay.Metrics)
	base.Energy = *m.MergeEnergy(&base.Energy, &overlay.Energy)
	base.Scheduler = *m.MergeScheduler(&base.Scheduler, &overlay.Scheduler)
	base.Auth = *m.MergeAuth(&base.Auth, &overlay.Auth)

	return nil
}

// MergeServer 合并服务器配置
func (m *ConfigMerger) MergeServer(base, overlay *ServerConfig) *ServerConfig {
	if overlay == nil {
		if base == nil {
			return &ServerConfig{}
		}
		return base
	}

	var result ServerConfig
	if base != nil {
		result = *base
	}

	// 使用合并函数合并字段
	if base != nil {
		result.Port = m.mergeInt(base.Port, overlay.Port)
		result.Host = m.mergeString(base.Host, overlay.Host)
		result.ReadTimeout = m.mergeDuration(base.ReadTimeout, overlay.ReadTimeout)
		result.WriteTimeout = m.mergeDuration(base.WriteTimeout, overlay.WriteTimeout)
		result.IdleTimeout = m.mergeDuration(base.IdleTimeout, overlay.IdleTimeout)
		result.EnablePprof = m.mergeBool(base.EnablePprof, overlay.EnablePprof)
	} else {
		// 如果 base 为 nil，直接使用 overlay 的值
		result.Port = overlay.Port
		result.Host = overlay.Host
		result.ReadTimeout = overlay.ReadTimeout
		result.WriteTimeout = overlay.WriteTimeout
		result.IdleTimeout = overlay.IdleTimeout
		result.EnablePprof = overlay.EnablePprof
	}

	return &result
}

// MergeWinPower 合并 WinPower 配置
func (m *ConfigMerger) MergeWinPower(base, overlay *WinPowerConfig) *WinPowerConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.URL = m.mergeString(base.URL, overlay.URL)
	result.Username = m.mergeString(base.Username, overlay.Username)
	result.Password = m.mergeString(base.Password, overlay.Password)
	result.Timeout = m.mergeDuration(base.Timeout, overlay.Timeout)
	result.RetryInterval = m.mergeDuration(base.RetryInterval, overlay.RetryInterval)
	result.MaxRetries = m.mergeInt(base.MaxRetries, overlay.MaxRetries)
	result.SkipSSLVerify = m.mergeBool(base.SkipSSLVerify, overlay.SkipSSLVerify)

	return &result
}

// MergeLogging 合并日志配置
func (m *ConfigMerger) MergeLogging(base, overlay *LoggingConfig) *LoggingConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.Level = m.mergeString(base.Level, overlay.Level)
	result.Format = m.mergeString(base.Format, overlay.Format)
	result.Output = m.mergeString(base.Output, overlay.Output)
	result.Filename = m.mergeString(base.Filename, overlay.Filename)
	result.MaxSize = m.mergeInt(base.MaxSize, overlay.MaxSize)
	result.MaxAge = m.mergeInt(base.MaxAge, overlay.MaxAge)
	result.MaxBackups = m.mergeInt(base.MaxBackups, overlay.MaxBackups)
	result.Compress = m.mergeBoolPtr(base.Compress, overlay.Compress)

	return &result
}

// MergeStorage 合并存储配置
func (m *ConfigMerger) MergeStorage(base, overlay *StorageConfig) *StorageConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.DataDir = m.mergeString(base.DataDir, overlay.DataDir)
	result.SyncWrite = m.mergeBoolPtr(base.SyncWrite, overlay.SyncWrite)

	return &result
}

// MergeCollector 合并采集器配置
func (m *ConfigMerger) MergeCollector(base, overlay *CollectorConfig) *CollectorConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.Enabled = m.mergeBoolPtr(base.Enabled, overlay.Enabled)
	result.Interval = m.mergeDuration(base.Interval, overlay.Interval)
	result.Timeout = m.mergeDuration(base.Timeout, overlay.Timeout)
	result.MaxConcurrent = m.mergeInt(base.MaxConcurrent, overlay.MaxConcurrent)

	return &result
}

// MergeMetrics 合并指标配置
func (m *ConfigMerger) MergeMetrics(base, overlay *MetricsConfig) *MetricsConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.Enabled = m.mergeBoolPtr(base.Enabled, overlay.Enabled)
	result.Path = m.mergeString(base.Path, overlay.Path)
	result.Namespace = m.mergeString(base.Namespace, overlay.Namespace)
	result.Subsystem = m.mergeString(base.Subsystem, overlay.Subsystem)
	result.HelpText = m.mergeString(base.HelpText, overlay.HelpText)

	return &result
}

// MergeEnergy 合并能耗配置
func (m *ConfigMerger) MergeEnergy(base, overlay *EnergyConfig) *EnergyConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.Enabled = m.mergeBoolPtr(base.Enabled, overlay.Enabled)
	result.Interval = m.mergeDuration(base.Interval, overlay.Interval)
	result.Precision = m.mergeInt(base.Precision, overlay.Precision)
	result.StoragePeriod = m.mergeDuration(base.StoragePeriod, overlay.StoragePeriod)

	return &result
}

// MergeScheduler 合并调度器配置
func (m *ConfigMerger) MergeScheduler(base, overlay *SchedulerConfig) *SchedulerConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.Enabled = m.mergeBoolPtr(base.Enabled, overlay.Enabled)
	result.Interval = m.mergeDuration(base.Interval, overlay.Interval)

	return &result
}

// MergeAuth 合并认证配置
func (m *ConfigMerger) MergeAuth(base, overlay *AuthConfig) *AuthConfig {
	if overlay == nil {
		return base
	}

	result := *base

	result.Enabled = m.mergeBool(base.Enabled, overlay.Enabled)
	result.Method = m.mergeString(base.Method, overlay.Method)
	result.TokenURL = m.mergeString(base.TokenURL, overlay.TokenURL)
	result.Username = m.mergeString(base.Username, overlay.Username)
	result.Password = m.mergeString(base.Password, overlay.Password)
	result.Timeout = m.mergeDuration(base.Timeout, overlay.Timeout)
	result.CacheTTL = m.mergeDuration(base.CacheTTL, overlay.CacheTTL)

	return &result
}

// mergeString 合并字符串字段，如果 overlay 不为空则使用 overlay 的值
func (m *ConfigMerger) mergeString(base, overlay string) string {
	if overlay != "" {
		return overlay
	}
	return base
}

// mergeInt 合并整数字段，如果 overlay 不为零值则使用 overlay 的值
func (m *ConfigMerger) mergeInt(base, overlay int) int {
	if overlay != 0 {
		return overlay
	}
	return base
}

// mergeBool 合并布尔字段，总是使用 overlay 的值（因为布尔值有明确的默认值）
func (m *ConfigMerger) mergeBool(base, overlay bool) bool {
	return overlay
}

// mergeBoolPtr 合并 *bool 字段，overlay 为 nil 时使用 base 的值
// 这允许我们区分"未设置"（nil）和"设置为 false"
func (m *ConfigMerger) mergeBoolPtr(base, overlay *bool) *bool {
	if overlay != nil {
		return overlay
	}
	return base
}

// mergeDuration 合并时间间隔字段，如果 overlay 不为零值则使用 overlay 的值
func (m *ConfigMerger) mergeDuration(base, overlay time.Duration) time.Duration {
	if overlay != 0 {
		return overlay
	}
	return base
}

// DeepMerge 深度合并两个配置对象，支持嵌套结构
func (m *ConfigMerger) DeepMerge(base, overlay interface{}) (interface{}, error) {
	baseVal := reflect.ValueOf(base)
	overlayVal := reflect.ValueOf(overlay)

	// 确保都是指针类型
	if baseVal.Kind() != reflect.Ptr || overlayVal.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("both parameters must be pointers")
	}

	if baseVal.IsNil() || overlayVal.IsNil() {
		return nil, fmt.Errorf("both parameters must be non-nil")
	}

	// 获取实际值
	baseElem := baseVal.Elem()
	overlayElem := overlayVal.Elem()

	// 确保类型相同
	if baseElem.Type() != overlayElem.Type() {
		return nil, fmt.Errorf("parameter types must match")
	}

	// 创建结果对象
	result := reflect.New(baseElem.Type())
	resultElem := result.Elem()

	// 复制基础值
	m.deepCopyStruct(baseElem, resultElem)

	// 合并覆盖值
	m.deepMergeStruct(overlayElem, resultElem)

	return result.Interface(), nil
}

// deepCopyStruct 深度复制结构体
func (m *ConfigMerger) deepCopyStruct(src, dst reflect.Value) {
	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		if !field.IsExported() {
			continue
		}

		srcField := src.Field(i)
		dstField := dst.Field(i)

		if dstField.CanSet() {
			dstField.Set(srcField)
		}
	}
}

// deepMergeStruct 深度合并结构体
func (m *ConfigMerger) deepMergeStruct(src, dst reflect.Value) {
	srcType := src.Type()
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		if !field.IsExported() {
			continue
		}

		srcField := src.Field(i)
		dstField := dst.Field(i)

		if !dstField.CanSet() {
			continue
		}

		// 检查源字段是否为零值
		if !m.isZeroValue(srcField) {
			dstField.Set(srcField)
		}
	}
}

// isZeroValue 检查值是否为零值
func (m *ConfigMerger) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Chan, reflect.Func:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理 time.Duration
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			return v.Interface().(time.Duration) == 0
		}
		return false
	default:
		return false
	}
}
