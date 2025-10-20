package storage

import (
	"strconv"
	"testing"
)

// MockConfigLoader 是一个用于测试的 ConfigLoader 实现
type MockConfigLoader struct {
	data map[string]interface{}
}

func NewMockConfigLoader() *MockConfigLoader {
	return &MockConfigLoader{
		data: make(map[string]interface{}),
	}
}

func (m *MockConfigLoader) Set(key string, value interface{}) {
	m.data[key] = value
}

func (m *MockConfigLoader) LoadString(key string) (string, bool) {
	if value, exists := m.data[key]; exists {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

func (m *MockConfigLoader) LoadInt(key string) (int, bool) {
	if value, exists := m.data[key]; exists {
		switch v := value.(type) {
		case int:
			return v, true
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i, true
			}
		}
	}
	return 0, false
}

func (m *MockConfigLoader) LoadBool(key string) (bool, bool) {
	if value, exists := m.data[key]; exists {
		switch v := value.(type) {
		case bool:
			return v, true
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b, true
			}
		}
	}
	return false, false
}

func (m *MockConfigLoader) LoadStringSlice(key string) ([]string, bool) {
	if value, exists := m.data[key]; exists {
		if slice, ok := value.([]string); ok {
			return slice, true
		}
	}
	return nil, false
}

// TestConfigLoader 测试 ConfigLoader 接口的基本功能
func TestConfigLoader(t *testing.T) {
	loader := NewMockConfigLoader()

	// 设置测试数据
	loader.Set("data_dir", "/tmp/test")
	loader.Set("sync_write", true)
	loader.Set("file_permissions", "0644")

	// 测试 LoadString
	if value, found := loader.LoadString("data_dir"); !found || value != "/tmp/test" {
		t.Errorf("LoadString failed: got %v, %v", value, found)
	}

	// 测试 LoadBool
	if value, found := loader.LoadBool("sync_write"); !found || !value {
		t.Errorf("LoadBool failed: got %v, %v", value, found)
	}

	// 测试 LoadString (用于文件权限)
	if value, found := loader.LoadString("file_permissions"); !found || value != "0644" {
		t.Errorf("LoadString for file_permissions failed: got %v, %v", value, found)
	}
}

// TestNewConfigFromLoader 测试从 ConfigLoader 创建配置
func TestNewConfigFromLoader(t *testing.T) {
	loader := NewMockConfigLoader()
	loader.Set("data_dir", "/tmp/loader-test")
	loader.Set("sync_write", true)
	loader.Set("create_dir", false)
	loader.Set("file_permissions", "0640")
	loader.Set("dir_permissions", "0750")

	config := NewConfigFromLoader(loader)

	if config.DataDir != "/tmp/loader-test" {
		t.Errorf("DataDir = %s, expected /tmp/loader-test", config.DataDir)
	}

	if !config.SyncWrite {
		t.Error("SyncWrite should be true")
	}

	if config.CreateDir {
		t.Error("CreateDir should be false")
	}

	if config.FilePermissions != 0640 {
		t.Errorf("FilePermissions = %o, expected 0640", config.FilePermissions)
	}

	if config.DirPermissions != 0750 {
		t.Errorf("DirPermissions = %o, expected 0750", config.DirPermissions)
	}
}

// TestNewConfigFromLoader_Nil 测试 nil loader 的情况
func TestNewConfigFromLoader_Nil(t *testing.T) {
	config := NewConfigFromLoader(nil)

	// 应该返回基于环境变量的配置
	if config == nil {
		t.Fatal("NewConfigFromLoader(nil) should not return nil")
	}

	// 应该有默认值
	if config.DataDir == "" {
		t.Error("DataDir should have a default value")
	}
}

// TestNewFileStorageManagerFromLoader 测试从 ConfigLoader 创建存储管理器
func TestNewFileStorageManagerFromLoader(t *testing.T) {
	loader := NewMockConfigLoader()
	loader.Set("data_dir", t.TempDir())
	loader.Set("sync_write", true)
	loader.Set("create_dir", true)

	manager, err := NewFileStorageManagerFromLoader(loader)
	if err != nil {
		t.Fatalf("NewFileStorageManagerFromLoader() error = %v", err)
	}

	if manager == nil {
		t.Fatal("NewFileStorageManagerFromLoader() returned nil")
	}

	// 测试存储管理器是否正常工作
	deviceID := "test-device"
	data := NewPowerData(100.0)

	err = manager.Write(deviceID, data)
	if err != nil {
		t.Errorf("Manager Write() error = %v", err)
	}

	readData, err := manager.Read(deviceID)
	if err != nil {
		t.Errorf("Manager Read() error = %v", err)
	}

	if readData.EnergyWH != data.EnergyWH {
		t.Error("Manager read/write consistency failed")
	}
}