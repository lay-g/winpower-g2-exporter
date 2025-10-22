package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("returns default configuration values", func(t *testing.T) {
		config := DefaultConfig()

		assert.Equal(t, 5*time.Second, config.CollectionInterval, "Default collection interval should be 5s")
		assert.Equal(t, 30*time.Second, config.GracefulShutdownTimeout, "Default graceful shutdown timeout should be 30s")
	})
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config should pass", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      5 * time.Second,
			GracefulShutdownTimeout: 30 * time.Second,
		}

		err := config.Validate()
		assert.NoError(t, err, "Valid config should not return error")
	})

	t.Run("nil config should error", func(t *testing.T) {
		var config *Config

		err := config.Validate()
		assert.Error(t, err, "Nil config should return error")
	})

	t.Run("zero collection interval should error", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      0,
			GracefulShutdownTimeout: 30 * time.Second,
		}

		err := config.Validate()
		assert.Error(t, err, "Zero collection interval should return error")
		assert.Contains(t, err.Error(), "collection interval cannot be zero or negative", "Error should mention collection interval")
	})

	t.Run("negative collection interval should error", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      -1 * time.Second,
			GracefulShutdownTimeout: 30 * time.Second,
		}

		err := config.Validate()
		assert.Error(t, err, "Negative collection interval should return error")
		assert.Contains(t, err.Error(), "collection interval cannot be zero or negative", "Error should mention collection interval")
	})

	t.Run("zero graceful shutdown timeout should error", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      5 * time.Second,
			GracefulShutdownTimeout: 0,
		}

		err := config.Validate()
		assert.Error(t, err, "Zero graceful shutdown timeout should return error")
		assert.Contains(t, err.Error(), "graceful shutdown timeout cannot be zero or negative", "Error should mention graceful shutdown timeout")
	})

	t.Run("negative graceful shutdown timeout should error", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      5 * time.Second,
			GracefulShutdownTimeout: -1 * time.Second,
		}

		err := config.Validate()
		assert.Error(t, err, "Negative graceful shutdown timeout should return error")
		assert.Contains(t, err.Error(), "graceful shutdown timeout cannot be zero or negative", "Error should mention graceful shutdown timeout")
	})

	t.Run("minimum valid values should pass", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Nanosecond,
			GracefulShutdownTimeout: 1 * time.Nanosecond,
		}

		err := config.Validate()
		assert.NoError(t, err, "Minimum positive values should be valid")
	})
}

func TestConfig_SetDefaults(t *testing.T) {
	t.Run("sets default values for empty config", func(t *testing.T) {
		config := &Config{}

		config.SetDefaults()

		assert.Equal(t, 5*time.Second, config.CollectionInterval, "Default collection interval should be set")
		assert.Equal(t, 30*time.Second, config.GracefulShutdownTimeout, "Default graceful shutdown timeout should be set")
	})

	t.Run("preserves existing non-zero values", func(t *testing.T) {
		config := &Config{
			CollectionInterval: 10 * time.Second,
		}

		config.SetDefaults()

		assert.Equal(t, 10*time.Second, config.CollectionInterval, "Existing collection interval should be preserved")
		assert.Equal(t, 30*time.Second, config.GracefulShutdownTimeout, "Default graceful shutdown timeout should be set")
	})

	t.Run("handles both values already set", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      15 * time.Second,
			GracefulShutdownTimeout: 45 * time.Second,
		}

		config.SetDefaults()

		assert.Equal(t, 15*time.Second, config.CollectionInterval, "Existing collection interval should be preserved")
		assert.Equal(t, 45*time.Second, config.GracefulShutdownTimeout, "Existing graceful shutdown timeout should be preserved")
	})
}

func TestConfig_String(t *testing.T) {
	t.Run("returns string representation", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      5 * time.Second,
			GracefulShutdownTimeout: 30 * time.Second,
		}

		str := config.String()

		assert.Contains(t, str, "SchedulerConfig{", "String should contain struct name")
		assert.Contains(t, str, "CollectionInterval: 5s", "String should contain collection interval")
		assert.Contains(t, str, "GracefulShutdownTimeout: 30s", "String should contain graceful shutdown timeout")
		assert.Contains(t, str, "}", "String should end with closing brace")
	})

	t.Run("handles nil config", func(t *testing.T) {
		var config *Config

		str := config.String()

		assert.Equal(t, "<nil>", str, "Nil config should return <nil>")
	})

	t.Run("handles different time durations", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      100 * time.Millisecond,
			GracefulShutdownTimeout: 2 * time.Minute,
		}

		str := config.String()

		assert.Contains(t, str, "CollectionInterval: 100ms", "String should handle milliseconds")
		assert.Contains(t, str, "GracefulShutdownTimeout: 2m0s", "String should handle minutes")
	})
}

func TestConfig_Clone(t *testing.T) {
	t.Run("creates deep copy", func(t *testing.T) {
		original := &Config{
			CollectionInterval:      5 * time.Second,
			GracefulShutdownTimeout: 30 * time.Second,
		}

		clonedInterface := original.Clone()
		cloned, ok := clonedInterface.(*Config)
		require.True(t, ok, "Clone should return *Config type")

		// Verify all fields are copied
		assert.Equal(t, original.CollectionInterval, cloned.CollectionInterval, "Collection interval should be copied")
		assert.Equal(t, original.GracefulShutdownTimeout, cloned.GracefulShutdownTimeout, "Graceful shutdown timeout should be copied")

		// Verify they are different objects
		assert.NotSame(t, original, cloned, "Clone should create a new object")
	})

	t.Run("modifying clone does not affect original", func(t *testing.T) {
		original := &Config{
			CollectionInterval:      5 * time.Second,
			GracefulShutdownTimeout: 30 * time.Second,
		}

		clonedInterface := original.Clone()
		cloned, ok := clonedInterface.(*Config)
		require.True(t, ok, "Clone should return *Config type")

		cloned.CollectionInterval = 10 * time.Second
		cloned.GracefulShutdownTimeout = 60 * time.Second

		assert.Equal(t, 5*time.Second, original.CollectionInterval, "Original collection interval should not be affected")
		assert.Equal(t, 30*time.Second, original.GracefulShutdownTimeout, "Original graceful shutdown timeout should not be affected")
		assert.Equal(t, 10*time.Second, cloned.CollectionInterval, "Cloned collection interval should be modified")
		assert.Equal(t, 60*time.Second, cloned.GracefulShutdownTimeout, "Cloned graceful shutdown timeout should be modified")
	})

	t.Run("handles nil config", func(t *testing.T) {
		var original *Config

		cloned := original.Clone()

		assert.Nil(t, cloned, "Clone of nil should return nil")
	})
}

func TestConfig_EdgeCases(t *testing.T) {
	t.Run("handles very large values", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      24 * time.Hour,
			GracefulShutdownTimeout: 7 * 24 * time.Hour, // 7 days
		}

		err := config.Validate()
		assert.NoError(t, err, "Large valid values should be accepted")

		str := config.String()
		assert.Contains(t, str, "24h0m0s", "String should handle hours")
		assert.Contains(t, str, "168h0m0s", "String should handle large durations")
	})

	t.Run("handles very small values", func(t *testing.T) {
		config := &Config{
			CollectionInterval:      1 * time.Nanosecond,
			GracefulShutdownTimeout: 1 * time.Microsecond,
		}

		err := config.Validate()
		assert.NoError(t, err, "Small positive values should be accepted")

		str := config.String()
		assert.Contains(t, str, "1ns", "String should handle nanoseconds")
		assert.Contains(t, str, "1µs", "String should handle microseconds")
	})
}

func TestSchedulerModuleConfig(t *testing.T) {
	t.Run("GetModuleConfig returns valid config", func(t *testing.T) {
		config := GetModuleConfig()
		assert.NotNil(t, config, "Module config should not be nil")
		assert.NotNil(t, config.Config, "Config should not be nil")
		assert.Equal(t, 5*time.Second, config.Config.CollectionInterval, "Should have default collection interval")
		assert.Equal(t, 30*time.Second, config.Config.GracefulShutdownTimeout, "Should have default graceful shutdown timeout")
	})

	t.Run("SetModuleConfig updates global config", func(t *testing.T) {
		originalConfig := GetModuleConfig()
		newConfig := &SchedulerModuleConfig{
			Config: &Config{
				CollectionInterval:      10 * time.Second,
				GracefulShutdownTimeout: 60 * time.Second,
			},
		}

		SetModuleConfig(newConfig)
		updatedConfig := GetModuleConfig()

		assert.Equal(t, 10*time.Second, updatedConfig.Config.CollectionInterval, "Should update collection interval")
		assert.Equal(t, 60*time.Second, updatedConfig.Config.GracefulShutdownTimeout, "Should update graceful shutdown timeout")

		// Restore original config
		SetModuleConfig(originalConfig)
	})

	t.Run("SetModuleConfig with nil does nothing", func(t *testing.T) {
		originalConfig := GetModuleConfig()

		SetModuleConfig(nil)
		configAfterNil := GetModuleConfig()

		assert.Equal(t, originalConfig, configAfterNil, "Setting nil should not change config")
	})
}
