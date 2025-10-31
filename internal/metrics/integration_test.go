package metrics_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lay-g/winpower-g2-exporter/internal/collector"
	"github.com/lay-g/winpower-g2-exporter/internal/metrics"
	"github.com/lay-g/winpower-g2-exporter/internal/metrics/mocks"
	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
)

func TestMetricsIntegration(t *testing.T) {
	// Setup
	logger := log.NewTestLogger()
	mockCollector := mocks.NewMockCollectorWithDevices()

	// Create metrics service
	service, err := metrics.NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)
	require.NotNil(t, service)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", service.HandleMetrics)

	// Create test request
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

	// Verify response body contains expected metrics
	body := w.Body.String()
	assert.Contains(t, body, "winpower_exporter_up")
	assert.Contains(t, body, "winpower_device_connected")
	assert.Contains(t, body, "winpower_device_load_total_watts")
	assert.Contains(t, body, "winpower_power_watts")
	assert.Contains(t, body, "winpower_device_cumulative_energy")
}

func TestMetricsIntegration_CollectionFailure(t *testing.T) {
	// Setup
	logger := log.NewTestLogger()

	// Create mock collector that returns error
	mockCollector := &mocks.MockCollector{
		CollectDeviceDataFunc: func(ctx context.Context) (*collector.CollectionResult, error) {
			return nil, assert.AnError
		},
	}

	// Create metrics service
	service, err := metrics.NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", service.HandleMetrics)

	// Create test request
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Serve request
	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestMetricsIntegration_MultipleRequests(t *testing.T) {
	// Setup
	logger := log.NewTestLogger()
	mockCollector := mocks.NewMockCollectorWithDevices()

	// Create metrics service
	service, err := metrics.NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", service.HandleMetrics)

	// Make multiple requests
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", "/metrics", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Small delay between requests
		time.Sleep(10 * time.Millisecond)
	}
}

func TestMetricsIntegration_DynamicDevices(t *testing.T) {
	// Setup
	logger := log.NewTestLogger()

	deviceCount := 1
	mockCollector := &mocks.MockCollector{
		CollectDeviceDataFunc: func(ctx context.Context) (*collector.CollectionResult, error) {
			devices := make(map[string]*collector.DeviceCollectionInfo)

			// Add devices dynamically
			for i := 1; i <= deviceCount; i++ {
				deviceID := "device" + string(rune('0'+i))
				devices[deviceID] = &collector.DeviceCollectionInfo{
					DeviceID:       deviceID,
					DeviceName:     "Test Device " + string(rune('0'+i)),
					DeviceType:     1,
					Connected:      true,
					LastUpdateTime: time.Now(),
					LoadTotalWatt:  float64(i * 100),
				}
			}

			return &collector.CollectionResult{
				Success:        true,
				DeviceCount:    deviceCount,
				CollectionTime: time.Now(),
				Duration:       100 * time.Millisecond,
				Devices:        devices,
			}, nil
		},
	}

	// Create metrics service
	service, err := metrics.NewMetricsService(mockCollector, logger, nil)
	require.NoError(t, err)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metrics", service.HandleMetrics)

	// First request - 1 device
	req1, _ := http.NewRequest("GET", "/metrics", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Increase device count
	deviceCount = 3

	// Second request - 3 devices
	req2, _ := http.NewRequest("GET", "/metrics", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Verify metrics for new devices exist
	body := w2.Body.String()
	assert.Contains(t, body, "device1")
	assert.Contains(t, body, "device2")
	assert.Contains(t, body, "device3")
}
