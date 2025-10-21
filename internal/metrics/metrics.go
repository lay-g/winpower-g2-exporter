package metrics

import (
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// NewMetricManager creates a new MetricManager instance with the given configuration and logger.
func NewMetricManager(config *MetricManagerConfig, logger *zap.Logger) (*MetricManager, error) {
	if config == nil {
		return nil, NewErrInvalidConfig("config", "cannot be nil")
	}
	if logger == nil {
		return nil, NewErrInvalidConfig("logger", "cannot be nil")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create or use provided registry
	registry := config.Registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	manager := &MetricManager{
		registry: registry,
		logger:   logger,
		config:   *config, // Make a copy
		logLevel: "info",  // Default log level
	}

	// Initialize metrics
	if err := manager.initializeMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	manager.initialized = true
	logger.Info("MetricManager initialized successfully",
		zap.String("namespace", config.Namespace),
		zap.String("subsystem", config.Subsystem))

	return manager, nil
}

// initializeMetrics creates and registers all Prometheus metrics.
func (m *MetricManager) initializeMetrics() error {
	namespace := m.config.Namespace
	subsystem := m.config.Subsystem

	// Initialize Exporter self-monitoring metrics
	m.exporterMetrics = ExporterMetrics{
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "up",
			Help:      "Indicates if the WinPower exporter is up and running.",
		}),
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests handled by the exporter.",
		}, []string{"winpower_host", "method", "status_code"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests handled by the exporter.",
			Buckets:   m.config.RequestDurationBuckets,
		}, []string{"winpower_host", "method", "status_code"}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "scrape_errors_total",
			Help:      "Total number of scrape errors encountered.",
		}, []string{"winpower_host", "error_type"}),
		collectionTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "collection_duration_seconds",
			Help:      "Time taken to collect and process data from WinPower.",
			Buckets:   m.config.CollectionDurationBuckets,
		}, []string{"winpower_host", "status"}),
		tokenRefresh: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "token_refresh_total",
			Help:      "Total number of token refresh attempts.",
		}, []string{"winpower_host", "result"}),
		deviceCount: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "device_count",
			Help:      "Number of devices discovered and monitored.",
		}, []string{"winpower_host", "device_type"}),
	}

	// Initialize WinPower connection/authentication metrics
	m.connectionMetrics = ConnectionMetrics{
		connectionStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "connection_status",
			Help:      "Status of connection to WinPower service (1=connected, 0=disconnected).",
		}, []string{"winpower_host", "connection_type"}),
		authStatus: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "auth_status",
			Help:      "Authentication status with WinPower service (1=authenticated, 0=not authenticated).",
		}, []string{"winpower_host", "auth_method"}),
		apiResponseTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "api_response_time_seconds",
			Help:      "Response time for WinPower API calls.",
			Buckets:   m.config.APIResponseBuckets,
		}, []string{"winpower_host", "api_endpoint"}),
		tokenExpiry: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "token_expiry_seconds",
			Help:      "Time remaining until token expiry in seconds.",
		}, []string{"winpower_host", "user_id"}),
		tokenValid: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "token_valid",
			Help:      "Indicates if the current token is valid (1=valid, 0=invalid).",
		}, []string{"winpower_host", "user_id"}),
	}

	// Initialize device/power supply metrics
	m.deviceMetrics = DeviceMetrics{
		deviceConnected: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "device_connected",
			Help:      "Device connection status (1=connected, 0=disconnected).",
		}, []string{"device_id", "device_name", "device_type"}),
		loadPercent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "load_percent",
			Help:      "Device load percentage.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		inputVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "input_voltage_volts",
			Help:      "Input voltage in volts.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		outputVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "output_voltage_volts",
			Help:      "Output voltage in volts.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		inputCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "input_current_amperes",
			Help:      "Input current in amperes.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		outputCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "output_current_amperes",
			Help:      "Output current in amperes.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		inputFrequency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "input_frequency_hertz",
			Help:      "Input frequency in hertz.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		outputFrequency: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "output_frequency_hertz",
			Help:      "Output frequency in hertz.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		inputWatts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "input_watts",
			Help:      "Input active power in watts.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		outputWatts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "output_watts",
			Help:      "Output active power in watts.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
		powerFactorOut: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "output_power_factor",
			Help:      "Output power factor.",
		}, []string{"device_id", "device_name", "device_type", "phase"}),
	}

	// Initialize energy metrics
	m.energyMetrics = EnergyMetrics{
		energyTotalWh: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "energy_total_wh",
			Help:      "Total accumulated energy in watt-hours (Wh). Can be negative for net energy.",
		}, []string{"device_id", "device_name", "device_type"}),
		powerWatts: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "power_watts",
			Help:      "Instantaneous power consumption in watts.",
		}, []string{"device_id", "device_name", "device_type"}),
	}

	// Register all metrics with the registry
	m.registerMetrics()

	return nil
}

// registerMetrics registers all metrics with the Prometheus registry.
func (m *MetricManager) registerMetrics() {
	// Register Exporter metrics
	m.registry.MustRegister(m.exporterMetrics.up)
	m.registry.MustRegister(m.exporterMetrics.requestsTotal)
	m.registry.MustRegister(m.exporterMetrics.requestDuration)
	m.registry.MustRegister(m.exporterMetrics.scrapeErrors)
	m.registry.MustRegister(m.exporterMetrics.collectionTime)
	m.registry.MustRegister(m.exporterMetrics.tokenRefresh)
	m.registry.MustRegister(m.exporterMetrics.deviceCount)

	// Register Connection metrics
	m.registry.MustRegister(m.connectionMetrics.connectionStatus)
	m.registry.MustRegister(m.connectionMetrics.authStatus)
	m.registry.MustRegister(m.connectionMetrics.apiResponseTime)
	m.registry.MustRegister(m.connectionMetrics.tokenExpiry)
	m.registry.MustRegister(m.connectionMetrics.tokenValid)

	// Register Device metrics
	m.registry.MustRegister(m.deviceMetrics.deviceConnected)
	m.registry.MustRegister(m.deviceMetrics.loadPercent)
	m.registry.MustRegister(m.deviceMetrics.inputVoltage)
	m.registry.MustRegister(m.deviceMetrics.outputVoltage)
	m.registry.MustRegister(m.deviceMetrics.inputCurrent)
	m.registry.MustRegister(m.deviceMetrics.outputCurrent)
	m.registry.MustRegister(m.deviceMetrics.inputFrequency)
	m.registry.MustRegister(m.deviceMetrics.outputFrequency)
	m.registry.MustRegister(m.deviceMetrics.inputWatts)
	m.registry.MustRegister(m.deviceMetrics.outputWatts)
	m.registry.MustRegister(m.deviceMetrics.powerFactorOut)

	// Register Energy metrics
	m.registry.MustRegister(m.energyMetrics.energyTotalWh)
	m.registry.MustRegister(m.energyMetrics.powerWatts)
}

// Handler returns the HTTP handler for exposing metrics.
func (m *MetricManager) Handler() http.Handler {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Metrics not initialized", http.StatusServiceUnavailable)
		})
	}

	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// SetUp sets the exporter up status.
func (m *MetricManager) SetUp(status bool) {
	if !m.safeCheckInitialized("SetUp") {
		return
	}

	value := 0.0
	if status {
		value = 1.0
	}

	if err := m.safeSetGauge(m.exporterMetrics.up, value); err != nil {
		m.logger.Error("Failed to set up status", zap.Error(err))
		return
	}

	m.logger.Debug("Exporter status updated", zap.Float64("status", value))
}

// ObserveRequest records an HTTP request with its duration.
func (m *MetricManager) ObserveRequest(host, method, code string, duration time.Duration) {
	if !m.safeCheckInitialized("ObserveRequest") {
		return
	}

	// Record request count
	if err := m.safeIncCounterVec(m.exporterMetrics.requestsTotal, []string{host, method, code}); err != nil {
		m.logger.Error("Failed to record request count", zap.Error(err))
	}

	// Record request duration
	if err := m.safeObserveHistogramVec(m.exporterMetrics.requestDuration, []string{host, method, code}, duration.Seconds()); err != nil {
		m.logger.Error("Failed to record request duration", zap.Error(err))
	}

	m.logger.Debug("HTTP request recorded",
		zap.String("host", host),
		zap.String("method", method),
		zap.String("code", code),
		zap.Duration("duration", duration))
}

// IncScrapeError increments the scrape error counter.
func (m *MetricManager) IncScrapeError(host, errorType string) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.exporterMetrics.scrapeErrors.WithLabelValues(host, errorType).Inc()

	m.logger.Debug("Scrape error recorded",
		zap.String("host", host),
		zap.String("error_type", errorType))
}

// ObserveCollection records the time taken to collect data from WinPower.
func (m *MetricManager) ObserveCollection(host, status string, duration time.Duration) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.exporterMetrics.collectionTime.WithLabelValues(host, status).Observe(duration.Seconds())

	m.logger.Debug("Data collection recorded",
		zap.String("host", host),
		zap.String("status", status),
		zap.Duration("duration", duration))
}

// IncTokenRefresh increments the token refresh counter.
func (m *MetricManager) IncTokenRefresh(host, result string) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.exporterMetrics.tokenRefresh.WithLabelValues(host, result).Inc()

	m.logger.Debug("Token refresh recorded",
		zap.String("host", host),
		zap.String("result", result))
}

// SetDeviceCount sets the device count metric.
func (m *MetricManager) SetDeviceCount(host, deviceType string, count float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.exporterMetrics.deviceCount.WithLabelValues(host, deviceType).Set(count)

	m.logger.Debug("Device count updated",
		zap.String("host", host),
		zap.String("device_type", deviceType),
		zap.Float64("count", count))
}

// SetConnectionStatus sets the connection status metric.
func (m *MetricManager) SetConnectionStatus(host, connectionType string, status float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.connectionMetrics.connectionStatus.WithLabelValues(host, connectionType).Set(status)

	m.logger.Debug("Connection status updated",
		zap.String("host", host),
		zap.String("connection_type", connectionType),
		zap.Float64("status", status))
}

// SetAuthStatus sets the authentication status metric.
func (m *MetricManager) SetAuthStatus(host, authMethod string, status float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.connectionMetrics.authStatus.WithLabelValues(host, authMethod).Set(status)

	m.logger.Debug("Authentication status updated",
		zap.String("host", host),
		zap.String("auth_method", authMethod),
		zap.Float64("status", status))
}

// ObserveAPI records API response time.
func (m *MetricManager) ObserveAPI(host, endpoint string, duration time.Duration) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.connectionMetrics.apiResponseTime.WithLabelValues(host, endpoint).Observe(duration.Seconds())

	m.logger.Debug("API response time recorded",
		zap.String("host", host),
		zap.String("endpoint", endpoint),
		zap.Duration("duration", duration))
}

// SetTokenExpiry sets the token expiry time metric.
func (m *MetricManager) SetTokenExpiry(host, userID string, seconds float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.connectionMetrics.tokenExpiry.WithLabelValues(host, userID).Set(seconds)

	m.logger.Debug("Token expiry time updated",
		zap.String("host", host),
		zap.String("user_id", userID),
		zap.Float64("seconds", seconds))
}

// SetTokenValid sets the token validity metric.
func (m *MetricManager) SetTokenValid(host, userID string, valid float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	m.connectionMetrics.tokenValid.WithLabelValues(host, userID).Set(valid)

	m.logger.Debug("Token validity updated",
		zap.String("host", host),
		zap.String("user_id", userID),
		zap.Float64("valid", valid))
}

// SetDeviceConnected sets the device connection status metric.
func (m *MetricManager) SetDeviceConnected(deviceID, deviceName, deviceType string, connected float64) {
	if !m.safeCheckInitialized("SetDeviceConnected") {
		return
	}

	// Validate input parameters
	if err := m.validateDeviceParameters(deviceID, deviceName, deviceType); err != nil {
		return
	}

	// Use read lock for concurrent safety
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.safeSetGaugeVec(m.deviceMetrics.deviceConnected, []string{deviceID, deviceName, deviceType}, connected); err != nil {
		m.logger.Error("Failed to set device connection status",
			zap.Error(err),
			zap.String("device_id", deviceID),
			zap.String("device_name", deviceName),
			zap.String("device_type", deviceType))
		return
	}

	m.logger.Debug("Device connection status updated",
		zap.String("device_id", deviceID),
		zap.String("device_name", deviceName),
		zap.String("device_type", deviceType),
		zap.Float64("connected", connected))
}

// SetLoadPercent sets the device load percentage metric.
func (m *MetricManager) SetLoadPercent(deviceID, deviceName, deviceType, phase string, pct float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	// Validate input parameters
	if deviceID == "" {
		m.logger.Error("device_id cannot be empty")
		return
	}
	if deviceName == "" {
		m.logger.Error("device_name cannot be empty")
		return
	}
	if deviceType == "" {
		m.logger.Error("device_type cannot be empty")
		return
	}

	m.deviceMetrics.loadPercent.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(pct)

	m.logger.Debug("Device load percent updated",
		zap.String("device_id", deviceID),
		zap.String("device_name", deviceName),
		zap.String("device_type", deviceType),
		zap.String("phase", phase),
		zap.Float64("load_percent", pct))
}

// SetElectricalData sets multiple electrical parameters for a device.
func (m *MetricManager) SetElectricalData(deviceID, deviceName, deviceType, phase string,
	inV, outV, inA, outA, inHz, outHz, inW, outW, pfo float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	// Validate input parameters
	if deviceID == "" {
		m.logger.Error("device_id cannot be empty")
		return
	}
	if deviceName == "" {
		m.logger.Error("device_name cannot be empty")
		return
	}
	if deviceType == "" {
		m.logger.Error("device_type cannot be empty")
		return
	}

	// Update all electrical metrics
	m.deviceMetrics.inputVoltage.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(inV)
	m.deviceMetrics.outputVoltage.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(outV)
	m.deviceMetrics.inputCurrent.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(inA)
	m.deviceMetrics.outputCurrent.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(outA)
	m.deviceMetrics.inputFrequency.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(inHz)
	m.deviceMetrics.outputFrequency.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(outHz)
	m.deviceMetrics.inputWatts.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(inW)
	m.deviceMetrics.outputWatts.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(outW)
	m.deviceMetrics.powerFactorOut.WithLabelValues(deviceID, deviceName, deviceType, phase).Set(pfo)

	m.logger.Debug("Device electrical data updated",
		zap.String("device_id", deviceID),
		zap.String("device_name", deviceName),
		zap.String("device_type", deviceType),
		zap.String("phase", phase),
		zap.Float64("input_voltage", inV),
		zap.Float64("output_voltage", outV),
		zap.Float64("input_current", inA),
		zap.Float64("output_current", outA),
		zap.Float64("input_frequency", inHz),
		zap.Float64("output_frequency", outHz),
		zap.Float64("input_watts", inW),
		zap.Float64("output_watts", outW),
		zap.Float64("power_factor_out", pfo))
}

// SetPowerWatts sets the instantaneous power consumption metric for a device.
func (m *MetricManager) SetPowerWatts(deviceID, deviceName, deviceType string, watts float64) {
	if !m.safeCheckInitialized("SetPowerWatts") {
		return
	}

	// Validate input parameters
	if err := m.validateDeviceParameters(deviceID, deviceName, deviceType); err != nil {
		return
	}

	if err := m.safeSetGaugeVec(m.energyMetrics.powerWatts, []string{deviceID, deviceName, deviceType}, watts); err != nil {
		m.logger.Error("Failed to set device power watts",
			zap.Error(err),
			zap.String("device_id", deviceID),
			zap.String("device_name", deviceName),
			zap.String("device_type", deviceType))
		return
	}

	m.logger.Debug("Device power watts updated",
		zap.String("device_id", deviceID),
		zap.String("device_name", deviceName),
		zap.String("device_type", deviceType),
		zap.Float64("power_watts", watts))
}

// SetEnergyTotalWh sets the total accumulated energy metric for a device.
// Supports negative values for net energy (energy export scenarios).
// Precision: 0.01 Wh (rounded to 2 decimal places).
func (m *MetricManager) SetEnergyTotalWh(deviceID, deviceName, deviceType string, wh float64) {
	if !m.initialized {
		m.logger.Error("MetricManager not initialized")
		return
	}

	// Validate input parameters
	if deviceID == "" {
		m.logger.Error("device_id cannot be empty")
		return
	}
	if deviceName == "" {
		m.logger.Error("device_name cannot be empty")
		return
	}
	if deviceType == "" {
		m.logger.Error("device_type cannot be empty")
		return
	}

	// Apply precision control (0.01 Wh precision)
	roundedWh := roundToPrecision(wh, 0.01)

	m.energyMetrics.energyTotalWh.WithLabelValues(deviceID, deviceName, deviceType).Set(roundedWh)

	m.logger.Debug("Device energy total updated",
		zap.String("device_id", deviceID),
		zap.String("device_name", deviceName),
		zap.String("device_type", deviceType),
		zap.Float64("energy_total_wh_input", wh),
		zap.Float64("energy_total_wh_rounded", roundedWh))
}

// roundToPrecision rounds a float64 to the specified precision.
func roundToPrecision(value, precision float64) float64 {
	if precision <= 0 {
		return value
	}

	// Calculate the multiplier
	multiplier := 1.0 / precision

	// Round to the specified precision
	return math.Round(value*multiplier) / multiplier
}

// Helper function to convert boolean to float64
// TODO: Keep this function for future use when converting boolean metrics
//
//nolint:unused // This function is kept for future use
func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

// safeCheckInitialized checks if the metric manager is initialized safely.
func (m *MetricManager) safeCheckInitialized(operation string) bool {
	if m == nil {
		// Can't log if m is nil, so we just return false
		return false
	}

	if !m.initialized {
		if m.logger != nil {
			m.logger.Error("MetricManager not initialized",
				zap.String("operation", operation),
				zap.Error(NewErrNotInitialized("MetricManager")))
		}
		return false
	}

	return true
}

// safeSetGauge safely sets a gauge metric value with error handling.
func (m *MetricManager) safeSetGauge(gauge prometheus.Gauge, value float64) error {
	defer func() {
		if r := recover(); r != nil {
			if m.logger != nil {
				m.logger.Error("Panic recovered while setting gauge",
					zap.Any("panic", r),
					zap.Float64("value", value))
			}
		}
	}()

	gauge.Set(value)
	return nil
}

// safeIncCounter safely increments a counter metric with error handling.
//
//nolint:unused // This method is kept for future use
func (m *MetricManager) safeIncCounter(counter prometheus.Counter) error {
	defer func() {
		if r := recover(); r != nil {
			if m.logger != nil {
				m.logger.Error("Panic recovered while incrementing counter",
					zap.Any("panic", r))
			}
		}
	}()

	counter.Inc()
	return nil
}

// safeObserveHistogram safely observes a histogram value with error handling.
//
//nolint:unused // This method is kept for future use
func (m *MetricManager) safeObserveHistogram(histogram prometheus.Observer, value float64) error {
	defer func() {
		if r := recover(); r != nil {
			if m.logger != nil {
				m.logger.Error("Panic recovered while observing histogram",
					zap.Any("panic", r),
					zap.Float64("value", value))
			}
		}
	}()

	histogram.Observe(value)
	return nil
}

// safeSetGaugeVec safely sets a gauge vector metric with error handling.
func (m *MetricManager) safeSetGaugeVec(gaugeVec *prometheus.GaugeVec, labelValues []string, value float64) error {
	defer func() {
		if r := recover(); r != nil {
			if m.logger != nil {
				m.logger.Error("Panic recovered while setting gauge vector",
					zap.Any("panic", r),
					zap.Strings("label_values", labelValues),
					zap.Float64("value", value))
			}
		}
	}()

	gaugeVec.WithLabelValues(labelValues...).Set(value)
	return nil
}

// safeIncCounterVec safely increments a counter vector metric with error handling.
func (m *MetricManager) safeIncCounterVec(counterVec *prometheus.CounterVec, labelValues []string) error {
	defer func() {
		if r := recover(); r != nil {
			if m.logger != nil {
				m.logger.Error("Panic recovered while incrementing counter vector",
					zap.Any("panic", r),
					zap.Strings("label_values", labelValues))
			}
		}
	}()

	counterVec.WithLabelValues(labelValues...).Inc()
	return nil
}

// safeObserveHistogramVec safely observes a histogram vector value with error handling.
func (m *MetricManager) safeObserveHistogramVec(histogramVec *prometheus.HistogramVec, labelValues []string, value float64) error {
	defer func() {
		if r := recover(); r != nil {
			if m.logger != nil {
				m.logger.Error("Panic recovered while observing histogram vector",
					zap.Any("panic", r),
					zap.Strings("label_values", labelValues),
					zap.Float64("value", value))
			}
		}
	}()

	histogramVec.WithLabelValues(labelValues...).Observe(value)
	return nil
}

// validateDeviceParameters validates common device parameters.
func (m *MetricManager) validateDeviceParameters(deviceID, deviceName, deviceType string) error {
	if deviceID == "" {
		err := NewErrInvalidLabelValue("device_id", deviceID, "cannot be empty")
		if m.logger != nil {
			m.logger.Error("device_id cannot be empty", zap.Error(err))
		}
		return err
	}
	if deviceName == "" {
		err := NewErrInvalidLabelValue("device_name", deviceName, "cannot be empty")
		if m.logger != nil {
			m.logger.Error("device_name cannot be empty", zap.Error(err))
		}
		return err
	}
	if deviceType == "" {
		err := NewErrInvalidLabelValue("device_type", deviceType, "cannot be empty")
		if m.logger != nil {
			m.logger.Error("device_type cannot be empty", zap.Error(err))
		}
		return err
	}
	return nil
}

// SetLogLevel sets the log level for the metrics module.
func (m *MetricManager) SetLogLevel(level string) error {
	if !m.safeCheckInitialized("SetLogLevel") {
		return NewErrNotInitialized("MetricManager")
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLevels[level] {
		err := NewErrInvalidConfig("log_level", fmt.Sprintf("invalid level '%s', must be one of: debug, info, warn, error, fatal", level))
		if m.logger != nil {
			m.logger.Error("Invalid log level", zap.String("level", level), zap.Error(err))
		}
		return err
	}

	m.mu.Lock()
	m.logLevel = level
	m.mu.Unlock()

	if m.logger != nil {
		m.logger.Info("Log level updated", zap.String("old_level", m.GetLogLevel()), zap.String("new_level", level))
	}

	return nil
}

// GetLogLevel returns the current log level.
func (m *MetricManager) GetLogLevel() string {
	if m == nil || !m.initialized {
		return "unknown"
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logLevel
}

// shouldLog checks if a message should be logged based on the current log level.
//
//nolint:unused // This method is kept for future use
func (m *MetricManager) shouldLog(messageLevel string) bool {
	if m == nil {
		return false
	}

	m.mu.RLock()
	currentLevel := m.logLevel
	m.mu.RUnlock()

	// Define log level hierarchy (higher value = higher priority)
	levelHierarchy := map[string]int{
		"debug": 0,
		"info":  1,
		"warn":  2,
		"error": 3,
		"fatal": 4,
	}

	messagePriority, messageExists := levelHierarchy[messageLevel]
	currentPriority, currentExists := levelHierarchy[currentLevel]

	if !messageExists || !currentExists {
		return false
	}

	return messagePriority >= currentPriority
}
