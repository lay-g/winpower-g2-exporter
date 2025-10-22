package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// MockConfig 用于测试的配置实现
type MockConfig struct {
	data  map[string]interface{}
	valid bool
}

func (m *MockConfig) Validate() error {
	if !m.valid {
		return errors.New("config is invalid")
	}
	return nil
}

func (m *MockConfig) SetDefaults() {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	if _, exists := m.data["field1"]; !exists {
		m.data["field1"] = "default_value"
	}
}

func (m *MockConfig) String() string {
	if m.data == nil {
		return "MockConfig{...}"
	}

	// 脱敏处理敏感信息
	var parts []string
	for k, v := range m.data {
		// 检查是否是敏感字段
		if isSensitiveField(k) {
			parts = append(parts, fmt.Sprintf("%s: ***", k))
		} else {
			parts = append(parts, fmt.Sprintf("%s: %v", k, v))
		}
	}

	if len(parts) == 0 {
		return "MockConfig{...}"
	}

	return fmt.Sprintf("MockConfig{%s}", strings.Join(parts, ", "))
}

// isSensitiveField 检查字段名是否为敏感字段
func isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{"password", "passwd", "secret", "token", "key", "api_key", "apikey"}
	lowerFieldName := strings.ToLower(fieldName)

	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerFieldName, sensitive) {
			return true
		}
	}
	return false
}

func (m *MockConfig) Clone() Config {
	// 深拷贝 data map
	newData := make(map[string]interface{})
	for k, v := range m.data {
		newData[k] = v
	}

	return &MockConfig{
		data:  newData,
		valid: m.valid,
	}
}

// MockProvider 用于测试的提供者实现
type MockProvider struct {
	config     Config
	configPath string
	loadError  error
}

func (m *MockProvider) Load() (Config, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	return m.config, nil
}

func (m *MockProvider) LoadFromEnv() Config {
	return m.config
}

func (m *MockProvider) GetConfigPath() string {
	return m.configPath
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config should not return error",
			config:  &MockConfig{valid: true},
			wantErr: false,
		},
		{
			name:    "invalid config should return error",
			config:  &MockConfig{valid: false},
			wantErr: true,
			errMsg:  "config is invalid",
		},
		{
			name:    "empty config should return error",
			config:  &MockConfig{data: nil, valid: false},
			wantErr: true,
			errMsg:  "config is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, expected to contain %s", err, tt.errMsg)
				}
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	mock := &MockConfig{data: make(map[string]interface{})}

	// 在设置默认值前，字段应该不存在
	if _, exists := mock.data["field1"]; exists {
		t.Errorf("Expected field1 to not exist before SetDefaults")
	}

	// 调用 SetDefaults
	mock.SetDefaults()

	// 设置默认值后，字段应该存在且有默认值
	if value, exists := mock.data["field1"]; !exists {
		t.Errorf("Expected field1 to exist after SetDefaults")
	} else if value != "default_value" {
		t.Errorf("Expected field1 to be 'default_value', got %v", value)
	}
}

func TestConfig_String(t *testing.T) {
	config := &MockConfig{}

	str := config.String()
	if str == "" {
		t.Errorf("Config.String() should not return empty string")
	}

	expected := "MockConfig{...}"
	if str != expected {
		t.Errorf("Config.String() = %v, want %v", str, expected)
	}

	// 测试包含敏感信息的配置
	sensitiveConfig := &MockConfig{
		data: map[string]interface{}{
			"username": "admin",
			"password": "secret123",     // 敏感信息
			"api_key":  "sk-1234567890", // 敏感信息
			"host":     "example.com",
		},
	}

	sensitiveStr := sensitiveConfig.String()

	// 验证字符串表示中不包含明文密码
	if strings.Contains(sensitiveStr, "secret123") {
		t.Errorf("Config.String() should not expose sensitive information like passwords")
	}

	if strings.Contains(sensitiveStr, "sk-1234567890") {
		t.Errorf("Config.String() should not expose sensitive information like API keys")
	}

	// 验证非敏感信息仍然存在用于调试
	if !strings.Contains(sensitiveStr, "admin") && !strings.Contains(sensitiveStr, "example.com") {
		t.Errorf("Config.String() should include non-sensitive information for debugging")
	}
}

func TestConfig_Clone(t *testing.T) {
	// 测试接口是否可以被正确实现
	var _ Config = &MockConfig{}

	// 测试深拷贝功能
	original := &MockConfig{
		data: map[string]interface{}{
			"field1": "value1",
			"field2": 42,
		},
		valid: true,
	}

	// 创建克隆
	cloned := original.Clone()

	// 验证克隆对象具有相同的值
	clonedConfig := cloned.(*MockConfig)
	if clonedConfig.data["field1"] != original.data["field1"] {
		t.Errorf("Clone should copy field1, got %v, want %v",
			clonedConfig.data["field1"], original.data["field1"])
	}

	if clonedConfig.data["field2"] != original.data["field2"] {
		t.Errorf("Clone should copy field2, got %v, want %v",
			clonedConfig.data["field2"], original.data["field2"])
	}

	if clonedConfig.valid != original.valid {
		t.Errorf("Clone should copy valid flag, got %v, want %v",
			clonedConfig.valid, original.valid)
	}

	// 验证是深拷贝：修改克隆不应影响原始对象
	clonedConfig.data["field1"] = "modified_value"
	clonedConfig.valid = false

	if original.data["field1"] == "modified_value" {
		t.Errorf("Clone should be deep copy, modifying clone should not affect original")
	}

	if original.valid == false {
		t.Errorf("Clone should be deep copy, modifying clone should not affect original")
	}

	// 通过反射检查 Clone 方法是否存在
	configType := reflect.TypeOf(original)
	cloneMethod, _ := configType.MethodByName("Clone")
	if cloneMethod.Name != "Clone" {
		t.Errorf("Expected Clone method to exist")
	}
}

func TestProvider_Load(t *testing.T) {
	tests := []struct {
		name      string
		provider  Provider
		wantErr   bool
		expectNil bool
	}{
		{
			name: "successful load should not return error",
			provider: &MockProvider{
				config: &MockConfig{valid: true},
			},
			wantErr:   false,
			expectNil: false,
		},
		{
			name: "failed load should return error",
			provider: &MockProvider{
				loadError: errors.New("load failed"),
			},
			wantErr:   true,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := tt.provider.Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expectNil && config != nil {
				t.Errorf("Provider.Load() config = %v, want nil", config)
			}
			if !tt.expectNil && config == nil {
				t.Errorf("Provider.Load() config = nil, want non-nil")
			}
		})
	}
}

func TestProvider_LoadFromEnv(t *testing.T) {
	expectedConfig := &MockConfig{valid: true}
	provider := &MockProvider{
		config: expectedConfig,
	}

	config := provider.LoadFromEnv()
	if config == nil {
		t.Errorf("Provider.LoadFromEnv() should not return nil")
	}

	if config != expectedConfig {
		t.Errorf("Provider.LoadFromEnv() = %v, want %v", config, expectedConfig)
	}
}

func TestProvider_GetConfigPath(t *testing.T) {
	expectedPath := "/path/to/config.yaml"
	provider := &MockProvider{
		configPath: expectedPath,
	}

	path := provider.GetConfigPath()
	if path != expectedPath {
		t.Errorf("Provider.GetConfigPath() = %v, want %v", path, expectedPath)
	}
}

func TestProvider_InterfaceCompatibility(t *testing.T) {
	// 测试 Provider 接口是否可以被正确实现
	var _ Provider = &MockProvider{}

	// 通过反射检查接口方法
	provider := &MockProvider{}
	providerType := reflect.TypeOf(provider)

	loadMethod, _ := providerType.MethodByName("Load")
	loadFromEnvMethod, _ := providerType.MethodByName("LoadFromEnv")
	getConfigPathMethod, _ := providerType.MethodByName("GetConfigPath")

	if loadMethod.Name != "Load" {
		t.Errorf("Expected Load method to exist")
	}

	if loadFromEnvMethod.Name != "LoadFromEnv" {
		t.Errorf("Expected LoadFromEnv method to exist")
	}

	if getConfigPathMethod.Name != "GetConfigPath" {
		t.Errorf("Expected GetConfigPath method to exist")
	}
}
