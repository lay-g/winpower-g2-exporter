// Package collector provides data collection and coordination for WinPower G2 Exporter.
//
// The collector module is the central coordinator in the system architecture,
// responsible for:
//   - Collecting device data from WinPower module
//   - Triggering energy calculations in the Energy module
//   - Coordinating data flow between components
//   - Providing unified collection interface for Scheduler and Metrics modules
//
// Key Design Principles:
//   - Single Responsibility: Focus on data collection coordination
//   - Dependency Inversion: Depend on interfaces, not concrete implementations
//   - Error Isolation: Single device failures don't affect overall collection
//   - Performance First: Minimize overhead in data collection and calculation
//
// Usage Example:
//
//	// Create collector service
//	collector := collector.NewCollectorService(winpowerClient, energyService, logger)
//
//	// Collect device data
//	result, err := collector.CollectDeviceData(ctx)
//	if err != nil {
//	    log.Fatal("Collection failed:", err)
//	}
//
//	// Process results
//	for deviceID, info := range result.Devices {
//	    fmt.Printf("Device %s: Power=%.2fW, Energy=%.2fWh\n",
//	        deviceID, info.LoadTotalWatt, info.EnergyValue)
//	}
package collector
