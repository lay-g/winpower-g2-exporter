package collector

import (
	"context"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// CollectorInterface defines the interface for data collection and coordination
type CollectorInterface interface {
	// CollectDeviceData collects device data and triggers energy calculations
	CollectDeviceData(ctx context.Context) (*CollectionResult, error)
}

// WinPowerClient defines the interface for WinPower data collection.
// Note: We define this interface in the collector package to follow the Dependency Inversion Principle.
// The collector module defines what it needs, rather than depending on the concrete winpower implementation.
// This makes the collector more testable and flexible.
//
// Alternative: We could directly use winpower.WinPowerClient, which would reduce duplication
// but would create a tighter coupling to the winpower module's interface definition.
type WinPowerClient interface {
	CollectDeviceData(ctx context.Context) ([]winpower.ParsedDeviceData, error)
	GetConnectionStatus() bool
	GetLastCollectionTime() time.Time
}

// EnergyCalculator defines the interface for energy calculation.
// Following the same principle as WinPowerClient, this interface is defined here
// to ensure the collector controls its own dependency contracts.
type EnergyCalculator interface {
	// Calculate calculates cumulative energy for a device
	Calculate(deviceID string, power float64) (float64, error)
	// Get retrieves the latest energy value for a device
	Get(deviceID string) (float64, error)
}
