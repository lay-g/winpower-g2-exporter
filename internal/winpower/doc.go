// Package winpower provides client implementation for WinPower G2 system data collection.
//
// The winpower package implements the core data collection capabilities for interacting
// with WinPower G2 systems, including authentication, token management, device data
// collection, data parsing, and error handling.
//
// Core Components:
//   - WinPowerClient: Unified interface for data collection
//   - HTTPClient: HTTP communication management
//   - TokenManager: Token lifecycle management
//   - DataParser: Data parsing and validation
//
// Usage Example:
//
//	cfg := winpower.DefaultConfig()
//	cfg.BaseURL = "https://winpower.example.com"
//	cfg.Username = "admin"
//	cfg.Password = "secret"
//
//	client, err := winpower.NewClient(cfg, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	data, err := client.CollectDeviceData(context.Background())
//	if err != nil {
//	    log.Printf("Failed to collect data: %v", err)
//	}
package winpower
