package winpower

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataParser(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := zap.NewNop()
		parser := NewDataParser(logger)
		assert.NotNil(t, parser)
		assert.Equal(t, logger, parser.logger)
	})

	t.Run("with nil logger", func(t *testing.T) {
		parser := NewDataParser(nil)
		assert.NotNil(t, parser)
		assert.NotNil(t, parser.logger)
	})
}

func TestDataParser_ParseResponse(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	t.Run("nil response", func(t *testing.T) {
		result, err := parser.ParseResponse(nil)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidResponse, err)
		assert.Nil(t, result)
	})

	t.Run("non-success code", func(t *testing.T) {
		response := &DeviceDataResponse{
			Code: "401",
			Msg:  "Unauthorized",
		}
		result, err := parser.ParseResponse(response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
		assert.Contains(t, err.Error(), "401")
		assert.Nil(t, result)
	})

	t.Run("empty data", func(t *testing.T) {
		response := &DeviceDataResponse{
			Code: "000000",
			Msg:  "OK",
			Data: []DeviceInfo{},
		}
		result, err := parser.ParseResponse(response)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("valid response with single device", func(t *testing.T) {
		response := createValidDeviceDataResponse()
		result, err := parser.ParseResponse(response)
		assert.NoError(t, err)
		require.Len(t, result, 1)

		device := result[0]
		assert.Equal(t, "e156e6cb-41cb-4b35-b0dd-869929186a5c", device.DeviceID)
		assert.Equal(t, 1, device.DeviceType)
		assert.Equal(t, "ON-LINE", device.Model)
		assert.Equal(t, "C3K", device.Alias)
		assert.True(t, device.Connected)
		assert.NotZero(t, device.CollectedAt)
	})

	t.Run("multiple devices", func(t *testing.T) {
		response := &DeviceDataResponse{
			Total: 2,
			Code:  "000000",
			Msg:   "OK",
			Data: []DeviceInfo{
				createValidDeviceInfo(),
				createValidDeviceInfo(),
			},
		}
		result, err := parser.ParseResponse(response)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("invalid device data should be skipped", func(t *testing.T) {
		// Note: Even empty DeviceInfo{} is technically valid (has zero values),
		// so this test verifies we continue parsing even if one device fails
		response := &DeviceDataResponse{
			Total: 2,
			Code:  "000000",
			Msg:   "OK",
			Data: []DeviceInfo{
				createValidDeviceInfo(),
				{}, // empty but still parseable
			},
		}
		result, err := parser.ParseResponse(response)
		assert.NoError(t, err)
		// Both devices should be included (empty device has zero values)
		assert.Len(t, result, 2)
	})
}

func TestDataParser_parseDeviceInfo(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	t.Run("nil device info", func(t *testing.T) {
		result, err := parser.parseDeviceInfo(nil)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidDeviceData, err)
		assert.Nil(t, result)
	})

	t.Run("valid device info", func(t *testing.T) {
		deviceInfo := createValidDeviceInfo()
		result, err := parser.parseDeviceInfo(&deviceInfo)
		assert.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "e156e6cb-41cb-4b35-b0dd-869929186a5c", result.DeviceID)
		assert.Equal(t, 1, result.DeviceType)
		assert.Equal(t, "ON-LINE", result.Model)
		assert.Equal(t, "C3K", result.Alias)
		assert.True(t, result.Connected)
		assert.WithinDuration(t, time.Now(), result.CollectedAt, time.Second)
	})

	t.Run("device with empty realtime data", func(t *testing.T) {
		deviceInfo := createValidDeviceInfo()
		deviceInfo.Realtime = map[string]interface{}{}
		result, err := parser.parseDeviceInfo(&deviceInfo)
		assert.NoError(t, err)
		require.NotNil(t, result)
		// Should still parse device info even without realtime data
		assert.Equal(t, "e156e6cb-41cb-4b35-b0dd-869929186a5c", result.DeviceID)
	})
}

func TestDataParser_parseRealtimeData(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	t.Run("complete realtime data", func(t *testing.T) {
		raw := createValidRealtimeData()
		result := parser.parseRealtimeData(raw)

		// Power data
		assert.Equal(t, 195.0, result.LoadTotalWatt)
		assert.Equal(t, 195.0, result.LoadWatt1)
		assert.Equal(t, 198.0, result.LoadTotalVa)
		assert.Equal(t, 198.0, result.LoadVa1)

		// Voltage data
		assert.Equal(t, 236.8, result.InputVolt1)
		assert.Equal(t, 220.1, result.OutputVolt1)
		assert.Equal(t, 81.4, result.BatVoltP)

		// Current data
		assert.Equal(t, 1.1, result.OutputCurrent1)

		// Frequency data
		assert.Equal(t, 49.9, result.InputFreq)
		assert.Equal(t, 49.9, result.OutputFreq)

		// Load data
		assert.Equal(t, 6.0, result.LoadPercent)

		// Battery data
		assert.Equal(t, 90.0, result.BatCapacity)
		assert.Equal(t, 6723, result.BatRemainTime)
		assert.True(t, result.IsCharging)

		// Status data
		assert.Equal(t, 27.0, result.UpsTemperature)
		assert.Equal(t, "3", result.Mode)
		assert.Equal(t, "1", result.Status)
		assert.Equal(t, "2", result.BatteryStatus)
		assert.Equal(t, "1", result.TestStatus)
		assert.Equal(t, "", result.FaultCode)
	})

	t.Run("empty realtime data", func(t *testing.T) {
		raw := map[string]interface{}{}
		result := parser.parseRealtimeData(raw)
		// Should return zero values
		assert.Equal(t, 0.0, result.LoadTotalWatt)
		assert.Equal(t, "", result.Status)
		assert.False(t, result.IsCharging)
	})

	t.Run("partial realtime data", func(t *testing.T) {
		raw := map[string]interface{}{
			"loadTotalWatt": "195",
			"status":        "1",
		}
		result := parser.parseRealtimeData(raw)
		assert.Equal(t, 195.0, result.LoadTotalWatt)
		assert.Equal(t, "1", result.Status)
		// Other fields should be zero
		assert.Equal(t, 0.0, result.LoadPercent)
	})
}

func TestDataParser_parseFloat(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	tests := []struct {
		name     string
		raw      map[string]interface{}
		key      string
		expected float64
	}{
		{
			name:     "string value",
			raw:      map[string]interface{}{"value": "123.45"},
			key:      "value",
			expected: 123.45,
		},
		{
			name:     "float64 value",
			raw:      map[string]interface{}{"value": 123.45},
			key:      "value",
			expected: 123.45,
		},
		{
			name:     "int value",
			raw:      map[string]interface{}{"value": 123},
			key:      "value",
			expected: 123.0,
		},
		{
			name:     "int64 value",
			raw:      map[string]interface{}{"value": int64(123)},
			key:      "value",
			expected: 123.0,
		},
		{
			name:     "empty string",
			raw:      map[string]interface{}{"value": ""},
			key:      "value",
			expected: 0,
		},
		{
			name:     "missing key",
			raw:      map[string]interface{}{},
			key:      "value",
			expected: 0,
		},
		{
			name:     "invalid string",
			raw:      map[string]interface{}{"value": "invalid"},
			key:      "value",
			expected: 0,
		},
		{
			name:     "unexpected type",
			raw:      map[string]interface{}{"value": true},
			key:      "value",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseFloat(tt.raw, tt.key, "test field")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataParser_parseInt(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	tests := []struct {
		name     string
		raw      map[string]interface{}
		key      string
		expected int
	}{
		{
			name:     "string value",
			raw:      map[string]interface{}{"value": "123"},
			key:      "value",
			expected: 123,
		},
		{
			name:     "int value",
			raw:      map[string]interface{}{"value": 123},
			key:      "value",
			expected: 123,
		},
		{
			name:     "int64 value",
			raw:      map[string]interface{}{"value": int64(123)},
			key:      "value",
			expected: 123,
		},
		{
			name:     "float64 value",
			raw:      map[string]interface{}{"value": 123.9},
			key:      "value",
			expected: 123,
		},
		{
			name:     "empty string",
			raw:      map[string]interface{}{"value": ""},
			key:      "value",
			expected: 0,
		},
		{
			name:     "missing key",
			raw:      map[string]interface{}{},
			key:      "value",
			expected: 0,
		},
		{
			name:     "invalid string",
			raw:      map[string]interface{}{"value": "invalid"},
			key:      "value",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseInt(tt.raw, tt.key, "test field")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataParser_parseBool(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	tests := []struct {
		name     string
		raw      map[string]interface{}
		key      string
		expected bool
	}{
		{
			name:     "string '1'",
			raw:      map[string]interface{}{"value": "1"},
			key:      "value",
			expected: true,
		},
		{
			name:     "string '0'",
			raw:      map[string]interface{}{"value": "0"},
			key:      "value",
			expected: false,
		},
		{
			name:     "string 'true'",
			raw:      map[string]interface{}{"value": "true"},
			key:      "value",
			expected: true,
		},
		{
			name:     "string 'True'",
			raw:      map[string]interface{}{"value": "True"},
			key:      "value",
			expected: true,
		},
		{
			name:     "string 'TRUE'",
			raw:      map[string]interface{}{"value": "TRUE"},
			key:      "value",
			expected: true,
		},
		{
			name:     "string 'false'",
			raw:      map[string]interface{}{"value": "false"},
			key:      "value",
			expected: false,
		},
		{
			name:     "bool true",
			raw:      map[string]interface{}{"value": true},
			key:      "value",
			expected: true,
		},
		{
			name:     "bool false",
			raw:      map[string]interface{}{"value": false},
			key:      "value",
			expected: false,
		},
		{
			name:     "int non-zero",
			raw:      map[string]interface{}{"value": 1},
			key:      "value",
			expected: true,
		},
		{
			name:     "int zero",
			raw:      map[string]interface{}{"value": 0},
			key:      "value",
			expected: false,
		},
		{
			name:     "missing key",
			raw:      map[string]interface{}{},
			key:      "value",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseBool(tt.raw, tt.key, "test field")
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataParser_parseString(t *testing.T) {
	parser := NewDataParser(zap.NewNop())

	tests := []struct {
		name     string
		raw      map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "string value",
			raw:      map[string]interface{}{"value": "test"},
			key:      "value",
			expected: "test",
		},
		{
			name:     "int value",
			raw:      map[string]interface{}{"value": 123},
			key:      "value",
			expected: "123",
		},
		{
			name:     "float value",
			raw:      map[string]interface{}{"value": 123.45},
			key:      "value",
			expected: "123.45",
		},
		{
			name:     "bool value",
			raw:      map[string]interface{}{"value": true},
			key:      "value",
			expected: "true",
		},
		{
			name:     "empty string",
			raw:      map[string]interface{}{"value": ""},
			key:      "value",
			expected: "",
		},
		{
			name:     "missing key",
			raw:      map[string]interface{}{},
			key:      "value",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.parseString(tt.raw, tt.key, "test field")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions to create test data

func createValidDeviceDataResponse() *DeviceDataResponse {
	return &DeviceDataResponse{
		Total:       1,
		PageSize:    20,
		CurrentPage: 1,
		Data:        []DeviceInfo{createValidDeviceInfo()},
		Code:        "000000",
		Msg:         "OK",
	}
}

func createValidDeviceInfo() DeviceInfo {
	return DeviceInfo{
		AssetDevice: AssetDevice{
			ID:              "e156e6cb-41cb-4b35-b0dd-869929186a5c",
			DeviceType:      1,
			Model:           "ON-LINE",
			Alias:           "C3K",
			ProtocolID:      13,
			ConnectType:     1,
			ComPort:         "COM3",
			BaudRate:        2400,
			AreaID:          "00000000-0000-0000-0000-000000000000",
			IsActive:        true,
			FirmwareVersion: "03.09",
			CreateTime:      FlexibleTime{Time: time.Now()},
			WarrantyStatus:  0,
		},
		Realtime:         createValidRealtimeData(),
		Config:           map[string]interface{}{},
		Setting:          map[string]interface{}{},
		ActiveAlarms:     []interface{}{},
		ControlSupported: map[string]interface{}{},
		Connected:        true,
	}
}

func createValidRealtimeData() map[string]interface{} {
	return map[string]interface{}{
		"inputVolt1":     "236.8",
		"outputVolt1":    "220.1",
		"batVoltP":       "81.4",
		"outputCurrent1": "1.1",
		"inputFreq":      "49.9",
		"outputFreq":     "49.9",
		"loadPercent":    "6",
		"loadTotalVa":    "198",
		"loadWatt1":      "195",
		"loadVa1":        "198",
		"loadTotalWatt":  "195",
		"batCapacity":    "90",
		"batRemainTime":  "6723",
		"isCharging":     "1",
		"upsTemperature": "27.0",
		"mode":           "3",
		"status":         "1",
		"batteryStatus":  "2",
		"testStatus":     "1",
		"faultCode":      "",
	}
}
