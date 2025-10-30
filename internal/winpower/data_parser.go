package winpower

import (
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// DataParser parses WinPower API responses into standardized data structures.
type DataParser struct {
	logger *zap.Logger
}

// NewDataParser creates a new DataParser instance.
func NewDataParser(logger *zap.Logger) *DataParser {
	if logger == nil {
		// Use a no-op logger as fallback
		logger = zap.NewNop()
	}
	return &DataParser{
		logger: logger,
	}
}

// ParseResponse parses the complete API response and extracts device data.
func (p *DataParser) ParseResponse(response *DeviceDataResponse) ([]ParsedDeviceData, error) {
	if response == nil {
		return nil, ErrInvalidResponse
	}

	// Check response code
	if response.Code != "000000" {
		p.logger.Warn("API returned non-success code",
			zap.String("code", response.Code),
			zap.String("msg", response.Msg))
		return nil, fmt.Errorf("API error: code=%s, msg=%s", response.Code, response.Msg)
	}

	// Check if data is empty
	if len(response.Data) == 0 {
		p.logger.Debug("No device data in response")
		return []ParsedDeviceData{}, nil
	}

	// Parse each device
	result := make([]ParsedDeviceData, 0, len(response.Data))
	for i, deviceInfo := range response.Data {
		parsed, err := p.parseDeviceInfo(&deviceInfo)
		if err != nil {
			p.logger.Error("Failed to parse device data",
				zap.Int("index", i),
				zap.String("device_id", deviceInfo.AssetDevice.ID),
				zap.Error(err))
			// Continue parsing other devices even if one fails
			continue
		}
		result = append(result, *parsed)
	}

	p.logger.Info("Successfully parsed device data",
		zap.Int("total_devices", response.Total),
		zap.Int("parsed_devices", len(result)))

	return result, nil
}

// parseDeviceInfo parses a single DeviceInfo into ParsedDeviceData.
func (p *DataParser) parseDeviceInfo(deviceInfo *DeviceInfo) (*ParsedDeviceData, error) {
	if deviceInfo == nil {
		return nil, ErrInvalidDeviceData
	}

	parsed := &ParsedDeviceData{
		DeviceID:    deviceInfo.AssetDevice.ID,
		DeviceType:  deviceInfo.AssetDevice.DeviceType,
		Model:       deviceInfo.AssetDevice.Model,
		Alias:       deviceInfo.AssetDevice.Alias,
		Connected:   deviceInfo.Connected,
		CollectedAt: time.Now(),
	}

	// Parse realtime data
	if len(deviceInfo.Realtime) > 0 {
		realtime := p.parseRealtimeData(deviceInfo.Realtime)
		parsed.Realtime = realtime
	} else {
		p.logger.Warn("No realtime data available",
			zap.String("device_id", parsed.DeviceID))
	}

	return parsed, nil
}

// parseRealtimeData parses realtime data map into RealtimeData structure.
func (p *DataParser) parseRealtimeData(raw map[string]interface{}) RealtimeData {
	data := RealtimeData{
		Raw: raw, // Keep raw data for reference
	}

	// Parse power data (most important for energy calculation)
	data.LoadTotalWatt = p.parseFloat(raw, "loadTotalWatt", "load total watt")

	// Parse voltage data
	data.InputVolt1 = p.parseFloat(raw, "inputVolt1", "input volt 1")
	data.OutputVolt1 = p.parseFloat(raw, "outputVolt1", "output volt 1")
	data.BatVoltP = p.parseFloat(raw, "batVoltP", "battery volt percentage")

	// Parse current data
	data.OutputCurrent1 = p.parseFloat(raw, "outputCurrent1", "output current 1")

	// Parse frequency data
	data.InputFreq = p.parseFloat(raw, "inputFreq", "input frequency")
	data.OutputFreq = p.parseFloat(raw, "outputFreq", "output frequency")

	// Parse load data
	data.LoadPercent = p.parseFloat(raw, "loadPercent", "load percent")
	data.LoadTotalVa = p.parseFloat(raw, "loadTotalVa", "load total VA")
	data.LoadWatt1 = p.parseFloat(raw, "loadWatt1", "load watt 1")
	data.LoadVa1 = p.parseFloat(raw, "loadVa1", "load VA 1")

	// Parse battery data
	data.BatCapacity = p.parseFloat(raw, "batCapacity", "battery capacity")
	data.BatRemainTime = p.parseInt(raw, "batRemainTime", "battery remain time")
	data.IsCharging = p.parseBool(raw, "isCharging", "is charging")

	// Parse status data
	data.UpsTemperature = p.parseFloat(raw, "upsTemperature", "UPS temperature")
	data.Mode = p.parseString(raw, "mode", "mode")
	data.Status = p.parseString(raw, "status", "status")
	data.BatteryStatus = p.parseString(raw, "batteryStatus", "battery status")
	data.TestStatus = p.parseString(raw, "testStatus", "test status")
	data.FaultCode = p.parseString(raw, "faultCode", "fault code")

	return data
}

// parseFloat extracts and converts a string field to float64.
func (p *DataParser) parseFloat(raw map[string]interface{}, key, fieldName string) float64 {
	val, ok := raw[key]
	if !ok {
		p.logger.Debug("Field not found in raw data",
			zap.String("field", key))
		return 0
	}

	// Handle different value types
	switch v := val.(type) {
	case string:
		if v == "" {
			return 0
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			p.logger.Warn("Failed to parse float field",
				zap.String("field", key),
				zap.String("value", v),
				zap.Error(err))
			return 0
		}
		return f
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		p.logger.Warn("Unexpected type for float field",
			zap.String("field", key),
			zap.String("type", fmt.Sprintf("%T", v)))
		return 0
	}
}

// parseInt extracts and converts a string field to int.
func (p *DataParser) parseInt(raw map[string]interface{}, key, fieldName string) int {
	val, ok := raw[key]
	if !ok {
		p.logger.Debug("Field not found in raw data",
			zap.String("field", key))
		return 0
	}

	// Handle different value types
	switch v := val.(type) {
	case string:
		if v == "" {
			return 0
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			p.logger.Warn("Failed to parse int field",
				zap.String("field", key),
				zap.String("value", v),
				zap.Error(err))
			return 0
		}
		return i
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		p.logger.Warn("Unexpected type for int field",
			zap.String("field", key),
			zap.String("type", fmt.Sprintf("%T", v)))
		return 0
	}
}

// parseBool extracts and converts a string field to bool.
func (p *DataParser) parseBool(raw map[string]interface{}, key, fieldName string) bool {
	val, ok := raw[key]
	if !ok {
		p.logger.Debug("Field not found in raw data",
			zap.String("field", key))
		return false
	}

	// Handle different value types
	switch v := val.(type) {
	case string:
		return v == "1" || v == "true" || v == "True" || v == "TRUE"
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	default:
		p.logger.Warn("Unexpected type for bool field",
			zap.String("field", key),
			zap.String("type", fmt.Sprintf("%T", v)))
		return false
	}
}

// parseString extracts a string field.
func (p *DataParser) parseString(raw map[string]interface{}, key, fieldName string) string {
	val, ok := raw[key]
	if !ok {
		p.logger.Debug("Field not found in raw data",
			zap.String("field", key))
		return ""
	}

	// Handle different value types
	switch v := val.(type) {
	case string:
		return v
	case int, int64, float64, bool:
		return fmt.Sprintf("%v", v)
	default:
		p.logger.Warn("Unexpected type for string field",
			zap.String("field", key),
			zap.String("type", fmt.Sprintf("%T", v)))
		return ""
	}
}
