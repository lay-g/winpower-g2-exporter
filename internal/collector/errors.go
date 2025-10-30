package collector

import (
	"errors"
	"fmt"
	"time"
)

// Error types for collector module
var (
	// ErrWinPowerCollection indicates failure in WinPower data collection
	ErrWinPowerCollection = errors.New("winpower data collection failed")

	// ErrEnergyCalculation indicates failure in energy calculation
	ErrEnergyCalculation = errors.New("energy calculation failed")

	// ErrDataConversion indicates failure in data conversion
	ErrDataConversion = errors.New("data conversion failed")

	// ErrInvalidContext indicates invalid or nil context
	ErrInvalidContext = errors.New("invalid context")

	// ErrNilDependency indicates nil dependency injection
	ErrNilDependency = errors.New("nil dependency provided")
)

// ErrorType represents the classification of errors
type ErrorType int

const (
	// ErrorTypeWinPowerCollection for WinPower collection errors
	ErrorTypeWinPowerCollection ErrorType = iota
	// ErrorTypeEnergyCalculation for energy calculation errors
	ErrorTypeEnergyCalculation
	// ErrorTypeDataConversion for data conversion errors
	ErrorTypeDataConversion
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeWinPowerCollection:
		return "WinPowerCollection"
	case ErrorTypeEnergyCalculation:
		return "EnergyCalculation"
	case ErrorTypeDataConversion:
		return "DataConversion"
	default:
		return "Unknown"
	}
}

// CollectionError represents a detailed error during collection
type CollectionError struct {
	Type      ErrorType `json:"type"`
	DeviceID  string    `json:"device_id,omitempty"`
	Message   string    `json:"message"`
	Cause     error     `json:"-"`
	Timestamp time.Time `json:"timestamp"`
}

// Error implements the error interface
func (e *CollectionError) Error() string {
	if e.DeviceID != "" {
		return fmt.Sprintf("[%s] Device %s: %s", e.Type, e.DeviceID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *CollectionError) Unwrap() error {
	return e.Cause
}

// NewCollectionError creates a new CollectionError
func NewCollectionError(errType ErrorType, deviceID, message string, cause error) *CollectionError {
	return &CollectionError{
		Type:      errType,
		DeviceID:  deviceID,
		Message:   message,
		Cause:     cause,
		Timestamp: time.Now(),
	}
}
