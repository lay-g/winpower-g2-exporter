// Package mocks provides mock implementations for testing
package mocks

import (
	"sync"

	"github.com/lay-g/winpower-g2-exporter/internal/storage"
)

// MockStorage 模拟存储管理器
type MockStorage struct {
	data  map[string]*storage.PowerData
	mutex sync.RWMutex

	// 用于测试的钩子
	WriteFunc func(deviceID string, data *storage.PowerData) error
	ReadFunc  func(deviceID string) (*storage.PowerData, error)
}

// NewMockStorage 创建模拟存储管理器
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]*storage.PowerData),
	}
}

// Write 写入设备电能数据
func (m *MockStorage) Write(deviceID string, data *storage.PowerData) error {
	if m.WriteFunc != nil {
		return m.WriteFunc(deviceID, data)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 复制数据以避免外部修改
	m.data[deviceID] = &storage.PowerData{
		Timestamp: data.Timestamp,
		EnergyWH:  data.EnergyWH,
	}

	return nil
}

// Read 读取设备电能数据
func (m *MockStorage) Read(deviceID string) (*storage.PowerData, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(deviceID)
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data, exists := m.data[deviceID]
	if !exists {
		return nil, storage.ErrFileNotFound
	}

	// 返回数据副本
	return &storage.PowerData{
		Timestamp: data.Timestamp,
		EnergyWH:  data.EnergyWH,
	}, nil
}

// GetData 获取所有存储的数据（用于测试验证）
func (m *MockStorage) GetData() map[string]*storage.PowerData {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[string]*storage.PowerData)
	for k, v := range m.data {
		result[k] = &storage.PowerData{
			Timestamp: v.Timestamp,
			EnergyWH:  v.EnergyWH,
		}
	}

	return result
}

// Clear 清空所有数据（用于测试重置）
func (m *MockStorage) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data = make(map[string]*storage.PowerData)
}
