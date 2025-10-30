package storage

import (
	"fmt"
	"math"
	"time"
)

// Validate checks if PowerData is valid.
//
// Validation rules:
//   - PowerData must not be nil
//   - Timestamp must be non-negative (0 or positive)
//   - Timestamp must not be more than 24 hours in the future
//   - EnergyWH must be a finite number (not NaN or Inf)
//   - EnergyWH must be non-negative
//
// Returns an error describing the first validation failure encountered,
// or nil if all validations pass.
//
// This method is called automatically by Write() before storing data,
// but can also be called explicitly for validation without storage.
//
// Example:
//
//	data := &storage.PowerData{
//	    Timestamp: time.Now().UnixMilli(),
//	    EnergyWH:  1234.5,
//	}
//	if err := data.Validate(); err != nil {
//	    log.Printf("invalid data: %v", err)
//	    return
//	}
func (d *PowerData) Validate() error {
	if d == nil {
		return fmt.Errorf("%w: PowerData cannot be nil", ErrInvalidData)
	}

	// Validate timestamp - should be positive and not in the far future
	if d.Timestamp < 0 {
		return fmt.Errorf("%w: timestamp cannot be negative", ErrInvalidData)
	}

	// Check if timestamp is too far in the future (more than 1 day)
	now := time.Now().UnixMilli()
	oneDayInMs := int64(24 * 60 * 60 * 1000)
	if d.Timestamp > now+oneDayInMs {
		return fmt.Errorf("%w: timestamp is too far in the future", ErrInvalidData)
	}

	// Validate energy - should be finite and non-negative
	if math.IsNaN(d.EnergyWH) || math.IsInf(d.EnergyWH, 0) {
		return fmt.Errorf("%w: energy value must be finite", ErrInvalidData)
	}

	if d.EnergyWH < 0 {
		return fmt.Errorf("%w: energy value cannot be negative", ErrInvalidData)
	}

	return nil
}
