package winpower

import (
	"testing"
	"time"
)

// TestDeviceData_Validate tests the Validate method for DeviceData
func TestDeviceData_Validate(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "valid_device_data_should_pass",
			test: func(t *testing.T) {
				data := &DeviceData{
					ID:            "device-001",
					Name:          "Test Device",
					Type:          "UPS",
					Status:        "online",
					Timestamp:     time.Now(),
					InputVoltage:  220.0,
					OutputVoltage: 220.0,
					Power:         1000.0,
				}

				err := data.Validate()
				if err != nil {
					t.Errorf("Validate() should not return error for valid data: %v", err)
				}
			},
		},
		{
			name: "empty_device_id_should_error",
			test: func(t *testing.T) {
				data := &DeviceData{
					ID: "",
				}

				err := data.Validate()
				if err == nil {
					t.Error("Validate() should return error for empty device ID")
				}
			},
		},
		{
			name: "negative_voltage_should_error",
			test: func(t *testing.T) {
				data := &DeviceData{
					ID:           "device-001",
					InputVoltage: -10.0,
				}

				err := data.Validate()
				if err == nil {
					t.Error("Validate() should return error for negative input voltage")
				}
			},
		},
		{
			name: "nil_device_data_should_error",
			test: func(t *testing.T) {
				var data *DeviceData
				err := data.Validate()
				if err == nil {
					t.Error("Validate() should return error for nil device data")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestDataParser_ParseDeviceData tests the ParseDeviceData method
func TestDataParser_ParseDeviceData(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "parse_empty_json_should_error",
			test: func(t *testing.T) {
				parser := NewDataParser()
				_, err := parser.ParseDeviceData([]byte(""))
				if err == nil {
					t.Error("ParseDeviceData() should return error for empty JSON")
				}
			},
		},
		{
			name: "parse_invalid_json_should_error",
			test: func(t *testing.T) {
				parser := NewDataParser()
				_, err := parser.ParseDeviceData([]byte("invalid json"))
				if err == nil {
					t.Error("ParseDeviceData() should return error for invalid JSON")
				}
			},
		},
		{
			name: "parse_nil_data_should_error",
			test: func(t *testing.T) {
				parser := NewDataParser()
				_, err := parser.ParseDeviceData(nil)
				if err == nil {
					t.Error("ParseDeviceData() should return error for nil data")
				}
			},
		},
		{
			name: "parse_valid_json_should_succeed",
			test: func(t *testing.T) {
				parser := NewDataParser()
				jsonData := []byte(`{
					"id": "device-001",
					"name": "Test Device",
					"type": "UPS",
					"status": "online",
					"input_voltage": 220.0,
					"output_voltage": 220.0,
					"power": 1000.0
				}`)

				data, err := parser.ParseDeviceData(jsonData)
				if err != nil {
					t.Errorf("ParseDeviceData() should not return error for valid JSON: %v", err)
				}

				if data == nil {
					t.Error("ParseDeviceData() should return non-nil data for valid JSON")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestDataParser_ParseMultipleDevices tests the ParseMultipleDevices method
func TestDataParser_ParseMultipleDevices(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "parse_empty_array_should_return_empty_slice",
			test: func(t *testing.T) {
				parser := NewDataParser()
				devices, err := parser.ParseMultipleDevices([]byte("[]"))
				if err != nil {
					t.Errorf("ParseMultipleDevices() should not return error for empty array: %v", err)
				}

				if len(devices) != 0 {
					t.Errorf("ParseMultipleDevices() should return empty slice for empty array, got: %d", len(devices))
				}
			},
		},
		{
			name: "parse_invalid_array_should_error",
			test: func(t *testing.T) {
				parser := NewDataParser()
				_, err := parser.ParseMultipleDevices([]byte("invalid"))
				if err == nil {
					t.Error("ParseMultipleDevices() should return error for invalid JSON")
				}
			},
		},
		{
			name: "parse_nil_data_should_error",
			test: func(t *testing.T) {
				parser := NewDataParser()
				_, err := parser.ParseMultipleDevices(nil)
				if err == nil {
					t.Error("ParseMultipleDevices() should return error for nil data")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestDeviceData_IsOnline tests the IsOnline method
func TestDeviceData_IsOnline(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "online_status_should_return_true",
			test: func(t *testing.T) {
				data := &DeviceData{Status: "online"}
				if !data.IsOnline() {
					t.Error("IsOnline() should return true for online status")
				}
			},
		},
		{
			name: "offline_status_should_return_false",
			test: func(t *testing.T) {
				data := &DeviceData{Status: "offline"}
				if data.IsOnline() {
					t.Error("IsOnline() should return false for offline status")
				}
			},
		},
		{
			name: "empty_status_should_return_false",
			test: func(t *testing.T) {
				data := &DeviceData{Status: ""}
				if data.IsOnline() {
					t.Error("IsOnline() should return false for empty status")
				}
			},
		},
		{
			name: "nil_data_should_return_false",
			test: func(t *testing.T) {
				var data *DeviceData
				if data.IsOnline() {
					t.Error("IsOnline() should return false for nil data")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestDeviceData_GetPower tests the GetPower method
func TestDeviceData_GetPower(t *testing.T) {
	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "should_return_power_value",
			test: func(t *testing.T) {
				data := &DeviceData{Power: 1500.5}
				power := data.GetPower()
				if power != 1500.5 {
					t.Errorf("GetPower() should return 1500.5, got: %f", power)
				}
			},
		},
		{
			name: "nil_data_should_return_zero",
			test: func(t *testing.T) {
				var data *DeviceData
				power := data.GetPower()
				if power != 0 {
					t.Errorf("GetPower() should return 0 for nil data, got: %f", power)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}
