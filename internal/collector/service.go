package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/lay-g/winpower-g2-exporter/internal/pkgs/log"
	"github.com/lay-g/winpower-g2-exporter/internal/winpower"
)

// CollectorService is the core implementation of data collection and coordination
type CollectorService struct {
	winpowerClient WinPowerClient
	energyCalc     EnergyCalculator
	logger         log.Logger
}

// NewCollectorService creates a new collector service with dependency injection
func NewCollectorService(
	winpowerClient WinPowerClient,
	energyCalc EnergyCalculator,
	logger log.Logger,
) (*CollectorService, error) {
	// Validate dependencies
	if winpowerClient == nil {
		return nil, fmt.Errorf("%w: winpowerClient", ErrNilDependency)
	}
	if energyCalc == nil {
		return nil, fmt.Errorf("%w: energyCalc", ErrNilDependency)
	}
	if logger == nil {
		return nil, fmt.Errorf("%w: logger", ErrNilDependency)
	}

	return &CollectorService{
		winpowerClient: winpowerClient,
		energyCalc:     energyCalc,
		logger:         logger,
	}, nil
}

// CollectDeviceData is the main entry point for data collection
func (cs *CollectorService) CollectDeviceData(ctx context.Context) (*CollectionResult, error) {
	if ctx == nil {
		return nil, ErrInvalidContext
	}

	start := time.Now()
	cs.logger.Debug("Starting device data collection")

	// Collect data from WinPower
	devices, err := cs.collectFromWinPower(ctx)
	if err != nil {
		cs.logger.Error("Failed to collect data from WinPower", log.Err(err))
		return &CollectionResult{
			Success:        false,
			DeviceCount:    0,
			Devices:        make(map[string]*DeviceCollectionInfo),
			CollectionTime: time.Now(),
			Duration:       time.Since(start),
			ErrorMessage:   err.Error(),
		}, err
	}

	// Process device data and trigger energy calculations
	result := cs.processDeviceData(ctx, devices, start)

	cs.logger.Info("Device data collection completed",
		log.Int("device_count", result.DeviceCount),
		log.Bool("success", result.Success),
		log.Duration("duration", result.Duration))

	return result, nil
}

// collectFromWinPower collects device data from WinPower module
func (cs *CollectorService) collectFromWinPower(ctx context.Context) ([]winpower.ParsedDeviceData, error) {
	devices, err := cs.winpowerClient.CollectDeviceData(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrWinPowerCollection, err)
	}

	cs.logger.Debug("Successfully collected data from WinPower",
		log.Int("device_count", len(devices)))

	return devices, nil
}

// processDeviceData processes each device and triggers energy calculation
func (cs *CollectorService) processDeviceData(
	ctx context.Context,
	devices []winpower.ParsedDeviceData,
	startTime time.Time,
) *CollectionResult {
	result := &CollectionResult{
		Success:        true,
		DeviceCount:    len(devices),
		Devices:        make(map[string]*DeviceCollectionInfo),
		CollectionTime: time.Now(),
	}

	for _, device := range devices {
		deviceInfo := cs.convertToDeviceInfo(device)

		// Trigger energy calculation for each device
		if err := cs.calculateEnergy(device.DeviceID, device.Realtime.LoadTotalWatt, deviceInfo); err != nil {
			cs.logger.Warn("Energy calculation failed for device",
				log.String("device_id", device.DeviceID),
				log.Err(err))
			// Continue processing other devices even if one fails
		}

		result.Devices[device.DeviceID] = deviceInfo
	}

	result.Duration = time.Since(startTime)
	return result
}

// calculateEnergy triggers energy calculation and updates device info
func (cs *CollectorService) calculateEnergy(
	deviceID string,
	power float64,
	deviceInfo *DeviceCollectionInfo,
) error {
	energy, err := cs.energyCalc.Calculate(deviceID, power)
	if err != nil {
		deviceInfo.EnergyCalculated = false
		deviceInfo.ErrorMsg = fmt.Sprintf("energy calculation failed: %v", err)
		return fmt.Errorf("%w: %v", ErrEnergyCalculation, err)
	}

	deviceInfo.EnergyCalculated = true
	deviceInfo.EnergyValue = energy
	return nil
}

// convertToDeviceInfo converts WinPower data to DeviceCollectionInfo
func (cs *CollectorService) convertToDeviceInfo(device winpower.ParsedDeviceData) *DeviceCollectionInfo {
	return &DeviceCollectionInfo{
		// Basic information
		DeviceID:       device.DeviceID,
		DeviceName:     device.Alias,
		DeviceType:     device.DeviceType,
		DeviceModel:    device.Model,
		Connected:      device.Connected,
		LastUpdateTime: device.CollectedAt,

		// Electrical parameters
		InputVolt1:     device.Realtime.InputVolt1,
		InputFreq:      device.Realtime.InputFreq,
		OutputVolt1:    device.Realtime.OutputVolt1,
		OutputCurrent1: device.Realtime.OutputCurrent1,
		OutputFreq:     device.Realtime.OutputFreq,

		// Load and power parameters
		LoadPercent:   device.Realtime.LoadPercent,
		LoadTotalWatt: device.Realtime.LoadTotalWatt,
		LoadTotalVa:   device.Realtime.LoadTotalVa,
		LoadWatt1:     device.Realtime.LoadWatt1,
		LoadVa1:       device.Realtime.LoadVa1,

		// Battery parameters
		IsCharging:    device.Realtime.IsCharging,
		BatVoltP:      device.Realtime.BatVoltP,
		BatCapacity:   device.Realtime.BatCapacity,
		BatRemainTime: device.Realtime.BatRemainTime,
		BatteryStatus: device.Realtime.BatteryStatus,

		// UPS status parameters
		UpsTemperature: device.Realtime.UpsTemperature,
		Mode:           device.Realtime.Mode,
		Status:         device.Realtime.Status,
		TestStatus:     device.Realtime.TestStatus,
		FaultCode:      device.Realtime.FaultCode,

		// Initialize energy fields (will be updated by calculateEnergy)
		EnergyCalculated: false,
		EnergyValue:      0,
		ErrorMsg:         "",
	}
}
