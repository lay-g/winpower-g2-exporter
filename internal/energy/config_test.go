package energy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0.01, config.Precision)
	assert.True(t, config.EnableStats)
	assert.Equal(t, int64(time.Second), config.MaxCalculationTime)
	assert.True(t, config.NegativePowerAllowed)
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := DefaultConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("nil config", func(t *testing.T) {
		var config *Config
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("invalid precision (zero)", func(t *testing.T) {
		config := &Config{Precision: 0}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "precision must be positive")
	})

	t.Run("invalid precision (negative)", func(t *testing.T) {
		config := &Config{Precision: -0.01}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "precision must be positive")
	})

	t.Run("invalid max calculation time (zero)", func(t *testing.T) {
		config := &Config{
			Precision:          0.01,
			MaxCalculationTime: 0,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max calculation time must be positive")
	})

	t.Run("invalid max calculation time (negative)", func(t *testing.T) {
		config := &Config{
			Precision:          0.01,
			MaxCalculationTime: -1000,
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max calculation time must be positive")
	})

	t.Run("max calculation time too long", func(t *testing.T) {
		config := &Config{
			Precision:          0.01,
			MaxCalculationTime: 11 * time.Second.Nanoseconds(),
		}
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "should not exceed 10 seconds")
	})
}

func TestConfig_String(t *testing.T) {
	t.Run("normal config", func(t *testing.T) {
		config := &Config{
			Precision:            0.01,
			EnableStats:          true,
			MaxCalculationTime:   time.Second.Nanoseconds(),
			NegativePowerAllowed: false,
		}

		str := config.String()
		assert.Contains(t, str, "0.0100")
		assert.Contains(t, str, "true")
		assert.Contains(t, str, "false")
		assert.Contains(t, str, "1s")
	})

	t.Run("nil config", func(t *testing.T) {
		var config *Config
		str := config.String()
		assert.Equal(t, "<nil>", str)
	})
}

func TestConfig_Clone(t *testing.T) {
	original := &Config{
		Precision:            0.001,
		EnableStats:          false,
		MaxCalculationTime:   2 * time.Second.Nanoseconds(),
		NegativePowerAllowed: true,
	}

	cloned := original.Clone()

	assert.Equal(t, original.Precision, cloned.Precision)
	assert.Equal(t, original.EnableStats, cloned.EnableStats)
	assert.Equal(t, original.MaxCalculationTime, cloned.MaxCalculationTime)
	assert.Equal(t, original.NegativePowerAllowed, cloned.NegativePowerAllowed)

	// Modify cloned to ensure they're independent
	cloned.Precision = 0.999
	assert.NotEqual(t, original.Precision, cloned.Precision)
}

func TestConfig_CloneNil(t *testing.T) {
	var config *Config
	cloned := config.Clone()
	assert.Nil(t, cloned)
}

func TestConfig_WithMethods(t *testing.T) {
	config := &Config{}

	// Test WithPrecision
	newConfig := config.WithPrecision(0.001)
	assert.Same(t, config, newConfig) // Should return same instance for chaining
	assert.Equal(t, 0.001, config.Precision)

	// Test WithStatsEnabled
	newConfig = config.WithStatsEnabled(false)
	assert.Same(t, config, newConfig)
	assert.False(t, config.EnableStats)

	// Test WithMaxCalculationTime
	newConfig = config.WithMaxCalculationTime(5 * time.Second)
	assert.Same(t, config, newConfig)
	assert.Equal(t, int64(5*time.Second), config.MaxCalculationTime)

	// Test WithNegativePowerAllowed
	newConfig = config.WithNegativePowerAllowed(false)
	assert.Same(t, config, newConfig)
	assert.False(t, config.NegativePowerAllowed)
}