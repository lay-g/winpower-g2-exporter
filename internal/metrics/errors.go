package metrics

import (
	"errors"
)

var (
	// ErrCollectorNil is returned when the collector interface is nil
	ErrCollectorNil = errors.New("collector interface cannot be nil")

	// ErrLoggerNil is returned when the logger is nil
	ErrLoggerNil = errors.New("logger cannot be nil")

	// ErrCollectionFailed is returned when data collection fails
	ErrCollectionFailed = errors.New("failed to collect device data")

	// ErrMetricsUpdateFailed is returned when metrics update fails
	ErrMetricsUpdateFailed = errors.New("failed to update metrics")

	// ErrDeviceNotFound is returned when a device is not found in the collection result
	ErrDeviceNotFound = errors.New("device not found in collection result")

	// ErrInvalidCollectionResult is returned when the collection result is invalid
	ErrInvalidCollectionResult = errors.New("invalid collection result")
)
