package winpower

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

// FlexibleTime is a custom time type that can parse multiple time formats.
// It handles WinPower's time format which doesn't include timezone information.
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It attempts to parse time in multiple formats:
// 1. RFC3339 with timezone (e.g., "2006-01-02T15:04:05Z07:00")
// 2. RFC3339 without timezone (e.g., "2006-01-02T15:04:05.999999")
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from JSON string
	s := strings.Trim(string(data), "\"")
	if s == "null" || s == "" {
		return nil
	}

	// Try parsing with RFC3339 format (with timezone)
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try parsing without timezone (WinPower format)
	// Format: "2006-01-02T15:04:05.999999"
	t, err = time.Parse("2006-01-02T15:04:05.999999", s)
	if err == nil {
		ft.Time = t
		return nil
	}

	// Try parsing without microseconds
	// Format: "2006-01-02T15:04:05"
	t, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		ft.Time = t
		return nil
	}

	return err
}

// MarshalJSON implements the json.Marshaler interface.
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Format(time.RFC3339))
}

// WinPowerClient defines the interface for WinPower G2 data collection.
type WinPowerClient interface {
	// CollectDeviceData collects device data from WinPower system.
	CollectDeviceData(ctx context.Context) ([]ParsedDeviceData, error)

	// GetConnectionStatus returns the current connection status.
	GetConnectionStatus() bool

	// GetLastCollectionTime returns the timestamp of the last successful collection.
	GetLastCollectionTime() time.Time
}

// ParsedDeviceData represents standardized device data structure.
type ParsedDeviceData struct {
	// Device basic information
	DeviceID   string `json:"device_id"`
	DeviceType int    `json:"device_type"`
	Model      string `json:"model"`
	Alias      string `json:"alias"`

	// Connection status
	Connected bool `json:"connected"`

	// Realtime data
	Realtime RealtimeData `json:"realtime"`

	// Collection metadata
	CollectedAt time.Time `json:"collected_at"`
}

// RealtimeData represents real-time device data.
type RealtimeData struct {
	// Power data (key for energy calculation)
	LoadTotalWatt float64 `json:"load_total_watt"` // Total active power in Watts

	// Voltage data
	InputVolt1  float64 `json:"input_volt_1"`  // Input voltage phase 1
	OutputVolt1 float64 `json:"output_volt_1"` // Output voltage phase 1
	BatVoltP    float64 `json:"bat_volt_p"`    // Battery voltage percentage

	// Current data
	OutputCurrent1 float64 `json:"output_current_1"` // Output current phase 1

	// Frequency data
	InputFreq  float64 `json:"input_freq"`  // Input frequency
	OutputFreq float64 `json:"output_freq"` // Output frequency

	// Load data
	LoadPercent float64 `json:"load_percent"`  // Load percentage
	LoadTotalVa float64 `json:"load_total_va"` // Total apparent power
	LoadWatt1   float64 `json:"load_watt_1"`   // Active power phase 1
	LoadVa1     float64 `json:"load_va_1"`     // Apparent power phase 1

	// Battery data
	BatCapacity   float64 `json:"bat_capacity"`    // Battery capacity percentage
	BatRemainTime int     `json:"bat_remain_time"` // Battery remaining time in seconds
	IsCharging    bool    `json:"is_charging"`     // Charging status

	// Status data
	UpsTemperature float64 `json:"ups_temperature"` // UPS temperature in Celsius
	Mode           string  `json:"mode"`            // UPS working mode
	Status         string  `json:"status"`          // Device status
	BatteryStatus  string  `json:"battery_status"`  // Battery status
	TestStatus     string  `json:"test_status"`     // Test status
	FaultCode      string  `json:"fault_code"`      // Fault code

	// Raw data for additional fields
	Raw map[string]interface{} `json:"raw,omitempty"`
}

// DeviceDataResponse represents the API response structure from WinPower.
type DeviceDataResponse struct {
	Total       int          `json:"total"`
	PageSize    int          `json:"pageSize"`
	CurrentPage int          `json:"currentPage"`
	Data        []DeviceInfo `json:"data"`
	Code        string       `json:"code"`
	Msg         string       `json:"msg"`
}

// ErrorResponse represents an error response structure from WinPower API.
// Used when the API returns an error (e.g., authentication failure).
// In error cases, the 'data' field is a string instead of an array.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"` // Error details as string
}

// DeviceInfo represents complete device information from WinPower API.
type DeviceInfo struct {
	AssetDevice      AssetDevice            `json:"assetDevice"`
	Realtime         map[string]interface{} `json:"realtime"`
	Config           map[string]interface{} `json:"config"`
	Setting          map[string]interface{} `json:"setting"`
	ActiveAlarms     []interface{}          `json:"activeAlarms"`
	ControlSupported map[string]interface{} `json:"controlSupported"`
	Connected        bool                   `json:"connected"`
}

// AssetDevice represents device basic information.
type AssetDevice struct {
	ID              string       `json:"id"`
	DeviceType      int          `json:"deviceType"`
	Model           string       `json:"model"`
	Alias           string       `json:"alias"`
	ProtocolID      int          `json:"protocolId"`
	ConnectType     int          `json:"connectType"`
	ComPort         string       `json:"comPort"`
	BaudRate        int          `json:"baudRate"`
	AreaID          string       `json:"areaId"`
	IsActive        bool         `json:"isActive"`
	FirmwareVersion string       `json:"firmwareVersion"`
	CreateTime      FlexibleTime `json:"createTime"`
	WarrantyStatus  int          `json:"warrantyStatus"`
}

// LoginRequest represents the authentication request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the authentication response.
type LoginResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DeviceID string `json:"deviceId"`
		Token    string `json:"token"`
	} `json:"data"`
}

// TokenCache represents cached token information.
type TokenCache struct {
	Token     string
	ExpiresAt time.Time
	DeviceID  string
}
