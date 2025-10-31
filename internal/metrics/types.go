package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

// MetricsService manages Prometheus metrics and provides HTTP handler for /metrics endpoint
type MetricsService struct {
	registry     *prometheus.Registry
	collector    collector.CollectorInterface
	logger       log.Logger
	winpowerHost string // Configuration value for WinPower host label

	// Exporter self-monitoring metrics
	exporterUp                prometheus.Gauge
	requestsTotal             *prometheus.CounterVec
	requestDuration           *prometheus.HistogramVec
	collectionDuration        *prometheus.HistogramVec
	scrapeErrorsTotal         *prometheus.CounterVec
	tokenRefreshTotal         *prometheus.CounterVec
	deviceCount               prometheus.Gauge
	memoryBytes               *prometheus.GaugeVec
	lastCollectionTimeSeconds prometheus.Gauge

	// WinPower connection/auth metrics
	connectionStatus   prometheus.Gauge
	authStatus         prometheus.Gauge
	apiResponseTime    *prometheus.HistogramVec
	tokenExpirySeconds prometheus.Gauge
	tokenValid         prometheus.Gauge

	// Device metrics - dynamically created per device
	deviceMetrics map[string]*DeviceMetrics
	mu            sync.RWMutex // Protects deviceMetrics map
}

// DeviceMetrics holds all Prometheus metrics for a single device
type DeviceMetrics struct {
	// Device status
	connected           prometheus.Gauge
	lastUpdateTimestamp prometheus.Gauge

	// Electrical parameters - Input
	inputVoltage   prometheus.Gauge
	inputFrequency prometheus.Gauge

	// Electrical parameters - Output
	outputVoltage     prometheus.Gauge
	outputCurrent     prometheus.Gauge
	outputFrequency   prometheus.Gauge
	outputVoltageType prometheus.Gauge

	// Load and power
	loadPercent    prometheus.Gauge
	loadTotalWatt  prometheus.Gauge // Core metric for energy calculation
	loadTotalVa    prometheus.Gauge
	loadWattPhase1 prometheus.Gauge
	loadVaPhase1   prometheus.Gauge
	powerWatts     prometheus.Gauge // Instantaneous power (same as LoadTotalWatt)

	// Battery parameters
	batteryCharging       prometheus.Gauge
	batteryVoltagePercent prometheus.Gauge
	batteryCapacity       prometheus.Gauge
	batteryRemainSeconds  prometheus.Gauge
	batteryStatus         prometheus.Gauge

	// UPS status
	upsTemperature prometheus.Gauge
	upsMode        prometheus.Gauge
	upsStatus      prometheus.Gauge
	upsTestStatus  prometheus.Gauge
	upsFaultCode   *prometheus.GaugeVec // Has fault_code label

	// Energy
	cumulativeEnergy prometheus.Gauge
}

// MetricsConfig holds configuration for the metrics service
type MetricsConfig struct {
	// Namespace is the Prometheus namespace for all metrics (default: "winpower")
	Namespace string

	// Subsystem is the Prometheus subsystem for exporter metrics (default: "exporter")
	Subsystem string

	// WinPowerHost is the label value for winpower_host
	WinPowerHost string

	// EnableMemoryMetrics enables memory usage monitoring
	EnableMemoryMetrics bool
}

// DefaultMetricsConfig returns default configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Namespace:           "winpower",
		Subsystem:           "exporter",
		WinPowerHost:        "localhost",
		EnableMemoryMetrics: true,
	}
}
