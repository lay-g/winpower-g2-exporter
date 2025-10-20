package storage

import (
	"math"
	"time"
)

// PowerData represents the power data structure for persistence.
// This structure is used to store cumulative energy data with timestamps.
type PowerData struct {
	// Timestamp is the Unix timestamp in milliseconds when the data was recorded
	Timestamp int64 `json:"timestamp" yaml:"timestamp"`

	// EnergyWH is the cumulative energy in watt-hours.
	// Can be negative to represent net energy (export - import).
	EnergyWH float64 `json:"energy_wh" yaml:"energy_wh"`
}

// NewPowerData creates a new PowerData instance with the current timestamp.
func NewPowerData(energyWH float64) *PowerData {
	return &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  energyWH,
	}
}

// NewPowerDataWithTimestamp creates a new PowerData instance with a specific timestamp.
func NewPowerDataWithTimestamp(energyWH float64, timestamp int64) *PowerData {
	return &PowerData{
		Timestamp: timestamp,
		EnergyWH:  energyWH,
	}
}

// IsZero checks if the PowerData represents a zero/uninitialized state.
// Note: PowerData with current timestamp and zero energy is not considered zero
// as it represents valid initialized data for a new device.
func (pd *PowerData) IsZero() bool {
	if pd == nil {
		return true
	}
	return pd.Timestamp == 0 && pd.EnergyWH == 0
}

// Validate checks if the PowerData contains valid values.
func (pd *PowerData) Validate() error {
	if pd == nil {
		return ErrInvalidData
	}

	if pd.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}

	// Check if energy value is finite (not NaN or Inf)
	if pd.EnergyWH != 0 && !isFinite(pd.EnergyWH) {
		return ErrInvalidEnergyValue
	}

	return nil
}

// isFinite checks if a float64 value is finite (not NaN or Inf).
func isFinite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}
