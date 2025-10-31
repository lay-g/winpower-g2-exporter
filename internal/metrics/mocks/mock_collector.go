// Package mocks provides mock implementations for testing the metrics module
package mocks

import (
	"context"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
)

// MockCollector is a mock implementation of collector.CollectorInterface for testing
type MockCollector struct {
	CollectDeviceDataFunc func(ctx context.Context) (*collector.CollectionResult, error)
}

// CollectDeviceData implements collector.CollectorInterface
func (m *MockCollector) CollectDeviceData(ctx context.Context) (*collector.CollectionResult, error) {
	if m.CollectDeviceDataFunc != nil {
		return m.CollectDeviceDataFunc(ctx)
	}
	// Default implementation returns successful empty result
	return &collector.CollectionResult{
		Success:        true,
		DeviceCount:    0,
		Devices:        make(map[string]*collector.DeviceCollectionInfo),
		CollectionTime: time.Now(),
		Duration:       100 * time.Millisecond,
	}, nil
}

// NewMockCollector creates a new MockCollector instance
func NewMockCollector() *MockCollector {
	return &MockCollector{}
}

// NewMockCollectorWithDevices creates a mock collector that returns test devices
func NewMockCollectorWithDevices() *MockCollector {
	return &MockCollector{
		CollectDeviceDataFunc: func(ctx context.Context) (*collector.CollectionResult, error) {
			return &collector.CollectionResult{
				Success:        true,
				DeviceCount:    2,
				CollectionTime: time.Now(),
				Duration:       200 * time.Millisecond,
				Devices: map[string]*collector.DeviceCollectionInfo{
					"device1": {
						DeviceID:          "device1",
						DeviceName:        "UPS-01",
						DeviceType:        1,
						DeviceModel:       "Model-X",
						Connected:         true,
						LastUpdateTime:    time.Now(),
						InputVolt1:        220.5,
						InputFreq:         50.0,
						OutputVolt1:       220.0,
						OutputCurrent1:    5.2,
						OutputFreq:        50.0,
						OutputVoltageType: "single",
						LoadPercent:       45.0,
						LoadTotalWatt:     1000.0,
						LoadTotalVa:       1100.0,
						LoadWatt1:         1000.0,
						LoadVa1:           1100.0,
						IsCharging:        true,
						BatVoltP:          95.0,
						BatCapacity:       90.0,
						BatRemainTime:     3600,
						BatteryStatus:     "normal",
						UpsTemperature:    35.0,
						Mode:              "online",
						Status:            "normal",
						TestStatus:        "no_test",
						FaultCode:         "",
						EnergyCalculated:  true,
						EnergyValue:       12345.67,
					},
					"device2": {
						DeviceID:          "device2",
						DeviceName:        "UPS-02",
						DeviceType:        1,
						DeviceModel:       "Model-Y",
						Connected:         true,
						LastUpdateTime:    time.Now(),
						InputVolt1:        220.0,
						InputFreq:         50.0,
						OutputVolt1:       220.0,
						OutputCurrent1:    3.0,
						OutputFreq:        50.0,
						OutputVoltageType: "single",
						LoadPercent:       30.0,
						LoadTotalWatt:     600.0,
						LoadTotalVa:       650.0,
						LoadWatt1:         600.0,
						LoadVa1:           650.0,
						IsCharging:        false,
						BatVoltP:          85.0,
						BatCapacity:       80.0,
						BatRemainTime:     1800,
						BatteryStatus:     "normal",
						UpsTemperature:    32.0,
						Mode:              "online",
						Status:            "normal",
						TestStatus:        "no_test",
						FaultCode:         "",
						EnergyCalculated:  true,
						EnergyValue:       9876.54,
					},
				},
			}, nil
		},
	}
}
