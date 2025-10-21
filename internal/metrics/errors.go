package metrics

import "fmt"

// ErrInvalidConfig represents an invalid configuration error.
type ErrInvalidConfig struct {
	Field   string
	Reason  string
	Message string
}

// Error implements the error interface.
func (e ErrInvalidConfig) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("invalid config: field '%s' %s", e.Field, e.Reason)
}

// NewErrInvalidConfig creates a new ErrInvalidConfig.
func NewErrInvalidConfig(field, reason string) ErrInvalidConfig {
	return ErrInvalidConfig{
		Field:  field,
		Reason: reason,
	}
}

// ErrMetricNotInitialized represents a metric not initialized error.
type ErrMetricNotInitialized struct {
	MetricName string
}

// Error implements the error interface.
func (e ErrMetricNotInitialized) Error() string {
	return fmt.Sprintf("metric not initialized: %s", e.MetricName)
}

// NewErrMetricNotInitialized creates a new ErrMetricNotInitialized.
func NewErrMetricNotInitialized(metricName string) ErrMetricNotInitialized {
	return ErrMetricNotInitialized{
		MetricName: metricName,
	}
}

// ErrInvalidLabelValue represents an invalid label value error.
type ErrInvalidLabelValue struct {
	LabelName  string
	LabelValue string
	Reason     string
}

// Error implements the error interface.
func (e ErrInvalidLabelValue) Error() string {
	return fmt.Sprintf("invalid label value for '%s': '%s' - %s", e.LabelName, e.LabelValue, e.Reason)
}

// NewErrInvalidLabelValue creates a new ErrInvalidLabelValue.
func NewErrInvalidLabelValue(labelName, labelValue, reason string) ErrInvalidLabelValue {
	return ErrInvalidLabelValue{
		LabelName:  labelName,
		LabelValue: labelValue,
		Reason:     reason,
	}
}

// ErrMetricUpdate represents a metric update error.
type ErrMetricUpdate struct {
	MetricName string
	Reason     string
	Cause      error
}

// Error implements the error interface.
func (e ErrMetricUpdate) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("metric update failed for '%s': %s: %v", e.MetricName, e.Reason, e.Cause)
	}
	return fmt.Sprintf("metric update failed for '%s': %s", e.MetricName, e.Reason)
}

// Unwrap returns the underlying cause.
func (e ErrMetricUpdate) Unwrap() error {
	return e.Cause
}

// NewErrMetricUpdate creates a new ErrMetricUpdate.
func NewErrMetricUpdate(metricName, reason string, cause error) ErrMetricUpdate {
	return ErrMetricUpdate{
		MetricName: metricName,
		Reason:     reason,
		Cause:      cause,
	}
}

// ErrNotInitialized represents a not initialized error.
type ErrNotInitialized struct {
	Component string
}

// Error implements the error interface.
func (e ErrNotInitialized) Error() string {
	return fmt.Sprintf("component not initialized: %s", e.Component)
}

// NewErrNotInitialized creates a new ErrNotInitialized.
func NewErrNotInitialized(component string) ErrNotInitialized {
	return ErrNotInitialized{
		Component: component,
	}
}
