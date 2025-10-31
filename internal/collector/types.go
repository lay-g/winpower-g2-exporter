package collector

import (
	"time"
)

// CollectionResult represents the result of a data collection operation
type CollectionResult struct {
	Success        bool                             `json:"success"`
	DeviceCount    int                              `json:"device_count"`
	Devices        map[string]*DeviceCollectionInfo `json:"devices"`
	CollectionTime time.Time                        `json:"collection_time"`
	Duration       time.Duration                    `json:"duration"`
	ErrorMessage   string                           `json:"error_message,omitempty"`

	// Token information
	TokenValid     bool      `json:"token_valid"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
}

// DeviceCollectionInfo contains comprehensive information about a collected device
type DeviceCollectionInfo struct {
	// Basic information
	DeviceID       string    `json:"device_id"`
	DeviceName     string    `json:"device_name"`
	DeviceType     int       `json:"device_type"`
	DeviceModel    string    `json:"device_model"`
	Connected      bool      `json:"connected"`
	LastUpdateTime time.Time `json:"last_update_time"`

	// Electrical parameters
	InputVolt1        float64 `json:"input_volt_1"`
	InputFreq         float64 `json:"input_freq"`
	OutputVolt1       float64 `json:"output_volt_1"`
	OutputCurrent1    float64 `json:"output_current_1"`
	OutputFreq        float64 `json:"output_freq"`
	OutputVoltageType string  `json:"output_voltage_type"`

	// Load and power parameters
	LoadPercent   float64 `json:"load_percent"`
	LoadTotalWatt float64 `json:"load_total_watt"` // Core field for energy calculation
	LoadTotalVa   float64 `json:"load_total_va"`
	LoadWatt1     float64 `json:"load_watt_1"`
	LoadVa1       float64 `json:"load_va_1"`

	// Battery parameters
	IsCharging    bool    `json:"is_charging"`
	BatVoltP      float64 `json:"bat_volt_p"`
	BatCapacity   float64 `json:"bat_capacity"`
	BatRemainTime int     `json:"bat_remain_time"`
	BatteryStatus string  `json:"battery_status"`

	// UPS status parameters
	UpsTemperature float64 `json:"ups_temperature"`
	Mode           string  `json:"mode"`
	Status         string  `json:"status"`
	TestStatus     string  `json:"test_status"`
	FaultCode      string  `json:"fault_code"`

	// Energy calculation result
	EnergyCalculated bool    `json:"energy_calculated"`
	EnergyValue      float64 `json:"energy_value"` // Cumulative energy in Wh

	// Error information
	ErrorMsg string `json:"error_msg,omitempty"`
}
