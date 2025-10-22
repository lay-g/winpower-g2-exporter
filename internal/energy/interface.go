package energy

import (
	"sync"

	"go.uber.org/zap"
)

// EnergyInterface defines the interface for energy calculation operations.
// This interface provides abstraction for power-to-energy conversion and data retrieval.
type EnergyInterface interface {
	// Calculate performs energy accumulation calculation based on power readings.
	// It loads historical data, calculates interval energy, and updates the accumulated value.
	// Returns the new total accumulated energy in watt-hours (Wh).
	Calculate(deviceID string, power float64) (float64, error)

	// Get retrieves the latest accumulated energy data for a device.
	// Returns the current total energy in watt-hours (Wh).
	Get(deviceID string) (float64, error)
}

// EnergyService implements the EnergyInterface with a simple single-threaded architecture.
// All calculations are executed serially through a global lock to ensure data consistency.
type EnergyService struct {
	storage StorageManager // Storage interface for data persistence
	logger  *zap.Logger    // Structured logger
	config  *Config        // Energy module configuration
	mutex   sync.RWMutex   // Global lock for serial execution
	stats   *SimpleStats   // Simple statistics tracking
}

// StorageManager defines the storage interface that energy module depends on.
// This is imported from the storage module.
type StorageManager interface {
	// Write persists power data for a specific device.
	Write(deviceID string, data *PowerData) error

	// Read retrieves power data for a specific device.
	Read(deviceID string) (*PowerData, error)
}

// PowerData represents the power data structure from storage module.
// This is imported from the storage module.
type PowerData struct {
	Timestamp int64   `json:"timestamp"`
	EnergyWH  float64 `json:"energy_wh"`
}

// IsZero checks if the PowerData represents a zero/uninitialized state.
func (pd *PowerData) IsZero() bool {
	if pd == nil {
		return true
	}
	return pd.Timestamp == 0 && pd.EnergyWH == 0
}

// Config represents the configuration for the energy module.
type Config struct {
	// Precision specifies the calculation precision for energy values in watt-hours.
	// Default: 0.01 Wh
	Precision float64 `yaml:"precision" json:"precision" env:"ENERGY_PRECISION" default:"0.01"`

	// EnableStats indicates whether to collect calculation statistics.
	// Default: true
	EnableStats bool `yaml:"enable_stats" json:"enable_stats" env:"ENERGY_ENABLE_STATS" default:"true"`

	// MaxCalculationTime specifies the maximum allowed time for a single calculation.
	// Default: 1 second
	MaxCalculationTime int64 `yaml:"max_calculation_time" json:"max_calculation_time" env:"ENERGY_MAX_CALCULATION_TIME" default:"1000000000"` // in nanoseconds

	// NegativePowerAllowed indicates whether negative power values are accepted.
	// Default: true (to support energy feedback scenarios)
	NegativePowerAllowed bool `yaml:"negative_power_allowed" json:"negative_power_allowed" env:"ENERGY_NEGATIVE_POWER_ALLOWED" default:"true"`
}

// SimpleStats tracks basic calculation statistics.
type SimpleStats struct {
	// TotalCalculations is the total number of calculations performed.
	TotalCalculations int64 `json:"total_calculations"`

	// TotalErrors is the total number of calculation errors.
	TotalErrors int64 `json:"total_errors"`

	// LastUpdateTime is the timestamp of the last calculation.
	LastUpdateTime int64 `json:"last_update_time"`

	// AvgCalculationTime is the average calculation time in nanoseconds.
	AvgCalculationTime int64 `json:"avg_calculation_time"`

	// mutex protects statistics updates.
	mutex sync.RWMutex
}
