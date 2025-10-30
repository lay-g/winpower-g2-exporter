package collector

import (
	"testing"
	"time"
)

func TestCollectionResult_Structure(t *testing.T) {
	result := &CollectionResult{
		Success:        true,
		DeviceCount:    2,
		Devices:        make(map[string]*DeviceCollectionInfo),
		CollectionTime: time.Now(),
		Duration:       100 * time.Millisecond,
		ErrorMessage:   "",
	}

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.DeviceCount != 2 {
		t.Errorf("Expected DeviceCount to be 2, got %d", result.DeviceCount)
	}
	if result.Devices == nil {
		t.Error("Expected Devices map to be initialized")
	}
}

func TestDeviceCollectionInfo_Structure(t *testing.T) {
	device := &DeviceCollectionInfo{
		DeviceID:         "device123",
		DeviceName:       "UPS-001",
		DeviceType:       1,
		DeviceModel:      "G2-3000",
		Connected:        true,
		LastUpdateTime:   time.Now(),
		InputVolt1:       220.5,
		InputFreq:        50.0,
		OutputVolt1:      220.0,
		OutputCurrent1:   5.5,
		OutputFreq:       50.0,
		LoadPercent:      65.5,
		LoadTotalWatt:    1500.0,
		LoadTotalVa:      1800.0,
		IsCharging:       false,
		BatVoltP:         95.5,
		BatCapacity:      100.0,
		BatRemainTime:    3600,
		UpsTemperature:   35.5,
		Mode:             "normal",
		Status:           "online",
		EnergyCalculated: true,
		EnergyValue:      1250.5,
		ErrorMsg:         "",
	}

	// Test basic fields
	if device.DeviceID != "device123" {
		t.Errorf("Expected DeviceID to be 'device123', got %s", device.DeviceID)
	}
	if device.DeviceModel != "G2-3000" {
		t.Errorf("Expected DeviceModel to be 'G2-3000', got %s", device.DeviceModel)
	}

	// Test core power field
	if device.LoadTotalWatt != 1500.0 {
		t.Errorf("Expected LoadTotalWatt to be 1500.0, got %f", device.LoadTotalWatt)
	}

	// Test energy calculation result
	if !device.EnergyCalculated {
		t.Error("Expected EnergyCalculated to be true")
	}
	if device.EnergyValue != 1250.5 {
		t.Errorf("Expected EnergyValue to be 1250.5, got %f", device.EnergyValue)
	}
}

func TestDeviceCollectionInfo_WithError(t *testing.T) {
	device := &DeviceCollectionInfo{
		DeviceID:         "device456",
		EnergyCalculated: false,
		ErrorMsg:         "calculation failed",
	}

	if device.EnergyCalculated {
		t.Error("Expected EnergyCalculated to be false")
	}
	if device.ErrorMsg != "calculation failed" {
		t.Errorf("Expected ErrorMsg to be 'calculation failed', got %s", device.ErrorMsg)
	}
}

func TestCollectionResult_WithDevices(t *testing.T) {
	device1 := &DeviceCollectionInfo{
		DeviceID:      "device1",
		LoadTotalWatt: 1000.0,
		EnergyValue:   500.0,
	}
	device2 := &DeviceCollectionInfo{
		DeviceID:      "device2",
		LoadTotalWatt: 2000.0,
		EnergyValue:   1000.0,
	}

	result := &CollectionResult{
		Success:     true,
		DeviceCount: 2,
		Devices: map[string]*DeviceCollectionInfo{
			"device1": device1,
			"device2": device2,
		},
		CollectionTime: time.Now(),
		Duration:       200 * time.Millisecond,
	}

	if len(result.Devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(result.Devices))
	}

	if result.Devices["device1"].LoadTotalWatt != 1000.0 {
		t.Error("Device1 power mismatch")
	}
	if result.Devices["device2"].EnergyValue != 1000.0 {
		t.Error("Device2 energy mismatch")
	}
}
