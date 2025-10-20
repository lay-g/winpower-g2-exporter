package energy

import (
	"time"

	"go.uber.org/zap"
)

// NewEnergyService creates a new EnergyService instance with the specified dependencies.
// If config is nil, default configuration will be used.
// If logger is nil, a default logger will be created.
func NewEnergyService(storage StorageManager, logger *zap.Logger, config *Config) *EnergyService {
	// Use default config if none provided
	if config == nil {
		config = DefaultConfig()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		// If config validation fails, use default config
		logger.Error("Invalid energy config, using defaults", zap.Error(err))
		config = DefaultConfig()
	}

	// Create default logger if none provided
	if logger == nil {
		logger = zap.NewNop() // No-op logger as fallback
	}

	return &EnergyService{
		storage: storage,
		logger:  logger,
		config:  config,
		stats:   &SimpleStats{},
	}
}

// Calculate performs energy accumulation calculation for a device.
// This method is thread-safe and uses a global lock to ensure serial execution.
// Returns the new total accumulated energy in watt-hours (Wh).
func (es *EnergyService) Calculate(deviceID string, power float64) (float64, error) {
	// Validate inputs
	if err := ValidateDeviceID(deviceID); err != nil {
		return 0, NewEnergyError("Calculate", deviceID, "invalid device ID", err)
	}

	if err := ValidatePower(power, es.config); err != nil {
		return 0, NewEnergyError("Calculate", deviceID, "invalid power value", err)
	}

	// Acquire write lock for serial execution
	es.mutex.Lock()
	defer es.mutex.Unlock()

	start := time.Now()
	logger := es.logger.With(
		zap.String("device_id", deviceID),
		zap.Float64("power", power),
	)

	logger.Debug("Starting energy calculation")

	// Load historical data
	historyData, err := es.loadHistoryData(deviceID)
	if err != nil {
		es.updateStats(false, time.Since(start))
		logger.Error("Failed to load history data", zap.Error(err))
		return 0, NewEnergyError("Calculate", deviceID, "failed to load history data", err)
	}

	// Calculate total energy
	totalEnergy, err := es.calculateTotalEnergy(historyData, power, time.Now())
	if err != nil {
		es.updateStats(false, time.Since(start))
		logger.Error("Failed to calculate energy", zap.Error(err))
		return 0, NewEnergyError("Calculate", deviceID, "failed to calculate energy", err)
	}

	// Save data to storage
	err = es.saveData(deviceID, totalEnergy)
	if err != nil {
		es.updateStats(false, time.Since(start))
		logger.Error("Failed to save data", zap.Error(err))
		return 0, NewEnergyError("Calculate", deviceID, "failed to save data", err)
	}

	// Update statistics
	es.updateStats(true, time.Since(start))

	duration := time.Since(start)
	logger.Info("Energy calculation completed",
		zap.Float64("total_energy", totalEnergy),
		zap.Duration("duration", duration))

	return totalEnergy, nil
}

// Get retrieves the latest accumulated energy data for a device.
// This method is thread-safe and allows concurrent reads.
func (es *EnergyService) Get(deviceID string) (float64, error) {
	// Validate input
	if err := ValidateDeviceID(deviceID); err != nil {
		return 0, NewEnergyError("Get", deviceID, "invalid device ID", err)
	}

	// Acquire read lock to allow concurrent reads
	es.mutex.RLock()
	defer es.mutex.RUnlock()

	data, err := es.storage.Read(deviceID)
	if err != nil {
		return 0, NewEnergyError("Get", deviceID, "failed to read from storage", err)
	}

	// Handle case where no data exists (new device)
	if data == nil {
		return 0, nil
	}

	return data.EnergyWH, nil
}

// GetStats returns the current calculation statistics.
// This method is thread-safe.
func (es *EnergyService) GetStats() *SimpleStats {
	if !es.config.EnableStats {
		return &SimpleStats{} // Return empty stats if disabled
	}

	es.stats.mutex.RLock()
	defer es.stats.mutex.RUnlock()

	return &SimpleStats{
		TotalCalculations:   es.stats.TotalCalculations,
		TotalErrors:         es.stats.TotalErrors,
		LastUpdateTime:      es.stats.LastUpdateTime,
		AvgCalculationTime:  es.stats.AvgCalculationTime,
	}
}

// loadHistoryData loads historical data from storage for a device.
// Returns nil if no historical data exists (for new devices).
func (es *EnergyService) loadHistoryData(deviceID string) (*PowerData, error) {
	data, err := es.storage.Read(deviceID)
	if err != nil {
		// Check if it's a "not found" error, which is normal for new devices
		if err.Error() == "file not found" {
			es.logger.Debug("No historical data found for device",
				zap.String("device_id", deviceID))
			return nil, nil
		}
		return nil, err
	}

	// Validate the loaded data
	if data.IsZero() {
		es.logger.Debug("Historical data is zero, treating as new device",
			zap.String("device_id", deviceID))
		return nil, nil
	}

	return data, nil
}

// calculateTotalEnergy calculates the new total accumulated energy.
// If historyData is nil, starts from zero.
func (es *EnergyService) calculateTotalEnergy(historyData *PowerData, currentPower float64, currentTime time.Time) (float64, error) {
	if historyData == nil {
		// First calculation for this device, start from zero
		return 0, nil
	}

	// Calculate time difference in hours
	timeDiff := currentTime.Sub(time.UnixMilli(historyData.Timestamp))
	hoursDiff := timeDiff.Hours()

	if hoursDiff <= 0 {
		// Invalid time interval, return historical value
		es.logger.Warn("Invalid time interval, returning historical value",
			zap.Float64("hours_diff", hoursDiff),
			zap.Int64("history_timestamp", historyData.Timestamp),
			zap.Int64("current_timestamp", currentTime.UnixMilli()))
		return historyData.EnergyWH, nil
	}

	// Calculate interval energy: power (W) × time (h) = energy (Wh)
	intervalEnergy := currentPower * hoursDiff

	// Calculate new total energy
	totalEnergy := historyData.EnergyWH + intervalEnergy

	// Apply precision rounding
	totalEnergy = es.roundToPrecision(totalEnergy)

	es.logger.Debug("Energy calculation details",
		zap.Float64("history_energy", historyData.EnergyWH),
		zap.Float64("current_power", currentPower),
		zap.Float64("hours_diff", hoursDiff),
		zap.Float64("interval_energy", intervalEnergy),
		zap.Float64("total_energy", totalEnergy))

	return totalEnergy, nil
}

// saveData saves energy data to storage.
func (es *EnergyService) saveData(deviceID string, energy float64) error {
	data := &PowerData{
		Timestamp: time.Now().UnixMilli(),
		EnergyWH:  energy,
	}

	return es.storage.Write(deviceID, data)
}

// updateStats updates the calculation statistics.
func (es *EnergyService) updateStats(success bool, duration time.Duration) {
	if !es.config.EnableStats {
		return
	}

	es.stats.mutex.Lock()
	defer es.stats.mutex.Unlock()

	es.stats.TotalCalculations++
	if !success {
		es.stats.TotalErrors++
	}

	// Update average calculation time using simple moving average
	durationNanos := duration.Nanoseconds()
	if es.stats.AvgCalculationTime == 0 {
		es.stats.AvgCalculationTime = durationNanos
	} else {
		// Weighted moving average with 0.9 weight for historical values
		es.stats.AvgCalculationTime = (es.stats.AvgCalculationTime*9 + durationNanos) / 10
	}

	es.stats.LastUpdateTime = time.Now().UnixMilli()
}

// roundToPrecision rounds a value to the configured precision.
func (es *EnergyService) roundToPrecision(value float64) float64 {
	if es.config.Precision <= 0 {
		return value
	}

	// Simple rounding to specified precision
	factor := 1.0 / es.config.Precision
	return float64(int(value*factor+0.5)) / factor
}