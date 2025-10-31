package metrics

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
)

// NewMetricsService creates a new MetricsService instance
// Parameters:
//   - collector: The collector interface for triggering data collection
//   - logger: The logger for structured logging
//   - config: Optional configuration (uses defaults if nil)
//
// Returns:
//   - *MetricsService: The initialized metrics service
//   - error: Error if initialization fails
func NewMetricsService(
	coll collector.CollectorInterface,
	logger *zap.Logger,
	config *MetricsConfig,
) (*MetricsService, error) {
	// Validate inputs
	if coll == nil {
		return nil, ErrCollectorNil
	}
	if logger == nil {
		return nil, ErrLoggerNil
	}

	// Use default config if not provided
	if config == nil {
		config = DefaultMetricsConfig()
	}

	// Create new registry to avoid conflicts with default registry
	registry := prometheus.NewRegistry()

	// Create service instance
	m := &MetricsService{
		registry:      registry,
		collector:     coll,
		logger:        logger,
		winpowerHost:  config.WinPowerHost,
		deviceMetrics: make(map[string]*DeviceMetrics),
	}

	// Initialize metrics
	m.initExporterMetrics(config)
	m.initConnectionMetrics(config)

	// Register all metrics with the registry
	m.registerMetrics()

	logger.Info("Metrics service initialized",
		zap.String("namespace", config.Namespace),
		zap.String("subsystem", config.Subsystem),
		zap.String("winpower_host", config.WinPowerHost),
		zap.Bool("memory_metrics_enabled", config.EnableMemoryMetrics),
	)

	return m, nil
}

// HandleMetrics is the Gin handler for the /metrics endpoint
// It triggers data collection and serves Prometheus-formatted metrics
func (m *MetricsService) HandleMetrics(c *gin.Context) {
	startTime := time.Now()

	// Increment request counter
	m.requestsTotal.WithLabelValues().Inc()

	// Log request
	m.logger.Debug("Handling /metrics request",
		zap.String("remote_addr", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	// Trigger data collection
	collectionResult, err := m.collector.CollectDeviceData(c.Request.Context())
	if err != nil {
		m.handleCollectionError(err)
		m.logger.Error("Failed to collect device data",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		c.String(http.StatusInternalServerError, "Failed to collect metrics: %v", err)
		return
	}

	// Update metrics based on collection result
	if err := m.updateMetrics(c.Request.Context(), collectionResult); err != nil {
		m.logger.Error("Failed to update metrics",
			zap.Error(err),
			zap.Duration("elapsed", time.Since(startTime)),
		)
		// Don't return error - still serve existing metrics
	}

	// Update self-monitoring metrics
	m.updateSelfMetrics(collectionResult)

	// Serve metrics in Prometheus format
	handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		ErrorLog:      &zapLogger{logger: m.logger},
		ErrorHandling: promhttp.ContinueOnError,
	})
	handler.ServeHTTP(c.Writer, c.Request)

	// Record request duration
	duration := time.Since(startTime).Seconds()
	m.requestDuration.WithLabelValues().Observe(duration)

	m.logger.Debug("Metrics request completed",
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("device_count", collectionResult.DeviceCount),
		zap.Bool("success", collectionResult.Success),
	)
}

// updateMetrics updates all metrics based on the collection result
func (m *MetricsService) updateMetrics(ctx context.Context, result *collector.CollectionResult) error {
	if result == nil {
		return ErrInvalidCollectionResult
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update collection timestamp
	m.lastCollectionTimeSeconds.Set(float64(result.CollectionTime.Unix()))

	// Update device count
	m.deviceCount.Set(float64(result.DeviceCount))

	// Update connection status based on collection success
	if result.Success {
		m.connectionStatus.Set(1)
	} else {
		m.connectionStatus.Set(0)
	}

	// Update each device's metrics
	for deviceID, deviceInfo := range result.Devices {
		if err := m.updateDeviceMetrics(deviceID, deviceInfo); err != nil {
			m.logger.Warn("Failed to update device metrics",
				zap.String("device_id", deviceID),
				zap.Error(err),
			)
			// Continue with other devices
			continue
		}
	}

	return nil
}

// updateDeviceMetrics updates metrics for a single device
func (m *MetricsService) updateDeviceMetrics(deviceID string, info *collector.DeviceCollectionInfo) error {
	if info == nil {
		return fmt.Errorf("device info is nil for device %s", deviceID)
	}

	// Get or create device metrics
	dm, exists := m.deviceMetrics[deviceID]
	if !exists {
		// Create new device metrics
		dm = m.createDeviceMetrics(
			deviceID,
			info.DeviceName,
			strconv.Itoa(info.DeviceType),
			m.winpowerHost,
		)
		m.deviceMetrics[deviceID] = dm
		m.logger.Info("Created metrics for new device",
			zap.String("device_id", deviceID),
			zap.String("device_name", info.DeviceName),
		)
	}

	// Update device status
	if info.Connected {
		dm.connected.Set(1)
	} else {
		dm.connected.Set(0)
	}
	dm.lastUpdateTimestamp.Set(float64(info.LastUpdateTime.Unix()))

	// Update input parameters
	dm.inputVoltage.Set(info.InputVolt1)
	dm.inputFrequency.Set(info.InputFreq)

	// Update output parameters
	dm.outputVoltage.Set(info.OutputVolt1)
	dm.outputCurrent.Set(info.OutputCurrent1)
	dm.outputFrequency.Set(info.OutputFreq)
	// Convert output voltage type to numeric value
	dm.outputVoltageType.Set(encodeOutputVoltageType(info.OutputVoltageType))

	// Update load and power - LoadTotalWatt is the core metric
	dm.loadPercent.Set(info.LoadPercent)
	dm.loadTotalWatt.Set(info.LoadTotalWatt)
	dm.loadTotalVa.Set(info.LoadTotalVa)
	dm.loadWattPhase1.Set(info.LoadWatt1)
	dm.loadVaPhase1.Set(info.LoadVa1)
	// PowerWatts is the same as LoadTotalWatt (instantaneous power)
	dm.powerWatts.Set(info.LoadTotalWatt)

	// Update battery parameters
	if info.IsCharging {
		dm.batteryCharging.Set(1)
	} else {
		dm.batteryCharging.Set(0)
	}
	dm.batteryVoltagePercent.Set(info.BatVoltP)
	dm.batteryCapacity.Set(info.BatCapacity)
	dm.batteryRemainSeconds.Set(float64(info.BatRemainTime))
	dm.batteryStatus.Set(encodeBatteryStatus(info.BatteryStatus))

	// Update UPS status
	dm.upsTemperature.Set(info.UpsTemperature)
	dm.upsMode.Set(encodeUPSMode(info.Mode))
	dm.upsStatus.Set(encodeUPSStatus(info.Status))
	dm.upsTestStatus.Set(encodeTestStatus(info.TestStatus))

	// Update fault code with label
	if info.FaultCode != "" {
		dm.upsFaultCode.WithLabelValues(info.FaultCode).Set(1)
	} else {
		dm.upsFaultCode.WithLabelValues("none").Set(0)
	}

	// Update energy if calculated
	if info.EnergyCalculated {
		dm.cumulativeEnergy.Set(info.EnergyValue)
	}

	return nil
}

// updateSelfMetrics updates exporter self-monitoring metrics
func (m *MetricsService) updateSelfMetrics(result *collector.CollectionResult) {
	// Record collection duration
	m.collectionDuration.WithLabelValues().Observe(result.Duration.Seconds())

	// Update memory metrics if enabled
	if m.memoryBytes != nil {
		m.updateMemoryMetrics()
	}
}

// updateMemoryMetrics updates memory usage metrics
func (m *MetricsService) updateMemoryMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	m.memoryBytes.WithLabelValues("alloc").Set(float64(memStats.Alloc))
	m.memoryBytes.WithLabelValues("sys").Set(float64(memStats.Sys))
	m.memoryBytes.WithLabelValues("heap").Set(float64(memStats.HeapAlloc))
}

// handleCollectionError handles collection errors and updates error metrics
func (m *MetricsService) handleCollectionError(err error) {
	// Classify error type
	errorType := "unknown"
	if err != nil {
		// Classify error by type using type switch
		switch err {
		case context.DeadlineExceeded:
			errorType = "timeout"
		case context.Canceled:
			errorType = "cancelled"
		default:
			errorType = "collection_failed"
		}
	}

	// Increment error counter
	m.scrapeErrorsTotal.WithLabelValues(errorType).Inc()

	// Set connection status to down
	m.connectionStatus.Set(0)
}

// Encoding functions for string values to numeric codes

func encodeOutputVoltageType(voltageType string) float64 {
	switch voltageType {
	case "single":
		return 1
	case "three":
		return 3
	default:
		return 0
	}
}

func encodeBatteryStatus(status string) float64 {
	// TODO: Define proper battery status codes
	switch status {
	case "normal":
		return 1
	case "low":
		return 2
	case "depleted":
		return 3
	default:
		return 0
	}
}

func encodeUPSMode(mode string) float64 {
	// TODO: Define proper UPS mode codes
	switch mode {
	case "online":
		return 1
	case "battery":
		return 2
	case "bypass":
		return 3
	default:
		return 0
	}
}

func encodeUPSStatus(status string) float64 {
	// TODO: Define proper UPS status codes
	switch status {
	case "normal":
		return 1
	case "warning":
		return 2
	case "alarm":
		return 3
	default:
		return 0
	}
}

func encodeTestStatus(testStatus string) float64 {
	// TODO: Define proper test status codes
	switch testStatus {
	case "no_test":
		return 0
	case "testing":
		return 1
	case "passed":
		return 2
	case "failed":
		return 3
	default:
		return 0
	}
}

// zapLogger wraps zap.Logger to implement promhttp.Logger interface
type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Println(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
}
