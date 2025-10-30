package winpower

import (
	"fmt"

	"go.uber.org/zap"
)

// DataValidator validates parsed device data for integrity and reasonableness.
type DataValidator struct {
	logger *zap.Logger
}

// NewDataValidator creates a new DataValidator instance.
func NewDataValidator(logger *zap.Logger) *DataValidator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DataValidator{
		logger: logger,
	}
}

// ValidationResult represents the result of data validation.
type ValidationResult struct {
	IsValid bool
	Errors  []ValidationError
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Validate validates a ParsedDeviceData structure.
func (v *DataValidator) Validate(data *ParsedDeviceData) *ValidationResult {
	result := &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
	}

	if data == nil {
		result.IsValid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "data",
			Message: "device data is nil",
		})
		return result
	}

	// Validate required fields
	v.validateRequiredFields(data, result)

	// Validate realtime data
	v.validateRealtimeData(&data.Realtime, result)

	// Update IsValid based on errors
	result.IsValid = len(result.Errors) == 0

	if !result.IsValid {
		v.logger.Warn("Device data validation failed",
			zap.String("device_id", data.DeviceID),
			zap.Int("error_count", len(result.Errors)))
	}

	return result
}

// validateRequiredFields validates required device fields.
func (v *DataValidator) validateRequiredFields(data *ParsedDeviceData, result *ValidationResult) {
	// DeviceID is required
	if data.DeviceID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "device_id",
			Value:   data.DeviceID,
			Message: "device_id cannot be empty",
		})
	}

	// DeviceType should be valid
	if data.DeviceType < 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "device_type",
			Value:   data.DeviceType,
			Message: "device_type must be non-negative",
		})
	}
}

// validateRealtimeData validates realtime data fields.
func (v *DataValidator) validateRealtimeData(data *RealtimeData, result *ValidationResult) {
	// Validate power values (cannot be negative)
	if data.LoadTotalWatt < 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "load_total_watt",
			Value:   data.LoadTotalWatt,
			Message: "load_total_watt cannot be negative",
		})
	}

	if data.LoadWatt1 < 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "load_watt1",
			Value:   data.LoadWatt1,
			Message: "load_watt1 cannot be negative",
		})
	}

	// Validate voltage values (reasonable range: 0-500V)
	v.validateVoltageRange("input_volt_1", data.InputVolt1, result)
	v.validateVoltageRange("output_volt_1", data.OutputVolt1, result)
	v.validateVoltageRange("bat_volt_p", data.BatVoltP, result)

	// Validate current values (cannot be negative)
	if data.OutputCurrent1 < 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "output_current_1",
			Value:   data.OutputCurrent1,
			Message: "output_current_1 cannot be negative",
		})
	}

	// Validate frequency values (reasonable range: 45-65Hz for power systems)
	v.validateFrequencyRange("input_freq", data.InputFreq, result)
	v.validateFrequencyRange("output_freq", data.OutputFreq, result)

	// Validate percentage values (0-100%)
	v.validatePercentageRange("load_percent", data.LoadPercent, result)
	v.validatePercentageRange("bat_capacity", data.BatCapacity, result)

	// Validate temperature (reasonable range: -20 to 100°C)
	if data.UpsTemperature < -20 || data.UpsTemperature > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "ups_temperature",
			Value:   data.UpsTemperature,
			Message: fmt.Sprintf("ups_temperature out of reasonable range (-20 to 100°C): %v", data.UpsTemperature),
		})
	}

	// Validate battery remain time (cannot be negative)
	if data.BatRemainTime < 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "bat_remain_time",
			Value:   data.BatRemainTime,
			Message: "bat_remain_time cannot be negative",
		})
	}
}

// validateVoltageRange validates voltage values are in reasonable range.
func (v *DataValidator) validateVoltageRange(field string, value float64, result *ValidationResult) {
	if value < 0 || value > 500 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Value:   value,
			Message: fmt.Sprintf("%s out of reasonable range (0-500V): %v", field, value),
		})
	}
}

// validateFrequencyRange validates frequency values are in reasonable range.
func (v *DataValidator) validateFrequencyRange(field string, value float64, result *ValidationResult) {
	// Allow zero for missing data, but if non-zero, check range
	if value != 0 && (value < 45 || value > 65) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Value:   value,
			Message: fmt.Sprintf("%s out of reasonable range (45-65Hz): %v", field, value),
		})
	}
}

// validatePercentageRange validates percentage values are in 0-100% range.
func (v *DataValidator) validatePercentageRange(field string, value float64, result *ValidationResult) {
	if value < 0 || value > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Value:   value,
			Message: fmt.Sprintf("%s out of valid range (0-100%%): %v", field, value),
		})
	}
}

// ValidateBatch validates a batch of device data.
func (v *DataValidator) ValidateBatch(dataList []ParsedDeviceData) []*ValidationResult {
	results := make([]*ValidationResult, len(dataList))
	for i := range dataList {
		results[i] = v.Validate(&dataList[i])
	}
	return results
}

// HasCriticalErrors checks if validation result contains critical errors.
// Critical errors are those that would prevent further processing.
func (v *DataValidator) HasCriticalErrors(result *ValidationResult) bool {
	if result == nil || result.IsValid {
		return false
	}

	// Check for critical error fields
	for _, err := range result.Errors {
		switch err.Field {
		case "data", "device_id":
			// These are critical - can't proceed without them
			return true
		}
	}

	return false
}
