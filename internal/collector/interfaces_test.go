package collector

import (
	"context"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// MockCollector is a mock implementation of CollectorInterface for testing
type MockCollector struct {
	CollectDeviceDataFunc func(ctx context.Context) (*CollectionResult, error)
}

func (m *MockCollector) CollectDeviceData(ctx context.Context) (*CollectionResult, error) {
	if m.CollectDeviceDataFunc != nil {
		return m.CollectDeviceDataFunc(ctx)
	}
	return &CollectionResult{Success: true}, nil
}

// MockWinPowerClient is a mock implementation of WinPowerClient for testing
type MockWinPowerClient struct {
	CollectDeviceDataFunc     func(ctx context.Context) ([]winpower.ParsedDeviceData, error)
	GetConnectionStatusFunc   func() bool
	GetLastCollectionTimeFunc func() time.Time
	GetTokenExpiresAtFunc     func() time.Time
	IsTokenValidFunc          func() bool
}

func (m *MockWinPowerClient) CollectDeviceData(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
	if m.CollectDeviceDataFunc != nil {
		return m.CollectDeviceDataFunc(ctx)
	}
	return []winpower.ParsedDeviceData{}, nil
}

func (m *MockWinPowerClient) GetConnectionStatus() bool {
	if m.GetConnectionStatusFunc != nil {
		return m.GetConnectionStatusFunc()
	}
	return true
}

func (m *MockWinPowerClient) GetLastCollectionTime() time.Time {
	if m.GetLastCollectionTimeFunc != nil {
		return m.GetLastCollectionTimeFunc()
	}
	return time.Now()
}

func (m *MockWinPowerClient) GetTokenExpiresAt() time.Time {
	if m.GetTokenExpiresAtFunc != nil {
		return m.GetTokenExpiresAtFunc()
	}
	return time.Now().Add(time.Hour)
}

func (m *MockWinPowerClient) IsTokenValid() bool {
	if m.IsTokenValidFunc != nil {
		return m.IsTokenValidFunc()
	}
	return true
}

// MockEnergyCalculator is a mock implementation of EnergyCalculator for testing
type MockEnergyCalculator struct {
	CalculateFunc func(deviceID string, power float64) (float64, error)
	GetFunc       func(deviceID string) (float64, error)
}

func (m *MockEnergyCalculator) Calculate(deviceID string, power float64) (float64, error) {
	if m.CalculateFunc != nil {
		return m.CalculateFunc(deviceID, power)
	}
	return 0, nil
}

func (m *MockEnergyCalculator) Get(deviceID string) (float64, error) {
	if m.GetFunc != nil {
		return m.GetFunc(deviceID)
	}
	return 0, nil
}

func TestCollectorInterface(t *testing.T) {
	mock := &MockCollector{
		CollectDeviceDataFunc: func(ctx context.Context) (*CollectionResult, error) {
			return &CollectionResult{
				Success:     true,
				DeviceCount: 1,
			}, nil
		},
	}

	ctx := context.Background()
	result, err := mock.CollectDeviceData(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !result.Success {
		t.Error("Expected success to be true")
	}
	if result.DeviceCount != 1 {
		t.Errorf("Expected device count to be 1, got %d", result.DeviceCount)
	}
}

func TestWinPowerClientInterface(t *testing.T) {
	mock := &MockWinPowerClient{
		CollectDeviceDataFunc: func(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
			return []winpower.ParsedDeviceData{
				{
					DeviceID: "test-device",
					Realtime: winpower.RealtimeData{
						LoadTotalWatt: 1000.0,
					},
				},
			}, nil
		},
		GetConnectionStatusFunc: func() bool {
			return true
		},
		GetLastCollectionTimeFunc: func() time.Time {
			return time.Now()
		},
	}

	ctx := context.Background()
	data, err := mock.CollectDeviceData(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(data) != 1 {
		t.Errorf("Expected 1 device, got %d", len(data))
	}

	status := mock.GetConnectionStatus()
	if !status {
		t.Error("Expected connection status to be true")
	}

	lastTime := mock.GetLastCollectionTime()
	if lastTime.IsZero() {
		t.Error("Expected non-zero last collection time")
	}
}

func TestEnergyCalculatorInterface(t *testing.T) {
	mock := &MockEnergyCalculator{
		CalculateFunc: func(deviceID string, power float64) (float64, error) {
			return power * 0.5, nil // Simple test calculation
		},
		GetFunc: func(deviceID string) (float64, error) {
			return 100.0, nil
		},
	}

	energy, err := mock.Calculate("test-device", 1000.0)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if energy != 500.0 {
		t.Errorf("Expected energy to be 500.0, got %f", energy)
	}

	retrieved, err := mock.Get("test-device")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if retrieved != 100.0 {
		t.Errorf("Expected retrieved energy to be 100.0, got %f", retrieved)
	}
}
