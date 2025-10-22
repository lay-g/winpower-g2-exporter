package main

import (
	"fmt"
	"log"
	"os"
	"time"

	cfg "github.com/lay-g/winpower-g2-exporter/internal/pkgs/config"
)

// ExampleConfig demonstrates Phase 2 configuration features
type ExampleConfig struct {
	Name       string        `yaml:"name" env:"EXAMPLE_NAME"`
	Timeout    time.Duration `yaml:"timeout" env:"EXAMPLE_TIMEOUT"`
	RetryCount int           `yaml:"retry_count" env:"EXAMPLE_RETRY_COUNT"`
	Enabled    bool          `yaml:"enabled" env:"EXAMPLE_ENABLED"`
	Password   string        `yaml:"password" env:"EXAMPLE_PASSWORD"`
	APIKey     string        `yaml:"api_key" env:"EXAMPLE_API_KEY"`
}

// Implement the Config interface
func (c *ExampleConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if c.RetryCount < 0 || c.RetryCount > 10 {
		return fmt.Errorf("retry_count must be between 0 and 10")
	}
	return nil
}

func (c *ExampleConfig) SetDefaults() {
	if c.Name == "" {
		c.Name = "example"
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.RetryCount == 0 {
		c.RetryCount = 3
	}
	if c.Password == "" {
		c.Password = "default123"
	}
	if c.APIKey == "" {
		c.APIKey = "default-api-key"
	}
}

func (c *ExampleConfig) String() string {
	// Mask sensitive information
	maskedPassword := maskSensitive(c.Password)
	maskedAPIKey := maskSensitive(c.APIKey)

	return fmt.Sprintf("ExampleConfig{Name: %s, Timeout: %v, RetryCount: %d, Enabled: %t, Password: %s, APIKey: %s}",
		c.Name, c.Timeout, c.RetryCount, c.Enabled, maskedPassword, maskedAPIKey)
}

func (c *ExampleConfig) Clone() cfg.Config {
	return &ExampleConfig{
		Name:       c.Name,
		Timeout:    c.Timeout,
		RetryCount: c.RetryCount,
		Enabled:    c.Enabled,
		Password:   c.Password,
		APIKey:     c.APIKey,
	}
}

func maskSensitive(value string) string {
	if value == "" {
		return "<empty>"
	}
	if len(value) <= 2 {
		return "***"
	}
	return value[:1] + "***" + value[len(value)-1:]
}

func main() {
	fmt.Println("=== Config Module Phase 2 Features Demo ===")

	// 1. Create a new config loader
	loader := cfg.NewLoader("EXAMPLE")

	// 2. Demonstrate enhanced validation with custom rules
	fmt.Println("\n--- Enhanced Validation ---")
	loader.AddValidationRule("timeout", func(value interface{}) error {
		if timeout, ok := value.(time.Duration); ok {
			if timeout > 300*time.Second {
				return fmt.Errorf("timeout should not exceed 5 minutes")
			}
		}
		return nil
	})

	loader.AddValidationRule("retry_count", func(value interface{}) error {
		if retryCount, ok := value.(int); ok {
			if retryCount > 5 {
				return fmt.Errorf("retry_count should not exceed 5 for performance")
			}
		}
		return nil
	})

	// 3. Create and validate configuration
	config := &ExampleConfig{
		Name:       "demo-app",
		Timeout:    60 * time.Second,
		RetryCount: 7, // This will fail our custom validation
		Enabled:    true,
		Password:   "secret123",
		APIKey:     "api-key-456",
	}

	fmt.Printf("Original config: %s\n", config.String())

	// Validate with custom rules
	if err := loader.Validate(config); err != nil {
		fmt.Printf("Validation error: %v\n", err)
	}

	// Fix the validation issue
	config.RetryCount = 3
	if err := loader.Validate(config); err == nil {
		fmt.Println("✓ Configuration passed validation")
	}

	// 4. Demonstrate deep cloning
	fmt.Println("\n--- Deep Cloning ---")
	clonedConfig := config.Clone().(*ExampleConfig)
	fmt.Printf("Cloned config: %s\n", clonedConfig.String())

	// Modify the clone
	clonedConfig.Name = "modified-app"
	clonedConfig.Password = "different-password"

	fmt.Printf("After modification:\n")
	fmt.Printf("Original: %s\n", config.String())
	fmt.Printf("Cloned:   %s\n", clonedConfig.String())

	// 5. Demonstrate configuration merging
	fmt.Println("\n--- Configuration Merging ---")
	baseConfig := &ExampleConfig{
		Name:       "base-app",
		Timeout:    30 * time.Second,
		RetryCount: 2,
		Enabled:    false,
		Password:   "base-password",
	}

	overlayConfig := &ExampleConfig{
		Name:    "overlay-app", // This will override base
		Enabled: true,          // This will override base
		// Timeout and RetryCount will be inherited from base
		// Password will be inherited from base
	}

	mergedConfig := &ExampleConfig{}
	if err := loader.MergeConfig(baseConfig, overlayConfig, mergedConfig); err != nil {
		log.Printf("Merge error: %v", err)
	} else {
		fmt.Printf("Base config:   %s\n", baseConfig.String())
		fmt.Printf("Overlay config: %s\n", overlayConfig.String())
		fmt.Printf("Merged config: %s\n", mergedConfig.String())
	}

	// 6. Demonstrate caching mechanism
	fmt.Println("\n--- Caching Mechanism ---")
	fmt.Printf("Cache enabled: %v\n", loader.IsCacheEnabled())

	// Load configuration multiple times
	for i := 0; i < 3; i++ {
		testConfig := &ExampleConfig{}
		// Simulate loading (this would normally load from file)
		testConfig.Name = fmt.Sprintf("cached-app-%d", i)
		testConfig.SetDefaults()

		fmt.Printf("Load %d: %s\n", i+1, testConfig.String())
	}

	// 7. Demonstrate cache control
	fmt.Println("\n--- Cache Control ---")
	loader.DisableCache()
	fmt.Printf("Cache disabled: %v\n", !loader.IsCacheEnabled())

	loader.EnableCache()
	fmt.Printf("Cache re-enabled: %v\n", loader.IsCacheEnabled())

	loader.ClearCache()
	fmt.Println("Cache cleared")

	// 8. Demonstrate environment variable override
	fmt.Println("\n--- Environment Variable Override ---")

	// Set some environment variables
	_ = os.Setenv("EXAMPLE_NAME", "env-override-app")
	_ = os.Setenv("EXAMPLE_TIMEOUT", "45s")
	_ = os.Setenv("EXAMPLE_RETRY_COUNT", "5")
	_ = os.Setenv("EXAMPLE_ENABLED", "true")
	_ = os.Setenv("EXAMPLE_PASSWORD", "env-secret")

	// Create a new loader to test environment variables
	envLoader := cfg.NewLoader("EXAMPLE")
	envConfig := &ExampleConfig{}
	envConfig.SetDefaults()

	// Note: In a real scenario, you would use LoadModule to load from config file
	// and apply environment variable overrides automatically
	// For demo purposes, we manually set values that would come from environment
	envConfig.Name = "env-override-app"
	envConfig.Timeout = 45 * time.Second
	envConfig.RetryCount = 5
	envConfig.Enabled = true
	envConfig.Password = "env-secret"

	fmt.Printf("Environment config: %s\n", envConfig.String())
	fmt.Printf("Environment loader cache enabled: %v\n", envLoader.IsCacheEnabled())

	// Clean up environment variables
	_ = os.Unsetenv("EXAMPLE_NAME")
	_ = os.Unsetenv("EXAMPLE_TIMEOUT")
	_ = os.Unsetenv("EXAMPLE_RETRY_COUNT")
	_ = os.Unsetenv("EXAMPLE_ENABLED")
	_ = os.Unsetenv("EXAMPLE_PASSWORD")

	fmt.Println("\n=== Demo Complete ===")
}
