package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// MetricManagerInterface defines the interface for metrics management operations.
// This interface provides abstraction for Prometheus metric registration, updates, and exposure.
type MetricManagerInterface interface {
	// HTTP暴露
	Handler() http.Handler

	// Exporter自监控
	SetUp(status bool)
	ObserveRequest(host, method, code string, duration time.Duration)
	IncScrapeError(host, errorType string)
	ObserveCollection(host, status string, duration time.Duration)
	IncTokenRefresh(host, result string)
	SetDeviceCount(host, deviceType string, count float64)

	// WinPower连接/认证
	SetConnectionStatus(host, connectionType string, status float64)
	SetAuthStatus(host, authMethod string, status float64)
	ObserveAPI(host, endpoint string, duration time.Duration)
	SetTokenExpiry(host, userID string, seconds float64)
	SetTokenValid(host, userID string, valid float64)

	// 设备/电源数据
	SetDeviceConnected(deviceID, deviceName, deviceType string, connected float64)
	SetLoadPercent(deviceID, deviceName, deviceType, phase string, pct float64)
	SetElectricalData(deviceID, deviceName, deviceType, phase string,
		inV, outV, inA, outA, inHz, outHz, inW, outW, pfo float64)

	// 能耗数据
	SetPowerWatts(deviceID, deviceName, deviceType string, watts float64)
	SetEnergyTotalWh(deviceID, deviceName, deviceType string, wh float64)

	// 日志和调试控制
	SetLogLevel(level string) error
	GetLogLevel() string
}

// MetricManager implements the MetricManagerInterface with a Prometheus-based architecture.
// It provides centralized metrics management for the WinPower G2 Exporter.
type MetricManager struct {
	registry *prometheus.Registry
	logger   *zap.Logger
	config   MetricManagerConfig

	// Exporter 自监控指标
	exporterMetrics ExporterMetrics

	// WinPower 连接/认证指标
	connectionMetrics ConnectionMetrics

	// 设备/电源指标
	deviceMetrics DeviceMetrics

	// 能耗指标
	energyMetrics EnergyMetrics

	initialized bool
	logLevel    string // Current log level for debugging
	mu          sync.RWMutex
}

// ExporterMetrics represents exporter self-monitoring metrics
type ExporterMetrics struct {
	up              prometheus.Gauge
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	scrapeErrors    *prometheus.CounterVec
	collectionTime  *prometheus.HistogramVec
	tokenRefresh    *prometheus.CounterVec
	deviceCount     *prometheus.GaugeVec
}

// ConnectionMetrics represents WinPower connection/authentication metrics
type ConnectionMetrics struct {
	connectionStatus *prometheus.GaugeVec
	authStatus       *prometheus.GaugeVec
	apiResponseTime  *prometheus.HistogramVec
	tokenExpiry      *prometheus.GaugeVec
	tokenValid       *prometheus.GaugeVec
}

// DeviceMetrics represents device/power supply metrics
type DeviceMetrics struct {
	deviceConnected *prometheus.GaugeVec
	loadPercent     *prometheus.GaugeVec
	inputVoltage    *prometheus.GaugeVec
	outputVoltage   *prometheus.GaugeVec
	inputCurrent    *prometheus.GaugeVec
	outputCurrent   *prometheus.GaugeVec
	inputFrequency  *prometheus.GaugeVec
	outputFrequency *prometheus.GaugeVec
	inputWatts      *prometheus.GaugeVec
	outputWatts     *prometheus.GaugeVec
	powerFactorOut  *prometheus.GaugeVec
}

// EnergyMetrics represents energy consumption metrics
type EnergyMetrics struct {
	energyTotalWh *prometheus.GaugeVec
	powerWatts    *prometheus.GaugeVec
}
