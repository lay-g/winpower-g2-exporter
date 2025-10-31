// Package metrics provides metrics management and exposure for the WinPower G2 Prometheus Exporter.
//
// This package is responsible for:
//   - Managing Prometheus metrics registry
//   - Coordinating with Collector module to fetch latest device data
//   - Exposing metrics via HTTP handler for Prometheus scraping
//   - Maintaining exporter self-monitoring metrics and WinPower device metrics
//
// The metrics module serves as the metrics management center, providing a standard
// Gin Handler as the single entry point to serve Prometheus-formatted metric data.
//
// Example usage:
//
//	// Create metrics service
//	metricsService := metrics.NewMetricsService(collector, logger)
//
//	// Use with Gin router
//	router := gin.Default()
//	router.GET("/metrics", metricsService.HandleMetrics)
//
// The module manages two categories of metrics:
//   - Exporter self-monitoring metrics: track exporter health and performance
//   - WinPower device metrics: track device status, electrical parameters, and energy consumption
package metrics
