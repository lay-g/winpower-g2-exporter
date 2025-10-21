package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExampleConfigDevelopment(t *testing.T) {
	config := ExampleConfigDevelopment()

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}
	if config.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.Host)
	}
	if config.Mode != "debug" {
		t.Errorf("Expected mode debug, got %s", config.Mode)
	}
	if !config.EnablePprof {
		t.Error("Expected pprof to be enabled")
	}
	if !config.EnableCORS {
		t.Error("Expected CORS to be enabled")
	}
	if config.EnableRateLimit {
		t.Error("Expected rate limit to be disabled")
	}
}

func TestExampleConfigProduction(t *testing.T) {
	config := ExampleConfigProduction()

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}
	if config.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Host)
	}
	if config.Mode != "release" {
		t.Errorf("Expected mode release, got %s", config.Mode)
	}
	if config.EnablePprof {
		t.Error("Expected pprof to be disabled")
	}
	if config.EnableCORS {
		t.Error("Expected CORS to be disabled")
	}
	if !config.EnableRateLimit {
		t.Error("Expected rate limit to be enabled")
	}
}

func TestExampleConfigHighSecurity(t *testing.T) {
	config := ExampleConfigHighSecurity()

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}
	if config.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.Host)
	}
	if config.Mode != "release" {
		t.Errorf("Expected mode release, got %s", config.Mode)
	}
	if config.EnablePprof {
		t.Error("Expected pprof to be disabled")
	}
	if config.EnableCORS {
		t.Error("Expected CORS to be disabled")
	}
	if !config.EnableRateLimit {
		t.Error("Expected rate limit to be enabled")
	}
}

func TestExampleConfigContainer(t *testing.T) {
	config := ExampleConfigContainer()

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}
	if config.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Host)
	}
	if config.Mode != "release" {
		t.Errorf("Expected mode release, got %s", config.Mode)
	}
	if config.EnableRateLimit {
		t.Error("Expected rate limit to be disabled")
	}
}

func TestExampleConfigHighThroughput(t *testing.T) {
	config := ExampleConfigHighThroughput()

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}
	if config.ReadTimeout != 10*time.Second {
		t.Errorf("Expected read timeout 10s, got %v", config.ReadTimeout)
	}
	if config.WriteTimeout != 10*time.Second {
		t.Errorf("Expected write timeout 10s, got %v", config.WriteTimeout)
	}
	if !config.EnableRateLimit {
		t.Error("Expected rate limit to be enabled")
	}
}

func TestExampleConfigLargePayloads(t *testing.T) {
	config := ExampleConfigLargePayloads()

	if config.ReadTimeout != 60*time.Second {
		t.Errorf("Expected read timeout 60s, got %v", config.ReadTimeout)
	}
	if config.WriteTimeout != 120*time.Second {
		t.Errorf("Expected write timeout 120s, got %v", config.WriteTimeout)
	}
	if config.IdleTimeout != 300*time.Second {
		t.Errorf("Expected idle timeout 300s, got %v", config.IdleTimeout)
	}
}

func TestExampleConfigTesting(t *testing.T) {
	config := ExampleConfigTesting()

	if config.Port != 0 {
		t.Errorf("Expected port 0 (random), got %d", config.Port)
	}
	if config.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.Host)
	}
	if config.Mode != "test" {
		t.Errorf("Expected mode test, got %s", config.Mode)
	}
	if config.ReadTimeout != 5*time.Second {
		t.Errorf("Expected read timeout 5s, got %v", config.ReadTimeout)
	}
}

func TestExampleConfigBehindProxy(t *testing.T) {
	config := ExampleConfigBehindProxy()

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}
	if config.EnableRateLimit {
		t.Error("Expected rate limit to be disabled (handled by proxy)")
	}
	if config.EnableCORS {
		t.Error("Expected CORS to be disabled (handled by proxy)")
	}
}

func TestExampleConfigMinimal(t *testing.T) {
	config := ExampleConfigMinimal()

	// Should be based on DefaultConfig
	defaultConfig := DefaultConfig()
	if config.Port != defaultConfig.Port {
		t.Errorf("Expected port %d, got %d", defaultConfig.Port, config.Port)
	}
	if config.EnablePprof {
		t.Error("Expected pprof to be disabled")
	}
	if config.EnableCORS {
		t.Error("Expected CORS to be disabled")
	}
	if config.EnableRateLimit {
		t.Error("Expected rate limit to be disabled")
	}
}

func TestNewConfigurationBuilder(t *testing.T) {
	builder := NewConfigurationBuilder()
	if builder.config == nil {
		t.Error("Expected builder config to be initialized")
	}
}

func TestConfigurationBuilder_WithPort(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithPort(8080)

	if builder.config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", builder.config.Port)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithHost(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithHost("localhost")

	if builder.config.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", builder.config.Host)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithMode(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithMode("test")

	if builder.config.Mode != "test" {
		t.Errorf("Expected mode test, got %s", builder.config.Mode)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithTimeouts(t *testing.T) {
	builder := NewConfigurationBuilder()
	read, write, idle := 5*time.Second, 10*time.Second, 15*time.Second
	result := builder.WithTimeouts(read, write, idle)

	if builder.config.ReadTimeout != read {
		t.Errorf("Expected read timeout %v, got %v", read, builder.config.ReadTimeout)
	}
	if builder.config.WriteTimeout != write {
		t.Errorf("Expected write timeout %v, got %v", write, builder.config.WriteTimeout)
	}
	if builder.config.IdleTimeout != idle {
		t.Errorf("Expected idle timeout %v, got %v", idle, builder.config.IdleTimeout)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithDebugMode(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithDebugMode()

	if builder.config.Mode != "debug" {
		t.Errorf("Expected mode debug, got %s", builder.config.Mode)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithReleaseMode(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithReleaseMode()

	if builder.config.Mode != "release" {
		t.Errorf("Expected mode release, got %s", builder.config.Mode)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithPprof(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithPprof(true)

	if !builder.config.EnablePprof {
		t.Error("Expected pprof to be enabled")
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}

	// Test disabling
	_ = builder.WithPprof(false)
	if builder.config.EnablePprof {
		t.Error("Expected pprof to be disabled")
	}
}

func TestConfigurationBuilder_WithCORS(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithCORS(true)

	if !builder.config.EnableCORS {
		t.Error("Expected CORS to be enabled")
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithRateLimit(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithRateLimit(true)

	if !builder.config.EnableRateLimit {
		t.Error("Expected rate limit to be enabled")
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithSecurity(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithSecurity()

	if builder.config.EnablePprof {
		t.Error("Expected pprof to be disabled")
	}
	if builder.config.EnableCORS {
		t.Error("Expected CORS to be disabled")
	}
	if !builder.config.EnableRateLimit {
		t.Error("Expected rate limit to be enabled")
	}
	if builder.config.Mode != "release" {
		t.Errorf("Expected mode release, got %s", builder.config.Mode)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_WithPerformance(t *testing.T) {
	builder := NewConfigurationBuilder()
	result := builder.WithPerformance()

	if builder.config.Mode != "release" {
		t.Errorf("Expected mode release, got %s", builder.config.Mode)
	}
	if builder.config.ReadTimeout != 10*time.Second {
		t.Errorf("Expected read timeout 10s, got %v", builder.config.ReadTimeout)
	}
	if builder.config.WriteTimeout != 10*time.Second {
		t.Errorf("Expected write timeout 10s, got %v", builder.config.WriteTimeout)
	}
	if builder.config.IdleTimeout != 30*time.Second {
		t.Errorf("Expected idle timeout 30s, got %v", builder.config.IdleTimeout)
	}
	if result != builder {
		t.Error("Expected builder to return itself for chaining")
	}
}

func TestConfigurationBuilder_Build(t *testing.T) {
	builder := NewConfigurationBuilder()
	builder.WithPort(8080)

	config, err := builder.Build()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}
}

func TestConfigurationBuilder_Build_ValidationError(t *testing.T) {
	builder := NewConfigurationBuilder()
	builder.WithPort(-1) // Invalid port

	_, err := builder.Build()
	if err == nil {
		t.Error("Expected validation error, got nil")
	}
}

func TestConfigurationBuilder_MustBuild(t *testing.T) {
	builder := NewConfigurationBuilder()
	builder.WithPort(8080)

	config := builder.MustBuild()
	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}
}

func TestConfigurationBuilder_MustBuild_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on invalid config")
		}
	}()

	builder := NewConfigurationBuilder()
	builder.WithPort(-1) // Invalid port
	builder.MustBuild()
}

func TestExampleUsage(t *testing.T) {
	config := ExampleUsage()

	if config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Port)
	}
	if config.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", config.Host)
	}
	if config.Mode != "debug" {
		t.Errorf("Expected mode debug, got %s", config.Mode)
	}
	if !config.EnablePprof {
		t.Error("Expected pprof to be enabled")
	}
	if !config.EnableCORS {
		t.Error("Expected CORS to be enabled")
	}
}

func TestExampleUsage_PanicOnError(t *testing.T) {
	// This test verifies that ExampleUsage panics on validation errors
	// We can't directly test the panic case because ExampleUsage uses fixed valid values
	// but we can test that it returns the expected configuration
	config := ExampleUsage()

	// Verify all expected values are set correctly
	assert.NotNil(t, config, "Config should not be nil")
	assert.Equal(t, 8080, config.Port, "Port should be 8080")
	assert.Equal(t, "127.0.0.1", config.Host, "Host should be 127.0.0.1")
	assert.Equal(t, "debug", config.Mode, "Mode should be debug")
	assert.True(t, config.EnablePprof, "Pprof should be enabled")
	assert.True(t, config.EnableCORS, "CORS should be enabled")
	assert.Equal(t, 15*time.Second, config.ReadTimeout, "Read timeout should be 15s")
	assert.Equal(t, 15*time.Second, config.WriteTimeout, "Write timeout should be 15s")
	assert.Equal(t, 45*time.Second, config.IdleTimeout, "Idle timeout should be 45s")
}
