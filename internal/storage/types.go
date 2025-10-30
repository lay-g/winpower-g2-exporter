package storage

// PowerData represents power information with timestamp and energy fields.
//
// This structure stores accumulated energy data for a device at a specific
// point in time. It is the primary data structure used by the storage module
// for persistence.
//
// Fields:
//   - Timestamp: Unix timestamp in milliseconds when the data was recorded
//   - EnergyWH: Accumulated energy in watt-hours (non-negative)
//
// The data is validated before storage to ensure:
//   - Timestamp is valid and not too far in the future
//   - Energy value is finite and non-negative
//
// Example:
//
//	data := &storage.PowerData{
//	    Timestamp: time.Now().UnixMilli(),
//	    EnergyWH:  1234.5,
//	}
//	if err := data.Validate(); err != nil {
//	    log.Printf("invalid data: %v", err)
//	}
type PowerData struct {
	// Timestamp is the Unix timestamp in milliseconds when the data was recorded
	Timestamp int64 `json:"timestamp"`

	// EnergyWH is the accumulated energy in watt-hours
	EnergyWH float64 `json:"energy_wh"`
}
