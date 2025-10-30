package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

func TestNewCollectorService(t *testing.T) {
	logger := log.NewTestLogger()

	tests := []struct {
		name           string
		winpowerClient WinPowerClient
		energyCalc     EnergyCalculator
		logger         log.Logger
		wantErr        bool
		errContains    string
	}{
		{
			name:           "Valid dependencies",
			winpowerClient: &MockWinPowerClient{},
			energyCalc:     &MockEnergyCalculator{},
			logger:         logger,
			wantErr:        false,
		},
		{
			name:           "Nil WinPower client",
			winpowerClient: nil,
			energyCalc:     &MockEnergyCalculator{},
			logger:         logger,
			wantErr:        true,
			errContains:    "winpowerClient",
		},
		{
			name:           "Nil Energy calculator",
			winpowerClient: &MockWinPowerClient{},
			energyCalc:     nil,
			logger:         logger,
			wantErr:        true,
			errContains:    "energyCalc",
		},
		{
			name:           "Nil logger",
			winpowerClient: &MockWinPowerClient{},
			energyCalc:     &MockEnergyCalculator{},
			logger:         nil,
			wantErr:        true,
			errContains:    "logger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewCollectorService(tt.winpowerClient, tt.energyCalc, tt.logger)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				if service != nil {
					t.Error("Expected nil service when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if service == nil {
					t.Error("Expected non-nil service")
				}
			}
		})
	}
}

func TestCollectorService_CollectDeviceData_Success(t *testing.T) {
	logger := log.NewTestLogger()

	// Mock WinPower client
	mockWinPower := &MockWinPowerClient{
		CollectDeviceDataFunc: func(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
			return []winpower.ParsedDeviceData{
				{
					DeviceID:   "device1",
					DeviceType: 1,
					Model:      "G2-3000",
					Alias:      "UPS-001",
					Connected:  true,
					Realtime: winpower.RealtimeData{
						LoadTotalWatt:  1500.0,
						InputVolt1:     220.5,
						OutputVolt1:    220.0,
						LoadPercent:    65.5,
						BatCapacity:    100.0,
						UpsTemperature: 35.5,
						Mode:           "normal",
						Status:         "online",
					},
					CollectedAt: time.Now(),
				},
			}, nil
		},
	}

	// Mock energy calculator
	mockEnergy := &MockEnergyCalculator{
		CalculateFunc: func(deviceID string, power float64) (float64, error) {
			return 1250.5, nil // Return calculated energy
		},
	}

	service, err := NewCollectorService(mockWinPower, mockEnergy, logger)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	result, err := service.CollectDeviceData(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Success {
		t.Error("Expected success to be true")
	}
	if result.DeviceCount != 1 {
		t.Errorf("Expected device count to be 1, got %d", result.DeviceCount)
	}
	if len(result.Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(result.Devices))
	}

	device, exists := result.Devices["device1"]
	if !exists {
		t.Fatal("Expected device1 to exist in results")
	}
	if device.DeviceID != "device1" {
		t.Errorf("Expected device ID to be 'device1', got %s", device.DeviceID)
	}
	if device.LoadTotalWatt != 1500.0 {
		t.Errorf("Expected power to be 1500.0, got %f", device.LoadTotalWatt)
	}
	if !device.EnergyCalculated {
		t.Error("Expected energy to be calculated")
	}
	if device.EnergyValue != 1250.5 {
		t.Errorf("Expected energy value to be 1250.5, got %f", device.EnergyValue)
	}
}

func TestCollectorService_CollectDeviceData_WinPowerError(t *testing.T) {
	logger := log.NewTestLogger()

	mockWinPower := &MockWinPowerClient{
		CollectDeviceDataFunc: func(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
			return nil, errors.New("connection timeout")
		},
	}

	mockEnergy := &MockEnergyCalculator{}

	service, err := NewCollectorService(mockWinPower, mockEnergy, logger)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	result, err := service.CollectDeviceData(ctx)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if result == nil {
		t.Fatal("Expected non-nil result even on error")
	}
	if result.Success {
		t.Error("Expected success to be false")
	}
	if result.DeviceCount != 0 {
		t.Errorf("Expected device count to be 0, got %d", result.DeviceCount)
	}
}

func TestCollectorService_CollectDeviceData_EnergyCalculationError(t *testing.T) {
	logger := log.NewTestLogger()

	mockWinPower := &MockWinPowerClient{
		CollectDeviceDataFunc: func(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
			return []winpower.ParsedDeviceData{
				{
					DeviceID:   "device1",
					DeviceType: 1,
					Model:      "G2-3000",
					Connected:  true,
					Realtime: winpower.RealtimeData{
						LoadTotalWatt: 1500.0,
					},
					CollectedAt: time.Now(),
				},
			}, nil
		},
	}

	mockEnergy := &MockEnergyCalculator{
		CalculateFunc: func(deviceID string, power float64) (float64, error) {
			return 0, errors.New("calculation failed")
		},
	}

	service, err := NewCollectorService(mockWinPower, mockEnergy, logger)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	result, err := service.CollectDeviceData(ctx)

	// Should not return error at top level - energy calculation errors are per-device
	if err != nil {
		t.Errorf("Expected no top-level error, got %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if !result.Success {
		t.Error("Expected overall success to be true")
	}

	device, exists := result.Devices["device1"]
	if !exists {
		t.Fatal("Expected device1 to exist in results")
	}
	if device.EnergyCalculated {
		t.Error("Expected energy calculation to fail")
	}
	if device.ErrorMsg == "" {
		t.Error("Expected error message to be set")
	}
}

func TestCollectorService_CollectDeviceData_MultipleDevices(t *testing.T) {
	logger := log.NewTestLogger()

	mockWinPower := &MockWinPowerClient{
		CollectDeviceDataFunc: func(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
			return []winpower.ParsedDeviceData{
				{
					DeviceID:  "device1",
					Connected: true,
					Realtime:  winpower.RealtimeData{LoadTotalWatt: 1000.0},
				},
				{
					DeviceID:  "device2",
					Connected: true,
					Realtime:  winpower.RealtimeData{LoadTotalWatt: 2000.0},
				},
				{
					DeviceID:  "device3",
					Connected: false,
					Realtime:  winpower.RealtimeData{LoadTotalWatt: 0},
				},
			}, nil
		},
	}

	mockEnergy := &MockEnergyCalculator{
		CalculateFunc: func(deviceID string, power float64) (float64, error) {
			// Simulate calculation error for device2
			if deviceID == "device2" {
				return 0, errors.New("device2 error")
			}
			return power * 0.5, nil
		},
	}

	service, err := NewCollectorService(mockWinPower, mockEnergy, logger)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx := context.Background()
	result, err := service.CollectDeviceData(ctx)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result.DeviceCount != 3 {
		t.Errorf("Expected 3 devices, got %d", result.DeviceCount)
	}

	// Check device1 (success)
	device1 := result.Devices["device1"]
	if !device1.EnergyCalculated {
		t.Error("Expected device1 energy to be calculated")
	}

	// Check device2 (calculation error)
	device2 := result.Devices["device2"]
	if device2.EnergyCalculated {
		t.Error("Expected device2 energy calculation to fail")
	}
	if device2.ErrorMsg == "" {
		t.Error("Expected device2 to have error message")
	}

	// Check device3 (disconnected)
	device3 := result.Devices["device3"]
	if device3.Connected {
		t.Error("Expected device3 to be disconnected")
	}
}

func TestCollectorService_CollectDeviceData_NilContext(t *testing.T) {
	logger := log.NewTestLogger()
	mockWinPower := &MockWinPowerClient{}
	mockEnergy := &MockEnergyCalculator{}

	service, err := NewCollectorService(mockWinPower, mockEnergy, logger)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	result, err := service.CollectDeviceData(nil)

	if err == nil {
		t.Error("Expected error for nil context")
	}
	if !errors.Is(err, ErrInvalidContext) {
		t.Errorf("Expected ErrInvalidContext, got %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nil context")
	}
}

func TestCollectorService_ConvertToDeviceInfo(t *testing.T) {
	logger := log.NewTestLogger()
	service, _ := NewCollectorService(&MockWinPowerClient{}, &MockEnergyCalculator{}, logger)

	now := time.Now()
	device := winpower.ParsedDeviceData{
		DeviceID:   "test-device",
		DeviceType: 1,
		Model:      "G2-3000",
		Alias:      "Test UPS",
		Connected:  true,
		Realtime: winpower.RealtimeData{
			LoadTotalWatt:  1500.0,
			InputVolt1:     220.5,
			OutputVolt1:    220.0,
			OutputCurrent1: 5.5,
			InputFreq:      50.0,
			OutputFreq:     50.0,
			LoadPercent:    65.5,
			LoadTotalVa:    1800.0,
			LoadWatt1:      1500.0,
			LoadVa1:        1800.0,
			IsCharging:     false,
			BatVoltP:       95.5,
			BatCapacity:    100.0,
			BatRemainTime:  3600,
			BatteryStatus:  "normal",
			UpsTemperature: 35.5,
			Mode:           "normal",
			Status:         "online",
			TestStatus:     "passed",
			FaultCode:      "",
		},
		CollectedAt: now,
	}

	info := service.convertToDeviceInfo(device)

	// Basic info
	if info.DeviceID != "test-device" {
		t.Errorf("Expected device ID 'test-device', got %s", info.DeviceID)
	}
	if info.DeviceName != "Test UPS" {
		t.Errorf("Expected device name 'Test UPS', got %s", info.DeviceName)
	}
	if info.DeviceModel != "G2-3000" {
		t.Errorf("Expected model 'G2-3000', got %s", info.DeviceModel)
	}
	if !info.Connected {
		t.Error("Expected connected to be true")
	}

	// Power data
	if info.LoadTotalWatt != 1500.0 {
		t.Errorf("Expected power 1500.0, got %f", info.LoadTotalWatt)
	}

	// Battery data
	if info.BatCapacity != 100.0 {
		t.Errorf("Expected battery capacity 100.0, got %f", info.BatCapacity)
	}

	// Initial energy state
	if info.EnergyCalculated {
		t.Error("Expected initial energy calculated to be false")
	}
	if info.EnergyValue != 0 {
		t.Errorf("Expected initial energy value to be 0, got %f", info.EnergyValue)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsRecursive(s, substr))
}

func containsRecursive(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsRecursive(s[1:], substr)
}
