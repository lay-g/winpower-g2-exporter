package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsConfig(t *testing.T) {
	t.Run("default config has correct values", func(t *testing.T) {
		config := DefaultMetricsConfig()

		assert.Equal(t, "winpower", config.Namespace)
		assert.Equal(t, "exporter", config.Subsystem)
		assert.Equal(t, "localhost", config.WinPowerHost)
		assert.True(t, config.EnableMemoryMetrics)
	})

	t.Run("custom config can be created", func(t *testing.T) {
		config := &MetricsConfig{
			Namespace:           "custom",
			Subsystem:           "test",
			WinPowerHost:        "test-host",
			EnableMemoryMetrics: false,
		}

		assert.Equal(t, "custom", config.Namespace)
		assert.Equal(t, "test", config.Subsystem)
		assert.Equal(t, "test-host", config.WinPowerHost)
		assert.False(t, config.EnableMemoryMetrics)
	})
}

func TestMetricsServiceStructure(t *testing.T) {
	// This test verifies that the MetricsService and DeviceMetrics
	// structs have the expected fields (compile-time check)

	var ms MetricsService
	assert.NotNil(t, &ms.registry)
	assert.NotNil(t, &ms.collector)
	assert.NotNil(t, &ms.logger)
	assert.NotNil(t, &ms.deviceMetrics)
	assert.NotNil(t, &ms.mu)

	// Verify exporter metrics fields exist
	assert.NotNil(t, &ms.exporterUp)
	assert.NotNil(t, &ms.requestsTotal)
	assert.NotNil(t, &ms.requestDuration)
	assert.NotNil(t, &ms.collectionDuration)
	assert.NotNil(t, &ms.scrapeErrorsTotal)
	assert.NotNil(t, &ms.tokenRefreshTotal)
	assert.NotNil(t, &ms.deviceCount)
	assert.NotNil(t, &ms.lastCollectionTimeSeconds)

	// Verify connection metrics fields exist
	assert.NotNil(t, &ms.connectionStatus)
	assert.NotNil(t, &ms.authStatus)
	assert.NotNil(t, &ms.apiResponseTime)
	assert.NotNil(t, &ms.tokenExpirySeconds)
	assert.NotNil(t, &ms.tokenValid)
}

func TestDeviceMetricsStructure(t *testing.T) {
	// This test verifies that DeviceMetrics struct has all expected fields

	var dm DeviceMetrics

	// Status fields
	assert.NotNil(t, &dm.connected)
	assert.NotNil(t, &dm.lastUpdateTimestamp)

	// Input parameters
	assert.NotNil(t, &dm.inputVoltage)
	assert.NotNil(t, &dm.inputFrequency)

	// Output parameters
	assert.NotNil(t, &dm.outputVoltage)
	assert.NotNil(t, &dm.outputCurrent)
	assert.NotNil(t, &dm.outputFrequency)
	assert.NotNil(t, &dm.outputVoltageType)

	// Load and power
	assert.NotNil(t, &dm.loadPercent)
	assert.NotNil(t, &dm.loadTotalWatt)
	assert.NotNil(t, &dm.loadTotalVa)
	assert.NotNil(t, &dm.loadWattPhase1)
	assert.NotNil(t, &dm.loadVaPhase1)
	assert.NotNil(t, &dm.powerWatts)

	// Battery
	assert.NotNil(t, &dm.batteryCharging)
	assert.NotNil(t, &dm.batteryVoltagePercent)
	assert.NotNil(t, &dm.batteryCapacity)
	assert.NotNil(t, &dm.batteryRemainSeconds)
	assert.NotNil(t, &dm.batteryStatus)

	// UPS
	assert.NotNil(t, &dm.upsTemperature)
	assert.NotNil(t, &dm.upsMode)
	assert.NotNil(t, &dm.upsStatus)
	assert.NotNil(t, &dm.upsTestStatus)
	assert.NotNil(t, &dm.upsFaultCode)

	// Energy
	assert.NotNil(t, &dm.cumulativeEnergy)
}
