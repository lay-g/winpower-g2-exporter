package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Metric namespaces and subsystems
	namespace = "winpower"
	subsystem = "exporter"

	// Common label names
	labelWinPowerHost = "winpower_host"
	labelDeviceID     = "device_id"
	labelDeviceName   = "device_name"
	labelDeviceType   = "device_type"
	labelFaultCode    = "fault_code"
	labelMemoryType   = "type"
	labelErrorType    = "error_type"
)

var (
	// Default histogram buckets for duration metrics
	durationBuckets = []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5}
	// API response time buckets (shorter range)
	apiResponseBuckets = []float64{0.05, 0.1, 0.2, 0.5, 1}
)

// initExporterMetrics initializes exporter self-monitoring metrics
func (m *MetricsService) initExporterMetrics(config *MetricsConfig) {
	labels := prometheus.Labels{labelWinPowerHost: config.WinPowerHost}

	m.exporterUp = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "up",
		Help:        "Whether the WinPower exporter is running (1 = up, 0 = down)",
		ConstLabels: labels,
	})

	m.requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "requests_total",
		Help:        "Total number of HTTP requests to the /metrics endpoint",
		ConstLabels: labels,
	}, []string{})

	m.requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "request_duration_seconds",
		Help:        "HTTP request duration in seconds",
		Buckets:     durationBuckets,
		ConstLabels: labels,
	}, []string{})

	m.collectionDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "collection_duration_seconds",
		Help:        "Data collection and calculation duration in seconds",
		Buckets:     durationBuckets,
		ConstLabels: labels,
	}, []string{})

	m.scrapeErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "scrape_errors_total",
		Help:        "Total number of data collection errors",
		ConstLabels: labels,
	}, []string{labelErrorType})

	m.tokenRefreshTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "token_refresh_total",
		Help:        "Total number of token refreshes",
		ConstLabels: labels,
	}, []string{})

	m.deviceCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "device_count",
		Help:        "Number of discovered devices",
		ConstLabels: labels,
	})

	m.lastCollectionTimeSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        "last_collection_time_seconds",
		Help:        "Unix timestamp of the last successful collection",
		ConstLabels: labels,
	})

	if config.EnableMemoryMetrics {
		m.memoryBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "memory_bytes",
			Help:        "Memory usage in bytes",
			ConstLabels: labels,
		}, []string{labelMemoryType})
	}
}

// initConnectionMetrics initializes WinPower connection and authentication metrics
func (m *MetricsService) initConnectionMetrics(config *MetricsConfig) {
	labels := prometheus.Labels{labelWinPowerHost: config.WinPowerHost}

	m.connectionStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "connection_status",
		Help:        "WinPower connection status (1 = connected, 0 = disconnected)",
		ConstLabels: labels,
	})

	m.authStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "auth_status",
		Help:        "WinPower authentication status (1 = authenticated, 0 = not authenticated)",
		ConstLabels: labels,
	})

	m.apiResponseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   namespace,
		Name:        "api_response_time_seconds",
		Help:        "WinPower API response time in seconds",
		Buckets:     apiResponseBuckets,
		ConstLabels: labels,
	}, []string{})

	m.tokenExpirySeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "token_expiry_seconds",
		Help:        "Remaining time until token expiry in seconds",
		ConstLabels: labels,
	})

	m.tokenValid = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Name:        "token_valid",
		Help:        "Whether the current token is valid (1 = valid, 0 = invalid)",
		ConstLabels: labels,
	})
}

// registerMetrics registers all metrics with the Prometheus registry
func (m *MetricsService) registerMetrics() {
	// Register exporter metrics
	m.registry.MustRegister(m.exporterUp)
	m.registry.MustRegister(m.requestsTotal)
	m.registry.MustRegister(m.requestDuration)
	m.registry.MustRegister(m.collectionDuration)
	m.registry.MustRegister(m.scrapeErrorsTotal)
	m.registry.MustRegister(m.tokenRefreshTotal)
	m.registry.MustRegister(m.deviceCount)
	m.registry.MustRegister(m.lastCollectionTimeSeconds)

	if m.memoryBytes != nil {
		m.registry.MustRegister(m.memoryBytes)
	}

	// Register connection metrics
	m.registry.MustRegister(m.connectionStatus)
	m.registry.MustRegister(m.authStatus)
	m.registry.MustRegister(m.apiResponseTime)
	m.registry.MustRegister(m.tokenExpirySeconds)
	m.registry.MustRegister(m.tokenValid)

	// Set exporter up to 1 on initialization
	m.exporterUp.Set(1)
}

// createDeviceMetrics creates a new DeviceMetrics instance for a device
func (m *MetricsService) createDeviceMetrics(deviceID, deviceName, deviceType, winpowerHost string) *DeviceMetrics {
	labels := prometheus.Labels{
		labelWinPowerHost: winpowerHost,
		labelDeviceID:     deviceID,
		labelDeviceName:   deviceName,
		labelDeviceType:   deviceType,
	}

	dm := &DeviceMetrics{
		// Device status
		connected: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_connected",
			Help:        "Device connection status (1 = connected, 0 = disconnected)",
			ConstLabels: labels,
		}),
		lastUpdateTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_last_update_timestamp",
			Help:        "Unix timestamp of the last device update",
			ConstLabels: labels,
		}),

		// Input electrical parameters
		inputVoltage: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_input_voltage",
			Help:        "Input voltage in volts",
			ConstLabels: labels,
		}),
		inputFrequency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_input_frequency",
			Help:        "Input frequency in hertz",
			ConstLabels: labels,
		}),

		// Output electrical parameters
		outputVoltage: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_output_voltage",
			Help:        "Output voltage in volts",
			ConstLabels: labels,
		}),
		outputCurrent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_output_current",
			Help:        "Output current in amperes",
			ConstLabels: labels,
		}),
		outputFrequency: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_output_frequency",
			Help:        "Output frequency in hertz",
			ConstLabels: labels,
		}),
		outputVoltageType: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_output_voltage_type",
			Help:        "Output voltage type (encoded as numeric value)",
			ConstLabels: labels,
		}),

		// Load and power
		loadPercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_load_percent",
			Help:        "Device load percentage",
			ConstLabels: labels,
		}),
		loadTotalWatt: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_load_total_watts",
			Help:        "Total load active power in watts (core metric for energy calculation)",
			ConstLabels: labels,
		}),
		loadTotalVa: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_load_total_va",
			Help:        "Total load apparent power in volt-amperes",
			ConstLabels: labels,
		}),
		loadWattPhase1: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_load_watts_phase1",
			Help:        "Phase 1 active power in watts",
			ConstLabels: labels,
		}),
		loadVaPhase1: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_load_va_phase1",
			Help:        "Phase 1 apparent power in volt-amperes",
			ConstLabels: labels,
		}),
		powerWatts: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "power_watts",
			Help:        "Instantaneous power in watts (same as load_total_watts)",
			ConstLabels: labels,
		}),

		// Battery parameters
		batteryCharging: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_battery_charging",
			Help:        "Battery charging status (1 = charging, 0 = not charging)",
			ConstLabels: labels,
		}),
		batteryVoltagePercent: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_battery_voltage_percent",
			Help:        "Battery voltage percentage",
			ConstLabels: labels,
		}),
		batteryCapacity: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_battery_capacity",
			Help:        "Battery capacity percentage",
			ConstLabels: labels,
		}),
		batteryRemainSeconds: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_battery_remain_seconds",
			Help:        "Battery remaining time in seconds",
			ConstLabels: labels,
		}),
		batteryStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_battery_status",
			Help:        "Battery status code (encoded as numeric value)",
			ConstLabels: labels,
		}),

		// UPS status
		upsTemperature: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_ups_temperature",
			Help:        "UPS temperature in Celsius",
			ConstLabels: labels,
		}),
		upsMode: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_ups_mode",
			Help:        "UPS operating mode (encoded as numeric value)",
			ConstLabels: labels,
		}),
		upsStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_ups_status",
			Help:        "UPS status code (encoded as numeric value)",
			ConstLabels: labels,
		}),
		upsTestStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_ups_test_status",
			Help:        "UPS test status code (encoded as numeric value)",
			ConstLabels: labels,
		}),
		upsFaultCode: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_ups_fault_code",
			Help:        "UPS fault code (with fault_code label for aggregation)",
			ConstLabels: labels,
		}, []string{labelFaultCode}),

		// Energy
		cumulativeEnergy: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "device_cumulative_energy",
			Help:        "Cumulative energy consumption in watt-hours",
			ConstLabels: labels,
		}),
	}

	// Register all device metrics
	m.registry.MustRegister(dm.connected)
	m.registry.MustRegister(dm.lastUpdateTimestamp)
	m.registry.MustRegister(dm.inputVoltage)
	m.registry.MustRegister(dm.inputFrequency)
	m.registry.MustRegister(dm.outputVoltage)
	m.registry.MustRegister(dm.outputCurrent)
	m.registry.MustRegister(dm.outputFrequency)
	m.registry.MustRegister(dm.outputVoltageType)
	m.registry.MustRegister(dm.loadPercent)
	m.registry.MustRegister(dm.loadTotalWatt)
	m.registry.MustRegister(dm.loadTotalVa)
	m.registry.MustRegister(dm.loadWattPhase1)
	m.registry.MustRegister(dm.loadVaPhase1)
	m.registry.MustRegister(dm.powerWatts)
	m.registry.MustRegister(dm.batteryCharging)
	m.registry.MustRegister(dm.batteryVoltagePercent)
	m.registry.MustRegister(dm.batteryCapacity)
	m.registry.MustRegister(dm.batteryRemainSeconds)
	m.registry.MustRegister(dm.batteryStatus)
	m.registry.MustRegister(dm.upsTemperature)
	m.registry.MustRegister(dm.upsMode)
	m.registry.MustRegister(dm.upsStatus)
	m.registry.MustRegister(dm.upsTestStatus)
	m.registry.MustRegister(dm.upsFaultCode)
	m.registry.MustRegister(dm.cumulativeEnergy)

	return dm
}
