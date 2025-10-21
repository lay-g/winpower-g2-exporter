package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("returns default configuration values", func(t *testing.T) {
		config := DefaultConfig()

		assert.Equal(t, 9090, config.Port, "Default port should be 9090")
		assert.Equal(t, "0.0.0.0", config.Host, "Default host should be 0.0.0.0")
		assert.Equal(t, "release", config.Mode, "Default mode should be release")
		assert.Equal(t, 30*time.Second, config.ReadTimeout, "Default read timeout should be 30s")
		assert.Equal(t, 30*time.Second, config.WriteTimeout, "Default write timeout should be 30s")
		assert.Equal(t, 60*time.Second, config.IdleTimeout, "Default idle timeout should be 60s")
		assert.False(t, config.EnablePprof, "Default pprof should be disabled")
		assert.False(t, config.EnableCORS, "Default CORS should be disabled")
		assert.False(t, config.EnableRateLimit, "Default rate limit should be disabled")
	})
}

func TestConfigValidateValidValues(t *testing.T) {
	t.Run("accepts valid configuration", func(t *testing.T) {
		config := &Config{
			Port:            8080,
			Host:            "localhost",
			Mode:            "release",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     30 * time.Second,
			EnablePprof:     false,
			EnableCORS:      false,
			EnableRateLimit: false,
		}

		err := config.Validate()
		assert.NoError(t, err, "Valid configuration should not return error")
	})

	t.Run("accepts debug mode", func(t *testing.T) {
		config := DefaultConfig()
		config.Mode = "debug"

		err := config.Validate()
		assert.NoError(t, err, "Debug mode should be valid")
	})

	t.Run("accepts test mode", func(t *testing.T) {
		config := DefaultConfig()
		config.Mode = "test"

		err := config.Validate()
		assert.NoError(t, err, "Test mode should be valid")
	})

	t.Run("accepts pprof enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnablePprof = true

		err := config.Validate()
		assert.NoError(t, err, "Pprof enabled should be valid")
	})

	t.Run("accepts CORS enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableCORS = true

		err := config.Validate()
		assert.NoError(t, err, "CORS enabled should be valid")
	})

	t.Run("accepts rate limit enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnableRateLimit = true

		err := config.Validate()
		assert.NoError(t, err, "Rate limit enabled should be valid")
	})
}

func TestConfigValidateInvalidValues(t *testing.T) {
	t.Run("rejects invalid port (0)", func(t *testing.T) {
		config := DefaultConfig()
		config.Port = 0

		err := config.Validate()
		assert.Error(t, err, "Port 0 should be invalid")
		assert.Contains(t, err.Error(), "port", "Error should mention port")
	})

	t.Run("rejects invalid port (-1)", func(t *testing.T) {
		config := DefaultConfig()
		config.Port = -1

		err := config.Validate()
		assert.Error(t, err, "Negative port should be invalid")
		assert.Contains(t, err.Error(), "port", "Error should mention port")
	})

	t.Run("rejects invalid port (65536)", func(t *testing.T) {
		config := DefaultConfig()
		config.Port = 65536

		err := config.Validate()
		assert.Error(t, err, "Port 65536 should be invalid")
		assert.Contains(t, err.Error(), "port", "Error should mention port")
	})

	t.Run("rejects empty host", func(t *testing.T) {
		config := DefaultConfig()
		config.Host = ""

		err := config.Validate()
		assert.Error(t, err, "Empty host should be invalid")
		assert.Contains(t, err.Error(), "host", "Error should mention host")
	})

	t.Run("rejects invalid mode", func(t *testing.T) {
		config := DefaultConfig()
		config.Mode = "invalid"

		err := config.Validate()
		assert.Error(t, err, "Invalid mode should be rejected")
		assert.Contains(t, err.Error(), "mode", "Error should mention mode")
	})

	t.Run("rejects zero read timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.ReadTimeout = 0

		err := config.Validate()
		assert.Error(t, err, "Zero read timeout should be invalid")
		assert.Contains(t, err.Error(), "read timeout", "Error should mention read timeout")
	})

	t.Run("rejects negative read timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.ReadTimeout = -1 * time.Second

		err := config.Validate()
		assert.Error(t, err, "Negative read timeout should be invalid")
		assert.Contains(t, err.Error(), "read timeout", "Error should mention read timeout")
	})

	t.Run("rejects zero write timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.WriteTimeout = 0

		err := config.Validate()
		assert.Error(t, err, "Zero write timeout should be invalid")
		assert.Contains(t, err.Error(), "write timeout", "Error should mention write timeout")
	})

	t.Run("rejects negative write timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.WriteTimeout = -1 * time.Second

		err := config.Validate()
		assert.Error(t, err, "Negative write timeout should be invalid")
		assert.Contains(t, err.Error(), "write timeout", "Error should mention write timeout")
	})

	t.Run("rejects zero idle timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.IdleTimeout = 0

		err := config.Validate()
		assert.Error(t, err, "Zero idle timeout should be invalid")
		assert.Contains(t, err.Error(), "idle timeout", "Error should mention idle timeout")
	})

	t.Run("rejects negative idle timeout", func(t *testing.T) {
		config := DefaultConfig()
		config.IdleTimeout = -1 * time.Second

		err := config.Validate()
		assert.Error(t, err, "Negative idle timeout should be invalid")
		assert.Contains(t, err.Error(), "idle timeout", "Error should mention idle timeout")
	})
}

func TestConfigSerialization(t *testing.T) {
	t.Run("JSON serialization and deserialization", func(t *testing.T) {
		original := &Config{
			Port:            8080,
			Host:            "127.0.0.1",
			Mode:            "debug",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    20 * time.Second,
			IdleTimeout:     45 * time.Second,
			EnablePprof:     true,
			EnableCORS:      true,
			EnableRateLimit: false,
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(original)
		require.NoError(t, err, "JSON marshaling should succeed")

		// Test JSON unmarshaling
		var restored Config
		err = json.Unmarshal(jsonData, &restored)
		require.NoError(t, err, "JSON unmarshaling should succeed")

		assert.Equal(t, original.Port, restored.Port, "Port should be preserved")
		assert.Equal(t, original.Host, restored.Host, "Host should be preserved")
		assert.Equal(t, original.Mode, restored.Mode, "Mode should be preserved")
		assert.Equal(t, original.ReadTimeout, restored.ReadTimeout, "Read timeout should be preserved")
		assert.Equal(t, original.WriteTimeout, restored.WriteTimeout, "Write timeout should be preserved")
		assert.Equal(t, original.IdleTimeout, restored.IdleTimeout, "Idle timeout should be preserved")
		assert.Equal(t, original.EnablePprof, restored.EnablePprof, "Pprof flag should be preserved")
		assert.Equal(t, original.EnableCORS, restored.EnableCORS, "CORS flag should be preserved")
		assert.Equal(t, original.EnableRateLimit, restored.EnableRateLimit, "Rate limit flag should be preserved")
	})

	t.Run("YAML serialization and deserialization", func(t *testing.T) {
		original := &Config{
			Port:            9090,
			Host:            "0.0.0.0",
			Mode:            "release",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			EnablePprof:     false,
			EnableCORS:      false,
			EnableRateLimit: true,
		}

		// Test YAML marshaling
		yamlData, err := yaml.Marshal(original)
		require.NoError(t, err, "YAML marshaling should succeed")

		// Test YAML unmarshaling
		var restored Config
		err = yaml.Unmarshal(yamlData, &restored)
		require.NoError(t, err, "YAML unmarshaling should succeed")

		assert.Equal(t, original.Port, restored.Port, "Port should be preserved")
		assert.Equal(t, original.Host, restored.Host, "Host should be preserved")
		assert.Equal(t, original.Mode, restored.Mode, "Mode should be preserved")
		assert.Equal(t, original.ReadTimeout, restored.ReadTimeout, "Read timeout should be preserved")
		assert.Equal(t, original.WriteTimeout, restored.WriteTimeout, "Write timeout should be preserved")
		assert.Equal(t, original.IdleTimeout, restored.IdleTimeout, "Idle timeout should be preserved")
		assert.Equal(t, original.EnablePprof, restored.EnablePprof, "Pprof flag should be preserved")
		assert.Equal(t, original.EnableCORS, restored.EnableCORS, "CORS flag should be preserved")
		assert.Equal(t, original.EnableRateLimit, restored.EnableRateLimit, "Rate limit flag should be preserved")
	})

	t.Run("partial JSON deserialization with defaults", func(t *testing.T) {
		// Test with partial JSON data
		partialJSON := `{"port": 8080, "mode": "debug"}`

		var restored Config
		err := json.Unmarshal([]byte(partialJSON), &restored)
		require.NoError(t, err, "Partial JSON unmarshaling should succeed")

		// Should have parsed fields
		assert.Equal(t, 8080, restored.Port, "Port should be from JSON")
		assert.Equal(t, "debug", restored.Mode, "Mode should be from JSON")

		// Should have zero values for missing fields
		assert.Equal(t, "", restored.Host, "Missing host should be empty")
		assert.Equal(t, time.Duration(0), restored.ReadTimeout, "Missing read timeout should be zero")
		assert.Equal(t, time.Duration(0), restored.WriteTimeout, "Missing write timeout should be zero")
		assert.Equal(t, time.Duration(0), restored.IdleTimeout, "Missing idle timeout should be zero")
		assert.False(t, restored.EnablePprof, "Missing pprof flag should be false")
		assert.False(t, restored.EnableCORS, "Missing CORS flag should be false")
		assert.False(t, restored.EnableRateLimit, "Missing rate limit flag should be false")
	})
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("accepts minimum valid port (1)", func(t *testing.T) {
		config := DefaultConfig()
		config.Port = 1

		err := config.Validate()
		assert.NoError(t, err, "Port 1 should be valid")
	})

	t.Run("accepts maximum valid port (65535)", func(t *testing.T) {
		config := DefaultConfig()
		config.Port = 65535

		err := config.Validate()
		assert.NoError(t, err, "Port 65535 should be valid")
	})

	t.Run("accepts minimum timeout (1ns)", func(t *testing.T) {
		config := DefaultConfig()
		config.ReadTimeout = 1 * time.Nanosecond
		config.WriteTimeout = 1 * time.Nanosecond
		config.IdleTimeout = 1 * time.Nanosecond

		err := config.Validate()
		assert.NoError(t, err, "Minimum timeout should be valid")
	})

	t.Run("accepts all optional features enabled", func(t *testing.T) {
		config := DefaultConfig()
		config.EnablePprof = true
		config.EnableCORS = true
		config.EnableRateLimit = true

		err := config.Validate()
		assert.NoError(t, err, "All optional features should be valid")
	})
}
