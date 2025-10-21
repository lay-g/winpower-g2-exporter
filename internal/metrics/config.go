package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// MetricManagerConfig represents the configuration for the metrics module.
type MetricManagerConfig struct {
	// Namespace specifies the metrics namespace prefix.
	// Default: "winpower"
	Namespace string `json:"namespace" yaml:"namespace" default:"winpower"`

	// Subsystem specifies the metrics subsystem prefix.
	// Default: "exporter"
	Subsystem string `json:"subsystem" yaml:"subsystem" default:"exporter"`

	// Registry allows injection of a custom Prometheus registry.
	// If nil, a new registry will be created.
	Registry *prometheus.Registry `json:"-" yaml:"-"`

	// RequestDurationBuckets defines the histogram buckets for HTTP request duration.
	// Default: [0.05, 0.1, 0.2, 0.5, 1, 2, 5] seconds
	RequestDurationBuckets []float64 `json:"request_duration_buckets" yaml:"request_duration_buckets"`

	// CollectionDurationBuckets defines the histogram buckets for data collection duration.
	// Default: [0.1, 0.2, 0.5, 1, 2, 5, 10] seconds
	CollectionDurationBuckets []float64 `json:"collection_duration_buckets" yaml:"collection_duration_buckets"`

	// APIResponseBuckets defines the histogram buckets for API response time.
	// Default: [0.05, 0.1, 0.2, 0.5, 1] seconds
	APIResponseBuckets []float64 `json:"api_response_buckets" yaml:"api_response_buckets"`
}

// DefaultConfig returns a default configuration for the metrics module.
func DefaultConfig() *MetricManagerConfig {
	return &MetricManagerConfig{
		Namespace:                 "winpower",
		Subsystem:                 "exporter",
		Registry:                  nil,
		RequestDurationBuckets:    []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5},
		CollectionDurationBuckets: []float64{0.1, 0.2, 0.5, 1, 2, 5, 10},
		APIResponseBuckets:        []float64{0.05, 0.1, 0.2, 0.5, 1},
	}
}

// Validate validates the configuration and returns an error if invalid.
func (c *MetricManagerConfig) Validate() error {
	if c.Namespace == "" {
		return ErrInvalidConfig{Field: "namespace", Reason: "cannot be empty"}
	}
	if c.Subsystem == "" {
		return ErrInvalidConfig{Field: "subsystem", Reason: "cannot be empty"}
	}
	if len(c.RequestDurationBuckets) == 0 {
		return ErrInvalidConfig{Field: "request_duration_buckets", Reason: "cannot be empty"}
	}
	if len(c.CollectionDurationBuckets) == 0 {
		return ErrInvalidConfig{Field: "collection_duration_buckets", Reason: "cannot be empty"}
	}
	if len(c.APIResponseBuckets) == 0 {
		return ErrInvalidConfig{Field: "api_response_buckets", Reason: "cannot be empty"}
	}
	return nil
}
