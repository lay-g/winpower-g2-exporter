package energy

import (
	"fmt"
	"math"
	"strings"
)

// Error types for the energy module
var (
	// ErrInvalidDeviceID is returned when device ID is empty or invalid
	ErrInvalidDeviceID = fmt.Errorf("device ID cannot be empty")

	// ErrInvalidPowerValue is returned when power value is invalid (NaN or Inf)
	ErrInvalidPowerValue = fmt.Errorf("power value must be a finite number")

	// ErrNegativePowerNotAllowed is returned when negative power is provided but not allowed
	ErrNegativePowerNotAllowed = fmt.Errorf("negative power is not allowed by configuration")

	// ErrCalculationTimeout is returned when calculation exceeds the configured timeout
	ErrCalculationTimeout = fmt.Errorf("energy calculation timed out")

	// ErrStorageOperation is returned when storage operation fails
	ErrStorageOperation = fmt.Errorf("storage operation failed")

	// ErrInvalidConfiguration is returned when configuration is invalid
	ErrInvalidConfiguration = fmt.Errorf("invalid energy module configuration")
)

// EnergyError represents an error with additional context for energy calculations
type EnergyError struct {
	Operation string        // Operation that failed (e.g., "Calculate", "Get")
	DeviceID  string        // Device ID involved in the operation
	Cause     error         // Underlying error
	Context   string        // Additional context
}

// Error implements the error interface
func (ee *EnergyError) Error() string {
	if ee.Context != "" {
		return fmt.Sprintf("energy %s operation failed for device '%s': %s (%s)",
			ee.Operation, ee.DeviceID, ee.Cause.Error(), ee.Context)
	}
	return fmt.Sprintf("energy %s operation failed for device '%s': %s",
		ee.Operation, ee.DeviceID, ee.Cause.Error())
}

// Unwrap returns the underlying cause
func (ee *EnergyError) Unwrap() error {
	return ee.Cause
}

// Is checks if the error matches the target
func (ee *EnergyError) Is(target error) bool {
	return ee.Cause == target
}

// NewEnergyError creates a new EnergyError with the specified parameters
func NewEnergyError(operation, deviceID, context string, cause error) *EnergyError {
	return &EnergyError{
		Operation: operation,
		DeviceID:  deviceID,
		Cause:     cause,
		Context:   context,
	}
}

// IsValidPower checks if a power value is valid for calculation
func IsValidPower(power float64) bool {
	return !math.IsNaN(power) && !math.IsInf(power, 0)
}

// ValidatePower validates power value against configuration
func ValidatePower(power float64, config *Config) error {
	if !IsValidPower(power) {
		return ErrInvalidPowerValue
	}

	if config != nil && !config.NegativePowerAllowed && power < 0 {
		return ErrNegativePowerNotAllowed
	}

	return nil
}

// ValidateDeviceID validates device ID
func ValidateDeviceID(deviceID string) error {
	if deviceID == "" || len(strings.TrimSpace(deviceID)) == 0 {
		return ErrInvalidDeviceID
	}
	return nil
}