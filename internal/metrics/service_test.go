package metrics

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
	"github.com/lay-g/winpower-g2-exporter/internal/metrics/mocks"
)

func TestNewMetricsService(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		collector  collector.CollectorInterface
		logger     *zap.Logger
		config     *MetricsConfig
		wantErr    error
		validateFn func(*testing.T, *MetricsService)
	}{
		{
			name:      "successful creation with defaults",
			collector: mocks.NewMockCollector(),
			logger:    logger,
			config:    nil,
			wantErr:   nil,
			validateFn: func(t *testing.T, m *MetricsService) {
				assert.NotNil(t, m)
				assert.NotNil(t, m.registry)
				assert.NotNil(t, m.collector)
				assert.NotNil(t, m.logger)
				assert.NotNil(t, m.deviceMetrics)
			},
		},
		{
			name:      "successful creation with custom config",
			collector: mocks.NewMockCollector(),
			logger:    logger,
			config: &MetricsConfig{
				Namespace:           "custom",
				Subsystem:           "test",
				WinPowerHost:        "test-host",
				EnableMemoryMetrics: false,
			},
			wantErr: nil,
			validateFn: func(t *testing.T, m *MetricsService) {
				assert.NotNil(t, m)
				assert.Nil(t, m.memoryBytes) // Memory metrics disabled
			},
		},
		{
			name:      "nil collector returns error",
			collector: nil,
			logger:    logger,
			config:    nil,
			wantErr:   ErrCollectorNil,
		},
		{
			name:      "nil logger returns error",
			collector: mocks.NewMockCollector(),
			logger:    nil,
			config:    nil,
			wantErr:   ErrLoggerNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewMetricsService(tt.collector, tt.logger, tt.config)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, service)
			} else {
				require.NoError(t, err)
				require.NotNil(t, service)
				if tt.validateFn != nil {
					tt.validateFn(t, service)
				}
			}
		})
	}
}

func TestMetricsService_updateMetrics(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollector()
	service, err := NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name    string
		result  *collector.CollectionResult
		wantErr bool
	}{
		{
			name: "successful update with devices",
			result: &collector.CollectionResult{
				Success:        true,
				DeviceCount:    1,
				CollectionTime: time.Now(),
				Duration:       100 * time.Millisecond,
				Devices: map[string]*collector.DeviceCollectionInfo{
					"test-device": {
						DeviceID:       "test-device",
						DeviceName:     "Test UPS",
						DeviceType:     1,
						Connected:      true,
						LastUpdateTime: time.Now(),
						LoadTotalWatt:  500.0,
						LoadPercent:    50.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "successful update with no devices",
			result: &collector.CollectionResult{
				Success:        true,
				DeviceCount:    0,
				CollectionTime: time.Now(),
				Duration:       50 * time.Millisecond,
				Devices:        make(map[string]*collector.DeviceCollectionInfo),
			},
			wantErr: false,
		},
		{
			name:    "nil result returns error",
			result:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := service.updateMetrics(tt.result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMetricsService_updateDeviceMetrics(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollector()
	service, err := NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	deviceID := "test-device-001"
	deviceInfo := &collector.DeviceCollectionInfo{
		DeviceID:          deviceID,
		DeviceName:        "Test UPS",
		DeviceType:        1,
		Connected:         true,
		LastUpdateTime:    time.Now(),
		InputVolt1:        220.0,
		InputFreq:         50.0,
		OutputVolt1:       220.0,
		OutputCurrent1:    5.0,
		OutputFreq:        50.0,
		OutputVoltageType: "single",
		LoadPercent:       45.0,
		LoadTotalWatt:     1000.0,
		LoadTotalVa:       1100.0,
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
	}

	// First update - should create device metrics
	err = service.updateDeviceMetrics(deviceID, deviceInfo)
	assert.NoError(t, err)
	assert.Contains(t, service.deviceMetrics, deviceID)

	// Second update - should update existing metrics
	deviceInfo.LoadTotalWatt = 1200.0
	err = service.updateDeviceMetrics(deviceID, deviceInfo)
	assert.NoError(t, err)

	// Test with nil device info
	err = service.updateDeviceMetrics("nil-device", nil)
	assert.Error(t, err)
}

func TestMetricsService_handleCollectionError(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollector()
	service, err := NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "timeout error",
			err:  context.DeadlineExceeded,
		},
		{
			name: "cancelled error",
			err:  context.Canceled,
		},
		{
			name: "generic error",
			err:  errors.New("collection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service.handleCollectionError(tt.err)
			// Verify that error counter was incremented
			// Note: We can't easily verify counter values without using testutil
			// but we can verify the function doesn't panic
		})
	}
}

func TestMetricsService_updateSelfMetrics(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollector()
	service, err := NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	result := &collector.CollectionResult{
		Success:        true,
		DeviceCount:    2,
		CollectionTime: time.Now(),
		Duration:       150 * time.Millisecond,
		Devices:        make(map[string]*collector.DeviceCollectionInfo),
	}

	service.updateSelfMetrics(result)
	// Verify that the function doesn't panic
}

func TestMetricsService_updateMemoryMetrics(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollector()
	config := DefaultMetricsConfig()
	config.EnableMemoryMetrics = true

	service, err := NewMetricsService(mockCollector, logger, config)
	require.NoError(t, err)

	service.updateMemoryMetrics()
	// Verify that the function doesn't panic
}

func TestEncodingFunctions(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(string) float64
		input    string
		expected float64
	}{
		// Output voltage type encoding
		{"voltage type single", encodeOutputVoltageType, "single", 1},
		{"voltage type three", encodeOutputVoltageType, "three", 3},
		{"voltage type unknown", encodeOutputVoltageType, "unknown", 0},

		// Battery status encoding
		{"battery normal", encodeBatteryStatus, "normal", 1},
		{"battery low", encodeBatteryStatus, "low", 2},
		{"battery depleted", encodeBatteryStatus, "depleted", 3},
		{"battery unknown", encodeBatteryStatus, "unknown", 0},

		// UPS mode encoding
		{"ups mode online", encodeUPSMode, "online", 1},
		{"ups mode battery", encodeUPSMode, "battery", 2},
		{"ups mode bypass", encodeUPSMode, "bypass", 3},
		{"ups mode unknown", encodeUPSMode, "unknown", 0},

		// UPS status encoding
		{"ups status normal", encodeUPSStatus, "normal", 1},
		{"ups status warning", encodeUPSStatus, "warning", 2},
		{"ups status alarm", encodeUPSStatus, "alarm", 3},
		{"ups status unknown", encodeUPSStatus, "unknown", 0},

		// Test status encoding
		{"test status no_test", encodeTestStatus, "no_test", 0},
		{"test status testing", encodeTestStatus, "testing", 1},
		{"test status passed", encodeTestStatus, "passed", 2},
		{"test status failed", encodeTestStatus, "failed", 3},
		{"test status unknown", encodeTestStatus, "unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultMetricsConfig(t *testing.T) {
	config := DefaultMetricsConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "winpower", config.Namespace)
	assert.Equal(t, "exporter", config.Subsystem)
	assert.Equal(t, "localhost", config.WinPowerHost)
	assert.True(t, config.EnableMemoryMetrics)
}

func TestZapLogger(t *testing.T) {
	logger := zap.NewNop()
	zapLog := &zapLogger{logger: logger}

	// Test that Println doesn't panic
	zapLog.Println("test message")
	zapLog.Println("test", "with", "multiple", "args")
}

func TestMetricsService_createDeviceMetrics(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollector()
	service, err := NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	dm := service.createDeviceMetrics("device-001", "Test Device", "1", "test-host")

	assert.NotNil(t, dm)
	assert.NotNil(t, dm.connected)
	assert.NotNil(t, dm.loadTotalWatt)
	assert.NotNil(t, dm.powerWatts)
	assert.NotNil(t, dm.cumulativeEnergy)
	assert.NotNil(t, dm.batteryCharging)
	assert.NotNil(t, dm.upsMode)
}

func TestMetricsService_concurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	mockCollector := mocks.NewMockCollectorWithDevices()
	service, err := NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	// Simulate concurrent access to updateMetrics
	ctx := context.Background()
	result, err := mockCollector.CollectDeviceData(ctx)
	require.NoError(t, err)

	// Run multiple updates concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			err := service.updateMetrics(result)
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
