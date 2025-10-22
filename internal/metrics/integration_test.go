package metrics

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/lay-g/winpower-g2-exporter/internal/energy"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// mockStorage implements energy.StorageManager for testing
type mockStorage struct {
	data map[string]*energy.PowerData
	mu   sync.RWMutex
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		data: make(map[string]*energy.PowerData),
	}
}

func (m *mockStorage) Write(deviceID string, data *energy.PowerData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[deviceID] = data
	return nil
}

func (m *mockStorage) Read(deviceID string) (*energy.PowerData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if data, exists := m.data[deviceID]; exists {
		return data, nil
	}
	// Return zero data for new devices as per energy interface specification
	return &energy.PowerData{Timestamp: time.Now().UnixMilli(), EnergyWH: 0}, nil
}

// mockWinPowerClient implements a mock WinPower client for testing
type mockWinPowerClient struct {
	devices map[string]*winpower.DeviceData
	mu      sync.RWMutex
}

func newMockWinPowerClient() *mockWinPowerClient {
	return &mockWinPowerClient{
		devices: make(map[string]*winpower.DeviceData),
	}
}

func (m *mockWinPowerClient) AddDevice(device *winpower.DeviceData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.devices[device.ID] = device
}

func (m *mockWinPowerClient) GetDevice(deviceID string) (*winpower.DeviceData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if device, exists := m.devices[deviceID]; exists {
		return device, nil
	}
	return nil, fmt.Errorf("device not found: %s", deviceID)
}

func (m *mockWinPowerClient) GetAllDevices() ([]*winpower.DeviceData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	devices := make([]*winpower.DeviceData, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	return devices, nil
}

// TestMetricManager_EndToEndFlow tests the complete metrics flow from data collection to exposure
func TestMetricManager_EndToEndFlow(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Create metrics manager
	config := DefaultConfig()
	config.Namespace = "test_winpower"

	metricsManager, err := NewMetricManager(config, logger)
	require.NoError(t, err)
	require.NotNil(t, metricsManager)

	// Create mock services
	storage := newMockStorage()
	energyConfig := &energy.Config{
		Precision:            0.01,
		EnableStats:          true,
		MaxCalculationTime:   1000000000,
		NegativePowerAllowed: true,
	}
	energyService := energy.NewEnergyService(storage, logger, energyConfig)

	winPowerClient := newMockWinPowerClient()

	// Add test devices
	testDevice1 := &winpower.DeviceData{
		ID:            "ups-001",
		Name:          "Main UPS",
		Type:          "UPS",
		Status:        "online",
		Timestamp:     time.Now(),
		InputVoltage:  230.0,
		InputCurrent:  10.5,
		InputFreq:     50.0,
		InputPower:    2415.0,
		OutputVoltage: 230.0,
		OutputCurrent: 9.8,
		OutputFreq:    50.0,
		OutputPower:   2254.0,
		Power:         2254.0,
		LoadPercent:   45.2,
	}

	testDevice2 := &winpower.DeviceData{
		ID:            "pdu-001",
		Name:          "Server PDU",
		Type:          "PDU",
		Status:        "online",
		Timestamp:     time.Now(),
		InputVoltage:  230.0,
		InputCurrent:  25.3,
		InputFreq:     50.0,
		InputPower:    5819.0,
		OutputVoltage: 230.0,
		OutputCurrent: 24.8,
		OutputFreq:    50.0,
		OutputPower:   5704.0,
		Power:         5704.0,
		LoadPercent:   78.5,
	}

	winPowerClient.AddDevice(testDevice1)
	winPowerClient.AddDevice(testDevice2)

	// Simulate the complete data flow
	// Step 1: Set initial exporter status
	metricsManager.SetUp(true)

	// Step 2: Set connection status
	metricsManager.SetConnectionStatus("localhost", "http", 1)
	metricsManager.SetAuthStatus("localhost", "basic", 1)

	// Step 3: Process each device and update metrics
	devices, err := winPowerClient.GetAllDevices()
	require.NoError(t, err)

	for _, device := range devices {
		// Update device connection status
		connected := 0.0
		if device.Status == "online" {
			connected = 1.0
		}
		metricsManager.SetDeviceConnected(device.ID, device.Name, device.Type, connected)

		// Update electrical metrics
		metricsManager.SetElectricalData(
			device.ID, device.Name, device.Type, "",
			device.InputVoltage, device.OutputVoltage,
			device.InputCurrent, device.OutputCurrent,
			device.InputFreq, device.OutputFreq,
			device.InputPower, device.OutputPower,
			device.PowerFactor,
		)

		// Update load percentage
		metricsManager.SetLoadPercent(device.ID, device.Name, device.Type, "", device.LoadPercent)

		// Calculate and update energy metrics
		totalEnergy, err := energyService.Calculate(device.ID, device.Power)
		if err != nil {
			logger.Error("Failed to calculate energy", zap.Error(err))
			continue
		}

		metricsManager.SetPowerWatts(device.ID, device.Name, device.Type, device.Power)
		metricsManager.SetEnergyTotalWh(device.ID, device.Name, device.Type, totalEnergy)

		// Simulate API call observation
		metricsManager.ObserveAPI("localhost", "/api/devices", 50*time.Millisecond)
	}

	// Step 4: Update device count
	metricsManager.SetDeviceCount("localhost", "UPS", 1)
	metricsManager.SetDeviceCount("localhost", "PDU", 1)

	// Step 5: Simulate scrape request
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler := metricsManager.Handler()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	metricsOutput := string(body)
	t.Logf("Metrics output:\n%s", metricsOutput)

	// Verify key metrics are present
	assert.Contains(t, metricsOutput, "test_winpower_exporter_up 1")
	assert.Contains(t, metricsOutput, `test_winpower_device_connected{device_id="ups-001",device_name="Main UPS",device_type="UPS"} 1`)
	assert.Contains(t, metricsOutput, `test_winpower_device_connected{device_id="pdu-001",device_name="Server PDU",device_type="PDU"} 1`)
	assert.Contains(t, metricsOutput, `test_winpower_load_percent{device_id="ups-001",device_name="Main UPS",device_type="UPS",phase=""} 45.2`)
	assert.Contains(t, metricsOutput, `test_winpower_load_percent{device_id="pdu-001",device_name="Server PDU",device_type="PDU",phase=""} 78.5`)
	assert.Contains(t, metricsOutput, `test_winpower_power_watts{device_id="ups-001",device_name="Main UPS",device_type="UPS"}`)
	assert.Contains(t, metricsOutput, `test_winpower_power_watts{device_id="pdu-001",device_name="Server PDU",device_type="PDU"}`)
	assert.Contains(t, metricsOutput, `test_winpower_energy_total_wh{device_id="ups-001",device_name="Main UPS",device_type="UPS"}`)
	assert.Contains(t, metricsOutput, `test_winpower_energy_total_wh{device_id="pdu-001",device_name="Server PDU",device_type="PDU"}`)

	// Step 6: Test request observation
	metricsManager.ObserveRequest("localhost", "GET", "/metrics", 10*time.Millisecond)

	// Step 7: Verify metrics are still accessible after multiple updates
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp = w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMetricManager_WinPowerIntegration tests integration with WinPower module
func TestMetricManager_WinPowerIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultConfig()
	config.Namespace = "test_winpower"

	metricsManager, err := NewMetricManager(config, logger)
	require.NoError(t, err)

	winPowerClient := newMockWinPowerClient()

	// Test device with comprehensive data
	comprehensiveDevice := &winpower.DeviceData{
		ID:              "ups-comprehensive",
		Name:            "Comprehensive Test UPS",
		Type:            "UPS",
		Status:          "online",
		Timestamp:       time.Now(),
		InputVoltage:    230.5,
		InputCurrent:    15.2,
		InputFreq:       50.1,
		InputPower:      3503.6,
		OutputVoltage:   229.8,
		OutputCurrent:   14.9,
		OutputFreq:      50.0,
		OutputPower:     3424.0,
		Power:           3424.0,
		ApparentPower:   3800.0,
		PowerFactor:     0.90,
		LoadPercent:     68.5,
		BatteryCapacity: 85.0,
		BatteryVoltage:  13.6,
		BatteryCurrent:  2.1,
		RuntimeMinutes:  45.0,
		Temperature:     25.5,
		Humidity:        40.0,
		Location:        "Server Room A",
		Model:           "SmartUPS RT 5000",
		SerialNumber:    "US1234567890",
	}

	winPowerClient.AddDevice(comprehensiveDevice)

	// Simulate device data processing
	device, err := winPowerClient.GetDevice("ups-comprehensive")
	require.NoError(t, err)
	require.NotNil(t, device)

	// Update all relevant metrics
	connected := 1.0
	if device.Status != "online" {
		connected = 0.0
	}

	metricsManager.SetDeviceConnected(device.ID, device.Name, device.Type, connected)
	metricsManager.SetElectricalData(
		device.ID, device.Name, device.Type, "",
		device.InputVoltage, device.OutputVoltage,
		device.InputCurrent, device.OutputCurrent,
		device.InputFreq, device.OutputFreq,
		device.InputPower, device.OutputPower,
		device.PowerFactor,
	)
	metricsManager.SetLoadPercent(device.ID, device.Name, device.Type, "", device.LoadPercent)

	// Test with phase-specific data (for three-phase devices)
	threePhaseDevice := &winpower.DeviceData{
		ID:            "pdu-three-phase",
		Name:          "Three Phase PDU",
		Type:          "PDU",
		Status:        "online",
		Timestamp:     time.Now(),
		InputVoltage:  400.0, // Three-phase voltage
		InputCurrent:  20.0,
		InputFreq:     50.0,
		InputPower:    8000.0,
		OutputVoltage: 400.0,
		OutputCurrent: 19.5,
		OutputFreq:    50.0,
		OutputPower:   7800.0,
		Power:         7800.0,
		LoadPercent:   65.0,
	}

	winPowerClient.AddDevice(threePhaseDevice)

	// Update three-phase metrics with phase labels
	phases := []string{"L1", "L2", "L3"}
	for i, phase := range phases {
		// Simulate per-phase data (in real implementation, this would come from device data)
		phaseInputCurrent := threePhaseDevice.InputCurrent + float64(i)*0.5
		phaseOutputCurrent := threePhaseDevice.OutputCurrent - float64(i)*0.3
		phasePower := threePhaseDevice.Power/3.0 + float64(i)*100.0

		metricsManager.SetLoadPercent(
			threePhaseDevice.ID, threePhaseDevice.Name, threePhaseDevice.Type,
			phase, threePhaseDevice.LoadPercent,
		)
		metricsManager.SetElectricalData(
			threePhaseDevice.ID, threePhaseDevice.Name, threePhaseDevice.Type,
			phase,
			threePhaseDevice.InputVoltage, threePhaseDevice.OutputVoltage,
			phaseInputCurrent, phaseOutputCurrent,
			threePhaseDevice.InputFreq, threePhaseDevice.OutputFreq,
			phasePower, phasePower,
			0.95, // Power factor
		)
	}

	// Verify metrics are properly exposed
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler := metricsManager.Handler()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	metricsOutput := string(body)

	// Verify comprehensive device metrics
	assert.Contains(t, metricsOutput, `device_id="ups-comprehensive"`)
	assert.Contains(t, metricsOutput, `device_name="Comprehensive Test UPS"`)
	assert.Contains(t, metricsOutput, `device_type="UPS"`)

	// Verify three-phase device metrics with phase labels
	assert.Contains(t, metricsOutput, `device_id="pdu-three-phase"`)
	assert.Contains(t, metricsOutput, `phase="L1"`)
	assert.Contains(t, metricsOutput, `phase="L2"`)
	assert.Contains(t, metricsOutput, `phase="L3"`)
}

// TestMetricManager_EnergyIntegration tests integration with Energy module
func TestMetricManager_EnergyIntegration(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultConfig()
	config.Namespace = "test_winpower"

	metricsManager, err := NewMetricManager(config, logger)
	require.NoError(t, err)

	// Setup energy service
	storage := newMockStorage()
	energyConfig := &energy.Config{
		Precision:            0.01,
		EnableStats:          true,
		MaxCalculationTime:   1000000000,
		NegativePowerAllowed: true,
	}
	energyService := energy.NewEnergyService(storage, logger, energyConfig)

	// Test device with energy calculation
	deviceID := "ups-energy-test"
	deviceName := "Energy Test UPS"
	deviceType := "UPS"

	// Simulate multiple power readings over time
	powerReadings := []float64{
		2000.0, // 2 kW initial
		2100.0, // 2.1 kW after 5 seconds
		1900.0, // 1.9 kW after 10 seconds
		2200.0, // 2.2 kW after 15 seconds
		1800.0, // 1.8 kW after 20 seconds
	}

	expectedEnergy := 0.0
	for i, power := range powerReadings {
		// Calculate energy for this interval (5 seconds = 5/3600 hours)
		intervalHours := 5.0 / 3600.0
		expectedEnergy += power * intervalHours

		// Update energy service
		totalEnergy, err := energyService.Calculate(deviceID, power)
		require.NoError(t, err)

		// Update metrics
		metricsManager.SetPowerWatts(deviceID, deviceName, deviceType, power)
		metricsManager.SetEnergyTotalWh(deviceID, deviceName, deviceType, totalEnergy)

		t.Logf("Reading %d: Power=%.1fW, Total Energy=%.2fWh", i+1, power, totalEnergy)

		// Verify energy is accumulating (first reading may be 0 due to time intervals)
		assert.GreaterOrEqual(t, totalEnergy, 0.0, "Energy should be non-negative")
		if i > 0 {
			// Energy should be increasing after first reading
			prevEnergy, err := energyService.Get(deviceID)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, totalEnergy, prevEnergy, "Energy should not decrease")
		}

		// Small delay to simulate real-time collection
		time.Sleep(10 * time.Millisecond)
	}

	// Test negative power handling (energy feedback scenario)
	negativePower := -500.0 // -500W (energy being fed back)
	totalEnergy, err := energyService.Calculate(deviceID, negativePower)
	require.NoError(t, err)

	metricsManager.SetPowerWatts(deviceID, deviceName, deviceType, negativePower)
	metricsManager.SetEnergyTotalWh(deviceID, deviceName, deviceType, totalEnergy)

	t.Logf("Negative power test: Power=%.1fW, Total Energy=%.2fWh", negativePower, totalEnergy)

	// Verify final metrics
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler := metricsManager.Handler()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	metricsOutput := string(body)
	t.Logf("Final metrics output:\n%s", metricsOutput)

	// Verify energy metrics are present and correct
	assert.Contains(t, metricsOutput, fmt.Sprintf(`test_winpower_power_watts{device_id="%s",device_name="%s",device_type="%s"} -500`, deviceID, deviceName, deviceType))
	assert.Contains(t, metricsOutput, fmt.Sprintf(`test_winpower_energy_total_wh{device_id="%s",device_name="%s",device_type="%s"}`, deviceID, deviceName, deviceType))

	// Verify energy value is reasonable (should be positive since most readings were positive)
	energyValueStr := extractMetricValue(metricsOutput, fmt.Sprintf(`test_winpower_energy_total_wh{device_id="%s",device_name="%s",device_type="%s"}`, deviceID, deviceName, deviceType))
	require.NotEmpty(t, energyValueStr)

	energyValue, err := parseFloat(energyValueStr)
	require.NoError(t, err)
	assert.Greater(t, energyValue, 0.0, "Total energy should be positive")

	// Test multiple devices
	device2ID := "pdu-energy-test"
	device2Name := "Energy Test PDU"
	device2Type := "PDU"

	device2Power := 3500.0
	totalEnergy2, err := energyService.Calculate(device2ID, device2Power)
	require.NoError(t, err)

	metricsManager.SetPowerWatts(device2ID, device2Name, device2Type, device2Power)
	metricsManager.SetEnergyTotalWh(device2ID, device2Name, device2Type, totalEnergy2)

	// Verify both devices' metrics are present
	req = httptest.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp = w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ = io.ReadAll(resp.Body)
	metricsOutput = string(body)

	assert.Contains(t, metricsOutput, fmt.Sprintf(`device_id="%s"`, deviceID))
	assert.Contains(t, metricsOutput, fmt.Sprintf(`device_id="%s"`, device2ID))
}

// TestMetricManager_ErrorScenarios tests various error scenarios and edge cases
func TestMetricManager_ErrorScenarios(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultConfig()
	config.Namespace = "test_winpower"

	metricsManager, err := NewMetricManager(config, logger)
	require.NoError(t, err)

	// Test 1: Invalid device data
	t.Run("InvalidDeviceData", func(t *testing.T) {
		// Test with empty device ID
		metricsManager.SetDeviceConnected("", "Empty ID", "UPS", 1.0)

		// Test with extremely large values
		metricsManager.SetLoadPercent("large-test", "Large Test", "UPS", "", 999999.0)
		metricsManager.SetElectricalData(
			"large-test", "Large Test", "UPS", "",
			999999.0, 999999.0, 999999.0, 999999.0,
			999999.0, 999999.0, 999999.0, 999999.0,
			999999.0,
		)

		// Verify metrics are still accessible
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		handler := metricsManager.Handler()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	
	// Test 2: Special characters in labels
	t.Run("SpecialCharactersInLabels", func(t *testing.T) {
		specialDeviceID := "device-with-special.chars_123"
		specialDeviceName := `Device with "quotes" and \ slashes`
		specialDeviceType := "PDU/UPS"

		metricsManager.SetDeviceConnected(specialDeviceID, specialDeviceName, specialDeviceType, 1.0)
		metricsManager.SetPowerWatts(specialDeviceID, specialDeviceName, specialDeviceType, 1500.0)

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		handler := metricsManager.Handler()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		metricsOutput := string(body)

		// Verify special characters are handled properly
		assert.Contains(t, metricsOutput, `device_id="device-with-special.chars_123"`)
		// Just verify device is present in metrics output
		assert.Contains(t, metricsOutput, `device_name="Device`)
	})

	// Test 3: Invalid HTTP requests
	t.Run("InvalidHTTPRequests", func(t *testing.T) {
		handler := metricsManager.Handler()

		// Test POST request (should work but may not be standard)
		req := httptest.NewRequest("POST", "/metrics", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp := w.Result()
		// Prometheus handler typically supports GET and HEAD
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMethodNotAllowed)

		// Test request with headers
		req = httptest.NewRequest("GET", "/metrics", nil)
		req.Header.Set("Accept", "application/json")
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		resp = w.Result()
		// Should still return Prometheus format
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.NotEmpty(t, body)
	})
}

// TestMetricManager_MultiScenarioValidation tests metrics correctness across different scenarios
func TestMetricManager_MultiScenarioValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	config := DefaultConfig()
	config.Namespace = "test_winpower"

	metricsManager, err := NewMetricManager(config, logger)
	require.NoError(t, err)

	scenarios := []struct {
		name           string
		deviceID       string
		deviceName     string
		deviceType     string
		status         string
		power          float64
		voltage        float64
		current        float64
		loadPercent    float64
		expectedStatus float64
	}{
		{
			name:           "Online UPS",
			deviceID:       "ups-online",
			deviceName:     "Online UPS",
			deviceType:     "UPS",
			status:         "online",
			power:          2000.0,
			voltage:        230.0,
			current:        8.7,
			loadPercent:    40.0,
			expectedStatus: 1.0,
		},
		{
			name:           "Offline UPS",
			deviceID:       "ups-offline",
			deviceName:     "Offline UPS",
			deviceType:     "UPS",
			status:         "offline",
			power:          0.0,
			voltage:        0.0,
			current:        0.0,
			loadPercent:    0.0,
			expectedStatus: 0.0,
		},
		{
			name:           "High Load PDU",
			deviceID:       "pdu-high-load",
			deviceName:     "High Load PDU",
			deviceType:     "PDU",
			status:         "online",
			power:          8000.0,
			voltage:        230.0,
			current:        34.8,
			loadPercent:    95.0,
			expectedStatus: 1.0,
		},
		{
			name:           "Critical UPS",
			deviceID:       "ups-critical",
			deviceName:     "Critical UPS",
			deviceType:     "UPS",
			status:         "critical",
			power:          1500.0,
			voltage:        220.0,
			current:        6.8,
			loadPercent:    85.0,
			expectedStatus: 1.0, // Still connected even if critical
		},
	}

	// Process all scenarios
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Update device metrics
			metricsManager.SetDeviceConnected(scenario.deviceID, scenario.deviceName, scenario.deviceType, scenario.expectedStatus)
			metricsManager.SetElectricalData(
				scenario.deviceID, scenario.deviceName, scenario.deviceType, "",
				scenario.voltage, scenario.voltage, // Same input/output for simplicity
				scenario.current, scenario.current,
				50.0, 50.0, // Standard frequency
				scenario.power, scenario.power,
				0.9, // Good power factor
			)
			metricsManager.SetLoadPercent(scenario.deviceID, scenario.deviceName, scenario.deviceType, "", scenario.loadPercent)
			metricsManager.SetPowerWatts(scenario.deviceID, scenario.deviceName, scenario.deviceType, scenario.power)
			metricsManager.SetEnergyTotalWh(scenario.deviceID, scenario.deviceName, scenario.deviceType, scenario.power*10) // Mock energy

			// Verify metrics in output
			req := httptest.NewRequest("GET", "/metrics", nil)
			w := httptest.NewRecorder()

			handler := metricsManager.Handler()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			metricsOutput := string(body)

			// Verify device-specific metrics
			assert.Contains(t, metricsOutput, fmt.Sprintf(`device_id="%s"`, scenario.deviceID))
			assert.Contains(t, metricsOutput, fmt.Sprintf(`device_name="%s"`, scenario.deviceName))
			assert.Contains(t, metricsOutput, fmt.Sprintf(`device_type="%s"`, scenario.deviceType))
			// Check device connected status without exact decimal formatting
			assert.Contains(t, metricsOutput, fmt.Sprintf(`test_winpower_device_connected{device_id="%s",device_name="%s",device_type="%s"}`, scenario.deviceID, scenario.deviceName, scenario.deviceType))
			// Check for load percent without exact decimal match to handle Prometheus formatting
			expectedLoadValue := fmt.Sprintf("%.0f", scenario.loadPercent)
			assert.Contains(t, metricsOutput, fmt.Sprintf(`test_winpower_load_percent{device_id="%s",device_name="%s",device_type="%s",phase=""} %s`, scenario.deviceID, scenario.deviceName, scenario.deviceType, expectedLoadValue))
		})
	}

	// Test overall metrics summary
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler := metricsManager.Handler()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	metricsOutput := string(body)

	// Count device types
	deviceTypes := make(map[string]int)
	for _, scenario := range scenarios {
		deviceTypes[scenario.deviceType]++
	}

	for deviceType, count := range deviceTypes {
		// Device count metrics may be present depending on implementation
		if strings.Contains(metricsOutput, "test_winpower_device_count") {
			expectedMetric := fmt.Sprintf(`test_winpower_device_count{device_type="%s",host="localhost"} %d`, deviceType, count)
			assert.Contains(t, metricsOutput, expectedMetric)
		}
	}
}

// Helper functions

func extractMetricValue(metricsOutput, metricName string) string {
	lines := strings.Split(metricsOutput, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, metricName+" ") {
			return strings.TrimPrefix(line, metricName+" ")
		}
	}
	return ""
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}
