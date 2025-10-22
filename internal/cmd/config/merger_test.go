package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestConfigMerger_Merge(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("single config", func(t *testing.T) {
		config := &Config{
			Server: ServerConfig{
				Port: 8080,
				Host: "localhost",
			},
		}

		result, err := merger.Merge(config)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 8080, result.Server.Port, "Port should be preserved")
		assert.Equal(t, "localhost", result.Server.Host, "Host should be preserved")
	})

	t.Run("multiple configs", func(t *testing.T) {
		config1 := &Config{
			Server: ServerConfig{
				Port:        8080,
				Host:        "localhost",
				ReadTimeout: 30 * time.Second,
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
		}

		config2 := &Config{
			Server: ServerConfig{
				Port: 9090,
				Host: "0.0.0.0",
			},
			Logging: LoggingConfig{
				Level: "debug",
			},
			WinPower: WinPowerConfig{
				URL:      "http://example.com",
				Username: "user",
				Password: "pass",
			},
		}

		config3 := &Config{
			Server: ServerConfig{
				EnablePprof: true,
			},
			WinPower: WinPowerConfig{
				Timeout: 60 * time.Second,
			},
		}

		result, err := merger.Merge(config1, config2, config3)
		require.NoError(t, err)
		require.NotNil(t, result)

		// 验证合并结果：后面的配置覆盖前面的配置
		assert.Equal(t, 9090, result.Server.Port, "Port should be from config2")
		assert.Equal(t, "0.0.0.0", result.Server.Host, "Host should be from config2")
		assert.Equal(t, 30*time.Second, result.Server.ReadTimeout, "ReadTimeout should be from config1")
		assert.Equal(t, true, result.Server.EnablePprof, "EnablePprof should be from config3")

		assert.Equal(t, "debug", result.Logging.Level, "Log level should be from config2")
		assert.Equal(t, "json", result.Logging.Format, "Log format should be from config1")

		assert.Equal(t, "http://example.com", result.WinPower.URL, "WinPower URL should be from config2")
		assert.Equal(t, "user", result.WinPower.Username, "WinPower username should be from config2")
		assert.Equal(t, "pass", result.WinPower.Password, "WinPower password should be from config2")
		assert.Equal(t, 60*time.Second, result.WinPower.Timeout, "WinPower timeout should be from config3")
	})

	t.Run("nil config in middle", func(t *testing.T) {
		config1 := &Config{
			Server: ServerConfig{Port: 8080},
		}

		config3 := &Config{
			Server: ServerConfig{Host: "localhost"},
		}

		result, err := merger.Merge(config1, nil, config3)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 8080, result.Server.Port, "Port should be from config1")
		assert.Equal(t, "localhost", result.Server.Host, "Host should be from config3")
	})

	t.Run("no configs", func(t *testing.T) {
		result, err := merger.Merge()
		assert.Error(t, err, "Should return error when no configs provided")
		assert.Nil(t, result, "Result should be nil when no configs provided")
		assert.Contains(t, err.Error(), "at least one configuration is required")
	})

	t.Run("all nil configs", func(t *testing.T) {
		result, err := merger.Merge(nil, nil, nil)
		assert.Error(t, err, "Should return error when all configs are nil")
		assert.Nil(t, result, "Result should be nil when all configs are nil")
		assert.Contains(t, err.Error(), "at least one non-nil configuration is required")
	})
}

func TestConfigMerger_MergeWithDefaults(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("merge with defaults", func(t *testing.T) {
		defaults := &Config{
			Server: ServerConfig{
				Port:        8080,
				Host:        "localhost",
				ReadTimeout: 30 * time.Second,
				WriteTimeout: 30 * time.Second,
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
		}

		overlay := &Config{
			Server: ServerConfig{
				Port: 9090,
			},
			Logging: LoggingConfig{
				Level: "debug",
			},
		}

		result, err := merger.MergeWithDefaults(defaults, overlay)
		require.NoError(t, err)
		require.NotNil(t, result)

		// 验证默认值被保留
		assert.Equal(t, "localhost", result.Server.Host, "Host should be from defaults")
		assert.Equal(t, 30*time.Second, result.Server.ReadTimeout, "ReadTimeout should be from defaults")
		assert.Equal(t, "json", result.Logging.Format, "Log format should be from defaults")

		// 验证覆盖值被应用
		assert.Equal(t, 9090, result.Server.Port, "Port should be from overlay")
		assert.Equal(t, "debug", result.Logging.Level, "Log level should be from overlay")
	})

	t.Run("nil defaults", func(t *testing.T) {
		overlay := &Config{
			Server: ServerConfig{Port: 9090},
		}

		result, err := merger.MergeWithDefaults(nil, overlay)
		assert.Error(t, err, "Should return error when defaults is nil")
		assert.Nil(t, result, "Result should be nil when defaults is nil")
		assert.Contains(t, err.Error(), "default configuration cannot be nil")
	})

	t.Run("only defaults", func(t *testing.T) {
		defaults := &Config{
			Server: ServerConfig{
				Port: 8080,
				Host: "localhost",
			},
		}

		result, err := merger.MergeWithDefaults(defaults)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, 8080, result.Server.Port, "Port should be from defaults")
		assert.Equal(t, "localhost", result.Server.Host, "Host should be from defaults")
	})
}

func TestConfigMerger_MergeServer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("normal merge", func(t *testing.T) {
		base := &ServerConfig{
			Port:        8080,
			Host:        "localhost",
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			EnablePprof: false,
		}

		overlay := &ServerConfig{
			Port:        9090,
			Host:        "0.0.0.0",
			EnablePprof: true,
		}

		result := merger.MergeServer(base, overlay)

		assert.Equal(t, 9090, result.Port, "Port should be from overlay")
		assert.Equal(t, "0.0.0.0", result.Host, "Host should be from overlay")
		assert.Equal(t, 30*time.Second, result.ReadTimeout, "ReadTimeout should be from base")
		assert.Equal(t, 30*time.Second, result.WriteTimeout, "WriteTimeout should be from base")
		assert.Equal(t, 60*time.Second, result.IdleTimeout, "IdleTimeout should be from base")
		assert.Equal(t, true, result.EnablePprof, "EnablePprof should be from overlay")
	})

	t.Run("nil overlay", func(t *testing.T) {
		base := &ServerConfig{
			Port: 8080,
			Host: "localhost",
		}

		result := merger.MergeServer(base, nil)
		assert.Equal(t, base, result, "Should return base when overlay is nil")
	})

	t.Run("nil base", func(t *testing.T) {
		overlay := &ServerConfig{
			Port: 9090,
			Host: "0.0.0.0",
		}

		result := merger.MergeServer(nil, overlay)
		assert.NotNil(t, result, "Should not return nil when base is nil")
		assert.Equal(t, 9090, result.Port, "Port should be from overlay")
		assert.Equal(t, "0.0.0.0", result.Host, "Host should be from overlay")
	})
}

func TestConfigMerger_MergeWinPower(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &WinPowerConfig{
		URL:           "http://base.example.com",
		Username:      "baseuser",
		Password:      "basepass",
		Timeout:       30 * time.Second,
		RetryInterval: 5 * time.Second,
		MaxRetries:    3,
		SkipSSLVerify: false,
	}

	overlay := &WinPowerConfig{
		URL:          "http://overlay.example.com",
		Username:     "overlayuser",
		MaxRetries:   5,
		SkipSSLVerify: true,
	}

	result := merger.MergeWinPower(base, overlay)

	assert.Equal(t, "http://overlay.example.com", result.URL, "URL should be from overlay")
	assert.Equal(t, "overlayuser", result.Username, "Username should be from overlay")
	assert.Equal(t, "basepass", result.Password, "Password should be from base")
	assert.Equal(t, 30*time.Second, result.Timeout, "Timeout should be from base")
	assert.Equal(t, 5*time.Second, result.RetryInterval, "RetryInterval should be from base")
	assert.Equal(t, 5, result.MaxRetries, "MaxRetries should be from overlay")
	assert.Equal(t, true, result.SkipSSLVerify, "SkipSSLVerify should be from overlay")
}

func TestConfigMerger_MergeLogging(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		Filename:   "/var/log/app.log",
		MaxSize:    100,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
	}

	overlay := &LoggingConfig{
		Level:      "debug",
		Output:     "file",
		MaxSize:    200,
		Compress:   false,
	}

	result := merger.MergeLogging(base, overlay)

	assert.Equal(t, "debug", result.Level, "Level should be from overlay")
	assert.Equal(t, "json", result.Format, "Format should be from base")
	assert.Equal(t, "file", result.Output, "Output should be from overlay")
	assert.Equal(t, "/var/log/app.log", result.Filename, "Filename should be from base")
	assert.Equal(t, 200, result.MaxSize, "MaxSize should be from overlay")
	assert.Equal(t, 7, result.MaxAge, "MaxAge should be from base")
	assert.Equal(t, 3, result.MaxBackups, "MaxBackups should be from base")
	assert.Equal(t, false, result.Compress, "Compress should be from overlay")
}

func TestConfigMerger_MergeStorage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &StorageConfig{
		DataDir:   "/data",
		SyncWrite: true,
	}

	overlay := &StorageConfig{
		DataDir:   "/new/data",
		SyncWrite: false,
	}

	result := merger.MergeStorage(base, overlay)

	assert.Equal(t, "/new/data", result.DataDir, "DataDir should be from overlay")
	assert.Equal(t, false, result.SyncWrite, "SyncWrite should be from overlay")
}

func TestConfigMerger_MergeCollector(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &CollectorConfig{
		Enabled:       true,
		Interval:      5 * time.Second,
		Timeout:       3 * time.Second,
		MaxConcurrent: 5,
	}

	overlay := &CollectorConfig{
		Enabled:  false,
		Interval: 10 * time.Second,
		Timeout:  6 * time.Second,
	}

	result := merger.MergeCollector(base, overlay)

	assert.Equal(t, false, result.Enabled, "Enabled should be from overlay")
	assert.Equal(t, 10*time.Second, result.Interval, "Interval should be from overlay")
	assert.Equal(t, 6*time.Second, result.Timeout, "Timeout should be from overlay")
	assert.Equal(t, 5, result.MaxConcurrent, "MaxConcurrent should be from base")
}

func TestConfigMerger_MergeMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &MetricsConfig{
		Enabled:   true,
		Path:      "/metrics",
		Namespace: "winpower",
		Subsystem: "exporter",
		HelpText:  "WinPower Exporter Metrics",
	}

	overlay := &MetricsConfig{
		Enabled:   false,
		Path:      "/newmetrics",
		Namespace: "custom",
		HelpText:  "Custom Metrics",
	}

	result := merger.MergeMetrics(base, overlay)

	assert.Equal(t, false, result.Enabled, "Enabled should be from overlay")
	assert.Equal(t, "/newmetrics", result.Path, "Path should be from overlay")
	assert.Equal(t, "custom", result.Namespace, "Namespace should be from overlay")
	assert.Equal(t, "exporter", result.Subsystem, "Subsystem should be from base")
	assert.Equal(t, "Custom Metrics", result.HelpText, "HelpText should be from overlay")
}

func TestConfigMerger_MergeEnergy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &EnergyConfig{
		Enabled:       true,
		Interval:      5 * time.Second,
		Precision:     3,
		StoragePeriod: 1 * time.Hour,
	}

	overlay := &EnergyConfig{
		Enabled:   false,
		Interval:  10 * time.Second,
		Precision: 5,
	}

	result := merger.MergeEnergy(base, overlay)

	assert.Equal(t, false, result.Enabled, "Enabled should be from overlay")
	assert.Equal(t, 10*time.Second, result.Interval, "Interval should be from overlay")
	assert.Equal(t, 5, result.Precision, "Precision should be from overlay")
	assert.Equal(t, 1*time.Hour, result.StoragePeriod, "StoragePeriod should be from base")
}

func TestConfigMerger_MergeScheduler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &SchedulerConfig{
		Enabled:  true,
		Interval: 5 * time.Second,
	}

	overlay := &SchedulerConfig{
		Enabled:  false,
		Interval: 10 * time.Second,
	}

	result := merger.MergeScheduler(base, overlay)

	assert.Equal(t, false, result.Enabled, "Enabled should be from overlay")
	assert.Equal(t, 10*time.Second, result.Interval, "Interval should be from overlay")
}

func TestConfigMerger_MergeAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	base := &AuthConfig{
		Enabled:  true,
		Method:   "token",
		TokenURL: "http://base.example.com/token",
		Username: "baseuser",
		Password: "basepass",
		Timeout:  30 * time.Second,
		CacheTTL: 1 * time.Hour,
	}

	overlay := &AuthConfig{
		Enabled:  false,
		Method:   "basic",
		TokenURL: "http://overlay.example.com/token",
		Timeout:  60 * time.Second,
	}

	result := merger.MergeAuth(base, overlay)

	assert.Equal(t, false, result.Enabled, "Enabled should be from overlay")
	assert.Equal(t, "basic", result.Method, "Method should be from overlay")
	assert.Equal(t, "http://overlay.example.com/token", result.TokenURL, "TokenURL should be from overlay")
	assert.Equal(t, "baseuser", result.Username, "Username should be from base")
	assert.Equal(t, "basepass", result.Password, "Password should be from base")
	assert.Equal(t, 60*time.Second, result.Timeout, "Timeout should be from overlay")
	assert.Equal(t, 1*time.Hour, result.CacheTTL, "CacheTTL should be from base")
}

func TestConfigMerger_mergeString(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("non-empty overlay", func(t *testing.T) {
		result := merger.mergeString("base", "overlay")
		assert.Equal(t, "overlay", result, "Should return overlay when it's non-empty")
	})

	t.Run("empty overlay", func(t *testing.T) {
		result := merger.mergeString("base", "")
		assert.Equal(t, "base", result, "Should return base when overlay is empty")
	})

	t.Run("both empty", func(t *testing.T) {
		result := merger.mergeString("", "")
		assert.Equal(t, "", result, "Should return empty string when both are empty")
	})

	t.Run("empty base", func(t *testing.T) {
		result := merger.mergeString("", "overlay")
		assert.Equal(t, "overlay", result, "Should return overlay when base is empty")
	})
}

func TestConfigMerger_mergeInt(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("non-zero overlay", func(t *testing.T) {
		result := merger.mergeInt(100, 200)
		assert.Equal(t, 200, result, "Should return overlay when it's non-zero")
	})

	t.Run("zero overlay", func(t *testing.T) {
		result := merger.mergeInt(100, 0)
		assert.Equal(t, 100, result, "Should return base when overlay is zero")
	})

	t.Run("both zero", func(t *testing.T) {
		result := merger.mergeInt(0, 0)
		assert.Equal(t, 0, result, "Should return zero when both are zero")
	})

	t.Run("zero base", func(t *testing.T) {
		result := merger.mergeInt(0, 200)
		assert.Equal(t, 200, result, "Should return overlay when base is zero")
	})

	t.Run("negative values", func(t *testing.T) {
		result := merger.mergeInt(-10, -20)
		assert.Equal(t, -20, result, "Should return overlay when it's non-zero (even negative)")
	})
}

func TestConfigMerger_mergeBool(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("overlay true", func(t *testing.T) {
		result := merger.mergeBool(false, true)
		assert.Equal(t, true, result, "Should always return overlay")
	})

	t.Run("overlay false", func(t *testing.T) {
		result := merger.mergeBool(true, false)
		assert.Equal(t, false, result, "Should always return overlay")
	})

	t.Run("both true", func(t *testing.T) {
		result := merger.mergeBool(true, true)
		assert.Equal(t, true, result, "Should always return overlay")
	})

	t.Run("both false", func(t *testing.T) {
		result := merger.mergeBool(false, false)
		assert.Equal(t, false, result, "Should always return overlay")
	})
}

func TestConfigMerger_mergeDuration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("non-zero overlay", func(t *testing.T) {
		base := 10 * time.Second
		overlay := 20 * time.Second
		result := merger.mergeDuration(base, overlay)
		assert.Equal(t, overlay, result, "Should return overlay when it's non-zero")
	})

	t.Run("zero overlay", func(t *testing.T) {
		base := 10 * time.Second
		overlay := 0 * time.Second
		result := merger.mergeDuration(base, overlay)
		assert.Equal(t, base, result, "Should return base when overlay is zero")
	})

	t.Run("both zero", func(t *testing.T) {
		base := 0 * time.Second
		overlay := 0 * time.Second
		result := merger.mergeDuration(base, overlay)
		assert.Equal(t, 0*time.Second, result, "Should return zero when both are zero")
	})

	t.Run("zero base", func(t *testing.T) {
		base := 0 * time.Second
		overlay := 20 * time.Second
		result := merger.mergeDuration(base, overlay)
		assert.Equal(t, overlay, result, "Should return overlay when base is zero")
	})
}

func TestConfigMerger_isZeroValue(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("string", func(t *testing.T) {
		assert.True(t, merger.isZeroValue(reflect.ValueOf("")), "Empty string should be zero value")
		assert.False(t, merger.isZeroValue(reflect.ValueOf("test")), "Non-empty string should not be zero value")
	})

	t.Run("int", func(t *testing.T) {
		assert.True(t, merger.isZeroValue(reflect.ValueOf(0)), "Zero int should be zero value")
		assert.False(t, merger.isZeroValue(reflect.ValueOf(10)), "Non-zero int should not be zero value")
	})

	t.Run("bool", func(t *testing.T) {
		assert.True(t, merger.isZeroValue(reflect.ValueOf(false)), "False should be zero value")
		assert.False(t, merger.isZeroValue(reflect.ValueOf(true)), "True should not be zero value")
	})

	t.Run("duration", func(t *testing.T) {
		assert.True(t, merger.isZeroValue(reflect.ValueOf(0 * time.Second)), "Zero duration should be zero value")
		assert.False(t, merger.isZeroValue(reflect.ValueOf(10 * time.Second)), "Non-zero duration should not be zero value")
	})

	t.Run("pointer", func(t *testing.T) {
		var ptr *int
		assert.True(t, merger.isZeroValue(reflect.ValueOf(ptr)), "Nil pointer should be zero value")

		value := 10
		assert.False(t, merger.isZeroValue(reflect.ValueOf(&value)), "Non-nil pointer should not be zero value")
	})

	t.Run("slice", func(t *testing.T) {
		var slice []int
		assert.True(t, merger.isZeroValue(reflect.ValueOf(slice)), "Nil slice should be zero value")

		nonNilSlice := []int{1, 2, 3}
		assert.False(t, merger.isZeroValue(reflect.ValueOf(nonNilSlice)), "Non-nil slice should not be zero value")

		emptySlice := []int{}
		assert.False(t, merger.isZeroValue(reflect.ValueOf(emptySlice)), "Empty non-nil slice should not be zero value")
	})

	t.Run("map", func(t *testing.T) {
		var m map[string]int
		assert.True(t, merger.isZeroValue(reflect.ValueOf(m)), "Nil map should be zero value")

		nonNilMap := map[string]int{"key": 1}
		assert.False(t, merger.isZeroValue(reflect.ValueOf(nonNilMap)), "Non-nil map should not be zero value")
	})
}

func TestConfigMerger_DeepMerge(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	t.Run("deep merge structs", func(t *testing.T) {
		base := &Config{
			Server: ServerConfig{
				Port:        8080,
				Host:        "localhost",
				ReadTimeout: 30 * time.Second,
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "json",
			},
		}

		overlay := &Config{
			Server: ServerConfig{
				Port:     9090,
				Host:     "0.0.0.0",
				EnablePprof: true,
			},
			Logging: LoggingConfig{
				Level: "debug",
			},
		}

		result, err := merger.DeepMerge(base, overlay)
		require.NoError(t, err)
		require.NotNil(t, result)

		resultConfig, ok := result.(*Config)
		require.True(t, ok, "Result should be *Config")

		// 验证深合并结果
		assert.Equal(t, 9090, resultConfig.Server.Port, "Port should be from overlay")
		assert.Equal(t, "0.0.0.0", resultConfig.Server.Host, "Host should be from overlay")
		assert.Equal(t, 30*time.Second, resultConfig.Server.ReadTimeout, "ReadTimeout should be from base")
		assert.Equal(t, true, resultConfig.Server.EnablePprof, "EnablePprof should be from overlay")

		assert.Equal(t, "debug", resultConfig.Logging.Level, "Log level should be from overlay")
		assert.Equal(t, "json", resultConfig.Logging.Format, "Log format should be from base")
	})

	t.Run("nil pointers", func(t *testing.T) {
		var base *Config
		var overlay *Config

		result, err := merger.DeepMerge(base, overlay)
		assert.Error(t, err, "Should return error for nil pointers")
		assert.Nil(t, result, "Result should be nil for nil pointers")
		assert.Contains(t, err.Error(), "must be pointers")
	})

	t.Run("mismatched types", func(t *testing.T) {
		base := &Config{}
		overlay := &ServerConfig{}

		result, err := merger.DeepMerge(base, overlay)
		assert.Error(t, err, "Should return error for mismatched types")
		assert.Nil(t, result, "Result should be nil for mismatched types")
		assert.Contains(t, err.Error(), "types must match")
	})
}

func TestConfigMerger_Integration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	merger := NewMerger(logger)

	// 测试完整的配置合并流程
	defaults := &Config{
		Server: ServerConfig{
			Port:        8080,
			Host:        "localhost",
			ReadTimeout: 30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
			EnablePprof: false,
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSize:    100,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		},
		WinPower: WinPowerConfig{
			Timeout:       30 * time.Second,
			RetryInterval: 5 * time.Second,
			MaxRetries:    3,
			SkipSSLVerify: false,
		},
		Storage: StorageConfig{
			DataDir:   "./data",
			SyncWrite: true,
		},
		Collector: CollectorConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Timeout:       3 * time.Second,
			MaxConcurrent: 5,
		},
		Metrics: MetricsConfig{
			Enabled:   true,
			Path:      "/metrics",
			Namespace: "winpower",
			Subsystem: "exporter",
			HelpText:  "WinPower G2 Exporter Metrics",
		},
		Energy: EnergyConfig{
			Enabled:       true,
			Interval:      5 * time.Second,
			Precision:     3,
			StoragePeriod: 1 * time.Hour,
		},
		Scheduler: SchedulerConfig{
			Enabled:  true,
			Interval: 5 * time.Second,
		},
		Auth: AuthConfig{
			Enabled: false,
			Method:  "token",
			Timeout: 30 * time.Second,
			CacheTTL: 1 * time.Hour,
		},
	}

	// 环境变量配置
	envConfig := &Config{
		Server: ServerConfig{
			Port: 9090,
		},
		Logging: LoggingConfig{
			Level: "debug",
		},
		WinPower: WinPowerConfig{
			URL:           "http://env.example.com",
			Username:      "envuser",
			Password:      "envpass",
			SkipSSLVerify: true,
		},
		Storage: StorageConfig{
			DataDir: "/env/data",
		},
	}

	// 命令行配置
	cliConfig := &Config{
		Server: ServerConfig{
			Host:        "0.0.0.0",
			EnablePprof: true,
		},
		Logging: LoggingConfig{
			Format: "console",
		},
		WinPower: WinPowerConfig{
			Timeout: 60 * time.Second,
		},
	}

	// 合并配置：默认值 -> 环境变量 -> 命令行参数
	result, err := merger.MergeWithDefaults(defaults, envConfig, cliConfig)
	require.NoError(t, err)
	require.NotNil(t, result)

	// 验证合并结果优先级
	assert.Equal(t, 9090, result.Server.Port, "Port should be from environment config")
	assert.Equal(t, "0.0.0.0", result.Server.Host, "Host should be from CLI config")
	assert.Equal(t, 30*time.Second, result.Server.ReadTimeout, "ReadTimeout should be from defaults")
	assert.Equal(t, true, result.Server.EnablePprof, "EnablePprof should be from CLI config")

	assert.Equal(t, "debug", result.Logging.Level, "Log level should be from environment config")
	assert.Equal(t, "console", result.Logging.Format, "Log format should be from CLI config")
	assert.Equal(t, "stdout", result.Logging.Output, "Log output should be from defaults")

	assert.Equal(t, "http://env.example.com", result.WinPower.URL, "WinPower URL should be from environment config")
	assert.Equal(t, "envuser", result.WinPower.Username, "WinPower username should be from environment config")
	assert.Equal(t, "envpass", result.WinPower.Password, "WinPower password should be from environment config")
	assert.Equal(t, 60*time.Second, result.WinPower.Timeout, "WinPower timeout should be from CLI config")
	assert.Equal(t, 5*time.Second, result.WinPower.RetryInterval, "WinPower retry interval should be from defaults")
	assert.Equal(t, true, result.WinPower.SkipSSLVerify, "WinPower skip SSL verify should be from environment config")

	assert.Equal(t, "/env/data", result.Storage.DataDir, "Data directory should be from environment config")
	assert.Equal(t, true, result.Storage.SyncWrite, "Sync write should be from defaults")

	// 验证未修改的字段保持默认值
	assert.Equal(t, true, result.Collector.Enabled, "Collector enabled should be from defaults")
	assert.Equal(t, 5*time.Second, result.Collector.Interval, "Collector interval should be from defaults")
	assert.Equal(t, true, result.Metrics.Enabled, "Metrics enabled should be from defaults")
	assert.Equal(t, "/metrics", result.Metrics.Path, "Metrics path should be from defaults")
	assert.Equal(t, true, result.Energy.Enabled, "Energy enabled should be from defaults")
	assert.Equal(t, true, result.Scheduler.Enabled, "Scheduler enabled should be from defaults")
	assert.Equal(t, false, result.Auth.Enabled, "Auth enabled should be from defaults")
}