package winpower

import (
	"encoding/json"
	"fmt"
	"time"
)

// DeviceData represents a WinPower device with its electrical measurements
type DeviceData struct {
	// ID is the unique identifier for the device
	ID string `json:"id" yaml:"id"`

	// Name is the human-readable name of the device
	Name string `json:"name" yaml:"name"`

	// Type is the type of device (e.g., "UPS", "PDU", "STS")
	Type string `json:"type" yaml:"type"`

	// Status is the current status of the device (e.g., "online", "offline")
	Status string `json:"status" yaml:"status"`

	// Timestamp is when the data was collected
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`

	// Input electrical measurements
	InputVoltage float64 `json:"input_voltage" yaml:"input_voltage"`     // Input voltage in volts
	InputCurrent float64 `json:"input_current" yaml:"input_current"`     // Input current in amperes
	InputFreq    float64 `json:"input_frequency" yaml:"input_frequency"` // Input frequency in Hz
	InputPower   float64 `json:"input_power" yaml:"input_power"`         // Input power in watts

	// Output electrical measurements
	OutputVoltage float64 `json:"output_voltage" yaml:"output_voltage"`     // Output voltage in volts
	OutputCurrent float64 `json:"output_current" yaml:"output_current"`     // Output current in amperes
	OutputFreq    float64 `json:"output_frequency" yaml:"output_frequency"` // Output frequency in Hz
	OutputPower   float64 `json:"output_power" yaml:"output_power"`         // Output power in watts

	// Power measurements
	Power         float64 `json:"power" yaml:"power"`                   // Active power in watts
	ApparentPower float64 `json:"apparent_power" yaml:"apparent_power"` // Apparent power in VA
	PowerFactor   float64 `json:"power_factor" yaml:"power_factor"`     // Power factor (0-1)

	// Load measurements
	LoadPercent float64 `json:"load_percent" yaml:"load_percent"` // Load percentage (0-100)

	// Battery measurements (for UPS devices)
	BatteryCapacity float64 `json:"battery_capacity" yaml:"battery_capacity"` // Battery capacity percentage (0-100)
	BatteryVoltage  float64 `json:"battery_voltage" yaml:"battery_voltage"`   // Battery voltage in volts
	BatteryCurrent  float64 `json:"battery_current" yaml:"battery_current"`   // Battery current in amperes
	RuntimeMinutes  float64 `json:"runtime_minutes" yaml:"runtime_minutes"`   // Estimated runtime in minutes

	// Environmental measurements
	Temperature float64 `json:"temperature" yaml:"temperature"` // Temperature in Celsius
	Humidity    float64 `json:"humidity" yaml:"humidity"`       // Humidity percentage (0-100)

	// Additional metadata
	Location     string            `json:"location" yaml:"location"`           // Physical location of device
	Model        string            `json:"model" yaml:"model"`                 // Device model
	SerialNumber string            `json:"serial_number" yaml:"serial_number"` // Device serial number
	Firmware     string            `json:"firmware" yaml:"firmware"`           // Firmware version
	Tags         map[string]string `json:"tags" yaml:"tags"`                   // Additional tags
}

// Validate checks if the device data is valid
func (dd *DeviceData) Validate() error {
	if dd == nil {
		return fmt.Errorf("device data cannot be nil")
	}

	if dd.ID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}

	// Validate voltage values are non-negative
	if dd.InputVoltage < 0 {
		return fmt.Errorf("input voltage cannot be negative: %f", dd.InputVoltage)
	}

	if dd.OutputVoltage < 0 {
		return fmt.Errorf("output voltage cannot be negative: %f", dd.OutputVoltage)
	}

	// Validate current values are non-negative
	if dd.InputCurrent < 0 {
		return fmt.Errorf("input current cannot be negative: %f", dd.InputCurrent)
	}

	if dd.OutputCurrent < 0 {
		return fmt.Errorf("output current cannot be negative: %f", dd.OutputCurrent)
	}

	// Validate frequency values are reasonable (typical range: 40-70 Hz)
	if dd.InputFreq > 0 && (dd.InputFreq < 40 || dd.InputFreq > 70) {
		return fmt.Errorf("input frequency out of reasonable range: %f Hz", dd.InputFreq)
	}

	if dd.OutputFreq > 0 && (dd.OutputFreq < 40 || dd.OutputFreq > 70) {
		return fmt.Errorf("output frequency out of reasonable range: %f Hz", dd.OutputFreq)
	}

	// Validate power factor is within valid range (0-1)
	if dd.PowerFactor < 0 || dd.PowerFactor > 1 {
		return fmt.Errorf("power factor must be between 0 and 1: %f", dd.PowerFactor)
	}

	// Validate load percentage is within valid range (0-100)
	if dd.LoadPercent < 0 || dd.LoadPercent > 100 {
		return fmt.Errorf("load percentage must be between 0 and 100: %f", dd.LoadPercent)
	}

	// Validate battery capacity is within valid range (0-100)
	if dd.BatteryCapacity < 0 || dd.BatteryCapacity > 100 {
		return fmt.Errorf("battery capacity must be between 0 and 100: %f", dd.BatteryCapacity)
	}

	// Validate humidity is within valid range (0-100)
	if dd.Humidity < 0 || dd.Humidity > 100 {
		return fmt.Errorf("humidity must be between 0 and 100: %f", dd.Humidity)
	}

	return nil
}

// IsOnline checks if the device is online
func (dd *DeviceData) IsOnline() bool {
	if dd == nil {
		return false
	}
	return dd.Status == "online"
}

// GetPower returns the active power of the device
func (dd *DeviceData) GetPower() float64 {
	if dd == nil {
		return 0
	}
	return dd.Power
}

// GetVoltage returns the output voltage (or input voltage if output is not available)
func (dd *DeviceData) GetVoltage() float64 {
	if dd == nil {
		return 0
	}
	if dd.OutputVoltage > 0 {
		return dd.OutputVoltage
	}
	return dd.InputVoltage
}

// GetCurrent returns the output current (or input current if output is not available)
func (dd *DeviceData) GetCurrent() float64 {
	if dd == nil {
		return 0
	}
	if dd.OutputCurrent > 0 {
		return dd.OutputCurrent
	}
	return dd.InputCurrent
}

// HasBattery checks if the device has battery information
func (dd *DeviceData) HasBattery() bool {
	if dd == nil {
		return false
	}
	return dd.BatteryCapacity > 0
}

// GetBatteryCapacity returns the battery capacity percentage
func (dd *DeviceData) GetBatteryCapacity() float64 {
	if dd == nil {
		return 0
	}
	return dd.BatteryCapacity
}

// GetRuntimeMinutes returns the estimated runtime in minutes
func (dd *DeviceData) GetRuntimeMinutes() float64 {
	if dd == nil {
		return 0
	}
	return dd.RuntimeMinutes
}

// Clone creates a deep copy of the device data
func (dd *DeviceData) Clone() *DeviceData {
	if dd == nil {
		return nil
	}

	clone := *dd

	// Deep copy tags map
	if dd.Tags != nil {
		clone.Tags = make(map[string]string)
		for k, v := range dd.Tags {
			clone.Tags[k] = v
		}
	}

	return &clone
}

// String returns a string representation of the device data
func (dd *DeviceData) String() string {
	if dd == nil {
		return "<nil>"
	}

	return fmt.Sprintf(
		"DeviceData{ID: %s, Name: %s, Type: %s, Status: %s, Power: %.2fW, Voltage: %.1fV, Current: %.2fA}",
		dd.ID, dd.Name, dd.Type, dd.Status, dd.Power, dd.GetVoltage(), dd.GetCurrent(),
	)
}

// DataParser handles parsing of WinPower API responses
type DataParser struct{}

// NewDataParser creates a new data parser
func NewDataParser() *DataParser {
	return &DataParser{}
}

// ParseDeviceData parses JSON data into a DeviceData structure
func (dp *DataParser) ParseDeviceData(data []byte) (*DeviceData, error) {
	if dp == nil {
		return nil, fmt.Errorf("data parser is nil")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var device DeviceData
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("failed to parse device data: %w", err)
	}

	// Set timestamp if not provided in the data
	if device.Timestamp.IsZero() {
		device.Timestamp = time.Now()
	}

	return &device, nil
}

// ParseMultipleDevices parses JSON array into multiple DeviceData structures
func (dp *DataParser) ParseMultipleDevices(data []byte) ([]*DeviceData, error) {
	if dp == nil {
		return nil, fmt.Errorf("data parser is nil")
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	var devices []DeviceData
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("failed to parse multiple devices: %w", err)
	}

	// Convert slice of DeviceData to slice of pointers to DeviceData
	result := make([]*DeviceData, len(devices))
	now := time.Now()

	for i, device := range devices {
		// Set timestamp if not provided in the data
		if device.Timestamp.IsZero() {
			device.Timestamp = now
		}

		// Create a copy to avoid pointer issues
		deviceCopy := device
		result[i] = &deviceCopy
	}

	return result, nil
}

// ValidateAndParse validates and parses device data in one step
func (dp *DataParser) ValidateAndParse(data []byte) (*DeviceData, error) {
	device, err := dp.ParseDeviceData(data)
	if err != nil {
		return nil, err
	}

	if err := device.Validate(); err != nil {
		return nil, fmt.Errorf("invalid device data: %w", err)
	}

	return device, nil
}

// ValidateAndParseMultiple validates and parses multiple devices in one step
func (dp *DataParser) ValidateAndParseMultiple(data []byte) ([]*DeviceData, error) {
	devices, err := dp.ParseMultipleDevices(data)
	if err != nil {
		return nil, err
	}

	for i, device := range devices {
		if err := device.Validate(); err != nil {
			return nil, fmt.Errorf("invalid device data at index %d: %w", i, err)
		}
	}

	return devices, nil
}
