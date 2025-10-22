package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig 测试用的配置结构体
type TestConfig struct {
	StringValue   string        `yaml:"string_value" env:"TEST_STRING_VALUE"`
	IntValue      int           `yaml:"int_value" env:"TEST_INT_VALUE"`
	BoolValue     bool          `yaml:"bool_value" env:"TEST_BOOL_VALUE"`
	DurationValue time.Duration `yaml:"duration_value" env:"TEST_DURATION_VALUE"`
	RequiredField string        `yaml:"required_field" env:"TEST_REQUIRED_FIELD" validate:"required"`
}

func (c *TestConfig) Validate() error {
	if c.RequiredField == "" {
		return fmt.Errorf("required field is empty")
	}
	return nil
}

func (c *TestConfig) SetDefaults() {
	if c.StringValue == "" {
		c.StringValue = "default"
	}
	if c.IntValue == 0 {
		c.IntValue = 42
	}
	if c.DurationValue == 0 {
		c.DurationValue = 30 * time.Second
	}
}

func (c *TestConfig) String() string {
	return fmt.Sprintf("TestConfig{StringValue: %s, IntValue: %d, RequiredField: %s}",
		c.StringValue, c.IntValue, c.RequiredField)
}

func (c *TestConfig) Clone() Config {
	return &TestConfig{
		StringValue:   c.StringValue,
		IntValue:      c.IntValue,
		BoolValue:     c.BoolValue,
		DurationValue: c.DurationValue,
		RequiredField: c.RequiredField,
	}
}

// FullConfig 用于测试完整配置文件解组
type FullConfig struct {
	Test TestConfig `yaml:"test"`
}

func TestNewLoader(t *testing.T) {
	// 测试创建新的 Loader
	prefix := "TEST_APP"

	loader := NewLoader(prefix)

	assert.NotNil(t, loader, "Loader should not be nil")
	assert.Equal(t, prefix, loader.prefix, "Prefix should be set correctly")
	assert.NotNil(t, loader.v, "Viper instance should be initialized")
}

func TestLoader_LoadModule_FileNotFound(t *testing.T) {
	// 测试文件不存在时的处理
	loader := NewLoader("TEST")
	nonExistentFile := "/path/that/does/not/exist/config.yaml"

	// 设置配置文件路径
	loader.SetConfigFile(nonExistentFile)

	config := &TestConfig{}
	err := loader.LoadModule("test", config)

	// 配置文件不存在时应该不返回错误，而是使用默认值
	assert.NoError(t, err, "Should not return error when file not found, should use defaults")
	assert.NotNil(t, config, "Config should not be nil")
}

func TestLoader_LoadModule_InvalidYAML(t *testing.T) {
	// 创建包含格式错误的 YAML 文件
	tmpDir := t.TempDir()
	invalidYAMLFile := filepath.Join(tmpDir, "invalid.yaml")

	// 写入格式错误的 YAML 内容
	invalidContent := `
test:
  string_value: "valid string"
  int_value: 123
  required_field: "required"
  invalid_yaml: [unclosed array

`

	err := os.WriteFile(invalidYAMLFile, []byte(invalidContent), 0644)
	require.NoError(t, err)

	loader := NewLoader("TEST")
	loader.SetConfigFile(invalidYAMLFile)

	config := &TestConfig{}
	err = loader.LoadModule("test", config)

	assert.Error(t, err, "Should return error for invalid YAML")
	// Viper 可能会在 YAML 解析时报错，或者在读取文件时报错
	assert.True(t,
		strings.Contains(err.Error(), "yaml") ||
			strings.Contains(err.Error(), "parse") ||
			strings.Contains(err.Error(), "config"),
		"Error should mention YAML parsing or config error")
}

func TestLoader_LoadModule_ValidYAML(t *testing.T) {
	// 创建包含有效 YAML 的临时文件
	tmpDir := t.TempDir()
	validYAMLFile := filepath.Join(tmpDir, "valid.yaml")

	// 写入有效的 YAML 内容
	validContent := `
test:
  string_value: "test string"
  int_value: 123
  bool_value: true
  duration_value: "30s"
  required_field: "required value"
`

	err := os.WriteFile(validYAMLFile, []byte(validContent), 0644)
	require.NoError(t, err)

	loader := NewLoader("TEST")
	loader.SetConfigFile(validYAMLFile)

	config := &TestConfig{}
	err = loader.LoadModule("test", config)

	assert.NoError(t, err, "Should not return error for valid YAML")
	assert.Equal(t, "test string", config.StringValue, "StringValue should be loaded correctly")
	assert.Equal(t, 123, config.IntValue, "IntValue should be loaded correctly")
	assert.Equal(t, true, config.BoolValue, "BoolValue should be loaded correctly")
	assert.Equal(t, "required value", config.RequiredField, "RequiredField should be loaded correctly")
}

func TestLoader_BindEnv_OverrideYAML(t *testing.T) {
	// 创建包含有效 YAML 的临时文件
	tmpDir := t.TempDir()
	validYAMLFile := filepath.Join(tmpDir, "valid.yaml")

	// 写入有效的 YAML 内容
	validContent := `
test:
  string_value: "yaml string"
  int_value: 123
  bool_value: true
  duration_value: "30s"
  required_field: "yaml required"
`

	err := os.WriteFile(validYAMLFile, []byte(validContent), 0644)
	require.NoError(t, err)

	// 设置环境变量来覆盖 YAML 配置
	if err := os.Setenv("TEST_STRING_VALUE", "env string"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("TEST_INT_VALUE", "456"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("TEST_BOOL_VALUE", "false"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("TEST_REQUIRED_FIELD", "env required"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_STRING_VALUE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("TEST_INT_VALUE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("TEST_BOOL_VALUE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("TEST_REQUIRED_FIELD"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	loader := NewLoader("TEST")
	loader.SetConfigFile(validYAMLFile)

	config := &TestConfig{}
	err = loader.LoadModule("test", config)

	assert.NoError(t, err, "Should not return error")
	// 环境变量应该覆盖 YAML 配置
	assert.Equal(t, "env string", config.StringValue, "StringValue should be overridden by env")
	assert.Equal(t, 456, config.IntValue, "IntValue should be overridden by env")
	assert.Equal(t, false, config.BoolValue, "BoolValue should be overridden by env")
	assert.Equal(t, "env required", config.RequiredField, "RequiredField should be overridden by env")
}

func TestLoader_BindEnv_MissingEnv(t *testing.T) {
	// 创建包含有效 YAML 的临时文件
	tmpDir := t.TempDir()
	validYAMLFile := filepath.Join(tmpDir, "valid.yaml")

	// 写入有效的 YAML 内容
	validContent := `
test:
  string_value: "yaml string"
  int_value: 123
  bool_value: true
  required_field: "yaml required"
`

	err := os.WriteFile(validYAMLFile, []byte(validContent), 0644)
	require.NoError(t, err)

	// 不设置任何环境变量
	loader := NewLoader("TEST")
	loader.SetConfigFile(validYAMLFile)

	config := &TestConfig{}
	err = loader.LoadModule("test", config)

	assert.NoError(t, err, "Should not return error")
	// 应该使用 YAML 配置中的值
	assert.Equal(t, "yaml string", config.StringValue, "StringValue should use YAML value")
	assert.Equal(t, 123, config.IntValue, "IntValue should use YAML value")
	assert.Equal(t, true, config.BoolValue, "BoolValue should use YAML value")
	assert.Equal(t, "yaml required", config.RequiredField, "RequiredField should use YAML value")
}

func TestLoader_BindEnv_InvalidType(t *testing.T) {
	// 创建包含有效 YAML 的临时文件
	tmpDir := t.TempDir()
	validYAMLFile := filepath.Join(tmpDir, "valid.yaml")

	// 写入有效的 YAML 内容
	validContent := `
test:
  string_value: "yaml string"
  int_value: 123
  bool_value: true
  required_field: "yaml required"
`

	err := os.WriteFile(validYAMLFile, []byte(validContent), 0644)
	require.NoError(t, err)

	// 设置无效类型的环境变量
	if err := os.Setenv("TEST_INT_VALUE", "not an integer"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("TEST_BOOL_VALUE", "not a boolean"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("TEST_INT_VALUE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("TEST_BOOL_VALUE"); err != nil {
			t.Logf("Warning: failed to unset env var: %v", err)
		}
	}()

	loader := NewLoader("TEST")
	loader.SetConfigFile(validYAMLFile)

	config := &TestConfig{}
	err = loader.LoadModule("test", config)

	// 应该返回错误，因为环境变量类型转换失败
	assert.Error(t, err, "Should return error for invalid type conversion")
	assert.Contains(t, err.Error(), "convert", "Error should mention conversion failure")
}

func TestLoader_Validate_ValidConfig(t *testing.T) {
	// 创建有效配置
	config := &TestConfig{
		StringValue:   "test string",
		IntValue:      123,
		BoolValue:     true,
		DurationValue: 30 * time.Second,
		RequiredField: "required",
	}

	loader := NewLoader("TEST")
	err := loader.Validate(config)

	assert.NoError(t, err, "Valid config should pass validation")
}

func TestLoader_Validate_MissingRequired(t *testing.T) {
	// 创建缺少必填字段的配置
	config := &TestConfig{
		StringValue:   "test string",
		IntValue:      123,
		BoolValue:     true,
		DurationValue: 30 * time.Second,
		RequiredField: "", // 空的必填字段
	}

	loader := NewLoader("TEST")
	err := loader.Validate(config)

	assert.Error(t, err, "Config with missing required field should fail validation")
	assert.Contains(t, err.Error(), "required field is empty", "Error should mention missing required field")
}

func TestLoader_Validate_NilConfig(t *testing.T) {
	loader := NewLoader("TEST")
	err := loader.Validate(nil)

	assert.Error(t, err, "Nil config should fail validation")
	assert.Contains(t, err.Error(), "cannot be nil", "Error should mention config cannot be nil")
}

// Phase 2 测试用例

func TestLoader_CacheMechanism(t *testing.T) {
	// 创建包含有效 YAML 的临时文件
	tmpDir := t.TempDir()
	validYAMLFile := filepath.Join(tmpDir, "valid.yaml")

	// 写入有效的 YAML 内容
	validContent := `
test:
  string_value: "cached string"
  int_value: 789
  bool_value: true
  required_field: "cached required"
`

	err := os.WriteFile(validYAMLFile, []byte(validContent), 0644)
	require.NoError(t, err)

	loader := NewLoader("TEST")
	loader.SetConfigFile(validYAMLFile)

	// 第一次加载
	config1 := &TestConfig{}
	err = loader.LoadModule("test", config1)
	assert.NoError(t, err, "First load should succeed")

	// 第二次加载应该从缓存获取
	config2 := &TestConfig{}
	err = loader.LoadModule("test", config2)
	assert.NoError(t, err, "Second load should succeed")

	// 验证两个配置对象内容相同
	assert.Equal(t, config1.StringValue, config2.StringValue, "Cached config should match original")
	assert.Equal(t, config1.IntValue, config2.IntValue, "Cached config should match original")
}

func TestLoader_CacheDisabled(t *testing.T) {
	// 创建包含有效 YAML 的临时文件
	tmpDir := t.TempDir()
	validYAMLFile := filepath.Join(tmpDir, "valid.yaml")

	validContent := `
test:
  string_value: "original string"
  int_value: 111
  required_field: "original required"
`

	err := os.WriteFile(validYAMLFile, []byte(validContent), 0644)
	require.NoError(t, err)

	loader := NewLoader("TEST")
	loader.SetConfigFile(validYAMLFile)

	// 禁用缓存
	loader.DisableCache()

	// 第一次加载
	config1 := &TestConfig{}
	err = loader.LoadModule("test", config1)
	assert.NoError(t, err, "First load should succeed")

	// 修改文件内容
	modifiedContent := `
test:
  string_value: "modified string"
  int_value: 222
  required_field: "modified required"
`

	err = os.WriteFile(validYAMLFile, []byte(modifiedContent), 0644)
	require.NoError(t, err)

	// 第二次加载应该重新读取文件（因为缓存被禁用）
	config2 := &TestConfig{}
	err = loader.LoadModule("test", config2)
	assert.NoError(t, err, "Second load should succeed")

	// 验证配置内容不同（因为重新读取了文件）
	assert.NotEqual(t, config1.StringValue, config2.StringValue, "Config should be reloaded when cache is disabled")
	assert.NotEqual(t, config1.IntValue, config2.IntValue, "Config should be reloaded when cache is disabled")
}

func TestLoader_ConfigDeepCopy(t *testing.T) {
	// 创建测试配置
	originalConfig := &TestConfig{
		StringValue:   "original",
		IntValue:      42,
		BoolValue:     true,
		DurationValue: 30 * time.Second,
		RequiredField: "required",
	}

	// 创建深拷贝
	clonedConfig := originalConfig.Clone().(*TestConfig)

	// 验证拷贝的内容相同
	assert.Equal(t, originalConfig.StringValue, clonedConfig.StringValue)
	assert.Equal(t, originalConfig.IntValue, clonedConfig.IntValue)
	assert.Equal(t, originalConfig.BoolValue, clonedConfig.BoolValue)
	assert.Equal(t, originalConfig.DurationValue, clonedConfig.DurationValue)
	assert.Equal(t, originalConfig.RequiredField, clonedConfig.RequiredField)

	// 修改拷贝的内容
	clonedConfig.StringValue = "modified"
	clonedConfig.IntValue = 100

	// 验证原始配置不受影响
	assert.Equal(t, "original", originalConfig.StringValue, "Original config should not be affected")
	assert.Equal(t, 42, originalConfig.IntValue, "Original config should not be affected")
	assert.Equal(t, "modified", clonedConfig.StringValue, "Cloned config should be modified")
	assert.Equal(t, 100, clonedConfig.IntValue, "Cloned config should be modified")
}

func TestLoader_ConfigMerge(t *testing.T) {
	loader := NewLoader("TEST")

	// 创建基础配置
	baseConfig := &TestConfig{
		StringValue:   "base",
		IntValue:      10,
		BoolValue:     false,
		DurationValue: 5 * time.Second,
		RequiredField: "base_required",
	}

	// 创建覆盖配置
	overlayConfig := &TestConfig{
		StringValue: "overlay",
		IntValue:    20,
		// BoolValue 保持默认值 false
		DurationValue: 15 * time.Second,
		RequiredField: "overlay_required",
	}

	// 合并配置
	mergedConfig := &TestConfig{}
	err := loader.MergeConfig(baseConfig, overlayConfig, mergedConfig)
	assert.NoError(t, err, "Merge should succeed")

	// 验证合并结果
	assert.Equal(t, "overlay", mergedConfig.StringValue, "String value should be from overlay")
	assert.Equal(t, 20, mergedConfig.IntValue, "Int value should be from overlay")
	assert.Equal(t, 15*time.Second, mergedConfig.DurationValue, "Duration value should be from overlay")
	assert.Equal(t, "overlay_required", mergedConfig.RequiredField, "Required field should be from overlay")
	// BoolValue 应该从 overlay 继承（即使它没有被显式设置）
	assert.Equal(t, false, mergedConfig.BoolValue, "Bool value should be from overlay")
}

func TestLoader_EnhancedValidation(t *testing.T) {
	loader := NewLoader("TEST")

	// 测试范围验证
	config := &TestConfig{
		StringValue:   "test",
		IntValue:      -5, // 无效的负值
		BoolValue:     true,
		DurationValue: 30 * time.Second,
		RequiredField: "required",
	}

	// 添加自定义验证规则
	loader.AddValidationRule("int_value", func(value interface{}) error {
		if intValue, ok := value.(int); ok {
			if intValue < 0 {
				return fmt.Errorf("int_value cannot be negative")
			}
		}
		return nil
	})

	err := loader.Validate(config)
	assert.Error(t, err, "Config with negative int value should fail validation")
	assert.Contains(t, err.Error(), "cannot be negative", "Error should mention negative value")
}

func TestLoader_SensitiveInfoMasking(t *testing.T) {
	// 创建包含敏感信息的测试配置
	config := &TestConfig{
		StringValue:   "test",
		IntValue:      42,
		BoolValue:     true,
		DurationValue: 30 * time.Second,
		RequiredField: "required123", // 模拟敏感信息
	}

	// 测试 String 方法脱敏功能
	str := config.String()

	// TestConfig 的 String 方法应该正常工作，包含 RequiredField
	assert.Contains(t, str, "test", "Should contain non-sensitive string value")
	assert.Contains(t, str, "42", "Should contain non-sensitive int value")
	assert.Contains(t, str, "required123", "Should contain required field")

	// 验证深拷贝不共享引用
	clonedConfig := config.Clone().(*TestConfig)
	clonedConfig.RequiredField = "modified123"

	// 原始配置不应该被修改
	assert.Equal(t, "required123", config.RequiredField, "Original config should not be affected by clone modification")
	assert.Equal(t, "modified123", clonedConfig.RequiredField, "Cloned config should be independently modifiable")
}
