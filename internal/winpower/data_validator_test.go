package winpower

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewDataValidator(t *testing.T) {
	t.Run("with logger", func(t *testing.T) {
		logger := zap.NewNop()
		validator := NewDataValidator(logger)
		assert.NotNil(t, validator)
		assert.Equal(t, logger, validator.logger)
	})

	t.Run("with nil logger", func(t *testing.T) {
		validator := NewDataValidator(nil)
		assert.NotNil(t, validator)
		assert.NotNil(t, validator.logger)
	})
}

func TestDataValidator_Validate(t *testing.T) {
	validator := NewDataValidator(zap.NewNop())

	t.Run("nil data", func(t *testing.T) {
		result := validator.Validate(nil)
		assert.False(t, result.IsValid)
		require.Len(t, result.Errors, 1)
		assert.Equal(t, "data", result.Errors[0].Field)
	})

	t.Run("valid data", func(t *testing.T) {
		data := createValidParsedDeviceData()
		result := validator.Validate(data)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("missing device_id", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.DeviceID = ""
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.Contains(t, getFieldNames(result.Errors), "device_id")
	})

	t.Run("negative device_type", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.DeviceType = -1
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.Contains(t, getFieldNames(result.Errors), "device_type")
	})

	t.Run("multiple errors", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.DeviceID = ""
		data.Realtime.LoadTotalWatt = -100
		data.Realtime.InputVolt1 = 1000 // out of range
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.GreaterOrEqual(t, len(result.Errors), 3)
	})
}

func TestDataValidator_validateRealtimeData(t *testing.T) {
	validator := NewDataValidator(zap.NewNop())

	t.Run("negative power values", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.Realtime.LoadTotalWatt = -100
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.Contains(t, getFieldNames(result.Errors), "load_total_watt")
	})

	t.Run("negative load_watt1", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.Realtime.LoadWatt1 = -50
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.Contains(t, getFieldNames(result.Errors), "load_watt1")
	})

	t.Run("voltage out of range", func(t *testing.T) {
		tests := []struct {
			name      string
			setValue  func(*ParsedDeviceData)
			fieldName string
		}{
			{
				name: "input_volt_1 too high",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.InputVolt1 = 600
				},
				fieldName: "input_volt_1",
			},
			{
				name: "output_volt_1 negative",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.OutputVolt1 = -10
				},
				fieldName: "output_volt_1",
			},
			{
				name: "bat_volt_p too high",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.BatVoltP = 501
				},
				fieldName: "bat_volt_p",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := createValidParsedDeviceData()
				tt.setValue(data)
				result := validator.Validate(data)
				assert.False(t, result.IsValid)
				assert.Contains(t, getFieldNames(result.Errors), tt.fieldName)
			})
		}
	})

	t.Run("negative current", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.Realtime.OutputCurrent1 = -5
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.Contains(t, getFieldNames(result.Errors), "output_current_1")
	})

	t.Run("frequency out of range", func(t *testing.T) {
		tests := []struct {
			name      string
			frequency float64
			fieldName string
		}{
			{
				name:      "input_freq too low",
				frequency: 40,
				fieldName: "input_freq",
			},
			{
				name:      "output_freq too high",
				frequency: 70,
				fieldName: "output_freq",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := createValidParsedDeviceData()
				if tt.fieldName == "input_freq" {
					data.Realtime.InputFreq = tt.frequency
				} else {
					data.Realtime.OutputFreq = tt.frequency
				}
				result := validator.Validate(data)
				assert.False(t, result.IsValid)
				assert.Contains(t, getFieldNames(result.Errors), tt.fieldName)
			})
		}
	})

	t.Run("frequency zero is allowed", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.Realtime.InputFreq = 0
		data.Realtime.OutputFreq = 0
		result := validator.Validate(data)
		// Should not fail on zero frequency (means missing data)
		assert.True(t, result.IsValid)
	})

	t.Run("percentage out of range", func(t *testing.T) {
		tests := []struct {
			name      string
			setValue  func(*ParsedDeviceData)
			fieldName string
		}{
			{
				name: "load_percent negative",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.LoadPercent = -10
				},
				fieldName: "load_percent",
			},
			{
				name: "load_percent over 100",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.LoadPercent = 150
				},
				fieldName: "load_percent",
			},
			{
				name: "bat_capacity negative",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.BatCapacity = -5
				},
				fieldName: "bat_capacity",
			},
			{
				name: "bat_capacity over 100",
				setValue: func(d *ParsedDeviceData) {
					d.Realtime.BatCapacity = 105
				},
				fieldName: "bat_capacity",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := createValidParsedDeviceData()
				tt.setValue(data)
				result := validator.Validate(data)
				assert.False(t, result.IsValid)
				assert.Contains(t, getFieldNames(result.Errors), tt.fieldName)
			})
		}
	})

	t.Run("temperature out of range", func(t *testing.T) {
		tests := []struct {
			name        string
			temperature float64
			shouldFail  bool
		}{
			{"normal temperature", 25.0, false},
			{"low temperature", -10.0, false},
			{"too cold", -25.0, true},
			{"high temperature", 80.0, false},
			{"too hot", 110.0, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := createValidParsedDeviceData()
				data.Realtime.UpsTemperature = tt.temperature
				result := validator.Validate(data)
				if tt.shouldFail {
					assert.False(t, result.IsValid)
					assert.Contains(t, getFieldNames(result.Errors), "ups_temperature")
				} else {
					assert.True(t, result.IsValid)
				}
			})
		}
	})

	t.Run("negative battery remain time", func(t *testing.T) {
		data := createValidParsedDeviceData()
		data.Realtime.BatRemainTime = -100
		result := validator.Validate(data)
		assert.False(t, result.IsValid)
		assert.Contains(t, getFieldNames(result.Errors), "bat_remain_time")
	})
}

func TestDataValidator_ValidateBatch(t *testing.T) {
	validator := NewDataValidator(zap.NewNop())

	t.Run("empty batch", func(t *testing.T) {
		results := validator.ValidateBatch([]ParsedDeviceData{})
		assert.Empty(t, results)
	})

	t.Run("single valid device", func(t *testing.T) {
		data := []ParsedDeviceData{*createValidParsedDeviceData()}
		results := validator.ValidateBatch(data)
		require.Len(t, results, 1)
		assert.True(t, results[0].IsValid)
	})

	t.Run("multiple devices with mixed validity", func(t *testing.T) {
		validData := createValidParsedDeviceData()
		invalidData := createValidParsedDeviceData()
		invalidData.DeviceID = ""

		data := []ParsedDeviceData{*validData, *invalidData}
		results := validator.ValidateBatch(data)
		require.Len(t, results, 2)
		assert.True(t, results[0].IsValid)
		assert.False(t, results[1].IsValid)
	})
}

func TestDataValidator_HasCriticalErrors(t *testing.T) {
	validator := NewDataValidator(zap.NewNop())

	t.Run("nil result", func(t *testing.T) {
		hasCritical := validator.HasCriticalErrors(nil)
		assert.False(t, hasCritical)
	})

	t.Run("valid result", func(t *testing.T) {
		result := &ValidationResult{IsValid: true}
		hasCritical := validator.HasCriticalErrors(result)
		assert.False(t, hasCritical)
	})

	t.Run("non-critical errors", func(t *testing.T) {
		result := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "load_total_watt", Message: "negative value"},
			},
		}
		hasCritical := validator.HasCriticalErrors(result)
		assert.False(t, hasCritical)
	})

	t.Run("critical error - nil data", func(t *testing.T) {
		result := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "data", Message: "data is nil"},
			},
		}
		hasCritical := validator.HasCriticalErrors(result)
		assert.True(t, hasCritical)
	})

	t.Run("critical error - missing device_id", func(t *testing.T) {
		result := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "device_id", Message: "device_id is empty"},
			},
		}
		hasCritical := validator.HasCriticalErrors(result)
		assert.True(t, hasCritical)
	})

	t.Run("mixed critical and non-critical errors", func(t *testing.T) {
		result := &ValidationResult{
			IsValid: false,
			Errors: []ValidationError{
				{Field: "load_total_watt", Message: "negative value"},
				{Field: "device_id", Message: "device_id is empty"},
			},
		}
		hasCritical := validator.HasCriticalErrors(result)
		assert.True(t, hasCritical)
	})
}

// Helper functions

func createValidParsedDeviceData() *ParsedDeviceData {
	return &ParsedDeviceData{
		DeviceID:   "test-device-001",
		DeviceType: 1,
		Model:      "ON-LINE",
		Alias:      "TestUPS",
		Connected:  true,
		Realtime: RealtimeData{
			LoadTotalWatt:  195.0,
			InputVolt1:     236.8,
			OutputVolt1:    220.1,
			BatVoltP:       81.4,
			OutputCurrent1: 1.1,
			InputFreq:      49.9,
			OutputFreq:     49.9,
			LoadPercent:    6.0,
			LoadTotalVa:    198.0,
			LoadWatt1:      195.0,
			LoadVa1:        198.0,
			BatCapacity:    90.0,
			BatRemainTime:  6723,
			IsCharging:     true,
			UpsTemperature: 27.0,
			Mode:           "3",
			Status:         "1",
			BatteryStatus:  "2",
			TestStatus:     "1",
			FaultCode:      "",
		},
		CollectedAt: time.Now(),
	}
}

func getFieldNames(errors []ValidationError) []string {
	names := make([]string, len(errors))
	for i, err := range errors {
		names[i] = err.Field
	}
	return names
}
